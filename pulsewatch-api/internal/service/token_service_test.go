package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"pulsewatch-api/internal/models"
)

func TestTokenService_JWTFlow(t *testing.T) {
	jwtSecret := "test_secret_key_minimum_length_required"
	svc := NewTokenService(jwtSecret, 2*time.Second)

	user := &models.User{
		ID:        uuid.New(),
		Email:     "user@example.com",
		Role:      models.RoleUser,
		FirstName: "John",
		LastName:  "Doe",
	}

	// 1. Generate token
	tokenStr, err := svc.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	// 2. Validate valid token
	claims, err := svc.ValidateAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}
	if claims.UserID != user.ID {
		t.Errorf("expected userID %v, got %v", user.ID, claims.UserID)
	}
	if claims.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, claims.Email)
	}
	if claims.Role != user.Role {
		t.Errorf("expected role %s, got %s", user.Role, claims.Role)
	}

	// 3. Reject tampered token
	tampered := tokenStr + "tampered"
	_, err = svc.ValidateAccessToken(tampered)
	if err == nil {
		t.Errorf("expected error on tampered token, got nil")
	}

	// 4. Reject token signed with different secret
	diffSvc := NewTokenService("different_secret_key_1234567890", 2*time.Second)
	_, err = diffSvc.ValidateAccessToken(tokenStr)
	if err == nil {
		t.Errorf("expected signature verification failure, got nil")
	}
}

func TestTokenService_OpaqueTokens(t *testing.T) {
	svc := NewTokenService("secret1234567890123456", time.Minute)

	raw1, hash1, err := svc.GenerateOpaqueToken()
	if err != nil {
		t.Fatalf("failed to generate opaque token: %v", err)
	}

	raw2, hash2, err := svc.GenerateOpaqueToken()
	if err != nil {
		t.Fatalf("failed to generate second opaque token: %v", err)
	}

	// Uniqueness
	if raw1 == raw2 || hash1 == hash2 {
		t.Fatalf("expected cryptographically unique tokens")
	}

	// Deterministic hashing
	if svc.HashToken(raw1) != hash1 {
		t.Errorf("expected hash to be deterministic")
	}

	// Constant time comparison
	if !svc.ConstantTimeCompare(hash1, hash1) {
		t.Errorf("expected identical hashes to match")
	}
	if svc.ConstantTimeCompare(hash1, hash2) {
		t.Errorf("expected different hashes not to match")
	}
}
