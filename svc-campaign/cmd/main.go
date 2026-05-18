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

	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/pkg/natsclient"

	svcconfig "svc-campaign/internal/config"
	"svc-campaign/internal/handler"
	"svc-campaign/internal/repository"
	"svc-campaign/internal/service"
)

func main() {
	cfg, err := svcconfig.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.Default(cfg.Server.Environment)
	ctx := logger.FromContext(context.Background()).Context(context.Background())

	// Connect to NATS (optional)
	nc := natsclient.New(nil, log)
	if err := nc.ConnectWithRetry(5, 2*time.Second); err != nil {
		log.Warn("NATS connection failed, continuing without NATS",
			slog.String("error", err.Error()),
		)
		nc = nil
	} else {
		defer nc.Close()
	}

	// Initialize repository
	repo := repository.NewCampaignRepo(cfg.Database.Pool)

	// Initialize service
	svc := service.NewCampaignService(repo, nc)

	// Initialize handler
	campaignHandler := handler.NewCampaignHandler(svc)

	log.Info("Starting Campaign service",
		slog.String("environment", cfg.Server.Environment),
		slog.String("port", cfg.Server.Port),
	)

	// Create router
	router := http.NewServeMux()

	// Health check
	router.HandleFunc("GET /health", handleHealth)

	// Campaign CRUD
	router.HandleFunc("POST /api/v1/campaigns", campaignHandler.HandleCreateCampaign)
	router.HandleFunc("GET /api/v1/campaigns", campaignHandler.HandleListCampaigns)
	router.HandleFunc("GET /api/v1/campaigns/", campaignHandler.HandleGetCampaign)
	router.HandleFunc("PATCH /api/v1/campaigns/", campaignHandler.HandleUpdateCampaign)

	// Campaign actions (start/pause/end)
	router.HandleFunc("POST /api/v1/campaigns/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if !strings.HasPrefix(path, "/api/v1/campaigns/") {
			http.NotFound(w, r)
			return
		}
		
		// Extract campaign ID and suffix
		remaining := strings.TrimPrefix(path, "/api/v1/campaigns/")
		
		// Check for actions
		actions := map[string]http.HandlerFunc{
			"start": campaignHandler.HandleStartCampaign,
			"pause": campaignHandler.HandlePauseCampaign,
			"end":  campaignHandler.HandleEndCampaign,
		}
		
		for action, h := range actions {
			suffix := "/" + action
			if strings.HasSuffix(remaining, suffix) {
				campaignID := strings.TrimSuffix(remaining, suffix)
				if campaignID != "" {
					r.URL.Path = "/api/v1/campaigns/" + campaignID
					h(w, r)
					return
				}
			}
		}
		
		http.NotFound(w, r)
	})

	// Campaign videos
	router.HandleFunc("POST /api/v1/campaigns/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/videos") {
			campaignID := strings.TrimSuffix(strings.TrimPrefix(path, "/api/v1/campaigns/"), "/videos")
			if campaignID != "" {
				r.URL.Path = "/api/v1/campaigns/" + campaignID
				campaignHandler.HandleAddVideo(w, r)
				return
			}
		}
		http.NotFound(w, r)
	})
	
	router.HandleFunc("DELETE /api/v1/campaigns/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Check pattern: /api/v1/campaigns/{id}/videos/{video_id}
		prefix := "/api/v1/campaigns/"
		if strings.HasPrefix(path, prefix) {
			remaining := strings.TrimPrefix(path, prefix)
			if strings.HasSuffix(remaining, "/videos") {
				http.NotFound(w, r)
				return
			}
			// Check for /videos/ pattern
			if idx := strings.Index(remaining, "/videos/"); idx > 0 {
				campaignID := remaining[:idx]
				r.URL.Path = "/api/v1/campaigns/" + campaignID
				campaignHandler.HandleRemoveVideo(w, r)
				return
			}
		}
		http.NotFound(w, r)
	})

	// Campaign budget
	router.HandleFunc("GET /api/v1/campaigns/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		prefix := "/api/v1/campaigns/"
		if strings.HasPrefix(path, prefix) {
			remaining := strings.TrimPrefix(path, prefix)
			if strings.HasSuffix(remaining, "/budget") {
				campaignID := strings.TrimSuffix(remaining, "/budget")
				if campaignID != "" && r.Method == "GET" {
					r.URL.Path = "/api/v1/campaigns/" + campaignID
					campaignHandler.HandleGetBudget(w, r)
					return
				}
			}
		}
	})
	
	router.HandleFunc("POST /api/v1/campaigns/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		prefix := "/api/v1/campaigns/"
		if strings.HasPrefix(path, prefix) {
			remaining := strings.TrimPrefix(path, prefix)
			if strings.HasSuffix(remaining, "/budget") {
				campaignID := strings.TrimSuffix(remaining, "/budget")
				if campaignID != "" && r.Method == "POST" {
					r.URL.Path = "/api/v1/campaigns/" + campaignID
					campaignHandler.HandleUpdateBudget(w, r)
					return
				}
			}
		}
		http.NotFound(w, r)
	})

	// Ad account routes (placeholder)
	router.HandleFunc("POST /api/v1/ad-accounts", campaignHandler.HandleCreateAdAccount)
	router.HandleFunc("GET /api/v1/ad-accounts", campaignHandler.HandleListAdAccounts)

	// Auth middleware (use nil public key for development)
	authMiddleware := middleware.NewAuthMiddleware(nil)

	// Chain middleware
	chain := middleware.Chain(router,
		middleware.RequestID,
		middleware.Recover,
		middleware.Logger,
		middleware.CORS("*"),
		authMiddleware.Authenticate,
	)

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
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

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}