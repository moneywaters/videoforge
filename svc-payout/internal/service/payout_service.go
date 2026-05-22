package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"github.com/videoforge/backend/pkg/natsclient"
	"github.com/videoforge/backend/svc-payout/internal/model"
	"github.com/videoforge/backend/svc-payout/internal/repository"
)

// PayoutService handles payout business logic
type PayoutService struct {
	repo     *repository.PayoutRepository
	nc       *natsclient.Client
	holdDays int
}

// NewPayoutService creates a new payout service
func NewPayoutService(repo *repository.PayoutRepository, nc *natsclient.Client, holdDays int) *PayoutService {
	return &PayoutService{
		repo:     repo,
		nc:       nc,
		holdDays: holdDays,
	}
}

// GetPayoutByID retrieves payout by ID
func (s *PayoutService) GetPayoutByID(ctx context.Context, id uuid.UUID) (*model.Payout, error) {
	return s.repo.GetPayoutByID(ctx, id)
}

// GetPayoutsForUser retrieves payouts for a user
func (s *PayoutService) GetPayoutsForUser(ctx context.Context, userID uuid.UUID) ([]model.Payout, error) {
	return s.repo.GetPayoutsByUserID(ctx, userID)
}

// GetAllPayouts retrieves all payouts (admin)
func (s *PayoutService) GetAllPayouts(ctx context.Context) ([]model.Payout, error) {
	return s.repo.GetAllPayouts(ctx)
}

// GetBalance retrieves user balance
func (s *PayoutService) GetBalance(ctx context.Context, userID uuid.UUID) (*model.BalanceResponse, error) {
	balance, err := s.repo.GetOrCreateBalance(ctx, userID, "USD")
	if err != nil {
		return nil, err
	}
	return &model.BalanceResponse{
		UserID:      balance.UserID,
		Available:  balance.Available,
		Pending:    balance.Pending,
		TotalEarned: balance.TotalEarned,
		Currency:   balance.Currency,
	}, nil
}

// CalculateEarnings calculates earnings for a user in a period
func (s *PayoutService) CalculateEarnings(ctx context.Context, userID uuid.UUID, periodStart, periodEnd time.Time) (*model.EarningsBreakdown, error) {
	// Get total verified sales in period
	totalVerifiedSales, err := s.repo.GetTotalVerifiedSalesByUserIDInPeriod(ctx, userID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	// Get payout rule
	rule, err := s.repo.GetDefaultPayoutRule(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate earnings based on rule
	var platformFee float64
	var userEarnings float64

	if totalVerifiedSales <= rule.ThresholdAmount {
		// First $500 (or threshold): no platform fee
		platformFee = 0
		userEarnings = totalVerifiedSales
	} else {
		// After threshold: platform takes 5% of user earnings
		platformFee = totalVerifiedSales * (rule.PlatformFeePercent / 100)
		userEarnings = totalVerifiedSales - platformFee
	}

	// User's actual share (proportional based on their contribution)
	// This is simplified - in reality would come from brief assignment
	userShareOfBounty := userEarnings

	holdUntil := time.Now().Add(time.Duration(s.holdDays) * 24 * time.Hour)

	return &model.EarningsBreakdown{
		UserID:             userID,
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		TotalVerifiedSales: totalVerifiedSales,
		UserShareOfBounty:  userShareOfBounty,
		PlatformFeePercent: rule.PlatformFeePercent,
		PlatformFee:       platformFee,
		UserEarnings:      userEarnings,
		HoldUntil:        &holdUntil,
	}, nil
}

// CreatePayout creates a new payout
func (s *PayoutService) CreatePayout(ctx context.Context, userID uuid.UUID, payoutType model.PayoutType, amount float64, currency string) (*model.Payout, error) {
	holdUntil := time.Now().Add(time.Duration(s.holdDays) * 24 * time.Hour)
	now := time.Now()

	payout := &model.Payout{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      payoutType,
		Amount:    amount,
		Currency:  currency,
		Status:   model.PayoutStatusPending,
		HoldUntil: &holdUntil,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreatePayout(ctx, payout); err != nil {
		return nil, err
	}

	// Update pending balance
	balance, err := s.repo.GetOrCreateBalance(ctx, userID, currency)
	if err != nil {
		return nil, err
	}

	balance.Pending += amount
	balance.TotalEarned += amount
	balance.UpdatedAt = time.Now()
	if err := s.repo.UpdateBalance(ctx, balance); err != nil {
		return nil, err
	}

	// Create transaction record
	txn := &model.Transaction{
		ID:           uuid.New(),
		PayoutID:     &payout.ID,
		Type:         model.TransactionTypeEarning,
		Amount:       amount,
		Currency:     currency,
		Description:  fmt.Sprintf("Payout earned: %s", payoutType),
		CreatedAt:    now,
	}
	if err := s.repo.CreateTransaction(ctx, txn); err != nil {
		return nil, err
	}

	return payout, nil
}

// ReleaseHolds releases pending payouts after hold period
func (s *PayoutService) ReleaseHolds(ctx context.Context) error {
	payouts, err := s.repo.GetPayoutsPendingRelease(ctx)
	if err != nil {
		return err
	}

	for _, payout := range payouts {
		// Update payout status to eligible
		if err := s.repo.UpdatePayoutStatus(ctx, payout.ID, model.PayoutStatusEligible, nil); err != nil {
			return err
		}

		// Move from pending to available in balance
		balance, err := s.repo.GetBalanceByUserID(ctx, payout.UserID)
		if err != nil {
			continue
		}

		balance.Available += payout.Amount
		balance.Pending -= payout.Amount
		balance.UpdatedAt = time.Now()
		if err := s.repo.UpdateBalance(ctx, balance); err != nil {
			continue
		}

		// Create hold release transaction
		txn := &model.Transaction{
			ID:           uuid.New(),
			PayoutID:     &payout.ID,
			Type:         model.TransactionTypeHoldRelease,
			Amount:       payout.Amount,
			Currency:     payout.Currency,
			Description:  "Hold released - funds available for payout",
			CreatedAt:    time.Now(),
		}
		s.repo.CreateTransaction(ctx, txn)
	}

	return nil
}

// CreateRuulBatch creates a new Ruul payout batch
func (s *PayoutService) CreateRuulBatch(ctx context.Context, userIDs []uuid.UUID, description string) (*model.RuulPayoutBatch, error) {
	batch := &model.RuulPayoutBatch{
		ID:         uuid.New(),
		BatchName:  description,
		Status:    model.RuulBatchStatusDraft,
		Currency: "USD",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateRuulBatch(ctx, batch); err != nil {
		return nil, err
	}

	// Create payout requests for each user
	for _, userID := range userIDs {
		balance, err := s.repo.GetBalanceByUserID(ctx, userID)
		if err != nil {
			continue
		}

		if balance.Available <= 0 {
			continue
		}

		req := &model.RuulPayoutRequest{
			ID:         uuid.New(),
			BatchID:    batch.ID,
			UserID:     userID,
			Amount:     balance.Available,
			Currency:   "USD",
			Status:     "pending",
			CreatedAt:  time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.repo.CreateRuulPayoutRequest(ctx, req); err != nil {
			continue
		}

		batch.TotalAmount += balance.Available
	}

	return batch, nil
}

// GetRuulBatch retrieves a batch
func (s *PayoutService) GetRuulBatch(ctx context.Context, id uuid.UUID) (*model.RuulPayoutBatch, error) {
	return s.repo.GetRuulBatchByID(ctx, id)
}

// GetAllRuulBatches retrieves all batches
func (s *PayoutService) GetAllRuulBatches(ctx context.Context) ([]model.RuulPayoutBatch, error) {
	return s.repo.GetAllRuulBatches(ctx)
}

// ProcessRuulBatch processes a batch via Ruul API
func (s *PayoutService) ProcessRuulBatch(ctx context.Context, batchID uuid.UUID) (*model.RuulPayoutBatch, error) {
	// Update status to processing
	if err := s.repo.UpdateRuulBatchStatus(ctx, batchID, model.RuulBatchStatusProcessing); err != nil {
		return nil, err
	}

	// TODO: Call Ruul.io API to process payout
	// This is a STUB - in production would call:
	// POST /payouts/bulk with batch details

	// Get requests for this batch
	requests, err := s.repo.GetRuulPayoutRequestsByBatchID(ctx, batchID)
	if err != nil {
		return nil, err
	}

	// For each request, mark as processing
	for _, req := range requests {
		// TODO: Call Ruul API per user
		// In production, this would create actual payment requests
		s.repo.UpdateRuulPayoutRequestStatus(ctx, req.ID, "processing", "")
	}

	// Update batch status
	if err := s.repo.UpdateRuulBatchStatus(ctx, batchID, model.RuulBatchStatusCompleted); err != nil {
		return nil, err
	}

	// Mark individual requests as completed
	for _, req := range requests {
		s.repo.UpdateRuulPayoutRequestStatus(ctx, req.ID, "completed", fmt.Sprintf("ruul-%s", req.ID))
	}

	return s.repo.GetRuulBatchByID(ctx, batchID)
}

// HandleSaleAttributed processes a sale.attributed NATS event
func (s *PayoutService) HandleSaleAttributed(ctx context.Context, event model.SaleAttributedEvent) error {
	// Determine payout type based on user type
	var payoutType model.PayoutType
	if event.UserType == "editor" {
		payoutType = model.PayoutTypeEditorFee
	} else if event.UserType == "specialist" {
		payoutType = model.PayoutTypeSpecialistFee
	} else {
		return fmt.Errorf("unknown user type: %s", event.UserType)
	}

	// Calculate earnings with tiered fee structure
	totalSales := event.TotalBounty
	userShare := totalSales * (event.UserShareOfBounty / event.TotalBounty)

	var platformFee float64
	var userEarnings float64

	if totalSales <= 500 {
		// First $500: no platform fee
		platformFee = 0
		userEarnings = userShare
	} else {
		// After $500: 5% platform fee on user earnings
		platformFee = userShare * 0.05
		userEarnings = userShare - platformFee
	}

	// Create payout with hold
	holdUntil := time.Now().Add(time.Duration(s.holdDays) * 24 * time.Hour)
	now := time.Now()

	payout := &model.Payout{
		ID:        uuid.New(),
		UserID:    event.UserID,
		Type:      payoutType,
		Amount:    userEarnings,
		Currency: event.Currency,
		Status:   model.PayoutStatusPending,
		HoldUntil: &holdUntil,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreatePayout(ctx, payout); err != nil {
		return err
	}

	// Create platform fee payout if applicable
	if platformFee > 0 {
		platformPayout := &model.Payout{
			ID:        uuid.New(),
			UserID:    event.UserID,
			Type:     model.PayoutTypePlatformFee,
			Amount:   platformFee,
			Currency: event.Currency,
			Status:   model.PayoutStatusPending,
			HoldUntil: &holdUntil,
			CreatedAt: now,
			UpdatedAt: now,
		}
		s.repo.CreatePayout(ctx, platformPayout)
	}

	// Update pending balance
	balance, err := s.repo.GetOrCreateBalance(ctx, event.UserID, event.Currency)
	if err != nil {
		return err
	}

	balance.Pending += userEarnings
	balance.TotalEarned += userEarnings
	balance.UpdatedAt = time.Now()
	if err := s.repo.UpdateBalance(ctx, balance); err != nil {
		return err
	}

	// Create transaction
	txn := &model.Transaction{
		ID:           uuid.New(),
		PayoutID:     &payout.ID,
		Type:         model.TransactionTypeEarning,
		Amount:       userEarnings,
		Currency:     event.Currency,
		Description:  fmt.Sprintf("Earning from sale %s", event.SaleID),
		CreatedAt:    now,
	}
	return s.repo.CreateTransaction(ctx, txn)
}

// SubscribeToEvents subscribes to NATS events
func (s *PayoutService) SubscribeToEvents(ctx context.Context) error {
	if s.nc == nil {
		return fmt.Errorf("NATS connection not available")
	}

	// Subscribe to sale.attributed events
	err := s.nc.Subscribe("sale.attributed", func(msg *nats.Msg) {
		var event model.SaleAttributedEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			return
		}
		if err := s.HandleSaleAttributed(ctx, event); err != nil {
			return
		}
	})
	return err
}

// HandleDodoWebhook handles DodoPayments webhook
func (s *PayoutService) HandleDodoWebhook(ctx context.Context, event model.DodoWebhookEvent) error {
	// TODO: Process DodoPayments webhook
	// This would typically update the brief status or emit an event
	return nil
}

// HandleRuulWebhook handles Ruul webhook
func (s *PayoutService) HandleRuulWebhook(ctx context.Context, event model.RuulWebhookEvent) error {
	// TODO: Process Ruul webhook to update payout request status
	return nil
}