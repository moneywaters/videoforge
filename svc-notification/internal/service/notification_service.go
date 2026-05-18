package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/pkg/natsclient"
	"github.com/videoforge/backend/svc-notification/internal/model"
	"github.com/videoforge/backend/svc-notification/internal/repository"
)

// NotificationService handles notification business logic
type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	preferenceRepo  *repository.UserPreferenceRepository
	connRepo       *repository.WSConnectionRepository
	nc             *natsclient.Client
	log            *logger.Logger
	connManager    *ConnectionManager
}

// NewNotificationService creates a new notification service
func NewNotificationService(pool *pgxpool.Pool, nc *natsclient.Client, log *logger.Logger) *NotificationService {
	notificationRepo := repository.NewNotificationRepository(pool, log)
	preferenceRepo := repository.NewUserPreferenceRepository(pool, log)
	connRepo := repository.NewWSConnectionRepository(pool, log)
	connManager := NewConnectionManager(log)

	return &NotificationService{
		notificationRepo: notificationRepo,
		preferenceRepo:  preferenceRepo,
		connRepo:       connRepo,
		nc:            nc,
		log:           log,
		connManager:    connManager,
	}
}

// StartEventConsumer starts consuming NATS events
func (s *NotificationService) StartEventConsumer(ctx context.Context) error {
	// Subscribe to all event subjects
	subjects := []string{
		"video.submitted",
		"video.approved",
		"video.rejected",
		"sale.attributed",
		"payout.released",
		"campaign.started",
		"campaign.ended",
	}

	for _, subject := range subjects {
		err := s.nc.Subscribe(subject, func(msg *nats.Msg) {
			s.handleEvent(msg)
		})
		if err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
		}
		s.log.Info("Subscribed to subject", slog.String("subject", subject))
	}

	return nil
}

// handleEvent handles incoming NATS events
func (s *NotificationService) handleEvent(msg *nats.Msg) {
	var eventPayload map[string]interface{}
	if err := json.Unmarshal(msg.Data, &eventPayload); err != nil {
		s.log.Error("Failed to unmarshal event payload", slog.String("error", err.Error()))
		return
	}

	subject := msg.Subject
	s.log.Info("Received event", slog.String("subject", subject), slog.String("payload", string(msg.Data)))

	// Determine recipients based on event type
	recipients := s.determineRecipients(subject, eventPayload)

	// Get enabled event type
	eventType := s.subjectToEventType(subject)

	for _, userID := range recipients {
		// Check user preferences
		pref, err := s.preferenceRepo.GetPreferences(context.Background(), userID)
		if err != nil {
			s.log.Error("Failed to get user preferences", slog.String("user_id", userID), slog.String("error", err.Error()))
			continue
		}

		// Check if event type is enabled
		if !s.isEventTypeEnabled(pref, eventType) {
			s.log.Debug("Event type not enabled for user", slog.String("user_id", userID), slog.String("type", string(eventType)))
			continue
		}

		// Create notification
		title, message := s.generateNotificationContent(eventType, eventPayload)
		input := model.CreateNotificationInput{
			UserID:  userID,
			Type:   eventType,
			Title:  title,
			Message: message,
			Data:   eventPayload,
		}

		notification, err := s.notificationRepo.CreateNotification(context.Background(), input)
		if err != nil {
			s.log.Error("Failed to create notification", slog.String("user_id", userID), slog.String("error", err.Error()))
			continue
		}

		// Send via WebSocket if preferred or as fallback
		wsPreferred := pref.ChannelPreference == model.ChannelWS || pref.ChannelPreference == model.ChannelBoth
		if wsPreferred {
			s.sendToUserViaWS(userID, notification)
		}

		s.log.Info("Notification created", slog.String("notification_id", notification.ID), slog.String("user_id", userID))
	}
}

// determineRecipients determines who should be notified for an event
func (s *NotificationService) determineRecipients(subject string, payload map[string]interface{}) []string {
	switch subject {
	case "video.submitted":
		// Notify brief owner (client)
		if briefOwner, ok := payload["brief_owner_id"].(string); ok {
			return []string{briefOwner}
		}
	case "video.approved", "video.rejected":
		// Notify editor
		if editorID, ok := payload["editor_id"].(string); ok {
			return []string{editorID}
		}
	case "sale.attributed":
		// Notify editor + ad_specialist + client
		var recipients []string
		if editorID, ok := payload["editor_id"].(string); ok {
			recipients = append(recipients, editorID)
		}
		if adSpecialistID, ok := payload["ad_specialist_id"].(string); ok {
			recipients = append(recipients, adSpecialistID)
		}
		if clientID, ok := payload["client_id"].(string); ok {
			recipients = append(recipients, clientID)
		}
		return recipients
	case "payout.released":
		// Notify recipient
		if recipientID, ok := payload["recipient_id"].(string); ok {
			return []string{recipientID}
		}
	case "campaign.started", "campaign.ended":
		// Notify client + ad_specialist
		var recipients []string
		if clientID, ok := payload["client_id"].(string); ok {
			recipients = append(recipients, clientID)
		}
		if adSpecialistID, ok := payload["ad_specialist_id"].(string); ok {
			recipients = append(recipients, adSpecialistID)
		}
		return recipients
	}
	return nil
}

// subjectToEventType converts NATS subject to notification type
func (s *NotificationService) subjectToEventType(subject string) model.NotificationType {
	switch subject {
	case "video.submitted":
		return model.NotificationVideoSubmitted
	case "video.approved":
		return model.NotificationVideoApproved
	case "video.rejected":
		return model.NotificationVideoRejected
	case "sale.attributed":
		return model.NotificationSaleAttributed
	case "payout.released":
		return model.NotificationPayoutReleased
	case "campaign.started":
		return model.NotificationCampaignStarted
	case "campaign.ended":
		return model.NotificationCampaignEnded
	default:
		return model.NotificationType(subject)
	}
}

// isEventTypeEnabled checks if an event type is enabled in preferences
func (s *NotificationService) isEventTypeEnabled(pref *model.UserPreference, eventType model.NotificationType) bool {
	if pref.EnabledTypes == nil {
		return true // Default to enabled
	}
	var enabledTypes []string
	_ = json.Unmarshal(pref.EnabledTypes, &enabledTypes)
	for _, t := range enabledTypes {
		if t == string(eventType) {
			return true
		}
	}
	return false
}

// generateNotificationContent generates title and message for a notification
func (s *NotificationService) generateNotificationContent(eventType model.NotificationType, payload map[string]interface{}) (title, message string) {
	switch eventType {
	case model.NotificationVideoSubmitted:
		title = "Video Submitted"
		message = fmt.Sprintf("A new video has been submitted for brief %s", payload["brief_id"])
	case model.NotificationVideoApproved:
		title = "Video Approved"
		message = fmt.Sprintf("Video %s has been approved", payload["video_id"])
	case model.NotificationVideoRejected:
		title = "Video Rejected"
		message = fmt.Sprintf("Video %s has been rejected", payload["video_id"])
	case model.NotificationSaleAttributed:
		title = "Sale Attributed"
		message = fmt.Sprintf("A new sale of $%v has been attributed to you", payload["amount"])
	case model.NotificationPayoutReleased:
		title = "Payout Released"
		message = fmt.Sprintf("Your payout of $%v has been released", payload["amount"])
	case model.NotificationCampaignStarted:
		title = "Campaign Started"
		message = fmt.Sprintf("Campaign %s has started", payload["campaign_id"])
	case model.NotificationCampaignEnded:
		title = "Campaign Ended"
		message = fmt.Sprintf("Campaign %s has ended", payload["campaign_id"])
	default:
		title = string(eventType)
		message = "You have a new notification"
	}
	return title, message
}

// sendToUserViaWS sends a notification to a user via WebSocket
func (s *NotificationService) sendToUserViaWS(userID string, notification *model.Notification) {
	s.connManager.SendToUser(userID, notification.ToResponse())
}

// ListNotifications lists notifications for a user
func (s *NotificationService) ListNotifications(ctx context.Context, userID string, filter model.ListNotificationsFilter) ([]model.NotificationResponse, int, error) {
	notifications, total, err := s.notificationRepo.ListNotificationsByUser(ctx, userID, filter)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.NotificationResponse, len(notifications))
	for i, n := range notifications {
		responses[i] = n.ToResponse()
	}
	return responses, total, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, userID, notificationID string) error {
	// Verify ownership
	notification, err := s.notificationRepo.GetNotificationByID(ctx, notificationID)
	if err != nil {
		return err
	}
	if notification == nil {
		return fmt.Errorf("notification not found")
	}
	if notification.UserID != userID {
		return fmt.Errorf("unauthorized")
	}
	return s.notificationRepo.MarkAsRead(ctx, notificationID)
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

// DeleteNotification deletes a notification
func (s *NotificationService) DeleteNotification(ctx context.Context, userID, notificationID string) error {
	// Verify ownership
	notification, err := s.notificationRepo.GetNotificationByID(ctx, notificationID)
	if err != nil {
		return err
	}
	if notification == nil {
		return fmt.Errorf("notification not found")
	}
	if notification.UserID != userID {
		return fmt.Errorf("unauthorized")
	}
	return s.notificationRepo.DeleteNotification(ctx, notificationID)
}

// GetPreferences retrieves user preferences
func (s *NotificationService) GetPreferences(ctx context.Context, userID string) (*model.UserPreferenceResponse, error) {
	pref, err := s.preferenceRepo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := pref.ToResponse()
	return &resp, nil
}

// UpdatePreferences updates user preferences
func (s *NotificationService) UpdatePreferences(ctx context.Context, userID string, input model.UpdatePreferenceInput) (*model.UserPreferenceResponse, error) {
	pref, err := s.preferenceRepo.UpdatePreferences(ctx, userID, input)
	if err != nil {
		return nil, err
	}
	resp := pref.ToResponse()
	return &resp, nil
}

// GetConnectionManager returns the connection manager
func (s *NotificationService) GetConnectionManager() *ConnectionManager {
	return s.connManager
}

// ConnectionManager manages WebSocket connections
type ConnectionManager struct {
	mu         sync.RWMutex
	connections map[string]map[string]*websocket.Conn // userID -> connectionID -> conn
	log        *logger.Logger
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(log *logger.Logger) *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]map[string]*websocket.Conn),
		log:        log,
	}
}

// AddConnection adds a WebSocket connection for a user
func (m *ConnectionManager) AddConnection(userID, connID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connections[userID] == nil {
		m.connections[userID] = make(map[string]*websocket.Conn)
	}
	m.connections[userID][connID] = conn

	m.log.Info("WebSocket connection added",
		slog.String("user_id", userID),
		slog.String("connection_id", connID),
	)
}

// RemoveConnection removes a WebSocket connection
func (m *ConnectionManager) RemoveConnection(userID, connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connections[userID] != nil {
		delete(m.connections[userID], connID)
		if len(m.connections[userID]) == 0 {
			delete(m.connections, userID)
		}
	}

	m.log.Info("WebSocket connection removed",
		slog.String("user_id", userID),
		slog.String("connection_id", connID),
	)
}

// SendToUser sends a message to all connections for a user
func (m *ConnectionManager) SendToUser(userID string, message interface{}) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	connections, ok := m.connections[userID]
	if !ok {
		return
	}

	for connID, conn := range connections {
		err := conn.WriteJSON(message)
		if err != nil {
			m.log.Error("Failed to send message",
				slog.String("connection_id", connID),
				slog.String("error", err.Error()),
			)
		}
	}
}

// SendToConnection sends a message to a specific connection
func (m *ConnectionManager) SendToConnection(userID, connID string, message interface{}) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.connections[userID] == nil {
		return false
	}

	conn, ok := m.connections[userID][connID]
	if !ok {
		return false
	}

	err := conn.WriteJSON(message)
	if err != nil {
m.log.Error("Failed to send message",
		slog.String("connection_id", connID),
		slog.String("error", err.Error()),
	)
		return false
	}
	return true
}

// GetConnections returns all connection IDs for a user
func (m *ConnectionManager) GetConnections(userID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.connections[userID] == nil {
		return nil
	}

	connIDs := make([]string, 0, len(m.connections[userID]))
	for connID := range m.connections[userID] {
		connIDs = append(connIDs, connID)
	}
	return connIDs
}

// IsUserConnected checks if a user has any active connections
func (m *ConnectionManager) IsUserConnected(userID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.connections[userID] != nil && len(m.connections[userID]) > 0
}