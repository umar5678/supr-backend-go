// internal/websocket/websocketutils/helpers.go
package websocketutils

import (
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/websocket"
)

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


