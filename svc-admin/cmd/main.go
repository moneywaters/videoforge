package main

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/videoforge/backend/pkg/config"
	"github.com/videoforge/backend/pkg/database"
	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/pkg/natsclient"

	"svc-admin/internal/config"
	"svc-admin/internal/handler"
	"svc-admin/internal/repository"
	"svc-admin/internal/service"
)

type route struct {
	pattern *regexp.Regexp
	handler http.HandlerFunc
	method  string
}

var routes = []route{
	{regexp.MustCompile(`^/api/v1/admin/users$`), nil, "GET"},
	{regexp.MustCompile(`^/api/v1/admin/users/([a-f0-9-]+)$`), nil, "GET"},
	{regexp.MustCompile(`^/api/v1/admin/users/([a-f0-9-]+)/ban$`), nil, "POST"},
	{regexp.MustCompile(`^/api/v1/admin/users/([a-f0-9-]+)/unban$`), nil, "POST"},
	{regexp.MustCompile(`^/api/v1/admin/users/([a-f0-9-]+)/roles$`), nil, "POST"},
	{regexp.MustCompile(`^/api/v1/admin/disputes$`), nil, "GET"},
	{regexp.MustCompile(`^/api/v1/admin/disputes/([a-f0-9-]+)$`), nil, "GET"},
	{regexp.MustCompile(`^/api/v1/admin/disputes/([a-f0-9-]+)/resolve$`), nil, "POST"},
	{regexp.MustCompile(`^/api/v1/admin/moderation-queue$`), nil, "GET"},
	{regexp.MustCompile(`^/api/v1/admin/moderation-queue/([a-f0-9-]+)/review$`), nil, "POST"},
	{regexp.MustCompile(`^/api/v1/admin/payouts/([a-f0-9-]+)/override$`), nil, "POST"},
	{regexp.MustCompile(`^/api/v1/admin/actions$`), nil, "GET"},
}

// routeHandler stores the handler functions
type routeHandler struct {
	adminHandler *handler.AdminHandler
}

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
		log.Warn("NATS not connected, events will be disabled",
			slog.String("error", err.Error()),
		)
		nc = nil
	} else {
		defer nc.Close()
	}

	// Initialize database
	db, err := database.NewPool(cfg.Database.ConnString, log)
	if err != nil {
		log.Error("Failed to connect to database",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize repository
	adminRepo := repository.NewAdminRepository(db)

	// Initialize service
	adminService := service.NewAdminService(adminRepo, nc)

	// Initialize handler
	adminHandler := handler.NewAdminHandler(adminService)

	// Parse JWT public key for auth middleware
	publicKey, err := parseRSAPublicKey(cfg.JWT.PublicKey)
	if err != nil {
		log.Warn("Failed to parse JWT public key, using default",
			slog.String("error", err.Error()),
		)
		publicKey = nil
	}

	// Initialize auth middleware
	var authMiddleware *middleware.AuthMiddleware
	if publicKey != nil {
		authMiddleware = middleware.NewAuthMiddleware(publicKey)
	}

	log.Info("Starting Admin service",
		slog.String("environment", cfg.Server.Environment),
		slog.String("port", cfg.Server.Port),
	)

	mux := http.NewServeMux()
	healthHandler := handler.NewHealthHandler()
	mux.HandleFunc("GET /health", healthHandler.HandleHealth)

	// Create router with handlers
	rh := &routeHandler{adminHandler: adminHandler}

	// Admin endpoints with auth
	mux.HandleFunc("GET /api/v1/admin/users", rh.requireAuth(authMiddleware, rh.listUsers))
	mux.HandleFunc("GET /api/v1/admin/users/", rh.requireAuth(authMiddleware, rh.getUser))
	mux.HandleFunc("POST /api/v1/admin/users/", rh.requireAuth(authMiddleware, rh.banUser))
	mux.HandleFunc("POST /api/v1/admin/users/", rh.requireAuth(authMiddleware, rh.unbanUser))
	mux.HandleFunc("POST /api/v1/admin/users/", rh.requireAuth(authMiddleware, rh.assignRoles))
	mux.HandleFunc("GET /api/v1/admin/disputes", rh.requireAuth(authMiddleware, rh.listDisputes))
	mux.HandleFunc("GET /api/v1/admin/disputes/", rh.requireAuth(authMiddleware, rh.getDispute))
	mux.HandleFunc("POST /api/v1/admin/disputes/", rh.requireAuth(authMiddleware, rh.resolveDispute))
	mux.HandleFunc("GET /api/v1/admin/moderation-queue", rh.requireAuth(authMiddleware, rh.listModerationQueue))
	mux.HandleFunc("POST /api/v1/admin/moderation-queue/", rh.requireAuth(authMiddleware, rh.reviewModerationItem))
	mux.HandleFunc("POST /api/v1/admin/payouts/", rh.requireAuth(authMiddleware, rh.overridePayout))
	mux.HandleFunc("GET /api/v1/admin/actions", rh.requireAuth(authMiddleware, rh.listAdminActions))

	// Chain middleware
	handler := mux
	handler = middleware.RequestID(handler)
	handler = middleware.Recover(handler)
	handler = middleware.Logger(handler)
	handler = middleware.CORS("*")(handler)

	server := &http.Server{
		Addr:    cfg.Server.GetAddress(),
		Handler: handler,
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

// requireAuth wraps a handler with authentication
func (rh *routeHandler) requireAuth(auth *middleware.AuthMiddleware, handler http.HandlerFunc) http.HandlerFunc {
	if auth == nil {
		return handler
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		role := middleware.GetUserRole(ctx)
		if role != "admin" {
			errors.WriteError(ctx, w, errors.Forbidden("admin access required"))
			return
		}
		handler(w, r)
	}
}

// listUsers handles GET /api/v1/admin/users
func (rh *routeHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	rh.adminHandler.ListUsers(w, r)
}

// getUser handles GET /api/v1/admin/users/:id
func (rh *routeHandler) getUser(w http.ResponseWriter, r *http.Request) {
	rh.adminHandler.GetUser(w, r)
}

// banUser handles POST /api/v1/admin/users/:id/ban
func (rh *routeHandler) banUser(w http.ResponseWriter, r *http.Request) {
	rh.adminHandler.BanUser(w, r)
}

// unbanUser handles POST /api/v1/admin/users/:id/unban
func (rh *routeHandler) unbanUser(w http.ResponseWriter, r *http.Request) {
	rh.adminHandler.UnbanUser(w, r)
}

// assignRoles handles POST /api/v1/admin/users/:id/roles
func (rh *routeHandler) assignRoles(w http.ResponseWriter, r *http.Request) {
	rh.adminHandler.AssignRoles(w, r)
}

// listDisputes handles GET /api/v1/admin/disputes
func (rh *routeHandler) listDisputes(w http.ResponseWriter, r *http.Request) {
	rh.adminHandler.ListDisputes(w, r)
}

// getDispute handles GET /api/v1/admin/disputes/:id
func (rh *routeHandler) getDispute(w http.ResponseWriter, r *http.Request) {
	rh.adminHandler.GetDispute(w, r)
}

// resolveDispute handles POST /api/v1/admin/disputes/:id/resolve
func (rh *routeHandler) resolveDispute(w http.ResponseWriter, r *http.Request) {
	rh.adminHandler.ResolveDispute(w, r)
}

// listModerationQueue handles GET /api/v1/admin/moderation-queue
func (rh *routeHandler) listModerationQueue(w http.ResponseWriter, r *http.Request) {
	rh.adminHandler.ListModerationQueue(w, r)
}

// reviewModerationItem handles POST /api/v1/admin/moderation-queue/:id/review
func (rh *routeHandler) reviewModerationItem(w http.ResponseWriter, r *http.Request) {
	rh.adminHandler.ReviewModerationItem(w, r)
}

// overridePayout handles POST /api/v1/admin/payouts/:id/override
func (rh *routeHandler) overridePayout(w http.ResponseWriter, r *http.Request) {
	rh.adminHandler.OverridePayout(w, r)
}

// listAdminActions handles GET /api/v1/admin/actions
func (rh *routeHandler) listAdminActions(w http.ResponseWriter, r *http.Request) {
	rh.adminHandler.ListAdminActions(w, r)
}

// parseRSAPublicKey parses a base64-encoded RSA public key
func parseRSAPublicKey(keyData string) (*rsa.PublicKey, error) {
	if keyData == "" {
		return nil, fmt.Errorf("empty key data")
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Parse RSA public key
	key, err := jwt.ParseRSAPublicKeyFromPEM(decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	return key, nil
}