package handlers

import (
	"github.com/umar5678/go-backend/internal/websocket"
	"github.com/umar5678/go-backend/internal/websocket/websocketutils"
)

func RegisterAllHandlers(manager *websocket.Manager) {

	RegisterRideHandlers(manager)
	RegisterAdminSupportHandlers(manager)
}

// Register handlers for admin support chat and SOS location updates
func RegisterAdminSupportHandlers(manager *websocket.Manager) {
	manager.RegisterHandler("admin_support_chat", HandleAdminSupportChat)
	manager.RegisterHandler("sos_alert", HandleSOSAlert)
}

// Handler for dedicated admin support chat channel
func HandleAdminSupportChat(client *websocket.Client, msg *websocket.Message) error {
	senderID := client.UserID
	senderRole := string(client.Role)
	content, _ := msg.Data["content"].(string)
	metadata, _ := msg.Data["metadata"].(map[string]interface{})
	return websocketutils.SendAdminSupportChat(senderID, senderRole, content, metadata)
}

// Handler for SOS live location updates to admin
func HandleSOSAlert(client *websocket.Client, msg *websocket.Message) error {
	userID := client.UserID
	location, _ := msg.Data["location"].(map[string]interface{})
	sosActive, _ := msg.Data["sosActive"].(bool)
	return websocketutils.SendSOSLocationUpdate(userID, location, sosActive)
}
