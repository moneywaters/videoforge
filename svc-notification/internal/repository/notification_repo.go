package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/svc-notification/internal/model"
)

// NotificationRepository handles notification database operations
type NotificationRepository struct {
	pool *pgxpool.Pool
	log  *logger.Logger
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(pool *pgxpool.Pool, log *logger.Logger) *NotificationRepository {
	return &NotificationRepository{
		pool: pool,
		log:  log,
	}
}

// CreateNotification creates a new notification
func (r *NotificationRepository) CreateNotification(ctx context.Context, input model.CreateNotificationInput) (*model.Notification, error) {
	var data []byte
	if input.Data != nil {
		var err error
		data, err = json.Marshal(input.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}
	}

	var notification model.Notification
	err := r.pool.QueryRow(ctx, `
		INSERT INTO notifications (user_id, type, title, message, data)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, type, title, message, data, read, read_at, created_at
	`, input.UserID, input.Type, input.Title, input.Message, data).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&notification.Data,
		&notification.Read,
		&notification.ReadAt,
		&notification.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}
	return &notification, nil
}

// GetNotificationByID retrieves a notification by ID
func (r *NotificationRepository) GetNotificationByID(ctx context.Context, id string) (*model.Notification, error) {
	var notification model.Notification
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, type, title, message, data, read, read_at, created_at
		FROM notifications
		WHERE id = $1
	`, id).Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&notification.Data,
		&notification.Read,
		&notification.ReadAt,
		&notification.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}
	return &notification, nil
}

// ListNotificationsByUser lists notifications for a user with pagination
func (r *NotificationRepository) ListNotificationsByUser(ctx context.Context, userID string, filter model.ListNotificationsFilter) ([]model.Notification, int, error) {
	filter.Default()

	// Build query
	baseQuery := `FROM notifications WHERE user_id = $1`
	args := []interface{}{userID}

	if filter.Read != nil {
		baseQuery += fmt.Sprintf(" AND read = $%d", len(args)+1)
		args = append(args, *filter.Read)
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf("SELECT id, user_id, type, title, message, data, read, read_at, created_at %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d",
		baseQuery, len(args)+1, len(args)+2)
	args = append(args, filter.PerPage, filter.Offset())

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	var notifications []model.Notification
	for rows.Next() {
		var n model.Notification
		err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Type,
			&n.Title,
			&n.Message,
			&n.Data,
			&n.Read,
			&n.ReadAt,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	if notifications == nil {
		notifications = []model.Notification{}
	}

	return notifications, total, nil
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE notifications
		SET read = true, read_at = $2
		WHERE id = $1 AND read = false
	`, id, time.Now())
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE notifications
		SET read = true, read_at = $2
		WHERE user_id = $1 AND read = false
	`, userID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}

// DeleteNotification deletes a notification
func (r *NotificationRepository) DeleteNotification(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM notifications WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}
	return nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read = false
	`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}

// UserPreferenceRepository handles user preference database operations
type UserPreferenceRepository struct {
	pool *pgxpool.Pool
	log  *logger.Logger
}

// NewUserPreferenceRepository creates a new user preference repository
func NewUserPreferenceRepository(pool *pgxpool.Pool, log *logger.Logger) *UserPreferenceRepository {
	return &UserPreferenceRepository{
		pool: pool,
		log:  log,
	}
}

// GetPreferences retrieves user preferences
func (r *UserPreferenceRepository) GetPreferences(ctx context.Context, userID string) (*model.UserPreference, error) {
	var pref model.UserPreference
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, channel_preference, enabled_types, created_at, updated_at
		FROM user_preferences
		WHERE user_id = $1
	`, userID).Scan(
		&pref.ID,
		&pref.UserID,
		&pref.ChannelPreference,
		&pref.EnabledTypes,
		&pref.CreatedAt,
		&pref.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Create default preferences
			return r.CreatePreferences(ctx, userID)
		}
		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}
	return &pref, nil
}

// CreatePreferences creates default preferences for a user
func (r *UserPreferenceRepository) CreatePreferences(ctx context.Context, userID string) (*model.UserPreference, error) {
	enabledTypes, _ := json.Marshal([]string{
		string(model.NotificationVideoSubmitted),
		string(model.NotificationVideoApproved),
		string(model.NotificationVideoRejected),
		string(model.NotificationSaleAttributed),
		string(model.NotificationPayoutReleased),
		string(model.NotificationCampaignStarted),
		string(model.NotificationCampaignEnded),
	})

	var pref model.UserPreference
	err := r.pool.QueryRow(ctx, `
		INSERT INTO user_preferences (user_id, channel_preference, enabled_types)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, channel_preference, enabled_types, created_at, updated_at
	`, userID, model.ChannelWS, enabledTypes).Scan(
		&pref.ID,
		&pref.UserID,
		&pref.ChannelPreference,
		&pref.EnabledTypes,
		&pref.CreatedAt,
		&pref.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create preferences: %w", err)
	}
	return &pref, nil
}

// UpdatePreferences updates user preferences
func (r *UserPreferenceRepository) UpdatePreferences(ctx context.Context, userID string, input model.UpdatePreferenceInput) (*model.UserPreference, error) {
	enabledTypes, err := json.Marshal(input.EnabledTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal enabled types: %w", err)
	}

	var pref model.UserPreference
	err = r.pool.QueryRow(ctx, `
		UPDATE user_preferences
		SET channel_preference = $2, enabled_types = $3, updated_at = NOW()
		WHERE user_id = $1
		RETURNING id, user_id, channel_preference, enabled_types, created_at, updated_at
	`, userID, input.ChannelPreference, enabledTypes).Scan(
		&pref.ID,
		&pref.UserID,
		&pref.ChannelPreference,
		&pref.EnabledTypes,
		&pref.CreatedAt,
		&pref.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Create preferences if not exists
			return r.CreatePreferences(ctx, userID)
		}
		return nil, fmt.Errorf("failed to update preferences: %w", err)
	}
	return &pref, nil
}

// WSConnectionRepository handles WebSocket connection database operations
type WSConnectionRepository struct {
	pool *pgxpool.Pool
	log  *logger.Logger
}

// NewWSConnectionRepository creates a new WS connection repository
func NewWSConnectionRepository(pool *pgxpool.Pool, log *logger.Logger) *WSConnectionRepository {
	return &WSConnectionRepository{
		pool: pool,
		log:  log,
	}
}

// CreateConnection creates a new WebSocket connection
func (r *WSConnectionRepository) CreateConnection(ctx context.Context, userID, connectionID string) (*model.WSConnection, error) {
	var conn model.WSConnection
	err := r.pool.QueryRow(ctx, `
		INSERT INTO ws_connections (user_id, connection_id)
		VALUES ($1, $2)
		RETURNING id, user_id, connection_id, connected_at, last_ping_at
	`, userID, connectionID).Scan(
		&conn.ID,
		&conn.UserID,
		&conn.ConnectionID,
		&conn.ConnectedAt,
		&conn.LastPingAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}
	return &conn, nil
}

// GetConnectionByID retrieves a connection by ID
func (r *WSConnectionRepository) GetConnectionByID(ctx context.Context, connectionID string) (*model.WSConnection, error) {
	var conn model.WSConnection
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, connection_id, connected_at, last_ping_at
		FROM ws_connections
		WHERE connection_id = $1
	`, connectionID).Scan(
		&conn.ID,
		&conn.UserID,
		&conn.ConnectionID,
		&conn.ConnectedAt,
		&conn.LastPingAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}
	return &conn, nil
}

// GetConnectionsByUser retrieves all connections for a user
func (r *WSConnectionRepository) GetConnectionsByUser(ctx context.Context, userID string) ([]model.WSConnection, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, connection_id, connected_at, last_ping_at
		FROM ws_connections
		WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w", err)
	}
	defer rows.Close()

	var connections []model.WSConnection
	for rows.Next() {
		var conn model.WSConnection
		err := rows.Scan(
			&conn.ID,
			&conn.UserID,
			&conn.ConnectionID,
			&conn.ConnectedAt,
			&conn.LastPingAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection: %w", err)
		}
		connections = append(connections, conn)
	}

	if connections == nil {
		connections = []model.WSConnection{}
	}
	return connections, nil
}

// DeleteConnection deletes a WebSocket connection
func (r *WSConnectionRepository) DeleteConnection(ctx context.Context, connectionID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM ws_connections WHERE connection_id = $1`, connectionID)
	if err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}
	return nil
}

// UpdateLastPing updates the last ping time for a connection
func (r *WSConnectionRepository) UpdateLastPing(ctx context.Context, connectionID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE ws_connections SET last_ping_at = NOW() WHERE connection_id = $1
	`, connectionID)
	if err != nil {
		return fmt.Errorf("failed to update last ping: %w", err)
	}
	return nil
}

// IsUserConnected checks if a user is connected via WebSocket
func (r *WSConnectionRepository) IsUserConnected(ctx context.Context, userID string) (bool, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM ws_connections WHERE user_id = $1
	`, userID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check user connection: %w", err)
	}
	return count > 0, nil
}