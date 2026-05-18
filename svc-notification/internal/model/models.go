package model

import (
	"encoding/json"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	// Video events
	NotificationVideoSubmitted NotificationType = "video.submitted"
	NotificationVideoApproved NotificationType = "video.approved"
	NotificationVideoRejected NotificationType = "video.rejected"

	// Sale events
	NotificationSaleAttributed NotificationType = "sale.attributed"

	// Payout events
	NotificationPayoutReleased NotificationType = "payout.released"

	// Campaign events
	NotificationCampaignStarted NotificationType = "campaign.started"
	NotificationCampaignEnded   NotificationType = "campaign.ended"
)

// Notification represents a user notification in the database
type Notification struct {
	ID        string          `json:"id"`
	UserID   string          `json:"user_id"`
	Type     NotificationType `json:"type"`
	Title    string          `json:"title"`
	Message  string          `json:"message"`
	Data     json.RawMessage `json:"data,omitempty"`
	Read     bool            `json:"read"`
	ReadAt   *time.Time     `json:"read_at,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

// NotificationResponse represents a notification response for API
type NotificationResponse struct {
	ID        string                 `json:"id"`
	UserID   string                 `json:"user_id"`
	Type     NotificationType      `json:"type"`
	Title    string                 `json:"title"`
	Message  string                 `json:"message"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Read     bool                   `json:"read"`
	ReadAt   *time.Time             `json:"read_at,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// ToResponse converts Notification to NotificationResponse
func (n *Notification) ToResponse() NotificationResponse {
	var data map[string]interface{}
	if n.Data != nil {
		_ = json.Unmarshal(n.Data, &data)
	}
	return NotificationResponse{
		ID:        n.ID,
		UserID:    n.UserID,
		Type:      n.Type,
		Title:     n.Title,
		Message:   n.Message,
		Data:      data,
		Read:      n.Read,
		ReadAt:    n.ReadAt,
		CreatedAt: n.CreatedAt,
	}
}

// UserPreference represents a user's notification preferences
type UserPreference struct {
	ID                 string          `json:"id"`
	UserID             string          `json:"user_id"`
	ChannelPreference  string          `json:"channel_preference"` // ws, email, both
	EnabledTypes       json.RawMessage `json:"enabled_types"`     // array of enabled event types
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// UserPreferenceResponse represents user preference response for API
type UserPreferenceResponse struct {
	ID                 string   `json:"id"`
	UserID             string   `json:"user_id"`
	ChannelPreference  string   `json:"channel_preference"`
	EnabledTypes      []string `json:"enabled_types"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ToResponse converts UserPreference to UserPreferenceResponse
func (p *UserPreference) ToResponse() UserPreferenceResponse {
	var enabledTypes []string
	if p.EnabledTypes != nil {
		_ = json.Unmarshal(p.EnabledTypes, &enabledTypes)
	}
	return UserPreferenceResponse{
		ID:                 p.ID,
		UserID:             p.UserID,
		ChannelPreference:  p.ChannelPreference,
		EnabledTypes:      enabledTypes,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

// WSConnection represents a WebSocket connection
type WSConnection struct {
	ID          string    `json:"id"`
	UserID     string    `json:"user_id"`
	ConnectionID string   `json:"connection_id"`
	ConnectedAt time.Time `json:"connected_at"`
	LastPingAt time.Time `json:"last_ping_at"`
}

// ChannelPreference enum
const (
	ChannelWS    = "ws"
	ChannelEmail = "email"
	ChannelBoth = "both"
)

// CreateNotificationInput represents input for creating a notification
type CreateNotificationInput struct {
	UserID   string          `json:"user_id"`
	Type     NotificationType `json:"type"`
	Title    string          `json:"title"`
	Message  string          `json:"message"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// UpdatePreferenceInput represents input for updating preferences
type UpdatePreferenceInput struct {
	ChannelPreference string   `json:"channel_preference"`
	EnabledTypes   []string `json:"enabled_types"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page    int `json:"page" form:"page"`
	PerPage int `json:"per_page" form:"per_page"`
}

// DefaultPagination returns default pagination
func (p *PaginationParams) Default() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 || p.PerPage > 100 {
		p.PerPage = 20
	}
}

// Offset calculates the offset for pagination
func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// ListNotificationsFilter represents filters for listing notifications
type ListNotificationsFilter struct {
	PaginationParams
	Read *bool `json:"read" form:"read"`
}