package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"pulsewatch-api/internal/config"
	"pulsewatch-api/internal/middleware"
	"pulsewatch-api/internal/repository"
	"pulsewatch-api/internal/service"
)

// AuthHandler handles HTTP requests for authentication and account security.
type AuthHandler struct {
	authService *service.AuthService
	cfg         *config.Config
}

func NewAuthHandler(authService *service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cfg:         cfg,
	}
}

// Signup handles user registration.
// POST /api/v1/auth/signup
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid_json", "Malformed JSON request body")
		return
	}

	if err := req.Validate(); err != nil {
		JSONError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}

	user, err := h.authService.Signup(r.Context(), service.SignupParams{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
	})
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			JSONError(w, http.StatusConflict, "user_exists", "An account with this email address already exists")
			return
		}
		JSONError(w, http.StatusInternalServerError, "server_error", "Failed to create account")
		return
	}

	JSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Account created successfully. Please check your email to verify your account.",
		"user":    user,
	})
}

// Login authenticates credentials and returns tokens.
// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid_json", "Malformed JSON request body")
		return
	}

	if err := req.Validate(); err != nil {
		JSONError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}

	userAgent := r.UserAgent()
	ipAddress := middleware.GetClientIP(r)

	resp, err := h.authService.Login(r.Context(), req.Email, req.Password, userAgent, ipAddress)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			JSONError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
			return
		}
		if errors.Is(err, service.ErrAccountUnverified) {
			JSONError(w, http.StatusForbidden, "account_unverified", err.Error())
			return
		}
		if errors.Is(err, service.ErrAccountDisabled) {
			JSONError(w, http.StatusForbidden, "account_disabled", err.Error())
			return
		}
		JSONError(w, http.StatusInternalServerError, "server_error", "An error occurred during authentication")
		return
	}

	// Set HttpOnly refresh token cookie
	h.setRefreshTokenCookie(w, resp.RefreshToken, h.cfg.RefreshTokenExpiry)

	JSON(w, http.StatusOK, resp)
}

// Refresh issues a new access token and rotated refresh token.
// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := h.getRefreshToken(r)
	if refreshToken == "" {
		JSONError(w, http.StatusUnauthorized, "missing_token", "Refresh token is missing")
		return
	}

	userAgent := r.UserAgent()
	ipAddress := middleware.GetClientIP(r)

	resp, err := h.authService.RefreshToken(r.Context(), refreshToken, userAgent, ipAddress)
	if err != nil {
		h.clearRefreshTokenCookie(w)
		if errors.Is(err, service.ErrRevokedTokenReused) {
			JSONError(w, http.StatusUnauthorized, "token_theft_detected", "Invalid token session. Please log in again.")
			return
		}
		if errors.Is(err, service.ErrInvalidCredentials) {
			JSONError(w, http.StatusUnauthorized, "invalid_token", "Refresh token is invalid or expired")
			return
		}
		JSONError(w, http.StatusInternalServerError, "server_error", "Failed to refresh authentication tokens")
		return
	}

	h.setRefreshTokenCookie(w, resp.RefreshToken, h.cfg.RefreshTokenExpiry)
	JSON(w, http.StatusOK, resp)
}

// Logout revokes the active refresh token and clears authentication cookies.
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken := h.getRefreshToken(r)
	if refreshToken != "" {
		_ = h.authService.Logout(r.Context(), refreshToken)
	}

	h.clearRefreshTokenCookie(w)
	JSONMessage(w, http.StatusOK, "Logged out successfully")
}

// Me returns the authenticated user's public profile.
// GET /api/v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		JSONError(w, http.StatusUnauthorized, "unauthorized", "User context not found")
		return
	}

	user, err := h.authService.GetCurrentUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			JSONError(w, http.StatusNotFound, "user_not_found", "User account not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "server_error", "Failed to load user profile")
		return
	}

	JSON(w, http.StatusOK, map[string]interface{}{
		"user": user,
	})
}

// ForgotPassword initiates an expiring password reset flow.
// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid_json", "Malformed JSON request body")
		return
	}

	if err := req.Validate(); err != nil {
		JSONError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}

	_ = h.authService.ForgotPassword(r.Context(), req.Email)

	// Return generic message to prevent account enumeration
	JSONMessage(w, http.StatusOK, "If an account exists with that email address, password reset instructions have been sent.")
}

// ResetPassword consumes the token and updates the user's password.
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid_json", "Malformed JSON request body")
		return
	}

	if err := req.Validate(); err != nil {
		JSONError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}

	if err := h.authService.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			JSONError(w, http.StatusBadRequest, "invalid_token", "Password reset token is invalid, expired, or already used")
			return
		}
		JSONError(w, http.StatusInternalServerError, "server_error", "Failed to reset password")
		return
	}

	JSONMessage(w, http.StatusOK, "Password has been reset successfully. You can now log in with your new password.")
}

// VerifyEmail consumes the verification token and activates the account.
// POST /api/v1/auth/verify-email
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid_json", "Malformed JSON request body")
		return
	}

	if err := req.Validate(); err != nil {
		JSONError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}

	if err := h.authService.VerifyEmail(r.Context(), req.Token); err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			JSONError(w, http.StatusBadRequest, "invalid_token", "Verification token is invalid, expired, or already consumed")
			return
		}
		JSONError(w, http.StatusInternalServerError, "server_error", "Failed to verify email address")
		return
	}

	JSONMessage(w, http.StatusOK, "Email address verified successfully. Your account is now active.")
}

// ResendVerification sends a new verification link.
// POST /api/v1/auth/resend-verification
func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid_json", "Malformed JSON request body")
		return
	}

	if err := req.Validate(); err != nil {
		JSONError(w, http.StatusUnprocessableEntity, "validation_error", err.Error())
		return
	}

	_ = h.authService.ResendVerification(r.Context(), req.Email)

	// Generic response preventing enumeration
	JSONMessage(w, http.StatusOK, "If an unverified account exists with that email address, a new verification link has been sent.")
}

// -----------------------------------------------------------------------------
// Cookie Helpers
// -----------------------------------------------------------------------------

func (h *AuthHandler) getRefreshToken(r *http.Request) string {
	// Check HttpOnly cookie
	if cookie, err := r.Cookie("refresh_token"); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Check JSON request body
	var req RefreshTokenRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	return req.RefreshToken
}

func (h *AuthHandler) setRefreshTokenCookie(w http.ResponseWriter, token string, duration time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth",
		Expires:  time.Now().Add(duration),
		HttpOnly: true,
		Secure:   h.cfg.Environment == "production",
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) clearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.cfg.Environment == "production",
		SameSite: http.SameSiteLaxMode,
	})
}
