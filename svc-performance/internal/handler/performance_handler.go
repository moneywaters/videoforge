package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/svc-performance/internal/model"
	"github.com/videoforge/backend/svc-performance/internal/service"
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
		errors.Write(r.Context(), w, errors.BadRequest("video id is required"))
		return
	}

	sales, err := h.service.GetVideoSales(r.Context(), videoID)
	if err != nil {
		if prob, ok := err.(*errors.ProblemDetails); ok {
			errors.Write(r.Context(), w, prob)
			return
		}
		errors.Write(r.Context(), w, errors.Internal("failed to get video sales"))
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
		errors.Write(r.Context(), w, errors.BadRequest("editor id is required"))
		return
	}

	sales, err := h.service.GetEditorSales(r.Context(), editorID)
	if err != nil {
		if prob, ok := err.(*errors.ProblemDetails); ok {
			errors.Write(r.Context(), w, prob)
			return
		}
		errors.Write(r.Context(), w, errors.Internal("failed to get editor sales"))
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
		errors.Write(r.Context(), w, errors.BadRequest("specialist id is required"))
		return
	}

	sales, err := h.service.GetSpecialistSales(r.Context(), specialistID)
	if err != nil {
		if prob, ok := err.(*errors.ProblemDetails); ok {
			errors.Write(r.Context(), w, prob)
			return
		}
		errors.Write(r.Context(), w, errors.Internal("failed to get specialist sales"))
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
		errors.Write(r.Context(), w, errors.BadRequest("campaign id is required"))
		return
	}

	sales, err := h.service.GetCampaignSales(r.Context(), campaignID)
	if err != nil {
		if prob, ok := err.(*errors.ProblemDetails); ok {
			errors.Write(r.Context(), w, prob)
			return
		}
		errors.Write(r.Context(), w, errors.Internal("failed to get campaign sales"))
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
		errors.Write(r.Context(), w, errors.BadRequest("brief id is required"))
		return
	}

	entityType := r.URL.Query().Get("entity_type")
	if entityType == "" {
		errors.Write(r.Context(), w, errors.BadRequest("entity_type query parameter is required"))
		return
	}

	if entityType != "editor" && entityType != "video" {
		errors.Write(r.Context(), w, errors.BadRequest("entity_type must be 'editor' or 'video'"))
		return
	}

	leaderboard, err := h.service.GetLeaderboard(r.Context(), briefID, entityType)
	if err != nil {
		if prob, ok := err.(*errors.ProblemDetails); ok {
			errors.Write(r.Context(), w, prob)
			return
		}
		errors.Write(r.Context(), w, errors.Internal("failed to get leaderboard"))
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
		errors.Write(r.Context(), w, errors.BadRequest("brief id is required"))
		return
	}

	entityType := r.URL.Query().Get("entity_type")
	if entityType == "" {
		errors.Write(r.Context(), w, errors.BadRequest("entity_type query parameter is required"))
		return
	}

	if entityType != "editor" && entityType != "video" {
		errors.Write(r.Context(), w, errors.BadRequest("entity_type must be 'editor' or 'video'"))
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
			errors.Write(r.Context(), w, errors.BadRequest("limit must be between 1 and 100"))
			return
		}
		limit = l
	}

	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil || o < 0 {
			errors.Write(r.Context(), w, errors.BadRequest("offset must be non-negative"))
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
		errors.Write(r.Context(), w, errors.Internal("failed to get rankings"))
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
		errors.Write(r.Context(), w, errors.BadRequest("entity_type query parameter is required"))
		return
	}
	if query.EntityID == "" {
		errors.Write(r.Context(), w, errors.BadRequest("entity_id query parameter is required"))
		return
	}
	if query.StartDate == "" {
		errors.Write(r.Context(), w, errors.BadRequest("start_date query parameter is required"))
		return
	}
	if query.EndDate == "" {
		errors.Write(r.Context(), w, errors.BadRequest("end_date query parameter is required"))
		return
	}

	// Validate entity_type
	if query.EntityType != "video" && query.EntityType != "campaign" {
		errors.Write(r.Context(), w, errors.BadRequest("entity_type must be 'video' or 'campaign'"))
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
		errors.Write(r.Context(), w, errors.Internal("failed to get analytics"))
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
		errors.Write(r.Context(), w, errors.Internal("failed to get anomalies"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(anomalies)
}