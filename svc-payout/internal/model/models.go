package model

import (
	"time"

	"github.com/google/uuid"
)

// PayoutType represents the type of payout
type PayoutType string

const (
	PayoutTypeEditorFee     PayoutType = "editor_fee"
	PayoutTypeSpecialistFee PayoutType = "specialist_fee"
	PayoutTypePlatformFee   PayoutType = "platform_fee"
)

// PayoutStatus represents the status of a payout
type PayoutStatus string

const (
	PayoutStatusPending PayoutStatus = "pending"
	PayoutStatusEligible PayoutStatus = "eligible"
	PayoutStatusPaid  PayoutStatus = "paid"
	PayoutStatusFailed PayoutStatus = "failed"
)

// Payout represents a payout record
type Payout struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Type        PayoutType `json:"type"`
	Amount      float64    `json:"amount"`
	Currency    string    `json:"currency"`
	Status      PayoutStatus `json:"status"`
	HoldUntil   *time.Time `json:"hold_until,omitempty"`
	PaidAt     *time.Time `json:"paid_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PayoutRule represents platform fee rules
type PayoutRule struct {
	ID                uuid.UUID `json:"id"`
	Name             string   `json:"name"`
	ThresholdAmount   float64  `json:"threshold_amount"`
	PlatformFeePercent float64 `json:"platform_fee_percent"`
	Description     string   `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

// TransactionType represents transaction types
type TransactionType string

const (
	TransactionTypeEarning     TransactionType = "earning"
	TransactionTypeHoldRelease TransactionType = "hold_release"
	TransactionTypeFee         TransactionType = "fee"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID          uuid.UUID       `json:"id"`
	PayoutID    *uuid.UUID    `json:"payout_id,omitempty"`
	Type        TransactionType `json:"type"`
	Amount     float64      `json:"amount"`
	Currency   string      `json:"currency"`
	Description string     `json:"description"`
	CreatedAt  time.Time   `json:"created_at"`
}

// Balance represents user balance
type Balance struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	Available     float64   `json:"available"`
	Pending       float64   `json:"pending"`
	TotalEarned   float64   `json:"total_earned"`
	Currency      string    `json:"currency"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RuulBatchStatus represents Ruul batch status
type RuulBatchStatus string

const (
	RuulBatchStatusDraft      RuulBatchStatus = "draft"
	RuulBatchStatusProcessing RuulBatchStatus = "processing"
	RuulBatchStatusCompleted RuulBatchStatus = "completed"
	RuulBatchStatusFailed    RuulBatchStatus = "failed"
)

// RuulPayoutBatch represents a Ruul payout batch
type RuulPayoutBatch struct {
	ID          uuid.UUID        `json:"id"`
	BatchName  string         `json:"batch_name"`
	Status    RuulBatchStatus `json:"status"`
	TotalAmount float64     `json:"total_amount"`
	Currency   string       `json:"currency"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// RuulPayoutRequest represents a payout request in a batch
type RuulPayoutRequest struct {
	ID              uuid.UUID  `json:"id"`
	BatchID         uuid.UUID `json:"batch_id"`
	UserID          uuid.UUID `json:"user_id"`
	Amount          float64  `json:"amount"`
	Currency       string   `json:"currency"`
	Status         string   `json:"status"` // pending, completed, failed
	RuulReferenceID string   `json:"ruul_reference_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SaleAttributedEvent represents a sale.attributed NATS event
type SaleAttributedEvent struct {
	SaleID           uuid.UUID `json:"sale_id"`
	UserID           uuid.UUID `json:"user_id"`
	UserType         string   `json:"user_type"` // editor, specialist
	Amount          float64  `json:"amount"`
	Currency        string   `json:"currency"`
	BriefID         uuid.UUID `json:"brief_id"`
	VerifiedAt      time.Time `json:"verified_at"`
	TotalBounty    float64  `json:"total_bounty"`
	UserShareOfBounty float64 `json:"user_share_of_bounty"`
}

// EarningsBreakdown represents earnings calculation breakdown
type EarningsBreakdown struct {
	UserID           uuid.UUID `json:"user_id"`
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	TotalVerifiedSales float64 `json:"total_verified_sales"`
	UserShareOfBounty float64 `json:"user_share_of_bounty"`
	PlatformFeePercent float64 `json:"platform_fee_percent"`
	PlatformFee    float64 `json:"platform_fee"`
	UserEarnings   float64 `json:"user_earnings"`
	HoldUntil     *time.Time `json:"hold_until,omitempty"`
}

// BalanceResponse represents balance API response
type BalanceResponse struct {
	UserID     uuid.UUID `json:"user_id"`
	Available float64  `json:"available"`
	Pending   float64  `json:"pending"`
	TotalEarned float64 `json:"total_earned"`
	Currency  string   `json:"currency"`
}

// CreateBatchRequest represents payout batch creation request
type CreateBatchRequest struct {
	UserIDs      []uuid.UUID `json:"user_ids"`
	Description string    `json:"description"`
}

// CalculateEarningsRequest represents earnings calculation request
type CalculateEarningsRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd  time.Time `json:"period_end"`
}

// DodoWebhookEvent represents DodoPayments webhook event
type DodoWebhookEvent struct {
	EventType   string `json:"event_type"`
	ExternalID string `json:"external_id"`
	Amount    float64 `json:"amount"`
	Currency  string `json:"currency"`
	Status   string `json:"status"`
}

// RuulWebhookEvent represents Ruul webhook event
type RuulWebhookEvent struct {
	EventType       string `json:"event_type"`
	ReferenceID    string `json:"reference_id"`
	Status        string `json:"status"`
	Amount        float64 `json:"amount"`
	Currency      string `json:"currency"`
}