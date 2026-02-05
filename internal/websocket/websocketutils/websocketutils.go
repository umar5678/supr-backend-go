// internal/utils/websocketutil/websocketutil.go
package websocketutils

import (
	"context"
	"errors"

	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/websocket"
)

var (
	wsManager *websocket.Manager
)

func Initialize(manager *websocket.Manager) {
	wsManager = manager
	logger.Info("websocket utility initialized")
}

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

func SendNotification(userID string, notification interface{}) error {
	if wsManager == nil {
		return nil
	}
	return wsManager.SendNotification(userID, notification)
}

func IsUserOnline(userID string) bool {
	if wsManager == nil {
		return false
	}
	return wsManager.Hub().IsUserConnected(userID)
}

func GetOnlineUsers() (int, int) {
	if wsManager == nil {
		return 0, 0
	}
	stats := wsManager.GetStats()
	return stats.ConnectedUsers, stats.TotalConnections
}

func SendRideLocationUpdate(riderID string, locationData map[string]interface{}) error {
	logger.Info("SendRideLocationUpdate CALLED",
		"riderUserID", riderID,
		"hasLocationData", locationData != nil,
	)

	if riderID == "" {
		logger.Error("SendRideLocationUpdate: empty riderID")
		return errors.New("empty riderID")
	}

	err := SendToUser(riderID, websocket.TypeRideLocation, locationData)

	if err != nil {
		logger.Error("SendRideLocationUpdate FAILED", "error", err, "riderID", riderID)
	} else {
		logger.Info("SendRideLocationUpdate SUCCESS", "riderID", riderID)
	}

	return err
}

func SendToUser(userID string, messageType websocket.MessageType, data map[string]interface{}) error {
	logger.Info("SendToUser CALLED",
		"userID", userID,
		"messageType", messageType,
		"wsManagerInitialized", wsManager != nil,
	)

	if wsManager == nil {
		logger.Error("websocket manager NOT INITIALIZED")
		return errors.New("websocket manager not initialized")
	}

	if data == nil {
		data = make(map[string]interface{})
	}

	msg := websocket.NewTargetedMessage(messageType, userID, data)

	logger.Info("Calling Hub().SendToUser",
		"userID", userID,
		"messageType", messageType,
		"msgData", msg,
	)

	wsManager.Hub().SendToUser(userID, msg)

	logger.Info("websocket message sent to user",
		"userID", userID,
		"type", messageType,
	)
	return nil
}

func SendRideRequest(driverID string, rideDetails map[string]interface{}) error {
	return SendToUser(driverID, websocket.TypeRideRequest, rideDetails)
}

func SendRideAccepted(riderID string, rideDetails map[string]interface{}) error {
	return SendToUser(riderID, websocket.TypeRideAccepted, rideDetails)
}

func SendRideStatusUpdate(riderID, driverID string, statusData map[string]interface{}) error {
	if riderID != "" {
		SendToUser(riderID, websocket.TypeRideStatusUpdate, statusData)
	}

	if driverID != "" {
		SendToUser(driverID, websocket.TypeRideStatusUpdate, statusData)
	}

	return nil
}

func SendPaymentUpdate(userID string, paymentData map[string]interface{}) error {
	return SendToUser(userID, websocket.TypePaymentCompleted, paymentData)
}

func SendSystemMessage(userID, message string, data map[string]interface{}) error {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["message"] = message
	return SendToUser(userID, websocket.TypeSystemMessage, data)
}

func SendToUserWithContext(ctx context.Context, userID string, messageType websocket.MessageType, data map[string]interface{}) error {
	return SendToUser(userID, messageType, data)
}

func SendRideLocationUpdateWithContext(ctx context.Context, riderID string, locationData map[string]interface{}) error {
	return SendRideLocationUpdate(riderID, locationData)
}
