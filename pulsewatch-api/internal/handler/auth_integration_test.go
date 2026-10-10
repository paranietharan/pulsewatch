package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"pulsewatch-api/internal/config"
	"pulsewatch-api/internal/database"
	"pulsewatch-api/internal/handler"
	"pulsewatch-api/internal/middleware"
	"pulsewatch-api/internal/models"
	"pulsewatch-api/internal/repository"
	"pulsewatch-api/internal/service"
)

func setupTestApp(t *testing.T) (*chi.Mux, *database.DB, *service.TokenService, func()) {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://pulsewatch:pulsewatch_secret@localhost:5432/pulsewatch?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, dbURL)
	if err != nil {
		t.Skipf("Skipping integration test: PostgreSQL is not accessible (%v)", err)
	}

	cfg := &config.Config{
		JWTSecret:          "test_integration_jwt_secret_1234567890",
		JWTAccessExpiry:    15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		VerifyTokenExpiry:  24 * time.Hour,
		ResetTokenExpiry:   1 * time.Hour,
		Environment:        "test",
	}

	userRepo := repository.NewUserRepository(db.Pool)
	tokenRepo := repository.NewTokenRepository(db.Pool)
	outboxRepo := repository.NewOutboxRepository(db.Pool)
	pwdService := service.NewPasswordService()
	tokenService := service.NewTokenService(cfg.JWTSecret, cfg.JWTAccessExpiry)

	authService := service.NewAuthService(cfg, db, userRepo, tokenRepo, outboxRepo, tokenService, pwdService)
	authHandler := handler.NewAuthHandler(authService, cfg)

	r := chi.NewRouter()
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.Signup)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)
		r.Post("/forgot-password", authHandler.ForgotPassword)
		r.Post("/reset-password", authHandler.ResetPassword)
		r.Post("/verify-email", authHandler.VerifyEmail)
		r.Post("/resend-verification", authHandler.ResendVerification)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(tokenService))
			r.Get("/me", authHandler.Me)
		})
	})

	cleanup := func() {
		// Clean up created test users
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE email LIKE '%@integration-test.dev';")
		db.Close()
	}

	return r, db, tokenService, cleanup
}

func TestAuthFlow_EndToEnd(t *testing.T) {
	router, db, _, cleanup := setupTestApp(t)
	defer cleanup()

	testEmail := fmt.Sprintf("user_%d@integration-test.dev", time.Now().UnixNano())
	testPassword := "CorrectPassword123!"

	// -------------------------------------------------------------------------
	// Signup
	// -------------------------------------------------------------------------
	signupBody, _ := json.Marshal(map[string]string{
		"first_name": "Integration",
		"last_name":  "Tester",
		"email":      testEmail,
		"password":   testPassword,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(signupBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Signup failed: expected 201 Created, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify duplicate signup returns 409 Conflict
	dupReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewReader(signupBody))
	dupReq.Header.Set("Content-Type", "application/json")
	dupW := httptest.NewRecorder()
	router.ServeHTTP(dupW, dupReq)

	if dupW.Code != http.StatusConflict {
		t.Errorf("Duplicate signup: expected 409 Conflict, got %d", dupW.Code)
	}

	// Verify Outbox event was recorded in DB
	var outboxCount int
	err := db.Pool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM auth_outbox_events WHERE routing_key = 'notification.email.verification.requested';",
	).Scan(&outboxCount)
	if err != nil || outboxCount == 0 {
		t.Fatalf("Expected outbox event for verification email to be created, count: %d, err: %v", outboxCount, err)
	}

	// -------------------------------------------------------------------------
	// Login Before Verification -> Should Fail (403 Forbidden)
	// -------------------------------------------------------------------------
	loginBody, _ := json.Marshal(map[string]string{
		"email":    testEmail,
		"password": testPassword,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("Login unverified: expected 403 Forbidden, got %d. Body: %s", w.Code, w.Body.String())
	}

	// -------------------------------------------------------------------------
	// Email Verification
	// -------------------------------------------------------------------------
	// Extract raw verification token from the outbox payload in PostgreSQL
	var eventPayloadJSON []byte
	err = db.Pool.QueryRow(context.Background(),
		"SELECT payload FROM auth_outbox_events WHERE routing_key = 'notification.email.verification.requested' ORDER BY created_at DESC LIMIT 1;",
	).Scan(&eventPayloadJSON)
	if err != nil {
		t.Fatalf("Failed to query outbox payload: %v", err)
	}

	var parsedEnvelope struct {
		Payload struct {
			VerificationToken string `json:"verification_token"`
		} `json:"payload"`
	}
	_ = json.Unmarshal(eventPayloadJSON, &parsedEnvelope)
	rawVerifyToken := parsedEnvelope.Payload.VerificationToken
	if rawVerifyToken == "" {
		t.Fatalf("Failed to extract verification token from outbox payload")
	}

	verifyBody, _ := json.Marshal(map[string]string{"token": rawVerifyToken})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", bytes.NewReader(verifyBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Verify email failed: expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Re-verification with same token must fail
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-email", bytes.NewReader(verifyBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Reusing verification token: expected 400 Bad Request, got %d", w.Code)
	}

	// -------------------------------------------------------------------------
	// Successful Login After Verification
	// -------------------------------------------------------------------------
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Login verified failed: expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var authResp service.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &authResp); err != nil {
		t.Fatalf("Failed to parse login response: %v", err)
	}
	if authResp.AccessToken == "" || authResp.RefreshToken == "" {
		t.Fatalf("Expected access and refresh tokens, got empty")
	}
	if !authResp.User.IsVerified {
		t.Errorf("Expected user.IsVerified to be true")
	}

	// Check HttpOnly cookie was set
	cookies := w.Result().Cookies()
	var foundCookie bool
	for _, c := range cookies {
		if c.Name == "refresh_token" && c.Value == authResp.RefreshToken && c.HttpOnly {
			foundCookie = true
			break
		}
	}
	if !foundCookie {
		t.Errorf("Expected HttpOnly refresh_token cookie to be set")
	}

	// -------------------------------------------------------------------------
	// Get Current User (/me)
	// -------------------------------------------------------------------------
	req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+authResp.AccessToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("/me failed: expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var meResp struct {
		User models.SafeUser `json:"user"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &meResp)
	if meResp.User.Email != testEmail {
		t.Errorf("Expected email %s, got %s", testEmail, meResp.User.Email)
	}

	// -------------------------------------------------------------------------
	// Refresh Token Rotation & Reuse Detection
	// -------------------------------------------------------------------------
	refreshBody, _ := json.Marshal(map[string]string{
		"refresh_token": authResp.RefreshToken,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(refreshBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Refresh failed: expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var rotatedResp service.AuthResponse
	_ = json.Unmarshal(w.Body.Bytes(), &rotatedResp)
	if rotatedResp.AccessToken == "" || rotatedResp.RefreshToken == "" {
		t.Fatalf("Expected new rotated tokens")
	}
	if rotatedResp.RefreshToken == authResp.RefreshToken {
		t.Fatalf("Expected new rotated refresh token, got identical token")
	}

	// THEFT/REUSE DETECTION: Try to use the old refresh token again!
	oldTokenReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(refreshBody))
	oldTokenReq.Header.Set("Content-Type", "application/json")
	oldW := httptest.NewRecorder()
	router.ServeHTTP(oldW, oldTokenReq)

	if oldW.Code != http.StatusUnauthorized {
		t.Errorf("Token reuse: expected 401 Unauthorized, got %d", oldW.Code)
	}

	// -------------------------------------------------------------------------
	// Forgot Password & Password Reset Flow
	// -------------------------------------------------------------------------
	forgotBody, _ := json.Marshal(map[string]string{"email": testEmail})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(forgotBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Forgot password failed: expected 200, got %d", w.Code)
	}

	// Extract reset token from outbox payload
	err = db.Pool.QueryRow(context.Background(),
		"SELECT payload FROM auth_outbox_events WHERE routing_key = 'notification.email.password_reset.requested' ORDER BY created_at DESC LIMIT 1;",
	).Scan(&eventPayloadJSON)
	if err != nil {
		t.Fatalf("Failed to query password reset outbox event: %v", err)
	}

	var resetEnvelope struct {
		Payload struct {
			ResetToken string `json:"reset_token"`
		} `json:"payload"`
	}
	_ = json.Unmarshal(eventPayloadJSON, &resetEnvelope)
	rawResetToken := resetEnvelope.Payload.ResetToken

	// Reset password with token
	newPassword := "BrandNewPassword123!"
	resetBody, _ := json.Marshal(map[string]string{
		"token":        rawResetToken,
		"new_password": newPassword,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(resetBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Reset password failed: expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Login with old password should fail
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Login with old password: expected 401, got %d", w.Code)
	}

	// Login with new password should succeed
	newLoginBody, _ := json.Marshal(map[string]string{
		"email":    testEmail,
		"password": newPassword,
	})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(newLoginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Login with new password failed: expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// -------------------------------------------------------------------------
	// Logout
	// -------------------------------------------------------------------------
	var newAuthResp service.AuthResponse
	_ = json.Unmarshal(w.Body.Bytes(), &newAuthResp)

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutReq.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: newAuthResp.RefreshToken,
	})
	logoutW := httptest.NewRecorder()
	router.ServeHTTP(logoutW, logoutReq)

	if logoutW.Code != http.StatusOK {
		t.Fatalf("Logout failed: expected 200, got %d", logoutW.Code)
	}

	// Verify cookie cleared
	foundCleared := false
	for _, c := range logoutW.Result().Cookies() {
		if c.Name == "refresh_token" && c.MaxAge == -1 {
			foundCleared = true
			break
		}
	}
	if !foundCleared {
		t.Errorf("Expected refresh_token cookie to be cleared with MaxAge -1")
	}
}
