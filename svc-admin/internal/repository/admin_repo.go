package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
	"svc-admin/internal/model"

	"github.com/videoforge/backend/pkg/errors"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found", 404)
	// ErrDisputeNotFound is returned when a dispute is not found
	ErrDisputeNotFound = errors.New("dispute not found", 404)
	// ErrModerationNotFound is returned when a moderation item is not found
	ErrModerationNotFound = errors.New("moderation item not found", 404)
)

// AdminRepository handles database operations for admin
type AdminRepository struct {
	db *pgxpool.Pool
}

// NewAdminRepository creates a new AdminRepository
func NewAdminRepository(db *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{db: db}
}

// AdminRepoInterface defines the interface for admin repository
type AdminRepoInterface interface {
	// User operations
	SearchUsers(ctx context.Context, req model.UserListRequest) ([]model.User, int, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error
	UpdateUserRoles(ctx context.Context, userID uuid.UUID, roles []string) error

	// Admin action operations
	CreateAdminAction(ctx context.Context, action *model.AdminAction) error
	ListAdminActions(ctx context.Context, req model.ActionsListRequest) ([]model.AdminAction, int, error)

	// Dispute operations
	ListDisputes(ctx context.Context, req model.DisputesListRequest) ([]model.Dispute, int, error)
	GetDisputeByID(ctx context.Context, id uuid.UUID) (*model.Dispute, error)
	UpdateDisputeStatus(ctx context.Context, disputeID uuid.UUID, status string, resolution string, resolvedBy uuid.UUID) error

	// Moderation queue operations
	ListModerationQueue(ctx context.Context, req model.ModerationQueueListRequest) ([]model.ModerationQueue, int, error)
	GetModerationQueueByID(ctx context.Context, id uuid.UUID) (*model.ModerationQueue, error)
	UpdateModerationQueueStatus(ctx context.Context, id uuid.UUID, status string, reviewedBy uuid.UUID, actionTaken string) error
}

// SearchUsers searches for users with filters
func (r *AdminRepository) SearchUsers(ctx context.Context, req model.UserListRequest) ([]model.User, int, error) {
	// Build query
	baseQuery := `
		SELECT id, email, first_name, last_name, role, status, created_at, updated_at, last_login_at
		FROM users
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if req.Role != "" {
		baseQuery += fmt.Sprintf(" AND role = $%d", argIdx)
		args = append(args, req.Role)
		argIdx++
	}

	if req.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, req.Status)
		argIdx++
	}

	if req.Email != "" {
		baseQuery += fmt.Sprintf(" AND email ILIKE $%d", argIdx)
		args = append(args, "%"+req.Email+"%")
		argIdx++
	}

	// Count query
	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") AS sub"
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Add pagination
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := r.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		var lastLoginAt pgtype.NullTime
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
			&lastLoginAt,
		); err != nil {
			return nil, 0, err
		}
		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}
		users = append(users, user)
	}

	return users, total, nil
}

// GetUserByID retrieves a user by ID
func (r *AdminRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, email, first_name, last_name, role, status, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`
	var user model.User
	var lastLoginAt pgtype.NullTime
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&lastLoginAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}
	return &user, nil
}

// UpdateUserStatus updates a user's status
func (r *AdminRepository) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	query := `
		UPDATE users
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	result, err := r.db.Exec(ctx, query, status, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// UpdateUserRoles updates a user's roles
func (r *AdminRepository) UpdateUserRoles(ctx context.Context, userID uuid.UUID, roles []string) error {
	// For now, just update the primary role
	role := "client"
	if len(roles) > 0 {
		role = roles[0]
	}

	query := `
		UPDATE users
		SET role = $1, updated_at = $2
		WHERE id = $3
	`
	result, err := r.db.Exec(ctx, query, role, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update user roles: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// CreateAdminAction creates a new admin action
func (r *AdminRepository) CreateAdminAction(ctx context.Context, action *model.AdminAction) error {
	query := `
		INSERT INTO admin_actions (id, admin_id, action_type, target_user_id, target_type, target_id, reason, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		action.ID,
		action.AdminID,
		action.ActionType,
		action.TargetUserID,
		action.TargetType,
		action.TargetID,
		action.Reason,
		action.Metadata,
		action.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create admin action: %w", err)
	}
	return nil
}

// ListAdminActions lists admin actions with filters
func (r *AdminRepository) ListAdminActions(ctx context.Context, req model.ActionsListRequest) ([]model.AdminAction, int, error) {
	baseQuery := `
		SELECT id, admin_id, action_type, target_user_id, target_type, target_id, reason, metadata, created_at
		FROM admin_actions
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if req.AdminID != "" {
		adminID, err := uuid.Parse(req.AdminID)
		if err == nil {
			baseQuery += fmt.Sprintf(" AND admin_id = $%d", argIdx)
			args = append(args, adminID)
			argIdx++
		}
	}

	if req.ActionType != "" {
		baseQuery += fmt.Sprintf(" AND action_type = $%d", argIdx)
		args = append(args, req.ActionType)
		argIdx++
	}

	// Count query
	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") AS sub"
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count admin actions: %w", err)
	}

	// Add pagination
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := r.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list admin actions: %w", err)
	}
	defer rows.Close()

	var actions []model.AdminAction
	for rows.Next() {
		var action model.AdminAction
		var targetUserID, targetID pgtype.UUID
		if err := rows.Scan(
			&action.ID,
			&action.AdminID,
			&action.ActionType,
			&targetUserID,
			&action.TargetType,
			&targetID,
			&action.Reason,
			&action.Metadata,
			&action.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		if targetUserID.Valid {
			action.TargetUserID = &targetUserID.UUID
		}
		if targetID.Valid {
			action.TargetID = &targetID.UUID
		}
		actions = append(actions, action)
	}

	return actions, total, nil
}

// ListDisputes lists disputes with filters
func (r *AdminRepository) ListDisputes(ctx context.Context, req model.DisputesListRequest) ([]model.Dispute, int, error) {
	baseQuery := `
		SELECT id, reporter_id, respondent_id, target_type, target_id, status, description, evidence, resolution, resolved_by, resolved_at, created_at, updated_at
		FROM disputes
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if req.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, req.Status)
		argIdx++
	}

	// Count query
	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") AS sub"
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count disputes: %w", err)
	}

	// Add pagination
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := r.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list disputes: %w", err)
	}
	defer rows.Close()

	var disputes []model.Dispute
	for rows.Next() {
		var dispute model.Dispute
		var resolution, resolvedBy pgtype.UUID
		var resolvedAt pgtype.NullTime
		if err := rows.Scan(
			&dispute.ID,
			&dispute.ReporterID,
			&dispute.RespondentID,
			&dispute.TargetType,
			&dispute.TargetID,
			&dispute.Status,
			&dispute.Description,
			&dispute.Evidence,
			&resolution,
			&resolvedBy,
			&resolvedAt,
			&dispute.CreatedAt,
			&dispute.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		if resolution.Valid {
			res := resolution.UUID.String()
			dispute.Resolution = &res
		}
		if resolvedBy.Valid {
			dispute.ResolvedBy = &resolvedBy.UUID
		}
		if resolvedAt.Valid {
			dispute.ResolvedAt = &resolvedAt.Time
		}
		disputes = append(disputes, dispute)
	}

	return disputes, total, nil
}

// GetDisputeByID retrieves a dispute by ID
func (r *AdminRepository) GetDisputeByID(ctx context.Context, id uuid.UUID) (*model.Dispute, error) {
	query := `
		SELECT id, reporter_id, respondent_id, target_type, target_id, status, description, evidence, resolution, resolved_by, resolved_at, created_at, updated_at
		FROM disputes
		WHERE id = $1
	`
	var dispute model.Dispute
	var resolution, resolvedBy pgtype.UUID
	var resolvedAt pgtype.NullTime
	err := r.db.QueryRow(ctx, query, id).Scan(
		&dispute.ID,
		&dispute.ReporterID,
		&dispute.RespondentID,
		&dispute.TargetType,
		&dispute.TargetID,
		&dispute.Status,
		&dispute.Description,
		&dispute.Evidence,
		&resolution,
		&resolvedBy,
		&resolvedAt,
		&dispute.CreatedAt,
		&dispute.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDisputeNotFound
		}
		return nil, fmt.Errorf("failed to get dispute: %w", err)
	}
	if resolution.Valid {
		res := resolution.UUID.String()
		dispute.Resolution = &res
	}
	if resolvedBy.Valid {
		dispute.ResolvedBy = &resolvedBy.UUID
	}
	if resolvedAt.Valid {
		dispute.ResolvedAt = &resolvedAt.Time
	}
	return &dispute, nil
}

// UpdateDisputeStatus updates a dispute's status
func (r *AdminRepository) UpdateDisputeStatus(ctx context.Context, disputeID uuid.UUID, status string, resolution string, resolvedBy uuid.UUID) error {
	query := `
		UPDATE disputes
		SET status = $1, resolution = $2, resolved_by = $3, resolved_at = $4, updated_at = $5
		WHERE id = $6
	`
	result, err := r.db.Exec(ctx, query, status, resolution, resolvedBy, time.Now(), time.Now(), disputeID)
	if err != nil {
		return fmt.Errorf("failed to update dispute status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrDisputeNotFound
	}
	return nil
}

// ListModerationQueue lists moderation queue with filters
func (r *AdminRepository) ListModerationQueue(ctx context.Context, req model.ModerationQueueListRequest) ([]model.ModerationQueue, int, error) {
	baseQuery := `
		SELECT id, target_type, target_id, flag_reason, status, reviewed_by, reviewed_at, action_taken, created_at
		FROM moderation_queue
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if req.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, req.Status)
		argIdx++
	}

	// Count query
	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") AS sub"
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count moderation queue: %w", err)
	}

	// Add pagination
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := r.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list moderation queue: %w", err)
	}
	defer rows.Close()

	var items []model.ModerationQueue
	for rows.Next() {
		var item model.ModerationQueue
		var reviewedBy pgtype.UUID
		var reviewedAt pgtype.NullTime
		var actionTaken pgtype.Text
		if err := rows.Scan(
			&item.ID,
			&item.TargetType,
			&item.TargetID,
			&item.FlagReason,
			&item.Status,
			&reviewedBy,
			&reviewedAt,
			&actionTaken,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		if reviewedBy.Valid {
			item.ReviewedBy = &reviewedBy.UUID
		}
		if reviewedAt.Valid {
			item.ReviewedAt = &reviewedAt.Time
		}
		if actionTaken.Valid {
			item.ActionTaken = &actionTaken.String
		}
		items = append(items, item)
	}

	return items, total, nil
}

// GetModerationQueueByID retrieves a moderation queue item by ID
func (r *AdminRepository) GetModerationQueueByID(ctx context.Context, id uuid.UUID) (*model.ModerationQueue, error) {
	query := `
		SELECT id, target_type, target_id, flag_reason, status, reviewed_by, reviewed_at, action_taken, created_at
		FROM moderation_queue
		WHERE id = $1
	`
	var item model.ModerationQueue
	var reviewedBy pgtype.UUID
	var reviewedAt pgtype.NullTime
	var actionTaken pgtype.Text
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.TargetType,
		&item.TargetID,
		&item.FlagReason,
		&item.Status,
		&reviewedBy,
		&reviewedAt,
		&actionTaken,
		&item.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrModerationNotFound
		}
		return nil, fmt.Errorf("failed to get moderation queue item: %w", err)
	}
	if reviewedBy.Valid {
		item.ReviewedBy = &reviewedBy.UUID
	}
	if reviewedAt.Valid {
		item.ReviewedAt = &reviewedAt.Time
	}
	if actionTaken.Valid {
		item.ActionTaken = &actionTaken.String
	}
	return &item, nil
}

// UpdateModerationQueueStatus updates a moderation queue item's status
func (r *AdminRepository) UpdateModerationQueueStatus(ctx context.Context, id uuid.UUID, status string, reviewedBy uuid.UUID, actionTaken string) error {
	query := `
		UPDATE moderation_queue
		SET status = $1, reviewed_by = $2, reviewed_at = $3, action_taken = $4
		WHERE id = $5
	`
	result, err := r.db.Exec(ctx, query, status, reviewedBy, time.Now(), actionTaken, id)
	if err != nil {
		return fmt.Errorf("failed to update moderation queue status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrModerationNotFound
	}
	return nil
}