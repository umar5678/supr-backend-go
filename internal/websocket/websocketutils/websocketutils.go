package websocketutils

import (
	"context"
	"errors"
	"time"

	"github.com/umar5678/go-backend/internal/modules/admin_support_chat"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/websocket"
)

var (
	wsManager           *websocket.Manager
	adminSupportService admin_support_chat.Service
)

func Initialize(manager *websocket.Manager, adminChatService admin_support_chat.Service) {
	wsManager = manager
	adminSupportService = adminChatService
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

// Send a chat message from any role to admin support channel
func SendAdminSupportChat(senderID string, senderRole string, content string, metadata map[string]interface{}) error {
	if wsManager == nil {
		logger.Warn("websocket manager not initialized")
		return nil
	}
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	// Use user ID as conversation ID (one conversation per user with admins)
	conversationID := senderID

	// Save message to database if service is available
	if adminSupportService != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := adminSupportService.SendMessage(
			ctx,
			conversationID,
			senderID,
			senderRole,
			content,
			metadata,
		)
		if err != nil {
			logger.Error("failed to save admin support chat to database", "error", err, "senderId", senderID)
			// Don't block the real-time message if persistence fails
		}
	}

	// Broadcast real-time message to admins
	msg := websocket.NewMessage(websocket.TypeChatMessage, map[string]interface{}{
		"senderId":       senderID,
		"senderRole":     senderRole,
		"content":        content,
		"metadata":       metadata,
		"timestamp":      time.Now(),
		"adminSupport":   true,
		"conversationId": conversationID,
	})
	wsManager.Hub().BroadcastToRole("admin", msg)
	logger.Info("admin support chat sent", "senderId", senderID, "role", senderRole, "conversationId", conversationID)
	return nil
}

// Send SOS live location updates to admin until resolved
func SendSOSLocationUpdate(userID string, location map[string]interface{}, sosActive bool) error {
	if wsManager == nil {
		logger.Warn("websocket manager not initialized")
		return nil
	}
	if location == nil {
		logger.Error("SOS location update: location missing")
		return errors.New("location missing")
	}
	msg := websocket.NewMessage(websocket.TypeSOSAlert, map[string]interface{}{
		"userId":    userID,
		"location":  location,
		"timestamp": time.Now(),
		"sosActive": sosActive,
	})
	wsManager.Hub().BroadcastToRole("admin", msg)
	logger.Info("SOS location update sent to admin", "userId", userID, "sosActive", sosActive)
	return nil
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

// GetAdminSupportChatHistory retrieves persisted chat messages for a conversation
func GetAdminSupportChatHistory(ctx context.Context, conversationID string, limit, offset int) (interface{}, error) {
	if adminSupportService == nil {
		logger.Warn("admin support service not initialized")
		return nil, errors.New("admin support service not initialized")
	}

	messages, err := adminSupportService.GetConversation(ctx, conversationID, limit, offset)
	if err != nil {
		logger.Error("failed to retrieve conversation history", "error", err, "conversationId", conversationID)
		return nil, err
	}

	return messages, nil
}

// ReplyToAdminSupportChat creates a reply to an admin support message (threading)
func ReplyToAdminSupportChat(senderID string, senderRole string, conversationID string, parentMessageID string, content string) error {
	if adminSupportService == nil {
		logger.Warn("admin support service not initialized")
		return errors.New("admin support service not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := adminSupportService.ReplyToMessage(
		ctx,
		conversationID,
		parentMessageID,
		senderID,
		senderRole,
		content,
		make(map[string]interface{}),
	)
	if err != nil {
		logger.Error("failed to create admin support reply", "error", err, "conversationId", conversationID)
		return err
	}

	return nil
}
