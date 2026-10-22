package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"pulsewatch-api/internal/config"
	"pulsewatch-api/internal/database"
	"pulsewatch-api/internal/models"
	"pulsewatch-api/internal/rabbitmq"
	"pulsewatch-api/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountUnverified  = errors.New("email address is not verified; please verify your email before logging in")
	ErrAccountDisabled    = errors.New("account is disabled; please contact support")
	ErrRevokedTokenReused = errors.New("security alert: refresh token reuse detected; session terminated")
)

// AuthResponse holds credentials and user profile returned on successful login or token refresh.
type AuthResponse struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	ExpiresIn    int64           `json:"expires_in"` // Access token expiry in seconds
	User         models.SafeUser `json:"user"`
}

// SignupParams holds input parameters for registration.
type SignupParams struct {
	FirstName   string
	LastName    string
	Email       string
	Password    string
	FrontendURL string
}

// AuthService handles all authentication and account security business logic.
type AuthService struct {
	cfg          *config.Config
	db           *database.DB
	userRepo     *repository.UserRepository
	tokenRepo    *repository.TokenRepository
	outboxRepo   *repository.OutboxRepository
	tokenService *TokenService
	pwdService   *PasswordService
}

func NewAuthService(
	cfg *config.Config,
	db *database.DB,
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	outboxRepo *repository.OutboxRepository,
	tokenService *TokenService,
	pwdService *PasswordService,
) *AuthService {
	return &AuthService{
		cfg:          cfg,
		db:           db,
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		outboxRepo:   outboxRepo,
		tokenService: tokenService,
		pwdService:   pwdService,
	}
}

// Signup registers a new user and queues a verification email atomically.
func (s *AuthService) Signup(ctx context.Context, params SignupParams) (*models.SafeUser, error) {
	email := strings.TrimSpace(strings.ToLower(params.Email))
	firstName := strings.TrimSpace(params.FirstName)
	lastName := strings.TrimSpace(params.LastName)

	passwordHash, err := s.pwdService.HashPassword(params.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &models.User{
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         models.RoleUser,
		IsActive:     true,
		IsVerified:   false,
	}

	rawVerifyToken, tokenHash, err := s.tokenService.GenerateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification token: %w", err)
	}

	var createdUser *models.User

	// Execute transactional write for User, Verification Token, and Outbox Event
	err = s.db.WithTx(ctx, func(tx pgx.Tx) error {
		var createErr error
		createdUser, createErr = s.userRepo.Create(ctx, tx, newUser)
		if createErr != nil {
			return createErr
		}

		// Store verification token
		evt := &models.EmailVerificationToken{
			UserID:    createdUser.ID,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().UTC().Add(s.cfg.VerifyTokenExpiry),
		}
		if err := s.tokenRepo.CreateVerificationToken(ctx, tx, evt); err != nil {
			return err
		}

		frontendURL := strings.TrimRight(params.FrontendURL, "/")
		if frontendURL == "" {
			frontendURL = strings.TrimRight(s.cfg.FrontendURL, "/")
		}

		// Prepare Outbox Event
		eventPayload := rabbitmq.BaseEvent{
			EventID:       uuid.New(),
			EventType:     rabbitmq.RoutingKeyEmailVerification,
			SchemaVersion: "1.0",
			OccurredAt:    time.Now().UTC(),
			Payload: rabbitmq.VerificationEmailPayload{
				UserID:            createdUser.ID,
				Email:             createdUser.Email,
				Name:              createdUser.FullName(),
				VerificationToken: rawVerifyToken,
				ExpiresAt:         evt.ExpiresAt,
				FrontendURL:       frontendURL,
			},
		}

		payloadBytes, err := json.Marshal(eventPayload)
		if err != nil {
			return fmt.Errorf("failed to serialize outbox event: %w", err)
		}

		outboxEvent := &models.OutboxEvent{
			EventType:  rabbitmq.RoutingKeyEmailVerification,
			RoutingKey: rabbitmq.RoutingKeyEmailVerification,
			Payload:    payloadBytes,
		}

		return s.outboxRepo.Create(ctx, tx, outboxEvent)
	})

	if err != nil {
		return nil, err
	}

	safe := createdUser.ToSafeUser()
	return &safe, nil
}

// Login authenticates credentials, issues JWT access token and rotating refresh token session.
func (s *AuthService) Login(ctx context.Context, email, password, userAgent, ipAddress string) (*AuthResponse, error) {
	cleanEmail := strings.TrimSpace(strings.ToLower(email))

	user, err := s.userRepo.GetByEmail(ctx, cleanEmail)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password using constant-time comparison
	if err := s.pwdService.ComparePassword(user.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrAccountDisabled
	}

	if !user.IsVerified {
		return nil, ErrAccountUnverified
	}

	// Generate Access Token (JWT)
	accessToken, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate Refresh Token
	rawRefreshToken, tokenHash, err := s.tokenService.GenerateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	familyID := uuid.New()
	refreshToken := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		FamilyID:  familyID,
		IsRevoked: false,
		ExpiresAt: time.Now().UTC().Add(s.cfg.RefreshTokenExpiry),
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}

	if err := s.tokenRepo.CreateRefreshToken(ctx, nil, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Update last login timestamp asynchronously
	go func(id uuid.UUID) {
		_ = s.userRepo.UpdateLastLogin(context.Background(), id)
	}(user.ID)

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresIn:    int64(s.cfg.JWTAccessExpiry.Seconds()),
		User:         user.ToSafeUser(),
	}, nil
}

// RefreshToken implements refresh token rotation and token family theft detection.
func (s *AuthService) RefreshToken(ctx context.Context, rawRefreshToken, userAgent, ipAddress string) (*AuthResponse, error) {
	if rawRefreshToken == "" {
		return nil, ErrInvalidCredentials
	}

	tokenHash := s.tokenService.HashToken(rawRefreshToken)
	rt, err := s.tokenRepo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrTokenNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// REUSE DETECTION: If this token was already revoked, a compromised token is being replayed!
	if rt.IsRevoked {
		// Invalidate the entire family to protect the user
		_ = s.tokenRepo.RevokeFamily(ctx, nil, rt.FamilyID)
		return nil, ErrRevokedTokenReused
	}

	if rt.IsExpired() {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrAccountDisabled
	}

	// Rotate token: revoke old token, issue new token in the same family
	newRawRefreshToken, newTokenHash, err := s.tokenService.GenerateOpaqueToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	newRefreshToken := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: newTokenHash,
		FamilyID:  rt.FamilyID, // Keep same family
		IsRevoked: false,
		ExpiresAt: time.Now().UTC().Add(s.cfg.RefreshTokenExpiry),
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}

	err = s.db.WithTx(ctx, func(tx pgx.Tx) error {
		if err := s.tokenRepo.RevokeRefreshToken(ctx, tx, tokenHash); err != nil {
			return err
		}
		return s.tokenRepo.CreateRefreshToken(ctx, tx, newRefreshToken)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to rotate refresh token: %w", err)
	}

	newAccessToken, err := s.tokenService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRawRefreshToken,
		ExpiresIn:    int64(s.cfg.JWTAccessExpiry.Seconds()),
		User:         user.ToSafeUser(),
	}, nil
}

// Logout revokes the given refresh token session.
func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	if rawRefreshToken == "" {
		return nil
	}
	tokenHash := s.tokenService.HashToken(rawRefreshToken)
	_ = s.tokenRepo.RevokeRefreshToken(ctx, nil, tokenHash)
	return nil
}

// GetCurrentUser returns the sanitized user profile for the authenticated user ID.
func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*models.SafeUser, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	safe := user.ToSafeUser()
	return &safe, nil
}

// ForgotPassword generates a password reset token and queues an email via outbox.
// Returns success without leaking whether the email exists.
func (s *AuthService) ForgotPassword(ctx context.Context, email, callerFrontendURL string) error {
	cleanEmail := strings.TrimSpace(strings.ToLower(email))
	user, err := s.userRepo.GetByEmail(ctx, cleanEmail)
	if err != nil {
		// Return generic success to prevent account enumeration
		return nil
	}

	if !user.IsActive {
		return nil
	}

	rawResetToken, tokenHash, err := s.tokenService.GenerateOpaqueToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	prt := &models.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(s.cfg.ResetTokenExpiry),
	}

	frontendURL := strings.TrimRight(callerFrontendURL, "/")
	if frontendURL == "" {
		frontendURL = strings.TrimRight(s.cfg.FrontendURL, "/")
	}

	eventPayload := rabbitmq.BaseEvent{
		EventID:       uuid.New(),
		EventType:     rabbitmq.RoutingKeyEmailPasswordReset,
		SchemaVersion: "1.0",
		OccurredAt:    time.Now().UTC(),
		Payload: rabbitmq.PasswordResetEmailPayload{
			UserID:      user.ID,
			Email:       user.Email,
			Name:        user.FullName(),
			ResetToken:  rawResetToken,
			ExpiresAt:   prt.ExpiresAt,
			FrontendURL: frontendURL,
		},
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return fmt.Errorf("failed to serialize outbox event: %w", err)
	}

	outboxEvent := &models.OutboxEvent{
		EventType:  rabbitmq.RoutingKeyEmailPasswordReset,
		RoutingKey: rabbitmq.RoutingKeyEmailPasswordReset,
		Payload:    payloadBytes,
	}

	return s.db.WithTx(ctx, func(tx pgx.Tx) error {
		if err := s.tokenRepo.CreatePasswordResetToken(ctx, tx, prt); err != nil {
			return err
		}
		return s.outboxRepo.Create(ctx, tx, outboxEvent)
	})
}

// ResetPassword consumes the reset token, updates password, and revokes all active sessions.
func (s *AuthService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	tokenHash := s.tokenService.HashToken(rawToken)
	prt, err := s.tokenRepo.GetPasswordResetTokenByHash(ctx, tokenHash)
	if err != nil {
		return ErrInvalidCredentials
	}

	if !prt.IsValid() {
		return ErrInvalidCredentials
	}

	newPasswordHash, err := s.pwdService.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.db.WithTx(ctx, func(tx pgx.Tx) error {
		// Consume single-use token atomically
		if err := s.tokenRepo.ConsumePasswordResetToken(ctx, tx, tokenHash); err != nil {
			return err
		}

		// Update user password
		if err := s.userRepo.UpdatePassword(ctx, tx, prt.UserID, newPasswordHash); err != nil {
			return err
		}

		// Revoke all existing sessions for this user across devices
		return s.tokenRepo.RevokeUserSessions(ctx, tx, prt.UserID)
	})
}

// VerifyEmail consumes the email verification token and marks user verified.
func (s *AuthService) VerifyEmail(ctx context.Context, rawToken string) error {
	tokenHash := s.tokenService.HashToken(rawToken)
	evt, err := s.tokenRepo.GetVerificationTokenByHash(ctx, tokenHash)
	if err != nil {
		return ErrInvalidCredentials
	}

	if !evt.IsValid() {
		return ErrInvalidCredentials
	}

	return s.db.WithTx(ctx, func(tx pgx.Tx) error {
		if err := s.tokenRepo.ConsumeVerificationToken(ctx, tx, tokenHash); err != nil {
			return err
		}
		return s.userRepo.MarkEmailVerified(ctx, tx, evt.UserID)
	})
}

// ResendVerification generates and sends a new verification token if account is unverified.
func (s *AuthService) ResendVerification(ctx context.Context, email, callerFrontendURL string) error {
	cleanEmail := strings.TrimSpace(strings.ToLower(email))
	user, err := s.userRepo.GetByEmail(ctx, cleanEmail)
	if err != nil {
		// Avoid account enumeration
		return nil
	}

	if user.IsVerified || !user.IsActive {
		return nil
	}

	rawVerifyToken, tokenHash, err := s.tokenService.GenerateOpaqueToken()
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}

	evt := &models.EmailVerificationToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(s.cfg.VerifyTokenExpiry),
	}

	frontendURL := strings.TrimRight(callerFrontendURL, "/")
	if frontendURL == "" {
		frontendURL = strings.TrimRight(s.cfg.FrontendURL, "/")
	}

	eventPayload := rabbitmq.BaseEvent{
		EventID:       uuid.New(),
		EventType:     rabbitmq.RoutingKeyEmailResendVerification,
		SchemaVersion: "1.0",
		OccurredAt:    time.Now().UTC(),
		Payload: rabbitmq.VerificationEmailPayload{
			UserID:            user.ID,
			Email:             user.Email,
			Name:              user.FullName(),
			VerificationToken: rawVerifyToken,
			ExpiresAt:         evt.ExpiresAt,
			FrontendURL:       frontendURL,
		},
	}

	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return fmt.Errorf("failed to serialize outbox event: %w", err)
	}

	outboxEvent := &models.OutboxEvent{
		EventType:  rabbitmq.RoutingKeyEmailResendVerification,
		RoutingKey: rabbitmq.RoutingKeyEmailResendVerification,
		Payload:    payloadBytes,
	}

	return s.db.WithTx(ctx, func(tx pgx.Tx) error {
		if err := s.tokenRepo.CreateVerificationToken(ctx, tx, evt); err != nil {
			return err
		}
		return s.outboxRepo.Create(ctx, tx, outboxEvent)
	})
}
