package models

import (
	"time"

	"github.com/google/uuid"
)

// CheckJob represents an individual probe execution task published to RabbitMQ.
type CheckJob struct {
	JobID              uuid.UUID         `json:"job_id"`
	MonitorID          uuid.UUID         `json:"monitor_id"`
	UserID             uuid.UUID         `json:"user_id"`
	Name               string            `json:"name"`
	URL                string            `json:"url"`
	Method             string            `json:"method"`
	TimeoutSeconds     int               `json:"timeout_seconds"`
	ExpectedStatusCode int               `json:"expected_status_code"`
	Headers            map[string]string `json:"headers"`
	Body               string            `json:"body,omitempty"`
	SSLCheckEnabled    bool              `json:"ssl_check_enabled"`
	ScheduledAt        time.Time         `json:"scheduled_at"`
}
