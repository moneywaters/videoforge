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

	"github.com/gorilla/mux"

	"github.com/videoforge/backend/pkg/config"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/pkg/natsclient"

	"github.com/videoforge/backend/svc-payout/internal/config"
	"github.com/videoforge/backend/svc-payout/internal/handler"
	"github.com/videoforge/backend/svc-payout/internal/repository"
	"github.com/videoforge/backend/svc-payout/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.Default(cfg.Server.Environment)
	ctx := logger.FromContext(context.Background()).Context(context.Background())

	// Initialize NATS client
	nc := natsclient.New(nil, log)
	if err := nc.ConnectWithRetry(5, 2*time.Second); err != nil {
		log.Warn("Failed to connect to NATS, continuing without event subscription", slog.String("error", err.Error()))
		nc = nil
	}
	if nc != nil {
		defer nc.Close()
	}

	// Initialize database repository
	repo := repository.NewPayoutRepository(cfg.Database.Pool)
	if err := repo.InitSchema(ctx); err != nil {
		log.Error("Failed to initialize database schema", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize payout service
	payoutSvc := service.NewPayoutService(repo, nc, cfg.HoldPeriod)

	// Subscribe to NATS events
	if nc != nil {
		if err := payoutSvc.SubscribeToEvents(ctx); err != nil {
			log.Warn("Failed to subscribe to NATS events", slog.String("error", err.Error()))
		}
	}

	// Start hold release cron (runs every hour)
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := payoutSvc.ReleaseHolds(ctx); err != nil {
					log.Error("Failed to release holds", slog.String("error", err.Error()))
				}
			}
		}
	}()

	log.Info("Starting Payout service",
		slog.String("environment", cfg.Server.Environment),
		slog.String("port", cfg.Server.Port),
	)

	mux := http.NewServeMux()

	// Initialize handlers
	healthHandler := handler.NewHealthHandler()
	payoutHandler := handler.NewPayoutHandler(payoutSvc)

	// Health endpoints
	mux.HandleFunc("GET /health", healthHandler.HandleHealth)

	// Payout API endpoints
	mux.HandleFunc("GET /api/v1/payouts", payoutHandler.GetPayouts)
	mux.HandleFunc("GET /api/v1/payouts/{id}", payoutHandler.GetPayoutByID)
	mux.HandleFunc("GET /api/v1/balance", payoutHandler.GetBalance)
	mux.HandleFunc("GET /api/v1/earnings", payoutHandler.GetEarnings)
	mux.HandleFunc("POST /api/v1/payouts/calculate", payoutHandler.CalculateEarnings)
	mux.HandleFunc("POST /api/v1/payouts/batches", payoutHandler.CreateBatch)
	mux.HandleFunc("GET /api/v1/payouts/batches", payoutHandler.GetBatches)
	mux.HandleFunc("GET /api/v1/payouts/batches/{id}", payoutHandler.GetBatchByID)
	mux.HandleFunc("POST /api/v1/payouts/batches/{id}/process", payoutHandler.ProcessBatch)

	// Webhook endpoints
	mux.HandleFunc("POST /api/v1/payouts/webhook/dodo", payoutHandler.HandleDodoWebhook)
	mux.HandleFunc("POST /api/v1/payouts/webhook/ruul", payoutHandler.HandleRuulWebhook)

	// Middleware chain
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

	// Close database
	cfg.Close()

	log.Info("Server stopped")
}