package middleware

import (
	"net/http"
	"sync"
	"time"
)

type clientRecord struct {
	tokens     float64
	lastUpdate time.Time
}

// RateLimiter implements a thread-safe token bucket rate limiter per client IP.
type RateLimiter struct {
	mu         sync.RWMutex
	records    map[string]*clientRecord
	rate       float64       // Tokens added per second
	capacity   float64       // Maximum burst capacity
	cleanupInt time.Duration // Interval to purge stale records
}

// NewRateLimiter creates a rate limiter allowing capacity requests with a refill rate.
func NewRateLimiter(rate float64, capacity float64) *RateLimiter {
	rl := &RateLimiter{
		records:    make(map[string]*clientRecord),
		rate:       rate,
		capacity:   capacity,
		cleanupInt: 5 * time.Minute,
	}

	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	rec, exists := rl.records[ip]
	if !exists {
		rl.records[ip] = &clientRecord{
			tokens:     rl.capacity - 1,
			lastUpdate: now,
		}
		return true
	}

	// Refill tokens
	elapsed := now.Sub(rec.lastUpdate).Seconds()
	rec.tokens += elapsed * rl.rate
	if rec.tokens > rl.capacity {
		rec.tokens = rl.capacity
	}
	rec.lastUpdate = now

	if rec.tokens >= 1.0 {
		rec.tokens -= 1.0
		return true
	}

	return false
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInt)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, rec := range rl.records {
			if now.Sub(rec.lastUpdate) > 10*time.Minute {
				delete(rl.records, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns an HTTP middleware wrapping handlers with rate limiting.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetClientIP(r)
			if !rl.allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"too_many_requests","message":"Rate limit exceeded. Please try again later."}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
