package middleware

import (
	"net/http"

	"github.com/videoforge/backend/pkg/middleware"
)

// AdminMiddleware handles admin-specific middleware
type AdminMiddleware struct {
	authMiddleware *middleware.AuthMiddleware
}

// NewAdminMiddleware creates a new AdminMiddleware
func NewAdminMiddleware(authMiddleware *middleware.AuthMiddleware) *AdminMiddleware {
	return &AdminMiddleware{authMiddleware: authMiddleware}
}

// RequireAdmin requires admin role for access
func (m *AdminMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := middleware.GetUserRole(r.Context())
		if role != "admin" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePermission requires a specific permission (stub implementation)
func (m *AdminMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For now, just check if user is admin
			role := middleware.GetUserRole(r.Context())
			if role != "admin" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}