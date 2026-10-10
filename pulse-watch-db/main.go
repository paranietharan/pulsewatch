package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
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

	migrationsDir := findMigrationsDir()
	m, err := migrate.New("file://"+migrationsDir, dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize migration driver: %v", err)
	}
	defer m.Close()

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "up":
		fmt.Println("Applying pending migrations...")
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("Database is already up to date. No pending migrations.")
				return
			}
			log.Fatalf("Migration UP failed: %v", err)
		}
		fmt.Println("All migrations applied successfully!")

	case "down":
		steps := 1
		if len(os.Args) > 2 {
			if s, err := strconv.Atoi(os.Args[2]); err == nil && s > 0 {
				steps = s
			}
		}
		fmt.Printf("Rolling back %d migration(s)...\n", steps)
		if err := m.Steps(-steps); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("No migrations to rollback.")
				return
			}
			log.Fatalf("Migration DOWN failed: %v", err)
		}
		fmt.Printf("Rolled back %d migration(s) successfully!\n", steps)

	case "status", "version":
		version, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				fmt.Println("Status: No migrations applied yet (version 0).")
				return
			}
			log.Fatalf("Failed to query status: %v", err)
		}
		fmt.Printf("Status: Active migration version: %d | Dirty state: %t\n", version, dirty)

	case "force":
		if len(os.Args) < 3 {
			log.Fatal("Usage: go run main.go force <version_number>")
		}
		targetVersion, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("Invalid version number: %s", os.Args[2])
		}
		if err := m.Force(targetVersion); err != nil {
			log.Fatalf("Force failed: %v", err)
		}
		fmt.Printf("Forced migration version to %d.\n", targetVersion)

	default:
		fmt.Printf("Unknown command: %s\nAvailable commands: up, down [steps], status, force <version>\n", command)
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func findMigrationsDir() string {
	candidates := []string{
		"migrations",
		"pulse-watch-db/migrations",
		"../migrations",
	}
	for _, c := range candidates {
		if stat, err := os.Stat(c); err == nil && stat.IsDir() {
			abs, err := filepath.Abs(c)
			if err == nil {
				return abs
			}
		}
	}
	abs, _ := filepath.Abs("migrations")
	return abs
}
