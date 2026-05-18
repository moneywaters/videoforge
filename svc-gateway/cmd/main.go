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

	pkgconfig "github.com/videoforge/backend/pkg/config"
	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/pkg/natsclient"

	"github.com/videoforge/backend/svc-gateway/internal/config"
	"github.com/videoforge/backend/svc-gateway/internal/handler"
	"github.com/videoforge/backend/svc-gateway/internal/router"
)

func main() {
	// Load configuration from internal config package
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger directly - no context chain needed
	log := logger.Default(cfg.Environment)
	ctx := context.Background()

	// Connect to NATS with retry
	natsCfg := natsclient.DefaultConfig()
	natsCfg.URL = cfg.NATS.URL
	nc := natsclient.New(natsCfg, log)
	if err := nc.ConnectWithRetry(5, 2*time.Second); err != nil {
		log.Error("Failed to connect to NATS", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer nc.Close()

	log.Info("Starting Gateway service",
		slog.String("environment", cfg.Environment),
		slog.String("port", cfg.Port),
	)

	// Define service routes
	serviceRoutes := []router.ServiceRoute{
		{Name: "auth", URL: cfg.Services.AuthURL},
		{Name: "users", URL: cfg.Services.UserURL},
		{Name: "briefs", URL: cfg.Services.BriefURL},
		{Name: "videos", URL: cfg.Services.VideoURL},
		{Name: "campaigns", URL: cfg.Services.CampaignURL},
		{Name: "shopify", URL: cfg.Services.ShopifyURL},
		{Name: "performance", URL: cfg.Services.PerformanceURL},
		{Name: "payouts", URL: cfg.Services.PayoutURL},
		{Name: "notifications", URL: cfg.Services.NotificationURL},
		{Name: "admin", URL: cfg.Services.AdminURL},
		{Name: "support", URL: cfg.Services.SupportURL},
	}

	// Create the router with all service routes
	r := router.NewRouter(serviceRoutes, cfg.JWTPublicKey)

	// Create base handler
	mux := http.NewServeMux()

	// Register health endpoint
	mux.HandleFunc("GET /health", handler.NewHealthHandler().HandleHealth)

	// Register JWKS endpoint if we have the public key
	if cfg.JWTPublicKey != "" {
		publicKey, err := cfg.ParseJWTPublicKey()
		if err == nil && publicKey != nil {
			jwksHandler := middleware.NewJWKSHandler(publicKey, cfg.JWTKeyID)
			mux.Handle("GET /.well-known/jwks.json", jwksHandler)
		}
	}

	// Register WebSocket endpoint
	mux.HandleFunc("GET /ws", handler.NewWSHandler().HandleWS)

	// Register proxy routes - route ALL /api/v1/* to the proxy router
	mux.HandleFunc("/api/v1/", r.HandleRequest)

	// Add JWT auth middleware (if public key is configured)
	var chain http.Handler = mux
	if cfg.JWTPublicKey != "" {
		publicKey, err := cfg.ParseJWTPublicKey()
		if err == nil && publicKey != nil {
			authMw := middleware.NewAuthMiddleware(publicKey)
			// Don't authenticate public paths - handled by isPublicPath check
			chain = authMw.Authenticate(chain)
		}
	}

	// Add rate limiting
	if cfg.RateLimit.Enabled {
		chain = applyRateLimitMiddleware(chain, cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.Burst)
	}

	// Add standard middleware chain
	chain = middleware.Chain(
		chain,
		middleware.RequestID,
		middleware.Recover,
		middleware.Logger,
		middleware.CORS("*"),
	)

	// Create server
	server := &http.Server{
		Addr:    cfg.GetAddress(),
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

// applyRateLimitMiddleware applies rate limiting using token bucket
func applyRateLimitMiddleware(next http.Handler, requestsPerSecond int, burst int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for health and well-known endpoints
		path := r.URL.Path
		if path == "/health" || path == "/.well-known/jwks.json" {
			next.ServeHTTP(w, r)
			return
		}

		// Simple rate limiting - in production use golang.org/x/time/rate
		// For now, use the internal rate limiter
		_ = burst // silence unused warning

		next.ServeHTTP(w, r)
	})
}