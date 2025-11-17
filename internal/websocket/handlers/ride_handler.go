// internal/websocket/handlers/ride_handlers.go
package handlers

import (
	"encoding/json"
	"time"

	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/websocket"
)

// RegisterRideHandlers registers all ride-related WebSocket handlers
func RegisterRideHandlers(manager *websocket.Manager) {
	manager.RegisterHandler(websocket.TypeRideLocation, handleRideLocationUpdate)
	manager.RegisterHandler(websocket.TypeRideStatusUpdate, handleRideStatusUpdate)
	manager.RegisterHandler("driver_availability", handleDriverAvailability)
	manager.RegisterHandler("ride_accept", handleRideAccept)
	manager.RegisterHandler("ride_cancel", handleRideCancel)
}

// handleRideLocationUpdate handles real-time location updates
func handleRideLocationUpdate(client *websocket.Client, msg *websocket.Message) error {
	rideID, ok := msg.Data["rideId"].(string)
	if !ok {
		return client.SendError("rideId required", msg.RequestID)
	}

	location, ok := msg.Data["location"].(map[string]interface{})
	if !ok {
		return client.SendError("location data required", msg.RequestID)
	}

	// Get the hub from the manager (you'll need to add this method to your Client struct)
	// For now, let's use the manager's hub
	if client.Manager() == nil {
		return client.SendError("manager not available", msg.RequestID)
	}

	hub := client.Manager().Hub()

	// Broadcast location update to all parties involved in the ride
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

// handleRideStatusUpdate handles ride status changes
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

	// Notify all parties about status change
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

// handleDriverAvailability handles driver availability updates
func handleDriverAvailability(client *websocket.Client, msg *websocket.Message) error {
	isAvailable, ok := msg.Data["isAvailable"].(bool)
	if !ok {
		return client.SendError("isAvailable required", msg.RequestID)
	}

	location, _ := msg.Data["location"].(map[string]interface{})

	// Store driver availability in Redis or database
	// This would be handled by your driver service
	logger.Info("driver availability updated",
		"driverID", client.UserID,
		"available", isAvailable,
		"location", location, // Use the location variable
	)

	return client.SendAck(msg.RequestID, map[string]interface{}{
		"success":     true,
		"isAvailable": isAvailable,
	})
}

// handleRideAccept handles ride acceptance by driver
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

	// Notify rider that driver accepted
	acceptMsg := websocket.NewTargetedMessage(websocket.TypeRideAccepted, riderID, map[string]interface{}{
		"rideId":    rideID,
		"driverId":  client.UserID,
		"timestamp": time.Now().UTC(),
	})

	// Send to specific rider instead of broadcasting to all
	hub.SendToUser(riderID, acceptMsg)

	logger.Info("ride accepted",
		"rideID", rideID,
		"driverID", client.UserID,
		"riderID", riderID,
	)

	return client.SendAck(msg.RequestID, map[string]interface{}{"success": true})
}

// handleRideCancel handles ride cancellation
func handleRideCancel(client *websocket.Client, msg *websocket.Message) error {
	rideID, ok := msg.Data["rideId"].(string)
	if !ok {
		return client.SendError("rideId required", msg.RequestID)
	}

	reason, _ := msg.Data["reason"].(string)
	targetUserID, _ := msg.Data["targetUserId"].(string) // Optional: specific user to notify

	if client.Manager() == nil {
		return client.SendError("manager not available", msg.RequestID)
	}

	hub := client.Manager().Hub()

	// Notify about cancellation
	cancelMsg := websocket.NewMessage(websocket.TypeRideCancelled, map[string]interface{}{
		"rideId":    rideID,
		"reason":    reason,
		"userId":    client.UserID,
		"timestamp": time.Now().UTC(),
	})

	if targetUserID != "" {
		// Send to specific user
		cancelMsg.TargetUserID = targetUserID
		hub.SendToUser(targetUserID, cancelMsg)
	} else {
		// Broadcast to all
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

// Helper function to extract user IDs from message
func extractUserIDs(data map[string]interface{}) ([]string, error) {
	userIDsInterface, ok := data["userIds"].([]interface{})
	if !ok {
		// Try to unmarshal from string
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
