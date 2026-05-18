package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/svc-video/internal/model"
	"github.com/videoforge/backend/svc-video/internal/service"
)

// VideoHandler handles video-related HTTP requests
type VideoHandler struct {
	service service.VideoService
	log     *logger.Logger
}

// NewVideoHandler creates a new video handler
func NewVideoHandler(svc service.VideoService, log *logger.Logger) *VideoHandler {
	return &VideoHandler{
		service: svc,
		log:     log,
	}
}

// Routes registers the video routes
func (h *VideoHandler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/videos", h.CreateVideo)
	mux.HandleFunc("GET /api/v1/videos", h.ListVideos)
	mux.HandleFunc("GET /api/v1/videos/", h.GetVideo)
	mux.HandleFunc("POST /api/v1/videos/", h.HandleVideoAction)
}

// HandleVideoAction routes individual video actions
func (h *VideoHandler) HandleVideoAction(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// Extract video ID from path: /api/v1/videos/{id}/action
	// We need to parse: /api/v1/videos/<id>/<action>
	if len(path) < len("/api/v1/videos/") {
		errors.WriteError(r.Context(), w, errors.NotFound("video not found"))
		return
	}

	// Parse video ID and action
	path = path[len("/api/v1/videos/"):]
	var videoID, action string
	if idx := findLastIndexByte(path, '/'); idx >= 0 {
		videoID = path[:idx]
		action = path[idx+1:]
	} else {
		// Direct ID access: GET /api/v1/videos/:id
		videoID = path
		action = ""
	}

	if action == "" {
		// GET /api/v1/videos/:id
		switch r.Method {
		case "GET":
			h.GetVideoByID(w, r, videoID)
		default:
			errors.WriteError(r.Context(), w, errors.NotFound("endpoint not found"))
		}
		return
	}

	switch action {
	case "upload-url":
		if r.Method != "POST" {
			errors.WriteError(r.Context(), w, errors.BadRequest("method not allowed"))
			return
		}
		h.GetUploadURL(w, r, videoID)
	case "download-url":
		switch r.Method {
		case "GET":
			h.GetDownloadURL(w, r, videoID)
		default:
			errors.WriteError(r.Context(), w, errors.BadRequest("method not allowed"))
		}
	case "thumbnail-url":
		switch r.Method {
		case "GET":
			h.GetThumbnailURL(w, r, videoID)
		default:
			errors.WriteError(r.Context(), w, errors.BadRequest("method not allowed"))
		}
	case "confirm-upload":
		if r.Method != "POST" {
			errors.WriteError(r.Context(), w, errors.BadRequest("method not allowed"))
			return
		}
		h.ConfirmUpload(w, r, videoID)
	case "submit":
		if r.Method != "POST" {
			errors.WriteError(r.Context(), w, errors.BadRequest("method not allowed"))
			return
		}
		h.SubmitVideo(w, r, videoID)
	case "approve":
		if r.Method != "POST" {
			errors.WriteError(r.Context(), w, errors.BadRequest("method not allowed"))
			return
		}
		h.ApproveVideo(w, r, videoID)
	case "reject":
		if r.Method != "POST" {
			errors.WriteError(r.Context(), w, errors.BadRequest("method not allowed"))
			return
		}
		h.RejectVideo(w, r, videoID)
	case "revise":
		if r.Method != "POST" {
			errors.WriteError(r.Context(), w, errors.BadRequest("method not allowed"))
			return
		}
		h.RequestRevision(w, r, videoID)
	case "revisions":
		switch r.Method {
		case "GET":
			h.ListRevisions(w, r, videoID)
		case "POST":
			h.CreateRevision(w, r, videoID)
		default:
			errors.WriteError(r.Context(), w, errors.BadRequest("method not allowed"))
		}
	case "feedback":
		if r.Method != "GET" {
			errors.WriteError(r.Context(), w, errors.BadRequest("method not allowed"))
			return
		}
		h.GetFeedback(w, r, videoID)
	case "delete":
		if r.Method != "DELETE" && r.Method != "POST" {
			errors.WriteError(r.Context(), w, errors.BadRequest("method not allowed"))
			return
		}
		h.DeleteVideo(w, r, videoID)
	default:
		errors.WriteError(r.Context(), w, errors.NotFound("endpoint not found"))
	}
}

// findLastIndexByte finds the last occurrence of a byte
func findLastIndexByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// CreateVideo handles POST /api/v1/videos
func (h *VideoHandler) CreateVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		errors.WriteError(r.Context(), w, errors.BadRequest("method not allowed"))
		return
	}

	// Get user from context
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	if userID == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("user not authenticated"))
		return
	}

	// Only editors can create videos
	if userRole != "editor" {
		errors.WriteError(r.Context(), w, errors.Forbidden("only editors can create videos"))
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}
	defer r.Body.Close()

	var req model.CreateVideoRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid JSON"))
		return
	}

	// Validate required fields
	if req.BriefID == "" || req.Title == "" {
		errors.WriteError(r.Context(), w, errors.BadRequest("brief_id and title are required"))
		return
	}

	// Create video
	video, err := h.service.CreateVideo(r.Context(), userID, req.BriefID, req.Title, req.Description)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.CreateVideoResponse{Video: video})
}

// GetVideo handles GET /api/v1/videos/:id
func (h *VideoHandler) GetVideo(w http.ResponseWriter, r *http.Request) {
	// Extract video ID from query or path
	videoID := r.URL.Query().Get("id")
	if videoID == "" {
		// Try to get from path
		path := r.URL.Path
		if len(path) > len("/api/v1/videos") {
			videoID = path[len("/api/v1/videos/"):]
		}
	}
	if videoID == "" {
		errors.WriteError(r.Context(), w, errors.BadRequest("video ID is required"))
		return
	}

	h.GetVideoByID(w, r, videoID)
}

// GetVideoByID handles GET /api/v1/videos/:id
func (h *VideoHandler) GetVideoByID(w http.ResponseWriter, r *http.Request, videoID string) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	if userID == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("user not authenticated"))
		return
	}

	// Brief client ID from query (used for client access)
	briefClientID := r.URL.Query().Get("brief_client_id")

	video, err := h.service.GetVideo(r.Context(), userID, userRole, videoID, briefClientID)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.GetVideoResponse{Video: video})
}

// ListVideos handles GET /api/v1/videos
func (h *VideoHandler) ListVideos(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	if userID == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("user not authenticated"))
		return
	}

	// Get query parameters
	briefID := r.URL.Query().Get("brief_id")
	editorID := r.URL.Query().Get("editor_id")
	status := r.URL.Query().Get("status")
	briefClientID := r.URL.Query().Get("brief_client_id")

	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize == 0 {
		pageSize = 20
	}

	videos, total, err := h.service.ListVideos(r.Context(), userID, userRole, briefClientID, briefID, editorID, status, page, pageSize)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.ListVideosResponse{
		Videos:    videos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetUploadURL handles POST /api/v1/videos/:id/upload-url
func (h *VideoHandler) GetUploadURL(w http.ResponseWriter, r *http.Request, videoID string) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("user not authenticated"))
		return
	}

	resp, err := h.service.GetUploadURL(r.Context(), videoID)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// SubmitVideo handles POST /api/v1/videos/:id/submit
func (h *VideoHandler) SubmitVideo(w http.ResponseWriter, r *http.Request, videoID string) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	if userID == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("user not authenticated"))
		return
	}

	// Only editors can submit
	if userRole != "editor" {
		errors.WriteError(r.Context(), w, errors.Forbidden("only editors can submit videos"))
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}
	defer r.Body.Close()

	var req model.SubmitVideoRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid JSON"))
		return
	}

	if req.StorjKey == "" {
		errors.WriteError(r.Context(), w, errors.BadRequest("storj_key is required"))
		return
	}

	video, err := h.service.SubmitVideo(r.Context(), userID, videoID, req.StorjKey, req.Duration, req.Resolution, req.Thumbnail)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.GetVideoResponse{Video: video})
}

// ApproveVideo handles POST /api/v1/videos/:id/approve
func (h *VideoHandler) ApproveVideo(w http.ResponseWriter, r *http.Request, videoID string) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("user not authenticated"))
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}
	defer r.Body.Close()

	var req model.ApproveVideoRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid JSON"))
		return
	}

	briefClientID := r.URL.Query().Get("brief_client_id")

	video, err := h.service.ApproveVideo(r.Context(), userID, briefClientID, videoID, req.Notes)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.GetVideoResponse{Video: video})
}

// RejectVideo handles POST /api/v1/videos/:id/reject
func (h *VideoHandler) RejectVideo(w http.ResponseWriter, r *http.Request, videoID string) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("user not authenticated"))
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}
	defer r.Body.Close()

	var req model.RejectVideoRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid JSON"))
		return
	}

	if req.Feedback == "" {
		errors.WriteError(r.Context(), w, errors.BadRequest("feedback is required"))
		return
	}

	briefClientID := r.URL.Query().Get("brief_client_id")

	video, err := h.service.RejectVideo(r.Context(), userID, briefClientID, videoID, req.Feedback)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.GetVideoResponse{Video: video})
}

// RequestRevision handles POST /api/v1/videos/:id/revise
func (h *VideoHandler) RequestRevision(w http.ResponseWriter, r *http.Request, videoID string) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("user not authenticated"))
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}
	defer r.Body.Close()

	var req model.ReviseVideoRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid JSON"))
		return
	}

	if req.Feedback == "" {
		errors.WriteError(r.Context(), w, errors.BadRequest("feedback is required"))
		return
	}

	briefClientID := r.URL.Query().Get("brief_client_id")

	video, err := h.service.RequestRevision(r.Context(), userID, briefClientID, videoID, req.Feedback)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.GetVideoResponse{Video: video})
}

// CreateRevision handles POST /api/v1/videos/:id/revisions
func (h *VideoHandler) CreateRevision(w http.ResponseWriter, r *http.Request, videoID string) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	if userID == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("user not authenticated"))
		return
	}

	// Only editors can create revisions
	if userRole != "editor" {
		errors.WriteError(r.Context(), w, errors.Forbidden("only editors can create revisions"))
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}
	defer r.Body.Close()

	var req model.CreateRevisionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid JSON"))
		return
	}

	if req.StorjKey == "" {
		errors.WriteError(r.Context(), w, errors.BadRequest("storj_key is required"))
		return
	}

	revision, err := h.service.CreateRevision(r.Context(), userID, videoID, req.StorjKey, req.Changelog)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(revision)
}

// ListRevisions handles GET /api/v1/videos/:id/revisions
func (h *VideoHandler) ListRevisions(w http.ResponseWriter, r *http.Request, videoID string) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	if userID == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("user not authenticated"))
		return
	}

	briefClientID := r.URL.Query().Get("brief_client_id")

	revisions, err := h.service.GetRevisions(r.Context(), userID, userRole, videoID, briefClientID)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.ListRevisionsResponse{Revisions: revisions})
}

// GetFeedback handles GET /api/v1/videos/:id/feedback
func (h *VideoHandler) GetFeedback(w http.ResponseWriter, r *http.Request, videoID string) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("user not authenticated"))
		return
	}

	feedback, err := h.service.GetFeedback(r.Context(), videoID)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.ListFeedbackResponse{Feedback: feedback})
}

// GetDownloadURL handles GET /api/v1/videos/:id/download-url
func (h *VideoHandler) GetDownloadURL(w http.ResponseWriter, r *http.Request, videoID string) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("user not authenticated"))
		return
	}

	resp, err := h.service.GetDownloadURL(r.Context(), videoID)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetThumbnailURL handles GET /api/v1/videos/:id/thumbnail-url
func (h *VideoHandler) GetThumbnailURL(w http.ResponseWriter, r *http.Request, videoID string) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("user not authenticated"))
		return
	}

	resp, err := h.service.GetThumbnailURL(r.Context(), videoID)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ConfirmUpload handles POST /api/v1/videos/:id/confirm-upload
func (h *VideoHandler) ConfirmUpload(w http.ResponseWriter, r *http.Request, videoID string) {
	userID := middleware.GetUserID(r.Context())
	userRole := middleware.GetUserRole(r.Context())
	if userID == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("user not authenticated"))
		return
	}

	// Only editors can confirm uploads
	if userRole != "editor" {
		errors.WriteError(r.Context(), w, errors.Forbidden("only editors can confirm uploads"))
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid request body"))
		return
	}
	defer r.Body.Close()

	var req model.ConfirmUploadRequest
	if err := json.Unmarshal(body, &req); err != nil {
		errors.WriteError(r.Context(), w, errors.BadRequest("invalid JSON"))
		return
	}

	// Validate required fields
	if req.StorjKey == "" {
		errors.WriteError(r.Context(), w, errors.BadRequest("storj_key is required"))
		return
	}
	if req.FileSize <= 0 {
		errors.WriteError(r.Context(), w, errors.BadRequest("file_size must be greater than 0"))
		return
	}
	if req.Duration <= 0 {
		errors.WriteError(r.Context(), w, errors.BadRequest("duration must be greater than 0"))
		return
	}

	video, err := h.service.ConfirmUpload(r.Context(), userID, videoID, &req)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.ConfirmUploadResponse{Video: video})
}

// DeleteVideo handles DELETE /api/v1/videos/:id/delete
func (h *VideoHandler) DeleteVideo(w http.ResponseWriter, r *http.Request, videoID string) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		errors.WriteError(r.Context(), w, errors.Unauthorized("user not authenticated"))
		return
	}

	err := h.service.DeleteVideo(r.Context(), userID, videoID)
	if err != nil {
		errors.WriteError(r.Context(), w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}