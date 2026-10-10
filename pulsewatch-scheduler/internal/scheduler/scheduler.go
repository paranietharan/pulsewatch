package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"pulsewatch-scheduler/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeName    = "pulsewatch.direct"
	ChecksQueue     = "monitor.checks"
	RoutingKeyCheck = "monitor.check"
	DLXExchangeName = "pulsewatch.dlx"
	ChecksDLQ       = "monitor.checks.dlq"
)

type Scheduler struct {
	db           *pgxpool.Pool
	rmqConn      *amqp.Connection
	rmqCh        *amqp.Channel
	logger       *slog.Logger
	pollInterval time.Duration
	batchSize    int
	mu           sync.Mutex
}

func New(db *pgxpool.Pool, rmqConn *amqp.Connection, pollInterval time.Duration, batchSize int, logger *slog.Logger) (*Scheduler, error) {
	ch, err := rmqConn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}

	// Enable Publisher Confirms
	if err := ch.Confirm(false); err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	// Declare Dead Letter Exchange & Queue
	if err := ch.ExchangeDeclare(DLXExchangeName, "direct", true, false, false, false, nil); err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to declare DLX: %w", err)
	}
	if _, err := ch.QueueDeclare(ChecksDLQ, true, false, false, false, nil); err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to declare Checks DLQ: %w", err)
	}
	if err := ch.QueueBind(ChecksDLQ, RoutingKeyCheck, DLXExchangeName, false, nil); err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to bind Checks DLQ: %w", err)
	}

	// Declare Primary Exchange
	if err := ch.ExchangeDeclare(ExchangeName, "direct", true, false, false, false, nil); err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to declare primary exchange: %w", err)
	}

	// Declare Primary Checks Queue
	if _, err := ch.QueueDeclare(ChecksQueue, true, false, false, false, nil); err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to declare Checks queue: %w", err)
	}
	if err := ch.QueueBind(ChecksQueue, RoutingKeyCheck, ExchangeName, false, nil); err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to bind Checks queue: %w", err)
	}

	return &Scheduler{
		db:           db,
		rmqConn:      rmqConn,
		rmqCh:        ch,
		logger:       logger,
		pollInterval: pollInterval,
		batchSize:    batchSize,
	}, nil
}

// Start runs the periodic scheduling loop until the context is canceled.
func (s *Scheduler) Start(ctx context.Context) {
	s.logger.Info("Scheduler loop started",
		"poll_interval", s.pollInterval.String(),
		"batch_size", s.batchSize,
	)

	// Execute an immediate sweep on startup
	s.sweepAndDispatch(ctx)

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Scheduler loop received cancellation signal")
			return
		case <-ticker.C:
			s.sweepAndDispatch(ctx)
		}
	}
}

// Close gracefully closes the RabbitMQ channel.
func (s *Scheduler) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rmqCh != nil {
		_ = s.rmqCh.Close()
	}
}

// sweepAndDispatch atomically claims due monitors and dispatches check jobs to RabbitMQ.
func (s *Scheduler) sweepAndDispatch(ctx context.Context) {
	jobs, err := s.ClaimDueMonitors(ctx, s.batchSize)
	if err != nil {
		s.logger.Error("Failed to claim due monitors from database", "error", err)
		return
	}

	if len(jobs) == 0 {
		return
	}

	s.logger.Info("Claimed due monitors for scheduling", "count", len(jobs))

	publishedCount := 0
	for _, job := range jobs {
		if err := s.publishJob(ctx, job); err != nil {
			s.logger.Error("Failed to publish check job to RabbitMQ",
				"monitor_id", job.MonitorID,
				"job_id", job.JobID,
				"error", err,
			)
			// Monitor's next_check_at was already safely updated in DB; subsequent tick will retry on failure
			continue
		}
		publishedCount++
	}

	s.logger.Debug("Dispatched check jobs batch",
		"total_claimed", len(jobs),
		"published", publishedCount,
	)
}

// ClaimDueMonitors uses FOR UPDATE SKIP LOCKED to atomically advance next_check_at and claim monitors.
func (s *Scheduler) ClaimDueMonitors(ctx context.Context, limit int) ([]models.CheckJob, error) {
	query := `
		WITH due_monitors AS (
			SELECT id
			FROM monitors
			WHERE is_active = TRUE
			  AND next_check_at <= NOW()
			ORDER BY next_check_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE monitors m
		SET next_check_at = NOW() + (m.interval_seconds * INTERVAL '1 second'),
		    updated_at = NOW()
		FROM due_monitors d
		WHERE m.id = d.id
		RETURNING m.id, m.user_id, m.name, m.url, m.method, m.interval_seconds,
		          m.timeout_seconds, m.expected_status_code, m.headers, m.body,
		          m.ssl_check_enabled;
	`

	rows, err := s.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query due monitors: %w", err)
	}
	defer rows.Close()

	var jobs []models.CheckJob
	now := time.Now().UTC()

	for rows.Next() {
		var job models.CheckJob
		var headersRaw []byte
		var body *string

		var intervalSeconds int
		err := rows.Scan(
			&job.MonitorID,
			&job.UserID,
			&job.Name,
			&job.URL,
			&job.Method,
			&intervalSeconds,
			&job.TimeoutSeconds,
			&job.ExpectedStatusCode,
			&headersRaw,
			&body,
			&job.SSLCheckEnabled,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monitor row: %w", err)
		}

		job.JobID = uuid.New()
		job.ScheduledAt = now

		if len(headersRaw) > 0 {
			_ = json.Unmarshal(headersRaw, &job.Headers)
		}
		if job.Headers == nil {
			job.Headers = make(map[string]string)
		}
		if body != nil {
			job.Body = *body
		}

		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

// publishJob publishes a CheckJob with publisher confirms.
func (s *Scheduler) publishJob(ctx context.Context, job models.CheckJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job payload: %w", err)
	}

	pubCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	confirmation, err := s.rmqCh.PublishWithDeferredConfirmWithContext(
		pubCtx,
		ExchangeName,
		RoutingKeyCheck,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    job.JobID.String(),
			Timestamp:    job.ScheduledAt,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish check job: %w", err)
	}

	// Block for confirmation
	ack, err := confirmation.WaitContext(pubCtx)
	if err != nil {
		return fmt.Errorf("timeout waiting for publisher confirm: %w", err)
	}
	if !ack {
		return fmt.Errorf("received NACK for check job %s", job.JobID)
	}

	return nil
}
