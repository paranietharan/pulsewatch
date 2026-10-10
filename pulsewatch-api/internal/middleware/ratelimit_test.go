package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter(t *testing.T) {
	// Limiter: 1 request/sec with burst capacity 2
	limiter := NewRateLimiter(1.0, 2.0)

	handler := limiter.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	// First request: OK
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req)
	if rec1.Code != http.StatusOK {
		t.Errorf("request 1: expected 200, got %d", rec1.Code)
	}

	// Second request: OK (within burst capacity)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusOK {
		t.Errorf("request 2: expected 200, got %d", rec2.Code)
	}

	// Third request: Exceeded limit -> 429
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req)
	if rec3.Code != http.StatusTooManyRequests {
		t.Errorf("request 3: expected 429 Too Many Requests, got %d", rec3.Code)
	}

	// Different IP: Still OK
	reqDiff := httptest.NewRequest(http.MethodPost, "/test", nil)
	reqDiff.RemoteAddr = "10.0.0.5:12345"
	recDiff := httptest.NewRecorder()
	handler.ServeHTTP(recDiff, reqDiff)
	if recDiff.Code != http.StatusOK {
		t.Errorf("different IP: expected 200, got %d", recDiff.Code)
	}
}
