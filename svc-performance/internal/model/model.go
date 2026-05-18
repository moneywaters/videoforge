package model

import "time"

// PerformanceEvent represents a tracked performance event
type PerformanceEvent struct {
	ID         string    `json:"id"`
	VideoID   string    `json:"video_id"`
	EventType string    `json:"event_type"` // view, click, conversion
	Source    string    `json:"source"` // youtube, tiktok, direct
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time `json:"timestamp"`
}

// Analytics represents aggregated analytics data
type Analytics struct {
	ID        string    `json:"id"`
	VideoID  string    `json:"video_id"`
	Date    time.Time `json:"date"`
	Views   int64     `json:"views"`
	Clicks  int64     `json:"clicks"`
	Conversions int64   `json:"conversions"`
	WatchTime int64    `json:"watch_time"` // seconds
	AvgViewDuration float64 `json:"avg_view_duration"`
}

// DashboardMetrics represents dashboard summary metrics
type DashboardMetrics struct {
	UserID        string  `json:"user_id"`
	TotalVideos  int64   `json:"total_videos"`
	TotalViews   int64   `json:"total_views"`
	TotalClicks int64   `json:"total_clicks"`
	TotalConversions int64 `json:"total_conversions"`
	Revenue     float64 `json:"revenue"`
}