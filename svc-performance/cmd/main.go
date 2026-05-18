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

	"github.com/videoforge/backend/pkg/config"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/pkg/natsclient"

	"github.com/videoforge/backend/svc-performance/internal/config"
	"github.com/videoforge/backend/svc-performance/internal/handler"
)

func main() {
	cfg, err := config.Load("SERVER")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.Default(cfg.Server.Environment)
	ctx := context.Background()

	nc := natsclient.New(natsclient.DefaultConfig(), log)
	if err := nc.ConnectWithRetry(5, 2*time.Second); err != nil {
		log.Error("Failed to connect to NATS", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer nc.Close()

	log.Info("Starting Performance service",
		slog.String("environment", cfg.Server.Environment),
		slog.String("port", cfg.Server.Port),
	)

	mux := http.NewServeMux()
	healthHandler := handler.NewHealthHandler()
	mux.HandleFunc("GET /health", healthHandler.HandleHealth)

	chain := middleware.Chain(mux, middleware.RequestID, middleware.Recover, middleware.Logger, middleware.CORS("*"))

	server := &http.Server{
		Addr:    cfg.Server.GetAddress(),
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