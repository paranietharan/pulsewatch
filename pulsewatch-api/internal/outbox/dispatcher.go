package outbox

import (
	"context"
	"log/slog"
	"time"

	"pulsewatch-api/internal/rabbitmq"
	"pulsewatch-api/internal/repository"
)

// Dispatcher periodically polls pending outbox events from PostgreSQL and reliably publishes them to RabbitMQ.
type Dispatcher struct {
	outboxRepo *repository.OutboxRepository
	rmqClient  *rabbitmq.Client
	logger     *slog.Logger
	interval   time.Duration
}

func NewDispatcher(
	outboxRepo *repository.OutboxRepository,
	rmqClient *rabbitmq.Client,
	logger *slog.Logger,
) *Dispatcher {
	return &Dispatcher{
		outboxRepo: outboxRepo,
		rmqClient:  rmqClient,
		logger:     logger,
		interval:   1 * time.Second,
	}
}

// Start runs the polling loop until the context is cancelled.
func (d *Dispatcher) Start(ctx context.Context) {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	d.logger.Info("Transactional Outbox Dispatcher started")

	for {
		select {
		case <-ctx.Done():
			d.logger.Info("Transactional Outbox Dispatcher stopping")
			return
		case <-ticker.C:
			d.processPendingEvents(ctx)
		}
	}
}

// processPendingEvents batches up to 50 pending events and dispatches them with publisher confirms.
func (d *Dispatcher) processPendingEvents(ctx context.Context) {
	events, err := d.outboxRepo.GetPending(ctx, 50)
	if err != nil {
		d.logger.Error("Failed to fetch pending outbox events", "error", err)
		return
	}

	for _, evt := range events {
		publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := d.rmqClient.PublishConfirm(publishCtx, evt.RoutingKey, evt.Payload)
		cancel()

		if err != nil {
			d.logger.Warn("Failed to publish outbox event to RabbitMQ",
				"event_id", evt.ID,
				"routing_key", evt.RoutingKey,
				"error", err,
			)
			_ = d.outboxRepo.MarkFailed(ctx, evt.ID, err.Error())
		} else {
			d.logger.Debug("Successfully published confirmed event to RabbitMQ",
				"event_id", evt.ID,
				"routing_key", evt.RoutingKey,
			)
			_ = d.outboxRepo.MarkPublished(ctx, evt.ID)
		}
	}
}
