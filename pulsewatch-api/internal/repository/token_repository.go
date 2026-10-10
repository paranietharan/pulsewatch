package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"pulsewatch-api/internal/models"
)

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenConsumed = errors.New("token already consumed")
	ErrTokenExpired  = errors.New("token expired")
)

// TokenRepository handles persistence for refresh tokens, verification tokens, and reset tokens.
type TokenRepository struct {
	pool *pgxpool.Pool
}

func NewTokenRepository(pool *pgxpool.Pool) *TokenRepository {
	return &TokenRepository{pool: pool}
}

// -----------------------------------------------------------------------------
// Refresh Tokens
// -----------------------------------------------------------------------------

// CreateRefreshToken inserts a new refresh token session.
func (r *TokenRepository) CreateRefreshToken(ctx context.Context, tx pgx.Tx, rt *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, family_id, is_revoked, expires_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at;
	`
	var (
		row pgx.Row
	)
	if tx != nil {
		row = tx.QueryRow(ctx, query, rt.UserID, rt.TokenHash, rt.FamilyID, rt.IsRevoked, rt.ExpiresAt, rt.UserAgent, rt.IPAddress)
	} else {
		row = r.pool.QueryRow(ctx, query, rt.UserID, rt.TokenHash, rt.FamilyID, rt.IsRevoked, rt.ExpiresAt, rt.UserAgent, rt.IPAddress)
	}

	if err := row.Scan(&rt.ID, &rt.CreatedAt); err != nil {
		return fmt.Errorf("failed to insert refresh token: %w", err)
	}
	return nil
}

// GetRefreshTokenByHash retrieves a refresh token by its SHA-256 hash.
func (r *TokenRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, family_id, is_revoked, expires_at, user_agent, ip_address, created_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1;
	`
	var rt models.RefreshToken
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.TokenHash,
		&rt.FamilyID,
		&rt.IsRevoked,
		&rt.ExpiresAt,
		&rt.UserAgent,
		&rt.IPAddress,
		&rt.CreatedAt,
		&rt.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	return &rt, nil
}

// RevokeRefreshToken marks a specific refresh token as revoked.
func (r *TokenRepository) RevokeRefreshToken(ctx context.Context, tx pgx.Tx, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = TRUE, revoked_at = $1
		WHERE token_hash = $2;
	`
	now := time.Now().UTC()
	var (
		res pgconn.CommandTag
		err error
	)
	if tx != nil {
		res, err = tx.Exec(ctx, query, now, tokenHash)
	} else {
		res, err = r.pool.Exec(ctx, query, now, tokenHash)
	}
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrTokenNotFound
	}
	return nil
}

// RevokeFamily revokes an entire family of refresh tokens (used when token reuse is detected).
func (r *TokenRepository) RevokeFamily(ctx context.Context, tx pgx.Tx, familyID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = TRUE, revoked_at = $1
		WHERE family_id = $2 AND is_revoked = FALSE;
	`
	now := time.Now().UTC()
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, query, now, familyID)
	} else {
		_, err = r.pool.Exec(ctx, query, now, familyID)
	}
	if err != nil {
		return fmt.Errorf("failed to revoke token family: %w", err)
	}
	return nil
}

// RevokeUserSessions revokes all refresh tokens for a user (used upon password reset).
func (r *TokenRepository) RevokeUserSessions(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = TRUE, revoked_at = $1
		WHERE user_id = $2 AND is_revoked = FALSE;
	`
	now := time.Now().UTC()
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, query, now, userID)
	} else {
		_, err = r.pool.Exec(ctx, query, now, userID)
	}
	if err != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}
	return nil
}

// -----------------------------------------------------------------------------
// Email Verification Tokens
// -----------------------------------------------------------------------------

// CreateVerificationToken stores a new email verification token.
func (r *TokenRepository) CreateVerificationToken(ctx context.Context, tx pgx.Tx, evt *models.EmailVerificationToken) error {
	query := `
		INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at;
	`
	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, query, evt.UserID, evt.TokenHash, evt.ExpiresAt)
	} else {
		row = r.pool.QueryRow(ctx, query, evt.UserID, evt.TokenHash, evt.ExpiresAt)
	}
	if err := row.Scan(&evt.ID, &evt.CreatedAt); err != nil {
		return fmt.Errorf("failed to insert verification token: %w", err)
	}
	return nil
}

// GetVerificationTokenByHash retrieves an email verification token by its hash.
func (r *TokenRepository) GetVerificationTokenByHash(ctx context.Context, tokenHash string) (*models.EmailVerificationToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, consumed_at, created_at
		FROM email_verification_tokens
		WHERE token_hash = $1;
	`
	var evt models.EmailVerificationToken
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&evt.ID,
		&evt.UserID,
		&evt.TokenHash,
		&evt.ExpiresAt,
		&evt.ConsumedAt,
		&evt.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get verification token: %w", err)
	}
	return &evt, nil
}

// ConsumeVerificationToken marks an email verification token as consumed.
func (r *TokenRepository) ConsumeVerificationToken(ctx context.Context, tx pgx.Tx, tokenHash string) error {
	query := `
		UPDATE email_verification_tokens
		SET consumed_at = $1
		WHERE token_hash = $2 AND consumed_at IS NULL AND expires_at > $1;
	`
	now := time.Now().UTC()
	var (
		res pgconn.CommandTag
		err error
	)
	if tx != nil {
		res, err = tx.Exec(ctx, query, now, tokenHash)
	} else {
		res, err = r.pool.Exec(ctx, query, now, tokenHash)
	}
	if err != nil {
		return fmt.Errorf("failed to consume verification token: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrTokenConsumed
	}
	return nil
}

// -----------------------------------------------------------------------------
// Password Reset Tokens
// -----------------------------------------------------------------------------

// CreatePasswordResetToken stores a new password reset token.
func (r *TokenRepository) CreatePasswordResetToken(ctx context.Context, tx pgx.Tx, prt *models.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at;
	`
	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, query, prt.UserID, prt.TokenHash, prt.ExpiresAt)
	} else {
		row = r.pool.QueryRow(ctx, query, prt.UserID, prt.TokenHash, prt.ExpiresAt)
	}
	if err := row.Scan(&prt.ID, &prt.CreatedAt); err != nil {
		return fmt.Errorf("failed to insert password reset token: %w", err)
	}
	return nil
}

// GetPasswordResetTokenByHash retrieves a password reset token by its hash.
func (r *TokenRepository) GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*models.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, consumed_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1;
	`
	var prt models.PasswordResetToken
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&prt.ID,
		&prt.UserID,
		&prt.TokenHash,
		&prt.ExpiresAt,
		&prt.ConsumedAt,
		&prt.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get password reset token: %w", err)
	}
	return &prt, nil
}

// ConsumePasswordResetToken marks a password reset token as consumed.
func (r *TokenRepository) ConsumePasswordResetToken(ctx context.Context, tx pgx.Tx, tokenHash string) error {
	query := `
		UPDATE password_reset_tokens
		SET consumed_at = $1
		WHERE token_hash = $2 AND consumed_at IS NULL AND expires_at > $1;
	`
	now := time.Now().UTC()
	var (
		res pgconn.CommandTag
		err error
	)
	if tx != nil {
		res, err = tx.Exec(ctx, query, now, tokenHash)
	} else {
		res, err = r.pool.Exec(ctx, query, now, tokenHash)
	}
	if err != nil {
		return fmt.Errorf("failed to consume password reset token: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrTokenConsumed
	}
	return nil
}
