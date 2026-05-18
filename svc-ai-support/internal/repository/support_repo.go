package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
	"svc-ai-support/internal/model"

	backenderrors "github.com/videoforge/backend/pkg/errors"
)

var (
	// ErrConversationNotFound is returned when a conversation is not found
	ErrConversationNotFound = errors.New("conversation not found")
	// ErrMessageNotFound is returned when a message is not found
	ErrMessageNotFound = errors.New("message not found")
	// ErrEscalationNotFound is returned when an escalation is not found
	ErrEscalationNotFound = errors.New("escalation not found")
	// ErrNotAuthorized is returned when user is not authorized
	ErrNotAuthorized = errors.New("not authorized")
)

// SupportRepository handles database operations for AI support
type SupportRepository struct {
	db *pgxpool.Pool
}

// NewSupportRepository creates a new SupportRepository
func NewSupportRepository(db *pgxpool.Pool) *SupportRepository {
	return &SupportRepository{db: db}
}

// SupportRepoInterface defines the interface for support repository
type SupportRepoInterface interface {
	CreateConversation(ctx context.Context, conv *model.Conversation) error
	GetConversation(ctx context.Context, id, userID uuid.UUID) (*model.Conversation, error)
	GetConversations(ctx context.Context, userID uuid.UUID) ([]model.Conversation, error)
	UpdateConversationStatus(ctx context.Context, id uuid.UUID, status string) error
	CreateMessage(ctx context.Context, msg *model.Message) error
	GetMessages(ctx context.Context, conversationID uuid.UUID) ([]model.Message, error)
	CreateEscalation(ctx context.Context, esc *model.Escalation) error
	GetEscalation(ctx context.Context, id uuid.UUID) (*model.Escalation, error)
	GetEscalations(ctx context.Context, adminID *uuid.UUID, status *string) ([]model.Escalation, error)
	UpdateEscalationStatus(ctx context.Context, id uuid.UUID, status, notes string) error
}

// CreateConversation creates a new conversation
func (r *SupportRepository) CreateConversation(ctx context.Context, conv *model.Conversation) error {
	conv.ID = uuid.Must(uuid.NewV7())
	conv.CreatedAt = time.Now()
	conv.UpdatedAt = time.Now()
	conv.Status = "active"

	query := `
		INSERT INTO conversations (id, user_id, status, topic, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query,
		conv.ID,
		conv.UserID,
		conv.Status,
		conv.Topic,
		conv.CreatedAt,
		conv.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}

	return nil
}

// GetConversation retrieves a conversation by ID
func (r *SupportRepository) GetConversation(ctx context.Context, id, userID uuid.UUID) (*model.Conversation, error) {
	query := `
		SELECT id, user_id, status, topic, created_at, updated_at, closed_at
		FROM conversations
		WHERE id = $1 AND user_id = $2
	`
	var conv model.Conversation
	var closedAt pgtype.NullTime
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&conv.ID,
		&conv.UserID,
		&conv.Status,
		&conv.Topic,
		&conv.CreatedAt,
		&conv.UpdatedAt,
		&closedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrConversationNotFound
		}
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	if closedAt.Valid {
		conv.ClosedAt = &closedAt.Time
	}

	return &conv, nil
}

// GetConversations retrieves all conversations for a user
func (r *SupportRepository) GetConversations(ctx context.Context, userID uuid.UUID) ([]model.Conversation, error) {
	query := `
		SELECT id, user_id, status, topic, created_at, updated_at, closed_at
		FROM conversations
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}
	defer rows.Close()

	var conversations []model.Conversation
	for rows.Next() {
		var conv model.Conversation
		var closedAt pgtype.NullTime
		if err := rows.Scan(
			&conv.ID,
			&conv.UserID,
			&conv.Status,
			&conv.Topic,
			&conv.CreatedAt,
			&conv.UpdatedAt,
			&closedAt,
		); err != nil {
			return nil, err
		}
		if closedAt.Valid {
			conv.ClosedAt = &closedAt.Time
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// UpdateConversationStatus updates a conversation's status
func (r *SupportRepository) UpdateConversationStatus(ctx context.Context, id uuid.UUID, status string) error {
	var closedAt *time.Time
	if status == "closed" {
		now := time.Now()
		closedAt = &now
	}

	query := `
		UPDATE conversations
		SET status = $1, updated_at = $2, closed_at = $3
		WHERE id = $4
	`
	result, err := r.db.Exec(ctx, query, status, time.Now(), closedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update conversation status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrConversationNotFound
	}

	return nil
}

// CreateMessage creates a new message
func (r *SupportRepository) CreateMessage(ctx context.Context, msg *model.Message) error {
	msg.ID = uuid.Must(uuid.NewV7())
	msg.CreatedAt = time.Now()

	// Serialize metadata to JSON
	metadataBytes, err := json.Marshal(msg.Metadata)
	if err != nil {
		metadataBytes = []byte("{}")
	}

	query := `
		INSERT INTO messages (id, conversation_id, sender_type, content, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = r.db.Exec(ctx, query,
		msg.ID,
		msg.ConversationID,
		msg.SenderType,
		msg.Content,
		metadataBytes,
		msg.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// Update conversation's updated_at
	updateQuery := `
		UPDATE conversations
		SET updated_at = $1
		WHERE id = $2
	`
	_, err = r.db.Exec(ctx, updateQuery, time.Now(), msg.ConversationID)
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	return nil
}

// GetMessages retrieves all messages for a conversation
func (r *SupportRepository) GetMessages(ctx context.Context, conversationID uuid.UUID) ([]model.Message, error) {
	query := `
		SELECT id, conversation_id, sender_type, content, metadata, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var msg model.Message
		var metadataBytes []byte
		if err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderType,
			&msg.Content,
			&metadataBytes,
			&msg.CreatedAt,
		); err != nil {
			return nil, err
		}
		if len(metadataBytes) > 0 {
			msg.Metadata = metadataBytes
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// CreateEscalation creates a new escalation
func (r *SupportRepository) CreateEscalation(ctx context.Context, esc *model.Escalation) error {
	esc.ID = uuid.Must(uuid.NewV7())
	esc.CreatedAt = time.Now()
	esc.Status = "open"

	query := `
		INSERT INTO escalations (id, conversation_id, escalated_by, reason, admin_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		esc.ID,
		esc.ConversationID,
		esc.EscalatedBy,
		esc.Reason,
		esc.AdminID,
		esc.Status,
		esc.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create escalation: %w", err)
	}

	return nil
}

// GetEscalation retrieves an escalation by ID
func (r *SupportRepository) GetEscalation(ctx context.Context, id uuid.UUID) (*model.Escalation, error) {
	query := `
		SELECT id, conversation_id, escalated_by, reason, admin_id, status, notes, created_at, resolved_at
		FROM escalations
		WHERE id = $1
	`
	var esc model.Escalation
	var adminID, notes, resolvedAt pgtype.NullUUID
	var reason pgtype.NullText
	err := r.db.QueryRow(ctx, query, id).Scan(
		&esc.ID,
		&esc.ConversationID,
		&esc.EscalatedBy,
		&reason,
		&adminID,
		&esc.Status,
		&notes,
		&esc.CreatedAt,
		&resolvedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEscalationNotFound
		}
		return nil, fmt.Errorf("failed to get escalation: %w", err)
	}
	if reason.Valid {
		esc.Reason = reason.String
	}
	if adminID.Valid {
		esc.AdminID = &adminID.UUID
	}
	if notes.Valid {
		esc.Notes = notes.String
	}
	if resolvedAt.Valid {
		esc.ResolvedAt = &resolvedAt.Time
	}

	return &esc, nil
}

// GetEscalations retrieves escalations (admin only)
func (r *SupportRepository) GetEscalations(ctx context.Context, adminID *uuid.UUID, status *string) ([]model.Escalation, error) {
	query := `
		SELECT id, conversation_id, escalated_by, reason, admin_id, status, notes, created_at, resolved_at
		FROM escalations
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if adminID != nil {
		query += fmt.Sprintf(" AND admin_id = $%d", argNum)
		args = append(args, *adminID)
		argNum++
	}
	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, *status)
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get escalations: %w", err)
	}
	defer rows.Close()

	var escalations []model.Escalation
	for rows.Next() {
		var esc model.Escalation
		var adminID, notes pgtype.NullUUID
		var reason pgtype.NullText
		var resolvedAt pgtype.NullTime
		if err := rows.Scan(
			&esc.ID,
			&esc.ConversationID,
			&esc.EscalatedBy,
			&reason,
			&adminID,
			&esc.Status,
			&notes,
			&esc.CreatedAt,
			&resolvedAt,
		); err != nil {
			return nil, err
		}
		if reason.Valid {
			esc.Reason = reason.String
		}
		if adminID.Valid {
			esc.AdminID = &adminID.UUID
		}
		if notes.Valid {
			esc.Notes = notes.String
		}
		if resolvedAt.Valid {
			esc.ResolvedAt = &resolvedAt.Time
		}
		escalations = append(escalations, esc)
	}

	return escalations, nil
}

// UpdateEscalationStatus updates an escalation's status
func (r *SupportRepository) UpdateEscalationStatus(ctx context.Context, id uuid.UUID, status, notes string) error {
	var resolvedAt *time.Time
	if status == "resolved" {
		now := time.Now()
		resolvedAt = &now
	}

	query := `
		UPDATE escalations
		SET status = $1, notes = $2, resolved_at = $3
		WHERE id = $4
	`
	result, err := r.db.Exec(ctx, query, status, notes, resolvedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update escalation status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrEscalationNotFound
	}

	return nil
}