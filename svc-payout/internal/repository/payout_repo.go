package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/videoforge/backend/svc-payout/internal/model"
)

// PayoutRepository handles payout database operations
type PayoutRepository struct {
	db *pgxpool.Pool
}

// NewPayoutRepository creates a new payout repository
func NewPayoutRepository(db *pgxpool.Pool) *PayoutRepository {
	return &PayoutRepository{db: db}
}

// CreatePayout creates a new payout record
func (r *PayoutRepository) CreatePayout(ctx context.Context, p *model.Payout) error {
	query := `
		INSERT INTO payouts (id, user_id, type, amount, currency, status, hold_until, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query, p.ID, p.UserID, p.Type, p.Amount, p.Currency, p.Status, p.HoldUntil, p.CreatedAt, p.UpdatedAt)
	return err
}

// GetPayoutByID retrieves a payout by ID
func (r *PayoutRepository) GetPayoutByID(ctx context.Context, id uuid.UUID) (*model.Payout, error) {
	query := `
		SELECT id, user_id, type, amount, currency, status, hold_until, paid_at, created_at, updated_at
		FROM payouts WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, id)
	return r.scanPayout(row)
}

// GetPayoutsByUserID retrieves all payouts for a user
func (r *PayoutRepository) GetPayoutsByUserID(ctx context.Context, userID uuid.UUID) ([]model.Payout, error) {
	query := `
		SELECT id, user_id, type, amount, currency, status, hold_until, paid_at, created_at, updated_at
		FROM payouts WHERE user_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payouts []model.Payout
	for rows.Next() {
		p, err := r.scanPayout(rows)
		if err != nil {
			return nil, err
		}
		payouts = append(payouts, *p)
	}
	return payouts, nil
}

// GetAllPayouts retrieves all payouts (admin)
func (r *PayoutRepository) GetAllPayouts(ctx context.Context) ([]model.Payout, error) {
	query := `
		SELECT id, user_id, type, amount, currency, status, hold_until, paid_at, created_at, updated_at
		FROM payouts ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payouts []model.Payout
	for rows.Next() {
		p, err := r.scanPayout(rows)
		if err != nil {
			return nil, err
		}
		payouts = append(payouts, *p)
	}
	return payouts, nil
}

// UpdatePayoutStatus updates payout status
func (r *PayoutRepository) UpdatePayoutStatus(ctx context.Context, id uuid.UUID, status model.PayoutStatus, paidAt *time.Time) error {
	query := `
		UPDATE payouts SET status = $2, paid_at = $3, updated_at = $4 WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id, status, paidAt, time.Now())
	return err
}

// GetPayoutsPendingRelease retrieves payouts ready for release
func (r *PayoutRepository) GetPayoutsPendingRelease(ctx context.Context) ([]model.Payout, error) {
	query := `
		SELECT id, user_id, type, amount, currency, status, hold_until, paid_at, created_at, updated_at
		FROM payouts WHERE status = 'pending' AND hold_until <= $1
	`
	rows, err := r.db.Query(ctx, query, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payouts []model.Payout
	for rows.Next() {
		p, err := r.scanPayout(rows)
		if err != nil {
			return nil, err
		}
		payouts = append(payouts, *p)
	}
	return payouts, nil
}

// CreateTransaction creates a new transaction record
func (r *PayoutRepository) CreateTransaction(ctx context.Context, t *model.Transaction) error {
	query := `
		INSERT INTO transactions (id, payout_id, type, amount, currency, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, t.ID, t.PayoutID, t.Type, t.Amount, t.Currency, t.Description, t.CreatedAt)
	return err
}

// CreateBalance creates a new balance record
func (r *PayoutRepository) CreateBalance(ctx context.Context, b *model.Balance) error {
	query := `
		INSERT INTO balances (id, user_id, available, pending, total_earned, currency, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, b.ID, b.UserID, b.Available, b.Pending, b.TotalEarned, b.Currency, b.UpdatedAt)
	return err
}

// GetBalanceByUserID retrieves balance for a user
func (r *PayoutRepository) GetBalanceByUserID(ctx context.Context, userID uuid.UUID) (*model.Balance, error) {
	query := `
		SELECT id, user_id, available, pending, total_earned, currency, updated_at
		FROM balances WHERE user_id = $1
	`
	row := r.db.QueryRow(ctx, query, userID)
	return r.scanBalance(row)
}

// UpdateBalance updates a balance record
func (r *PayoutRepository) UpdateBalance(ctx context.Context, b *model.Balance) error {
	query := `
		UPDATE balances SET available = $2, pending = $3, total_earned = $4, updated_at = $5 WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query, b.ID, b.Available, b.Pending, b.TotalEarned, b.UpdatedAt, b.ID)
	return err
}

// GetOrCreateBalance gets or creates balance for user
func (r *PayoutRepository) GetOrCreateBalance(ctx context.Context, userID uuid.UUID, currency string) (*model.Balance, error) {
	b, err := r.GetBalanceByUserID(ctx, userID)
	if err == nil {
		return b, nil
	}
	if err != pgx.ErrNoRows {
		return nil, err
	}

	// Create new balance
	newBalance := &model.Balance{
		ID:          uuid.New(),
		UserID:      userID,
		Available:   0,
		Pending:     0,
		TotalEarned: 0,
		Currency:    currency,
		UpdatedAt:  time.Now(),
	}
	if err := r.CreateBalance(ctx, newBalance); err != nil {
		return nil, err
	}
	return newBalance, nil
}

// GetPayoutRuleByName retrieves payout rule by name
func (r *PayoutRepository) GetPayoutRuleByName(ctx context.Context, name string) (*model.PayoutRule, error) {
	query := `
		SELECT id, name, threshold_amount, platform_fee_percent, description, created_at
		FROM payout_rules WHERE name = $1
	`
	row := r.db.QueryRow(ctx, query, name)
	return r.scanPayoutRule(row)
}

// GetDefaultPayoutRule retrieves default payout rule
func (r *PayoutRepository) GetDefaultPayoutRule(ctx context.Context) (*model.PayoutRule, error) {
	return r.GetPayoutRuleByName(ctx, "default")
}

// CreateRuulBatch creates a new Ruul payout batch
func (r *PayoutRepository) CreateRuulBatch(ctx context.Context, batch *model.RuulPayoutBatch) error {
	query := `
		INSERT INTO ruul_payout_batches (id, batch_name, status, total_amount, currency, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, batch.ID, batch.BatchName, batch.Status, batch.TotalAmount, batch.Currency, batch.CreatedAt, batch.UpdatedAt)
	return err
}

// GetRuulBatchByID retrieves a batch by ID
func (r *PayoutRepository) GetRuulBatchByID(ctx context.Context, id uuid.UUID) (*model.RuulPayoutBatch, error) {
	query := `
		SELECT id, batch_name, status, total_amount, currency, created_at, updated_at
		FROM ruul_payout_batches WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, id)
	return r.scanRuulBatch(row)
}

// GetAllRuulBatches retrieves all batches
func (r *PayoutRepository) GetAllRuulBatches(ctx context.Context) ([]model.RuulPayoutBatch, error) {
	query := `
		SELECT id, batch_name, status, total_amount, currency, created_at, updated_at
		FROM ruul_payout_batches ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var batches []model.RuulPayoutBatch
	for rows.Next() {
		b, err := r.scanRuulBatch(rows)
		if err != nil {
			return nil, err
		}
		batches = append(batches, *b)
	}
	return batches, nil
}

// UpdateRuulBatchStatus updates batch status
func (r *PayoutRepository) UpdateRuulBatchStatus(ctx context.Context, id uuid.UUID, status model.RuulBatchStatus) error {
	query := `
		UPDATE ruul_payout_batches SET status = $2, updated_at = $3 WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id, status, time.Now())
	return err
}

// CreateRuulPayoutRequest creates a payout request
func (r *PayoutRepository) CreateRuulPayoutRequest(ctx context.Context, req *model.RuulPayoutRequest) error {
	query := `
		INSERT INTO ruul_payout_requests (id, batch_id, user_id, amount, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query, req.ID, req.BatchID, req.UserID, req.Amount, req.Currency, req.Status, req.CreatedAt, req.UpdatedAt)
	return err
}

// GetRuulPayoutRequestsByBatchID retrieves requests for a batch
func (r *PayoutRepository) GetRuulPayoutRequestsByBatchID(ctx context.Context, batchID uuid.UUID) ([]model.RuulPayoutRequest, error) {
	query := `
		SELECT id, batch_id, user_id, amount, currency, status, ruul_reference_id, created_at, updated_at
		FROM ruul_payout_requests WHERE batch_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, batchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []model.RuulPayoutRequest
	for rows.Next() {
		req, err := r.scanRuulPayoutRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, *req)
	}
	return requests, nil
}

// UpdateRuulPayoutRequestStatus updates payout request status
func (r *PayoutRepository) UpdateRuulPayoutRequestStatus(ctx context.Context, id uuid.UUID, status string, ruulRefID string) error {
	query := `
		UPDATE ruul_payout_requests SET status = $2, ruul_reference_id = $3, updated_at = $4 WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id, status, ruulRefID, time.Now())
	return err
}

// scanPayout scans a payout row
func (r *PayoutRepository) scanPayout(row pgx.Row) (*model.Payout, error) {
	var p model.Payout
	var holdUntil, paidAt *time.Time
	err := row.Scan(&p.ID, &p.UserID, &p.Type, &p.Amount, &p.Currency, &p.Status, &holdUntil, &paidAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.HoldUntil = holdUntil
	p.PaidAt = paidAt
	return &p, nil
}

// scanPayoutRule scans a payout rule row
func (r *PayoutRepository) scanPayoutRule(row pgx.Row) (*model.PayoutRule, error) {
	var pr model.PayoutRule
	err := row.Scan(&pr.ID, &pr.Name, &pr.ThresholdAmount, &pr.PlatformFeePercent, &pr.Description, &pr.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

// scanBalance scans a balance row
func (r *PayoutRepository) scanBalance(row pgx.Row) (*model.Balance, error) {
	var b model.Balance
	err := row.Scan(&b.ID, &b.UserID, &b.Available, &b.Pending, &b.TotalEarned, &b.Currency, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// scanRuulBatch scans a Ruul batch row
func (r *PayoutRepository) scanRuulBatch(row pgx.Row) (*model.RuulPayoutBatch, error) {
	var b model.RuulPayoutBatch
	err := row.Scan(&b.ID, &b.BatchName, &b.Status, &b.TotalAmount, &b.Currency, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// scanRuulPayoutRequest scans a payout request row
func (r *PayoutRepository) scanRuulPayoutRequest(row pgx.Row) (*model.RuulPayoutRequest, error) {
	var req model.RuulPayoutRequest
	var ruulRefID *string
	err := row.Scan(&req.ID, &req.BatchID, &req.UserID, &req.Amount, &req.Currency, &req.Status, &ruulRefID, &req.CreatedAt, &req.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if ruulRefID != nil {
		req.RuulReferenceID = *ruulRefID
	}
	return &req, nil
}

// GetEarningsByUserIDInPeriod retrieves earnings for user in period
func (r *PayoutRepository) GetEarningsByUserIDInPeriod(ctx context.Context, userID uuid.UUID, periodStart, periodEnd time.Time) ([]model.Payout, error) {
	query := `
		SELECT id, user_id, type, amount, currency, status, hold_until, paid_at, created_at, updated_at
		FROM payouts 
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3 AND type IN ('editor_fee', 'specialist_fee')
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payouts []model.Payout
	for rows.Next() {
		p, err := r.scanPayout(rows)
		if err != nil {
			return nil, err
		}
		payouts = append(payouts, *p)
	}
	return payouts, nil
}

// GetTotalVerifiedSalesByUserIDInPeriod retrieves total verified sales for user in period
func (r *PayoutRepository) GetTotalVerifiedSalesByUserIDInPeriod(ctx context.Context, userID uuid.UUID, periodStart, periodEnd time.Time) (float64, error) {
	// This would typically come from the Performance service
	// For now, sum up the earnings as a proxy
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM payouts
		WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3 AND type IN ('editor_fee', 'specialist_fee')
	`
	var total float64
	err := r.db.QueryRow(ctx, query, userID, periodStart, periodEnd).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// EnsurePayoutRuleExists ensures default payout rule exists
func (r *PayoutRepository) EnsurePayoutRuleExists(ctx context.Context) error {
	rule := &model.PayoutRule{
		ID:                uuid.New(),
		Name:              "default",
		ThresholdAmount:   500.0,
		PlatformFeePercent: 5.0,
		Description:      "Default rule: $0 platform fee for first $500, then 5%",
		CreatedAt:         time.Now(),
	}
	query := `
		INSERT INTO payout_rules (id, name, threshold_amount, platform_fee_percent, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (name) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, rule.ID, rule.Name, rule.ThresholdAmount, rule.PlatformFeePercent, rule.Description, rule.CreatedAt)
	return err
}

// InitSchema initializes the database schema
func (r *PayoutRepository) InitSchema(ctx context.Context) error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS payouts (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL,
			type VARCHAR(50) NOT NULL,
			amount DECIMAL(18,2) NOT NULL,
			currency VARCHAR(10) NOT NULL DEFAULT 'USD',
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			hold_until TIMESTAMP,
			paid_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS payout_rules (
			id UUID PRIMARY KEY,
			name VARCHAR(100) UNIQUE NOT NULL,
			threshold_amount DECIMAL(18,2) NOT NULL,
			platform_fee_percent DECIMAL(5,2) NOT NULL,
			description TEXT,
			created_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id UUID PRIMARY KEY,
			payout_id UUID,
			type VARCHAR(50) NOT NULL,
			amount DECIMAL(18,2) NOT NULL,
			currency VARCHAR(10) NOT NULL DEFAULT 'USD',
			description TEXT,
			created_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS balances (
			id UUID PRIMARY KEY,
			user_id UUID UNIQUE NOT NULL,
			available DECIMAL(18,2) NOT NULL DEFAULT 0,
			pending DECIMAL(18,2) NOT NULL DEFAULT 0,
			total_earned DECIMAL(18,2) NOT NULL DEFAULT 0,
			currency VARCHAR(10) NOT NULL DEFAULT 'USD',
			updated_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS ruul_payout_batches (
			id UUID PRIMARY KEY,
			batch_name VARCHAR(255) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'draft',
			total_amount DECIMAL(18,2) NOT NULL DEFAULT 0,
			currency VARCHAR(10) NOT NULL DEFAULT 'USD',
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS ruul_payout_requests (
			id UUID PRIMARY KEY,
			batch_id UUID NOT NULL,
			user_id UUID NOT NULL,
			amount DECIMAL(18,2) NOT NULL,
			currency VARCHAR(10) NOT NULL DEFAULT 'USD',
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			ruul_reference_id VARCHAR(255),
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_payouts_user_id ON payouts(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_payouts_status ON payouts(status)`,
		`CREATE INDEX IF NOT EXISTS idx_payouts_hold_until ON payouts(hold_until)`,
		`CREATE INDEX IF NOT EXISTS idx_balances_user_id ON balances(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_ruul_payout_requests_batch_id ON ruul_payout_requests(batch_id)`,
	}

	for _, schema := range schemas {
		if _, err := r.db.Exec(ctx, schema); err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
	}

	return r.EnsurePayoutRuleExists(ctx)
}