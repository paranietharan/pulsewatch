package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recoverer catches unhandled panics, logs the stack trace, and returns a clean 500 JSON response.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("Unhandled panic in HTTP handler",
						"panic", rec,
						"stack", string(debug.Stack()),
						"path", r.URL.Path,
						"method", r.Method,
					)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":"internal_server_error","message":"An unexpected error occurred. Please try again later."}`))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
