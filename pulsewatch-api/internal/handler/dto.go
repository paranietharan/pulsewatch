package handler

import (
	"errors"
	"net/mail"
	"strings"
)

type SignupRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func (r *SignupRequest) Validate() error {
	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))

	if r.FirstName == "" {
		return errors.New("first_name is required")
	}
	if r.Email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return errors.New("invalid email address format")
	}
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if len(r.Password) > 72 { // bcrypt limit is 72 bytes
		return errors.New("password must not exceed 72 characters")
	}
	return nil
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() error {
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

func (r *ForgotPasswordRequest) Validate() error {
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	if r.Email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return errors.New("invalid email address format")
	}
	return nil
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
	Password    string `json:"password,omitempty"`
}

func (r *ResetPasswordRequest) Validate() error {
	r.Token = strings.TrimSpace(r.Token)
	if r.Token == "" {
		return errors.New("reset token is required")
	}
	if r.NewPassword == "" && r.Password != "" {
		r.NewPassword = r.Password
	}
	if len(r.NewPassword) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	if len(r.NewPassword) > 72 {
		return errors.New("password must not exceed 72 characters")
	}
	return nil
}

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

func (r *VerifyEmailRequest) Validate() error {
	r.Token = strings.TrimSpace(r.Token)
	if r.Token == "" {
		return errors.New("verification token is required")
	}
	return nil
}

type ResendVerificationRequest struct {
	Email string `json:"email"`
}

func (r *ResendVerificationRequest) Validate() error {
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	if r.Email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return errors.New("invalid email address format")
	}
	return nil
}
