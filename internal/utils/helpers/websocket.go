package helpers

import (
	"context"

	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/websocket"
)

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

func IsUserOnline(userID string) (bool, error) {
	ctx := context.Background()
	return cache.IsOnline(ctx, userID)
}

func GetUserDeviceCount(userID string) (int64, error) {
	ctx := context.Background()
	return cache.GetDeviceCount(ctx, userID)
}
