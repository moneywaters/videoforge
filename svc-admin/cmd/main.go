package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/videoforge/backend/pkg/config"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/pkg/natsclient"

	"github.com/videoforge/backend/svc-admin/internal/handler"
	"github.com/videoforge/backend/svc-admin/internal/repository"
	"github.com/videoforge/backend/svc-admin/internal/service"
)

func main() {
	cfg, err := config.Load("SERVER")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	dbURL := cfg.DatabaseURL
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		fmt.Fprintf(os.Stderr, "DATABASE_URL is required\n")
		os.Exit(1)
	}

	jwtSecret := cfg.JWTSecret
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	log := logger.Default(cfg.Environment)

	// NATS client (optional)
	var natsClient *natsclient.Client
	if cfg.NatsURL != "" {
		natsCfg := natsclient.DefaultConfig()
		natsCfg.URL = cfg.NatsURL
		nc := natsclient.New(natsCfg, log)
		if err := nc.Connect(); err == nil {
			natsClient = nc
		}
	}

	// Setup repositories, services, handlers
	repo := repository.NewAdminRepository(pool)
	svc := service.NewAdminService(repo, natsClient)
	adminHandler := handler.NewAdminHandler(svc)

	// Setup router
	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Admin routes
	mux.HandleFunc("GET /api/v1/admin/users", adminHandler.ListUsers)
	mux.HandleFunc("GET /api/v1/admin/users/{id}", adminHandler.GetUser)
	mux.HandleFunc("PATCH /api/v1/admin/users/{id}/role", adminHandler.AssignRoles)
	mux.HandleFunc("PATCH /api/v1/admin/users/{id}/status", adminHandler.BanUser)
	mux.HandleFunc("GET /api/v1/admin/disputes", adminHandler.ListDisputes)
	mux.HandleFunc("GET /api/v1/admin/disputes/{id}", adminHandler.GetDispute)
	mux.HandleFunc("PATCH /api/v1/admin/disputes/{id}/resolve", adminHandler.ResolveDispute)
	mux.HandleFunc("GET /api/v1/admin/moderation", adminHandler.ListModerationQueue)
	mux.HandleFunc("POST /api/v1/admin/moderation/{id}/action", adminHandler.ReviewModerationItem)

	// Wrap with middleware
	var handler http.Handler = mux
	handler = middleware.CORS("*")(handler)
	handler = middleware.Logger(handler)
	handler = middleware.Recover(handler)
	handler = middleware.RequestID(handler)

	addr := fmt.Sprintf(":%s", cfg.Port)
	if cfg.Port == "" {
		addr = ":8080"
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("admin server starting", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down admin server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown failed", slog.String("error", err.Error()))
	}
	log.Info("admin server stopped")
}
