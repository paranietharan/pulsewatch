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

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	_ = godotenv.Load(".env", "../.env")

	rmqURL := os.Getenv("RABBITMQ_URL")
	if rmqURL == "" {
		rmqURL = "amqp://pulsewatch:pulsewatch_secret@localhost:5672/"
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	logger.Info("Starting PulseWatch Notification Service...")

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
		false, // exclusive
		false, // no-local
		false, // no-wait
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
				_ = d.Nack(false, false) // dead-letter invalid message
				continue
			}

			switch d.RoutingKey {
			case RoutingKeyEmailVerification, RoutingKeyEmailResendVerification:
				var p VerificationEmailPayload
				if err := json.Unmarshal(evt.Payload, &p); err != nil {
					logger.Error("Failed to deserialize verification payload", "error", err)
					_ = d.Nack(false, false)
					continue
				}

				verificationLink := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, p.VerificationToken)
				_ = verificationLink // Link generated for email transmission

				logger.Info("Verification email processed",
					"event_id", evt.EventID,
					"event_type", d.RoutingKey,
					"recipient", maskEmail(p.Email),
					"expires_at", p.ExpiresAt.Format(time.RFC3339),
					"duration_ms", time.Since(start).Milliseconds(),
				)

			case RoutingKeyEmailPasswordReset:
				var p PasswordResetEmailPayload
				if err := json.Unmarshal(evt.Payload, &p); err != nil {
					logger.Error("Failed to deserialize password reset payload", "error", err)
					_ = d.Nack(false, false)
					continue
				}

				resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, p.ResetToken)
				_ = resetLink // Link generated for email transmission

				logger.Info("Password reset email processed",
					"event_id", evt.EventID,
					"event_type", d.RoutingKey,
					"recipient", maskEmail(p.Email),
					"expires_at", p.ExpiresAt.Format(time.RFC3339),
					"duration_ms", time.Since(start).Milliseconds(),
				)

			default:
				logger.Warn("Unknown routing key received", "routing_key", d.RoutingKey)
			}

			// Acknowledge message successfully
			if err := d.Ack(false); err != nil {
				logger.Error("Failed to ack message", "error", err)
			}
		}
	}()

	// Wait for termination signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	logger.Info("PulseWatch Notification Service stopped gracefully")
	_ = ctx
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
