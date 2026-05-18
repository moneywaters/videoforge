package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/pkg/middleware"

	"svc-brief/internal/model"
	"svc-brief/internal/service"
)

// BriefHandler handles brief HTTP requests
type BriefHandler struct {
	service service.BriefServiceInterface
}

// NewBriefHandler creates a new BriefHandler
func NewBriefHandler(svc service.BriefServiceInterface) *BriefHandler {
	return &BriefHandler{service: svc}
}

// HandleCreateBrief handles POST /api/v1/briefs
func (h *BriefHandler) HandleCreateBrief(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userIDStr := middleware.GetUserID(ctx)
	if userIDStr == "" {
		errors.WriteError(ctx, w, errors.Unauthorized("user not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid user ID"))
		return
	}

	// Parse request body
	var req model.CreateBriefRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid request body"))
		return
	}

	// Create brief
	brief, err := h.service.CreateBrief(ctx, userID, req)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(brief)
}

// HandleGetBrief handles GET /api/v1/briefs/:id
func (h *BriefHandler) HandleGetBrief(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract brief ID from path
	idStr := extractPathVar(r, "id")
	if idStr == "" {
		errors.WriteError(ctx, w, errors.BadRequest("brief ID required"))
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid brief ID"))
		return
	}

	// Get brief
	brief, err := h.service.GetBrief(ctx, id)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brief)
}

// HandleUpdateBrief handles PATCH /api/v1/briefs/:id
func (h *BriefHandler) HandleUpdateBrief(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userIDStr := middleware.GetUserID(ctx)
	if userIDStr == "" {
		errors.WriteError(ctx, w, errors.Unauthorized("user not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid user ID"))
		return
	}

	// Extract brief ID from path
	idStr := extractPathVar(r, "id")
	if idStr == "" {
		errors.WriteError(ctx, w, errors.BadRequest("brief ID required"))
		return
	}

	briefID, err := uuid.Parse(idStr)
	if err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid brief ID"))
		return
	}

	// Parse request body
	var req model.UpdateBriefRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid request body"))
		return
	}

	// Update brief
	brief, err := h.service.UpdateBrief(ctx, userID, briefID, req)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brief)
}

// HandleListBriefs handles GET /api/v1/briefs
func (h *BriefHandler) HandleListBriefs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get query parameters
	userIDStr := r.URL.Query().Get("user_id")
	status := r.URL.Query().Get("status")
	tagsStr := r.URL.Query().Get("tags")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	var userID *uuid.UUID
	if userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			errors.WriteError(ctx, w, errors.BadRequest("invalid user_id"))
			return
		}
		userID = &id
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	var tags []string
	if tagsStr != "" {
		tags = splitCSV(tagsStr)
	}

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// List briefs
	resp, err := h.service.ListBriefs(ctx, userID, statusPtr, tags, page, limit)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandlePublishBrief handles POST /api/v1/briefs/:id/publish
func (h *BriefHandler) HandlePublishBrief(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userIDStr := middleware.GetUserID(ctx)
	if userIDStr == "" {
		errors.WriteError(ctx, w, errors.Unauthorized("user not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid user ID"))
		return
	}

	// Extract brief ID from path
	idStr := extractPathVar(r, "id")
	if idStr == "" {
		errors.WriteError(ctx, w, errors.BadRequest("brief ID required"))
		return
	}

	briefID, err := uuid.Parse(idStr)
	if err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid brief ID"))
		return
	}

	// Parse request body
	var req model.PublishBriefRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid request body"))
		return
	}

	// Publish brief
	brief, err := h.service.PublishBrief(ctx, userID, briefID, req)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brief)
}

// HandleCloseBrief handles POST /api/v1/briefs/:id/close
func (h *BriefHandler) HandleCloseBrief(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context
	userIDStr := middleware.GetUserID(ctx)
	if userIDStr == "" {
		errors.WriteError(ctx, w, errors.Unauthorized("user not authenticated"))
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid user ID"))
		return
	}

	// Extract brief ID from path
	idStr := extractPathVar(r, "id")
	if idStr == "" {
		errors.WriteError(ctx, w, errors.BadRequest("brief ID required"))
		return
	}

	briefID, err := uuid.Parse(idStr)
	if err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid brief ID"))
		return
	}

	// Close brief
	brief, err := h.service.CloseBrief(ctx, userID, briefID)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brief)
}

// HandleInterview handles POST /api/v1/briefs/:id/interview
func (h *BriefHandler) HandleInterview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract brief ID from path
	idStr := extractPathVar(r, "id")
	if idStr == "" {
		errors.WriteError(ctx, w, errors.BadRequest("brief ID required"))
		return
	}

	briefID, err := uuid.Parse(idStr)
	if err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid brief ID"))
		return
	}

	// Parse request body
	var req model.InterviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid request body"))
		return
	}

	// Process interview
	resp, err := h.service.Interview(ctx, briefID, req)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleMatchingBriefs handles GET /api/v1/briefs/matching
func (h *BriefHandler) HandleMatchingBriefs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get query parameters
	tagsStr := r.URL.Query().Get("editor_tags")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	var tags []string
	if tagsStr != "" {
		tags = splitCSV(tagsStr)
	}

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Get matching briefs
	resp, err := h.service.GetMatchingBriefs(ctx, tags, page, limit)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleViewBrief handles POST /api/v1/briefs/:id/view
func (h *BriefHandler) HandleViewBrief(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract brief ID from path
	idStr := extractPathVar(r, "id")
	if idStr == "" {
		errors.WriteError(ctx, w, errors.BadRequest("brief ID required"))
		return
	}

	briefID, err := uuid.Parse(idStr)
	if err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid brief ID"))
		return
	}

	// Parse request body
	var req model.ViewBriefRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid request body"))
		return
	}

	// Mark brief as viewed
	if err := h.service.MarkBriefViewed(ctx, briefID, req); err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func extractPathVar(r *http.Request, key string) string {
	// Simple path extraction - in production use gorilla/mux or chi
	// For now, extract from URL path directly
	path := r.URL.Path
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if i+1 < len(path) {
				return path[i+1:]
			}
			break
		}
	}
	return ""
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	var current string
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			if current != "" {
				result = append(result, current)
			}
			current = ""
		} else {
			current += string(s[i])
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}