package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/pkg/middleware"

	"github.com/videoforge/backend/svc-campaign/internal/model"
	"github.com/videoforge/backend/svc-campaign/internal/service"
)

// CampaignHandler handles campaign HTTP requests
type CampaignHandler struct {
	service *service.CampaignService
}

// NewCampaignHandler creates a new CampaignHandler
func NewCampaignHandler(svc *service.CampaignService) *CampaignHandler {
	return &CampaignHandler{service: svc}
}

// HandleCreateCampaign handles POST /api/v1/campaigns
func (h *CampaignHandler) HandleCreateCampaign(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from context (use shared middleware function)
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

	role := middleware.GetUserRole(ctx)

	// Parse request body
	var req model.CreateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid request body"))
		return
	}

	// Create campaign
	resp, err := h.service.CreateCampaign(ctx, userID, role, req)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// HandleGetCampaign handles GET /api/v1/campaigns/:id
func (h *CampaignHandler) HandleGetCampaign(w http.ResponseWriter, r *http.Request) {
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

	role := middleware.GetUserRole(ctx)

	// Extract campaign ID from path
	campaignID := extractPathVar(r, "id")
	if campaignID == "" {
		errors.WriteError(ctx, w, errors.BadRequest("campaign ID required"))
		return
	}

	// Get campaign
	resp, err := h.service.GetCampaign(ctx, userID, role, campaignID)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleUpdateCampaign handles PATCH /api/v1/campaigns/:id
func (h *CampaignHandler) HandleUpdateCampaign(w http.ResponseWriter, r *http.Request) {
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

	role := middleware.GetUserRole(ctx)

	// Extract campaign ID from path
	campaignID := extractPathVar(r, "id")
	if campaignID == "" {
		errors.WriteError(ctx, w, errors.BadRequest("campaign ID required"))
		return
	}

	// Parse request body
	var req model.UpdateCampaignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid request body"))
		return
	}

	// Update campaign
	resp, err := h.service.UpdateCampaign(ctx, userID, role, campaignID, req)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleListCampaigns handles GET /api/v1/campaigns
func (h *CampaignHandler) HandleListCampaigns(w http.ResponseWriter, r *http.Request) {
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

	role := middleware.GetUserRole(ctx)

	// Get query parameters
	status := r.URL.Query().Get("status")
	clientID := r.URL.Query().Get("client_id")
	adSpecialistID := r.URL.Query().Get("ad_specialist_id")
	briefID := r.URL.Query().Get("brief_id")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	var clientIDPtr *string
	if clientID != "" {
		clientIDPtr = &clientID
	}

	var adSpecIDPtr *string
	if adSpecialistID != "" {
		adSpecIDPtr = &adSpecialistID
	}

	var briefIDPtr *string
	if briefID != "" {
		briefIDPtr = &briefID
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

	// List campaigns
	resp, err := h.service.ListCampaigns(ctx, userID, role, statusPtr, clientIDPtr, adSpecIDPtr, briefIDPtr, page, limit)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleStartCampaign handles POST /api/v1/campaigns/:id/start
func (h *CampaignHandler) HandleStartCampaign(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	role := middleware.GetUserRole(ctx)

	campaignID := extractPathVar(r, "id")
	if campaignID == "" {
		errors.WriteError(ctx, w, errors.BadRequest("campaign ID required"))
		return
	}

	resp, err := h.service.StartCampaign(ctx, userID, role, campaignID)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandlePauseCampaign handles POST /api/v1/campaigns/:id/pause
func (h *CampaignHandler) HandlePauseCampaign(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	role := middleware.GetUserRole(ctx)

	campaignID := extractPathVar(r, "id")
	if campaignID == "" {
		errors.WriteError(ctx, w, errors.BadRequest("campaign ID required"))
		return
	}

	resp, err := h.service.PauseCampaign(ctx, userID, role, campaignID)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleEndCampaign handles POST /api/v1/campaigns/:id/end
func (h *CampaignHandler) HandleEndCampaign(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	role := middleware.GetUserRole(ctx)

	campaignID := extractPathVar(r, "id")
	if campaignID == "" {
		errors.WriteError(ctx, w, errors.BadRequest("campaign ID required"))
		return
	}

	resp, err := h.service.EndCampaign(ctx, userID, role, campaignID)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleAddVideo handles POST /api/v1/campaigns/:id/videos
func (h *CampaignHandler) HandleAddVideo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	role := middleware.GetUserRole(ctx)

	campaignID := extractPathVar(r, "id")
	if campaignID == "" {
		errors.WriteError(ctx, w, errors.BadRequest("campaign ID required"))
		return
	}

	var req model.AddVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid request body"))
		return
	}

	resp, err := h.service.AddVideo(ctx, userID, role, campaignID, req)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// HandleRemoveVideo handles DELETE /api/v1/campaigns/:id/videos/:video_id
func (h *CampaignHandler) HandleRemoveVideo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	role := middleware.GetUserRole(ctx)

	campaignID := extractPathVar(r, "id")
	if campaignID == "" {
		errors.WriteError(ctx, w, errors.BadRequest("campaign ID required"))
		return
	}

	videoID := extractPathVar(r, "video_id")
	if videoID == "" {
		errors.WriteError(ctx, w, errors.BadRequest("video ID required"))
		return
	}

	if err := h.service.RemoveVideo(ctx, userID, role, campaignID, videoID); err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetBudget handles GET /api/v1/campaigns/:id/budget
func (h *CampaignHandler) HandleGetBudget(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	role := middleware.GetUserRole(ctx)

	campaignID := extractPathVar(r, "id")
	if campaignID == "" {
		errors.WriteError(ctx, w, errors.BadRequest("campaign ID required"))
		return
	}

	resp, err := h.service.GetBudget(ctx, userID, role, campaignID)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleUpdateBudget handles POST /api/v1/campaigns/:id/budget
func (h *CampaignHandler) HandleUpdateBudget(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	role := middleware.GetUserRole(ctx)

	campaignID := extractPathVar(r, "id")
	if campaignID == "" {
		errors.WriteError(ctx, w, errors.BadRequest("campaign ID required"))
		return
	}

	var req model.UpdateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid request body"))
		return
	}

	resp, err := h.service.UpdateBudget(ctx, userID, role, campaignID, req)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleCreateAdAccount handles POST /api/v1/ad-accounts
func (h *CampaignHandler) HandleCreateAdAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	role := middleware.GetUserRole(ctx)

	var req model.AdAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(ctx, w, errors.BadRequest("invalid request body"))
		return
	}

	resp, err := h.service.CreateAdAccount(ctx, userID, role, req)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// HandleListAdAccounts handles GET /api/v1/ad-accounts
func (h *CampaignHandler) HandleListAdAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

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

	resp, err := h.service.ListAdAccounts(ctx, userID, page, limit)
	if err != nil {
		errors.WriteError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Service interface for CampaignHandler
type CampaignServiceInterface interface {
	CreateCampaign(ctx context.Context, userID uuid.UUID, role string, req model.CreateCampaignRequest) (model.CampaignResponse, error)
	GetCampaign(ctx context.Context, userID uuid.UUID, role string, campaignID string) (model.CampaignResponse, error)
	UpdateCampaign(ctx context.Context, userID uuid.UUID, role string, campaignID string, req model.UpdateCampaignRequest) (model.CampaignResponse, error)
	ListCampaigns(ctx context.Context, userID uuid.UUID, role string, status *string, clientID *string, adSpecialistID *string, briefID *string, page, limit int) (*model.ListCampaignsResponse, error)
	StartCampaign(ctx context.Context, userID uuid.UUID, role string, campaignID string) (model.CampaignResponse, error)
	PauseCampaign(ctx context.Context, userID uuid.UUID, role string, campaignID string) (model.CampaignResponse, error)
	EndCampaign(ctx context.Context, userID uuid.UUID, role string, campaignID string) (model.CampaignResponse, error)
	AddVideo(ctx context.Context, userID uuid.UUID, role string, campaignID string, req model.AddVideoRequest) (model.CampaignVideoResponse, error)
	RemoveVideo(ctx context.Context, userID uuid.UUID, role string, campaignID, videoID string) error
	GetBudget(ctx context.Context, userID uuid.UUID, role string, campaignID string) (model.CampaignBudgetResponse, error)
	UpdateBudget(ctx context.Context, userID uuid.UUID, role string, campaignID string, req model.UpdateBudgetRequest) (model.CampaignBudgetResponse, error)
	CreateAdAccount(ctx context.Context, userID uuid.UUID, role string, req model.AdAccountRequest) (model.AdAccountResponse, error)
	ListAdAccounts(ctx context.Context, userID uuid.UUID, page, limit int) (*model.ListAdAccountsResponse, error)
}

// Helper functions

func extractPathVar(r *http.Request, key string) string {
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