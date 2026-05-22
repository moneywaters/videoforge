package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/videoforge/backend/svc-admin/internal/model"
	"github.com/videoforge/backend/svc-admin/internal/repository"

	"github.com/videoforge/backend/pkg/errors"
	"github.com/videoforge/backend/pkg/middleware"
	"github.com/videoforge/backend/pkg/natsclient"
)

var (
	// ErrUnauthorized is returned when user is not authorized
	ErrUnauthorized = errors.Forbidden("admin access required")
	// ErrUserNotFound is returned when user is not found
	ErrUserNotFound = errors.NotFound("user not found")
	// ErrDisputeNotFound is returned when dispute is not found
	ErrDisputeNotFound = errors.NotFound("dispute not found")
	// ErrModerationNotFound is returned when moderation item is not found
	ErrModerationNotFound = errors.NotFound("moderation item not found")
	// ErrInvalidStatus is returned for invalid status
	ErrInvalidStatus = errors.BadRequest("invalid status")
)

// AdminService handles admin business logic
type AdminService struct {
	repo         repository.AdminRepoInterface
	natsClient   *natsclient.Client
}

// AdminServiceInterface defines the interface for admin service
type AdminServiceInterface interface {
	// User management
	SearchUsers(ctx context.Context, req model.UserListRequest) (model.UserListResponse, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (model.UserResponse, error)
	BanUser(ctx context.Context, adminID, userID uuid.UUID, req model.BanRequest) error
	UnbanUser(ctx context.Context, adminID, userID uuid.UUID, req model.UnbanRequest) error
	AssignRoles(ctx context.Context, adminID, userID uuid.UUID, req model.AssignRolesRequest) error

	// Disputes
	ListDisputes(ctx context.Context, req model.DisputesListRequest) (model.DisputesListResponse, error)
	GetDisputeByID(ctx context.Context, disputeID uuid.UUID) (model.DisputeResponse, error)
	ResolveDispute(ctx context.Context, adminID, disputeID uuid.UUID, req model.ResolveDisputeRequest) error

	// Moderation
	ListModerationQueue(ctx context.Context, req model.ModerationQueueListRequest) (model.ModerationQueueListResponse, error)
	ReviewModerationItem(ctx context.Context, adminID, itemID uuid.UUID, req model.ReviewModerationRequest) error

	// Payouts
	OverridePayout(ctx context.Context, adminID, payoutID uuid.UUID, req model.OverridePayoutRequest) error

	// Audit
	ListAdminActions(ctx context.Context, req model.ActionsListRequest) (model.ActionsListResponse, error)
}

// NewAdminService creates a new AdminService
func NewAdminService(repo repository.AdminRepoInterface, natsClient *natsclient.Client) *AdminService {
	return &AdminService{
		repo:       repo,
		natsClient: natsClient,
	}
}

// checkAdminPermission checks if the user has admin role
func checkAdminPermission(ctx context.Context) error {
	role := middleware.GetUserRole(ctx)
	if role != "admin" {
		return ErrUnauthorized
	}
	return nil
}

// checkPermission checks if user has specific permission (stub for now)
func checkPermission(ctx context.Context, permission string) error {
	role := middleware.GetUserRole(ctx)
	if role != "admin" {
		return ErrUnauthorized
	}
	return nil
}

// SearchUsers searches for users
func (s *AdminService) SearchUsers(ctx context.Context, req model.UserListRequest) (model.UserListResponse, error) {
	if err := checkPermission(ctx, "users:read"); err != nil {
		return model.UserListResponse{}, err
	}

	users, total, err := s.repo.SearchUsers(ctx, req)
	if err != nil {
		return model.UserListResponse{}, fmt.Errorf("failed to search users: %w", err)
	}

	userResponses := make([]model.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	limit := req.Limit
	if limit == 0 {
		limit = 20
	}

	return model.UserListResponse{
		Users:  userResponses,
		Total:  total,
		Limit:  limit,
		Offset: req.Offset,
	}, nil
}

// GetUserByID gets a user by ID
func (s *AdminService) GetUserByID(ctx context.Context, userID uuid.UUID) (model.UserResponse, error) {
	if err := checkPermission(ctx, "users:read"); err != nil {
		return model.UserResponse{}, err
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return model.UserResponse{}, ErrUserNotFound
		}
		return model.UserResponse{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user.ToResponse(), nil
}

// BanUser bans a user
func (s *AdminService) BanUser(ctx context.Context, adminID, userID uuid.UUID, req model.BanRequest) error {
	if err := checkPermission(ctx, "users:ban"); err != nil {
		return err
	}

	// Update user status to banned
	if err := s.repo.UpdateUserStatus(ctx, userID, "banned"); err != nil {
		if err == repository.ErrUserNotFound {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to ban user: %w", err)
	}

	// Create admin action record
	action := &model.AdminAction{
		ID:           uuid.Must(uuid.NewV7()),
		AdminID:     adminID,
		ActionType:  "ban",
		TargetUserID: &userID,
		TargetType:  "user",
		TargetID:    &userID,
		Reason:      req.Reason,
		CreatedAt:   time.Now(),
	}

	metadata := map[string]interface{}{
		"duration": req.Duration,
	}
	metadataBytes, _ := json.Marshal(metadata)
	action.Metadata = metadataBytes

	if err := s.repo.CreateAdminAction(ctx, action); err != nil {
		return fmt.Errorf("failed to create admin action: %w", err)
	}

	// Emit user.banned event
	if s.natsClient != nil && s.natsClient.IsConnected() {
		event := map[string]interface{}{
			"type":      "user.banned",
			"user_id":   userID.String(),
			"admin_id":  adminID.String(),
			"reason":    req.Reason,
			"duration":  req.Duration,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		eventBytes, _ := json.Marshal(event)
		s.natsClient.Publish("user.banned", eventBytes)
	}

	return nil
}

// UnbanUser unbans a user
func (s *AdminService) UnbanUser(ctx context.Context, adminID, userID uuid.UUID, req model.UnbanRequest) error {
	if err := checkPermission(ctx, "users:ban"); err != nil {
		return err
	}

	// Update user status to active
	if err := s.repo.UpdateUserStatus(ctx, userID, "active"); err != nil {
		if err == repository.ErrUserNotFound {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to unban user: %w", err)
	}

	// Create admin action record
	action := &model.AdminAction{
		ID:           uuid.Must(uuid.NewV7()),
		AdminID:     adminID,
		ActionType:  "unban",
		TargetUserID: &userID,
		TargetType:  "user",
		TargetID:    &userID,
		Reason:      req.Reason,
		CreatedAt:   time.Now(),
	}

	if err := s.repo.CreateAdminAction(ctx, action); err != nil {
		return fmt.Errorf("failed to create admin action: %w", err)
	}

	return nil
}

// AssignRoles assigns roles to a user
func (s *AdminService) AssignRoles(ctx context.Context, adminID, userID uuid.UUID, req model.AssignRolesRequest) error {
	if err := checkPermission(ctx, "users:ban"); err != nil {
		return err
	}

	if len(req.Roles) == 0 {
		return errors.BadRequest("at least one role is required")
	}

	// Update user roles
	if err := s.repo.UpdateUserRoles(ctx, userID, req.Roles); err != nil {
		if err == repository.ErrUserNotFound {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to assign roles: %w", err)
	}

	// Create admin action record
	action := &model.AdminAction{
		ID:           uuid.Must(uuid.NewV7()),
		AdminID:     adminID,
		ActionType:  "role_change",
		TargetUserID: &userID,
		TargetType:  "user",
		TargetID:    &userID,
		Reason:      "Role assignment",
		CreatedAt:   time.Now(),
	}

	metadata := map[string]interface{}{
		"roles": req.Roles,
	}
	metadataBytes, _ := json.Marshal(metadata)
	action.Metadata = metadataBytes

	if err := s.repo.CreateAdminAction(ctx, action); err != nil {
		return fmt.Errorf("failed to create admin action: %w", err)
	}

	return nil
}

// ListDisputes lists disputes
func (s *AdminService) ListDisputes(ctx context.Context, req model.DisputesListRequest) (model.DisputesListResponse, error) {
	if err := checkPermission(ctx, "campaigns:audit"); err != nil {
		return model.DisputesListResponse{}, err
	}

	disputes, total, err := s.repo.ListDisputes(ctx, req)
	if err != nil {
		return model.DisputesListResponse{}, fmt.Errorf("failed to list disputes: %w", err)
	}

	disputeResponses := make([]model.DisputeResponse, len(disputes))
	for i, dispute := range disputes {
		disputeResponses[i] = dispute.ToResponse()
	}

	limit := req.Limit
	if limit == 0 {
		limit = 20
	}

	return model.DisputesListResponse{
		Disputes: disputeResponses,
		Total:   total,
		Limit:  limit,
		Offset: req.Offset,
	}, nil
}

// GetDisputeByID gets a dispute by ID
func (s *AdminService) GetDisputeByID(ctx context.Context, disputeID uuid.UUID) (model.DisputeResponse, error) {
	if err := checkPermission(ctx, "campaigns:audit"); err != nil {
		return model.DisputeResponse{}, err
	}

	dispute, err := s.repo.GetDisputeByID(ctx, disputeID)
	if err != nil {
		if err == repository.ErrDisputeNotFound {
			return model.DisputeResponse{}, ErrDisputeNotFound
		}
		return model.DisputeResponse{}, fmt.Errorf("failed to get dispute: %w", err)
	}

	return dispute.ToResponse(), nil
}

// ResolveDispute resolves a dispute
func (s *AdminService) ResolveDispute(ctx context.Context, adminID, disputeID uuid.UUID, req model.ResolveDisputeRequest) error {
	if err := checkPermission(ctx, "campaigns:audit"); err != nil {
		return err
	}

	// Update dispute status
	status := "resolved"
	if err := s.repo.UpdateDisputeStatus(ctx, disputeID, status, req.Resolution, adminID); err != nil {
		if err == repository.ErrDisputeNotFound {
			return ErrDisputeNotFound
		}
		return fmt.Errorf("failed to resolve dispute: %w", err)
	}

	// Create admin action record
	targetID := disputeID
	action := &model.AdminAction{
		ID:          uuid.Must(uuid.NewV7()),
		AdminID:    adminID,
		ActionType: "dispute_resolved",
		TargetType: "dispute",
		TargetID:   &targetID,
		Reason:     req.Resolution,
		CreatedAt:  time.Now(),
	}

	metadata := map[string]interface{}{
		"action_taken": req.ActionTaken,
	}
	metadataBytes, _ := json.Marshal(metadata)
	action.Metadata = metadataBytes

	if err := s.repo.CreateAdminAction(ctx, action); err != nil {
		return fmt.Errorf("failed to create admin action: %w", err)
	}

	return nil
}

// ListModerationQueue lists moderation queue items
func (s *AdminService) ListModerationQueue(ctx context.Context, req model.ModerationQueueListRequest) (model.ModerationQueueListResponse, error) {
	if err := checkPermission(ctx, "support:escalate"); err != nil {
		return model.ModerationQueueListResponse{}, err
	}

	items, total, err := s.repo.ListModerationQueue(ctx, req)
	if err != nil {
		return model.ModerationQueueListResponse{}, fmt.Errorf("failed to list moderation queue: %w", err)
	}

	itemResponses := make([]model.ModerationQueueResponse, len(items))
	for i, item := range items {
		itemResponses[i] = item.ToResponse()
	}

	limit := req.Limit
	if limit == 0 {
		limit = 20
	}

	return model.ModerationQueueListResponse{
		Items:  itemResponses,
		Total:  total,
		Limit:  limit,
		Offset: req.Offset,
	}, nil
}

// ReviewModerationItem reviews a moderation queue item
func (s *AdminService) ReviewModerationItem(ctx context.Context, adminID, itemID uuid.UUID, req model.ReviewModerationRequest) error {
	if err := checkPermission(ctx, "support:escalate"); err != nil {
		return err
	}

	// Update moderation queue status
	if err := s.repo.UpdateModerationQueueStatus(ctx, itemID, "reviewed", adminID, req.ActionTaken); err != nil {
		if err == repository.ErrModerationNotFound {
			return ErrModerationNotFound
		}
		return fmt.Errorf("failed to review moderation item: %w", err)
	}

	// Create admin action record
	targetID := itemID
	action := &model.AdminAction{
		ID:          uuid.Must(uuid.NewV7()),
		AdminID:    adminID,
		ActionType: "content_review",
		TargetType: "moderation",
		TargetID:   &targetID,
		Reason:     req.ActionTaken,
		CreatedAt:  time.Now(),
	}

	if err := s.repo.CreateAdminAction(ctx, action); err != nil {
		return fmt.Errorf("failed to create admin action: %w", err)
	}

	return nil
}

// OverridePayout overrides a payout
func (s *AdminService) OverridePayout(ctx context.Context, adminID, payoutID uuid.UUID, req model.OverridePayoutRequest) error {
	if err := checkPermission(ctx, "payouts:override"); err != nil {
		return err
	}

	if req.NewAmount < 0 {
		return errors.BadRequest("new amount must be non-negative")
	}

	// Create admin action record
	targetID := payoutID
	action := &model.AdminAction{
		ID:          uuid.Must(uuid.NewV7()),
		AdminID:    adminID,
		ActionType: "payout_override",
		TargetType: "payout",
		TargetID:   &targetID,
		Reason:    req.Reason,
		CreatedAt: time.Now(),
	}

	metadata := map[string]interface{}{
		"new_amount": req.NewAmount,
	}
	metadataBytes, _ := json.Marshal(metadata)
	action.Metadata = metadataBytes

	if err := s.repo.CreateAdminAction(ctx, action); err != nil {
		return fmt.Errorf("failed to create admin action: %w", err)
	}

	// Emit payout.overridden event
	if s.natsClient != nil && s.natsClient.IsConnected() {
		event := map[string]interface{}{
			"type":       "payout.overridden",
			"payout_id":  payoutID.String(),
			"admin_id":   adminID.String(),
			"new_amount": req.NewAmount,
			"reason":    req.Reason,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		eventBytes, _ := json.Marshal(event)
		s.natsClient.Publish("payout.overridden", eventBytes)
	}

	return nil
}

// ListAdminActions lists admin actions
func (s *AdminService) ListAdminActions(ctx context.Context, req model.ActionsListRequest) (model.ActionsListResponse, error) {
	if err := checkPermission(ctx, "campaigns:audit"); err != nil {
		return model.ActionsListResponse{}, err
	}

	actions, total, err := s.repo.ListAdminActions(ctx, req)
	if err != nil {
		return model.ActionsListResponse{}, fmt.Errorf("failed to list admin actions: %w", err)
	}

	actionResponses := make([]model.AdminActionResponse, len(actions))
	for i, action := range actions {
		var targetUserID, targetID *string
		if action.TargetUserID != nil {
			id := action.TargetUserID.String()
			targetUserID = &id
		}
		if action.TargetID != nil {
			id := action.TargetID.String()
			targetID = &id
		}
		actionResponses[i] = model.AdminActionResponse{
			ID:           action.ID.String(),
			AdminID:      action.AdminID.String(),
			ActionType:  action.ActionType,
			TargetUserID: targetUserID,
			TargetType:  action.TargetType,
			TargetID:    targetID,
			Reason:      action.Reason,
			Metadata:    action.Metadata,
			CreatedAt:   action.CreatedAt,
		}
	}

	limit := req.Limit
	if limit == 0 {
		limit = 20
	}

	return model.ActionsListResponse{
		Actions: actionResponses,
		Total:   total,
		Limit:  limit,
		Offset: req.Offset,
	}, nil
}