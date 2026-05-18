package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"svc-ai-support/internal/config"
	"svc-ai-support/internal/handler"
	"svc-ai-support/internal/service"

	"github.com/videoforge/backend/pkg/logger"
	backendmiddleware "github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/pkg/natsclient"
)

// @title VideoForge AI Support Service API
// @version 1.0
// @description AI-powered conversational support service for VideoForge
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

	logger.Info("starting AI support service",
		"port", cfg.Server.Port,
	)

	// Initialize NATS client (optional - service can run without it)
	var natsClient natsclient.NATSClient
	natsCfg := natsclient.DefaultConfig()
	natsCfg.URL = cfg.NATS.URL
	natsClient = natsclient.New(natsCfg, logger.Default("development"))

	// Try to connect to NATS
	if err := natsClient.Connect(); err != nil {
		logger.Warn("NATS not available, running without event publishing",
			"error", err,
		)
		natsClient = nil
	} else {
		logger.Info("NATS connected")
		defer natsClient.Close()
	}

	// Initialize repository and service
	supportService := service.NewSupportService(cfg.Database.Pool, natsClient)

	// Initialize middleware
	recoverMw := backendmiddleware.Recover()
	loggerMw := backendmiddleware.Logger()
	authMiddleware := handler.NewAuthMiddleware(cfg.JWT.PublicKey)

	// Setup router
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API routes
	api := http.NewServeMux()

	// Support handlers
	supportH := handler.NewSupportHandler(supportService)
	api.HandleFunc("/api/v1/support/chat", supportH.Chat)

	// Protected routes
	protected := http.NewServeMux()
	protected.HandleFunc("/api/v1/support/conversations", supportH.ListConversations)
	protected.HandleFunc("/api/v1/support/conversations/", supportH.GetConversation)

	// Escalation endpoints
	escalateHandler := http.NewServeMux()
	escalateHandler.HandleFunc("/api/v1/support/conversations/", supportH.Escalate)
	escalateHandler.HandleFunc("/api/v1/support/escalations", supportH.ListEscalations)
	escalateHandler.HandleFunc("/api/v1/support/escalations/", supportH.ResolveEscalation)

	// Wrap with auth middleware
	protected = authMiddleware.Authenticate(protected)
	escalateHandler = authMiddleware.Authenticate(escalateHandler)

	api.Handle("/api/v1/support/conversations", protected)
	api.Handle("/api/v1/support/conversations/", escalateHandler)
	api.Handle("/api/v1/support/escalations", escalateHandler)
	api.Handle("/api/v1/support/escalations/", escalateHandler)

	// Apply global middleware
	handler := loggerMw(recoverMw(api))

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