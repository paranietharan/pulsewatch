package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	RabbitMQURL   string
	PollInterval  time.Duration
	BatchSize     int
	LogLevel      string
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env", "../.env")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			getEnv("POSTGRES_USER", "pulsewatch"),
			getEnv("POSTGRES_PASSWORD", "pulsewatch_secret"),
			getEnv("POSTGRES_HOST", "localhost"),
			getEnv("POSTGRES_PORT", "5432"),
			getEnv("POSTGRES_DB", "pulsewatch"),
			getEnv("POSTGRES_SSLMODE", "disable"),
		)
	}

	rmqURL := getEnv("RABBITMQ_URL", "amqp://pulsewatch:pulsewatch_secret@localhost:5672/")

	pollSeconds, err := strconv.Atoi(getEnv("SCHEDULER_POLL_INTERVAL_SEC", "5"))
	if err != nil || pollSeconds < 1 {
		pollSeconds = 5
	}

	batchSize, err := strconv.Atoi(getEnv("SCHEDULER_BATCH_SIZE", "50"))
	if err != nil || batchSize < 1 {
		batchSize = 50
	}

	return &Config{
		DatabaseURL:  dbURL,
		RabbitMQURL:  rmqURL,
		PollInterval: time.Duration(pollSeconds) * time.Second,
		BatchSize:    batchSize,
		LogLevel:     getEnv("LOG_LEVEL", "debug"),
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
