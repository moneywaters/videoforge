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

	"github.com/videoforge/backend/pkg/logger"
	pkgmiddleware "github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/svc-brief/internal/handler"
	briefmiddleware "github.com/videoforge/backend/svc-brief/internal/middleware"
	"github.com/videoforge/backend/svc-brief/internal/repository"
	"github.com/videoforge/backend/svc-brief/internal/service"

	briefconfig "github.com/videoforge/backend/svc-brief/internal/config"
)

func main() {
	// Load configuration
	cfg, err := briefconfig.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	defer cfg.Close()

	// Initialize logger
	log := logger.Default(cfg.Server.Environment)

	log.Info("Starting Brief service",
		slog.String("environment", cfg.Server.Environment),
		slog.String("port", cfg.Server.Port),
	)

	// Initialize repository
	repo := repository.NewBriefRepo(cfg.Database.Pool)

	// Initialize service
	svc := service.NewBriefService(repo, cfg.Storage.Client)

	// Initialize handler
	briefHandler := handler.NewBriefHandler(svc)

	// Load auth configuration
	authConfig, _ := briefmiddleware.LoadAuthConfig()
	authMiddleware := briefmiddleware.NewAuthMiddleware(authConfig)

	// Set up HTTP router
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /api/v1/briefs", wrapAuth(authMiddleware, briefHandler.HandleCreateBrief))
	mux.HandleFunc("GET /api/v1/briefs", wrapAuth(authMiddleware, briefHandler.HandleListBriefs))
	mux.HandleFunc("GET /api/v1/briefs/$", wrapAuth(authMiddleware, briefHandler.HandleGetBrief))
	mux.HandleFunc("PATCH /api/v1/briefs/$", wrapAuth(authMiddleware, briefHandler.HandleUpdateBrief))
	mux.HandleFunc("POST /api/v1/briefs/$/publish", wrapAuth(authMiddleware, briefHandler.HandlePublishBrief))
	mux.HandleFunc("POST /api/v1/briefs/$/close", wrapAuth(authMiddleware, briefHandler.HandleCloseBrief))
	mux.HandleFunc("POST /api/v1/briefs/$/interview", briefHandler.HandleInterview)
	mux.HandleFunc("GET /api/v1/briefs/matching", wrapAuth(authMiddleware, briefHandler.HandleMatchingBriefs))
	mux.HandleFunc("POST /api/v1/briefs/$/view", wrapAuth(authMiddleware, briefHandler.HandleViewBrief))
	mux.HandleFunc("POST /api/v1/briefs/$/raw-footage/upload-url", wrapAuth(authMiddleware, briefHandler.HandleGetRawFootageUploadURL))
	mux.HandleFunc("POST /api/v1/briefs/$/raw-footage/confirm", wrapAuth(authMiddleware, briefHandler.HandleConfirmRawFootageUpload))
	mux.HandleFunc("GET /api/v1/briefs/$/raw-footage/download-url", wrapAuth(authMiddleware, briefHandler.HandleGetRawFootageDownloadURL))

	// Add middleware chain
	chain := pkgmiddleware.Chain(
		mux,
		pkgmiddleware.RequestID,
		pkgmiddleware.Recover,
		pkgmiddleware.Logger,
		pkgmiddleware.CORS("*"),
	)

	// Create server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
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

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Shutdown error", slog.String("error", err.Error()))
	}

	log.Info("Server stopped")
}

// handleHealth is the health check endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

// wrapAuth wraps a handler with auth middleware
func wrapAuth(auth *briefmiddleware.AuthMiddleware, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth.Handler(handler).ServeHTTP(w, r)
	}
}