package middleware

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/videoforge/backend/pkg/errors"
)

// Claims represents JWT claims
type Claims struct {
	jwt.RegisteredClaims
	UserID   string   `json:"sub"`
	Email   string   `json:"email"`
	Role    string   `json:"role"`
	ClientID string  `json:"client_id"`
}

// AuthMiddleware provides JWT authentication
type AuthMiddleware struct {
	publicKey *rsa.PublicKey
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(publicKey *rsa.PublicKey) *AuthMiddleware {
	return &AuthMiddleware{publicKey: publicKey}
}

// Authenticate authenticates requests using JWT
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Skip auth for public endpoints
		path := r.URL.Path
		if isPublicEndpoint(path) {
			next.ServeHTTP(w, r)
			return
		}

		// Get token from header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errors.WriteError(ctx, w, errors.Unauthorized("authorization header required"))
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			errors.WriteError(ctx, w, errors.Unauthorized("invalid authorization header format"))
			return
		}
		tokenString := parts[1]

		// Parse and validate token
		claims, err := m.validateToken(tokenString)
		if err != nil {
			errors.WriteError(ctx, w, errors.Unauthorized(fmt.Sprintf("invalid token: %v", err)))
			return
		}

		// Add claims to context
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "user_role", claims.Role)
		ctx = context.WithValue(ctx, "client_id", claims.ClientID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateToken validates a JWT token and returns claims
func (m *AuthMiddleware) validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Check expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

// isPublicEndpoint checks if the endpoint is public (no auth required)
func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/health",
		"/api/v1/shopify/webhook",
	}

	for _, p := range publicPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	return false
}

// GetUserID retrieves the user ID from context
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value("user_id").(string); ok {
		return id
	}
	return ""
}

// GetUserRole retrieves the user role from context
func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value("user_role").(string); ok {
		return role
	}
	return ""
}

// GetClientID retrieves the client ID from context
func GetClientID(ctx context.Context) string {
	if id, ok := ctx.Value("client_id").(string); ok {
		return id
	}
	return ""
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userRole := GetUserRole(ctx)

			if userRole == "" {
				errors.WriteError(ctx, w, errors.Forbidden("role required"))
				return
			}

			if userRole != "admin" && userRole != role {
				errors.WriteError(ctx, w, errors.Forbidden("insufficient permissions"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireClient creates a middleware that ensures client ID is present
func RequireClient(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		clientID := GetClientID(ctx)

		if clientID == "" {
			errors.WriteError(ctx, w, errors.Forbidden("client access required"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Ensure UUID validity
var _ = uuid.MustParse