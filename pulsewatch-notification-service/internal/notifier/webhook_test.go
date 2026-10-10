package notifier

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestComputeHMACSHA256(t *testing.T) {
	secret := "super-secret-key-1234"
	payload := []byte(`{"event":"test"}`)

	sig := computeHMACSHA256(payload, secret)
	if sig == "" {
		t.Fatal("expected non-empty signature")
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	expected := hex.EncodeToString(h.Sum(nil))

	if sig != expected {
		t.Fatalf("expected signature %s, got %s", expected, sig)
	}
}

func TestValidateWebhookURL(t *testing.T) {
	validURLs := []string{
		"https://example.com/webhook",
		"http://example.com:8080/hook",
		"http://localhost:9000/webhook",
	}
	for _, u := range validURLs {
		if err := validateWebhookURL(u); err != nil {
			t.Errorf("expected URL %s to be valid, got: %v", u, err)
		}
	}

	invalidURLs := []string{
		"ftp://example.com/webhook",
		"file:///etc/passwd",
		"javascript:alert(1)",
		"",
	}
	for _, u := range invalidURLs {
		if err := validateWebhookURL(u); err == nil {
			t.Errorf("expected URL %s to be invalid, got nil", u)
		}
	}
}

func TestWebhookDelivery_Success(t *testing.T) {
	receivedHeader := ""
	receivedSignature := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-PulseWatch-Event")
		receivedSignature = r.Header.Get("X-PulseWatch-Signature")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"received":true}`))
	}))
	defer server.Close()

	wn := &WebhookNotifier{
		client: server.Client(),
	}

	payload := WebhookPayload{
		EventID:     uuid.New(),
		EventType:   "incident.opened",
		IncidentID:  uuid.New(),
		MonitorID:   uuid.New(),
		MonitorName: "Production App",
		TargetURL:   server.URL,
		Status:      "OPEN",
		Timestamp:   time.Now().UTC(),
	}

	secret := "secret-webhook-key"
	deliveryID := uuid.New()

	status, attempts, err := wn.Deliver(context.Background(), deliveryID, server.URL, secret, payload)
	if err != nil {
		t.Fatalf("webhook deliver failed: %v", err)
	}

	if status != "DELIVERED" {
		t.Errorf("expected status DELIVERED, got %s", status)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
	if receivedHeader != "incident.opened" {
		t.Errorf("expected header incident.opened, got %s", receivedHeader)
	}
	if receivedSignature == "" {
		t.Error("expected non-empty signature header")
	}
}
