package processor_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"pulsewatch-api/internal/processor"
	"pulsewatch-api/internal/rabbitmq"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://pulsewatch:pulsewatch_secret@localhost:5432/pulsewatch?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("Skipping integration test: PostgreSQL is not accessible (%v)", err)
	}

	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Skipping integration test: PostgreSQL ping failed (%v)", err)
	}

	return pool
}

func TestResultProcessor_IncidentLifecycleAndThresholds(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()

	userID := uuid.New()
	monitorID := uuid.New()

	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, first_name, last_name, email, password_hash, role, is_active, is_verified)
		VALUES ($1, 'Proc', 'Test', $2, 'hash', 'USER'::user_role, true, true)
		ON CONFLICT (id) DO NOTHING;
	`, userID, fmt.Sprintf("proc_%s@example.com", userID.String()[:8]))
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO monitors (
			id, user_id, name, url, method, interval_seconds, timeout_seconds,
			expected_status_code, failure_threshold, recovery_threshold,
			consecutive_failures, consecutive_successes, status, is_active
		) VALUES (
			$1, $2, 'Incident Test Site', 'https://example.com', 'GET'::http_method, 60, 10,
			200, 2, 2,
			0, 0, 'PENDING'::monitor_status, true
		);
	`, monitorID, userID)
	if err != nil {
		t.Fatalf("failed to insert test monitor: %v", err)
	}

	logger := slog.Default()
	proc := processor.NewResultProcessor(pool, nil, logger)

	defer func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM monitors WHERE id = $1;", monitorID)
		_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1;", userID)
	}()

	// First Failure (Threshold = 2, so should NOT open incident yet)
	job1 := uuid.New()
	res1 := rabbitmq.CheckResult{
		JobID:          job1,
		MonitorID:      monitorID,
		Status:         "DOWN",
		StatusCode:     500,
		ResponseTimeMs: 150,
		ErrorMessage:   "HTTP 500 Internal Server Error",
		Region:         "us-east",
		CheckedAt:      time.Now().UTC(),
	}

	if err := proc.ProcessResult(ctx, res1); err != nil {
		t.Fatalf("failed to process res1: %v", err)
	}

	var status string
	var consFailures int
	err = pool.QueryRow(ctx, "SELECT status, consecutive_failures FROM monitors WHERE id = $1;", monitorID).Scan(&status, &consFailures)
	if err != nil {
		t.Fatalf("failed to query monitor after res1: %v", err)
	}
	if consFailures != 1 {
		t.Errorf("expected consecutive_failures = 1, got %d", consFailures)
	}

	var incidentCount int
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM incidents WHERE monitor_id = $1;", monitorID).Scan(&incidentCount)
	if incidentCount != 0 {
		t.Errorf("expected 0 incidents after 1 failure (threshold=2), got %d", incidentCount)
	}

	// Duplicate res1 delivery (idempotency check)
	// Processing the same JobID must NOT double-increment consecutive failures
	if err := proc.ProcessResult(ctx, res1); err != nil {
		t.Fatalf("failed to process duplicate res1: %v", err)
	}
	err = pool.QueryRow(ctx, "SELECT consecutive_failures FROM monitors WHERE id = $1;", monitorID).Scan(&consFailures)
	if err != nil {
		t.Fatalf("failed to query monitor after duplicate res1: %v", err)
	}
	if consFailures != 1 {
		t.Errorf("duplicate job_id should be deduplicated: expected 1, got %d", consFailures)
	}

	// Second Failure (Threshold = 2 reached -> Incident OPENED)
	job2 := uuid.New()
	res2 := rabbitmq.CheckResult{
		JobID:          job2,
		MonitorID:      monitorID,
		Status:         "DOWN",
		StatusCode:     502,
		ResponseTimeMs: 200,
		ErrorMessage:   "HTTP 502 Bad Gateway",
		Region:         "us-east",
		CheckedAt:      time.Now().UTC(),
	}

	if err := proc.ProcessResult(ctx, res2); err != nil {
		t.Fatalf("failed to process res2: %v", err)
	}

	err = pool.QueryRow(ctx, "SELECT status, consecutive_failures FROM monitors WHERE id = $1;", monitorID).Scan(&status, &consFailures)
	if err != nil {
		t.Fatalf("failed to query monitor after res2: %v", err)
	}
	if status != "DOWN" {
		t.Errorf("expected monitor status DOWN, got %s", status)
	}
	if consFailures != 2 {
		t.Errorf("expected consecutive_failures = 2, got %d", consFailures)
	}

	var openIncidents int
	var incidentID uuid.UUID
	err = pool.QueryRow(ctx, "SELECT id, COUNT(*) OVER() FROM incidents WHERE monitor_id = $1 AND status = 'OPEN';", monitorID).Scan(&incidentID, &openIncidents)
	if err != nil {
		t.Fatalf("failed to query open incidents: %v", err)
	}
	if openIncidents != 1 {
		t.Fatalf("expected 1 open incident, got %d", openIncidents)
	}

	// First Recovery Success (Recovery Threshold = 2, so should NOT resolve yet)
	job3 := uuid.New()
	res3 := rabbitmq.CheckResult{
		JobID:          job3,
		MonitorID:      monitorID,
		Status:         "UP",
		StatusCode:     200,
		ResponseTimeMs: 50,
		Region:         "us-east",
		CheckedAt:      time.Now().UTC(),
	}

	if err := proc.ProcessResult(ctx, res3); err != nil {
		t.Fatalf("failed to process res3: %v", err)
	}

	var consSuccesses int
	err = pool.QueryRow(ctx, "SELECT consecutive_successes, status FROM monitors WHERE id = $1;", monitorID).Scan(&consSuccesses, &status)
	if err != nil {
		t.Fatalf("failed to query monitor after res3: %v", err)
	}
	if consSuccesses != 1 {
		t.Errorf("expected consecutive_successes = 1, got %d", consSuccesses)
	}
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM incidents WHERE id = $1 AND status = 'OPEN';", incidentID).Scan(&openIncidents)
	if openIncidents != 1 {
		t.Errorf("incident should remain OPEN after 1 success (recovery threshold=2)")
	}

	// Second Recovery Success (Recovery Threshold = 2 reached -> Incident RESOLVED)
	job4 := uuid.New()
	res4 := rabbitmq.CheckResult{
		JobID:          job4,
		MonitorID:      monitorID,
		Status:         "UP",
		StatusCode:     200,
		ResponseTimeMs: 45,
		Region:         "us-east",
		CheckedAt:      time.Now().UTC(),
	}

	if err := proc.ProcessResult(ctx, res4); err != nil {
		t.Fatalf("failed to process res4: %v", err)
	}

	err = pool.QueryRow(ctx, "SELECT status, consecutive_successes FROM monitors WHERE id = $1;", monitorID).Scan(&status, &consSuccesses)
	if err != nil {
		t.Fatalf("failed to query monitor after res4: %v", err)
	}
	if status != "UP" {
		t.Errorf("expected monitor status UP after recovery, got %s", status)
	}

	var incStatus string
	var resolvedAt *time.Time
	err = pool.QueryRow(ctx, "SELECT status, resolved_at FROM incidents WHERE id = $1;", incidentID).Scan(&incStatus, &resolvedAt)
	if err != nil {
		t.Fatalf("failed to query incident after res4: %v", err)
	}
	if incStatus != "RESOLVED" {
		t.Errorf("expected incident status RESOLVED, got %s", incStatus)
	}
	if resolvedAt == nil {
		t.Error("expected resolved_at timestamp to be non-nil")
	}
}
