package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pulsewatch-probe-worker/internal/config"
	"pulsewatch-probe-worker/internal/probe"
	"pulsewatch-probe-worker/internal/worker"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	logger.Info("Starting PulseWatch Probe Worker Service...")

	// Load Configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Connect to RabbitMQ
	rmqConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer rmqConn.Close()
	logger.Info("Connected to RabbitMQ successfully")

	// Initialize Prober & Worker Pool
	prober := probe.NewProber(cfg.Region)
	pool := worker.NewWorkerPool(rmqConn, prober, cfg.WorkerCount, cfg.PrefetchCount, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := pool.Start(ctx); err != nil {
		logger.Error("Failed to start probe worker pool", "error", err)
		os.Exit(1)
	}

	// Graceful Shutdown on SIGINT / SIGTERM
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	sig := <-shutdown
	logger.Info("Shutdown signal received, draining probe workers...", "signal", sig.String())

	cancel()

	// Wait for active probes to complete
	drainDone := make(chan struct{})
	go func() {
		pool.Stop()
		close(drainDone)
	}()

	select {
	case <-drainDone:
		logger.Info("All probe workers drained cleanly")
	case <-time.After(10 * time.Second):
		logger.Warn("Timed out waiting for probe workers to drain")
	}

	logger.Info("PulseWatch Probe Worker Service shut down gracefully")
}
