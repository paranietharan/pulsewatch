package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"pulsewatch-scheduler/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCheckJobSerialization(t *testing.T) {
	jobID := uuid.New()
	monitorID := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Millisecond)

	job := models.CheckJob{
		JobID:              jobID,
		MonitorID:          monitorID,
		UserID:             userID,
		Name:               "Production Gateway",
		URL:                "https://api.example.com/health",
		Method:             "GET",
		TimeoutSeconds:     10,
		ExpectedStatusCode: 200,
		Headers: map[string]string{
			"User-Agent": "PulseWatch-Probe/1.0",
		},
		Body:            "",
		SSLCheckEnabled: true,
		ScheduledAt:     now,
	}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("failed to marshal CheckJob: %v", err)
	}

	var decoded models.CheckJob
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal CheckJob: %v", err)
	}

	if decoded.JobID != jobID || decoded.MonitorID != monitorID {
		t.Errorf("IDs do not match after serialization")
	}
	if decoded.URL != "https://api.example.com/health" {
		t.Errorf("URL mismatch: expected https://api.example.com/health, got %s", decoded.URL)
	}
	if decoded.Headers["User-Agent"] != "PulseWatch-Probe/1.0" {
		t.Errorf("Headers mismatch: %v", decoded.Headers)
	}
}

func setupSchedulerTestDB(t *testing.T) *pgxpool.Pool {
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

func TestClaimDueMonitors_Integration(t *testing.T) {
	pool := setupSchedulerTestDB(t)
	defer pool.Close()

	ctx := context.Background()

	userID := uuid.New()
	monitorID := uuid.New()

	// Insert test user & monitor due immediately
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, first_name, last_name, email, password_hash, role, is_active, is_verified)
		VALUES ($1, 'Sched', 'User', $2, 'hash', 'USER'::user_role, true, true)
		ON CONFLICT (id) DO NOTHING;
	`, userID, fmt.Sprintf("sched_%s@example.com", userID.String()[:8]))
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO monitors (
			id, user_id, name, url, method, interval_seconds, timeout_seconds,
			expected_status_code, next_check_at, is_active
		) VALUES (
			$1, $2, 'Due Monitor', 'https://example.com/health', 'GET'::http_method, 60, 10,
			200, NOW() - INTERVAL '5 seconds', true
		);
	`, monitorID, userID)
	if err != nil {
		t.Fatalf("failed to insert test monitor: %v", err)
	}

	defer func() {
		_, _ = pool.Exec(context.Background(), "DELETE FROM monitors WHERE id = $1;", monitorID)
		_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1;", userID)
	}()

	sched := &Scheduler{
		db:           pool,
		logger:       slog.Default(),
		pollInterval: 5 * time.Second,
		batchSize:    10,
	}

	// Claim due monitors -> should return our monitor
	jobs, err := sched.ClaimDueMonitors(ctx, 10)
	if err != nil {
		t.Fatalf("failed to claim due monitors: %v", err)
	}

	var found bool
	for _, j := range jobs {
		if j.MonitorID == monitorID {
			found = true
			if j.URL != "https://example.com/health" {
				t.Errorf("expected URL https://example.com/health, got %s", j.URL)
			}
			if j.Method != "GET" {
				t.Errorf("expected method GET, got %s", j.Method)
			}
			if j.TimeoutSeconds != 10 {
				t.Errorf("expected timeout 10, got %d", j.TimeoutSeconds)
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected monitor %s in claimed jobs, but was not found", monitorID)
	}

	// Immediately claim again -> next_check_at was advanced by 60s, so it must NOT be claimed again
	jobsSecond, err := sched.ClaimDueMonitors(ctx, 10)
	if err != nil {
		t.Fatalf("failed to claim second time: %v", err)
	}
	for _, j := range jobsSecond {
		if j.MonitorID == monitorID {
			t.Fatalf("monitor %s claimed twice! FOR UPDATE SKIP LOCKED / next_check_at advance failed", monitorID)
		}
	}

	// Verify in DB that next_check_at is now in the future
	var nextCheckAt time.Time
	err = pool.QueryRow(ctx, "SELECT next_check_at FROM monitors WHERE id = $1;", monitorID).Scan(&nextCheckAt)
	if err != nil {
		t.Fatalf("failed to query next_check_at: %v", err)
	}
	if !nextCheckAt.After(time.Now().UTC()) {
		t.Errorf("expected next_check_at to be in the future, got %v", nextCheckAt)
	}
}
