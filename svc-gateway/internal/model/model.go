package model

import "time"

// Route represents an API route configuration
type Route struct {
	ID        string   `json:"id"`
	Path     string   `json:"path"`
	Methods  []string `json:"methods"`
	Service  string   `json:"service"`
	Endpoint string   `json:"endpoint"`
	Active   bool     `json:"active"`
}

// APIKey represents an API key for authentication
type APIKey struct {
	ID        string    `json:"id"`
	Key      string    `json:"key"`
	UserID   string    `json:"user_id"`
	Name     string    `json:"name"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Active   bool      `json:"active"`
}

// Time is a custom time type for JSON serialization
type Time struct {
	Time time.Time
}