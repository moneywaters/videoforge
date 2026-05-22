package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/pkg/natsclient"

	cfg "github.com/videoforge/backend/svc-shopify/internal/config"
	shopifyhandler "github.com/videoforge/backend/svc-shopify/internal/handler"
	"github.com/videoforge/backend/svc-shopify/internal/repository"
	"github.com/videoforge/backend/svc-shopify/internal/service"
)

func main() {
	// Load configuration
	configuration, err := cfg.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.Default(configuration.Server.Environment)
	_ = logger.FromContext(context.Background()).Context(context.Background())

	// Connect to NATS
	nc := natsclient.New(nil, log)
	if err := nc.ConnectWithRetry(5, 2*time.Second); err != nil {
		log.Warn("Failed to connect to NATS, continuing without NATS",
			slog.String("error", err.Error()))
		nc = nil
	}
	defer func() {
		if nc != nil {
			nc.Close()
		}
	}()

	// Initialize repository
	repo := repository.NewShopifyRepository(configuration.Database.Pool)

	// Get store URL for generating links
	storeURL := os.Getenv("STORE_URL")
	if storeURL == "" {
		storeURL = "https://store.example.com"
	}

	// Initialize service
	svc := service.NewShopifyService(repo, nc, log, storeURL)

	// Initialize handler
	shopifyHandler := shopifyhandler.NewShopifyHandler(svc, log)

	log.Info("Starting Shopify service",
		slog.String("environment", configuration.Server.Environment),
		slog.String("port", configuration.Server.Port),
	)

	mux := http.NewServeMux()

	// Register routes
	registerRoutes(mux, shopifyHandler)

	// Health check (public)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"status":"ok","service":"shopify"}`)
	})

	// Apply base middleware chain
	chain := middleware.Chain(mux,
		middleware.RequestID,
		middleware.Recover,
		middleware.Logger,
		middleware.CORS("*"),
	)

	// Try to load JWT public key for authentication
	var publicKey *rsa.PublicKey
	jwtPublicKey := os.Getenv("JWT_PUBLIC_KEY")
	jwtPublicKeyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")

	if jwtPublicKey != "" {
		publicKey, err = loadPublicKeyFromPEM(jwtPublicKey)
		if err != nil {
			log.Error("Failed to parse JWT public key", slog.String("error", err.Error()))
			os.Exit(1)
		}
	} else if jwtPublicKeyPath != "" {
		keyData, err := os.ReadFile(jwtPublicKeyPath)
		if err != nil {
			log.Error("Failed to read JWT public key file", slog.String("error", err.Error()))
			os.Exit(1)
		}
		publicKey, err = loadPublicKeyFromPEM(string(keyData))
		if err != nil {
			log.Error("Failed to parse JWT public key from file", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	// Add auth middleware if public key is available
	if publicKey != nil {
		authMiddleware := middleware.NewAuthMiddleware(publicKey)
		chain = middleware.Chain(chain, func(next http.Handler) http.Handler {
			return authMiddleware.Authenticate(next)
		})
		log.Info("JWT authentication enabled")
	} else {
		log.Warn("JWT_PUBLIC_KEY not set, running without authentication")
	}

	server := &http.Server{
		Addr:         ":" + configuration.Server.Port,
		Handler:      chain,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
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

	if configuration.Database.Pool != nil {
		configuration.Database.Pool.Close()
	}

	log.Info("Server stopped")
}

// registerRoutes registers all HTTP routes
func registerRoutes(mux *http.ServeMux, h *shopifyhandler.ShopifyHandler) {
	h.Routes(mux)
}

// loadPublicKeyFromPEM loads an RSA public key from PEM format
func loadPublicKeyFromPEM(pemData string) (*rsa.PublicKey, error) {
	// Remove any whitespace and newlines
	pemData = strings.TrimSpace(pemData)

	// Decode PEM block
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse public key
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}