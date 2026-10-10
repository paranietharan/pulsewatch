package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pulsewatch-api/internal/config"
	"pulsewatch-api/internal/database"
	"pulsewatch-api/internal/handler"
	"pulsewatch-api/internal/middleware"
	"pulsewatch-api/internal/outbox"
	"pulsewatch-api/internal/processor"
	"pulsewatch-api/internal/rabbitmq"
	"pulsewatch-api/internal/repository"
	"pulsewatch-api/internal/service"

	"github.com/go-chi/chi/v5"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	logger.Info("Starting PulseWatch API Service...")

	// Load Configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Connect to PostgreSQL
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("Connected to PostgreSQL successfully")

	// Connect to RabbitMQ
	rmqClient, err := rabbitmq.Connect(cfg.RabbitMQURL)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer rmqClient.Close()
	logger.Info("Connected to RabbitMQ with publisher confirms enabled")

	// Initialize Repositories
	userRepo := repository.NewUserRepository(db.Pool)
	tokenRepo := repository.NewTokenRepository(db.Pool)
	outboxRepo := repository.NewOutboxRepository(db.Pool)
	monitorRepo := repository.NewMonitorRepository(db.Pool)

	// Initialize Services
	pwdService := service.NewPasswordService()
	tokenService := service.NewTokenService(cfg.JWTSecret, cfg.JWTAccessExpiry)
	authService := service.NewAuthService(
		cfg,
		db,
		userRepo,
		tokenRepo,
		outboxRepo,
		tokenService,
		pwdService,
	)

	// Start Transactional Outbox Dispatcher in background
	outboxCtx, outboxCancel := context.WithCancel(context.Background())
	defer outboxCancel()
	dispatcher := outbox.NewDispatcher(outboxRepo, rmqClient, logger)
	go dispatcher.Start(outboxCtx)

	// Start Check-Result and Outcome Processor in background
	resultProcessor := processor.NewResultProcessor(db.Pool, rmqClient, logger)
	go func() {
		if err := resultProcessor.Start(outboxCtx); err != nil {
			logger.Error("Result processor consumer stopped", "error", err)
		}
	}()

	// Setup Router and Middlewares
	authHandler := handler.NewAuthHandler(authService, cfg)
	monitorHandler := handler.NewMonitorHandler(monitorRepo)
	strictRateLimiter := middleware.NewRateLimiter(0.2, 5) // 1 token every 5 sec, burst capacity 5

	r := chi.NewRouter()

	r.Use(middleware.SetupClientIP(cfg))
	r.Use(middleware.Recoverer(logger))
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.CORS(cfg.FrontendURL))

	// Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		handler.JSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"service": "pulsewatch-api",
		})
	})

	r.Route("/api/v1", func(r chi.Router) {
		// Authentication routes
		r.Route("/auth", func(r chi.Router) {
			// Rate-limited sensitive public endpoints
			r.With(strictRateLimiter.Middleware()).Post("/signup", authHandler.Signup)
			r.With(strictRateLimiter.Middleware()).Post("/login", authHandler.Login)
			r.With(strictRateLimiter.Middleware()).Post("/forgot-password", authHandler.ForgotPassword)
			r.With(strictRateLimiter.Middleware()).Post("/reset-password", authHandler.ResetPassword)
			r.With(strictRateLimiter.Middleware()).Post("/resend-verification", authHandler.ResendVerification)

			// Standard public auth endpoints
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
			r.Post("/verify-email", authHandler.VerifyEmail)

			// Protected auth endpoints
			r.Group(func(r chi.Router) {
				r.Use(middleware.AuthMiddleware(tokenService))
				r.Get("/me", authHandler.Me)
			})
		})

		// Protected endpoints
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(tokenService))

			// Monitors
			r.Post("/monitors", monitorHandler.CreateMonitor)
			r.Get("/monitors", monitorHandler.ListMonitors)
			r.Get("/monitors/{monitorID}", monitorHandler.GetMonitor)
			r.Patch("/monitors/{monitorID}", monitorHandler.UpdateMonitor)
			r.Delete("/monitors/{monitorID}", monitorHandler.DeleteMonitor)
			r.Post("/monitors/{monitorID}/enable", monitorHandler.EnableMonitor)
			r.Post("/monitors/{monitorID}/disable", monitorHandler.DisableMonitor)
			r.Get("/monitors/{monitorID}/results", monitorHandler.GetResults)
			r.Get("/monitors/{monitorID}/incidents", monitorHandler.GetIncidents)

			// Dashboard Summary
			r.Get("/dashboard/summary", monitorHandler.GetDashboardSummary)

			// Alert Channels
			r.Get("/alert-channels", monitorHandler.ListAlertChannels)
			r.Post("/alert-channels", monitorHandler.CreateAlertChannel)
			r.Delete("/alert-channels/{channelID}", monitorHandler.DeleteAlertChannel)
		})
	})

	serverAddr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info(fmt.Sprintf("HTTP Server listening on http://%s", serverAddr))
		serverErrors <- server.ListenAndServe()
	}()

	// Graceful Shutdown on SIGINT / SIGTERM
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP Server failed", "error", err)
			os.Exit(1)
		}
	case sig := <-shutdown:
		logger.Info("Shutdown signal received, draining connections...", "signal", sig.String())
		outboxCancel()

		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelShutdown()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Could not stop server gracefully", "error", err)
			_ = server.Close()
		}
		logger.Info("PulseWatch API Service shut down gracefully")
	}
}
