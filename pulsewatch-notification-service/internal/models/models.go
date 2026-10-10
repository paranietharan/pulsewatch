package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

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

type Monitor struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	Name                 string
	URL                  string
	Method               string
	Status               string
	IsActive             bool
	FailureThreshold     int
	RecoveryThreshold    int
	ConsecutiveFailures  int
	ConsecutiveSuccesses int
	AvgResponseTimeMs    int
}

type Incident struct {
	ID              uuid.UUID
	MonitorID       uuid.UUID
	Status          string
	CauseSummary    string
	StartedAt       time.Time
	ResolvedAt      *time.Time
	DurationSeconds *int
}

type AlertChannel struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      string // EMAIL, WEBHOOK, SLACK, etc.
	Name      string
	Config    json.RawMessage
	IsEnabled bool
}

type WebhookConfig struct {
	URL    string `json:"url"`
	Secret string `json:"secret,omitempty"`
}

type EmailConfig struct {
	Email string `json:"email"`
}

type AlertDelivery struct {
	ID             uuid.UUID
	IncidentID     uuid.UUID
	AlertChannelID *uuid.UUID
	ChannelType    string
	Status         string // PENDING, SENDING, DELIVERED, FAILED
	AttemptCount   int
	LastError      string
	Payload        json.RawMessage
}
