package handler

import (
	"testing"
)

func TestDTOValidation(t *testing.T) {
	// Signup validation
	t.Run("SignupRequest", func(t *testing.T) {
		valid := SignupRequest{
			FirstName: "Alice",
			LastName:  "Smith",
			Email:     "alice@example.com",
			Password:  "password123",
		}
		if err := valid.Validate(); err != nil {
			t.Errorf("expected valid, got error: %v", err)
		}

		// Invalid short password
		shortPwd := valid
		shortPwd.Password = "short"
		if err := shortPwd.Validate(); err == nil {
			t.Errorf("expected error for short password, got nil")
		}

		// Invalid email
		invalidEmail := valid
		invalidEmail.Email = "notanemail"
		if err := invalidEmail.Validate(); err == nil {
			t.Errorf("expected error for invalid email, got nil")
		}

		// Missing first name
		missingName := valid
		missingName.FirstName = ""
		if err := missingName.Validate(); err == nil {
			t.Errorf("expected error for missing first name, got nil")
		}
	})

	// Login validation
	t.Run("LoginRequest", func(t *testing.T) {
		valid := LoginRequest{Email: "user@example.com", Password: "anyPassword"}
		if err := valid.Validate(); err != nil {
			t.Errorf("expected valid, got: %v", err)
		}

		missingEmail := LoginRequest{Email: "", Password: "pwd"}
		if err := missingEmail.Validate(); err == nil {
			t.Errorf("expected error for missing email, got nil")
		}
	})

	// ResetPassword validation
	t.Run("ResetPasswordRequest", func(t *testing.T) {
		valid := ResetPasswordRequest{Token: "someToken123", NewPassword: "newpassword123"}
		if err := valid.Validate(); err != nil {
			t.Errorf("expected valid, got: %v", err)
		}

		missingToken := ResetPasswordRequest{Token: "", NewPassword: "newpassword123"}
		if err := missingToken.Validate(); err == nil {
			t.Errorf("expected error for missing token, got nil")
		}
	})
}
