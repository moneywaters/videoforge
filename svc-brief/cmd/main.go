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

	"github.com/golang-jwt/jwt/v5"
	"github.com/videoforge/backend/pkg/logger"
	pkgmiddleware "github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/svc-brief/internal/handler"
	"github.com/videoforge/backend/svc-brief/internal/middleware"
	"github.com/videoforge/backend/svc-brief/internal/repository"
	"github.com/videoforge/backend/svc-brief/internal/service"

	briefconfig "github.com/videoforge/backend/svc-brief/internal/config"
)

func main() {
	cfg, err := briefconfig.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	defer cfg.Close()

	log := logger.Default(cfg.Server.Environment)
	log.Info("Starting Brief service", slog.String("port", cfg.Server.Port))

	repo := repository.NewBriefRepo(cfg.Database.Pool)
	svc := service.NewBriefService(repo, cfg.Storage.Client)
	briefHandler := handler.NewBriefHandler(svc)

	// Auth wrapper: extracts token into request context for all routes except /health
	authMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenStr := parts[1]
					token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
					if err == nil {
						if claims, ok := token.Claims.(jwt.MapClaims); ok {
							if sub, ok := claims["sub"].(string); ok && sub != "" {
								role, _ := claims["role"].(string)
								ctx := context.WithValue(r.Context(), middleware.UserIDKey, sub)
								ctx = context.WithValue(ctx, middleware.UserRoleKey, role)
								r = r.WithContext(ctx)
							}
						}
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})
	mux.HandleFunc("POST /api/v1/briefs", briefHandler.HandleCreateBrief)
	mux.HandleFunc("GET /api/v1/briefs", briefHandler.HandleListBriefs)
	mux.HandleFunc("GET /api/v1/briefs/$", briefHandler.HandleGetBrief)
	mux.HandleFunc("PATCH /api/v1/briefs/$", briefHandler.HandleUpdateBrief)
	mux.HandleFunc("POST /api/v1/briefs/$/publish", briefHandler.HandlePublishBrief)
	mux.HandleFunc("POST /api/v1/briefs/$/close", briefHandler.HandleCloseBrief)
	mux.HandleFunc("POST /api/v1/briefs/$/interview", briefHandler.HandleInterview)
	mux.HandleFunc("GET /api/v1/briefs/matching", briefHandler.HandleMatchingBriefs)
	mux.HandleFunc("POST /api/v1/briefs/$/view", briefHandler.HandleViewBrief)
	mux.HandleFunc("POST /api/v1/briefs/$/raw-footage/upload-url", briefHandler.HandleGetRawFootageUploadURL)
	mux.HandleFunc("POST /api/v1/briefs/$/raw-footage/confirm", briefHandler.HandleConfirmRawFootageUpload)
	mux.HandleFunc("GET /api/v1/briefs/$/raw-footage/download-url", briefHandler.HandleGetRawFootageDownloadURL)

	chain := pkgmiddleware.Chain(
		mux,
		pkgmiddleware.RequestID,
		pkgmiddleware.Recover,
		pkgmiddleware.Logger,
		pkgmiddleware.CORS("*"),
		authMw,
	)

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
