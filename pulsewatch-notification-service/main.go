package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/smtp"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeName                      = "pulsewatch.direct"
	EmailQueue                        = "notifications.email"
	RoutingKeyEmailVerification       = "notification.email.verification.requested"
	RoutingKeyEmailPasswordReset      = "notification.email.password_reset.requested"
	RoutingKeyEmailResendVerification = "notification.email.resend_verification.requested"
)

type BaseEvent struct {
	EventID       uuid.UUID       `json:"event_id"`
	EventType     string          `json:"event_type"`
	SchemaVersion string          `json:"schema_version"`
	OccurredAt    time.Time       `json:"occurred_at"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	Payload       json.RawMessage `json:"payload"`
}

type VerificationEmailPayload struct {
	UserID            uuid.UUID `json:"user_id"`
	Email             string    `json:"email"`
	Name              string    `json:"name"`
	VerificationToken string    `json:"verification_token"`
	ExpiresAt         time.Time `json:"expires_at"`
}

type PasswordResetEmailPayload struct {
	UserID     uuid.UUID `json:"user_id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	ResetToken string    `json:"reset_token"`
	ExpiresAt  time.Time `json:"expires_at"`
}

type EmailSender struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

func NewEmailSender() *EmailSender {
	port, err := strconv.Atoi(getEnv("SMTP_PORT", "1025"))
	if err != nil {
		port = 1025
	}
	return &EmailSender{
		host:     getEnv("SMTP_HOST", "localhost"),
		port:     port,
		user:     getEnv("SMTP_USER", ""),
		password: getEnv("SMTP_PASSWORD", ""),
		from:     getEnv("SMTP_FROM", "PulseWatch <alerts@pulsewatch.dev>"),
	}
}

func (s *EmailSender) Send(to, subject, textBody, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	var auth smtp.Auth
	if s.user != "" {
		auth = smtp.PlainAuth("", s.user, s.password, s.host)
	}

	boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary))

	// Plain text
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	msg.WriteString(textBody)
	msg.WriteString("\r\n\r\n")

	// HTML
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n\r\n")

	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	// Extract raw email address for MAIL FROM envelope
	fromAddr := s.from
	if start := strings.Index(fromAddr, "<"); start != -1 {
		if end := strings.Index(fromAddr, ">"); end != -1 && end > start {
			fromAddr = fromAddr[start+1 : end]
		}
	}

	return smtp.SendMail(addr, auth, fromAddr, []string{to}, msg.Bytes())
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	_ = godotenv.Load(".env", "../.env")

	rmqURL := getEnv("RABBITMQ_URL", "amqp://pulsewatch:pulsewatch_secret@localhost:5672/")
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:3000")
	emailSender := NewEmailSender()

	logger.Info("Starting PulseWatch Notification Service...",
		"smtp_host", emailSender.host,
		"smtp_port", emailSender.port,
		"smtp_from", emailSender.from,
	)

	conn, err := amqp.Dial(rmqURL)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		logger.Error("Failed to open RabbitMQ channel", "error", err)
		os.Exit(1)
	}
	defer ch.Close()

	// 1. Declare Exchange
	err = ch.ExchangeDeclare(
		ExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Error("Failed to declare exchange", "error", err)
		os.Exit(1)
	}

	// 2. Declare Queue
	_, err = ch.QueueDeclare(
		EmailQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Error("Failed to declare queue", "error", err)
		os.Exit(1)
	}

	// 3. Bind routing keys
	for _, key := range []string{RoutingKeyEmailVerification, RoutingKeyEmailPasswordReset, RoutingKeyEmailResendVerification} {
		if err := ch.QueueBind(EmailQueue, key, ExchangeName, false, nil); err != nil {
			logger.Error("Failed to bind queue", "key", key, "error", err)
			os.Exit(1)
		}
	}

	// 4. Set prefetch count
	if err := ch.Qos(10, 0, false); err != nil {
		logger.Error("Failed to set QoS", "error", err)
		os.Exit(1)
	}

	msgs, err := ch.Consume(
		EmailQueue,
		"notification-worker",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Error("Failed to register consumer", "error", err)
		os.Exit(1)
	}

	logger.Info("Notification Service listening for email jobs on queue", "queue", EmailQueue)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for d := range msgs {
			start := time.Now()
			var evt BaseEvent
			if err := json.Unmarshal(d.Body, &evt); err != nil {
				logger.Error("Failed to deserialize event envelope", "error", err)
				_ = d.Nack(false, false)
				continue
			}

			var sendErr error
			var recipient string

			switch d.RoutingKey {
			case RoutingKeyEmailVerification, RoutingKeyEmailResendVerification:
				var p VerificationEmailPayload
				if err := json.Unmarshal(evt.Payload, &p); err != nil {
					logger.Error("Failed to deserialize verification payload", "error", err)
					_ = d.Nack(false, false)
					continue
				}

				recipient = p.Email
				verificationLink := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, p.VerificationToken)
				subject := "Verify your PulseWatch account"
				textBody := fmt.Sprintf("Hi %s,\n\nPlease verify your email address by visiting:\n%s\n\nThis link will expire at %s.\n\nThanks,\nPulseWatch Team",
					p.Name, verificationLink, p.ExpiresAt.Format(time.RFC1123))

				htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #0b0f19; color: #f3f4f6; margin: 0; padding: 40px 20px; }
    .card { max-width: 540px; margin: 0 auto; background-color: #111827; border: 1px solid #1f2937; border-radius: 12px; padding: 32px; box-shadow: 0 10px 25px rgba(0,0,0,0.5); }
    .logo { font-size: 24px; font-weight: bold; color: #38bdf8; margin-bottom: 24px; letter-spacing: -0.5px; }
    h1 { font-size: 20px; color: #ffffff; margin-bottom: 16px; }
    p { font-size: 14px; line-height: 1.6; color: #9ca3af; margin-bottom: 24px; }
    .btn { display: inline-block; background-color: #0284c7; color: #ffffff !important; text-decoration: none; padding: 12px 24px; border-radius: 8px; font-weight: 600; font-size: 14px; }
    .btn:hover { background-color: #0369a1; }
    .footer { font-size: 12px; color: #6b7280; margin-top: 32px; border-top: 1px solid #1f2937; padding-top: 16px; }
  </style>
</head>
<body>
  <div class="card">
    <div class="logo">⚡ PulseWatch</div>
    <h1>Verify your email address</h1>
    <p>Hi %s,</p>
    <p>Welcome to PulseWatch! Please confirm your email address by clicking the button below to activate your monitoring workspace.</p>
    <p><a href="%s" class="btn" target="_blank">Verify Email Address</a></p>
    <p>If the button doesn't work, copy and paste this URL into your browser:<br><span style="color: #38bdf8; word-break: break-all;">%s</span></p>
    <div class="footer">
      This verification link expires at %s.<br>If you did not sign up for PulseWatch, you can safely ignore this email.
    </div>
  </div>
</body>
</html>`, p.Name, verificationLink, verificationLink, p.ExpiresAt.Format(time.RFC1123))

				sendErr = emailSender.Send(p.Email, subject, textBody, htmlBody)

			case RoutingKeyEmailPasswordReset:
				var p PasswordResetEmailPayload
				if err := json.Unmarshal(evt.Payload, &p); err != nil {
					logger.Error("Failed to deserialize password reset payload", "error", err)
					_ = d.Nack(false, false)
					continue
				}

				recipient = p.Email
				resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, p.ResetToken)
				subject := "Reset your PulseWatch password"
				textBody := fmt.Sprintf("Hi %s,\n\nA password reset was requested for your PulseWatch account. Visit:\n%s\n\nThis link will expire at %s.\n\nIf you did not request this, please ignore this email.\n\nThanks,\nPulseWatch Team",
					p.Name, resetLink, p.ExpiresAt.Format(time.RFC1123))

				htmlBody := fmt.Sprintf(`<!DOCTYPE html>
						<html>
						<head>
						<meta charset="utf-8">
						<style>
							body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #0b0f19; color: #f3f4f6; margin: 0; padding: 40px 20px; }
							.card { max-width: 540px; margin: 0 auto; background-color: #111827; border: 1px solid #1f2937; border-radius: 12px; padding: 32px; box-shadow: 0 10px 25px rgba(0,0,0,0.5); }
							.logo { font-size: 24px; font-weight: bold; color: #38bdf8; margin-bottom: 24px; letter-spacing: -0.5px; }
							h1 { font-size: 20px; color: #ffffff; margin-bottom: 16px; }
							p { font-size: 14px; line-height: 1.6; color: #9ca3af; margin-bottom: 24px; }
							.btn { display: inline-block; background-color: #dc2626; color: #ffffff !important; text-decoration: none; padding: 12px 24px; border-radius: 8px; font-weight: 600; font-size: 14px; }
							.btn:hover { background-color: #b91c1c; }
							.footer { font-size: 12px; color: #6b7280; margin-top: 32px; border-top: 1px solid #1f2937; padding-top: 16px; }
						</style>
						</head>
						<body>
						<div class="card">
							<div class="logo">⚡ PulseWatch</div>
							<h1>Reset your password</h1>
							<p>Hi %s,</p>
							<p>We received a request to reset the password for your PulseWatch account. Click the button below to choose a new password.</p>
							<p><a href="%s" class="btn" target="_blank">Reset Password</a></p>
							<p>If the button doesn't work, copy and paste this URL into your browser:<br><span style="color: #38bdf8; word-break: break-all;">%s</span></p>
							<div class="footer">
							This link expires at %s.<br>If you did not request a password reset, please ignore this email or review your security settings.
							</div>
						</div>
						</body>
						</html>`, p.Name, resetLink, resetLink, p.ExpiresAt.Format(time.RFC1123))

				sendErr = emailSender.Send(p.Email, subject, textBody, htmlBody)

			default:
				logger.Warn("Unknown routing key received", "routing_key", d.RoutingKey)
			}

			if sendErr != nil {
				logger.Error("Failed to deliver email via SMTP (Mailpit)",
					"recipient", maskEmail(recipient),
					"error", sendErr,
				)
				_ = d.Nack(false, true) // Requeue on transient SMTP error
				continue
			}

			logger.Info("Email successfully sent via SMTP (Mailpit)",
				"event_id", evt.EventID,
				"event_type", d.RoutingKey,
				"recipient", maskEmail(recipient),
				"duration_ms", time.Since(start).Milliseconds(),
			)

			if err := d.Ack(false); err != nil {
				logger.Error("Failed to ack message", "error", err)
			}
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	logger.Info("PulseWatch Notification Service stopped gracefully")
	_ = ctx
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}
	name := parts[0]
	if len(name) <= 2 {
		return "**@" + parts[1]
	}
	return string(name[0]) + "***" + string(name[len(name)-1]) + "@" + parts[1]
}
