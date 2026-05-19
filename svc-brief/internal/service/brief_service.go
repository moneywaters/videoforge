package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/videoforge/backend/svc-brief/internal/model"
	"github.com/videoforge/backend/svc-brief/internal/repository"

	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/pkg/storage"
)

var (
	ErrCannotUpdatePublished = errors.BadRequest("cannot update published brief")
	ErrCannotPublishClosed  = errors.BadRequest("cannot publish closed brief")
	ErrBountyNotDeposited     = errors.BadRequest("bounty must be deposited before publishing")
	ErrInterviewNoBrief      = errors.BadRequest("no brief found for interview")
	ErrNoStorageClient       = errors.BadRequest("storage not configured")
	ErrNoRawFootage          = errors.BadRequest("no raw footage uploaded")
)

// BriefService handles brief business logic
type BriefService struct {
	repo    repository.BriefRepoInterface
	storage storage.Storage
}

// BriefServiceInterface defines the brief service interface
type BriefServiceInterface interface {
	CreateBrief(ctx context.Context, userID uuid.UUID, req model.CreateBriefRequest) (model.BriefResponse, error)
	GetBrief(ctx context.Context, id uuid.UUID) (model.BriefResponse, error)
	UpdateBrief(ctx context.Context, userID, briefID uuid.UUID, req model.UpdateBriefRequest) (model.BriefResponse, error)
	DeleteBrief(ctx context.Context, userID, briefID uuid.UUID) error
	ListBriefs(ctx context.Context, userID *uuid.UUID, status *string, tags []string, page, limit int) (*model.ListBriefsResponse, error)
	PublishBrief(ctx context.Context, userID, briefID uuid.UUID, req model.PublishBriefRequest) (model.BriefResponse, error)
	CloseBrief(ctx context.Context, userID, briefID uuid.UUID) (model.BriefResponse, error)
	Interview(ctx context.Context, briefID uuid.UUID, req model.InterviewRequest) (model.InterviewResponse, error)
	GetMatchingBriefs(ctx context.Context, tags []string, page, limit int) (*model.ListBriefsResponse, error)
	MarkBriefViewed(ctx context.Context, briefID uuid.UUID, req model.ViewBriefRequest) error

	// Raw footage
	GetRawFootageUploadURL(ctx context.Context, userID, briefID uuid.UUID, filename string) (*model.UploadRawFootageResponse, error)
	ConfirmRawFootageUpload(ctx context.Context, userID, briefID uuid.UUID, storjKey string) error
	GetRawFootageDownloadURL(ctx context.Context, userID, briefID uuid.UUID) (*model.RawFootageDownloadURLResponse, error)
}

// NewBriefService creates a new BriefService
func NewBriefService(repo repository.BriefRepoInterface, storageClient storage.Storage) *BriefService {
	return &BriefService{repo: repo, storage: storageClient}
}

// CreateBrief creates a new brief
func (s *BriefService) CreateBrief(ctx context.Context, userID uuid.UUID, req model.CreateBriefRequest) (model.BriefResponse, error) {
	// Validate required fields
	if req.Title == "" {
		return model.BriefResponse{}, errors.BadRequest("title is required")
	}

	if req.SubmissionsLimit < 1 {
		req.SubmissionsLimit = 1
	}

	now := time.Now()
	brief := &model.Brief{
		ID:               uuid.Must(uuid.NewV7()),
		ClientID:         userID,
		Title:            req.Title,
		Description:      req.Description,
		Goals:            req.Goals,
		TargetAudience:   req.TargetAudience,
		Tone:             req.Tone,
		StylePreferences: req.StylePreferences,
		CTA:              req.CTA,
		Status:           model.BriefStatusDraft,
		BountyBudget:    req.BountyBudget,
		BountyDeposited: false,
		SubmissionsLimit: req.SubmissionsLimit,
		IsBlind:         req.IsBlind,
		Tags:             req.Tags,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	// Default is_blind to true
	if !req.IsBlind {
		brief.IsBlind = true
	}

	if err := s.repo.CreateBrief(ctx, brief); err != nil {
		return model.BriefResponse{}, fmt.Errorf("failed to create brief: %w", err)
	}

	return brief.ToResponse(), nil
}

// GetBrief gets a brief by ID
func (s *BriefService) GetBrief(ctx context.Context, id uuid.UUID) (model.BriefResponse, error) {
	brief, err := s.repo.GetBriefByID(ctx, id)
	if err != nil {
		if err == repository.ErrBriefNotFound {
			return model.BriefResponse{}, errors.NotFound("brief not found")
		}
		return model.BriefResponse{}, fmt.Errorf("failed to get brief: %w", err)
	}

	return brief.ToResponse(), nil
}

// UpdateBrief updates a brief
func (s *BriefService) UpdateBrief(ctx context.Context, userID, briefID uuid.UUID, req model.UpdateBriefRequest) (model.BriefResponse, error) {
	brief, err := s.repo.GetBriefByID(ctx, briefID)
	if err != nil {
		if err == repository.ErrBriefNotFound {
			return model.BriefResponse{}, errors.NotFound("brief not found")
		}
		return model.BriefResponse{}, fmt.Errorf("failed to get brief: %w", err)
	}

	// Check ownership
	if brief.ClientID != userID {
		return model.BriefResponse{}, errors.Forbidden("not authorized to update this brief")
	}

	// Check status
	if brief.Status != model.BriefStatusDraft {
		return model.BriefResponse{}, ErrCannotUpdatePublished
	}

	// Update fields
	if req.Title != "" {
		brief.Title = req.Title
	}
	if req.Description != "" {
		brief.Description = req.Description
	}
	if req.Goals != "" {
		brief.Goals = req.Goals
	}
	if req.TargetAudience != "" {
		brief.TargetAudience = req.TargetAudience
	}
	if req.Tone != "" {
		brief.Tone = req.Tone
	}
	if req.StylePreferences != "" {
		brief.StylePreferences = req.StylePreferences
	}
	if req.CTA != "" {
		brief.CTA = req.CTA
	}
	if req.BountyBudget != nil {
		brief.BountyBudget = *req.BountyBudget
	}
	if req.SubmissionsLimit != nil {
		brief.SubmissionsLimit = *req.SubmissionsLimit
	}
	if req.IsBlind != nil {
		brief.IsBlind = *req.IsBlind
	}
	if len(req.Tags) > 0 {
		brief.Tags = req.Tags
	}

	brief.UpdatedAt = time.Now()

	if err := s.repo.UpdateBrief(ctx, brief); err != nil {
		return model.BriefResponse{}, fmt.Errorf("failed to update brief: %w", err)
	}

	return brief.ToResponse(), nil
}

// DeleteBrief deletes a brief
func (s *BriefService) DeleteBrief(ctx context.Context, userID, briefID uuid.UUID) error {
	brief, err := s.repo.GetBriefByID(ctx, briefID)
	if err != nil {
		if err == repository.ErrBriefNotFound {
			return errors.NotFound("brief not found")
		}
		return fmt.Errorf("failed to get brief: %w", err)
	}

	// Check ownership
	if brief.ClientID != userID {
		return errors.Forbidden("not authorized to delete this brief")
	}

	// Only allow deleting draft briefs
	if brief.Status != model.BriefStatusDraft {
		return errors.BadRequest("can only delete draft briefs")
	}

	if err := s.repo.DeleteBrief(ctx, briefID); err != nil {
		return fmt.Errorf("failed to delete brief: %w", err)
	}

	return nil
}

// ListBriefs lists briefs with filtering and pagination
func (s *BriefService) ListBriefs(ctx context.Context, userID *uuid.UUID, status *string, tags []string, page, limit int) (*model.ListBriefsResponse, error) {
	return s.repo.ListBriefs(ctx, userID, status, tags, page, limit)
}

// PublishBrief publishes a brief
func (s *BriefService) PublishBrief(ctx context.Context, userID, briefID uuid.UUID, req model.PublishBriefRequest) (model.BriefResponse, error) {
	brief, err := s.repo.GetBriefByID(ctx, briefID)
	if err != nil {
		if err == repository.ErrBriefNotFound {
			return model.BriefResponse{}, errors.NotFound("brief not found")
		}
		return model.BriefResponse{}, fmt.Errorf("failed to get brief: %w", err)
	}

	// Check ownership
	if brief.ClientID != userID {
		return model.BriefResponse{}, errors.Forbidden("not authorized to publish this brief")
	}

	// Check status
	if brief.Status == model.BriefStatusPublished {
		return model.BriefResponse{}, errors.BadRequest("brief is already published")
	}
	if brief.Status == model.BriefStatusClosed {
		return model.BriefResponse{}, ErrCannotPublishClosed
	}

	// For MVP, accept bounty_deposited from request (stub for DodoPayments check)
	// In production, this would verify with DodoPayments service
	if req.BountyDeposited {
		brief.BountyDeposited = true
	}

	// Require bounty deposited for publishing
	if !brief.BountyDeposited {
		return model.BriefResponse{}, ErrBountyNotDeposited
	}

	brief.Status = model.BriefStatusPublished
	brief.UpdatedAt = time.Now()

	if err := s.repo.PublishBrief(ctx, briefID); err != nil {
		return model.BriefResponse{}, fmt.Errorf("failed to publish brief: %w", err)
	}

	return brief.ToResponse(), nil
}

// CloseBrief closes a brief
func (s *BriefService) CloseBrief(ctx context.Context, userID, briefID uuid.UUID) (model.BriefResponse, error) {
	brief, err := s.repo.GetBriefByID(ctx, briefID)
	if err != nil {
		if err == repository.ErrBriefNotFound {
			return model.BriefResponse{}, errors.NotFound("brief not found")
		}
		return model.BriefResponse{}, fmt.Errorf("failed to get brief: %w", err)
	}

	// Check ownership
	if brief.ClientID != userID {
		return model.BriefResponse{}, errors.Forbidden("not authorized to close this brief")
	}

	// Check status
	if brief.Status == model.BriefStatusDraft {
		return model.BriefResponse{}, errors.BadRequest("cannot close draft brief, publish it first or delete it")
	}
	if brief.Status == model.BriefStatusClosed {
		return model.BriefResponse{}, errors.BadRequest("brief is already closed")
	}

	if err := s.repo.CloseBrief(ctx, briefID); err != nil {
		return model.BriefResponse{}, fmt.Errorf("failed to close brief: %w", err)
	}

	brief.Status = model.BriefStatusClosed
	return brief.ToResponse(), nil
}

// Interview processes an AI interview message and returns a structured brief (STUB)
func (s *BriefService) Interview(ctx context.Context, briefID uuid.UUID, req model.InterviewRequest) (model.InterviewResponse, error) {
	// This is a STUB implementation for MVP
	// In production, this would integrate with an AI service (OpenAI, Anthropic, etc.)

	// First, create or get the brief
	brief, err := s.repo.GetBriefByID(ctx, briefID)
	if err != nil {
		if err == repository.ErrBriefNotFound {
			return model.InterviewResponse{}, ErrInterviewNoBrief
		}
		return model.InterviewResponse{}, fmt.Errorf("failed to get brief: %w", err)
	}

	// Parse keywords from the message to build a mock structured brief
	// This is a STUB - real implementation would use AI
	message := req.Message

	// Update brief with interview answers if meaningful content found
	if message != "" {
		// Simple keyword matching to populate brief fields (STUB)
		if containsAny(message, []string{"brand", "product", "service", "business"}) {
			if brief.Goals == "" {
				brief.Goals = "Showcase our " + extractNoun(message) + " professionally"
			}
		}
		if containsAny(message, []string{"young", "gen z", "millennial", "professional", "b2b"}) {
			if brief.TargetAudience == "" {
				brief.TargetAudience = extractAudience(message)
			}
		}
		if containsAny(message, []string{"fun", "professional", "casual", "energetic", "serious", "humor"}) {
			if brief.Tone == "" {
				brief.Tone = extractTone(message)
			}
		}
		if containsAny(message, []string{"animated", "live action", "whiteboard", "motion graphics", "logo"}) {
			if brief.StylePreferences == "" {
				brief.StylePreferences = extractStyle(message)
			}
		}
		if containsAny(message, []string{"visit", "sign up", "buy", "click", "learn more", "register"}) {
			if brief.CTA == "" {
				brief.CTA = extractCTA(message)
			}
		}

		brief.UpdatedAt = time.Now()
		_ = s.repo.UpdateBrief(ctx, brief)

		// Store Q&A
		now := time.Now()
		q := &model.BriefQuestion{
			ID:        uuid.Must(uuid.NewV7()),
			BriefID:   briefID,
			Question:  "User message",
			Answer:    message,
			CreatedAt: now,
		}
		_ = s.repo.CreateBriefQuestion(ctx, q)
	}

	// Return mock structured brief (STUB)
	response := model.InterviewResponse{
		Message: "Thank you for sharing. Here's a summary of your brief:",
		StructuredBrief: model.StructuredBrief{
			Goals:            brief.Goals,
			TargetAudience:  brief.TargetAudience,
			Tone:            brief.Tone,
			StylePreferences: brief.StylePreferences,
			CTA:             brief.CTA,
		},
	}

	return response, nil
}

// containsAny checks if any of the keywords are in the message
func containsAny(message string, keywords []string) bool {
	for _, kw := range keywords {
		if contains(message, kw) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func extractNoun(message string) string {
	if contains(message, "brand") {
		return "brand"
	}
	if contains(message, "product") {
		return "product"
	}
	if contains(message, "service") {
		return "service"
	}
	return "offering"
}

func extractAudience(message string) string {
	if contains(message, "young") || contains(message, "gen z") {
		return "Gen Z"
	}
	if contains(message, "millennial") {
		return "Millennials"
	}
	if contains(message, "b2b") || contains(message, "professional") {
		return "Business professionals"
	}
	return "General audience"
}

func extractTone(message string) string {
	if contains(message, "fun") || contains(message, "humor") {
		return "Fun and playful"
	}
	if contains(message, "energetic") {
		return "Energetic"
	}
	if contains(message, "professional") || contains(message, "serious") {
		return "Professional"
	}
	return "Neutral"
}

func extractStyle(message string) string {
	if contains(message, "animated") {
		return "Animated"
	}
	if contains(message, "live action") {
		return "Live action"
	}
	if contains(message, "whiteboard") {
		return "Whiteboard"
	}
	if contains(message, "motion graphics") {
		return "Motion graphics"
	}
	return "Dynamic"
}

func extractCTA(message string) string {
	if contains(message, "visit") {
		return "Visit our website"
	}
	if contains(message, "sign up") || contains(message, "register") {
		return "Sign up now"
	}
	if contains(message, "buy") {
		return "Buy now"
	}
	if contains(message, "click") {
		return "Learn more"
	}
	return "Get started"
}

// GetMatchingBriefs gets briefs matching editor tags
func (s *BriefService) GetMatchingBriefs(ctx context.Context, tags []string, page, limit int) (*model.ListBriefsResponse, error) {
	return s.repo.GetMatchingBriefs(ctx, tags, page, limit)
}

// MarkBriefViewed marks a brief as viewed by an editor
func (s *BriefService) MarkBriefViewed(ctx context.Context, briefID uuid.UUID, req model.ViewBriefRequest) error {
	editorID, err := uuid.Parse(req.EditorID)
	if err != nil {
		return errors.BadRequest("invalid editor_id")
	}

	// Verify brief exists
	_, err = s.repo.GetBriefByID(ctx, briefID)
	if err != nil {
		if err == repository.ErrBriefNotFound {
			return errors.NotFound("brief not found")
		}
		return fmt.Errorf("failed to get brief: %w", err)
	}

	if err := s.repo.MarkBriefViewed(ctx, briefID, editorID); err != nil {
		return fmt.Errorf("failed to mark brief viewed: %w", err)
	}

	return nil
}

// GetRawFootageUploadURL generates a presigned upload URL for raw footage
func (s *BriefService) GetRawFootageUploadURL(ctx context.Context, userID, briefID uuid.UUID, filename string) (*model.UploadRawFootageResponse, error) {
	// Verify storage is available
	if s.storage == nil {
		return nil, ErrNoStorageClient
	}

	// Get brief and verify ownership
	brief, err := s.repo.GetBriefByID(ctx, briefID)
	if err != nil {
		if err == repository.ErrBriefNotFound {
			return nil, errors.NotFound("brief not found")
		}
		return nil, fmt.Errorf("failed to get brief: %w", err)
	}

	// Only the brief owner (client) can upload raw footage
	if brief.ClientID != userID {
		return nil, errors.Forbidden("not authorized to upload raw footage for this brief")
	}

	// Validate filename
	if filename == "" {
		return nil, errors.BadRequest("filename is required")
	}

	// Generate Storj key: briefs/{briefID}/raw_footage/{filename}
	storjKey := fmt.Sprintf("briefs/%s/raw_footage/%s", briefID.String(), filename)

	// Generate presigned upload URL
	uploadURL, expiresIn, err := s.storage.GeneratePresignedUploadURL(ctx, storjKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return &model.UploadRawFootageResponse{
		UploadURL: uploadURL,
		StorjKey:  storjKey,
		ExpiresIn: expiresIn,
	}, nil
}

// ConfirmRawFootageUpload confirms raw footage upload is complete
func (s *BriefService) ConfirmRawFootageUpload(ctx context.Context, userID, briefID uuid.UUID, storjKey string) error {
	// Verify storage is available
	if s.storage == nil {
		return ErrNoStorageClient
	}

	// Get brief and verify ownership
	brief, err := s.repo.GetBriefByID(ctx, briefID)
	if err != nil {
		if err == repository.ErrBriefNotFound {
			return errors.NotFound("brief not found")
		}
		return fmt.Errorf("failed to get brief: %w", err)
	}

	// Only the brief owner (client) can confirm upload
	if brief.ClientID != userID {
		return errors.Forbidden("not authorized to confirm upload for this brief")
	}

	// Validate storj key
	if storjKey == "" {
		return errors.BadRequest("storj_key is required")
	}

	// Verify the file exists in storage
	exists, err := s.storage.FileExists(ctx, storjKey)
	if err != nil {
		return fmt.Errorf("failed to verify file: %w", err)
	}
	if !exists {
		return errors.BadRequest("file not found in storage")
	}

	// Update brief with raw footage info
	if err := s.repo.UpdateRawFootage(ctx, briefID, storjKey); err != nil {
		return fmt.Errorf("failed to update brief: %w", err)
	}

	return nil
}

// GetRawFootageDownloadURL generates a presigned download URL for raw footage
func (s *BriefService) GetRawFootageDownloadURL(ctx context.Context, userID, briefID uuid.UUID) (*model.RawFootageDownloadURLResponse, error) {
	// Verify storage is available
	if s.storage == nil {
		return nil, ErrNoStorageClient
	}

	// Get brief
	brief, err := s.repo.GetBriefByID(ctx, briefID)
	if err != nil {
		if err == repository.ErrBriefNotFound {
			return nil, errors.NotFound("brief not found")
		}
		return nil, fmt.Errorf("failed to get brief: %w", err)
	}

	// Check if raw footage exists
	if !brief.HasRawFootage || brief.RawFootageStorjKey == "" {
		return nil, ErrNoRawFootage
	}

	// Only the brief owner (client) or matched editors can download
	isOwner := brief.ClientID == userID
	isEditorMatched := false

	if !isOwner {
		// Check if user has viewed the brief (indicating they're a matched editor)
		viewers, err := s.repo.GetBriefViewers(ctx, briefID)
		if err == nil {
			for _, viewerID := range viewers {
				if viewerID == userID {
					isEditorMatched = true
					break
				}
			}
		}
	}

	if !isOwner && !isEditorMatched {
		return nil, errors.Forbidden("not authorized to download raw footage for this brief")
	}

	// Generate presigned download URL
	downloadURL, expiresIn, err := s.storage.GeneratePresignedDownloadURL(ctx, brief.RawFootageStorjKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return &model.RawFootageDownloadURLResponse{
		DownloadURL: downloadURL,
		ExpiresIn:   expiresIn,
	}, nil
}