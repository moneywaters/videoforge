package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/videoforge/backend/pkg/config"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/pkg/natsclient"
	"github.com/videoforge/backend/svc-notification/internal/handler"
	"github.com/videoforge/backend/svc-notification/internal/service"
)

func main() {
	cfg, err := config.Load("SERVER")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.Default(cfg.Environment)
	ctx := logger.FromContext(context.Background()).Context(context.Background())

	// Create database pool
	dbURL := cfg.DatabaseURL
	if dbURL == "" {
		dbURL = "postgres://videoforge:password@localhost:5432/videoforge?sslmode=disable"
	}

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Error("Failed to parse database URL", slog.String("error", err.Error()))
		os.Exit(1)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Error("Failed to create database pool", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// Test database connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Error("Failed to ping database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("Database connected")

	// Create NATS client
	nc := natsclient.New(nil, log)
	if err := nc.ConnectWithRetry(5, 2*time.Second); err != nil {
		log.Error("Failed to connect to NATS", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer nc.Close()

	// Create notification service with raw client (pointer to the implementation)
	// Note: We pass the underlying *natsclient.Client which implements NATSClient interface
	notificationService := service.NewNotificationService(pool, nc, log)

	// Start event consumer
	if err := notificationService.StartEventConsumer(ctx); err != nil {
		log.Error("Failed to start event consumer", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("Event consumer started")

	// Create handlers
	notificationHandler := handler.NewNotificationHandler(notificationService)
	wsHandler := handler.NewWebSocketHandler(pool, notificationService.GetConnectionManager(), log)
	healthHandler := handler.NewHealthHandler()

	// Create router
	r := chi.NewRouter()

	// Apply global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.Recover)
	r.Use(middleware.Logger)
	r.Use(middleware.CORS("*"))

	// Health check
	r.Get("/health", healthHandler.HandleHealth)

	// WebSocket endpoint
	r.Get("/ws", wsHandler.HandleWebSocket)

	// API v1 routes with authentication
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware(cfg.JWTSecret, log))

		r.Route("/api/v1/notifications", func(r chi.Router) {
			r.Get("/", notificationHandler.ListNotifications)
			r.Post("/read-all", notificationHandler.MarkAllAsRead)

			r.Route("/preferences", func(r chi.Router) {
				r.Get("/", notificationHandler.GetPreferences)
				r.Put("/", notificationHandler.UpdatePreferences)
			})

			r.Route("/{id}", func(r chi.Router) {
				r.Post("/read", notificationHandler.MarkAsRead)
				r.Delete("/", notificationHandler.DeleteNotification)
			})
		})
	})

	log.Info("Starting Notification service",
		slog.String("environment", cfg.Environment),
		slog.String("port", cfg.Port),
	)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: r,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("Server starting", slog.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Error("Server error", slog.String("error", err.Error()))
		os.Exit(1)
	case <-sigCh:
		log.Info("Shutting down server...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Shutdown error", slog.String("error", err.Error()))
	}

	log.Info("Server stopped")
}

// authMiddleware creates authentication middleware
func authMiddleware(jwtSecret string, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from header first
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// Try query param for WebSocket fallback
				authHeader = r.URL.Query().Get("token")
				if authHeader == "" {
					http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
					return
				}
			} else {
				// Extract Bearer token
				authHeader = strings.TrimPrefix(authHeader, "Bearer ")
			}

			// For now, accept any valid JWT format
			// In production, properly validate JWT with jwtSecret
			parts := strings.Split(authHeader, ".")
			if len(parts) != 3 {
				http.Error(w, "Invalid token format", http.StatusUnauthorized)
				return
			}

			// Decode payload to get user_id
			// Note: In production, use proper JWT library to validate signature
			userID := extractUserIDFromJWT(parts[1])
			if userID == "" {
				http.Error(w, "Invalid token payload", http.StatusUnauthorized)
				return
			}

			// Add userID to context
			ctx := context.WithValue(r.Context(), "user_id", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractUserIDFromJWT extracts user ID from JWT payload
func extractUserIDFromJWT(payload string) string {
	// Replace base64url characters with standard base64
	output := strings.ReplaceAll(payload, "-", "+")
	output = strings.ReplaceAll(output, "_", "/")

	// Add padding
	switch len(output) % 4 {
	case 2:
		output += "=="
	case 3:
		output += "="
	}

	// Decode from base64
	decoded, err := base64.StdEncoding.DecodeString(output)
	if err != nil {
		return ""
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return ""
	}

	// Try sub claim first
	if sub, ok := claims["sub"].(string); ok {
		return sub
	}

	// Try user_id
	if uid, ok := claims["user_id"].(string); ok {
		return uid
	}

	return ""
}