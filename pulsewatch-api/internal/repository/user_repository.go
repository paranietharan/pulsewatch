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
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

// UserRepository handles database queries for users.
type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create inserts a new user record inside an optional transaction.
func (r *UserRepository) Create(ctx context.Context, tx pgx.Tx, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (first_name, last_name, email, password_hash, role, is_active, is_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, first_name, last_name, email, role, is_active, is_verified, created_at, updated_at;
	`
	var (
		created models.User
		row     pgx.Row
	)

	if tx != nil {
		row = tx.QueryRow(ctx, query, user.FirstName, user.LastName, user.Email, user.PasswordHash, user.Role, user.IsActive, user.IsVerified)
	} else {
		row = r.pool.QueryRow(ctx, query, user.FirstName, user.LastName, user.Email, user.PasswordHash, user.Role, user.IsActive, user.IsVerified)
	}

	err := row.Scan(
		&created.ID,
		&created.FirstName,
		&created.LastName,
		&created.Email,
		&created.Role,
		&created.IsActive,
		&created.IsVerified,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	created.PasswordHash = user.PasswordHash
	return &created, nil
}

// GetByID finds a user by their UUID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, first_name, last_name, email, password_hash, role, is_active, is_verified, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1;
	`
	var user models.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}
	return &user, nil
}

// GetByEmail finds a user by email address.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, first_name, last_name, email, password_hash, role, is_active, is_verified, created_at, updated_at, last_login_at
		FROM users
		WHERE LOWER(email) = LOWER($1);
	`
	var user models.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.IsVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}
	return &user, nil
}

// UpdatePassword updates the user's password hash.
func (r *UserRepository) UpdatePassword(ctx context.Context, tx pgx.Tx, userID uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2;`
	var (
		res pgconn.CommandTag
		err error
	)

	if tx != nil {
		res, err = tx.Exec(ctx, query, passwordHash, userID)
	} else {
		res, err = r.pool.Exec(ctx, query, passwordHash, userID)
	}
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// MarkEmailVerified updates is_verified to true.
func (r *UserRepository) MarkEmailVerified(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	query := `UPDATE users SET is_verified = TRUE, updated_at = NOW() WHERE id = $1;`
	var (
		res pgconn.CommandTag
		err error
	)

	if tx != nil {
		res, err = tx.Exec(ctx, query, userID)
	} else {
		res, err = r.pool.Exec(ctx, query, userID)
	}
	if err != nil {
		return fmt.Errorf("failed to mark email verified: %w", err)
	}
	if res.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdateLastLogin updates the last_login_at timestamp.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET last_login_at = $1 WHERE id = $2;`
	_, err := r.pool.Exec(ctx, query, time.Now().UTC(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last login timestamp: %w", err)
	}
	return nil
}
