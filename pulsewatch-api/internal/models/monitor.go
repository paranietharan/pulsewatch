package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Monitor struct {
	ID                   uuid.UUID         `json:"id"`
	UserID               uuid.UUID         `json:"user_id"`
	Name                 string            `json:"name"`
	URL                  string            `json:"url"`
	Method               string            `json:"method"`
	IntervalSeconds      int               `json:"interval_seconds"`
	TimeoutSeconds       int               `json:"timeout_seconds"`
	ExpectedStatusCode   int               `json:"expected_status_code"`
	Headers              map[string]string `json:"headers"`
	Body                 string            `json:"body,omitempty"`
	Status               string            `json:"status"` // PENDING, UP, DOWN, DEGRADED
	IsActive             bool              `json:"is_active"`
	SSLCheckEnabled      bool              `json:"ssl_check_enabled"`
	FailureThreshold     int               `json:"failure_threshold"`
	RecoveryThreshold    int               `json:"recovery_threshold"`
	ConsecutiveFailures  int               `json:"consecutive_failures"`
	ConsecutiveSuccesses int               `json:"consecutive_successes"`
	LastCheckedAt        *time.Time        `json:"last_checked_at"`
	NextCheckAt          time.Time         `json:"next_check_at"`
	Uptime24h            float64           `json:"uptime_24h"`
	Uptime7d             float64           `json:"uptime_7d"`
	Uptime30d            float64           `json:"uptime_30d"`
	AvgResponseTimeMs    int               `json:"avg_response_time_ms"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}

type Heartbeat struct {
	ID             int64     `json:"id"`
	MonitorID      uuid.UUID `json:"monitor_id"`
	Status         string    `json:"status"` // UP, DOWN, DEGRADED
	StatusCode     int       `json:"status_code"`
	ResponseTimeMs int       `json:"response_time_ms"`
	DNSLookupMs    int       `json:"dns_lookup_ms"`
	TCPConnectMs   int       `json:"tcp_connect_ms"`
	TLSHandshakeMs int       `json:"tls_handshake_ms"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	Region         string    `json:"region"`
	CheckedAt      time.Time `json:"checked_at"`
	JobID          *uuid.UUID `json:"job_id,omitempty"`
}

type Incident struct {
	ID              uuid.UUID        `json:"id"`
	MonitorID       uuid.UUID        `json:"monitor_id"`
	MonitorName     string           `json:"monitor_name,omitempty"`
	Status          string           `json:"status"` // OPEN, ACKNOWLEDGED, RESOLVED
	CauseSummary    string           `json:"cause_summary"`
	StartedAt       time.Time        `json:"started_at"`
	ResolvedAt      *time.Time       `json:"resolved_at,omitempty"`
	DurationSeconds *int             `json:"duration_seconds,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	Updates         []IncidentUpdate `json:"updates,omitempty"`
}

type IncidentUpdate struct {
	ID         int64     `json:"id"`
	IncidentID uuid.UUID `json:"incident_id"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

type AlertChannel struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Type      string          `json:"type"` // EMAIL, WEBHOOK, SLACK, etc.
	Name      string          `json:"name"`
	Config    json.RawMessage `json:"config"`
	IsEnabled bool            `json:"is_enabled"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type DashboardSummary struct {
	TotalMonitors     int     `json:"total_monitors"`
	HealthyMonitors   int     `json:"healthy_monitors"`
	DegradedMonitors  int     `json:"degraded_monitors"`
	DownMonitors      int     `json:"down_monitors"`
	PausedMonitors    int     `json:"paused_monitors"`
	ActiveIncidents   int     `json:"active_incidents"`
	AvgUptime24h      float64 `json:"avg_uptime_24h"`
	AvgResponseTimeMs int     `json:"avg_response_time_ms"`
}
