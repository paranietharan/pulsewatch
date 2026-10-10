package notifier

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

type WebhookNotifier struct {
	client *http.Client
}

type WebhookPayload struct {
	EventID         uuid.UUID  `json:"event_id"`
	EventType       string     `json:"event_type"` // incident.opened, incident.resolved
	IncidentID      uuid.UUID  `json:"incident_id"`
	MonitorID       uuid.UUID  `json:"monitor_id"`
	MonitorName     string     `json:"monitor_name"`
	TargetURL       string     `json:"target_url"`
	Status          string     `json:"status"` // OPEN, RESOLVED
	CauseSummary    string     `json:"cause_summary,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty"`
	DurationSeconds *int       `json:"duration_seconds,omitempty"`
	Timestamp       time.Time  `json:"timestamp"`
}

func NewWebhookNotifier() *WebhookNotifier {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
		MaxIdleConns:        50,
		IdleConnTimeout:     60 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return validateWebhookURL(req.URL.String())
		},
	}

	return &WebhookNotifier{
		client: client,
	}
}

// Deliver sends a signed webhook POST request with retry logic and returns the delivery outcome status.
func (wn *WebhookNotifier) Deliver(ctx context.Context, deliveryID uuid.UUID, targetURL, secret string, payload WebhookPayload) (status string, attempts int, lastErr error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "FAILED", 0, fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	if err := validateWebhookURL(targetURL); err != nil {
		return "FAILED", 1, fmt.Errorf("invalid webhook target URL: %w", err)
	}

	signature := computeHMACSHA256(payloadBytes, secret)

	maxAttempts := 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		attempts = attempt
		req, err := http.NewRequestWithContext(ctx, "POST", targetURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			lastErr = err
			break
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "PulseWatch-Webhook/1.0")
		req.Header.Set("X-PulseWatch-Event", payload.EventType)
		req.Header.Set("X-PulseWatch-Delivery", deliveryID.String())
		if signature != "" {
			req.Header.Set("X-PulseWatch-Signature", "sha256="+signature)
		}

		resp, err := wn.client.Do(req)
		if err == nil {
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
			resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return "DELIVERED", attempt, nil
			}
			lastErr = fmt.Errorf("endpoint returned HTTP %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		// Wait before retry with backoff
		if attempt < maxAttempts {
			select {
			case <-ctx.Done():
				return "FAILED", attempt, ctx.Err()
			case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
			}
		}
	}

	return "FAILED", attempts, lastErr
}

func validateWebhookURL(raw string) error {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("malformed URL: %w", err)
	}

	// Must be HTTP or HTTPS
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("scheme must be http or https, got %s", scheme)
	}

	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	// Basic check for loopback/localhost in development
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return nil
	}

	return nil
}

func computeHMACSHA256(payload []byte, secret string) string {
	if secret == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
