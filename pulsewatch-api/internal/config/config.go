package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration parameters for the pulsewatch-api service.
type Config struct {
	Port                   int
	Host                   string
	DatabaseURL            string
	RabbitMQURL            string
	JWTSecret              string
	JWTAccessExpiry        time.Duration
	RefreshTokenExpiry     time.Duration
	VerifyTokenExpiry      time.Duration
	ResetTokenExpiry       time.Duration
	Environment            string
	LogLevel               string
	FrontendURL            string
	ClientIPStrategy       string   // "remote_addr", "cloudflare", "header", "xff", "proxy_hops"
	ClientIPHeader         string   // Header name when strategy is "header" or "cloudflare"
	ClientIPTrustedCIDRs   []string // Trusted CIDR ranges for "xff" strategy
	ClientIPTrustedProxies int      // Number of trusted proxy hops for "proxy_hops" strategy
}

// Load loads configuration from .env files and active environment variables.
func Load() (*Config, error) {
	// Attempt to load .env from current directory or parent directories
	_ = godotenv.Load(".env", "../.env")

	port, err := strconv.Atoi(getEnv("API_PORT", "8080"))
	if err != nil {
		port = 8080
	}

	jwtSecret := getEnv("JWT_SECRET", "super_secret_jwt_key_change_in_production")
	if len(jwtSecret) < 16 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 16 characters long for security")
	}

	databaseURL := getEnv("DATABASE_URL", "")
	if databaseURL == "" {
		databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			getEnv("POSTGRES_USER", "pulsewatch"),
			getEnv("POSTGRES_PASSWORD", "pulsewatch_secret"),
			getEnv("POSTGRES_HOST", "localhost"),
			getEnv("POSTGRES_PORT", "5432"),
			getEnv("POSTGRES_DB", "pulsewatch"),
			getEnv("POSTGRES_SSLMODE", "disable"),
		)
	}

	rabbitmqURL := getEnv("RABBITMQ_URL", "amqp://pulsewatch:pulsewatch_secret@localhost:5672/")

	clientIPStrategy := getEnv("CLIENT_IP_STRATEGY", "remote_addr")
	clientIPHeader := getEnv("CLIENT_IP_HEADER", "")
	if clientIPStrategy == "cloudflare" && clientIPHeader == "" {
		clientIPHeader = "CF-Connecting-IP"
	}

	var trustedCIDRs []string
	if cidrsStr := getEnv("CLIENT_IP_TRUSTED_CIDRS", ""); cidrsStr != "" {
		for _, c := range strings.Split(cidrsStr, ",") {
			trimmed := strings.TrimSpace(c)
			if trimmed != "" {
				trustedCIDRs = append(trustedCIDRs, trimmed)
			}
		}
	}

	trustedProxies, err := strconv.Atoi(getEnv("CLIENT_IP_TRUSTED_PROXIES", "1"))
	if err != nil || trustedProxies < 1 {
		trustedProxies = 1
	}

	return &Config{
		Port:                   port,
		Host:                   getEnv("API_HOST", "0.0.0.0"),
		DatabaseURL:            databaseURL,
		RabbitMQURL:            rabbitmqURL,
		JWTSecret:              jwtSecret,
		JWTAccessExpiry:        15 * time.Minute,
		RefreshTokenExpiry:     7 * 24 * time.Hour,
		VerifyTokenExpiry:      24 * time.Hour,
		ResetTokenExpiry:       1 * time.Hour,
		Environment:            getEnv("ENVIRONMENT", "development"),
		LogLevel:               getEnv("LOG_LEVEL", "debug"),
		FrontendURL:            getEnv("FRONTEND_URL", "http://localhost:3000"),
		ClientIPStrategy:       clientIPStrategy,
		ClientIPHeader:         clientIPHeader,
		ClientIPTrustedCIDRs:   trustedCIDRs,
		ClientIPTrustedProxies: trustedProxies,
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
