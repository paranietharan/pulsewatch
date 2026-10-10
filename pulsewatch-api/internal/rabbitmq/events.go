package rabbitmq

import (
	"time"

	"github.com/google/uuid"
)

const (
	ExchangeName = "pulsewatch.direct"
	EmailQueue   = "notifications.email"

	RoutingKeyEmailVerification       = "notification.email.verification.requested"
	RoutingKeyEmailPasswordReset      = "notification.email.password_reset.requested"
	RoutingKeyEmailResendVerification = "notification.email.resend_verification.requested"
)

// BaseEvent provides standard metadata envelope for all platform events.
type BaseEvent struct {
	EventID       uuid.UUID   `json:"event_id"`
	EventType     string      `json:"event_type"`
	SchemaVersion string      `json:"schema_version"`
	OccurredAt    time.Time   `json:"occurred_at"`
	CorrelationID string      `json:"correlation_id,omitempty"`
	Payload       interface{} `json:"payload"`
}

// VerificationEmailPayload contains details for sending account verification emails.
type VerificationEmailPayload struct {
	UserID            uuid.UUID `json:"user_id"`
	Email             string    `json:"email"`
	Name              string    `json:"name"`
	VerificationToken string    `json:"verification_token"` // Raw single-use token sent ONLY to email provider
	ExpiresAt         time.Time `json:"expires_at"`
}

// PasswordResetEmailPayload contains details for sending password reset emails.
type PasswordResetEmailPayload struct {
	UserID     uuid.UUID `json:"user_id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	ResetToken string    `json:"reset_token"` // Raw single-use token sent ONLY to email provider
	ExpiresAt  time.Time `json:"expires_at"`
}
