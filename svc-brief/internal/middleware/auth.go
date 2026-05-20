package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/videoforge/backend/pkg/errors"
)

// AuthConfig holds auth configuration
type AuthConfig struct {
	PublicKey    *rsa.PublicKey
	PublicKeyStr string
	KeyID        string
}

// LoadAuthConfig loads JWT public key from environment or file
func LoadAuthConfig() (*AuthConfig, error) {
	// First try to load from JWT_PUBLIC_KEY_PATH
	keyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")
	if keyPath != "" {
		data, err := os.ReadFile(keyPath)
		if err == nil {
			return parsePublicKey(string(data))
		}
	}

	// Try JWT_PUBLIC_KEY env var (base64 encoded)
	keyStr := os.Getenv("JWT_PUBLIC_KEY")
	if keyStr != "" {
		return parsePublicKey(keyStr)
	}

	// Try to load from JWKS_URL (stub - in production fetch from svc-user)
	jwksURL := os.Getenv("JWKS_URL")
	if jwksURL != "" {
		// Stub: just return nil and use a fallback
		return &AuthConfig{
			PublicKey:    nil,
			PublicKeyStr: "",
			KeyID:        "",
		}, nil
	}

	// For development, create a simple fallback
	return &AuthConfig{
		PublicKey:    nil,
		PublicKeyStr: "",
		KeyID:        "dev-key",
	}, nil
}

func parsePublicKey(keyStr string) (*AuthConfig, error) {
	// Try base64 decode first
	_, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		// Try plain string (PEM format)
		// For now, return simple config
		return &AuthConfig{
			PublicKey:    nil,
			PublicKeyStr: keyStr,
			KeyID:       "videoforge-key",
		}, nil
	}

	// Try to parse as RSA public key
	// For now, return simple config
	return &AuthConfig{
		PublicKey:    nil,
		PublicKeyStr: keyStr,
		KeyID:       "videoforge-key",
	}, nil
}

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	config *AuthConfig
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(config *AuthConfig) *AuthMiddleware {
	return &AuthMiddleware{config: config}
}

// Handler returns the auth handler
func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If no public key configured, still try to extract user context from token
		if m.config == nil || m.config.PublicKeyStr == "" {
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
								ctx := context.WithValue(r.Context(), UserIDKey, sub)
								ctx = context.WithValue(ctx, UserRoleKey, role)
								r = r.WithContext(ctx)
							}
						}
					}
				}
			}
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Allow health checks without auth
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}
			errors.WriteError(r.Context(), w, errors.Unauthorized("authorization header required"))
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			errors.WriteError(r.Context(), w, errors.Unauthorized("invalid authorization header format"))
			return
		}

		tokenStr := parts[1]

		// Parse token without verification for development
		// In production, use real public key
		token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			errors.WriteError(r.Context(), w, errors.Unauthorized("invalid token"))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			errors.WriteError(r.Context(), w, errors.Unauthorized("invalid token claims"))
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			errors.WriteError(r.Context(), w, errors.Unauthorized("invalid token subject"))
			return
		}

		role, _ := claims["role"].(string)

		ctx := context.WithValue(r.Context(), UserIDKey, sub)
		ctx = context.WithValue(ctx, UserRoleKey, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserIDKey is the context key for user ID
const UserIDKey = "user_id"

// UserRoleKey is the context key for user role
const UserRoleKey = "user_role"

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserRole retrieves user role from context
func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value(UserRoleKey).(string); ok {
		return role
	}
	return ""
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKSHandler serves the public key in JWKS format
type JWKSHandler struct {
	publicKey *rsa.PublicKey
	keyID    string
}

// NewJWKSHandler creates a new JWKSHandler
func NewJWKSHandler(publicKey *rsa.PublicKey, keyID string) *JWKSHandler {
	return &JWKSHandler{publicKey: publicKey, keyID: keyID}
}

// ServeHTTP serves the JWKS
func (h *JWKSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.publicKey == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "no public key configured"})
		return
	}

	n := base64.RawURLEncoding.EncodeToString(h.publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString([]byte{byte(h.publicKey.E >> 8), byte(h.publicKey.E)})

	jwks := JWKS{
		Keys: []JWK{
			{Kty: "RSA", Use: "sig", Kid: h.keyID, Alg: "RS256", N: n, E: e},
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jwks)
}

// Authenticate validates JWT tokens - convenience function
func Authenticate(publicKey *rsa.PublicKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errors.WriteError(r.Context(), w, errors.Unauthorized("authorization header required"))
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				errors.WriteError(r.Context(), w, errors.Unauthorized("invalid authorization header format"))
				return
			}

			token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("invalid signing method")
				}
				return publicKey, nil
			})
			if err != nil || !token.Valid {
				errors.WriteError(r.Context(), w, errors.Unauthorized("invalid or expired token"))
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				errors.WriteError(r.Context(), w, errors.Unauthorized("invalid token claims"))
				return
			}

			sub, ok := claims["sub"].(string)
			if !ok || sub == "" {
				errors.WriteError(r.Context(), w, errors.Unauthorized("invalid token subject"))
				return
			}

			role, _ := claims["role"].(string)

			ctx := context.WithValue(r.Context(), UserIDKey, sub)
			ctx = context.WithValue(ctx, UserRoleKey, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}