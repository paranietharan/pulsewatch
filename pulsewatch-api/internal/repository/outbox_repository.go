package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"pulsewatch-api/internal/models"
)

// OutboxRepository manages database operations for transactional outbox events.
type OutboxRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{pool: pool}
}

// Create inserts a new outbox event within a transaction.
func (r *OutboxRepository) Create(ctx context.Context, tx pgx.Tx, event *models.OutboxEvent) error {
	query := `
		INSERT INTO auth_outbox_events (event_type, routing_key, payload, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at;
	`
	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, query, event.EventType, event.RoutingKey, event.Payload, models.OutboxStatusPending)
	} else {
		row = r.pool.QueryRow(ctx, query, event.EventType, event.RoutingKey, event.Payload, models.OutboxStatusPending)
	}
	if err := row.Scan(&event.ID, &event.CreatedAt); err != nil {
		return fmt.Errorf("failed to create outbox event: %w", err)
	}
	event.Status = models.OutboxStatusPending
	return nil
}

// GetPending retrieves pending events with FOR UPDATE SKIP LOCKED for concurrent safe workers.
func (r *OutboxRepository) GetPending(ctx context.Context, limit int) ([]models.OutboxEvent, error) {
	query := `
		SELECT id, event_type, routing_key, payload, status, retry_count, error_message, created_at, published_at
		FROM auth_outbox_events
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED;
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending outbox events: %w", err)
	}
	defer rows.Close()

	var events []models.OutboxEvent
	for rows.Next() {
		var evt models.OutboxEvent
		if err := rows.Scan(
			&evt.ID,
			&evt.EventType,
			&evt.RoutingKey,
			&evt.Payload,
			&evt.Status,
			&evt.RetryCount,
			&evt.ErrorMessage,
			&evt.CreatedAt,
			&evt.PublishedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan outbox event: %w", err)
		}
		events = append(events, evt)
	}

	return events, rows.Err()
}

// MarkPublished marks an outbox event as successfully dispatched.
func (r *OutboxRepository) MarkPublished(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE auth_outbox_events
		SET status = 'PUBLISHED', published_at = $1
		WHERE id = $2;
	`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to mark outbox event published: %w", err)
	}
	return nil
}

// MarkFailed increments retry_count and updates status if max retries exceeded.
func (r *OutboxRepository) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error {
	query := `
		UPDATE auth_outbox_events
		SET retry_count = retry_count + 1,
		    error_message = $1,
		    status = CASE WHEN retry_count + 1 >= 5 THEN 'FAILED' ELSE 'PENDING' END
		WHERE id = $2;
	`
	_, err := r.pool.Exec(ctx, query, errMsg, id)
	if err != nil {
		return fmt.Errorf("failed to mark outbox event failed: %w", err)
	}
	return nil
}
