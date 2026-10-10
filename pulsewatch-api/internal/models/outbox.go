package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "PENDING"
	OutboxStatusPublished OutboxStatus = "PUBLISHED"
	OutboxStatusFailed    OutboxStatus = "FAILED"
)

// OutboxEvent represents an asynchronous event stored transactionally in PostgreSQL.
type OutboxEvent struct {
	ID           uuid.UUID       `json:"id"`
	EventType    string          `json:"event_type"`
	RoutingKey   string          `json:"routing_key"`
	Payload      json.RawMessage `json:"payload"`
	Status       OutboxStatus    `json:"status"`
	RetryCount   int             `json:"retry_count"`
	ErrorMessage *string         `json:"error_message,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	PublishedAt  *time.Time      `json:"published_at,omitempty"`
}
