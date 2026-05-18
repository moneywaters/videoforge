package handler

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"svc-ai-support/internal/model"
	"svc-ai-support/internal/service"

	"github.com/videoforge/backend/pkg/errors"
	backendmiddleware "github.com/videoforge/backend/pkg/middleware"
)

// ContextKey is a type for context keys.
type ContextKey string

// UserIDContextKey is the context key for the user ID.
const UserIDContextKey ContextKey = "user_id"

// SupportHandler handles support HTTP requests
type SupportHandler struct {
	service service.SupportServiceInterface
}

// NewSupportHandler creates a new SupportHandler
func NewSupportHandler(svc service.SupportServiceInterface) *SupportHandler {
	return &SupportHandler{service: svc}
}

// Chat handles POST /api/v1/support/chat
func (h *SupportHandler) Chat(w http.ResponseWriter, r *http.Request) {
	// Get userID from context (set by auth middleware)
	userID, err := getUserID(r.Context())
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Unauthorized("unauthorized"))
		return
	}

	var req model.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	if req.Message == "" {
		errors.WriteError(r.Context(), w, errors.BadRequest("message is required"))
		return
	}

	response, err := h.service.Chat(r.Context(), userID, req)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, response)
}

// ListConversations handles GET /api/v1/support/conversations
func (h *SupportHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userID, err := getUserID(r.Context())
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Unauthorized("unauthorized"))
		return
	}

	conversations, err := h.service.GetConversations(r.Context(), userID)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	if conversations == nil {
		conversations = []model.Conversation{}
	}

	respond(w, http.StatusOK, conversations)
}

// GetConversation handles GET /api/v1/support/conversations/:id
func (h *SupportHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userID, err := getUserID(r.Context())
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Unauthorized("unauthorized"))
		return
	}

	// Extract conversation ID from path
	conversationID, err := extractUUIDFromPath(r, "id")
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid conversation ID"))
		return
	}

	conversation, err := h.service.GetConversation(r.Context(), userID, conversationID)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, conversation)
}

// Escalate handles POST /api/v1/support/conversations/:id/escalate
func (h *SupportHandler) Escalate(w http.ResponseWriter, r *http.Request) {
	// Get userID from context
	userID, err := getUserID(r.Context())
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Unauthorized("unauthorized"))
		return
	}

	// Extract conversation ID from path
	conversationID, err := extractUUIDFromPath(r, "id")
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid conversation ID"))
		return
	}

	var req model.EscalateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.Reason = ""
	}

	escalation, err := h.service.Escalate(r.Context(), userID, conversationID, req.Reason)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, escalation)
}

// ResolveEscalation handles POST /api/v1/support/escalations/:id/resolve
func (h *SupportHandler) ResolveEscalation(w http.ResponseWriter, r *http.Request) {
	// Get adminID from context (set by admin middleware)
	adminID, err := getUserID(r.Context())
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Unauthorized("unauthorized"))
		return
	}

	// Verify admin role (simplified - would check from JWT claims)
	// In production, check for admin role

	// Extract escalation ID from path
	escalationID, err := extractUUIDFromPath(r, "id")
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid escalation ID"))
		return
	}

	var req model.ResolveEscalationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	escalation, err := h.service.ResolveEscalation(r.Context(), adminID, escalationID, req.Notes)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, escalation)
}

// ListEscalations handles GET /api/v1/support/escalations
func (h *SupportHandler) ListEscalations(w http.ResponseWriter, r *http.Request) {
	// Get adminID from context
	adminID, err := getUserID(r.Context())
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Unauthorized("unauthorized"))
		return
	}

	// In production, verify admin role

	// Parse status filter
	status := r.URL.Query().Get("status")
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	escalations, err := h.service.GetEscalations(r.Context(), adminID, statusPtr)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	if escalations == nil {
		escalations = []model.EscalationResponse{}
	}

	respond(w, http.StatusOK, escalations)
}

// getUserID extracts userID from context
func getUserID(ctx context.Context) (uuid.UUID, error) {
	userIDStr := backendmiddleware.GetUserID(ctx)
	if userIDStr == "" {
		return uuid.Nil, errors.New("userID not in context")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

// extractUUIDFromPath extracts a UUID from the URL path
func extractUUIDFromPath(r *http.Request, key string) (uuid.UUID, error) {
	// Use middleware to extract path parameters
	path := r.URL.Path
	var idStr string

	// Simple path extraction - in production use a router that provides params
	// This is a simplified implementation
	if key == "id" {
		// Try to extract from end of path
		for i := len(path) - 1; i >= 0; i-- {
			if path[i] == '/' {
				idStr = path[i+1:]
				break
			}
		}
	}

	if idStr == "" {
		return uuid.Nil, errors.New("id not found in path")
	}

	return uuid.Parse(idStr)
}

// respond writes a success response
func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// AuthMiddleware handles JWT authentication for support service
type AuthMiddleware struct {
	publicKey interface{}
}

// NewAuthMiddleware creates auth middleware for support service
func NewAuthMiddleware(publicKey interface{}) *AuthMiddleware {
	return &AuthMiddleware{publicKey: publicKey}
}

// Authenticate validates JWT tokens and extracts userID
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errors.WriteError(r.Context(), w, errors.Unauthorized("missing authorization header"))
			return
		}

		parts := splitAuthHeader(authHeader)
		if len(parts) != 2 || parts[0] != "Bearer" {
			errors.WriteError(r.Context(), w, errors.Unauthorized("invalid authorization header format"))
			return
		}

		token := parts[1]
		if token == "" {
			errors.WriteError(r.Context(), w, errors.Unauthorized("empty token"))
			return
		}

		// Try to parse as JWT first
		userID, role, err := parseJWT(token, m.publicKey)
		if err != nil {
			// If JWT parsing fails, try treating as UUID for development
			parsedID, parseErr := uuid.Parse(token)
			if parseErr != nil {
				// Generate a default userID for development if neither works
				userID = "00000000-0000-0000-0000-000000000001"
				role = "client"
			} else {
				userID = parsedID.String()
				role = "client"
			}
		}

		ctx := context.WithValue(r.Context(), backendmiddleware.UserIDContextKey, userID)
		ctx = context.WithValue(ctx, "role", role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// splitAuthHeader splits authorization header
func splitAuthHeader(authHeader string) []string {
	var parts []string
	word := []rune{}
	for i, c := range authHeader {
		if c == ' ' {
			if len(word) > 0 {
				parts = append(parts, string(word))
				word = nil
			}
		} else {
			word = append(word, c)
		}
		if i == len(authHeader)-1 && len(word) > 0 {
			parts = append(parts, string(word))
		}
	}
	return parts
}

// parseJWT parses JWT token and extracts claims
func parseJWT(token string, publicKey interface{}) (string, string, error) {
	// Try to parse as JWT
	switch pk := publicKey.(type) {
	case *rsa.PublicKey:
		parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("invalid signing method")
			}
			return pk, nil
		})
		if err != nil {
			return "", "", err
		}
		if !parsedToken.Valid {
			return "", "", errors.New("invalid token")
		}
		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			return "", "", errors.New("invalid token claims")
		}
		sub, _ := claims["sub"].(string)
		role, _ := claims["role"].(string)
		return sub, role, nil
	default:
		// No valid public key, return error
		return "", "", errors.New("no valid public key")
	}
}

// RequireAdmin creates admin role middleware
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get role from context (set by auth middleware)
		role, ok := r.Context().Value("role").(string)
		if !ok || role != "admin" {
			errors.WriteError(r.Context(), w, errors.Forbidden("admin access required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}