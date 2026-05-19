package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/videoforge/backend/svc-user/internal/config"
	"github.com/videoforge/backend/svc-user/internal/handler"
	"github.com/videoforge/backend/svc-user/internal/repository"
	"github.com/videoforge/backend/svc-user/internal/service"

	"github.com/videoforge/backend/pkg/logger"
	backendmiddleware "github.com/videoforge/backend/pkg/middleware"
)

// @title VideoForge User Service API
// @version 1.0
// @description User authentication and management service for VideoForge
// @termsOfService https://videoforge.io/terms

// @contact.name VideoForge Support
// @contact.url https://videoforge.io/support
// @contact.email support@videoforge.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	defer cfg.Close()

	logger.Info("starting user service",
		"port", cfg.Server.Port,
	)

	// Initialize repository
	userRepo := repository.NewUserRepository(cfg.Database.Pool)

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWT.PrivateKey, cfg.JWT.KeyID)
	userService := service.NewUserService(userRepo, cfg.JWT.PrivateKey, cfg.JWT.KeyID)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

	// Initialize Neon Auth handler (optional, falls back if not configured)
	var neonAuthHandler *handler.NeonAuthHandler
	if cfg.NeonAuth != nil && cfg.NeonAuth.IsConfigured() {
		neonAuthHandler = handler.NewNeonAuthHandler(cfg.NeonAuth, authService)
		logger.Info("neon auth enabled")
	}

	// Initialize middleware
	authMiddleware := backendmiddleware.NewAuthMiddleware(cfg.JWT.PublicKey)
	jwksHandler := backendmiddleware.NewJWKSHandler(cfg.JWT.PublicKey, cfg.JWT.KeyID)

	// Setup router
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// JWKS endpoint
	mux.Handle("/.well-known/jwks.json", jwksHandler)

	// API routes
	api := http.NewServeMux()
	api.HandleFunc("/api/v1/auth/register", authHandler.Register)
	api.HandleFunc("/api/v1/auth/login", authHandler.Login)
	api.HandleFunc("/api/v1/auth/refresh", authHandler.Refresh)
	api.HandleFunc("/api/v1/auth/logout", authHandler.Logout)

	// Neon Auth routes (optional)
	if neonAuthHandler != nil {
		api.HandleFunc("/api/v1/auth/neon/register", neonAuthHandler.Register)
		api.HandleFunc("/api/v1/auth/neon/login", neonAuthHandler.Login)
		api.HandleFunc("/api/v1/auth/neon/session", neonAuthHandler.GetSession)
		api.HandleFunc("/api/v1/auth/neon/logout", neonAuthHandler.Logout)
		api.HandleFunc("/api/v1/auth/neon/refresh", neonAuthHandler.Refresh)
	}

	// Protected routes
	protected := http.NewServeMux()
	protected.HandleFunc("/api/v1/users/me", userHandler.GetMe)
	protected.HandleFunc("/api/v1/users/me", userHandler.UpdateMe)

	// Wrap with auth middleware
	protected = authMiddleware.Authenticate(protected)
	api.Handle("/api/v1/auth/", protected)

	// Apply global middleware
	handler := backendmiddleware.Chain(api, backendmiddleware.Recover, backendmiddleware.Logger)

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		logger.Info("server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}