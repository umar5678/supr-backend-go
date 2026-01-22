// internal/websocket/handlers/handlers.go
package handlers

import (
	"github.com/umar5678/go-backend/internal/websocket"
)

// RegisterAllHandlers registers all WebSocket handlers
// Note: Message handlers must be registered separately after messages service is initialized
func RegisterAllHandlers(manager *websocket.Manager) {
	// Register Ride Handlers
	RegisterRideHandlers(manager)

	// Add others as you build them:
	// RegisterChatHandlers(manager)
	// RegisterNotificationHandlers(manager)
}
