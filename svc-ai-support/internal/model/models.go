package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Conversation represents an AI support conversation
type Conversation struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Status    string     `json:"status"`
	Topic     string     `json:"topic"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
}

// ConversationResponse is the JSON response for conversation data
type ConversationResponse struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Status    string     `json:"status"`
	Topic     string     `json:"topic"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
	Messages  []Message `json:"messages,omitempty"`
}

// ToResponse converts Conversation to ConversationResponse
func (c *Conversation) ToResponse() ConversationResponse {
	return ConversationResponse{
		ID:        c.ID,
		UserID:    c.UserID,
		Status:    c.Status,
		Topic:     c.Topic,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		ClosedAt:  c.ClosedAt,
	}
}

// Message represents a message in a conversation
type Message struct {
	ID             uuid.UUID       `json:"id"`
	ConversationID uuid.UUID      `json:"conversation_id"`
	SenderType    string          `json:"sender_type"`
	Content       string          `json:"content"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
}

// MessageMetadata contains additional message metadata
type MessageMetadata struct {
	ConfidenceScore *float64 `json:"confidence_score,omitempty"`
	Sources         []string `json:"sources,omitempty"`
}

// Escalation represents a support escalation
type Escalation struct {
	ID             uuid.UUID  `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	EscalatedBy   string     `json:"escalated_by"`
	Reason        string     `json:"reason,omitempty"`
	AdminID       *uuid.UUID `json:"admin_id,omitempty"`
	Status        string     `json:"status"`
	Notes         string     `json:"notes,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
}

// EscalationResponse is the JSON response for escalation data
type EscalationResponse struct {
	ID             uuid.UUID  `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	EscalatedBy   string     `json:"escalated_by"`
	Reason        string     `json:"reason,omitempty"`
	Status        string     `json:"status"`
	Notes         string     `json:"notes,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
}

// ToResponse converts Escalation to EscalationResponse
func (e *Escalation) ToResponse() EscalationResponse {
	return EscalationResponse{
		ID:             e.ID,
		ConversationID: e.ConversationID,
		EscalatedBy:    e.EscalatedBy,
		Reason:        e.Reason,
		Status:        e.Status,
		Notes:         e.Notes,
		CreatedAt:     e.CreatedAt,
		ResolvedAt:    e.ResolvedAt,
	}
}

// ChatRequest represents a chat API request
type ChatRequest struct {
	Message        string `json:"message"`
	ConversationID *uuid.UUID `json:"conversation_id,omitempty"`
}

// ChatResponse represents a chat API response
type ChatResponse struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	Messages       []Message `json:"messages"`
	AIResponse     string   `json:"ai_response"`
	ShouldEscalate bool     `json:"should_escalate"`
}

// EscalateRequest represents an escalation API request
type EscalateRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ResolveEscalationRequest represents a resolve escalation API request
type ResolveEscalationRequest struct {
	Notes string `json:"notes"`
}

// UserContext represents the user's context (stub)
type UserContext struct {
	UserID uuid.UUID `json:"user_id"`
	// TODO: Implement context loading from User Service
	// FirstName string `json:"first_name"`
	// LastName string `json:"last_name"`
	// Email string `json:"email"`
	// TODO: Implement loading briefs from Brief Service
	// BriefCount int `json:"brief_count"`
	// TODO: Implement loading videos from Video Service
	// VideoCount int `json:"video_count"`
	// TODO: Implement loading balance from Payout Service
	// Balance float64 `json:"balance"`
}

// NATSEvent represents a NATS event for escalation
type NATSEvent struct {
	Type           string    `json:"type"`
	ConversationID uuid.UUID `json:"conversation_id"`
	EscalationID   uuid.UUID `json:"escalation_id"`
	UserID         uuid.UUID `json:"user_id"`
	Reason         string   `json:"reason"`
	Timestamp     time.Time `json:"timestamp"`
}