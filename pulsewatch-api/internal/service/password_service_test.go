package service

import (
	"testing"
)

func TestPasswordService_HashAndCompare(t *testing.T) {
	svc := NewPasswordService()
	password := "SecretPassword123!"

	hash, err := svc.HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	if hash == "" || hash == password {
		t.Fatalf("expected non-empty encrypted hash, got %s", hash)
	}

	// Successful comparison
	if err := svc.ComparePassword(hash, password); err != nil {
		t.Fatalf("expected password match, got: %v", err)
	}

	// Failed comparison
	if err := svc.ComparePassword(hash, "WrongPassword123!"); err == nil {
		t.Fatalf("expected password mismatch error, got nil")
	}
}
