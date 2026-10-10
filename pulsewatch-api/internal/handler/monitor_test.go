package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"pulsewatch-api/internal/database"
	"pulsewatch-api/internal/handler"
	"pulsewatch-api/internal/middleware"
	"pulsewatch-api/internal/models"
	"pulsewatch-api/internal/repository"
	"pulsewatch-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestMonitorDTO_Validation(t *testing.T) {
	t.Run("Valid CreateMonitorRequest", func(t *testing.T) {
		req := handler.CreateMonitorRequest{
			Name: "Production Web App",
			URL:  "https://example.com/health",
		}
		if err := req.Validate(); err != nil {
			t.Fatalf("expected valid, got: %v", err)
		}
		if req.Method != "GET" {
			t.Errorf("expected default method GET, got %s", req.Method)
		}
		if req.IntervalSeconds != 60 {
			t.Errorf("expected default interval 60, got %d", req.IntervalSeconds)
		}
		if req.TimeoutSeconds != 10 {
			t.Errorf("expected default timeout 10, got %d", req.TimeoutSeconds)
		}
		if req.FailureThreshold != 3 {
			t.Errorf("expected default failure threshold 3, got %d", req.FailureThreshold)
		}
		if req.RecoveryThreshold != 2 {
			t.Errorf("expected default recovery threshold 2, got %d", req.RecoveryThreshold)
		}
	})

	t.Run("Missing Name", func(t *testing.T) {
		req := handler.CreateMonitorRequest{
			Name: "   ",
			URL:  "https://example.com",
		}
		if err := req.Validate(); err == nil {
			t.Fatal("expected error for empty name, got nil")
		}
	})

	t.Run("Invalid URL Scheme", func(t *testing.T) {
		req := handler.CreateMonitorRequest{
			Name: "Bad Scheme",
			URL:  "ftp://example.com",
		}
		if err := req.Validate(); err == nil {
			t.Fatal("expected error for non-http URL, got nil")
		}
	})

	t.Run("Timeout greater than interval", func(t *testing.T) {
		req := handler.CreateMonitorRequest{
			Name:            "Timeout too high",
			URL:             "https://example.com",
			IntervalSeconds: 30,
			TimeoutSeconds:  35,
		}
		if err := req.Validate(); err == nil {
			t.Fatal("expected error when timeout > interval, got nil")
		}
	})

	t.Run("Valid UpdateMonitorRequest", func(t *testing.T) {
		name := "Updated Name"
		req := handler.UpdateMonitorRequest{
			Name: &name,
		}
		if err := req.Validate(); err != nil {
			t.Fatalf("expected valid update, got: %v", err)
		}
	})

	t.Run("Valid CreateAlertChannelRequest", func(t *testing.T) {
		req := handler.CreateAlertChannelRequest{
			Type:   "webhook",
			Name:   "Slack Alerts",
			Config: json.RawMessage(`{"url":"https://hooks.slack.com/services/123"}`),
		}
		if err := req.Validate(); err != nil {
			t.Fatalf("expected valid alert channel, got: %v", err)
		}
		if req.Type != "WEBHOOK" {
			t.Errorf("expected uppercase WEBHOOK, got %s", req.Type)
		}
	})
}

func setupMonitorTestApp(t *testing.T) (*chi.Mux, *database.DB, *service.TokenService, uuid.UUID, uuid.UUID) {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://pulsewatch:pulsewatch_secret@localhost:5432/pulsewatch?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, dbURL)
	if err != nil {
		t.Skipf("Skipping integration test: PostgreSQL is not accessible (%v)", err)
	}

	tokenService := service.NewTokenService("test_monitor_jwt_secret_1234567890", 1*time.Hour)
	monitorRepo := repository.NewMonitorRepository(db.Pool)
	monitorHandler := handler.NewMonitorHandler(monitorRepo)

	// Create two distinct users in db for ownership tests
	userA_ID := uuid.New()
	userB_ID := uuid.New()

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, first_name, last_name, email, password_hash, role, is_active, is_verified, created_at, updated_at)
		VALUES ($1, 'User', 'A', $2, 'hash', 'USER'::user_role, true, true, NOW(), NOW()),
		       ($3, 'User', 'B', $4, 'hash', 'USER'::user_role, true, true, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING;
	`, userA_ID, fmt.Sprintf("usera_%s@example.com", userA_ID.String()[:8]),
		userB_ID, fmt.Sprintf("userb_%s@example.com", userB_ID.String()[:8]))
	if err != nil {
		t.Fatalf("failed to insert test users: %v", err)
	}

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(tokenService))

			r.Post("/monitors", monitorHandler.CreateMonitor)
			r.Get("/monitors", monitorHandler.ListMonitors)
			r.Get("/monitors/{monitorID}", monitorHandler.GetMonitor)
			r.Patch("/monitors/{monitorID}", monitorHandler.UpdateMonitor)
			r.Delete("/monitors/{monitorID}", monitorHandler.DeleteMonitor)
			r.Post("/monitors/{monitorID}/enable", monitorHandler.EnableMonitor)
			r.Post("/monitors/{monitorID}/disable", monitorHandler.DisableMonitor)
			r.Get("/monitors/{monitorID}/results", monitorHandler.GetResults)
			r.Get("/monitors/{monitorID}/incidents", monitorHandler.GetIncidents)
			r.Get("/dashboard/summary", monitorHandler.GetDashboardSummary)
		})
	})

	return r, db, tokenService, userA_ID, userB_ID
}

func TestMonitorFlow_AndAuthorization(t *testing.T) {
	router, _, tokenService, userA, userB := setupMonitorTestApp(t)

	tokenA, err := tokenService.GenerateAccessToken(&models.User{
		ID:    userA,
		Email: "usera@example.com",
		Role:  models.RoleUser,
	})
	if err != nil {
		t.Fatalf("failed to generate token A: %v", err)
	}
	tokenB, err := tokenService.GenerateAccessToken(&models.User{
		ID:    userB,
		Email: "userb@example.com",
		Role:  models.RoleUser,
	})
	if err != nil {
		t.Fatalf("failed to generate token B: %v", err)
	}

	// User A creates a monitor
	createBody := map[string]interface{}{
		"name":                "User A Site",
		"url":                 "https://example.com",
		"method":              "GET",
		"interval_seconds":    60,
		"timeout_seconds":     10,
		"failure_threshold":   3,
		"recovery_threshold":  2,
	}
	payload, _ := json.Marshal(createBody)

	req := httptest.NewRequest("POST", "/api/v1/monitors", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	var created models.Monitor
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal created monitor: %v", err)
	}
	if created.ID == uuid.Nil {
		t.Fatal("expected non-nil monitor ID")
	}

	// User A retrieves the monitor -> 200 OK
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/monitors/%s", created.ID), nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for User A, got %d", w.Code)
	}

	// User B tries to retrieve User A's monitor -> 404 Not Found (Cross-user isolation)
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/monitors/%s", created.ID), nil)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found for cross-user access, got %d", w.Code)
	}

	// User B tries to update User A's monitor -> 404 Not Found
	updateBody := map[string]interface{}{
		"name": "Hacked Name",
	}
	upPayload, _ := json.Marshal(updateBody)
	req = httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/monitors/%s", created.ID), bytes.NewReader(upPayload))
	req.Header.Set("Authorization", "Bearer "+tokenB)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found when User B updates User A's monitor, got %d", w.Code)
	}

	// User A updates their own monitor -> 200 OK
	newName := "Updated User A Site"
	updateBody["name"] = newName
	upPayload, _ = json.Marshal(updateBody)
	req = httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/monitors/%s", created.ID), bytes.NewReader(upPayload))
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for monitor update, got %d: %s", w.Code, w.Body.String())
	}

	// User A disables monitor -> 200 OK
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/monitors/%s/disable", created.ID), nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK disabling monitor, got %d", w.Code)
	}

	// User A enables monitor -> 200 OK
	req = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/monitors/%s/enable", created.ID), nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK enabling monitor, got %d", w.Code)
	}

	// User A gets dashboard summary -> 200 OK
	req = httptest.NewRequest("GET", "/api/v1/dashboard/summary", nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK getting dashboard summary, got %d", w.Code)
	}

	// User A deletes monitor -> 200 OK
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/monitors/%s", created.ID), nil)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK deleting monitor, got %d", w.Code)
	}
}
