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

	cfg "github.com/videoforge/backend/svc-video/internal/config"
	"github.com/videoforge/backend/svc-video/internal/handler"
	"github.com/videoforge/backend/svc-video/internal/repository"
	"github.com/videoforge/backend/svc-video/internal/service"
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
		log.Error("Failed to connect to NATS", slog.String("error", err.Error()))
		// Continue without NATS for development
	}

	// Load JWT public key for auth middleware
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
	} else {
		log.Warn("JWT_PUBLIC_KEY not set, running without authentication")
	}

	// Initialize repository
	repo := repository.NewVideoRepository(configuration.Database.Pool)

	// Initialize service with storage (if configured)
	// If no storage credentials, storageClient will be nil and mock storage will be used
	var storageClient interface{} = nil
	if configuration.Storage.Enabled && configuration.Storage.Client != nil {
		storageClient = configuration.Storage.Client
	}
	svc := service.NewVideoService(repo, nc, log, storageClient)

	// Initialize handler
	videoHandler := handler.NewVideoHandler(svc, log)

	log.Info("Starting Video service",
		slog.String("environment", configuration.Server.Environment),
		slog.String("port", configuration.Server.Port),
	)

	mux := http.NewServeMux()

	// Register routes
	registerRoutes(mux, videoHandler, publicKey)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"status":"ok"}`)
	})

	// Apply middleware chain
	chain := middleware.Chain(mux,
		middleware.RequestID,
		middleware.Recover,
		middleware.Logger,
		middleware.CORS("*"),
	)

	// Add auth middleware if public key is available
	if publicKey != nil {
		auth := middleware.NewAuthMiddleware(publicKey)
		chain = middleware.Chain(chain, func(next http.Handler) http.Handler {
			return auth.Authenticate(next)
		})
	}

	server := &http.Server{
		Addr:    configuration.Server.GetAddress(),
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

	if nc != nil {
		nc.Close()
	}
	if configuration.Database.Pool != nil {
		configuration.Database.Pool.Close()
	}

	log.Info("Server stopped")
}

// registerRoutes registers all HTTP routes
func registerRoutes(mux *http.ServeMux, h *handler.VideoHandler, publicKey *rsa.PublicKey) {
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