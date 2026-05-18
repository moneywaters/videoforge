package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/natsclient"
	"github.com/videoforge/backend/svc-video/internal/model"
	"github.com/videoforge/backend/svc-video/internal/repository"
)

// VideoService handles video business logic
type VideoService interface {
	CreateVideo(ctx context.Context, editorID, briefID, title, description string) (*model.Video, error)
	GetVideo(ctx context.Context, userID, userRole, videoID string, briefClientID string) (*model.Video, error)
	ListVideos(ctx context.Context, userID, userRole, briefClientID string, briefID, editorID, status string, page, pageSize int) ([]*model.Video, int, error)
	GetUploadURL(ctx context.Context, videoID string) (*model.UploadURLResponse, error)
	SubmitVideo(ctx context.Context, userID, videoID string, storjKey string, duration int, resolution, thumbnail string) (*model.Video, error)
	ApproveVideo(ctx context.Context, userID, briefClientID, videoID, notes string) (*model.Video, error)
	RejectVideo(ctx context.Context, userID, briefClientID, videoID, feedback string) (*model.Video, error)
	RequestRevision(ctx context.Context, userID, briefClientID, videoID, feedback string) (*model.Video, error)
	CreateRevision(ctx context.Context, userID, videoID, storjKey, changelog string) (*model.VideoRevision, error)
	GetRevisions(ctx context.Context, userID, userRole, videoID string, briefClientID string) ([]*model.VideoRevision, error)
	GetFeedback(ctx context.Context, videoID string) ([]*model.VideoFeedback, error)
}

type videoService struct {
	repo      repository.VideoRepository
	natsClient natsclient.NATSClient
	log       *logger.Logger
}

// NewVideoService creates a new video service
func NewVideoService(repo repository.VideoRepository, nc natsclient.NATSClient, log *logger.Logger) VideoService {
	return &videoService{
		repo:      repo,
		natsClient: nc,
		log:       log,
	}
}

// CreateVideo creates a new video for a brief
func (s *videoService) CreateVideo(ctx context.Context, editorID, briefID, title, description string) (*model.Video, error) {
	// Validate brief exists and is published
	brief, err := s.repo.GetBrief(ctx, briefID)
	if err != nil {
		return nil, errors.BadRequest("invalid brief ID")
	}
	if brief == nil {
		return nil, errors.NotFound("brief not found")
	}
	if !brief.IsPublished {
		return nil, errors.BadRequest("brief is not published")
	}

	// Create video
	now := time.Now()
	video := &model.Video{
		ID:          uuid.New().String(),
		BriefID:     briefID,
		EditorID:    editorID,
		Title:       title,
		Description: description,
		Status:      model.StatusDraft,
		Duration:    0,
		Resolution: "1080p",
		CreatedAt:   now,
		UpdatedAt:  now,
	}

	if err := s.repo.Create(ctx, video); err != nil {
		s.log.Error("failed to create video",
			slog.String("error", err.Error())),
		)
		return nil, errors.Internal("failed to create video")
	}

	return video, nil
}

// GetVideo retrieves a video by ID with blind submission enforcement
func (s *videoService) GetVideo(ctx context.Context, userID, userRole, videoID string, briefClientID string) (*model.Video, error) {
	video, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return nil, errors.Internal("failed to get video")
	}
	if video == nil {
		return nil, errors.NotFound("video not found")
	}

	// Get the brief to check ownership
	brief, err := s.repo.GetBrief(ctx, video.BriefID)
	if err != nil || brief == nil {
		return nil, errors.NotFound("brief not found")
	}

	// Enforce blind submissions
	if userRole == "editor" {
		// Editors can only see their own videos for a brief
		if video.EditorID != userID {
			if brief.ClientID != userID {
				return nil, errors.Forbidden("you can only view your own videos")
			}
		}
	} else if userRole == "client" {
		// Clients can only see videos on their briefs
		if brief.ClientID != userID {
			return nil, errors.Forbidden("you can only view videos on your briefs")
		}
	}
	// Admin can view all

	return video, nil
}

// ListVideos lists videos with blind submission filtering
func (s *videoService) ListVideos(ctx context.Context, userID, userRole, briefClientID string, briefID, editorID, status string, page, pageSize int) ([]*model.Video, int, error) {
	// Enforce blind submissions based on role
	if userRole == "editor" && briefID != "" {
		// Editors can only see their own videos for a brief
		return s.repo.ListByBriefForEditor(ctx, briefID, userID, page, pageSize)
	} else if userRole == "client" && briefID != "" {
		// Clients viewing their briefs don't need extra filtering, they see all
		return s.repo.List(ctx, briefID, "", status, page, pageSize)
	}

	return s.repo.List(ctx, briefID, editorID, status, page, pageSize)
}

// GetUploadURL generates a mock presigned URL for upload (STUB)
func (s *videoService) GetUploadURL(ctx context.Context, videoID string) (*model.UploadURLResponse, error) {
	video, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return nil, errors.Internal("failed to get video")
	}
	if video == nil {
		return nil, errors.NotFound("video not found")
	}

	// Stub: generate mock presigned URL (in production, call Storj API)
	mockURL := fmt.Sprintf("https://link.storj.io/broadcast/%s/mock-presigned-url-%s", videoID, time.Now().Unix())
	storjKey := fmt.Sprintf("videos/%s/%s.mp4", video.ID, uuid.New().String())

	return &model.UploadURLResponse{
		UploadURL: mockURL,
		StorjKey: storjKey,
	}, nil
}

// SubmitVideo submits a video for approval
func (s *videoService) SubmitVideo(ctx context.Context, userID, videoID string, storjKey string, duration int, resolution, thumbnail string) (*model.Video, error) {
	video, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return nil, errors.Internal("failed to get video")
	}
	if video == nil {
		return nil, errors.NotFound("video not found")
	}

	// Only the editor who created the video can submit
	if video.EditorID != userID {
		return nil, errors.Forbidden("only the editor who created this video can submit it")
	}

	// Can only submit from draft or revision_requested status
	if video.Status != model.StatusDraft && video.Status != model.StatusRevisionRequested {
		return nil, errors.BadRequest("video cannot be submitted in current status")
	}

	now := time.Now()
	video.StorjKey = storjKey
	video.Status = model.StatusSubmitted
	video.Duration = duration
	video.Resolution = resolution
	video.ThumbnailURL = thumbnail
	video.SubmittedAt = &now
	video.UpdatedAt = now

	if err := s.repo.Update(ctx, video); err != nil {
		return nil, errors.Internal("failed to submit video")
	}

	// Emit event to NATS
	if s.natsClient != nil {
		event := map[string]interface{}{
			"event":     "video.submitted",
			"video_id": video.ID,
			"brief_id": video.BriefID,
			"editor_id": video.EditorID,
			"timestamp": now,
		}
		if data, err := json.Marshal(event); err == nil {
			s.natsClient.Publish("video.submitted", data)
		}
	}

	return video, nil
}

// ApproveVideo approves a video
func (s *videoService) ApproveVideo(ctx context.Context, userID, briefClientID, videoID, notes string) (*model.Video, error) {
	video, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return nil, errors.Internal("failed to get video")
	}
	if video == nil {
		return nil, errors.NotFound("video not found")
	}

	// Get brief to verify client owns it
	brief, err := s.repo.GetBrief(ctx, video.BriefID)
	if err != nil || brief == nil {
		return nil, errors.Forbidden("brief not found")
	}

	// Only the client who owns the brief can approve
	if brief.ClientID != userID {
		return nil, errors.Forbidden("only the client can approve videos for their briefs")
	}

	// Can only approve submitted videos
	if video.Status != model.StatusSubmitted {
		return nil, errors.BadRequest("can only approve submitted videos")
	}

	now := time.Now()
	video.Status = model.StatusApproved
	video.UpdatedAt = now

	if err := s.repo.Update(ctx, video); err != nil {
		return nil, errors.Internal("failed to approve video")
	}

	// Create approval record
	approval := &model.VideoApproval{
		ID:         uuid.New().String(),
		VideoID:    videoID,
		Status:    "approved",
		ApprovedBy: userID,
		ApprovedAt: now,
		Notes:     notes,
	}
	if err := s.repo.CreateApproval(ctx, approval); err != nil {
		s.log.Error("failed to create approval record",
			slog.String("error", err.Error())),
		)
	}

	// Emit event to NATS
	if s.natsClient != nil {
		event := map[string]interface{}{
			"event":     "video.approved",
			"video_id": video.ID,
			"brief_id": video.BriefID,
			"approved_by": userID,
			"timestamp": now,
		}
		if data, err := json.Marshal(event); err == nil {
			s.natsClient.Publish("video.approved", data)
		}
	}

	return video, nil
}

// RejectVideo rejects a video
func (s *videoService) RejectVideo(ctx context.Context, userID, briefClientID, videoID, feedback string) (*model.Video, error) {
	video, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return nil, errors.Internal("failed to get video")
	}
	if video == nil {
		return nil, errors.NotFound("video not found")
	}

	// Get brief to verify client owns it
	brief, err := s.repo.GetBrief(ctx, video.BriefID)
	if err != nil || brief == nil {
		return nil, errors.Forbidden("brief not found")
	}

	// Only the client who owns the brief can reject
	if brief.ClientID != userID {
		return nil, errors.Forbidden("only the client can reject videos for their briefs")
	}

	// Can only reject submitted videos
	if video.Status != model.StatusSubmitted {
		return nil, errors.BadRequest("can only reject submitted videos")
	}

	now := time.Now()
	video.Status = model.StatusRejected
	video.UpdatedAt = now

	if err := s.repo.Update(ctx, video); err != nil {
		return nil, errors.Internal("failed to reject video")
	}

	// Create approval record
	approval := &model.VideoApproval{
		ID:         uuid.New().String(),
		VideoID:    videoID,
		Status:    "rejected",
		ApprovedBy: userID,
		ApprovedAt: now,
		Notes:    feedback,
	}
	if err := s.repo.CreateApproval(ctx, approval); err != nil {
		s.log.Error("failed to create rejection record",
			slog.String("error", err.Error())),
		)
	}

	// Emit event to NATS
	if s.natsClient != nil {
		event := map[string]interface{}{
			"event":      "video.rejected",
			"video_id":  video.ID,
			"brief_id":  video.BriefID,
			"rejected_by": userID,
			"feedback":  feedback,
			"timestamp": now,
		}
		if data, err := json.Marshal(event); err == nil {
			s.natsClient.Publish("video.rejected", data)
		}
	}

	return video, nil
}

// RequestRevision requests revisions for a video
func (s *videoService) RequestRevision(ctx context.Context, userID, briefClientID, videoID, feedback string) (*model.Video, error) {
	video, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return nil, errors.Internal("failed to get video")
	}
	if video == nil {
		return nil, errors.NotFound("video not found")
	}

	// Get brief to verify client owns it
	brief, err := s.repo.GetBrief(ctx, video.BriefID)
	if err != nil || brief == nil {
		return nil, errors.Forbidden("brief not found")
	}

	// Only the client who owns the brief can request revisions
	if brief.ClientID != userID {
		return nil, errors.Forbidden("only the client can request revisions for their briefs")
	}

	// Can only request revisions for submitted videos
	if video.Status != model.StatusSubmitted {
		return nil, errors.BadRequest("can only request revisions for submitted videos")
	}

	now := time.Now()
	video.Status = model.StatusRevisionRequested
	video.UpdatedAt = now

	if err := s.repo.Update(ctx, video); err != nil {
		return nil, errors.Internal("failed to request revision")
	}

	// Get current revision for feedback
	revisionID := video.CurrentRevisionID
	if revisionID == "" {
		// If no revision yet, create a placeholder ID
		revisionID = uuid.New().String()
	}

	// Create feedback record
	fb := &model.VideoFeedback{
		ID:        uuid.New().String(),
		VideoID:   videoID,
		RevisionID: revisionID,
		Feedback:  feedback,
		CreatedBy: userID,
		CreatedAt: now,
	}
	if err := s.repo.CreateFeedback(ctx, fb); err != nil {
		s.log.Error("failed to create feedback record",
			slog.String("error", err.Error())),
		)
	}

	// Emit event to NATS
	if s.natsClient != nil {
		event := map[string]interface{}{
			"event":      "video.revision_requested",
			"video_id":  video.ID,
			"brief_id":  video.BriefID,
			"requested_by": userID,
			"feedback":  feedback,
			"timestamp": now,
		}
		if data, err := json.Marshal(event); err == nil {
			s.natsClient.Publish("video.revision_requested", data)
		}
	}

	return video, nil
}

// CreateRevision creates a new revision for a video
func (s *videoService) CreateRevision(ctx context.Context, userID, videoID, storjKey, changelog string) (*model.VideoRevision, error) {
	video, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return nil, errors.Internal("failed to get video")
	}
	if video == nil {
		return nil, errors.NotFound("video not found")
	}

	// Only the editor who created the video can create revisions
	if video.EditorID != userID {
		return nil, errors.Forbidden("only the editor who created this video can add revisions")
	}

	// Get latest revision number
	latestNum, err := s.repo.GetLatestRevisionNumber(ctx, videoID)
	if err != nil {
		return nil, errors.Internal("failed to get revision number")
	}

	// Create revision
	now := time.Now()
	revision := &model.VideoRevision{
		ID:             uuid.New().String(),
		VideoID:        videoID,
		RevisionNumber: latestNum + 1,
		StorjKey:       storjKey,
		Changelog:      changelog,
		CreatedAt:      now,
	}

	if err := s.repo.CreateRevision(ctx, revision); err != nil {
		return nil, errors.Internal("failed to create revision")
	}

	// Update video's current revision and reset status to draft
	video.CurrentRevisionID = revision.ID
	if video.Status == model.StatusRevisionRequested {
		video.Status = model.StatusDraft
	}
	video.UpdatedAt = now

	if err := s.repo.Update(ctx, video); err != nil {
		return nil, errors.Internal("failed to update video")
	}

	return revision, nil
}

// GetRevisions gets all revisions for a video (only visible to submitting editor and client)
func (s *videoService) GetRevisions(ctx context.Context, userID, userRole, videoID string, briefClientID string) ([]*model.VideoRevision, error) {
	video, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return nil, errors.Internal("failed to get video")
	}
	if video == nil {
		return nil, errors.NotFound("video not found")
	}

	// Check access - only editor who submitted or the client can view revisions
	if userRole == "editor" && video.EditorID != userID {
		if briefClientID != "" {
			// Check if user is the client
			brief, err := s.repo.GetBrief(ctx, video.BriefID)
			if err != nil || brief == nil || brief.ClientID != briefClientID {
				return nil, errors.Forbidden("not authorized to view revisions")
			}
		} else {
			return nil, errors.Forbidden("not authorized to view revisions")
		}
	} else if userRole == "client" {
		// Client can view if they own the brief
		brief, err := s.repo.GetBrief(ctx, video.BriefID)
		if err != nil || brief == nil || brief.ClientID != userID {
			return nil, errors.Forbidden("not authorized to view revisions")
		}
	}
	// Admin can view all

	return s.repo.ListRevisions(ctx, videoID)
}

// GetFeedback gets feedback history for a video
func (s *videoService) GetFeedback(ctx context.Context, videoID string) ([]*model.VideoFeedback, error) {
	video, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return nil, errors.Internal("failed to get video")
	}
	if video == nil {
		return nil, errors.NotFound("video not found")
	}

	return s.repo.ListFeedback(ctx, videoID)
}