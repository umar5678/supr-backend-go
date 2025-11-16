package helpers

import (
	"context"

	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/websocket"
)

// Notification Helpers

// SendNotification sends notification to a single user (all devices)
func SendNotification(userID string, notification interface{}) error {
	msg := websocket.NewTargetedMessage(
		websocket.TypeNotification,
		userID,
		map[string]interface{}{
			"notification": notification,
		},
	)

	ctx := context.Background()
	if err := cache.PublishMessage(ctx, "websocket:broadcast", msg); err != nil {
		logger.Error("failed to send notification",
			"error", err,
			"userID", userID,
		)
		return err
	}

	return nil
}

// SendNotificationToMultiple sends notification to multiple users
func SendNotificationToMultiple(userIDs []string, notification interface{}) error {
	for _, userID := range userIDs {
		if err := SendNotification(userID, notification); err != nil {
			logger.Error("failed to send notification to user",
				"error", err,
				"userID", userID,
			)
		}
	}
	return nil
}

// BroadcastNotification sends notification to all connected users
func BroadcastNotification(notification interface{}) error {
	msg := websocket.NewMessage(
		websocket.TypeNotification,
		map[string]interface{}{
			"notification": notification,
		},
	)

	ctx := context.Background()
	return cache.PublishMessage(ctx, "websocket:broadcast", msg)
}

// Chat Helpers

// SendChatMessage sends chat message to receiver
func SendChatMessage(receiverID string, message interface{}) error {
	msg := websocket.NewTargetedMessage(
		websocket.TypeChatMessage,
		receiverID,
		map[string]interface{}{
			"message": message,
		},
	)

	ctx := context.Background()
	if err := cache.PublishMessage(ctx, "websocket:broadcast", msg); err != nil {
		logger.Error("failed to send chat message",
			"error", err,
			"receiverID", receiverID,
		)
		return err
	}

	return nil
}

// SendChatMessageSent sends confirmation to sender
func SendChatMessageSent(senderID string, message interface{}) error {
	msg := websocket.NewTargetedMessage(
		websocket.TypeChatMessageSent,
		senderID,
		map[string]interface{}{
			"message": message,
		},
	)

	ctx := context.Background()
	return cache.PublishMessage(ctx, "websocket:broadcast", msg)
}

// SendTypingIndicator sends typing indicator to receiver
func SendTypingIndicator(receiverID, senderID string, isTyping bool) error {
	msg := websocket.NewTargetedMessage(
		websocket.TypeTyping,
		receiverID,
		map[string]interface{}{
			"senderId": senderID,
			"isTyping": isTyping,
		},
	)

	ctx := context.Background()
	return cache.PublishMessage(ctx, "websocket:broadcast", msg)
}

// SendReadReceipt sends read receipt to sender
func SendReadReceipt(senderID, receiverID string, messageIDs []string) error {
	msg := websocket.NewTargetedMessage(
		websocket.TypeReadReceipt,
		senderID,
		map[string]interface{}{
			"readBy":     receiverID,
			"messageIds": messageIDs,
		},
	)

	ctx := context.Background()
	return cache.PublishMessage(ctx, "websocket:broadcast", msg)
}

// SendMessageEdited notifies both users about message edit
func SendMessageEdited(senderID, receiverID string, message interface{}) error {
	msg := websocket.NewMessage(
		websocket.TypeChatEdit,
		map[string]interface{}{
			"message": message,
		},
	)

	ctx := context.Background()
	return cache.PublishMessage(ctx, "websocket:broadcast", msg)
}

// SendMessageDeleted notifies both users about message deletion
func SendMessageDeleted(senderID, receiverID, messageID string) error {
	msg := websocket.NewMessage(
		websocket.TypeChatDelete,
		map[string]interface{}{
			"messageId":  messageID,
			"senderId":   senderID,
			"receiverId": receiverID,
		},
	)

	ctx := context.Background()
	return cache.PublishMessage(ctx, "websocket:broadcast", msg)
}

// Presence Helpers

// SendPresenceUpdate broadcasts user online/offline status
func SendPresenceUpdate(userID string, isOnline bool) error {
	msgType := websocket.TypeUserOnline
	if !isOnline {
		msgType = websocket.TypeUserOffline
	}

	msg := websocket.NewMessage(msgType, map[string]interface{}{
		"userId":   userID,
		"isOnline": isOnline,
	})

	ctx := context.Background()
	return cache.PublishMessage(ctx, "websocket:broadcast", msg)
}

// System Helpers

// SendSystemMessage broadcasts system message to all users
func SendSystemMessage(message string, data map[string]interface{}) error {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["message"] = message

	msg := websocket.NewMessage(
		websocket.TypeSystemMessage,
		data,
	)

	ctx := context.Background()
	return cache.PublishMessage(ctx, "websocket:broadcast", msg)
}

// SendSystemMessageToUser sends system message to specific user
func SendSystemMessageToUser(userID, message string, data map[string]interface{}) error {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["message"] = message

	msg := websocket.NewTargetedMessage(
		websocket.TypeSystemMessage,
		userID,
		data,
	)

	ctx := context.Background()
	return cache.PublishMessage(ctx, "websocket:broadcast", msg)
}

// Utility Helpers

// IsUserOnline checks if user is currently online
func IsUserOnline(userID string) (bool, error) {
	ctx := context.Background()
	return cache.IsOnline(ctx, userID)
}

// GetUserDeviceCount returns number of active devices for user
func GetUserDeviceCount(userID string) (int64, error) {
	ctx := context.Background()
	return cache.GetDeviceCount(ctx, userID)
}
