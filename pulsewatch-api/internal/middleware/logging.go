package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

// RequestLogger logs incoming HTTP requests with latency, status code, and correlation ID.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = uuid.New().String()
			}
			w.Header().Set("X-Request-ID", reqID)

			wrapper := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			start := time.Now()
			next.ServeHTTP(wrapper, r)
			duration := time.Since(start)

			logger.Info("HTTP request",
				"request_id", reqID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapper.statusCode,
				"duration_ms", duration.Milliseconds(),
				"ip", GetClientIP(r),
			)
		})
	}
}
