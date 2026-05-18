package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Upgrader configures WebSocket upgrades
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// WSHandler handles WebSocket connections
type WSHandler struct {
	 upgrader *websocket.Upgrader
}

// NewWSHandler creates a new WSHandler
func NewWSHandler() *WSHandler {
	return &WSHandler{
		upgrader: &upgrader,
	}
}

// HandleWS upgrades the connection to WebSocket and handles echo communication
// TODO: Replace with actual WebSocket handling via NATS for real-time features
func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	// Upgrade to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket client connected from %s", r.RemoteAddr)

	// Handle incoming messages - echo back for now
	// TODO: Route messages to appropriate internal services via NATS
	for {
		// Read message from client
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Echo message back
		// TODO: Process message and route to internal services
		if msgType == websocket.TextMessage {
			// For demo, echo back with timestamp
			response := map[string]interface{}{
				"type":    "echo",
				"message": string(msg),
			}
			respBytes, _ := json.Marshal(response)
			err := conn.WriteMessage(websocket.TextMessage, respBytes)
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				break
			}
		}
	}

	log.Printf("WebSocket client disconnected from %s", r.RemoteAddr)
}