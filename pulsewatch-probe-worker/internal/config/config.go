package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	RabbitMQURL         string
	WorkerCount         int
	PrefetchCount       int
	Region              string
	LogLevel            string
	AllowInternalProbes bool
}

func Load() (*Config, error) {
	_ = godotenv.Load(".env", "../.env")

	workerCount, err := strconv.Atoi(getEnv("PROBE_WORKER_CONCURRENCY", "20"))
	if err != nil || workerCount < 1 {
		workerCount = 20
	}

	prefetchCount, err := strconv.Atoi(getEnv("PROBE_WORKER_PREFETCH", "20"))
	if err != nil || prefetchCount < 1 {
		prefetchCount = 20
	}

	return &Config{
		RabbitMQURL:         getEnv("RABBITMQ_URL", "amqp://pulsewatch:pulsewatch_secret@localhost:5672/"),
		WorkerCount:         workerCount,
		PrefetchCount:       prefetchCount,
		Region:              getEnv("PROBE_REGION", "default"),
		LogLevel:            getEnv("LOG_LEVEL", "debug"),
		AllowInternalProbes: getEnv("ALLOW_INTERNAL_PROBES", "false") == "true",
	}, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
