package rabbitmq

import (
	"time"

	"github.com/google/uuid"
)

const (
	ExchangeName         = "pulsewatch.direct"
	EmailQueue           = "notifications.email"
	WebhookQueue         = "notifications.webhook"
	DeliveryOutcomeQueue = "notifications.delivery_outcome"
	ResultsQueue         = "monitor.results"

	RoutingKeyEmailVerification       = "notification.email.verification.requested"
	RoutingKeyEmailPasswordReset      = "notification.email.password_reset.requested"
	RoutingKeyEmailResendVerification = "notification.email.resend_verification.requested"
	RoutingKeyEmailIncidentOpened     = "notification.email.incident.opened"
	RoutingKeyEmailIncidentResolved   = "notification.email.incident.resolved"
	RoutingKeyWebhookDispatch         = "notification.webhook.dispatch"
	RoutingKeyDeliveryOutcome         = "notification.delivery.outcome"
	RoutingKeyResult                  = "monitor.result"
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
	FrontendURL       string    `json:"frontend_url,omitempty"`
}

// PasswordResetEmailPayload contains details for sending password reset emails.
type PasswordResetEmailPayload struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	ResetToken  string    `json:"reset_token"` // Raw single-use token sent ONLY to email provider
	ExpiresAt   time.Time `json:"expires_at"`
	FrontendURL string    `json:"frontend_url,omitempty"`
}

// IncidentOpenedEmailPayload contains details for sending outage alert emails.
type IncidentOpenedEmailPayload struct {
	MonitorID   uuid.UUID `json:"monitor_id"`
	MonitorName string    `json:"monitor_name"`
	MonitorURL  string    `json:"monitor_url"`
	Email       string    `json:"email"`
	Cause       string    `json:"cause"`
	StartedAt   time.Time `json:"started_at"`
}

// IncidentResolvedEmailPayload contains details for sending recovery alert emails.
type IncidentResolvedEmailPayload struct {
	MonitorID       uuid.UUID `json:"monitor_id"`
	MonitorName     string    `json:"monitor_name"`
	MonitorURL      string    `json:"monitor_url"`
	Email           string    `json:"email"`
	DurationSeconds int       `json:"duration_seconds"`
	ResolvedAt      time.Time `json:"resolved_at"`
}

// WebhookDispatchPayload contains details for sending webhook alerts.
type WebhookDispatchPayload struct {
	DeliveryID      uuid.UUID  `json:"delivery_id"`
	ChannelID       uuid.UUID  `json:"channel_id"`
	TargetURL       string     `json:"target_url"`
	Secret          string     `json:"secret,omitempty"`
	EventType       string     `json:"event_type"` // incident.opened, incident.resolved
	IncidentID      uuid.UUID  `json:"incident_id"`
	MonitorID       uuid.UUID  `json:"monitor_id"`
	MonitorName     string     `json:"monitor_name"`
	Status          string     `json:"status"` // OPEN, RESOLVED
	CauseSummary    string     `json:"cause_summary,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty"`
	Timestamp       time.Time  `json:"timestamp"`
}

// DeliveryOutcomePayload records the outcome of a webhook or notification attempt.
type DeliveryOutcomePayload struct {
	DeliveryID   uuid.UUID `json:"delivery_id"`
	IncidentID   uuid.UUID `json:"incident_id"`
	ChannelID    uuid.UUID `json:"channel_id"`
	ChannelType  string    `json:"channel_type"`
	Status       string    `json:"status"` // DELIVERED, FAILED
	AttemptCount int       `json:"attempt_count"`
	LastError    string    `json:"last_error,omitempty"`
}

// CheckResult represents the structured outcome of a health probe published by workers.
type CheckResult struct {
	JobID          uuid.UUID `json:"job_id"`
	MonitorID      uuid.UUID `json:"monitor_id"`
	UserID         uuid.UUID `json:"user_id"`
	Status         string    `json:"status"` // UP, DOWN, DEGRADED
	StatusCode     int       `json:"status_code"`
	ResponseTimeMs int       `json:"response_time_ms"`
	DNSLookupMs    int       `json:"dns_lookup_ms"`
	TCPConnectMs   int       `json:"tcp_connect_ms"`
	TLSHandshakeMs int       `json:"tls_handshake_ms"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	Region         string    `json:"region"`
	CheckedAt      time.Time `json:"checked_at"`
}
