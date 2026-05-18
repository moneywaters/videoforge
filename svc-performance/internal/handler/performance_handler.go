package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/videoforge/backend/pkg/errors"
	"svc-performance/internal/model"
	"svc-performance/internal/service"
)

type PerformanceHandler struct {
	service *service.PerformanceService
}

func NewPerformanceHandler(svc *service.PerformanceService) *PerformanceHandler {
	return &PerformanceHandler{service: svc}
}

// GetVideoSales handles GET /api/v1/performance/videos/:id
func (h *PerformanceHandler) GetVideoSales(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	if videoID == "" {
		errors.BadRequest("video id is required").Write(r.Context(), w)
		return
	}

	sales, err := h.service.GetVideoSales(r.Context(), videoID)
	if err != nil {
		if prob, ok := err.(*errors.ProblemDetails); ok {
			errors.Write(r.Context(), w, prob)
			return
		}
		errors.Internal("failed to get video sales").Write(r.Context(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sales)
}

// GetEditorSales handles GET /api/v1/performance/editors/:id
func (h *PerformanceHandler) GetEditorSales(w http.ResponseWriter, r *http.Request) {
	editorID := chi.URLParam(r, "id")
	if editorID == "" {
		errors.BadRequest("editor id is required").Write(r.Context(), w)
		return
	}

	sales, err := h.service.GetEditorSales(r.Context(), editorID)
	if err != nil {
		if prob, ok := err.(*errors.ProblemDetails); ok {
			errors.Write(r.Context(), w, prob)
			return
		}
		errors.Internal("failed to get editor sales").Write(r.Context(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sales)
}

// GetSpecialistSales handles GET /api/v1/performance/specialists/:id
func (h *PerformanceHandler) GetSpecialistSales(w http.ResponseWriter, r *http.Request) {
	specialistID := chi.URLParam(r, "id")
	if specialistID == "" {
		errors.BadRequest("specialist id is required").Write(r.Context(), w)
		return
	}

	sales, err := h.service.GetSpecialistSales(r.Context(), specialistID)
	if err != nil {
		if prob, ok := err.(*errors.ProblemDetails); ok {
			errors.Write(r.Context(), w, prob)
			return
		}
		errors.Internal("failed to get specialist sales").Write(r.Context(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sales)
}

// GetCampaignSales handles GET /api/v1/performance/campaigns/:id
func (h *PerformanceHandler) GetCampaignSales(w http.ResponseWriter, r *http.Request) {
	campaignID := chi.URLParam(r, "id")
	if campaignID == "" {
		errors.BadRequest("campaign id is required").Write(r.Context(), w)
		return
	}

	sales, err := h.service.GetCampaignSales(r.Context(), campaignID)
	if err != nil {
		if prob, ok := err.(*errors.ProblemDetails); ok {
			errors.Write(r.Context(), w, prob)
			return
		}
		errors.Internal("failed to get campaign sales").Write(r.Context(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sales)
}

// GetLeaderboard handles GET /api/v1/performance/briefs/:id/leaderboard
func (h *PerformanceHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	briefID := chi.URLParam(r, "id")
	if briefID == "" {
		errors.BadRequest("brief id is required").Write(r.Context(), w)
		return
	}

	entityType := r.URL.Query().Get("entity_type")
	if entityType == "" {
		errors.BadRequest("entity_type query parameter is required").Write(r.Context(), w)
		return
	}

	if entityType != "editor" && entityType != "video" {
		errors.BadRequest("entity_type must be 'editor' or 'video'").Write(r.Context(), w)
		return
	}

	leaderboard, err := h.service.GetLeaderboard(r.Context(), briefID, entityType)
	if err != nil {
		if prob, ok := err.(*errors.ProblemDetails); ok {
			errors.Write(r.Context(), w, prob)
			return
		}
		errors.Internal("failed to get leaderboard").Write(r.Context(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(leaderboard)
}

// GetRankings handles GET /api/v1/performance/briefs/:id/rankings
func (h *PerformanceHandler) GetRankings(w http.ResponseWriter, r *http.Request) {
	briefID := chi.URLParam(r, "id")
	if briefID == "" {
		errors.BadRequest("brief id is required").Write(r.Context(), w)
		return
	}

	entityType := r.URL.Query().Get("entity_type")
	if entityType == "" {
		errors.BadRequest("entity_type query parameter is required").Write(r.Context(), w)
		return
	}

	if entityType != "editor" && entityType != "video" {
		errors.BadRequest("entity_type must be 'editor' or 'video'").Write(r.Context(), w)
		return
	}

	// Pagination with defaults
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // default
	offset := 0 // default

	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 || l > 100 {
			errors.BadRequest("limit must be between 1 and 100").Write(r.Context(), w)
			return
		}
		limit = l
	}

	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil || o < 0 {
			errors.BadRequest("offset must be non-negative").Write(r.Context(), w)
			return
		}
		offset = o
	}

	rankings, err := h.service.GetRankings(r.Context(), briefID, entityType, limit, offset)
	if err != nil {
		if prob, ok := err.(*errors.ProblemDetails); ok {
			errors.Write(r.Context(), w, prob)
			return
		}
		errors.Internal("failed to get rankings").Write(r.Context(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rankings)
}

// GetAnalytics handles GET /api/v1/performance/analytics
func (h *PerformanceHandler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	query := model.AnalyticsQuery{
		EntityType:  r.URL.Query().Get("entity_type"),
		EntityID:   r.URL.Query().Get("entity_id"),
		StartDate:  r.URL.Query().Get("start_date"),
		EndDate:    r.URL.Query().Get("end_date"),
		Granularity: r.URL.Query().Get("granularity"),
	}

	// Validate required params
	if query.EntityType == "" {
		errors.BadRequest("entity_type query parameter is required").Write(r.Context(), w)
		return
	}
	if query.EntityID == "" {
		errors.BadRequest("entity_id query parameter is required").Write(r.Context(), w)
		return
	}
	if query.StartDate == "" {
		errors.BadRequest("start_date query parameter is required").Write(r.Context(), w)
		return
	}
	if query.EndDate == "" {
		errors.BadRequest("end_date query parameter is required").Write(r.Context(), w)
		return
	}

	// Validate entity_type
	if query.EntityType != "video" && query.EntityType != "campaign" {
		errors.BadRequest("entity_type must be 'video' or 'campaign'").Write(r.Context(), w)
		return
	}

	// Set default granularity
	if query.Granularity == "" {
		query.Granularity = "daily"
	}

	analytics, err := h.service.GetAnalytics(r.Context(), query)
	if err != nil {
		if prob, ok := err.(*errors.ProblemDetails); ok {
			errors.Write(r.Context(), w, prob)
			return
		}
		errors.Internal("failed to get analytics").Write(r.Context(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(analytics)
}

// GetAnomalies handles GET /api/v1/performance/anomalies
func (h *PerformanceHandler) GetAnomalies(w http.ResponseWriter, r *http.Request) {
	anomalies, err := h.service.GetAnomalies(r.Context())
	if err != nil {
		if prob, ok := err.(*errors.ProblemDetails); ok {
			errors.Write(r.Context(), w, prob)
			return
		}
		errors.Internal("failed to get anomalies").Write(r.Context(), w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(anomalies)
}