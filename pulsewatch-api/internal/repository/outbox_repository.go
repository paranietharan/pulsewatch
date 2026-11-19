package repository

import (
	"context"
	"errors"
	"fmt"

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

// DispatchOne holds the claim until the matching confirmation is persisted.
// If the connection dies after broker acceptance, redelivery uses the same ID.
func (r *OutboxRepository) DispatchOne(ctx context.Context, publish func(context.Context, string, []byte, string) error) (bool, error) {
	var found bool
	var publishErr error
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		var event models.OutboxEvent
		err := tx.QueryRow(ctx, `SELECT id,routing_key,payload FROM auth_outbox_events
            WHERE status='PENDING' AND ready_at<=now() ORDER BY created_at,id
            LIMIT 1 FOR UPDATE SKIP LOCKED`).Scan(&event.ID, &event.RoutingKey, &event.Payload)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		found = true
		publishErr = publish(ctx, event.RoutingKey, event.Payload, event.ID.String())
		if publishErr != nil {
			_, err = tx.Exec(ctx, `UPDATE auth_outbox_events SET retry_count=retry_count+1,
                ready_at=now()+LEAST(300,POWER(2,LEAST(retry_count,9))) * interval '1 second',
                error_message=$2 WHERE id=$1`, event.ID, publishErr.Error())
		} else {
			_, err = tx.Exec(ctx, `UPDATE auth_outbox_events SET status='PUBLISHED',published_at=now(),
              error_message=NULL WHERE id=$1`, event.ID)
		}
		return err
	})
	return found, errors.Join(err, publishErr)
}
