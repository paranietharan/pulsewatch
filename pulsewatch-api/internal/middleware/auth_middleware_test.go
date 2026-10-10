package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"pulsewatch-api/internal/models"
	"pulsewatch-api/internal/service"
)

func TestAuthMiddleware(t *testing.T) {
	jwtSecret := "test_secret_for_auth_middleware_testing"
	tokenService := service.NewTokenService(jwtSecret, 1*time.Hour)

	user := &models.User{
		ID:        uuid.New(),
		Email:     "auth@example.com",
		Role:      models.RoleUser,
		FirstName: "Auth",
		LastName:  "Test",
	}

	validToken, err := tokenService.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, ok := GetUserID(r.Context())
		if !ok || uid != user.ID {
			http.Error(w, "invalid user id in context", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	mw := AuthMiddleware(tokenService)(testHandler)

	// Case 1: Missing header -> 401
	r1 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w1 := httptest.NewRecorder()
	mw.ServeHTTP(w1, r1)
	if w1.Code != http.StatusUnauthorized {
		t.Errorf("missing header: expected 401, got %d", w1.Code)
	}

	// Case 2: Malformed header -> 401
	r2 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r2.Header.Set("Authorization", "InvalidFormat")
	w2 := httptest.NewRecorder()
	mw.ServeHTTP(w2, r2)
	if w2.Code != http.StatusUnauthorized {
		t.Errorf("malformed header: expected 401, got %d", w2.Code)
	}

	// Case 3: Invalid token -> 401
	r3 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r3.Header.Set("Authorization", "Bearer invalid.jwt.signature")
	w3 := httptest.NewRecorder()
	mw.ServeHTTP(w3, r3)
	if w3.Code != http.StatusUnauthorized {
		t.Errorf("invalid token: expected 401, got %d", w3.Code)
	}

	// Case 4: Valid token -> 200 and context attached
	r4 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r4.Header.Set("Authorization", "Bearer "+validToken)
	w4 := httptest.NewRecorder()
	mw.ServeHTTP(w4, r4)
	if w4.Code != http.StatusOK {
		t.Errorf("valid token: expected 200, got %d. Body: %s", w4.Code, w4.Body.String())
	}
}
