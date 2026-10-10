package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"pulsewatch-api/internal/models"
	"pulsewatch-api/internal/service"
)

type contextKey string

const (
	UserIDKey    contextKey = "userID"
	UserEmailKey contextKey = "userEmail"
	UserRoleKey  contextKey = "userRole"
)

// AuthMiddleware validates JWT Bearer tokens and attaches claims to the request context.
func AuthMiddleware(tokenService *service.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"unauthorized","message":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, `{"error":"unauthorized","message":"invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := tokenService.ValidateAccessToken(parts[1])
			if err != nil {
				http.Error(w, `{"error":"unauthorized","message":"invalid or expired access token"}`, http.StatusUnauthorized)
				return
			}

			// Store in context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID retrieves the authenticated user's UUID from the request context.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	val, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return val, ok
}

// GetUserEmail retrieves the authenticated user's email address from context.
func GetUserEmail(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(UserEmailKey).(string)
	return val, ok
}

// GetUserRole retrieves the authenticated user's role from context.
func GetUserRole(ctx context.Context) (models.Role, bool) {
	val, ok := ctx.Value(UserRoleKey).(models.Role)
	return val, ok
}
