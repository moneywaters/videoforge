package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/videoforge/backend/svc-user/internal/model"
	"github.com/videoforge/backend/svc-user/internal/service"

	"github.com/videoforge/backend/pkg/errors"
)

// UserHandler handles user HTTP requests
type UserHandler struct {
	service service.UserServiceInterface
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(service service.UserServiceInterface) *UserHandler {
	return &UserHandler{service: service}
}

// GetMe returns the current user profile
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Get userID from context (set by auth middleware)
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		errors.WriteError(r.Context(), w, errors.Unauthorized("unauthorized"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid user ID"))
		return
	}

	user, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, user)
}

// UpdateMe updates the current user profile
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	// Get userID from context (set by auth middleware)
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		errors.WriteError(r.Context(), w, errors.Unauthorized("unauthorized"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid user ID"))
		return
	}

	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	user, err := h.service.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, user)
}

// respond writes a success response
func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}