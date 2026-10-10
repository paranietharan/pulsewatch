package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"pulsewatch-probe-worker/internal/probe"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeName     = "pulsewatch.direct"
	ChecksQueue      = "monitor.checks"
	RoutingKeyCheck  = "monitor.check"
	ResultsQueue     = "monitor.results"
	RoutingKeyResult = "monitor.result"
)

type WorkerPool struct {
	rmqConn     *amqp.Connection
	prober      *probe.Prober
	logger      *slog.Logger
	workerCount int
	prefetch    int
	wg          sync.WaitGroup
}

func NewWorkerPool(rmqConn *amqp.Connection, prober *probe.Prober, workerCount, prefetch int, logger *slog.Logger) *WorkerPool {
	return &WorkerPool{
		rmqConn:     rmqConn,
		prober:      prober,
		logger:      logger,
		workerCount: workerCount,
		prefetch:    prefetch,
	}
}

// Start launches the RabbitMQ consumer and worker goroutines.
func (wp *WorkerPool) Start(ctx context.Context) error {
	ch, err := wp.rmqConn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open consumer channel: %w", err)
	}

	// Declare Topology
	if err := ch.ExchangeDeclare(ExchangeName, "direct", true, false, false, false, nil); err != nil {
		ch.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare Results queue
	if _, err := ch.QueueDeclare(ResultsQueue, true, false, false, false, nil); err != nil {
		ch.Close()
		return fmt.Errorf("failed to declare results queue: %w", err)
	}
	if err := ch.QueueBind(ResultsQueue, RoutingKeyResult, ExchangeName, false, nil); err != nil {
		ch.Close()
		return fmt.Errorf("failed to bind results queue: %w", err)
	}

	// Declare Checks queue
	if _, err := ch.QueueDeclare(ChecksQueue, true, false, false, false, nil); err != nil {
		ch.Close()
		return fmt.Errorf("failed to declare checks queue: %w", err)
	}
	if err := ch.QueueBind(ChecksQueue, RoutingKeyCheck, ExchangeName, false, nil); err != nil {
		ch.Close()
		return fmt.Errorf("failed to bind checks queue: %w", err)
	}

	// Set Prefetch QoS
	if err := ch.Qos(wp.prefetch, 0, false); err != nil {
		ch.Close()
		return fmt.Errorf("failed to set prefetch QoS: %w", err)
	}

	// Start consuming checks
	msgs, err := ch.Consume(
		ChecksQueue,
		"pulsewatch-probe-worker",
		false, // manual ack
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to start consuming from checks queue: %w", err)
	}

	wp.logger.Info("Probe worker consumer started",
		"concurrency", wp.workerCount,
		"prefetch", wp.prefetch,
		"queue", ChecksQueue,
	)

	// Create dedicated publisher channel with confirms for results dispatch
	pubCh, err := wp.rmqConn.Channel()
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to open publisher channel: %w", err)
	}
	if err := pubCh.Confirm(false); err != nil {
		ch.Close()
		pubCh.Close()
		return fmt.Errorf("failed to enable publisher confirms: %w", err)
	}

	var pubMu sync.Mutex

	// Launch worker pool
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go func(workerID int) {
			defer wp.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case d, ok := <-msgs:
					if !ok {
						return
					}
					wp.processDelivery(ctx, d, pubCh, &pubMu)
				}
			}
		}(i)
	}

	// Wait for context cancellation to close channels
	go func() {
		<-ctx.Done()
		_ = ch.Close()
		_ = pubCh.Close()
	}()

	return nil
}

// Stop waits for all worker goroutines to complete their current checks.
func (wp *WorkerPool) Stop() {
	wp.wg.Wait()
}

func (wp *WorkerPool) processDelivery(ctx context.Context, d amqp.Delivery, pubCh *amqp.Channel, pubMu *sync.Mutex) {
	var job probe.CheckJob
	if err := json.Unmarshal(d.Body, &job); err != nil {
		wp.logger.Error("Malformed check job body, dead-lettering message", "error", err)
		_ = d.Nack(false, false)
		return
	}

	start := time.Now()
	result := wp.prober.Execute(ctx, job)
	duration := time.Since(start)

	wp.logger.Debug("Probe executed",
		"job_id", job.JobID,
		"monitor_id", job.MonitorID,
		"status", result.Status,
		"status_code", result.StatusCode,
		"latency_ms", result.ResponseTimeMs,
		"duration_ms", duration.Milliseconds(),
	)

	// Publish result to RabbitMQ
	resultData, err := json.Marshal(result)
	if err != nil {
		wp.logger.Error("Failed to serialize check result", "error", err)
		_ = d.Nack(false, false)
		return
	}

	pubMu.Lock()
	pubCtx, pubCancel := context.WithTimeout(ctx, 5*time.Second)
	confirmation, err := pubCh.PublishWithDeferredConfirmWithContext(
		pubCtx,
		ExchangeName,
		RoutingKeyResult,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    result.JobID.String(),
			Timestamp:    result.CheckedAt,
			Body:         resultData,
		},
	)
	if err != nil {
		pubMu.Unlock()
		pubCancel()
		wp.logger.Error("Failed to publish check result, requeuing job", "error", err)
		_ = d.Nack(false, true)
		return
	}

	ack, err := confirmation.WaitContext(pubCtx)
	pubMu.Unlock()
	pubCancel()

	if err != nil || !ack {
		wp.logger.Error("Failed to receive publisher confirmation for check result, requeuing job", "error", err)
		_ = d.Nack(false, true)
		return
	}

	// Acknowledge original job only after result is safely confirmed in RabbitMQ
	if err := d.Ack(false); err != nil {
		wp.logger.Error("Failed to ack check job delivery", "error", err)
	}
}
