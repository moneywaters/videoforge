package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/videoforge/backend/svc-ai-support/internal/model"
	"github.com/videoforge/backend/svc-ai-support/internal/repository"

	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/natsclient"
)

// SupportService handles business logic for AI support
type SupportService struct {
	repo     repository.SupportRepoInterface
	nats     natsclient.NATSClient
	logger   *logger.Logger
}

// NewSupportService creates a new SupportService
func NewSupportService(db *pgxpool.Pool, natsClient natsclient.NATSClient) *SupportService {
	return &SupportService{
		repo:   repository.NewSupportRepository(db),
		nats:   natsClient,
		logger: logger.Default("development"),
	}
}

// SupportServiceInterface defines the interface for support service
type SupportServiceInterface interface {
	Chat(ctx context.Context, userID uuid.UUID, req model.ChatRequest) (*model.ChatResponse, error)
	GetConversations(ctx context.Context, userID uuid.UUID) ([]model.Conversation, error)
	GetConversation(ctx context.Context, userID, conversationID uuid.UUID) (*model.ConversationResponse, error)
	Escalate(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID, reason string) (*model.EscalationResponse, error)
	ResolveEscalation(ctx context.Context, adminID uuid.UUID, escalationID uuid.UUID, notes string) (*model.EscalationResponse, error)
	GetEscalations(ctx context.Context, adminID uuid.UUID, status *string) ([]model.EscalationResponse, error)
	LoadUserContext(ctx context.Context, userID uuid.UUID) (model.UserContext, error)
}

// Chat handles a chat message from a user
func (s *SupportService) Chat(ctx context.Context, userID uuid.UUID, req model.ChatRequest) (*model.ChatResponse, error) {
	// Determine if we need to create a new conversation or use existing
	var conv *model.Conversation
	var convID uuid.UUID

	if req.ConversationID != nil {
		// Use existing conversation
		var err error
		conv, err = s.repo.GetConversation(ctx, *req.ConversationID, userID)
		if err != nil {
			return nil, err
		}
		convID = conv.ID
	} else {
		// Create new conversation
		// Extract topic from first message (simple extraction)
		topic := extractTopic(req.Message)
		conv = &model.Conversation{
			UserID: userID,
			Topic:  topic,
		}
		if err := s.repo.CreateConversation(ctx, conv); err != nil {
			return nil, fmt.Errorf("failed to create conversation: %w", err)
		}
		convID = conv.ID
	}

	// Store user message
	userMsg := &model.Message{
		ConversationID: convID,
		SenderType:    "user",
		Content:       req.Message,
		Metadata:      json.RawMessage(`{}`),
	}
	if err := s.repo.CreateMessage(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("failed to store user message: %w", err)
	}

	// Load user context (stub)
	userContext, err := s.LoadUserContext(ctx, userID)
	if err != nil {
		s.logger.Warn("failed to load user context", "error", err)
	}

	// Generate AI response
	aiResponse, confidenceScore := generateAIResponse(req.Message, userContext)

	// Determine if we should escalate
	shouldEscalate := shouldEscalateToHuman(req.Message, confidenceScore)

	// Store AI response with metadata
	metadata := model.MessageMetadata{
		ConfidenceScore: &confidenceScore,
		Sources:        []string{},
	}
	metadataBytes, _ := json.Marshal(metadata)

	aiMsg := &model.Message{
		ConversationID: convID,
		SenderType:     "ai",
		Content:        aiResponse,
		Metadata:       metadataBytes,
	}
	if err := s.repo.CreateMessage(ctx, aiMsg); err != nil {
		return nil, fmt.Errorf("failed to store AI message: %w", err)
	}

	// Auto-escalate if confidence is low
	if shouldEscalate {
		reason := fmt.Sprintf("Low confidence score: %.2f", confidenceScore)
		if err := s.autoEscalate(ctx, convID, userID, reason); err != nil {
			s.logger.Warn("failed to auto-escalate", "error", err)
		}
	}

	// Get all messages for response
	messages, err := s.repo.GetMessages(ctx, convID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return &model.ChatResponse{
		ConversationID: convID,
		Messages:      messages,
		AIResponse:    aiResponse,
		ShouldEscalate: shouldEscalate,
	}, nil
}

// GetConversations returns all conversations for a user
func (s *SupportService) GetConversations(ctx context.Context, userID uuid.UUID) ([]model.Conversation, error) {
	return s.repo.GetConversations(ctx, userID)
}

// GetConversation returns a specific conversation with all messages
func (s *SupportService) GetConversation(ctx context.Context, userID, conversationID uuid.UUID) (*model.ConversationResponse, error) {
	conv, err := s.repo.GetConversation(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}

	messages, err := s.repo.GetMessages(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	response := conv.ToResponse()
	response.Messages = messages

	return &response, nil
}

// Escalate escalates a conversation to human support
func (s *SupportService) Escalate(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID, reason string) (*model.EscalationResponse, error) {
	// Verify conversation belongs to user
	conv, err := s.repo.GetConversation(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}

	// Update conversation status
	if err := s.repo.UpdateConversationStatus(ctx, conversationID, "escalated"); err != nil {
		return nil, err
	}

	// Create escalation record
	escalation := &model.Escalation{
		ConversationID: conversationID,
		EscalatedBy:    "user",
		Reason:         reason,
	}
	if err := s.repo.CreateEscalation(ctx, escalation); err != nil {
		return nil, fmt.Errorf("failed to create escalation: %w", err)
	}

	// Emit NATS event
	s.emitEscalationEvent(ctx, conv, escalation)

	return escalation.ToResponse(), nil
}

// ResolveEscalation resolves an escalation (admin only)
func (s *SupportService) ResolveEscalation(ctx context.Context, adminID uuid.UUID, escalationID uuid.UUID, notes string) (*model.EscalationResponse, error) {
	// Update escalation status
	if err := s.repo.UpdateEscalationStatus(ctx, escalationID, "resolved", notes); err != nil {
		return nil, err
	}

	// Get escalation to find conversation ID
	escalation, err := s.repo.GetEscalation(ctx, escalationID)
	if err != nil {
		return nil, err
	}

	// Update conversation status to closed
	if err := s.repo.UpdateConversationStatus(ctx, escalation.ConversationID, "closed"); err != nil {
		return nil, err
	}

	return escalation.ToResponse(), nil
}

// GetEscalations returns escalations (admin only)
func (s *SupportService) GetEscalations(ctx context.Context, adminID uuid.UUID, status *string) ([]model.EscalationResponse, error) {
	escalations, err := s.repo.GetEscalations(ctx, &adminID, status)
	if err != nil {
		return nil, err
	}

	responses := make([]model.EscalationResponse, len(escalations))
	for i, esc := range escalations {
		responses[i] = *esc.ToResponse()
	}

	return responses, nil
}

// LoadUserContext loads user context for AI response generation
// TODO: Implement actual context loading from User Service, Brief Service, Video Service, and Payout Service
func (s *SupportService) LoadUserContext(ctx context.Context, userID uuid.UUID) (model.UserContext, error) {
	// For MVP, just return user ID
	// TODO: Implement context loading from cross-service calls
	// - Get user profile from User Service
	// - Get briefs from Brief Service
	// - Get videos from Video Service
	// - Get balance from Payout Service
	return model.UserContext{
		UserID: userID,
	}, nil
}

// generateAIResponse generates an AI response based on the user's message
// This is a STUB implementation
func generateAIResponse(message string, context model.UserContext) (string, float64) {
	msg := toLower(message)

	// Check for payout/payment related queries
	if contains(msg, "payout") || contains(msg, "payment") {
		// TODO: Load actual balance from Payout Service
		return "I can help with payout questions. Your current balance is available in the Payouts section. Earnings are held for 14 days before becoming available.", 0.9
	}

	// Check for brief/project related queries
	if contains(msg, "brief") || contains(msg, "project") {
		// TODO: Load actual brief count from Brief Service
		return "For brief-related questions, you can view your active briefs in the Briefs dashboard. Would you like help creating a new brief?", 0.85
	}

	// Check for video/edit related queries
	if contains(msg, "video") || contains(msg, "edit") {
		// TODO: Load actual video count from Video Service
		return "Video submissions are reviewed by clients. You can check the status of your videos in the Videos section.", 0.85
	}

	// Check for campaign/ad related queries
	if contains(msg, "campaign") || contains(msg, "ad") {
		return "Campaign performance data is available in the Performance dashboard. Would you like tips on optimizing your campaigns?", 0.85
	}

	// Default response
	return "I'm here to help! Could you provide more details about your question? You can also type 'human' to speak with a support agent.", 0.7
}

// shouldEscalateToHuman determines if the conversation should be escalated
func shouldEscalateToHuman(message string, confidenceScore float64) bool {
	msg := toLower(message)

	// Escalation trigger words
	escalationTriggers := []string{"human", "agent", "support", "help me", "escalate", "talk to a person"}

	for _, trigger := range escalationTriggers {
		if contains(msg, trigger) {
			return true
		}
	}

	// Low confidence score - random 10% chance
	if confidenceScore < 0.5 && rand.Float64() < 0.1 {
		return true
	}

	return false
}

// extractTopic extracts a simple topic from a message
func extractTopic(message string) string {
	msg := toLower(message)

	if contains(msg, "payout") || contains(msg, "payment") {
		return "Payout Question"
	}
	if contains(msg, "brief") || contains(msg, "project") {
		return "Brief/Project Question"
	}
	if contains(msg, "video") || contains(msg, "edit") {
		return "Video Question"
	}
	if contains(msg, "campaign") || contains(msg, "ad") {
		return "Campaign Question"
	}

	return "General Inquiry"
}

// autoEscalate automatically escalates a conversation
func (s *SupportService) autoEscalate(ctx context.Context, conversationID, userID uuid.UUID, reason string) error {
	// Update conversation status
	if err := s.repo.UpdateConversationStatus(ctx, conversationID, "escalated"); err != nil {
		return err
	}

	// Create escalation record
	escalation := &model.Escalation{
		ConversationID: conversationID,
		EscalatedBy:   "auto",
		Reason:        reason,
	}
	if err := s.repo.CreateEscalation(ctx, escalation); err != nil {
		return err
	}

	// Get conversation for event
	conv, err := s.repo.GetConversation(ctx, conversationID, userID)
	if err != nil {
		return err
	}

	// Emit NATS event
	s.emitEscalationEvent(ctx, conv, escalation)

	return nil
}

// emitEscalationEvent emits a NATS event for escalation
func (s *SupportService) emitEscalationEvent(ctx context.Context, conv *model.Conversation, escalation *model.Escalation) {
	event := model.NATSEvent{
		Type:           "support.escalated",
		ConversationID: conv.ID,
		EscalationID:   escalation.ID,
		UserID:         conv.UserID,
		Reason:         escalation.Reason,
		Timestamp:     time.Now(),
	}

	data, err := json.Marshal(event)
	if err != nil {
		s.logger.Error("failed to marshal escalation event", "error", err)
		return
	}

	if s.nats != nil && s.nats.IsConnected() {
		if err := s.nats.Publish("support.escalated", data); err != nil {
			s.logger.Error("failed to publish escalation event", "error", err)
		}
	}
}

// Helper functions
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}