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
	"github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/pkg/natsclient"

	pkgconfig "github.com/videoforge/backend/svc-performance/internal/config"
	"github.com/videoforge/backend/svc-performance/internal/handler"
)

func main() {
	cfg, err := pkgconfig.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.Default(cfg.Server.Environment)

	// Initialize NATS
	nc := natsclient.New(&natsclient.Config{
		URL: cfg.NATS.URL,
	}, log)
	if err := nc.ConnectWithRetry(5, 2*time.Second); err != nil {
		log.Warn("NATS not connected, events will be disabled",
			slog.String("error", err.Error()),
		)
		nc = nil
	} else {
		defer nc.Close()
	}

	log.Info("Starting Performance service",
		slog.String("environment", cfg.Server.Environment),
		slog.String("port", cfg.Server.Port),
	)

	// Note: Not initializing database or handlers for now
	// TODO: Initialize repository, service, and handlers when database is ready

	mux := http.NewServeMux()
	healthHandler := handler.NewHealthHandler()
	mux.HandleFunc("GET /health", healthHandler.HandleHealth)

	chain := middleware.Chain(mux, middleware.RequestID, middleware.Recover, middleware.Logger, middleware.CORS("*"))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Server.Port),
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