package router

import (
	"context"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/svc-gateway/internal/handler"
	"github.com/videoforge/backend/svc-gateway/internal/middleware"
)

// ServiceRoute maps service name to URL
type ServiceRoute struct {
	Name string
	URL  string
}

// Router holds the router configuration
type Router struct {
	serviceRoutes map[string]string
	proxyHandler *handler.ProxyHandler
	jwtPublicKey string
}

// NewRouter creates a new Router
func NewRouter(serviceRoutes []ServiceRoute, jwtPublicKey string) *Router {
	// Convert service routes to map
	routesMap := make(map[string]string)
	for _, sr := range serviceRoutes {
		routesMap[sr.Name] = sr.URL
	}

	r := &Router{
		serviceRoutes: routesMap,
		jwtPublicKey: jwtPublicKey,
	}

	// Initialize handlers
	r.proxyHandler = handler.NewProxyHandler(routesMap)

	return r
}

// HandleRequest handles API requests by proxying to internal services
// This is registered as the handler for /api/v1/* routes
func (r *Router) HandleRequest(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	// Extract service name from path
	// Format: /api/v1/service_name/...
	parts := strings.SplitN(path, "/", 4)
	if len(parts) < 3 {
		http.NotFound(w, req)
		return
	}

	serviceName := parts[2] // "users", "briefs", etc.
	baseURL, ok := r.serviceRoutes[serviceName]
	if !ok {
		// Service not found
		errors.WriteError(req.Context(), w, errors.NotFound("service not found: "+serviceName))
		return
	}

	// Check if this is a public route
	isPublic := isPublicRoute(path)

	// For protected routes, we need to ensure user is authenticated
	// The JWT middleware should have already set the user context
	// But we extract user info here to forward to internal services
	if !isPublic {
		userID := middleware.GetUserID(req.Context())
		role := middleware.GetUserRole(req.Context())

		if userID == "" {
			// Not authenticated - return 401
			errors.WriteError(req.Context(), w, errors.Unauthorized("authentication required"))
			return
		}

		// Forward user info in headers
		req.Header.Set("X-User-ID", userID)
		req.Header.Set("X-User-Role", role)
	}

	// Build target URL
	targetURL := baseURL + path

	// For /api/v1 stripping
	targetURL = baseURL + strings.TrimPrefix(path, "/api/v1")

	r.proxyHandler.ProxyTo(targetURL, w, req)
}

// isPublicRoute checks if the route is public (no JWT required)
func isPublicRoute(path string) bool {
	// Public paths that don't require authentication
	publicPrefixes := []string{
		"/api/v1/auth/",        // Auth endpoints
		"/api/v1/users/me",    // User profile (requires auth actually)
		"/.well-known/",      // JWKS
		"/health",           // Health
	}

	for _, prefix := range publicPrefixes {
		if strings.HasPrefix(path, prefix) {
			// But /api/v1/auth/register and /api/v1/auth/login are truly public
			if prefix == "/api/v1/auth/" {
				authOnlyPaths := []string{
					"/api/v1/auth/register",
					"/api/v1/auth/login",
					"/api/v1/auth/refresh",
				}
				for _, p := range authOnlyPaths {
					if path == p || path == p+"/" {
						return true
					}
				}
			}
			// /health is public
			if path == "/health" || path == "/health/" {
				return true
			}
		}
	}

	return false
}

// GetServiceURL returns the service URL for a given service name
func (r *Router) GetServiceURL(serviceName string) string {
	return r.serviceRoutes[serviceName]
}

// ServiceRoutes returns the default service route mappings
func ServiceRoutes() []ServiceRoute {
	return []ServiceRoute{
		{Name: "auth", URL: "http://svc-user:8080"},
		{Name: "users", URL: "http://svc-user:8080"},
		{Name: "briefs", URL: "http://svc-brief:8080"},
		{Name: "videos", URL: "http://svc-video:8080"},
		{Name: "campaigns", URL: "http://svc-campaign:8080"},
		{Name: "shopify", URL: "http://svc-shopify:8080"},
		{Name: "performance", URL: "http://svc-performance:8080"},
		{Name: "payouts", URL: "http://svc-payout:8080"},
		{Name: "notifications", URL: "http://svc-notification:8080"},
		{Name: "admin", URL: "http://svc-admin:8080"},
		{Name: "support", URL: "http://svc-ai-support:8080"},
	}
}

// NewReverseProxy creates a new reverse proxy for the given URL
func NewReverseProxy(target string) *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(target)
}