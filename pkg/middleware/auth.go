package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/videoforge/backend/pkg/errors"
)

// ContextKey is a type for context keys.
type ContextKey string

// UserIDContextKey is the context key for the user ID.
const UserIDContextKey ContextKey = "user_id"

// UserRoleContextKey is the context key for the user role.
const UserRoleContextKey ContextKey = "user_role"

// AuthMiddleware validates JWT tokens
type AuthMiddleware struct {
	publicKey *rsa.PublicKey
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(publicKey *rsa.PublicKey) *AuthMiddleware {
	return &AuthMiddleware{publicKey: publicKey}
}

// Authenticate validates JWT tokens from Authorization header
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
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
			return m.publicKey, nil
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

		ctx := context.WithValue(r.Context(), UserIDContextKey, sub)
		ctx = context.WithValue(ctx, UserRoleContextKey, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
	keyID     string
}

// NewJWKSHandler creates a new JWKSHandler
func NewJWKSHandler(publicKey *rsa.PublicKey, keyID string) *JWKSHandler {
	return &JWKSHandler{publicKey: publicKey, keyID: keyID}
}

// ServeHTTP serves the JWKS
func (h *JWKSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Convert RSA public key to JWKS format
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

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDContextKey).(string); ok {
		return id
	}
	return ""
}

// GetUserRole retrieves user role from context
func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value(UserRoleContextKey).(string); ok {
		return role
	}
	return ""
}