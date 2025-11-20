// internal/utils/websocketutil/websocketutil.go
package websocketutils

import (
	"context"

	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/websocket"
)

var (
	wsManager *websocket.Manager
)

// Initialize sets up the WebSocket utility with the manager
func Initialize(manager *websocket.Manager) {
	wsManager = manager
	logger.Info("websocket utility initialized")
}

// SendToUser sends a message to a specific user
func SendToUser(userID string, messageType websocket.MessageType, data map[string]interface{}) error {
	if wsManager == nil {
		logger.Warn("websocket manager not initialized")
		return nil
	}

	if data == nil {
		data = make(map[string]interface{})
	}

	msg := websocket.NewTargetedMessage(messageType, userID, data)
	wsManager.Hub().SendToUser(userID, msg)

	logger.Debug("websocket message sent to user",
		"userID", userID,
		"type", messageType,
	)
	return nil
}

// BroadcastToAll sends a message to all connected users
func BroadcastToAll(messageType websocket.MessageType, data map[string]interface{}) error {
	if wsManager == nil {
		logger.Warn("websocket manager not initialized")
		return nil
	}

	if data == nil {
		data = make(map[string]interface{})
	}

	msg := websocket.NewMessage(messageType, data)
	wsManager.Hub().BroadcastToAll(msg)

	logger.Debug("websocket message broadcasted",
		"type", messageType,
	)
	return nil
}

// SendNotification sends a notification to a user
func SendNotification(userID string, notification interface{}) error {
	if wsManager == nil {
		return nil
	}
	return wsManager.SendNotification(userID, notification)
}

// IsUserOnline checks if a user is currently connected
func IsUserOnline(userID string) bool {
	if wsManager == nil {
		return false
	}
	return wsManager.Hub().IsUserConnected(userID)
}

// GetOnlineUsers returns statistics about connected users
func GetOnlineUsers() (int, int) {
	if wsManager == nil {
		return 0, 0
	}
	stats := wsManager.GetStats()
	return stats.ConnectedUsers, stats.TotalConnections
}

// Ride-specific helper functions

// SendRideRequest sends ride request to driver
func SendRideRequest(driverID string, rideDetails map[string]interface{}) error {
	return SendToUser(driverID, websocket.TypeRideRequest, rideDetails)
}

// SendRideAccepted notifies rider that driver accepted
func SendRideAccepted(riderID string, rideDetails map[string]interface{}) error {
	return SendToUser(riderID, websocket.TypeRideAccepted, rideDetails)
}

// SendRideLocationUpdate sends location update to rider
func SendRideLocationUpdate(riderID string, locationData map[string]interface{}) error {
	return SendToUser(riderID, websocket.TypeRideLocation, locationData)
}

// SendRideStatusUpdate sends status update to both rider and driver
func SendRideStatusUpdate(riderID, driverID string, statusData map[string]interface{}) error {
	// Send to rider
	if riderID != "" {
		SendToUser(riderID, websocket.TypeRideStatusUpdate, statusData)
	}

	// Send to driver
	if driverID != "" {
		SendToUser(driverID, websocket.TypeRideStatusUpdate, statusData)
	}

	return nil
}

// SendPaymentUpdate sends payment status update
func SendPaymentUpdate(userID string, paymentData map[string]interface{}) error {
	return SendToUser(userID, websocket.TypePaymentCompleted, paymentData)
}

// SendSystemMessage sends system message to user
func SendSystemMessage(userID, message string, data map[string]interface{}) error {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["message"] = message
	return SendToUser(userID, websocket.TypeSystemMessage, data)
}

// Context-aware functions (for when you need context)

// SendToUserWithContext sends message with context (for future use)
func SendToUserWithContext(ctx context.Context, userID string, messageType websocket.MessageType, data map[string]interface{}) error {
	// Currently just calls the regular function, but structure allows for future context usage
	return SendToUser(userID, messageType, data)
}

// SendRideLocationUpdateWithContext sends location update with context
func SendRideLocationUpdateWithContext(ctx context.Context, riderID string, locationData map[string]interface{}) error {
	return SendRideLocationUpdate(riderID, locationData)
}
