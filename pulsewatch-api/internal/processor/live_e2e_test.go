package processor_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"pulsewatch-api/internal/processor"
	"pulsewatch-api/internal/rabbitmq"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestLiveE2E_DistributedPipeline(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	rmqURL := os.Getenv("RABBITMQ_URL")
	if rmqURL == "" {
		rmqURL = "amqp://pulsewatch:pulsewatch_secret@localhost:5672/"
	}

	rmqClient, err := rabbitmq.Connect(rmqURL)
	if err != nil {
		t.Skipf("Skipping live E2E test: RabbitMQ is not accessible (%v)", err)
	}
	defer rmqClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Setup test user and monitor
	userID := uuid.New()
	monitorID := uuid.New()
	jobID := uuid.New()

	_, err = pool.Exec(ctx, `
		INSERT INTO users (id, first_name, last_name, email, password_hash, role, is_active, is_verified)
		VALUES ($1, 'E2E', 'User', $2, 'hash', 'USER'::user_role, true, true)
		ON CONFLICT (id) DO NOTHING;
	`, userID, fmt.Sprintf("e2e_%s@example.com", userID.String()[:8]))
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO monitors (
			id, user_id, name, url, method, interval_seconds, timeout_seconds,
			expected_status_code, failure_threshold, recovery_threshold,
			status, is_active, next_check_at
		) VALUES (
			$1, $2, 'Live E2E Monitor', 'https://example.com', 'GET'::http_method, 60, 10,
			200, 3, 2,
			'PENDING'::monitor_status, true, NOW()
		);
	`, monitorID, userID)
	if err != nil {
		t.Fatalf("failed to insert test monitor: %v", err)
	}

	defer func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM monitors WHERE id = $1;", monitorID)
		_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1;", userID)
	}()

	// Start Result Processor
	logger := slog.Default()
	proc := processor.NewResultProcessor(pool, rmqClient, logger)

	procCtx, procCancel := context.WithCancel(ctx)
	defer procCancel()

	go func() {
		_ = proc.Start(procCtx)
	}()

	time.Sleep(200 * time.Millisecond)

	// Publish mock check result to RabbitMQ queue monitor.results
	conn := rmqClient.Connection()
	pubCh, err := conn.Channel()
	if err != nil {
		t.Fatalf("failed to open publisher channel: %v", err)
	}
	defer pubCh.Close()

	checkResult := rabbitmq.CheckResult{
		JobID:          jobID,
		MonitorID:      monitorID,
		UserID:         userID,
		Status:         "UP",
		StatusCode:     200,
		ResponseTimeMs: 82,
		DNSLookupMs:    12,
		TCPConnectMs:   25,
		TLSHandshakeMs: 45,
		Region:         "us-east-1",
		CheckedAt:      time.Now().UTC(),
	}

	body, err := json.Marshal(checkResult)
	if err != nil {
		t.Fatalf("failed to serialize check result: %v", err)
	}

	err = pubCh.PublishWithContext(
		ctx,
		rabbitmq.ExchangeName,
		rabbitmq.RoutingKeyResult,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now().UTC(),
			Body:         body,
		},
	)
	if err != nil {
		t.Fatalf("failed to publish check result: %v", err)
	}

	// Poll until check result is recorded in database
	deadline := time.Now().Add(5 * time.Second)
	found := false
	for time.Now().Before(deadline) {
		var recordedCount int
		err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM heartbeats WHERE job_id = $1;", jobID).Scan(&recordedCount)
		if err == nil && recordedCount > 0 {
			found = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !found {
		t.Fatalf("check result with job_id %s was not processed into heartbeats within deadline", jobID)
	}

	// Verify monitor status updated to UP
	var monitorStatus string
	err = pool.QueryRow(ctx, "SELECT status FROM monitors WHERE id = $1;", monitorID).Scan(&monitorStatus)
	if err != nil {
		t.Fatalf("failed to query monitor status: %v", err)
	}
	if monitorStatus != "UP" {
		t.Errorf("expected monitor status 'UP', got '%s'", monitorStatus)
	}
}
