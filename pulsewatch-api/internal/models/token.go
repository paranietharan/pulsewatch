package models

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a long-lived user session stored in the database.
type RefreshToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TokenHash string     `json:"-"` // SHA-256 hash of opaque token
	FamilyID  uuid.UUID  `json:"family_id"`
	IsRevoked bool       `json:"is_revoked"`
	ExpiresAt time.Time  `json:"expires_at"`
	UserAgent string     `json:"user_agent"`
	IPAddress string     `json:"ip_address"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// IsExpired checks if the refresh token has expired.
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// EmailVerificationToken represents an expiring single-use email verification token.
type EmailVerificationToken struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	TokenHash  string     `json:"-"`
	ExpiresAt  time.Time  `json:"expires_at"`
	ConsumedAt *time.Time `json:"consumed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// IsValid checks if the token can still be used.
func (evt *EmailVerificationToken) IsValid() bool {
	return evt.ConsumedAt == nil && time.Now().Before(evt.ExpiresAt)
}

// PasswordResetToken represents an expiring single-use password reset token.
type PasswordResetToken struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	TokenHash  string     `json:"-"`
	ExpiresAt  time.Time  `json:"expires_at"`
	ConsumedAt *time.Time `json:"consumed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// IsValid checks if the token can still be used.
func (prt *PasswordResetToken) IsValid() bool {
	return prt.ConsumedAt == nil && time.Now().Before(prt.ExpiresAt)
}
