package handlers

import (
	"encoding/json"
	"time"

	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/websocket"
)

func RegisterRideHandlers(manager *websocket.Manager) {
	manager.RegisterHandler(websocket.TypeRideLocation, handleRideLocationUpdate)
	manager.RegisterHandler(websocket.TypeRideStatusUpdate, handleRideStatusUpdate)
	manager.RegisterHandler("driver_availability", handleDriverAvailability)
	manager.RegisterHandler("ride_accept", handleRideAccept)
	manager.RegisterHandler("ride_cancel", handleRideCancel)
}

func handleRideLocationUpdate(client *websocket.Client, msg *websocket.Message) error {
	rideID, ok := msg.Data["rideId"].(string)
	if !ok {
		return client.SendError("rideId required", msg.RequestID)
	}

	location, ok := msg.Data["location"].(map[string]interface{})
	if !ok {
		return client.SendError("location data required", msg.RequestID)
	}

	if client.Manager() == nil {
		return client.SendError("manager not available", msg.RequestID)
	}

	hub := client.Manager().Hub()

	broadcastMsg := websocket.NewMessage(websocket.TypeRideLocation, map[string]interface{}{
		"rideId":    rideID,
		"location":  location,
		"driverId":  client.UserID,
		"timestamp": time.Now().UTC(),
	})

	hub.BroadcastToAll(broadcastMsg)

	logger.Info("ride location update broadcasted",
		"rideID", rideID,
		"driverID", client.UserID,
	)

	return client.SendAck(msg.RequestID, map[string]interface{}{"success": true})
}

func handleRideStatusUpdate(client *websocket.Client, msg *websocket.Message) error {
	rideID, ok := msg.Data["rideId"].(string)
	if !ok {
		return client.SendError("rideId required", msg.RequestID)
	}

	status, ok := msg.Data["status"].(string)
	if !ok {
		return client.SendError("status required", msg.RequestID)
	}

	if client.Manager() == nil {
		return client.SendError("manager not available", msg.RequestID)
	}

	hub := client.Manager().Hub()

	broadcastMsg := websocket.NewMessage(websocket.TypeRideStatusUpdate, map[string]interface{}{
		"rideId":    rideID,
		"status":    status,
		"userId":    client.UserID,
		"timestamp": time.Now().UTC(),
	})

	hub.BroadcastToAll(broadcastMsg)

	logger.Info("ride status update broadcasted",
		"rideID", rideID,
		"status", status,
		"userID", client.UserID,
	)

	return client.SendAck(msg.RequestID, map[string]interface{}{"success": true})
}

func handleDriverAvailability(client *websocket.Client, msg *websocket.Message) error {
	isAvailable, ok := msg.Data["isAvailable"].(bool)
	if !ok {
		return client.SendError("isAvailable required", msg.RequestID)
	}

	location, _ := msg.Data["location"].(map[string]interface{})

	logger.Info("driver availability updated",
		"driverID", client.UserID,
		"available", isAvailable,
		"location", location,
	)

	return client.SendAck(msg.RequestID, map[string]interface{}{
		"success":     true,
		"isAvailable": isAvailable,
	})
}

func handleRideAccept(client *websocket.Client, msg *websocket.Message) error {
	rideID, ok := msg.Data["rideId"].(string)
	if !ok {
		return client.SendError("rideId required", msg.RequestID)
	}

	riderID, ok := msg.Data["riderId"].(string)
	if !ok {
		return client.SendError("riderId required", msg.RequestID)
	}

	if client.Manager() == nil {
		return client.SendError("manager not available", msg.RequestID)
	}

	hub := client.Manager().Hub()

	acceptMsg := websocket.NewTargetedMessage(websocket.TypeRideAccepted, riderID, map[string]interface{}{
		"rideId":    rideID,
		"driverId":  client.UserID,
		"timestamp": time.Now().UTC(),
	})

	hub.SendToUser(riderID, acceptMsg)

	logger.Info("ride accepted",
		"rideID", rideID,
		"driverID", client.UserID,
		"riderID", riderID,
	)

	return client.SendAck(msg.RequestID, map[string]interface{}{"success": true})
}

func handleRideCancel(client *websocket.Client, msg *websocket.Message) error {
	rideID, ok := msg.Data["rideId"].(string)
	if !ok {
		return client.SendError("rideId required", msg.RequestID)
	}

	reason, _ := msg.Data["reason"].(string)
	targetUserID, _ := msg.Data["targetUserId"].(string)

	if client.Manager() == nil {
		return client.SendError("manager not available", msg.RequestID)
	}

	hub := client.Manager().Hub()

	cancelMsg := websocket.NewMessage(websocket.TypeRideCancelled, map[string]interface{}{
		"rideId":    rideID,
		"reason":    reason,
		"userId":    client.UserID,
		"timestamp": time.Now().UTC(),
	})

	if targetUserID != "" {
		cancelMsg.TargetUserID = targetUserID
		hub.SendToUser(targetUserID, cancelMsg)
	} else {
		hub.BroadcastToAll(cancelMsg)
	}

	logger.Info("ride cancelled",
		"rideID", rideID,
		"userID", client.UserID,
		"reason", reason,
		"targetUserID", targetUserID,
	)

	return client.SendAck(msg.RequestID, map[string]interface{}{"success": true})
}

func extractUserIDs(data map[string]interface{}) ([]string, error) {
	userIDsInterface, ok := data["userIds"].([]interface{})
	if !ok {
		if userIDsStr, ok := data["userIds"].(string); ok {
			var userIDs []string
			if err := json.Unmarshal([]byte(userIDsStr), &userIDs); err != nil {
				return nil, err
			}
			return userIDs, nil
		}
		return nil, json.Unmarshal([]byte(data["userIds"].(string)), &userIDsInterface)
	}

	userIDs := make([]string, len(userIDsInterface))
	for i, id := range userIDsInterface {
		if str, ok := id.(string); ok {
			userIDs[i] = str
		}
	}

	return userIDs, nil
}
