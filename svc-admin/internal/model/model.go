package model

import "time"

// User represents an admin user view
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"` // admin, super_admin
	Status    string    `json:"status"` // active, suspended
	LastLogin *time.Time `json:"last_login,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        string    `json:"id"`
	UserID   string    `json:"user_id"`
	Action  string    `json:"action"` // user_create, user_update, video_approve
	Resource string   `json:"resource"` // user, video, campaign
	ResourceID string `json:"resource_id"`
	Details  map[string]interface{} `json:"details"`
	IPAddress string  `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}

// SystemStats represents system-wide statistics
type SystemStats struct {
	TotalUsers     int64   `json:"total_users"`
	ActiveUsers   int64   `json:"active_users"`
	TotalVideos  int64   `json:"total_videos"`
	TotalCampaigns int64   `json:"total_campaigns"`
	TotalRevenue float64 `json:"total_revenue"`
}