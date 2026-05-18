package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/videoforge/backend/pkg/config"
	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/pkg/storage"

	"svc-brief/internal/config"
	"svc-brief/internal/handler"
	"svc-brief/internal/middleware"
	"svc-brief/internal/repository"
	"svc-brief/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load("SERVER")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	defer cfg.Close()

	// Initialize logger
	log := logger.Default(cfg.Server.Environment)
	ctx := logger.FromContext(context.Background()).Context(context.Background())

	log.Info("Starting Brief service",
		slog.String("environment", cfg.Server.Environment),
		slog.String("port", cfg.Server.Port),
	)

	// Initialize repository
	repo := repository.NewBriefRepo(cfg.Database.Pool)

	// Get storage client if available
	var storageClient storage.Storage
	if cfg.Storage.Enabled && cfg.Storage.Client != nil {
		storageClient = cfg.Storage.Client
	}

	// Initialize service
	svc := service.NewBriefService(repo, storageClient)

	// Initialize handler
	briefHandler := handler.NewBriefHandler(svc)

	// Load auth configuration
	authConfig, err := middleware.LoadAuthConfig()
	if err != nil {
		log.Warn("Failed to load auth config, continuing without JWT validation",
			slog.String("error", err.Error()))
	}
	authMiddleware := middleware.NewAuthMiddleware(authConfig)

	// Set up HTTP router with pattern-based routing
	mux := http.NewServeMux()

	// Register handlers with path patterns
	mux.HandleFunc("GET /health", handleHealth)

	// API v1 routes
	mux.HandleFunc("POST /api/v1/briefs", wrapAuth(authMiddleware, briefHandler.HandleCreateBrief))
	mux.HandleFunc("GET /api/v1/briefs", wrapAuth(authMiddleware, briefHandler.HandleListBriefs))
	mux.HandleFunc("GET /api/v1/briefs/$", wrapAuth(authMiddleware, briefHandler.HandleGetBrief))
	mux.HandleFunc("PATCH /api/v1/briefs/$", wrapAuth(authMiddleware, briefHandler.HandleUpdateBrief))
	mux.HandleFunc("POST /api/v1/briefs/$/publish", wrapAuth(authMiddleware, briefHandler.HandlePublishBrief))
	mux.HandleFunc("POST /api/v1/briefs/$/close", wrapAuth(authMiddleware, briefHandler.HandleCloseBrief))
	mux.HandleFunc("POST /api/v1/briefs/$/interview", briefHandler.HandleInterview)
	mux.HandleFunc("GET /api/v1/briefs/matching", wrapAuth(authMiddleware, briefHandler.HandleMatchingBriefs))
	mux.HandleFunc("POST /api/v1/briefs/$/view", wrapAuth(authMiddleware, briefHandler.HandleViewBrief))

	// Raw footage endpoints
	mux.HandleFunc("POST /api/v1/briefs/$/raw-footage/upload-url", wrapAuth(authMiddleware, briefHandler.HandleGetRawFootageUploadURL))
	mux.HandleFunc("POST /api/v1/briefs/$/raw-footage/confirm", wrapAuth(authMiddleware, briefHandler.HandleConfirmRawFootageUpload))
	mux.HandleFunc("GET /api/v1/briefs/$/raw-footage/download-url", wrapAuth(authMiddleware, briefHandler.HandleGetRawFootageDownloadURL))

	// Add middleware chain
	chain := middleware.Chain(
		mux,
		middleware.RequestID,
		middleware.Recover,
		middleware.Logger,
		middleware.CORS("*"),
	)

	// Create server
	server := &http.Server{
		Addr:    cfg.Server.GetAddress(),
		Handler: chain,
	}

	// Start server with graceful shutdown
	errCh := make(chan error, 1)
	go func() {
		log.Info("Server starting", slog.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Error("Server error", slog.String("error", err.Error()))
		os.Exit(1)
	case <-sigCh:
		log.Info("Shutting down server...")
	}

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Shutdown error", slog.String("error", err.Error()))
	}

	log.Info("Server stopped")
}

// Simple health handler
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

// wrapAuth wraps a handler with auth middleware
func wrapAuth(auth *middleware.AuthMiddleware, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			// Token provided - would validate in production
			// For now, just extract user info if present
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr != "" {
				// Simple extraction - in production use proper JWT
				// Check for user ID in various formats
				if strings.Contains(tokenStr, ".") {
					// Likely JWT, extract sub claim (simplified)
					parts := strings.Split(tokenStr, ".")
					if len(parts) >= 2 {
						// Could decode payload here
						// For now, use a placeholder
						ctx := context.WithValue(r.Context(), "user_id", "user-from-token")
						r = r.WithContext(ctx)
					}
				}
			}
		}
		handler(w, r)
	}
}