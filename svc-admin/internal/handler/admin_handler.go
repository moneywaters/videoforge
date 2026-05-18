package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"svc-admin/internal/model"
	"svc-admin/internal/service"

	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/pkg/middleware"
)

// AdminHandler handles admin HTTP requests
type AdminHandler struct {
	service service.AdminServiceInterface
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(service service.AdminServiceInterface) *AdminHandler {
	return &AdminHandler{service: service}
}

// ListUsers handles GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	req := model.UserListRequest{
		Role:    r.URL.Query().Get("role"),
		Status:  r.URL.Query().Get("status"),
		Email:   r.URL.Query().Get("email"),
		Limit:   20,
		Offset:  0,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	result, err := h.service.SearchUsers(r.Context(), req)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

// GetUser handles GET /api/v1/admin/users/:id
func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(getPathVar(r, "id"))
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid user ID"))
		return
	}

	result, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

// BanUser handles POST /api/v1/admin/users/:id/ban
func (h *AdminHandler) BanUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(getPathVar(r, "id"))
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid user ID"))
		return
	}

	var req model.BanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	adminID, err := uuid.Parse(middleware.GetUserID(r.Context()))
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Unauthorized("invalid admin ID"))
		return
	}

	if err := h.service.BanUser(r.Context(), adminID, userID, req); err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, map[string]string{
		"message": "user banned",
		"user_id": userID.String(),
	})
}

// UnbanUser handles POST /api/v1/admin/users/:id/unban
func (h *AdminHandler) UnbanUser(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(getPathVar(r, "id"))
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid user ID"))
		return
	}

	var req model.UnbanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	adminID, err := uuid.Parse(middleware.GetUserID(r.Context()))
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Unauthorized("invalid admin ID"))
		return
	}

	if err := h.service.UnbanUser(r.Context(), adminID, userID, req); err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, map[string]string{
		"message": "user unbanned",
		"user_id": userID.String(),
	})
}

// AssignRoles handles POST /api/v1/admin/users/:id/roles
func (h *AdminHandler) AssignRoles(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(getPathVar(r, "id"))
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid user ID"))
		return
	}

	var req model.AssignRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	adminID, err := uuid.Parse(middleware.GetUserID(r.Context()))
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Unauthorized("invalid admin ID"))
		return
	}

	if err := h.service.AssignRoles(r.Context(), adminID, userID, req); err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, map[string]interface{}{
		"message": "roles assigned",
		"user_id": userID.String(),
		"roles":   req.Roles,
	})
}

// ListDisputes handles GET /api/v1/admin/disputes
func (h *AdminHandler) ListDisputes(w http.ResponseWriter, r *http.Request) {
	req := model.DisputesListRequest{
		Status: r.URL.Query().Get("status"),
		Limit:  20,
		Offset: 0,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	result, err := h.service.ListDisputes(r.Context(), req)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

// GetDispute handles GET /api/v1/admin/disputes/:id
func (h *AdminHandler) GetDispute(w http.ResponseWriter, r *http.Request) {
	disputeID, err := uuid.Parse(getPathVar(r, "id"))
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid dispute ID"))
		return
	}

	result, err := h.service.GetDisputeByID(r.Context(), disputeID)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

// ResolveDispute handles POST /api/v1/admin/disputes/:id/resolve
func (h *AdminHandler) ResolveDispute(w http.ResponseWriter, r *http.Request) {
	disputeID, err := uuid.Parse(getPathVar(r, "id"))
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid dispute ID"))
		return
	}

	var req model.ResolveDisputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	adminID, err := uuid.Parse(middleware.GetUserID(r.Context()))
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Unauthorized("invalid admin ID"))
		return
	}

	if err := h.service.ResolveDispute(r.Context(), adminID, disputeID, req); err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, map[string]string{
		"message":    "dispute resolved",
		"dispute_id": disputeID.String(),
	})
}

// ListModerationQueue handles GET /api/v1/admin/moderation-queue
func (h *AdminHandler) ListModerationQueue(w http.ResponseWriter, r *http.Request) {
	req := model.ModerationQueueListRequest{
		Status: r.URL.Query().Get("status"),
		Limit:  20,
		Offset: 0,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	result, err := h.service.ListModerationQueue(r.Context(), req)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

// ReviewModerationItem handles POST /api/v1/admin/moderation-queue/:id/review
func (h *AdminHandler) ReviewModerationItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(getPathVar(r, "id"))
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid moderation item ID"))
		return
	}

	var req model.ReviewModerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	adminID, err := uuid.Parse(middleware.GetUserID(r.Context()))
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Unauthorized("invalid admin ID"))
		return
	}

	if err := h.service.ReviewModerationItem(r.Context(), adminID, itemID, req); err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, map[string]string{
		"message": "moderation item reviewed",
		"item_id": itemID.String(),
	})
}

// OverridePayout handles POST /api/v1/admin/payouts/:id/override
func (h *AdminHandler) OverridePayout(w http.ResponseWriter, r *http.Request) {
	payoutID, err := uuid.Parse(getPathVar(r, "id"))
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid payout ID"))
		return
	}

	var req model.OverridePayoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}

	adminID, err := uuid.Parse(middleware.GetUserID(r.Context()))
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Unauthorized("invalid admin ID"))
		return
	}

	if err := h.service.OverridePayout(r.Context(), adminID, payoutID, req); err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, map[string]interface{}{
		"message":   "payout overridden",
		"payout_id": payoutID.String(),
		"amount":   req.NewAmount,
	})
}

// ListAdminActions handles GET /api/v1/admin/actions
func (h *AdminHandler) ListAdminActions(w http.ResponseWriter, r *http.Request) {
	req := model.ActionsListRequest{
		AdminID:    r.URL.Query().Get("admin_id"),
		ActionType: r.URL.Query().Get("action_type"),
		Limit:     20,
		Offset:    0,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	result, err := h.service.ListAdminActions(r.Context(), req)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	respond(w, http.StatusOK, result)
}

// getPathVar extracts a path variable from the URL path
// Expected path format: /api/v1/admin/{resource}/:id or /api/v1/admin/{resource}/:id/{action}
func getPathVar(r *http.Request, name string) string {
	path := r.URL.Path

	// Split the path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == ":"+name && i > 0 {
			return parts[i-1]
		}
	}

	// If we can't find a path param format, try to extract from the end
	// This is a fallback - the actual implementation should use a router
	if name == "id" {
		// Find the last UUID-like segment
		for i := len(parts) - 1; i >= 0; i-- {
			if _, err := uuid.Parse(parts[i]); err == nil {
				return parts[i]
			}
		}
	}

	return ""
}

// respond writes a JSON response
func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}