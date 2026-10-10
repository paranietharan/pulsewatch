package service

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// PasswordService handles secure password hashing and verification.
type PasswordService struct{}

func NewPasswordService() *PasswordService {
	return &PasswordService{}
}

// HashPassword generates a bcrypt hash of the given plaintext password.
func (s *PasswordService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// ComparePassword compares a bcrypt hashed password with its possible plaintext equivalent.
func (s *PasswordService) ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
