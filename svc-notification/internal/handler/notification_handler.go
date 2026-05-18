package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/svc-notification/internal/model"
	"github.com/videoforge/backend/svc-notification/internal/service"
)

// NotificationHandler handles notification API endpoints
type NotificationHandler struct {
	service *service.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		service: svc,
	}
}

// ListNotifications handles GET /api/v1/notifications
func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		errors.WriteError(r.Context(), w, errors.New(http.StatusUnauthorized, "unauthorized"))
		return
	}
	userIDStr := userID.(string)

	// Parse pagination and filters
	filter := model.ListNotificationsFilter{
		PaginationParams: model.PaginationParams{
			Page:    1,
			PerPage: 20,
		},
	}

	// Parse query params
	if page := r.URL.Query().Get("page"); page != "" {
		p, err := strconv.Atoi(page)
		if err != nil || p < 1 {
			p = 1
		}
		filter.Page = p
	}
	if perPage := r.URL.Query().Get("per_page"); perPage != "" {
		p, err := strconv.Atoi(perPage)
		if err != nil || p < 1 || p > 100 {
			p = 20
		}
		filter.PerPage = p
	}
	if readStr := r.URL.Query().Get("read"); readStr != "" {
		read := readStr == "true"
		filter.Read = &read
	}

	notifications, total, err := h.service.ListNotifications(r.Context(), userIDStr, filter)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Internal(err.Error()))
		return
	}

	response := map[string]interface{}{
		"data":       notifications,
		"total":      total,
		"page":       filter.Page,
		"per_page":   filter.PerPage,
		"total_page": (total + filter.PerPage - 1) / filter.PerPage,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// MarkAsRead handles POST /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		errors.WriteError(r.Context(), w, errors.New(http.StatusUnauthorized, "unauthorized"))
		return
	}
	userIDStr := userID.(string)

	notificationID := chi.URLParam(r, "id")
	if notificationID == "" {
		errors.WriteError(r.Context(), w, errors.BadRequest("Notification ID required"))
		return
	}

	err := h.service.MarkAsRead(r.Context(), userIDStr, notificationID)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Internal(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "marked_as_read"})
}

// MarkAllAsRead handles POST /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		errors.WriteError(r.Context(), w, errors.New(http.StatusUnauthorized, "unauthorized"))
		return
	}
	userIDStr := userID.(string)

	err := h.service.MarkAllAsRead(r.Context(), userIDStr)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Internal(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "all_marked_as_read"})
}

// GetPreferences handles GET /api/v1/notifications/preferences
func (h *NotificationHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		errors.WriteError(r.Context(), w, errors.New(http.StatusUnauthorized, "unauthorized"))
		return
	}
	userIDStr := userID.(string)

	pref, err := h.service.GetPreferences(r.Context(), userIDStr)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Internal(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pref)
}

// UpdatePreferences handles PUT /api/v1/notifications/preferences
func (h *NotificationHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		errors.WriteError(r.Context(), w, errors.New(http.StatusUnauthorized, "unauthorized"))
		return
	}
	userIDStr := userID.(string)

	var input model.UpdatePreferenceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("Invalid request body"))
		return
	}

	// Validate channel preference
	if input.ChannelPreference != "" &&
		input.ChannelPreference != model.ChannelWS &&
		input.ChannelPreference != model.ChannelEmail &&
		input.ChannelPreference != model.ChannelBoth {
		errors.WriteError(r.Context(), w, errors.BadRequest("Channel must be ws, email, or both"))
		return
	}

	pref, err := h.service.UpdatePreferences(r.Context(), userIDStr, input)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Internal(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(pref)
}

// DeleteNotification handles DELETE /api/v1/notifications/:id
func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		errors.WriteError(r.Context(), w, errors.New(http.StatusUnauthorized, "unauthorized"))
		return
	}
	userIDStr := userID.(string)

	notificationID := chi.URLParam(r, "id")
	if notificationID == "" {
		errors.WriteError(r.Context(), w, errors.BadRequest("Notification ID required"))
		return
	}

	err := h.service.DeleteNotification(r.Context(), userIDStr, notificationID)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.Internal(fmt.Sprintf("delete_failed: %s", err.Error())))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}