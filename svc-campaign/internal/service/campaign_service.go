package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/pkg/natsclient"

	"github.com/videoforge/backend/svc-campaign/internal/model"
	"github.com/videoforge/backend/svc-campaign/internal/repository"
)

var (
	ErrUnauthorized          = errors.Unauthorized("unauthorized")
	ErrForbidden             = errors.Forbidden("access denied")
	ErrCampaignNotFound       = errors.NotFound("campaign not found")
	ErrInvalidStatus         = errors.BadRequest("invalid campaign status")
	ErrCannotStartCampaign  = errors.BadRequest("cannot start campaign")
	ErrCannotPauseCampaign   = errors.BadRequest("cannot pause campaign")
	ErrCannotEndCampaign    = errors.BadRequest("cannot end campaign")
	ErrCannotUpdateCampaign = errors.BadRequest("cannot update campaign")
	ErrVideoNotFound         = errors.NotFound("video not found")
	ErrVideoAlreadyExists     = errors.Conflict("video already in campaign")
	ErrAdAccountNotFound    = errors.NotFound("ad account not found")
	ErrBudgetNotFound        = errors.NotFound("budget not found")
	ErrInvalidPlatform      = errors.BadRequest("invalid platform")
	ErrInvalidBudgetType    = errors.BadRequest("invalid budget type")
)

// CampaignService handles campaign business logic
type CampaignService struct {
	repo repository.CampaignRepo
	nats natsclient.NATSClient
}

// NewCampaignService creates a new CampaignService
func NewCampaignService(repo repository.CampaignRepo, nats natsclient.NATSClient) *CampaignService {
	return &CampaignService{
		repo: repo,
		nats: nats,
	}
}

// CreateCampaign creates a new campaign
func (s *CampaignService) CreateCampaign(ctx context.Context, userID uuid.UUID, role string, req model.CreateCampaignRequest) (model.CampaignResponse, error) {
	// Role check - only ad_specialist can create
	if role != model.UserRoleAdSpecialist {
		return model.CampaignResponse{}, ErrForbidden
	}

	// Validate required fields
	if req.Name == "" {
		return model.CampaignResponse{}, errors.BadRequest("name is required")
	}
	if req.Platform == "" {
		return model.CampaignResponse{}, errors.BadRequest("platform is required")
	}
	if req.Platform != model.CampaignPlatformMeta && req.Platform != model.CampaignPlatformTikTok {
		return model.CampaignResponse{}, ErrInvalidPlatform
	}

	// Parse dates
	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return model.CampaignResponse{}, errors.BadRequest("invalid start_date format")
	}

	var endDate *time.Time
	if req.EndDate != "" {
		t, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			return model.CampaignResponse{}, errors.BadRequest("invalid end_date format")
		}
		endDate = &t
	}

	// Parse brief_id if provided
	var briefID *uuid.UUID
	if req.BriefID != "" {
		id, err := uuid.Parse(req.BriefID)
		if err != nil {
			return model.CampaignResponse{}, errors.BadRequest("invalid brief_id")
		}
		briefID = &id
	}

	now := time.Now()
	campaign := &model.Campaign{
		ID:            uuid.Must(uuid.NewV7()).String(),
		AdSpecialistID: userID,
		ClientID:      userID, // For MVP, ad specialist is also the client
		BriefID:       briefID,
		Name:          req.Name,
		Description:   req.Description,
		Status:        model.CampaignStatusDraft,
		Platform:      req.Platform,
		TotalBudget:   req.TotalBudget,
		DailyBudget:   req.DailyBudget,
		StartDate:     startDate,
		EndDate:       endDate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Create campaign
	if err := s.repo.CreateCampaign(ctx, campaign); err != nil {
		return model.CampaignResponse{}, fmt.Errorf("failed to create campaign: %w", err)
	}

	// Add videos if provided
	for _, videoIDStr := range req.VideoIDs {
		videoID, err := uuid.Parse(videoIDStr)
		if err != nil {
			continue
		}

		cv := &model.CampaignVideo{
			ID:         uuid.Must(uuid.NewV7()).String(),
			CampaignID: campaign.ID,
			VideoID:    videoID,
			Status:     model.CampaignVideoStatusActive,
			AddedAt:    now,
		}

		_ = s.repo.CreateCampaignVideo(ctx, cv)
	}

	// Create initial budget
	budget := &model.CampaignBudget{
		ID:          uuid.Must(uuid.NewV7()).String(),
		CampaignID: campaign.ID,
		Amount:     req.TotalBudget,
		Type:       model.BudgetTypeTotal,
		Spent:      0,
		Remaining:  req.TotalBudget,
		CreatedAt:  now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateCampaignBudget(ctx, budget); err != nil {
		return model.CampaignResponse{}, fmt.Errorf("failed to create campaign budget: %w", err)
	}

	// Emit NATS event
	s.emitEvent(ctx, "campaign.created", map[string]string{
		"campaign_id": campaign.ID,
	})

	return campaign.ToResponse(), nil
}

// GetCampaign gets a campaign by ID
func (s *CampaignService) GetCampaign(ctx context.Context, userID uuid.UUID, role string, campaignID string) (model.CampaignResponse, error) {
	campaign, err := s.repo.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == repository.ErrCampaignNotFound {
			return model.CampaignResponse{}, ErrCampaignNotFound
		}
		return model.CampaignResponse{}, fmt.Errorf("failed to get campaign: %w", err)
	}

	// Authorization check
	if role != model.UserRoleAdmin && role != model.UserRoleClient {
		if campaign.AdSpecialistID != userID {
			return model.CampaignResponse{}, ErrForbidden
		}
	} else if role == model.UserRoleClient {
		// Client can only view their own campaigns
		if campaign.ClientID != userID {
			return model.CampaignResponse{}, ErrForbidden
		}
	}

	// Get videos
	videos, _ := s.repo.GetCampaignVideos(ctx, campaignID)

	resp := campaign.ToResponse()
	resp.Videos = make([]model.CampaignVideoResponse, len(videos))
	for i, v := range videos {
		resp.Videos[i] = v.ToResponse()
	}

	return resp, nil
}

// UpdateCampaign updates a campaign
func (s *CampaignService) UpdateCampaign(ctx context.Context, userID uuid.UUID, role string, campaignID string, req model.UpdateCampaignRequest) (model.CampaignResponse, error) {
	campaign, err := s.repo.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == repository.ErrCampaignNotFound {
			return model.CampaignResponse{}, ErrCampaignNotFound
		}
		return model.CampaignResponse{}, fmt.Errorf("failed to get campaign: %w", err)
	}

	// Authorization check - only ad_specialist can update
	if role != model.UserRoleAdmin {
		if campaign.AdSpecialistID != userID {
			return model.CampaignResponse{}, ErrForbidden
		}
		// Can only update if draft or paused
		if campaign.Status != model.CampaignStatusDraft && campaign.Status != model.CampaignStatusPaused {
			return model.CampaignResponse{}, ErrCannotUpdateCampaign
		}
	}

	// Update fields
	if req.Name != "" {
		campaign.Name = req.Name
	}
	if req.Description != "" {
		campaign.Description = req.Description
	}
	if req.Platform != "" {
		if req.Platform != model.CampaignPlatformMeta && req.Platform != model.CampaignPlatformTikTok {
			return model.CampaignResponse{}, ErrInvalidPlatform
		}
		campaign.Platform = req.Platform
	}
	if req.TotalBudget > 0 {
		campaign.TotalBudget = req.TotalBudget
	}
	if req.DailyBudget > 0 {
		campaign.DailyBudget = req.DailyBudget
	}
	if req.StartDate != "" {
		startDate, err := time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			return model.CampaignResponse{}, errors.BadRequest("invalid start_date format")
		}
		campaign.StartDate = startDate
	}
	if req.EndDate != "" {
		endDate, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			return model.CampaignResponse{}, errors.BadRequest("invalid end_date format")
		}
		campaign.EndDate = &endDate
	}

	campaign.UpdatedAt = time.Now()

	if err := s.repo.UpdateCampaign(ctx, campaign); err != nil {
		return model.CampaignResponse{}, fmt.Errorf("failed to update campaign: %w", err)
	}

	return campaign.ToResponse(), nil
}

// ListCampaigns lists campaigns with filtering and pagination
func (s *CampaignService) ListCampaigns(ctx context.Context, userID uuid.UUID, role string, status *string, clientID *string, adSpecialistID *string, briefID *string, page, limit int) (*model.ListCampaignsResponse, error) {
	filter := &repository.CampaignFilter{}

	// Apply role-based filtering
	if role == model.UserRoleClient {
		clientIDStr := userID.String()
		filter.ClientID = &clientIDStr
	} else if role == model.UserRoleAdSpecialist && adSpecialistID == nil {
		adSpecIDStr := userID.String()
		filter.AdSpecialistID = &adSpecIDStr
	}

	// Apply additional filters
	if status != nil {
		filter.Status = status
	}
	if clientID != nil {
		filter.ClientID = clientID
	}
	if adSpecialistID != nil {
		filter.AdSpecialistID = adSpecialistID
	}
	if briefID != nil {
		filter.BriefID = briefID
	}

	campaigns, total, err := s.repo.ListCampaigns(ctx, filter, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list campaigns: %w", err)
	}

	resp := &model.ListCampaignsResponse{
		Campaigns: make([]model.CampaignResponse, len(campaigns)),
		Page:      page,
		Limit:     limit,
		Total:     total,
	}

	for i, c := range campaigns {
		resp.Campaigns[i] = c.ToResponse()
	}

	return resp, nil
}

// StartCampaign starts a campaign
func (s *CampaignService) StartCampaign(ctx context.Context, userID uuid.UUID, role string, campaignID string) (model.CampaignResponse, error) {
	campaign, err := s.repo.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == repository.ErrCampaignNotFound {
			return model.CampaignResponse{}, ErrCampaignNotFound
		}
		return model.CampaignResponse{}, fmt.Errorf("failed to get campaign: %w", err)
	}

	// Authorization check
	if role != model.UserRoleAdmin {
		if campaign.AdSpecialistID != userID {
			return model.CampaignResponse{}, ErrForbidden
		}
	}

	// Can only start if draft or paused
	if campaign.Status != model.CampaignStatusDraft && campaign.Status != model.CampaignStatusPaused {
		return model.CampaignResponse{}, ErrCannotStartCampaign
	}

	campaign.Status = model.CampaignStatusActive
	campaign.UpdatedAt = time.Now()

	if err := s.repo.UpdateCampaign(ctx, campaign); err != nil {
		return model.CampaignResponse{}, fmt.Errorf("failed to start campaign: %w", err)
	}

	// Emit NATS event
	s.emitEvent(ctx, "campaign.started", map[string]string{
		"campaign_id": campaign.ID,
	})

	return campaign.ToResponse(), nil
}

// PauseCampaign pauses a campaign
func (s *CampaignService) PauseCampaign(ctx context.Context, userID uuid.UUID, role string, campaignID string) (model.CampaignResponse, error) {
	campaign, err := s.repo.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == repository.ErrCampaignNotFound {
			return model.CampaignResponse{}, ErrCampaignNotFound
		}
		return model.CampaignResponse{}, fmt.Errorf("failed to get campaign: %w", err)
	}

	// Authorization check
	if role != model.UserRoleAdmin {
		if campaign.AdSpecialistID != userID {
			return model.CampaignResponse{}, ErrForbidden
		}
	}

	// Can only pause if active
	if campaign.Status != model.CampaignStatusActive {
		return model.CampaignResponse{}, ErrCannotPauseCampaign
	}

	campaign.Status = model.CampaignStatusPaused
	campaign.UpdatedAt = time.Now()

	if err := s.repo.UpdateCampaign(ctx, campaign); err != nil {
		return model.CampaignResponse{}, fmt.Errorf("failed to pause campaign: %w", err)
	}

	return campaign.ToResponse(), nil
}

// EndCampaign ends a campaign
func (s *CampaignService) EndCampaign(ctx context.Context, userID uuid.UUID, role string, campaignID string) (model.CampaignResponse, error) {
	campaign, err := s.repo.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == repository.ErrCampaignNotFound {
			return model.CampaignResponse{}, ErrCampaignNotFound
		}
		return model.CampaignResponse{}, fmt.Errorf("failed to get campaign: %w", err)
	}

	// Authorization check
	if role != model.UserRoleAdmin {
		if campaign.AdSpecialistID != userID {
			return model.CampaignResponse{}, ErrForbidden
		}
	}

	// Can only end if active or paused
	if campaign.Status != model.CampaignStatusActive && campaign.Status != model.CampaignStatusPaused {
		return model.CampaignResponse{}, ErrCannotEndCampaign
	}

	campaign.Status = model.CampaignStatusEnded
	campaign.UpdatedAt = time.Now()

	if err := s.repo.UpdateCampaign(ctx, campaign); err != nil {
		return model.CampaignResponse{}, fmt.Errorf("failed to end campaign: %w", err)
	}

	// Emit NATS event
	s.emitEvent(ctx, "campaign.ended", map[string]string{
		"campaign_id": campaign.ID,
	})

	return campaign.ToResponse(), nil
}

// AddVideo adds a video to a campaign
func (s *CampaignService) AddVideo(ctx context.Context, userID uuid.UUID, role string, campaignID string, req model.AddVideoRequest) (model.CampaignVideoResponse, error) {
	campaign, err := s.repo.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == repository.ErrCampaignNotFound {
			return model.CampaignVideoResponse{}, ErrCampaignNotFound
		}
		return model.CampaignVideoResponse{}, fmt.Errorf("failed to get campaign: %w", err)
	}

	// Authorization check
	if role != model.UserRoleAdmin {
		if campaign.AdSpecialistID != userID {
			return model.CampaignVideoResponse{}, ErrForbidden
		}
	}

	videoID, err := uuid.Parse(req.VideoID)
	if err != nil {
		return model.CampaignVideoResponse{}, errors.BadRequest("invalid video_id")
	}

	// Check if video already exists
	existing, _ := s.repo.GetCampaignVideo(ctx, campaignID, req.VideoID)
	if existing != nil {
		return model.CampaignVideoResponse{}, ErrVideoAlreadyExists
	}

	cv := &model.CampaignVideo{
		ID:         uuid.Must(uuid.NewV7()).String(),
		CampaignID: campaignID,
		VideoID:    videoID,
		Status:    model.CampaignVideoStatusActive,
		AddedAt:   time.Now(),
	}

	if err := s.repo.CreateCampaignVideo(ctx, cv); err != nil {
		return model.CampaignVideoResponse{}, fmt.Errorf("failed to add video: %w", err)
	}

	return cv.ToResponse(), nil
}

// RemoveVideo removes a video from a campaign
func (s *CampaignService) RemoveVideo(ctx context.Context, userID uuid.UUID, role string, campaignID, videoID string) error {
	campaign, err := s.repo.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == repository.ErrCampaignNotFound {
			return ErrCampaignNotFound
		}
		return fmt.Errorf("failed to get campaign: %w", err)
	}

	// Authorization check
	if role != model.UserRoleAdmin {
		if campaign.AdSpecialistID != userID {
			return ErrForbidden
		}
	}

	if err := s.repo.DeleteCampaignVideo(ctx, campaignID, videoID); err != nil {
		if err == repository.ErrCampaignVideoNotFound {
			return ErrVideoNotFound
		}
		return fmt.Errorf("failed to remove video: %w", err)
	}

	return nil
}

// GetBudget gets campaign budget
func (s *CampaignService) GetBudget(ctx context.Context, userID uuid.UUID, role string, campaignID string) (model.CampaignBudgetResponse, error) {
	campaign, err := s.repo.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == repository.ErrCampaignNotFound {
			return model.CampaignBudgetResponse{}, ErrCampaignNotFound
		}
		return model.CampaignBudgetResponse{}, fmt.Errorf("failed to get campaign: %w", err)
	}

	// Authorization check
	if role != model.UserRoleAdmin {
		if campaign.AdSpecialistID != userID {
			return model.CampaignBudgetResponse{}, ErrForbidden
		}
	}

	budget, err := s.repo.GetCampaignBudget(ctx, campaignID)
	if err != nil {
		if err == repository.ErrBudgetNotFound {
			return model.CampaignBudgetResponse{}, ErrBudgetNotFound
		}
		return model.CampaignBudgetResponse{}, fmt.Errorf("failed to get budget: %w", err)
	}

	return budget.ToResponse(), nil
}

// UpdateBudget updates campaign budget
func (s *CampaignService) UpdateBudget(ctx context.Context, userID uuid.UUID, role string, campaignID string, req model.UpdateBudgetRequest) (model.CampaignBudgetResponse, error) {
	campaign, err := s.repo.GetCampaignByID(ctx, campaignID)
	if err != nil {
		if err == repository.ErrCampaignNotFound {
			return model.CampaignBudgetResponse{}, ErrCampaignNotFound
		}
		return model.CampaignBudgetResponse{}, fmt.Errorf("failed to get campaign: %w", err)
	}

	// Authorization check
	if role != model.UserRoleAdmin {
		if campaign.AdSpecialistID != userID {
			return model.CampaignBudgetResponse{}, ErrForbidden
		}
	}

	// Validate budget type
	if req.Type != model.BudgetTypeDaily && req.Type != model.BudgetTypeTotal {
		return model.CampaignBudgetResponse{}, ErrInvalidBudgetType
	}

	existingBudget, err := s.repo.GetCampaignBudget(ctx, campaignID)
	if err != nil {
		if err == repository.ErrBudgetNotFound {
			return model.CampaignBudgetResponse{}, ErrBudgetNotFound
		}
		return model.CampaignBudgetResponse{}, fmt.Errorf("failed to get budget: %w", err)
	}

	// Update budget
	oldSpent := existingBudget.Spent
	existingBudget.Amount = req.Amount
	existingBudget.Type = req.Type
	existingBudget.Spent = oldSpent
	existingBudget.Remaining = req.Amount - oldSpent
	existingBudget.UpdatedAt = time.Now()

	if err := s.repo.UpdateCampaignBudget(ctx, existingBudget); err != nil {
		return model.CampaignBudgetResponse{}, fmt.Errorf("failed to update budget: %w", err)
	}

	return existingBudget.ToResponse(), nil
}

// CreateAdAccount creates an ad account placeholder
func (s *CampaignService) CreateAdAccount(ctx context.Context, userID uuid.UUID, role string, req model.AdAccountRequest) (model.AdAccountResponse, error) {
	// Validate platform
	if req.Platform != model.CampaignPlatformMeta && req.Platform != model.CampaignPlatformTikTok {
		return model.AdAccountResponse{}, ErrInvalidPlatform
	}

	now := time.Now()
	adAccount := &model.AdAccount{
		ID:         uuid.Must(uuid.NewV7()).String(),
		UserID:     userID,
		Platform:   req.Platform,
		AccountID: req.AccountID,
		Name:      req.Name,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateAdAccount(ctx, adAccount); err != nil {
		return model.AdAccountResponse{}, fmt.Errorf("failed to create ad account: %w", err)
	}

	return adAccount.ToResponse(), nil
}

// ListAdAccounts lists ad accounts
func (s *CampaignService) ListAdAccounts(ctx context.Context, userID uuid.UUID, page, limit int) (*model.ListAdAccountsResponse, error) {
	accounts, total, err := s.repo.GetAdAccountsByUserID(ctx, userID.String(), page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list ad accounts: %w", err)
	}

	resp := &model.ListAdAccountsResponse{
		AdAccounts: make([]model.AdAccountResponse, len(accounts)),
		Page:       page,
		Limit:      limit,
		Total:      total,
	}

	for i, aa := range accounts {
		resp.AdAccounts[i] = aa.ToResponse()
	}

	return resp, nil
}

// emitEvent emits a NATS event
func (s *CampaignService) emitEvent(ctx context.Context, subject string, data map[string]string) {
	if s.nats == nil {
		return
	}

	payload, _ := json.Marshal(data)
	_ = s.nats.Publish(subject, payload)
}