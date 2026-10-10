package probe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestProber_SSRFBlocking(t *testing.T) {
	_ = os.Unsetenv("ALLOW_INTERNAL_PROBES")
	prober := NewProber("test-region")

	job := CheckJob{
		JobID:          uuid.New(),
		MonitorID:      uuid.New(),
		UserID:         uuid.New(),
		Name:           "Internal Admin Probe",
		URL:            "http://127.0.0.1:8080/admin",
		Method:         "GET",
		TimeoutSeconds: 5,
		ScheduledAt:    time.Now().UTC(),
	}

	result := prober.Execute(context.Background(), job)

	if result.Status != "DOWN" {
		t.Errorf("expected status DOWN for SSRF blocked target, got %s", result.Status)
	}
	if !strings.Contains(result.ErrorMessage, "SSRF") {
		t.Errorf("expected error message to mention SSRF, got %s", result.ErrorMessage)
	}
}

func TestProber_InvalidURL(t *testing.T) {
	prober := NewProber("test-region")

	job := CheckJob{
		JobID:          uuid.New(),
		MonitorID:      uuid.New(),
		UserID:         uuid.New(),
		Name:           "Malformed Probe",
		URL:            "file:///etc/shadow",
		Method:         "GET",
		TimeoutSeconds: 5,
		ScheduledAt:    time.Now().UTC(),
	}

	result := prober.Execute(context.Background(), job)

	if result.Status != "DOWN" {
		t.Errorf("expected status DOWN for invalid scheme, got %s", result.Status)
	}
}

func TestProber_SuccessfulExecution(t *testing.T) {
	_ = os.Setenv("ALLOW_INTERNAL_PROBES", "true")
	defer os.Unsetenv("ALLOW_INTERNAL_PROBES")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	}))
	defer server.Close()

	prober := NewProber("eu-central")
	job := CheckJob{
		JobID:              uuid.New(),
		MonitorID:          uuid.New(),
		UserID:             uuid.New(),
		Name:               "Healthy Endpoint",
		URL:                server.URL,
		Method:             "GET",
		TimeoutSeconds:     5,
		ExpectedStatusCode: 200,
		ScheduledAt:        time.Now().UTC(),
	}

	result := prober.Execute(context.Background(), job)

	if result.Status != "UP" {
		t.Fatalf("expected status UP, got %s (err: %s)", result.Status, result.ErrorMessage)
	}
	if result.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", result.StatusCode)
	}
	if result.Region != "eu-central" {
		t.Errorf("expected region eu-central, got %s", result.Region)
	}
	if result.JobID != job.JobID {
		t.Errorf("expected JobID %s, got %s", job.JobID, result.JobID)
	}
}

func TestProber_TimeoutAndFailure(t *testing.T) {
	_ = os.Setenv("ALLOW_INTERNAL_PROBES", "true")
	defer os.Unsetenv("ALLOW_INTERNAL_PROBES")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	prober := NewProber("us-east")
	job := CheckJob{
		JobID:              uuid.New(),
		MonitorID:          uuid.New(),
		UserID:             uuid.New(),
		Name:               "Slow Endpoint",
		URL:                server.URL,
		Method:             "GET",
		TimeoutSeconds:     1, // 1s timeout
		ExpectedStatusCode: 200,
		ScheduledAt:        time.Now().UTC(),
	}

	result := prober.Execute(context.Background(), job)

	if result.Status != "DOWN" {
		t.Errorf("expected status DOWN due to timeout, got %s", result.Status)
	}
	if result.ErrorMessage == "" {
		t.Error("expected non-empty error message on timeout")
	}
}
