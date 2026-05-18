package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	resp := UserResponse{
		ID:        u.ID.String(),
		Email:     u.Email,
		Name:      u.FirstName + " " + u.LastName,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
	if u.LastLoginAt != nil {
		resp.LastLogin = u.LastLoginAt
	}
	return resp
}

// UserResponse is the API response for a user
type UserResponse struct {
	ID        uuid.UUID  `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	LastLogin *time.Time `json:"last_login,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// UserListRequest is the request for listing users
type UserListRequest struct {
	Role    string `json:"role"`
	Status string `json:"status"`
	Email  string `json:"email"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// UserListResponse is the response for listing users
type UserListResponse struct {
	Users  []UserResponse `json:"users"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// BanRequest is the request for banning a user
type BanRequest struct {
	Reason   string `json:"reason"`
	Duration string `json:"duration,omitempty"` // e.g., "24h", "7d", "permanent"
}

// UnbanRequest is the request for unbanning a user
type UnbanRequest struct {
	Reason string `json:"reason"`
}

// AssignRolesRequest is the request for assigning roles
type AssignRolesRequest struct {
	Roles []string `json:"roles"`
}

// AdminAction represents an admin action
type AdminAction struct {
	ID           uuid.UUID              `json:"id"`
	AdminID     uuid.UUID            `json:"admin_id"`
	ActionType  string               `json:"action_type"`
	TargetUserID *uuid.UUID         `json:"target_user_id,omitempty"`
	TargetType  string               `json:"target_type"`
	TargetID    *uuid.UUID          `json:"target_id,omitempty"`
	Reason      string               `json:"reason"`
	Metadata    json.RawMessage     `json:"metadata,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
}

// AdminActionResponse is the API response for an admin action
type AdminActionResponse struct {
	ID           string          `json:"id"`
	AdminID      string         `json:"admin_id"`
	ActionType  string         `json:"action_type"`
	TargetUserID *string      `json:"target_user_id,omitempty"`
	TargetType  string         `json:"target_type"`
	TargetID    *string       `json:"target_id,omitempty"`
	Reason      string         `json:"reason"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}

// DisputesListRequest is the request for listing disputes
type DisputesListRequest struct {
	Status string `json:"status"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// DisputesListResponse is the response for listing disputes
type DisputesListResponse struct {
	Disputes []DisputeResponse `json:"disputes"`
	Total   int               `json:"total"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
}

// Dispute represents a dispute
type Dispute struct {
	ID           uuid.UUID       `json:"id"`
	ReporterID   uuid.UUID      `json:"reporter_id"`
	RespondentID uuid.UUID      `json:"respondent_id"`
	TargetType   string         `json:"target_type"`
	TargetID     uuid.UUID      `json:"target_id"`
	Status       string         `json:"status"`
	Description  string        `json:"description"`
	Evidence     json.RawMessage `json:"evidence,omitempty"`
	Resolution   *string        `json:"resolution,omitempty"`
	ResolvedBy   *uuid.UUID     `json:"resolved_by,omitempty"`
	ResolvedAt   *time.Time     `json:"resolved_at,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// ToResponse converts Dispute to DisputeResponse
func (d *Dispute) ToResponse() DisputeResponse {
	resp := DisputeResponse{
		ID:           d.ID.String(),
		ReporterID:   d.ReporterID.String(),
		RespondentID: d.RespondentID.String(),
		TargetType:  d.TargetType,
		TargetID:    d.TargetID.String(),
		Status:      d.Status,
		Description: d.Description,
		Evidence:   d.Evidence,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}
	if d.Resolution != nil {
		resp.Resolution = *d.Resolution
	}
	if d.ResolvedBy != nil {
		resp.ResolvedBy = *d.ResolvedBy
	}
	if d.ResolvedAt != nil {
		resp.ResolvedAt = *d.ResolvedAt
	}
	return resp
}

// DisputeResponse is the API response for a dispute
type DisputeResponse struct {
	ID           string          `json:"id"`
	ReporterID   string         `json:"reporter_id"`
	RespondentID string         `json:"respondent_id"`
	TargetType  string         `json:"target_type"`
	TargetID    string         `json:"target_id"`
	Status      string         `json:"status"`
	Description string        `json:"description"`
	Evidence   json.RawMessage `json:"evidence,omitempty"`
	Resolution string        `json:"resolution,omitempty"`
	ResolvedBy string        `json:"resolved_by,omitempty"`
	ResolvedAt time.Time     `json:"resolved_at,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// ResolveDisputeRequest is the request for resolving a dispute
type ResolveDisputeRequest struct {
	Resolution string `json:"resolution"`
	ActionTaken string `json:"action_taken"`
}

// ModerationQueueListRequest is the request for listing moderation queue
type ModerationQueueListRequest struct {
	Status string `json:"status"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// ModerationQueueListResponse is the response for listing moderation queue
type ModerationQueueListResponse struct {
	Items  []ModerationQueueResponse `json:"items"`
	Total  int                    `json:"total"`
	Limit  int                    `json:"limit"`
	Offset int                    `json:"offset"`
}

// ModerationQueue represents a moderation queue item
type ModerationQueue struct {
	ID          uuid.UUID  `json:"id"`
	TargetType  string    `json:"target_type"`
	TargetID   uuid.UUID `json:"target_id"`
	FlagReason string    `json:"flag_reason"`
	Status    string    `json:"status"`
	ReviewedBy *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	ActionTaken *string  `json:"action_taken,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts ModerationQueue to ModerationQueueResponse
func (m *ModerationQueue) ToResponse() ModerationQueueResponse {
	resp := ModerationQueueResponse{
		ID:          m.ID.String(),
		TargetType:  m.TargetType,
		TargetID:   m.TargetID.String(),
		FlagReason: m.FlagReason,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
	}
	if m.ReviewedBy != nil {
		resp.ReviewedBy = *m.ReviewedBy
	}
	if m.ReviewedAt != nil {
		resp.ReviewedAt = *m.ReviewedAt
	}
	if m.ActionTaken != nil {
		resp.ActionTaken = *m.ActionTaken
	}
	return resp
}

// ModerationQueueResponse is the API response for a moderation queue item
type ModerationQueueResponse struct {
	ID          string    `json:"id"`
	TargetType  string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	FlagReason string    `json:"flag_reason"`
	Status    string    `json:"status"`
	ReviewedBy string    `json:"reviewed_by,omitempty"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	ActionTaken string    `json:"action_taken,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ReviewModerationRequest is the request for reviewing a moderation item
type ReviewModerationRequest struct {
	ActionTaken string `json:"action_taken"`
}

// OverridePayoutRequest is the request for overriding a payout
type OverridePayoutRequest struct {
	NewAmount int64  `json:"new_amount"`
	Reason    string `json:"reason"`
}

// ActionsListRequest is the request for listing admin actions
type ActionsListRequest struct {
	AdminID    string `json:"admin_id"`
	ActionType string `json:"action_type"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

// ActionsListResponse is the response for listing admin actions
type ActionsListResponse struct {
	Actions []AdminActionResponse `json:"actions"`
	Total   int                  `json:"total"`
	Limit  int                  `json:"limit"`
	Offset int                  `json:"offset"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Limit  int
	Offset int
}

// DefaultPagination returns default pagination params
func DefaultPagination() PaginationParams {
	return PaginationParams{
		Limit:  20,
		Offset: 0,
	}
}