package model

import (
	"time"

	"github.com/google/uuid"
)

// Campaign statuses
const (
	CampaignStatusDraft   = "draft"
	CampaignStatusActive = "active"
	CampaignStatusPaused = "paused"
	CampaignStatusEnded = "ended"
)

// Campaign platforms
const (
	CampaignPlatformMeta  = "meta"
	CampaignPlatformTikTok = "tiktok"
)

// CampaignVideo statuses
const (
	CampaignVideoStatusActive = "active"
	CampaignVideoStatusPaused = "paused"
)

// Budget types
const (
	BudgetTypeDaily = "daily"
	BudgetTypeTotal = "total"
)

// User roles
const (
	UserRoleAdmin        = "admin"
	UserRoleAdSpecialist = "ad_specialist"
	UserRoleClient       = "client"
)

// Campaign represents an advertising campaign
type Campaign struct {
	ID           string    `json:"id"`
	AdSpecialistID uuid.UUID `json:"ad_specialist_id"`
	ClientID    uuid.UUID `json:"client_id"`
	BriefID     *uuid.UUID `json:"brief_id,omitempty"`
	Name        string    `json:"name"`
	Description string   `json:"description"`
	Status      string    `json:"status"`
	Platform    string    `json:"platform"`
	AdAccountID string   `json:"ad_account_id,omitempty"`
	TotalBudget float64  `json:"total_budget"`
	DailyBudget float64  `json:"daily_budget"`
	StartDate   time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CampaignVideo links a video to a campaign
type CampaignVideo struct {
	ID         string    `json:"id"`
	CampaignID string    `json:"campaign_id"`
	VideoID   uuid.UUID `json:"video_id"`
	Status    string    `json:"status"`
	AddedAt   time.Time `json:"added_at"`
}

// AdAccount represents a placeholder for ad platform accounts
type AdAccount struct {
	ID          string    `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Platform   string   `json:"platform"`
	AccountID string   `json:"account_id"`
	Name       string   `json:"name"`
	Status     string   `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CampaignBudget tracks campaign spending
type CampaignBudget struct {
	ID         string    `json:"id"`
	CampaignID string   `json:"campaign_id"`
	Amount    float64   `json:"amount"`
	Type      string    `json:"type"` // daily or total
	Spent     float64   `json:"spent"`
	Remaining float64  `json:"remaining"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Request/Response types

// CreateCampaignRequest represents the request to create a campaign
type CreateCampaignRequest struct {
	Name         string   `json:"name"`
	Description string   `json:"description"`
	BriefID     string   `json:"brief_id"`
	Platform    string   `json:"platform"`
	TotalBudget float64  `json:"total_budget"`
	DailyBudget float64  `json:"daily_budget"`
	StartDate   string   `json:"start_date"`
	EndDate     string   `json:"end_date"`
	VideoIDs    []string `json:"video_ids"`
}

// UpdateCampaignRequest represents the request to update a campaign
type UpdateCampaignRequest struct {
	Name         string  `json:"name"`
	Description string  `json:"description"`
	Platform    string  `json:"platform"`
	TotalBudget float64 `json:"total_budget"`
	DailyBudget float64 `json:"daily_budget"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
}

// CampaignResponse represents the campaign response
type CampaignResponse struct {
	ID           string              `json:"id"`
	AdSpecialistID string            `json:"ad_specialist_id"`
	ClientID     string              `json:"client_id"`
	BriefID     *string            `json:"brief_id,omitempty"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Status      string              `json:"status"`
	Platform    string              `json:"platform"`
	AdAccountID string             `json:"ad_account_id,omitempty"`
	TotalBudget float64            `json:"total_budget"`
	DailyBudget float64           `json:"daily_budget"`
	StartDate   string             `json:"start_date"`
	EndDate     *string           `json:"end_date,omitempty"`
	Videos      []CampaignVideoResponse `json:"videos"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

// CampaignVideoResponse represents a campaign video response
type CampaignVideoResponse struct {
	ID         string `json:"id"`
	CampaignID string `json:"campaign_id"`
	VideoID   string `json:"video_id"`
	Status    string `json:"status"`
	AddedAt   string `json:"added_at"`
}

// AddVideoRequest represents the request to add a video to a campaign
type AddVideoRequest struct {
	VideoID string `json:"video_id"`
}

// UpdateBudgetRequest represents the request to update budget
type UpdateBudgetRequest struct {
	Amount float64 `json:"amount"`
	Type   string  `json:"type"`
}

// CampaignBudgetResponse represents the campaign budget response
type CampaignBudgetResponse struct {
	ID          string  `json:"id"`
	CampaignID string  `json:"campaign_id"`
	Amount     float64 `json:"amount"`
	Type       string  `json:"type"`
	Spent      float64 `json:"spent"`
	Remaining  float64 `json:"remaining"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// ListCampaignsResponse represents the response for listing campaigns
type ListCampaignsResponse struct {
	Campaigns []CampaignResponse `json:"campaigns"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	Total     int                `json:"total"`
}

// AdAccountRequest represents the request to create an ad account
type AdAccountRequest struct {
	Platform   string `json:"platform"`
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
}

// AdAccountResponse represents the ad account response
type AdAccountResponse struct {
	ID         string `json:"id"`
	UserID    string `json:"user_id"`
	Platform  string `json:"platform"`
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListAdAccountsResponse represents the response for listing ad accounts
type ListAdAccountsResponse struct {
	AdAccounts []AdAccountResponse `json:"ad_accounts"`
	Page      int                   `json:"page"`
	Limit    int                   `json:"limit"`
	Total    int                   `json:"total"`
}

// ToResponse converts Campaign to CampaignResponse
func (c *Campaign) ToResponse() CampaignResponse {
	resp := CampaignResponse{
		ID:            c.ID,
		AdSpecialistID: c.AdSpecialistID.String(),
		ClientID:     c.ClientID.String(),
		Name:         c.Name,
		Description: c.Description,
		Status:       c.Status,
		Platform:     c.Platform,
		AdAccountID:  c.AdAccountID,
		TotalBudget:  c.TotalBudget,
		DailyBudget:  c.DailyBudget,
		StartDate:    c.StartDate.Format(time.RFC3339),
		CreatedAt:   c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
	}

	if c.BriefID != nil {
		briefID := c.BriefID.String()
		resp.BriefID = &briefID
	}

	if c.EndDate != nil {
		endDate := c.EndDate.Format(time.RFC3339)
		resp.EndDate = &endDate
	}

	return resp
}

// ToResponse converts CampaignVideo to CampaignVideoResponse
func (cv *CampaignVideo) ToResponse() CampaignVideoResponse {
	return CampaignVideoResponse{
		ID:         cv.ID,
		CampaignID: cv.CampaignID,
		VideoID:    cv.VideoID.String(),
		Status:    cv.Status,
		AddedAt:   cv.AddedAt.Format(time.RFC3339),
	}
}

// ToResponse converts CampaignBudget to CampaignBudgetResponse
func (cb *CampaignBudget) ToResponse() CampaignBudgetResponse {
	return CampaignBudgetResponse{
		ID:         cb.ID,
		CampaignID: cb.CampaignID,
		Amount:    cb.Amount,
		Type:      cb.Type,
		Spent:     cb.Spent,
		Remaining: cb.Remaining,
		CreatedAt: cb.CreatedAt.Format(time.RFC3339),
		UpdatedAt: cb.UpdatedAt.Format(time.RFC3339),
	}
}

// ToResponse converts AdAccount to AdAccountResponse
func (aa *AdAccount) ToResponse() AdAccountResponse {
	return AdAccountResponse{
		ID:         aa.ID,
		UserID:    aa.UserID.String(),
		Platform:  aa.Platform,
		AccountID: aa.AccountID,
		Name:      aa.Name,
		Status:    aa.Status,
		CreatedAt: aa.CreatedAt.Format(time.RFC3339),
		UpdatedAt: aa.UpdatedAt.Format(time.RFC3339),
	}
}