package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"pulsewatch-notification-service/internal/notifier"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeName                      = "pulsewatch.direct"
	EmailQueue                        = "notifications.email"
	WebhookQueue                      = "notifications.webhook"
	DeliveryOutcomeQueue              = "notifications.delivery_outcome"

	RoutingKeyEmailVerification       = "notification.email.verification.requested"
	RoutingKeyEmailPasswordReset      = "notification.email.password_reset.requested"
	RoutingKeyEmailResendVerification = "notification.email.resend_verification.requested"
	RoutingKeyEmailIncidentOpened     = "notification.email.incident.opened"
	RoutingKeyEmailIncidentResolved   = "notification.email.incident.resolved"
	RoutingKeyWebhookDispatch         = "notification.webhook.dispatch"
	RoutingKeyDeliveryOutcome         = "notification.delivery.outcome"
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

type IncidentOpenedEmailPayload struct {
	MonitorID   uuid.UUID `json:"monitor_id"`
	MonitorName string    `json:"monitor_name"`
	MonitorURL  string    `json:"monitor_url"`
	Email       string    `json:"email"`
	Cause       string    `json:"cause"`
	StartedAt   time.Time `json:"started_at"`
}

type IncidentResolvedEmailPayload struct {
	MonitorID       uuid.UUID `json:"monitor_id"`
	MonitorName     string    `json:"monitor_name"`
	MonitorURL      string    `json:"monitor_url"`
	Email           string    `json:"email"`
	DurationSeconds int       `json:"duration_seconds"`
	ResolvedAt      time.Time `json:"resolved_at"`
}

type WebhookDispatchPayload struct {
	DeliveryID      uuid.UUID  `json:"delivery_id"`
	ChannelID       uuid.UUID  `json:"channel_id"`
	TargetURL       string     `json:"target_url"`
	Secret          string     `json:"secret,omitempty"`
	EventType       string     `json:"event_type"`
	IncidentID      uuid.UUID  `json:"incident_id"`
	MonitorID       uuid.UUID  `json:"monitor_id"`
	MonitorName     string     `json:"monitor_name"`
	Status          string     `json:"status"`
	CauseSummary    string     `json:"cause_summary,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty"`
	Timestamp       time.Time  `json:"timestamp"`
}

type DeliveryOutcomePayload struct {
	DeliveryID   uuid.UUID `json:"delivery_id"`
	IncidentID   uuid.UUID `json:"incident_id"`
	ChannelID    uuid.UUID `json:"channel_id"`
	ChannelType  string    `json:"channel_type"`
	Status       string    `json:"status"`
	AttemptCount int       `json:"attempt_count"`
	LastError    string    `json:"last_error,omitempty"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	_ = godotenv.Load(".env", "../.env")

	rmqURL := getEnv("RABBITMQ_URL", "amqp://pulsewatch:pulsewatch_secret@localhost:5672/")
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:3000")

	logger.Info("Starting PulseWatch Notification Dispatch Service (Database-Free)...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to RabbitMQ
	rmqConn, err := amqp.Dial(rmqURL)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer rmqConn.Close()
	logger.Info("Connected to RabbitMQ successfully")

	// Initialize Notifiers
	emailSender := notifier.NewEmailSender()
	webhookNotifier := notifier.NewWebhookNotifier()

	// Start Email Dispatch Consumer in background
	go startEmailConsumer(ctx, rmqConn, emailSender, frontendURL, logger)

	// Start Webhook Dispatch Consumer in background
	go startWebhookConsumer(ctx, rmqConn, webhookNotifier, logger)

	// Graceful Shutdown on SIGINT / SIGTERM
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	sig := <-shutdown
	logger.Info("Shutdown signal received, shutting down gracefully...", "signal", sig.String())
	cancel()

	time.Sleep(2 * time.Second)
	logger.Info("PulseWatch Notification Service stopped gracefully")
}

func startEmailConsumer(ctx context.Context, conn *amqp.Connection, sender *notifier.EmailSender, frontendURL string, logger *slog.Logger) {
	ch, err := conn.Channel()
	if err != nil {
		logger.Error("Failed to open email channel", "error", err)
		return
	}
	defer ch.Close()

	// Declare exchange and email queue
	_ = ch.ExchangeDeclare(ExchangeName, "direct", true, false, false, false, nil)
	_, _ = ch.QueueDeclare(EmailQueue, true, false, false, false, nil)

	emailBindings := []string{
		RoutingKeyEmailVerification,
		RoutingKeyEmailPasswordReset,
		RoutingKeyEmailResendVerification,
		RoutingKeyEmailIncidentOpened,
		RoutingKeyEmailIncidentResolved,
	}
	for _, key := range emailBindings {
		_ = ch.QueueBind(EmailQueue, key, ExchangeName, false, nil)
	}

	_ = ch.Qos(10, 0, false)

	msgs, err := ch.Consume(
		EmailQueue,
		"notification-email-worker",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Error("Failed to register email consumer", "error", err)
		return
	}

	logger.Info("Email notification worker listening on queue", "queue", EmailQueue)

	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}
			processEmailDelivery(d, sender, frontendURL, logger)
		}
	}
}

func processEmailDelivery(d amqp.Delivery, sender *notifier.EmailSender, frontendURL string, logger *slog.Logger) {
	start := time.Now()

	var payloadBytes []byte
	var evt BaseEvent
	if err := json.Unmarshal(d.Body, &evt); err == nil && len(evt.Payload) > 0 {
		payloadBytes = evt.Payload
	} else {
		payloadBytes = d.Body
	}

	var sendErr error
	var recipient string

	switch d.RoutingKey {
	case RoutingKeyEmailVerification, RoutingKeyEmailResendVerification:
		var p VerificationEmailPayload
		if err := json.Unmarshal(payloadBytes, &p); err != nil {
			logger.Error("Failed to deserialize verification payload", "error", err)
			_ = d.Nack(false, false)
			return
		}

		recipient = p.Email
		verificationLink := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, p.VerificationToken)
		subject := "Verify your PulseWatch account"
		textBody := fmt.Sprintf("Hi %s,\n\nPlease verify your email address by visiting:\n%s\n\nThis link expires at %s.\n\nPulseWatch Team",
			p.Name, verificationLink, p.ExpiresAt.Format(time.RFC1123))

		htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: -apple-system, sans-serif; background: #0b0f19; color: #f3f4f6; padding: 30px;">
  <div style="max-width: 520px; margin: 0 auto; background: #111827; border: 1px solid #1f2937; border-radius: 12px; padding: 28px;">
    <div style="font-size: 22px; font-weight: bold; color: #38bdf8; margin-bottom: 16px;">⚡ PulseWatch</div>
    <h1 style="font-size: 18px; color: #fff;">Verify your email address</h1>
    <p style="color: #9ca3af; font-size: 14px; line-height: 1.5;">Hi %s, please click the button below to verify your account.</p>
    <p style="margin: 24px 0;"><a href="%s" style="background: #0284c7; color: #fff; padding: 12px 24px; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: 14px;" target="_blank">Verify Email Address</a></p>
    <p style="color: #6b7280; font-size: 12px;">Link expires at %s.</p>
  </div>
</body>
</html>`, p.Name, verificationLink, p.ExpiresAt.Format(time.RFC1123))

		sendErr = sender.Send(p.Email, subject, textBody, htmlBody)

	case RoutingKeyEmailPasswordReset:
		var p PasswordResetEmailPayload
		if err := json.Unmarshal(payloadBytes, &p); err != nil {
			logger.Error("Failed to deserialize password reset payload", "error", err)
			_ = d.Nack(false, false)
			return
		}

		recipient = p.Email
		resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, p.ResetToken)
		subject := "Reset your PulseWatch password"
		textBody := fmt.Sprintf("Hi %s,\n\nA password reset was requested. Visit:\n%s\n\nExpires at: %s.\n\nPulseWatch Team",
			p.Name, resetLink, p.ExpiresAt.Format(time.RFC1123))

		htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: -apple-system, sans-serif; background: #0b0f19; color: #f3f4f6; padding: 30px;">
  <div style="max-width: 520px; margin: 0 auto; background: #111827; border: 1px solid #1f2937; border-radius: 12px; padding: 28px;">
    <div style="font-size: 22px; font-weight: bold; color: #38bdf8; margin-bottom: 16px;">⚡ PulseWatch</div>
    <h1 style="font-size: 18px; color: #fff;">Reset your password</h1>
    <p style="color: #9ca3af; font-size: 14px; line-height: 1.5;">Hi %s, click below to set a new password for your account.</p>
    <p style="margin: 24px 0;"><a href="%s" style="background: #dc2626; color: #fff; padding: 12px 24px; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: 14px;" target="_blank">Reset Password</a></p>
    <p style="color: #6b7280; font-size: 12px;">Link expires at %s.</p>
  </div>
</body>
</html>`, p.Name, resetLink, p.ExpiresAt.Format(time.RFC1123))

		sendErr = sender.Send(p.Email, subject, textBody, htmlBody)

	case RoutingKeyEmailIncidentOpened:
		var p IncidentOpenedEmailPayload
		if err := json.Unmarshal(payloadBytes, &p); err != nil {
			logger.Error("Failed to deserialize incident opened payload", "error", err)
			_ = d.Nack(false, false)
			return
		}

		recipient = p.Email
		sendErr = sender.SendIncidentOpened(p.Email, p.MonitorName, p.MonitorURL, p.Cause, p.StartedAt)

	case RoutingKeyEmailIncidentResolved:
		var p IncidentResolvedEmailPayload
		if err := json.Unmarshal(payloadBytes, &p); err != nil {
			logger.Error("Failed to deserialize incident resolved payload", "error", err)
			_ = d.Nack(false, false)
			return
		}

		recipient = p.Email
		sendErr = sender.SendIncidentResolved(p.Email, p.MonitorName, p.MonitorURL, p.DurationSeconds, p.ResolvedAt)

	default:
		logger.Warn("Unknown routing key received on email queue", "routing_key", d.RoutingKey)
	}

	if sendErr != nil {
		logger.Error("Failed to deliver email via SMTP",
			"recipient", maskEmail(recipient),
			"error", sendErr,
		)
		_ = d.Nack(false, true)
		return
	}

	logger.Info("Email successfully delivered via SMTP",
		"event_type", d.RoutingKey,
		"recipient", maskEmail(recipient),
		"duration_ms", time.Since(start).Milliseconds(),
	)

	_ = d.Ack(false)
}

func startWebhookConsumer(ctx context.Context, conn *amqp.Connection, notifierInstance *notifier.WebhookNotifier, logger *slog.Logger) {
	ch, err := conn.Channel()
	if err != nil {
		logger.Error("Failed to open webhook channel", "error", err)
		return
	}
	defer ch.Close()

	pubCh, err := conn.Channel()
	if err != nil {
		logger.Error("Failed to open webhook response publisher channel", "error", err)
		return
	}
	defer pubCh.Close()

	_ = ch.ExchangeDeclare(ExchangeName, "direct", true, false, false, false, nil)
	_, _ = ch.QueueDeclare(WebhookQueue, true, false, false, false, nil)
	_ = ch.QueueBind(WebhookQueue, RoutingKeyWebhookDispatch, ExchangeName, false, nil)

	_ = ch.Qos(10, 0, false)

	msgs, err := ch.Consume(
		WebhookQueue,
		"notification-webhook-worker",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Error("Failed to register webhook consumer", "error", err)
		return
	}

	logger.Info("Webhook notification worker listening on queue", "queue", WebhookQueue)

	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}
			processWebhookDelivery(ctx, d, notifierInstance, pubCh, logger)
		}
	}
}

func processWebhookDelivery(ctx context.Context, d amqp.Delivery, notifierInstance *notifier.WebhookNotifier, pubCh *amqp.Channel, logger *slog.Logger) {
	start := time.Now()

	var payloadBytes []byte
	var evt BaseEvent
	if err := json.Unmarshal(d.Body, &evt); err == nil && len(evt.Payload) > 0 {
		payloadBytes = evt.Payload
	} else {
		payloadBytes = d.Body
	}

	var p WebhookDispatchPayload
	if err := json.Unmarshal(payloadBytes, &p); err != nil {
		logger.Error("Failed to deserialize webhook dispatch payload", "error", err)
		_ = d.Nack(false, false)
		return
	}

	webhookPayload := notifier.WebhookPayload{
		EventID:         p.DeliveryID,
		EventType:       p.EventType,
		IncidentID:      p.IncidentID,
		MonitorID:       p.MonitorID,
		MonitorName:     p.MonitorName,
		TargetURL:       p.TargetURL,
		Status:          p.Status,
		CauseSummary:    p.CauseSummary,
		StartedAt:       p.StartedAt,
		ResolvedAt:      p.ResolvedAt,
		DurationSeconds: p.DurationSeconds,
		Timestamp:       p.Timestamp,
	}

	status, attempts, deliverErr := notifierInstance.Deliver(ctx, p.DeliveryID, p.TargetURL, p.Secret, webhookPayload)
	var lastErrStr string
	if deliverErr != nil {
		lastErrStr = deliverErr.Error()
	}

	// Publish delivery outcome back to RabbitMQ for pulsewatch-api persistence
	outcome := DeliveryOutcomePayload{
		DeliveryID:   p.DeliveryID,
		IncidentID:   p.IncidentID,
		ChannelID:    p.ChannelID,
		ChannelType:  "WEBHOOK",
		Status:       status,
		AttemptCount: attempts,
		LastError:    lastErrStr,
	}

	if outcomeBytes, err := json.Marshal(outcome); err == nil {
		_ = pubCh.PublishWithContext(
			ctx,
			ExchangeName,
			RoutingKeyDeliveryOutcome,
			false,
			false,
			amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				Timestamp:    time.Now().UTC(),
				Body:         outcomeBytes,
			},
		)
	}

	logger.Info("Webhook dispatch completed",
		"delivery_id", p.DeliveryID,
		"target_url", p.TargetURL,
		"status", status,
		"attempts", attempts,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	_ = d.Ack(false)
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
