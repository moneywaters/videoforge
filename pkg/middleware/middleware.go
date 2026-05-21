package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/videoforge/backend/pkg/logger"
)

// RequestIDKey is the context key for the request ID.
const RequestIDKey = "request_id"

// RequestID middleware adds a unique request ID to the request context.
// The request ID is added as a UUID and can be used for tracing requests
// across services and logs.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID is already set in header
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		// Add to response header
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID Retrieves the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// Recover middleware recovers from panics in the handler.
// If a panic occurs, it logs the error and returns a 500 Internal Server Error.
// This prevents the entire server from crashing due to a single request.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a custom response writer to capture the status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		defer func() {
			if recovered := recover(); recovered != nil {
				// Log the panic
				requestID := GetRequestID(r.Context())
				slog.Error("panic recovered", slog.String("request_id", requestID), slog.Any("error", recovered))

				// Write the 500 response
				wrapper.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(wrapper, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	// Ensure status code is set if not already
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

// Logger middleware logs the incoming request and response details.
// It logs method, path, status code, and duration.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get logger from context if available
		ctx := r.Context()
		log := logger.FromContext(ctx)
		if log == nil {
			log = logger.Default(os.Getenv("ENVIRONMENT"))
		}

		requestID := GetRequestID(ctx)
		if requestID == "" {
			requestID = "none"
		}

		// Wrap response writer to capture status code
		wrapper := &responseWriter{ResponseWriter: w}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)

		// Log the request details
		log.Info("HTTP request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("query", r.URL.RawQuery),
			slog.Int("status", wrapper.statusCode),
			slog.Duration("duration", duration),
			slog.String("request_id", requestID),
		)
	})
}

// CORS middleware adds cross-origin headers and handles preflight.
// It mirrors the request's Origin when allowAll=true (required for credentialed requests).
func CORS(origin string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestOrigin := r.Header.Get("Origin")

			// Determine allowed origin: echo caller's Origin if wildcard mode.
			// Browsers reject ‘*’ when credentials are used (e.g. Authorization header).
			allowedOrigin := origin
			if origin == "*" || origin == "" {
				allowedOrigin = requestOrigin
			}

			if allowedOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
				w.Header().Add("Vary", "Origin")
			}

			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID, X-Requested-With")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Chain applies multiple middleware in sequence.
// It returns a handler that applies each middleware from left to right.
func Chain(next http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		next = middlewares[i](next)
	}
	return next
}

// ContentTypeJSON sets the Content-Type header to application/json.
func ContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// MaxBodySize limits the request body size.
// It returns 413 Payload Too Large if the body exceeds the limit.
func MaxBodySize(maxBytes int64) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				w.Header().Set("Content-Type", "application/problem+json")
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				w.Write([]byte(`{"title":"Payload Too Large","status":413,"detail":"Request body exceeds maximum allowed size"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Headers adds additional response headers.
func Headers(headers map[string]string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for key, value := range headers {
				w.Header().Set(key, value)
			}
			next.ServeHTTP(w, r)
		})
	}
}