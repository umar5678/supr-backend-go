// internal/websocket/handlers/handlers.go
package handlers

import (
	"github.com/umar5678/go-backend/internal/websocket"
)

// RegisterAllHandlers registers all WebSocket handlers
func RegisterAllHandlers(manager *websocket.Manager) {
	RegisterRideHandlers(manager)
	// Register other handler groups here
	// RegisterNotificationHandlers(manager)
	// RegisterChatHandlers(manager)
}
