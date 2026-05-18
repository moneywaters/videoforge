package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"svc-campaign/internal/model"
)

// CampaignRepo defines the interface for campaign data access
type CampaignRepo interface {
	CreateCampaign(ctx context.Context, campaign *model.Campaign) error
	GetCampaignByID(ctx context.Context, id string) (*model.Campaign, error)
	UpdateCampaign(ctx context.Context, campaign *model.Campaign) error
	DeleteCampaign(ctx context.Context, id string) error
	ListCampaigns(ctx context.Context, filter *CampaignFilter, page, limit int) ([]*model.Campaign, int, error)

	CreateCampaignVideo(ctx context.Context, cv *model.CampaignVideo) error
	GetCampaignVideo(ctx context.Context, campaignID, videoID string) (*model.CampaignVideo, error)
	DeleteCampaignVideo(ctx context.Context, campaignID, videoID string) error
	GetCampaignVideos(ctx context.Context, campaignID string) ([]*model.CampaignVideo, error)

	CreateAdAccount(ctx context.Context, aa *model.AdAccount) error
	GetAdAccountByID(ctx context.Context, id string) (*model.AdAccount, error)
	GetAdAccountsByUserID(ctx context.Context, userID string, page, limit int) ([]*model.AdAccount, int, error)
	UpdateAdAccount(ctx context.Context, aa *model.AdAccount) error

	CreateCampaignBudget(ctx context.Context, cb *model.CampaignBudget) error
	GetCampaignBudget(ctx context.Context, campaignID string) (*model.CampaignBudget, error)
	UpdateCampaignBudget(ctx context.Context, cb *model.CampaignBudget) error
}

// CampaignFilter filters campaigns
type CampaignFilter struct {
	Status         *string
	ClientID      *string
	AdSpecialistID *string
	BriefID       *string
}

// PGCampaignRepo is a PostgreSQL implementation of CampaignRepo
type PGCampaignRepo struct {
	pool *pgxpool.Pool
}

// NewCampaignRepo creates a new campaign repository
func NewCampaignRepo(pool *pgxpool.Pool) *PGCampaignRepo {
	return &PGCampaignRepo{pool: pool}
}

// CreateCampaign creates a new campaign
func (r *PGCampaignRepo) CreateCampaign(ctx context.Context, campaign *model.Campaign) error {
	query := `
		INSERT INTO campaigns (
			id, ad_specialist_id, client_id, brief_id, name, description,
			status, platform, ad_account_id, total_budget, daily_budget,
			start_date, end_date, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err := r.pool.Exec(ctx, query,
		campaign.ID,
		campaign.AdSpecialistID,
		campaign.ClientID,
		campaign.BriefID,
		campaign.Name,
		campaign.Description,
		campaign.Status,
		campaign.Platform,
		campaign.AdAccountID,
		campaign.TotalBudget,
		campaign.DailyBudget,
		campaign.StartDate,
		campaign.EndDate,
		campaign.CreatedAt,
		campaign.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create campaign: %w", err)
	}

	return nil
}

// GetCampaignByID gets a campaign by ID
func (r *PGCampaignRepo) GetCampaignByID(ctx context.Context, id string) (*model.Campaign, error) {
	query := `
		SELECT id, ad_specialist_id, client_id, brief_id, name, description,
			status, platform, ad_account_id, total_budget, daily_budget,
			start_date, end_date, created_at, updated_at
		FROM campaigns
		WHERE id = $1
	`

	var campaign model.Campaign
	var briefID *uuid.UUID
	var endDate *time.Time

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&campaign.ID,
		&campaign.AdSpecialistID,
		&campaign.ClientID,
		&briefID,
		&campaign.Name,
		&campaign.Description,
		&campaign.Status,
		&campaign.Platform,
		&campaign.AdAccountID,
		&campaign.TotalBudget,
		&campaign.DailyBudget,
		&campaign.StartDate,
		&endDate,
		&campaign.CreatedAt,
		&campaign.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCampaignNotFound
		}
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	campaign.BriefID = briefID
	campaign.EndDate = endDate

	return &campaign, nil
}

// UpdateCampaign updates a campaign
func (r *PGCampaignRepo) UpdateCampaign(ctx context.Context, campaign *model.Campaign) error {
	query := `
		UPDATE campaigns SET
			name = $2,
			description = $3,
			status = $4,
			platform = $5,
			ad_account_id = $6,
			total_budget = $7,
			daily_budget = $8,
			start_date = $9,
			end_date = $10,
			updated_at = $11
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		campaign.ID,
		campaign.Name,
		campaign.Description,
		campaign.Status,
		campaign.Platform,
		campaign.AdAccountID,
		campaign.TotalBudget,
		campaign.DailyBudget,
		campaign.StartDate,
		campaign.EndDate,
		campaign.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update campaign: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrCampaignNotFound
	}

	return nil
}

// DeleteCampaign deletes a campaign
func (r *PGCampaignRepo) DeleteCampaign(ctx context.Context, id string) error {
	query := `DELETE FROM campaigns WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete campaign: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrCampaignNotFound
	}

	return nil
}

// ListCampaigns lists campaigns with filtering and pagination
func (r *PGCampaignRepo) ListCampaigns(ctx context.Context, filter *CampaignFilter, page, limit int) ([]*model.Campaign, int, error) {
	// Build query
	baseQuery := `FROM campaigns WHERE 1=1`
	var args []interface{}
	argNum := 1

	if filter != nil {
		if filter.Status != nil {
			baseQuery += fmt.Sprintf(" AND status = $%d", argNum)
			args = append(args, *filter.Status)
			argNum++
		}
		if filter.ClientID != nil {
			baseQuery += fmt.Sprintf(" AND client_id = $%d", argNum)
			args = append(args, *filter.ClientID)
			argNum++
		}
		if filter.AdSpecialistID != nil {
			baseQuery += fmt.Sprintf(" AND ad_specialist_id = $%d", argNum)
			args = append(args, *filter.AdSpecialistID)
			argNum++
		}
		if filter.BriefID != nil {
			baseQuery += fmt.Sprintf(" AND brief_id = $%d", argNum)
			args = append(args, *filter.BriefID)
			argNum++
		}
	}

	// Get total count
	countQuery := `SELECT COUNT(*) ` + baseQuery
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count campaigns: %w", err)
	}

	// Get campaigns with pagination
	offset := (page - 1) * limit
	selectQuery := `
		SELECT id, ad_specialist_id, client_id, brief_id, name, description,
			status, platform, ad_account_id, total_budget, daily_budget,
			start_date, end_date, created_at, updated_at
	` + baseQuery + ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", argNum) + ` OFFSET $` + fmt.Sprintf("%d", argNum+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []*model.Campaign
	for rows.Next() {
		var campaign model.Campaign
		var briefID *uuid.UUID
		var endDate *time.Time

		err := rows.Scan(
			&campaign.ID,
			&campaign.AdSpecialistID,
			&campaign.ClientID,
			&briefID,
			&campaign.Name,
			&campaign.Description,
			&campaign.Status,
			&campaign.Platform,
			&campaign.AdAccountID,
			&campaign.TotalBudget,
			&campaign.DailyBudget,
			&campaign.StartDate,
			&endDate,
			&campaign.CreatedAt,
			&campaign.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan campaign: %w", err)
		}

		campaign.BriefID = briefID
		campaign.EndDate = endDate
		campaigns = append(campaigns, &campaign)
	}

	return campaigns, total, nil
}

// CreateCampaignVideo creates a new campaign video
func (r *PGCampaignRepo) CreateCampaignVideo(ctx context.Context, cv *model.CampaignVideo) error {
	query := `
		INSERT INTO campaign_videos (id, campaign_id, video_id, status, added_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(ctx, query,
		cv.ID,
		cv.CampaignID,
		cv.VideoID,
		cv.Status,
		cv.AddedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create campaign video: %w", err)
	}

	return nil
}

// GetCampaignVideo gets a campaign video
func (r *PGCampaignRepo) GetCampaignVideo(ctx context.Context, campaignID, videoID string) (*model.CampaignVideo, error) {
	query := `
		SELECT id, campaign_id, video_id, status, added_at
		FROM campaign_videos
		WHERE campaign_id = $1 AND video_id = $2
	`

	var cv model.CampaignVideo
	err := r.pool.QueryRow(ctx, query, campaignID, videoID).Scan(
		&cv.ID,
		&cv.CampaignID,
		&cv.VideoID,
		&cv.Status,
		&cv.AddedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCampaignVideoNotFound
		}
		return nil, fmt.Errorf("failed to get campaign video: %w", err)
	}

	return &cv, nil
}

// DeleteCampaignVideo deletes a campaign video
func (r *PGCampaignRepo) DeleteCampaignVideo(ctx context.Context, campaignID, videoID string) error {
	query := `DELETE FROM campaign_videos WHERE campaign_id = $1 AND video_id = $2`

	result, err := r.pool.Exec(ctx, query, campaignID, videoID)
	if err != nil {
		return fmt.Errorf("failed to delete campaign video: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrCampaignVideoNotFound
	}

	return nil
}

// GetCampaignVideos gets all videos for a campaign
func (r *PGCampaignRepo) GetCampaignVideos(ctx context.Context, campaignID string) ([]*model.CampaignVideo, error) {
	query := `
		SELECT id, campaign_id, video_id, status, added_at
		FROM campaign_videos
		WHERE campaign_id = $1
		ORDER BY added_at DESC
	`

	rows, err := r.pool.Query(ctx, query, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign videos: %w", err)
	}
	defer rows.Close()

	var videos []*model.CampaignVideo
	for rows.Next() {
		var cv model.CampaignVideo
		err := rows.Scan(
			&cv.ID,
			&cv.CampaignID,
			&cv.VideoID,
			&cv.Status,
			&cv.AddedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan campaign video: %w", err)
		}
		videos = append(videos, &cv)
	}

	return videos, nil
}

// CreateAdAccount creates a new ad account
func (r *PGCampaignRepo) CreateAdAccount(ctx context.Context, aa *model.AdAccount) error {
	query := `
		INSERT INTO ad_accounts (id, user_id, platform, account_id, name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		aa.ID,
		aa.UserID,
		aa.Platform,
		aa.AccountID,
		aa.Name,
		aa.Status,
		aa.CreatedAt,
		aa.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create ad account: %w", err)
	}

	return nil
}

// GetAdAccountByID gets an ad account by ID
func (r *PGCampaignRepo) GetAdAccountByID(ctx context.Context, id string) (*model.AdAccount, error) {
	query := `
		SELECT id, user_id, platform, account_id, name, status, created_at, updated_at
		FROM ad_accounts
		WHERE id = $1
	`

	var aa model.AdAccount
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&aa.ID,
		&aa.UserID,
		&aa.Platform,
		&aa.AccountID,
		&aa.Name,
		&aa.Status,
		&aa.CreatedAt,
		&aa.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrAdAccountNotFound
		}
		return nil, fmt.Errorf("failed to get ad account: %w", err)
	}

	return &aa, nil
}

// GetAdAccountsByUserID gets ad accounts by user ID
func (r *PGCampaignRepo) GetAdAccountsByUserID(ctx context.Context, userID string, page, limit int) ([]*model.AdAccount, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM ad_accounts WHERE user_id = $1`
	var total int
	err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count ad accounts: %w", err)
	}

	// Get ad accounts with pagination
	offset := (page - 1) * limit
	selectQuery := `
		SELECT id, user_id, platform, account_id, name, status, created_at, updated_at
		FROM ad_accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, selectQuery, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list ad accounts: %w", err)
	}
	defer rows.Close()

	var accounts []*model.AdAccount
	for rows.Next() {
		var aa model.AdAccount
		err := rows.Scan(
			&aa.ID,
			&aa.UserID,
			&aa.Platform,
			&aa.AccountID,
			&aa.Name,
			&aa.Status,
			&aa.CreatedAt,
			&aa.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan ad account: %w", err)
		}
		accounts = append(accounts, &aa)
	}

	return accounts, total, nil
}

// UpdateAdAccount updates an ad account
func (r *PGCampaignRepo) UpdateAdAccount(ctx context.Context, aa *model.AdAccount) error {
	query := `
		UPDATE ad_accounts SET
			platform = $2,
			account_id = $3,
			name = $4,
			status = $5,
			updated_at = $6
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		aa.ID,
		aa.Platform,
		aa.AccountID,
		aa.Name,
		aa.Status,
		aa.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update ad account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrAdAccountNotFound
	}

	return nil
}

// CreateCampaignBudget creates a new campaign budget
func (r *PGCampaignRepo) CreateCampaignBudget(ctx context.Context, cb *model.CampaignBudget) error {
	query := `
		INSERT INTO campaign_budgets (id, campaign_id, amount, type, spent, remaining, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		cb.ID,
		cb.CampaignID,
		cb.Amount,
		cb.Type,
		cb.Spent,
		cb.Remaining,
		cb.CreatedAt,
		cb.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create campaign budget: %w", err)
	}

	return nil
}

// GetCampaignBudget gets a campaign budget
func (r *PGCampaignRepo) GetCampaignBudget(ctx context.Context, campaignID string) (*model.CampaignBudget, error) {
	query := `
		SELECT id, campaign_id, amount, type, spent, remaining, created_at, updated_at
		FROM campaign_budgets
		WHERE campaign_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var cb model.CampaignBudget
	err := r.pool.QueryRow(ctx, query, campaignID).Scan(
		&cb.ID,
		&cb.CampaignID,
		&cb.Amount,
		&cb.Type,
		&cb.Spent,
		&cb.Remaining,
		&cb.CreatedAt,
		&cb.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrBudgetNotFound
		}
		return nil, fmt.Errorf("failed to get campaign budget: %w", err)
	}

	return &cb, nil
}

// UpdateCampaignBudget updates a campaign budget
func (r *PGCampaignRepo) UpdateCampaignBudget(ctx context.Context, cb *model.CampaignBudget) error {
	query := `
		UPDATE campaign_budgets SET
			amount = $2,
			type = $3,
			spent = $4,
			remaining = $5,
			updated_at = $6
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		cb.ID,
		cb.Amount,
		cb.Type,
		cb.Spent,
		cb.Remaining,
		cb.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update campaign budget: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrBudgetNotFound
	}

	return nil
}

// Errors
var (
	ErrCampaignNotFound    = fmt.Errorf("campaign not found")
	ErrCampaignVideoNotFound = fmt.Errorf("campaign video not found")
	ErrAdAccountNotFound   = fmt.Errorf("ad account not found")
	ErrBudgetNotFound      = fmt.Errorf("budget not found")
)