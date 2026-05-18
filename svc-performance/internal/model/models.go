package model

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// VideoSales represents sales data for a video
type VideoSales struct {
	ID           pgtype.UUID         `json:"id"`
	VideoID      pgtype.UUID         `json:"video_id"`
	CampaignID  pgtype.UUID         `json:"campaign_id"`
	TotalOrders  int                `json:"total_orders"`
	TotalRevenue pgtype.Numeric     `json:"total_revenue"`
	Currency     string             `json:"currency"`
	FirstSaleAt  pgtype.Timestamptz `json:"first_sale_at"`
	LastSaleAt   pgtype.Timestamptz `json:"last_sale_at"`
	UpdatedAt    pgtype.Timestamptz `json:"updated_at"`
}

// EditorSales represents sales data for an editor
type EditorSales struct {
	ID           pgtype.UUID     `json:"id"`
	EditorID    pgtype.UUID    `json:"editor_id"`
	TotalVideos int           `json:"total_videos"`
	TotalOrders int           `json:"total_orders"`
	TotalRevenue pgtype.Numeric `json:"total_revenue"`
	Currency    string        `json:"currency"`
	UpdatedAt   pgtype.Timestamptz `json:"updated_at"`
}

// SpecialistSales represents sales data for an ad specialist
type SpecialistSales struct {
	ID              pgtype.UUID     `json:"id"`
	SpecialistID    pgtype.UUID    `json:"specialist_id"`
	TotalCampaigns  int           `json:"total_campaigns"`
	TotalOrders     int           `json:"total_orders"`
	TotalRevenue    pgtype.Numeric `json:"total_revenue"`
	Currency        string        `json:"currency"`
	UpdatedAt       pgtype.Timestamptz `json:"updated_at"`
}

// CampaignSales represents sales data for a campaign
type CampaignSales struct {
	ID           pgtype.UUID         `json:"id"`
	CampaignID  pgtype.UUID         `json:"campaign_id"`
	TotalOrders int                `json:"total_orders"`
	TotalRevenue pgtype.Numeric    `json:"total_revenue"`
	Currency    string             `json:"currency"`
	StartDate   pgtype.Date        `json:"start_date"`
	EndDate     pgtype.Date        `json:"end_date"`
	UpdatedAt   pgtype.Timestamptz `json:"updated_at"`
}

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	ID           pgtype.UUID     `json:"id"`
	BriefID      pgtype.UUID    `json:"brief_id"`
	EntityType   string         `json:"entity_type"`
	EntityID     pgtype.UUID    `json:"entity_id"`
	Rank         int            `json:"rank"`
	TotalRevenue pgtype.Numeric  `json:"total_revenue"`
	TotalOrders  int            `json:"total_orders"`
	UpdatedAt   pgtype.Timestamptz `json:"updated_at"`
}

// DailyMetric represents daily sales metrics
type DailyMetric struct {
	ID          pgtype.UUID    `json:"id"`
	Date        pgtype.Date   `json:"date"`
	VideoID     pgtype.UUID   `json:"video_id"`
	CampaignID pgtype.UUID   `json:"campaign_id"`
	Orders      int          `json:"orders"`
	Revenue     pgtype.Numeric `json:"revenue"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
}

// Anomaly represents a flagged anomaly (placeholder)
type Anomaly struct {
	ID          string    `json:"id"`
	Type       string    `json:"type"`
	EntityID   string    `json:"entity_id"`
	EntityType string    `json:"entity_type"`
	Message    string    `json:"message"`
	DetectedAt time.Time `json:"detected_at"`
}

// SaleAttributedEvent represents a sale.attributed NATS event
type SaleAttributedEvent struct {
	SaleID        string    `json:"sale_id"`
	VideoID       string    `json:"video_id"`
	CampaignID   string    `json:"campaign_id"`
	EditorID     string    `json:"editor_id"`
	SpecialistID string    `json:"specialist_id"`
	BriefID      string    `json:"brief_id"`
	OrderID      string    `json:"order_id"`
	Amount       float64   `json:"amount"`
	Currency     string    `json:"currency"`
	Timestamp   time.Time `json:"timestamp"`
}

// AnalyticsQuery represents query parameters for analytics
type AnalyticsQuery struct {
	EntityType  string    `query:"entity_type"`
	EntityID   string    `query:"entity_id"`
	StartDate  string    `query:"start_date"`
	EndDate    string    `query:"end_date"`
	Granularity string   `query:"granularity"` // daily, weekly, monthly
}

// LeaderboardQuery represents query parameters for leaderboard
type LeaderboardQuery struct {
	EntityType string `query:"entity_type"`
}