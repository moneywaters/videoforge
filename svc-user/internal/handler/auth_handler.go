package handler

import (
	"encoding/json"
	"net/http"

	"svc-user/internal/model"
	"svc-user/internal/service"

	"github.com/videoforge/backend/pkg/errors"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	service service.AuthServiceInterface
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(service service.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	user, err := h.service.Register(r.Context(), req)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusCreated, user)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	response, err := h.service.Login(r.Context(), req)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, response)
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	if req.RefreshToken == "" {
		errors.WriteError(r.Context(), w, errors.BadRequest("refresh token is required"))
		return
	}

	response, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, response)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get userID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		errors.WriteError(r.Context(), w, errors.Unauthorized("unauthorized"))
		return
	}

	var req model.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	if req.RefreshToken == "" {
		errors.WriteError(r.Context(), w, errors.BadRequest("refresh token is required"))
		return
	}

	// Parse userID
	// In a real implementation, we'd use the proper UUID type
	// For now, we'll just pass an empty UUID if parsing fails
	if err := h.service.Logout(r.Context(), parseUUID(userID), req.RefreshToken); err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusNoContent, nil)
}

// parseUUID parses a UUID string (placeholder implementation)
func parseUUID(s string) interface{} {
	// This is a placeholder - in real implementation we'd use google/uuid
	return nil
}

// respond writes a success response
func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}