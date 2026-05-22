package model

import "time"

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