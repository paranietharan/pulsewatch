package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"pulsewatch-api/internal/models"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrTokenExpired = errors.New("token has expired")
)

// JWTClaims represents the custom and standard JWT claims.
type JWTClaims struct {
	UserID uuid.UUID   `json:"sub"`
	Email  string      `json:"email"`
	Role   models.Role `json:"role"`
	jwt.RegisteredClaims
}

// TokenService manages access JWTs and cryptographic opaque tokens.
type TokenService struct {
	jwtSecret    []byte
	accessExpiry time.Duration
	issuer       string
	audience     string
}

func NewTokenService(jwtSecret string, accessExpiry time.Duration) *TokenService {
	return &TokenService{
		jwtSecret:    []byte(jwtSecret),
		accessExpiry: accessExpiry,
		issuer:       "pulsewatch-api",
		audience:     "pulsewatch-client",
	}
}

// GenerateAccessToken signs a short-lived JWT for authenticated requests.
func (s *TokenService) GenerateAccessToken(user *models.User) (string, error) {
	now := time.Now().UTC()
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessExpiry)),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}
	return signed, nil
}

// ValidateAccessToken verifies the JWT signature, algorithm, and claims.
func (s *TokenService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	}, jwt.WithIssuer(s.issuer), jwt.WithAudience(s.audience))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateOpaqueToken generates a cryptographically secure random token (64 hex characters = 32 bytes)
// and returns both the raw token (to give to client) and its SHA-256 hash (to store in DB).
func (s *TokenService) GenerateOpaqueToken() (rawToken string, tokenHash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random token: %w", err)
	}

	rawToken = hex.EncodeToString(bytes)
	tokenHash = s.HashToken(rawToken)
	return rawToken, tokenHash, nil
}

// HashToken computes the deterministic SHA-256 hex hash of a raw token.
func (s *TokenService) HashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

// ConstantTimeCompare compares two token hashes safely without timing leakage.
func (s *TokenService) ConstantTimeCompare(hashA, hashB string) bool {
	return subtle.ConstantTimeCompare([]byte(hashA), []byte(hashB)) == 1
}
