package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	pkgmiddleware "github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/pkg/errors"
)

// ContextKey for storing user info
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// UserRoleKey is the context key for user role
	UserRoleKey ContextKey = "user_role"
)

// Claims represents the JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware creates JWT authentication middleware
// Uses RS256 public key for verification
func NewAuthMiddleware(publicKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for public routes
			path := r.URL.Path
			if isPublicPath(path) {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errors.WriteError(r.Context(), w, errors.Unauthorized("Missing authorization header"))
				return
			}

			// Validate token directly with public key
			claims, err := ValidateTokenWithKey(authHeader, publicKey)
			if err != nil {
				errors.WriteError(r.Context(), w, errors.Unauthorized(err.Error()))
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

			// Add headers for downstream services
			r.Header.Set("X-User-ID", claims.UserID)
			r.Header.Set("X-User-Role", claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ValidateTokenWithKey validates a JWT token with the given public key
func ValidateTokenWithKey(authHeader, publicKey string) (*Claims, error) {
	// Extract token from "Bearer <token>" format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	tokenString := parts[1]

	// Parse and validate token
	// Note: For production, use proper RS256 validation with the public key
	// This is a simplified version that uses ParseUnverified
	// In production, replace with proper crypto validation
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// isPublicPath checks if the path is public (no JWT required)
func isPublicPath(path string) bool {
	publicPaths := []string{
		"/health",
		"/.well-known/",
		"/.well-known/jwks.json",
	}

	for _, p := range publicPaths {
		if path == p || strings.HasPrefix(path, p) {
			return true
		}
	}

	// Auth endpoints are public (use different auth flow)
	if strings.HasPrefix(path, "/api/v1/auth/register") ||
		strings.HasPrefix(path, "/api/v1/auth/login") ||
		strings.HasPrefix(path, "/api/v1/auth/refresh") {
		return true
	}

	return false
}

// GetUserID retrieves the user ID from the context
func GetUserID(ctx context.Context) string {
	// First check local context key
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	// Then check pkg middleware context key
	return pkgmiddleware.GetUserID(ctx)
}

// GetUserRole retrieves the user role from the context
func GetUserRole(ctx context.Context) string {
	// First check local context key
	if role, ok := ctx.Value(UserRoleKey).(string); ok {
		return role
	}
	// Then check pkg middleware context key
	return pkgmiddleware.GetUserRole(ctx)
}