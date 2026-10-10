package models

import (
	"time"

	"github.com/google/uuid"
)

// Role represents the user's authorization level in the system.
type Role string

const (
	RoleUser  Role = "USER"
	RoleAdmin Role = "ADMIN"
)

// User represents a user account record in the database.
type User struct {
	ID           uuid.UUID  `json:"id"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // Never expose in JSON responses
	Role         Role       `json:"role"`
	IsActive     bool       `json:"is_active"`
	IsVerified   bool       `json:"is_verified"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

// FullName returns the combined first and last name.
func (u *User) FullName() string {
	if u.LastName == "" {
		return u.FirstName
	}
	return u.FirstName + " " + u.LastName
}

// SafeUser is the public sanitized DTO returned to clients.
type SafeUser struct {
	ID          uuid.UUID  `json:"id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Email       string     `json:"email"`
	Role        Role       `json:"role"`
	IsActive    bool       `json:"is_active"`
	IsVerified  bool       `json:"is_verified"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// ToSafeUser converts a User model to SafeUser.
func (u *User) ToSafeUser() SafeUser {
	return SafeUser{
		ID:          u.ID,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		Email:       u.Email,
		Role:        u.Role,
		IsActive:    u.IsActive,
		IsVerified:  u.IsVerified,
		CreatedAt:   u.CreatedAt,
		LastLoginAt: u.LastLoginAt,
	}
}
