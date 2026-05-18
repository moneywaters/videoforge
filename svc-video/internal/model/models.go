package model

import "time"

// VideoStatus represents the status of a video
type VideoStatus string

const (
	StatusDraft            VideoStatus = "draft"
	StatusSubmitted       VideoStatus = "submitted"
	StatusApproved       VideoStatus = "approved"
	StatusRejected       VideoStatus = "rejected"
	StatusRevisionRequested VideoStatus = "revision_requested"
)

// Video represents a video submission for a brief
type Video struct {
	ID                string     `json:"id"`
	BriefID          string     `json:"brief_id"`
	EditorID         string     `json:"editor_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description,omitempty"`
	StorjKey        string     `json:"storj_key,omitempty"`
	Status          VideoStatus `json:"status"`
	CurrentRevisionID string     `json:"current_revision_id,omitempty"`
	Duration        int        `json:"duration"`
	Resolution      string     `json:"resolution"`
	ThumbnailURL    string     `json:"thumbnail_url,omitempty"`
	SubmittedAt     *time.Time `json:"submitted_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// VideoRevision represents a revision of a video
type VideoRevision struct {
	ID             string    `json:"id"`
	VideoID        string   `json:"video_id"`
	RevisionNumber int      `json:"revision_number"`
	StorjKey      string   `json:"storj_key,omitempty"`
	Changelog     string   `json:"changelog,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// VideoApproval represents an approval or rejection decision
type VideoApproval struct {
	ID         string    `json:"id"`
	VideoID    string   `json:"video_id"`
	Status    string   `json:"status"` // "approved" or "rejected"
	ApprovedBy string   `json:"approved_by"`
	ApprovedAt time.Time `json:"approved_at"`
	Notes     string   `json:"notes,omitempty"`
}

// VideoFeedback represents feedback from the client
type VideoFeedback struct {
	ID        string    `json:"id"`
	VideoID   string   `json:"video_id"`
	RevisionID string   `json:"revision_id"`
	Feedback  string   `json:"feedback"`
	CreatedBy string   `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// Request/Response types

// CreateVideoRequest represents the request to create a new video
type CreateVideoRequest struct {
	BriefID      string `json:"brief_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// CreateVideoResponse represents the response after creating a video
type CreateVideoResponse struct {
	Video *Video `json:"video"`
}

// GetVideoResponse represents the response for getting a video
type GetVideoResponse struct {
	Video *Video `json:"video"`
}

// ListVideosRequest represents the request to list videos with filters
type ListVideosRequest struct {
	BriefID   string `json:"brief_id,omitempty"`
	EditorID  string `json:"editor_id,omitempty"`
	Status   string `json:"status,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
}

// ListVideosResponse represents the response for listing videos
type ListVideosResponse struct {
	Videos    []*Video `json:"videos"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// UploadURLResponse represents the response for getting an upload URL
type UploadURLResponse struct {
	UploadURL string `json:"upload_url"`
	StorjKey  string `json:"storj_key"`
}

// SubmitVideoRequest represents the request to submit a video
type SubmitVideoRequest struct {
	StorjKey   string `json:"storj_key"`
	Duration  int    `json:"duration"`
	Resolution string `json:"resolution"`
	Thumbnail string `json:"thumbnail"`
}

// ApproveVideoRequest represents the request to approve a video
type ApproveVideoRequest struct {
	Notes string `json:"notes,omitempty"`
}

// RejectVideoRequest represents the request to reject a video
type RejectVideoRequest struct {
	Feedback string `json:"feedback"`
}

// ReviseVideoRequest represents the request to request revisions
type ReviseVideoRequest struct {
	Feedback string `json:"feedback"`
}

// CreateRevisionRequest represents the request to create a new revision
type CreateRevisionRequest struct {
	StorjKey  string `json:"storj_key"`
	Changelog string `json:"changelog"`
}

// ListRevisionsResponse represents the response for listing revisions
type ListRevisionsResponse struct {
	Revisions []*VideoRevision `json:"revisions"`
}

// ListFeedbackResponse represents the response for listing feedback
type ListFeedbackResponse struct {
	Feedback []*VideoFeedback `json:"feedback"`
}

// Brief represents a brief (for validation - in production this would come from svc-brief)
type Brief struct {
	ID          string `json:"id"`
	ClientID    string `json:"client_id"`
	Status     string `json:"status"`
	IsPublished bool  `json:"is_published"`
}