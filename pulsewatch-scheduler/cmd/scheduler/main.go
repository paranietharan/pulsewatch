package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pulsewatch-scheduler/internal/config"
	"pulsewatch-scheduler/internal/scheduler"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	logger.Info("Starting PulseWatch Scheduler Service...")

	// Load Configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Connect to PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pgConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to parse database URL", "error", err)
		os.Exit(1)
	}
	pgConfig.MaxConns = 15
	pgConfig.MinConns = 2
	pgConfig.MaxConnLifetime = 30 * time.Minute

	dbPool, err := pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(ctx); err != nil {
		logger.Error("Failed to ping PostgreSQL", "error", err)
		os.Exit(1)
	}
	logger.Info("Connected to PostgreSQL successfully")

	// Connect to RabbitMQ
	rmqConn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer rmqConn.Close()
	logger.Info("Connected to RabbitMQ successfully")

	// Initialize Scheduler
	sched, err := scheduler.New(dbPool, rmqConn, cfg.PollInterval, cfg.BatchSize, logger)
	if err != nil {
		logger.Error("Failed to initialize scheduler", "error", err)
		os.Exit(1)
	}
	defer sched.Close()

	// Start Scheduler Loop
	schedCtx, schedCancel := context.WithCancel(context.Background())
	defer schedCancel()

	go sched.Start(schedCtx)

	// Graceful Shutdown on SIGINT / SIGTERM
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	sig := <-shutdown
	logger.Info("Shutdown signal received, shutting down scheduler...", "signal", sig.String())

	schedCancel()

	// Allow current batch to complete within 5-second deadline
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	<-shutdownCtx.Done()
	logger.Info("PulseWatch Scheduler Service shut down gracefully")
}
