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
	for i := 0; i < 50; i++ {
		found, err := d.outboxRepo.DispatchOne(ctx, d.rmqClient.PublishConfirm)
		if err != nil {
			d.logger.Error("Outbox dispatch deferred", "error", err)
			return
		}
		if !found {
			return
		}
	}
}
