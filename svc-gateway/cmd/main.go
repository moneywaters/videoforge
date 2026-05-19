package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/videoforge/backend/pkg/config"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/middleware"
)

func main() {
	cfg, err := config.Load("SERVER")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.Default(cfg.Environment)
	log.Info("Starting Gateway service",
		slog.String("environment", cfg.Environment),
		slog.String("port", cfg.Port),
	)

	userTarget, _ := url.Parse("http://videoforge-user.flycast:8080")
	briefTarget, _ := url.Parse("http://videoforge-brief.flycast:8080")

	userProxy := httputil.NewSingleHostReverseProxy(userTarget)
	briefProxy := httputil.NewSingleHostReverseProxy(briefTarget)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Proxy routes
	mux.Handle("/api/v1/auth/", userProxy)
	mux.Handle("/api/v1/users/", userProxy)
	mux.Handle("/api/v1/briefs/", briefProxy)
	mux.Handle("/api/v1/projects/", briefProxy)
	mux.Handle("/api/v1/proposals/", briefProxy)
	mux.Handle("/api/v1/milestones/", briefProxy)
	mux.Handle("/api/v1/submissions/", briefProxy)
	mux.Handle("/api/v1/payments/", briefProxy)
	mux.Handle("/api/v1/disputes/", briefProxy)
	mux.Handle("/api/v1/reviews/", briefProxy)
	mux.Handle("/api/v1/notifications/", briefProxy)
	mux.Handle("/api/v1/analytics/", briefProxy)
	mux.Handle("/api/v1/assets/", briefProxy)

	chain := middleware.Chain(
		mux,
		middleware.RequestID,
		middleware.Recover,
		middleware.Logger,
		middleware.CORS("*"),
	)

	server := &http.Server{
		Addr:    cfg.GetAddress(),
		Handler: chain,
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
