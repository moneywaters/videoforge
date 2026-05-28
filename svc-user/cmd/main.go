package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/videoforge/backend/pkg/config"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/middleware"
)

// User represents a user in the system.
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name,omitempty"`
	Role      string `json:"role"`
	Status    string `json:"status"`
}

// RegisterRequest represents a registration request.
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Role      string `json:"role,omitempty"`
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

var (
	pool      *pgxpool.Pool
	jwtSecret string
	googleOAuth *oauth2.Config
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

	jwtSecret = cfg.JWTSecret
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}
	if jwtSecret == "" {
		slog.Warn("JWT_SECRET not set, using insecure default for development")
		jwtSecret = "insecure-dev-secret-change-in-production"
	}

	pool, err = pgxpool.New(context.Background(), dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create database pool: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	ctx := context.Background()
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL DEFAULT '',
			first_name VARCHAR(255) NOT NULL DEFAULT '',
			last_name VARCHAR(255) NOT NULL DEFAULT '',
			role VARCHAR(50) NOT NULL DEFAULT 'client',
			status VARCHAR(50) NOT NULL DEFAULT 'active',
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			last_login_at TIMESTAMP WITH TIME ZONE,
			google_id VARCHAR(255) DEFAULT '',
			picture TEXT DEFAULT ''
		)
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create users table: %v\n", err)
		os.Exit(1)
	}

	// Add columns if they don't exist (for existing tables from before)
	pool.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS google_id VARCHAR(255) DEFAULT ''`)
	pool.Exec(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS picture TEXT DEFAULT ''`)

	slog.Info("database initialized", "table", "users")

	// Setup Google OAuth if credentials present
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID != "" && clientSecret != "" && clientID != "PLACEHOLDER" {
		googleOAuth = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}
		if googleOAuth.RedirectURL == "" {
			googleOAuth.RedirectURL = "https://videoforge-gateway.fly.dev/api/v1/auth/google/callback"
		}
		slog.Info("Google OAuth configured")
	} else {
		slog.Warn("Google OAuth credentials not set, skipping")
	}

	log := logger.Default(cfg.Environment)
	_ = log

	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Auth routes
	mux.HandleFunc("POST /api/v1/auth/register", handleRegister)
	mux.HandleFunc("POST /api/v1/auth/login", handleLogin)
	mux.HandleFunc("POST /api/v1/auth/logout", handleLogout)
	mux.HandleFunc("GET /api/v1/auth/google/login", handleGoogleLogin)
	mux.HandleFunc("GET /api/v1/auth/google/callback", handleGoogleCallback)

	// User routes
	mux.HandleFunc("GET /api/v1/users/me", handleGetMe)

	handler := middleware.Chain(mux,
		middleware.RequestID,
		middleware.Recover,
		middleware.Logger,
		middleware.CORS("*"),
	)

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

	go func() {
		slog.Info("server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}

// --- JWT helpers ---

func generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtSecret))
}

func generateTokenWithClaims(userID, email, name, role, picture string) (string, error) {
	claims := jwt.MapClaims{
		"sub":     userID,
		"email":   email,
		"name":    name,
		"role":    role,
		"picture": picture,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtSecret))
}

// --- Standard auth handlers ---

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}

	// Validate role if provided
	validRoles := map[string]bool{
		"client":       true,
		"editor":       true,
		"ad_specialist": true,
	}
	role := req.Role
	if role == "" {
		role = "client"
	} else if !validRoles[role] {
		http.Error(w, `{"error":"invalid role"}`, http.StatusBadRequest)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	var userID, status string
	err := pool.QueryRow(r.Context(),
		`INSERT INTO users (email, password_hash, first_name, last_name, role, status) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, role, status`,
		req.Email, string(hash), req.FirstName, req.LastName, role, "active",
	).Scan(&userID, &role, &status)
	if err != nil {
		http.Error(w, `{"error":"user already exists or invalid"}`, http.StatusConflict)
		return
	}

	// Construct name from first_name and last_name
	name := req.FirstName
	if req.LastName != "" {
		name = req.FirstName + " " + req.LastName
	}

	token, _ := generateToken(userID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: User{ID: userID, Email: req.Email, Name: name, Role: role, Status: status}})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}
	var userID, hash, fname, lname, role, status string
	err := pool.QueryRow(r.Context(),
		`SELECT id, password_hash, first_name, last_name, role, status FROM users WHERE email=$1`, req.Email,
	).Scan(&userID, &hash, &fname, &lname, &role, &status)
	if err != nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}
	token, _ := generateToken(userID)
	name := fname
	if lname != "" {
		name = fname + " " + lname
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{Token: token, User: User{ID: userID, Email: req.Email, Name: name, Role: role, Status: status}})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}

func handleGetMe(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	tokenStr := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenStr = authHeader[7:]
	}
	claims, err := validateToken(tokenStr)
	if err != nil {
		http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
		return
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
		return
	}
	var id, email, fname, lname, role, status, picture string
	err = pool.QueryRow(r.Context(),
		`SELECT id, email, first_name, last_name, role, status, COALESCE(picture,'') FROM users WHERE id=$1`, sub,
	).Scan(&id, &email, &fname, &lname, &role, &status, &picture)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}
	name := fname
	if lname != "" { name = fname + " " + lname }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(User{ID: id, Email: email, Name: name, Role: role, Status: status})
}

func validateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil { return nil, err }
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

// --- Google OAuth handlers ---

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if googleOAuth == nil {
		http.Error(w, `{"error":"Google OAuth not configured"}`, http.StatusNotImplemented)
		return
	}
	state := randState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	})
	url := googleOAuth.AuthCodeURL(state, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if googleOAuth == nil {
		http.Error(w, `{"error":"Google OAuth not configured"}`, http.StatusNotImplemented)
		return
	}
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value == "" {
		http.Error(w, `{"error":"invalid oauth state"}`, http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "oauth_state",
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Unix(0, 0),
	})
	if r.URL.Query().Get("state") != stateCookie.Value {
		http.Error(w, `{"error":"oauth state mismatch"}`, http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, `{"error":"missing code"}`, http.StatusBadRequest)
		return
	}

	token, err := googleOAuth.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to exchange code: %s"}`, err), http.StatusInternalServerError)
		return
	}

	// Fetch user info from Google
	client := googleOAuth.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to fetch user info: %s"}`, err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var gUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gUser); err != nil {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf(`{"error":"failed to parse user info: %s"}`, string(body)), http.StatusInternalServerError)
		return
	}

	// Find or create user
	var userID, role, status string
	err = pool.QueryRow(r.Context(),
		`SELECT id, role, status FROM users WHERE google_id=$1 OR email=$2 LIMIT 1`,
		gUser.ID, gUser.Email,
	).Scan(&userID, &role, &status)

	if err != nil {
		// Create new user
		parts := splitName(gUser.Name)
		err = pool.QueryRow(r.Context(),
		`INSERT INTO users (email, password_hash, first_name, last_name, role, status, google_id, picture) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id, role, status`,
		gUser.Email, "", parts[0], parts[1], "client", "active", gUser.ID, gUser.Picture,
		).Scan(&userID, &role, &status)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"failed to create user: %s"}`, err), http.StatusInternalServerError)
			return
		}
	} else {
		// Update login time and picture
		pool.Exec(r.Context(), `UPDATE users SET last_login_at=NOW(), picture=$1, google_id=$2 WHERE id=$3`, gUser.Picture, gUser.ID, userID)
	}

	parts := splitName(gUser.Name)
	jwtToken, err := generateTokenWithClaims(userID, gUser.Email, gUser.Name, role, gUser.Picture)
	_ = parts
	if err != nil {
		http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	// Redirect to frontend with token
	redirectURL := os.Getenv("FRONTEND_URL")
	if redirectURL == "" {
		redirectURL = "https://cutthroatreels.com"
	}
	http.Redirect(w, r, redirectURL+"/auth/google-callback?token="+jwtToken, http.StatusTemporaryRedirect)
}

func randState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func splitName(fullName string) [2]string {
	if fullName == "" {
		return [2]string{"", ""}
	}
	parts := [2]string{"", ""}
	for i := len(fullName) - 1; i >= 0; i-- {
		if fullName[i] == ' ' {
			parts[0] = fullName[:i]
			parts[1] = fullName[i+1:]
			return parts
		}
	}
	parts[0] = fullName
	return parts
}
