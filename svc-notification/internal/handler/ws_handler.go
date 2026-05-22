package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
	"github.com/videoforge/backend/pkg/logger"
	"github.com/videoforge/backend/svc-notification/internal/model"
	"github.com/videoforge/backend/svc-notification/internal/repository"
	"github.com/videoforge/backend/svc-notification/internal/service"
)

// Constants for ping/pong
const (
	pingInterval  = 30 * time.Second
	pingWriteWait = 10 * time.Second
	pingReadWait  = 35 * time.Second
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	upgrader          websocket.Upgrader
	connManager       *service.ConnectionManager
	connRepo          *repository.WSConnectionRepository
	notificationRepo *repository.NotificationRepository
	log               *logger.Logger
}

// NewWebSocketHandler creates a new WebSocketHandler
func NewWebSocketHandler(pool *pgxpool.Pool, connManager *service.ConnectionManager, log *logger.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		connManager:       connManager,
		connRepo:          repository.NewWSConnectionRepository(pool, log),
		notificationRepo: repository.NewNotificationRepository(pool, log),
		log:               log,
	}
}

// HandleWebSocket upgrades HTTP connection to WebSocket and handles it
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Authenticate via JWT token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		// Try Authorization header
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if token == "" {
		http.Error(w, "Missing authentication token", http.StatusUnauthorized)
		return
	}

	// Parse JWT token to get userID
	userID, err := h.parseJWT(token)
	if err != nil {
		h.log.Error("Failed to parse JWT", slog.String("error", err.Error()))
		http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
		return
	}

	// Generate connection ID
	connectionID := uuid.New().String()

	// Upgrade connection
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("Failed to upgrade WebSocket", slog.String("error", err.Error()))
		return
	}
	defer func() {
		conn.Close()
		// Remove from manager and database
		h.connManager.RemoveConnection(userID, connectionID)
		h.connRepo.DeleteConnection(context.Background(), connectionID)
	}()

	// Register connection in database
	_, err = h.connRepo.CreateConnection(context.Background(), userID, connectionID)
	if err != nil {
		h.log.Error("Failed to create connection in database", slog.String("error", err.Error()))
		conn.WriteMessage(websocket.CloseMessage, []byte(`{"type":"error","message":"Failed to register connection"}`))
		return
	}

	// Add to connection manager
	h.connManager.AddConnection(userID, connectionID, conn)

	// Send welcome message with connection info
	welcomeMsg := map[string]interface{}{
		"event":         "connected",
		"connection_id": connectionID,
		"message":       "WebSocket connection established",
	}
	if err := conn.WriteJSON(welcomeMsg); err != nil {
		h.log.Error("Failed to send welcome message", slog.String("error", err.Error()))
		return
	}

	// Send unread notifications
	unreads, err := h.getUnreadNotifications(context.Background(), userID)
	if err != nil {
		h.log.Error("Failed to get unread notifications", slog.String("error", err.Error()))
	} else if len(unreads) > 0 {
		for _, n := range unreads {
			if err := conn.WriteJSON(n); err != nil {
				break
			}
		}
	}

	// Start ping/pong goroutine
	pingChan := make(chan struct{})
	go h.writePingLoop(conn, connectionID, pingChan)
	go h.readPongLoop(conn, connectionID, pingChan)

	// Handle incoming messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.log.Debug("WebSocket connection closed", slog.String("connection_id", connectionID))
			}
			break
		}

		if messageType == websocket.TextMessage {
			h.handleClientMessage(userID, connectionID, message)
		}
	}

	// Cleanup
	close(pingChan)
}

// writePingLoop sends periodic pings to the client
func (h *WebSocketHandler) writePingLoop(conn *websocket.Conn, connectionID string, pingChan chan struct{}) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(pingWriteWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				h.log.Error("Failed to send ping", slog.String("error", err.Error()), slog.String("connection_id", connectionID))
				return
			}
		case <-pingChan:
			return
		}
	}
}

// readPongLoop waits for pong responses and updates last ping
func (h *WebSocketHandler) readPongLoop(conn *websocket.Conn, connectionID string, pingChan chan struct{}) {
	for {
		conn.SetReadDeadline(time.Now().Add(pingReadWait))
		_, _, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Connection might be closed or stale
			}
			pingChan <- struct{}{}
			return
		}
		// Update last ping in database
		h.connRepo.UpdateLastPing(context.Background(), connectionID)
	}
}

// handleClientMessage handles messages from the client
func (h *WebSocketHandler) handleClientMessage(userID, connectionID string, message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "ping":
		// Respond with pong
		h.connManager.SendToConnection(userID, connectionID, map[string]string{
			"type":    "pong",
			"status": "ok",
		})
	case "subscribe":
		// Handle subscription to notification types (future feature)
	}
}

// parseJWT parses a JWT token and returns the userID
func (h *WebSocketHandler) parseJWT(token string) (string, error) {
	// For now, implement basic JWT parsing
	// In production, use proper JWT validation with the secret from config

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format")
	}

	// Decode payload (second part)
	payload, err := jwtDecodeBase64Url(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode token: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("failed to parse claims: %w", err)
	}

	// Extract userID from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		// Try alternate claim names
		if uid, ok := claims["user_id"].(string); ok {
			return uid, nil
		}
		return "", fmt.Errorf("user ID not found in token")
	}

	return userID, nil
}

// jwtDecodeBase64Url decodes a base64url encoded string
func jwtDecodeBase64Url(input string) ([]byte, error) {
	// Replace base64url specific characters with standard base64
	output := strings.ReplaceAll(input, "-", "+")
	output = strings.ReplaceAll(output, "_", "/")

	// Add padding if needed
	switch len(output) % 4 {
	case 2:
		output += "=="
	case 3:
		output += "="
	}

	// Decode from base64
	decoded, err := base64.StdEncoding.DecodeString(output)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}

// getUnreadNotifications retrieves unread notifications for a user
func (h *WebSocketHandler) getUnreadNotifications(ctx context.Context, userID string) ([]model.NotificationResponse, error) {
	filter := model.ListNotificationsFilter{
		PaginationParams: model.PaginationParams{
			Page:    1,
			PerPage: 50,
		},
	}
	read := false
	filter.Read = &read

	notifications, _, err := h.notificationRepo.ListNotificationsByUser(ctx, userID, filter)
	if err != nil {
		return nil, err
	}

	responses := make([]model.NotificationResponse, len(notifications))
	for i, n := range notifications {
		responses[i] = n.ToResponse()
	}
	return responses, nil
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Event    string          `json:"event"`
	Type    string          `json:"type,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}