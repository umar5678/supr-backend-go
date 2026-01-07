// internal/websocket/websocketutils/helpers.go
package websocketutils

import (
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/websocket"
)

// // SendToUser sends a message to a specific user by their user ID
// func SendToUser(userID string, messageType string, payload interface{}) {
// 	hub := websocket.GetHub()
// 	if hub == nil {
// 		logger.Error("websocket hub not initialized")
// 		return
// 	}

// 	hub.SendToUser(userID, messageType, payload)
// }

// BroadcastToRole sends a message to all users with a specific role
func BroadcastToRole(role string, messageType websocket.MessageType, payload interface{}) error {
	if wsManager == nil {
		logger.Warn("websocket manager not initialized")
		return nil
	}

	if data, ok := payload.(map[string]interface{}); ok {
		msg := websocket.NewMessage(messageType, data)
		wsManager.Hub().BroadcastToRole(role, msg)
		return nil
	}

	logger.Error("invalid payload type for BroadcastToRole")
	return nil
}


