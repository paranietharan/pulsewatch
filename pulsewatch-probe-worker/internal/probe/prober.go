package probe

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	"pulsewatch-probe-worker/internal/ssrf"

	"github.com/google/uuid"
)

type CheckJob struct {
	JobID              uuid.UUID         `json:"job_id"`
	MonitorID          uuid.UUID         `json:"monitor_id"`
	UserID             uuid.UUID         `json:"user_id"`
	Name               string            `json:"name"`
	URL                string            `json:"url"`
	Method             string            `json:"method"`
	TimeoutSeconds     int               `json:"timeout_seconds"`
	ExpectedStatusCode int               `json:"expected_status_code"`
	Headers            map[string]string `json:"headers"`
	Body               string            `json:"body,omitempty"`
	SSLCheckEnabled    bool              `json:"ssl_check_enabled"`
	ScheduledAt        time.Time         `json:"scheduled_at"`
}

type CheckResult struct {
	JobID          uuid.UUID `json:"job_id"`
	MonitorID      uuid.UUID `json:"monitor_id"`
	UserID         uuid.UUID `json:"user_id"`
	Status         string    `json:"status"` // UP, DOWN, DEGRADED
	StatusCode     int       `json:"status_code"`
	ResponseTimeMs int       `json:"response_time_ms"`
	DNSLookupMs    int       `json:"dns_lookup_ms"`
	TCPConnectMs   int       `json:"tcp_connect_ms"`
	TLSHandshakeMs int       `json:"tls_handshake_ms"`
	ErrorMessage   string    `json:"error_message,omitempty"`
	Region         string    `json:"region"`
	CheckedAt      time.Time `json:"checked_at"`
}

type Prober struct {
	region string
	client *http.Client
}

func NewProber(region string) *Prober {
	transport := &http.Transport{
		DialContext:           ssrf.SafeDialContext(),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	client := &http.Client{
		Transport:     transport,
		CheckRedirect: ssrf.SafeCheckRedirect,
	}

	return &Prober{
		region: region,
		client: client,
	}
}

// Execute performs an HTTP/HTTPS probe with bounded timeout, SSRF protection, and latency telemetry.
func (p *Prober) Execute(ctx context.Context, job CheckJob) CheckResult {
	checkedAt := time.Now().UTC()
	result := CheckResult{
		JobID:     job.JobID,
		MonitorID: job.MonitorID,
		UserID:    job.UserID,
		Region:    p.region,
		CheckedAt: checkedAt,
	}

	// SSRF Validation
	parsedURL, err := ssrf.ValidateTargetURL(job.URL)
	if err != nil {
		result.Status = "DOWN"
		result.ErrorMessage = fmt.Sprintf("SSRF blocked or invalid URL: %v", err)
		return result
	}

	// Setup Timeout
	timeout := time.Duration(job.TimeoutSeconds) * time.Second
	if timeout <= 0 || timeout > 60*time.Second {
		timeout = 10 * time.Second
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Setup Tracing for Latency Breakdown
	var (
		dnsStart, dnsDone   time.Time
		connStart, connDone time.Time
		tlsStart, tlsDone   time.Time
	)

	trace := &httptrace.ClientTrace{
		DNSStart:          func(info httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone:           func(info httptrace.DNSDoneInfo) { dnsDone = time.Now() },
		ConnectStart:      func(network, addr string) { connStart = time.Now() },
		ConnectDone:       func(network, addr string, err error) { connDone = time.Now() },
		TLSHandshakeStart: func() { tlsStart = time.Now() },
		TLSHandshakeDone:  func(state tls.ConnectionState, err error) { tlsDone = time.Now() },
	}

	var reqBody io.Reader
	if job.Body != "" {
		reqBody = bytes.NewBufferString(job.Body)
	}

	method := strings.ToUpper(job.Method)
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(reqCtx, trace), method, parsedURL.String(), reqBody)
	if err != nil {
		result.Status = "DOWN"
		result.ErrorMessage = fmt.Sprintf("failed to build request: %v", err)
		return result
	}

	req.Header.Set("User-Agent", "PulseWatch-ProbeWorker/1.0 (+https://pulsewatch.dev)")
	for k, v := range job.Headers {
		req.Header.Set(k, v)
	}

	// Execute HTTP Request
	start := time.Now()
	resp, err := p.client.Do(req)
	totalDuration := time.Since(start)

	result.ResponseTimeMs = int(totalDuration.Milliseconds())
	if !dnsStart.IsZero() && !dnsDone.IsZero() {
		result.DNSLookupMs = int(dnsDone.Sub(dnsStart).Milliseconds())
	}
	if !connStart.IsZero() && !connDone.IsZero() {
		result.TCPConnectMs = int(connDone.Sub(connStart).Milliseconds())
	}
	if !tlsStart.IsZero() && !tlsDone.IsZero() {
		result.TLSHandshakeMs = int(tlsDone.Sub(tlsStart).Milliseconds())
	}

	if err != nil {
		result.Status = "DOWN"
		result.ErrorMessage = sanitizeError(err)
		return result
	}
	defer resp.Body.Close()

	// Drain response body safely (max 64KB)
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))

	result.StatusCode = resp.StatusCode

	// Outcome Classification
	expectedStatus := job.ExpectedStatusCode
	if expectedStatus == 0 {
		expectedStatus = 200
	}

	isExpected := (result.StatusCode == expectedStatus) ||
		(job.ExpectedStatusCode == 0 && result.StatusCode >= 200 && result.StatusCode < 300)

	if isExpected {
		if result.ResponseTimeMs > 2500 {
			result.Status = "DEGRADED"
			result.ErrorMessage = fmt.Sprintf("High response latency: %dms", result.ResponseTimeMs)
		} else {
			result.Status = "UP"
		}
	} else {
		result.Status = "DOWN"
		result.ErrorMessage = fmt.Sprintf("HTTP status %d does not match expected %d", result.StatusCode, expectedStatus)
	}

	return result
}

func sanitizeError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if strings.Contains(msg, "context deadline exceeded") {
		return "Request timed out"
	}
	if strings.Contains(msg, "connection refused") {
		return "Connection refused"
	}
	if strings.Contains(msg, "no such host") {
		return "DNS resolution failed: host not found"
	}
	if strings.Contains(msg, "certificate") {
		return "TLS/SSL certificate validation failed"
	}
	return msg
}
