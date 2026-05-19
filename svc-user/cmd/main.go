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

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/videoforge/backend/pkg/config"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/middleware"
)

// User represents a user in the system.
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name     string `json:"name,omitempty"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// RegisterRequest represents a registration request.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents an authentication response.
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// UserResponse represents a user response.
type UserResponse struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Name  string `json:"name,omitempty"`
	Role  string `json:"role"`
	Status string `json:"status"`
}

func main() {
	// Load configuration
	cfg, err := config.Load("SERVER")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Get database URL from config or environment
	dbURL := cfg.DatabaseURL
	if dbURL == "" {
		dbURL = os.Getenv("DATABASE_URL")
	}
	if dbURL == "" {
		fmt.Fprintf(os.Stderr, "DATABASE_URL is required\n")
		os.Exit(1)
	}

	// Get JWT secret
	jwtSecret := cfg.JWTSecret
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}
	if jwtSecret == "" {
		slog.Warn("JWT_SECRET not set, using insecure default for development")
		jwtSecret = "insecure-dev-secret-change-in-production"
	}

	// Create database connection pool
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create database pool: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Ensure users table exists
	ctx := context.Background()
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			name TEXT,
			role TEXT DEFAULT 'client',
			status TEXT DEFAULT 'active',
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create users table: %v\n", err)
		os.Exit(1)
	}

	slog.Info("database initialized", "table", "users")

	// Create logger
	log := logger.Default(cfg.Environment)
	if log == nil {
		log = logger.New(nil, false, logger.LevelInfo)
	}

	// Setup router
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API routes
	api := http.NewServeMux()

	// Register handler
	api.HandleFunc("POST /api/v1/auth/register", handleRegister(pool, jwtSecret, log))

	// Login handler
	api.HandleFunc("POST /api/v1/auth/login", handleLogin(pool, jwtSecret, log))

	// Logout handler
	api.HandleFunc("POST /api/v1/auth/logout", handleLogout)

	// Get current user
	api.HandleFunc("GET /api/v1/users/me", handleGetMe(pool, jwtSecret, log))

	// Apply global middleware
	handler := middleware.Chain(api,
		middleware.RequestID,
		middleware.Recover,
		middleware.Logger,
		middleware.CORS("*"),
	)

	// Create server
	addr := fmt.Sprintf(":%s", cfg.Port)
	if addr == ":" {
		addr = ":8080"
	}
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		slog.Info("server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")

	// Graceful shutdown with 10s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}

// generateToken generates a JWT token for the given user ID.
func generateToken(userID, jwtSecret string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// validateToken validates a JWT token and returns the claims.
func validateToken(tokenString, jwtSecret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

func handleRegister(pool *pgxpool.Pool, jwtSecret string, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		if req.Email == "" || req.Password == "" {
			http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
			return
		}

		// Hash password with bcrypt
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, `{"error":"failed to hash password"}`, http.StatusInternalServerError)
			return
		}

		// Insert user
		var userID, role, status string
		err = pool.QueryRow(context.Background(),
			`INSERT INTO users (email, password_hash, name, role, status) VALUES ($1, $2, $3, $4, $5) RETURNING id, role, status`,
			req.Email, string(hashedPassword), req.Name, "client", "active",
		).Scan(&userID, &role, &status)
		if err != nil {
			http.Error(w, `{"error":"user already exists or invalid"}`, http.StatusConflict)
			return
		}

		// Generate JWT
		token, err := generateToken(userID, jwtSecret)
		if err != nil {
			http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
			return
		}

		// Return response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(AuthResponse{
			Token: token,
			User: User{
				ID:     userID,
				Email:  req.Email,
				Name:  req.Name,
				Role:  role,
				Status: status,
			},
		})
	}
}

func handleLogin(pool *pgxpool.Pool, jwtSecret string, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}

		if req.Email == "" || req.Password == "" {
			http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
			return
		}

		// Get user by email
		var userID, passwordHash, name, role, status string
		err := pool.QueryRow(context.Background(),
			`SELECT id, password_hash, COALESCE(name, ''), role, status FROM users WHERE email = $1`,
			req.Email,
		).Scan(&userID, &passwordHash, &name, &role, &status)
		if err != nil {
			http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
			return
		}

		// Generate JWT
		token, err := generateToken(userID, jwtSecret)
		if err != nil {
			http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
			return
		}

		// Return response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(AuthResponse{
			Token: token,
			User: User{
				ID:     userID,
				Email:  req.Email,
				Name:  name,
				Role:  role,
				Status: status,
			},
		})
	}
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}

func handleGetMe(pool *pgxpool.Pool, jwtSecret string, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"authorization header required"}`, http.StatusUnauthorized)
			return
		}

		// Extract token
		tokenString := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		// Validate JWT
		claims, err := validateToken(tokenString, jwtSecret)
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Get user ID from claims
		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			http.Error(w, `{"error":"invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		// Get user from database
		var email, name, role, status string
		err = pool.QueryRow(context.Background(),
			`SELECT email, COALESCE(name, ''), role, status FROM users WHERE id = $1`,
			sub,
		).Scan(&email, &name, &role, &status)
		if err != nil {
			http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
			return
		}

		// Return response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(UserResponse{
			ID:     sub,
			Email:  email,
			Name:  name,
			Role:  role,
			Status: status,
		})
	}
}