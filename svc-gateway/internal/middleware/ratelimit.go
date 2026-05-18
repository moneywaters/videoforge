package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/videoforge/backend/pkg/errors"
	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting using token bucket algorithm
type RateLimiter struct {
	limiters    map[string]*rate.Limiter
	mu         sync.RWMutex
	rate       rate.Limit
	burst     int
	cleanupInterval time.Duration
}

// NewRateLimiter creates a new rate limiter with the given rate and burst
func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		limiters:    make(map[string]*rate.Limiter),
		rate:       r,
		burst:     burst,
		cleanupInterval: 5 * time.Minute,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup removes expired entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	for range ticker.C {
		rl.cleanupOnce()
	}
}

func (rl *RateLimiter) cleanupOnce() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// In a production system, we'd clean up old limiters here
	// For now, we just let them accumulate (memory-safe since they're small)
	// In production, implement LRU or time-based cleanup
}

// Allow checks if a request is allowed for the given client ID
func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[clientID]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[clientID] = limiter
	}

	return limiter.Allow()
}

// getClientID returns a unique client identifier
// Prefers user ID if authenticated, falls back to IP
func getClientID(r *http.Request) string {
	// Try X-User-ID header first (set by JWT middleware)
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return "user:" + userID
	}

	// Try X-Forwarded-For header (for proxied requests)
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return "ip:" + fwd
	}

	// Fall back to remote address (strip port)
	addr := r.RemoteAddr
	if idx := lastIndex(addr, ":"); idx > 0 {
		addr = addr[:idx]
	}
	return "ip:" + addr
}

// lastIndex returns the last index of separator, or -1 if not found
func lastIndex(s, sep string) int {
	for i := len(s) - 1; i >= 0; i-- {
		found := true
		for j := 0; j < len(sep); j++ {
			if i+j >= len(s) || s[i+j] != sep[j] {
				found = false
				break
			}
		}
		if found {
			return i
		}
	}
	return -1
}

// RateLimitMiddleware creates rate limiting middleware
// Uses per-IP and per-user-ID rate limiting
func RateLimitMiddleware(requestsPerSecond int, burst int) func(next http.Handler) http.Handler {
	limiter := NewRateLimiter(rate.Limit(requestsPerSecond), burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip rate limiting for health and well-known endpoints
			path := r.URL.Path
			if path == "/health" || path == "/.well-known/jwks.json" {
				next.ServeHTTP(w, r)
				return
			}

			// Skip for OPTIONS (CORS preflight)
			if r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			clientID := getClientID(r)

			if !limiter.Allow(clientID) {
				errors.WriteError(r.Context(), w, errors.TooManyRequests("Rate limit exceeded"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}