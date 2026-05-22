package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/videoforge/backend/svc-ai-support/internal/config"
	"github.com/videoforge/backend/svc-ai-support/internal/handler"
	"github.com/videoforge/backend/svc-ai-support/internal/service"

	"github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/pkg/natsclient"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	defer cfg.Close()

	slog.Info("starting AI support service", "port", cfg.Server.Port)

	// Initialize NATS client (optional - service can run without it)
	var natsClient natsclient.NATSClient
	natsCfg := natsclient.DefaultConfig()
	natsCfg.URL = cfg.NATS.URL
	natsClient = natsclient.New(natsCfg, nil)

	// Try to connect to NATS
	if err := natsClient.Connect(); err != nil {
		slog.Warn("NATS not available, running without event publishing", "error", err)
		natsClient = nil
	} else {
		slog.Info("NATS connected")
		defer natsClient.Close()
	}

	// Initialize repository and service
	supportService := service.NewSupportService(cfg.Database.Pool, natsClient)

	// Initialize middleware
	authMiddleware := handler.NewAuthMiddleware(cfg.JWT.PublicKey)

	// Setup router
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API routes
	api := http.NewServeMux()

	// Support handlers
	supportH := handler.NewSupportHandler(supportService)
	api.HandleFunc("POST /api/v1/support/chat", supportH.Chat)

	// Protected routes
	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/v1/support/conversations", supportH.ListConversations)
	protected.HandleFunc("GET /api/v1/support/conversations/", supportH.GetConversation)

	// Escalation endpoints
	escalateHandler := http.NewServeMux()
	escalateHandler.HandleFunc("POST /api/v1/support/conversations/", supportH.Escalate)
	escalateHandler.HandleFunc("GET /api/v1/support/escalations", supportH.ListEscalations)
	escalateHandler.HandleFunc("POST /api/v1/support/escalations/", supportH.ResolveEscalation)

	// Wrap with auth middleware - use http.Handler directly
	protected = wrapHandler(authMiddleware.Authenticate(protected))
	escalateHandler = wrapHandler(authMiddleware.Authenticate(escalateHandler))

	api.Handle("/api/v1/support/conversations", protected)
	api.Handle("/api/v1/support/conversations/", escalateHandler)
	api.Handle("/api/v1/support/escalations", escalateHandler)
	api.Handle("/api/v1/support/escalations/", escalateHandler)

	// Apply global middleware using Chain
	handler := middleware.Chain(api,
		middleware.RequestID,
		middleware.Recover,
		middleware.Logger,
	)

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
		slog.Info("server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

// wrapHandler wraps an http.Handler to return *http.ServeMux
func wrapHandler(h http.Handler) *http.ServeMux {
	m := http.NewServeMux()
	m.Handle("/", h)
	return m
}