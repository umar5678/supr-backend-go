package websocket

import "time"

// MessageType defines different WebSocket message types
type MessageType string

const (
	// Notifications
	TypeNotification     MessageType = "notification"
	TypeNotificationRead MessageType = "notification_read"
	TypeNotificationBulk MessageType = "notification_bulk"

	// Chat
	TypeChatMessage     MessageType = "chat_message"
	TypeChatMessageSent MessageType = "chat_message_sent"
	TypeChatEdit        MessageType = "chat_edit"
	TypeChatDelete      MessageType = "chat_delete"
	TypeTyping          MessageType = "typing"
	TypeReadReceipt     MessageType = "read_receipt"

	// Presence
	TypeUserOnline  MessageType = "user_online"
	TypeUserOffline MessageType = "user_offline"
	TypePresence    MessageType = "presence"

	// System
	TypeSystemMessage MessageType = "system"
	TypeError         MessageType = "error"
	TypePing          MessageType = "ping"
	TypePong          MessageType = "pong"
	TypeAck           MessageType = "ack"
)

// Message represents a WebSocket message
type Message struct {
	Type         MessageType            `json:"type"`
	TargetUserID string                 `json:"targetUserId,omitempty"` // For targeted messages
	Data         map[string]interface{} `json:"data"`
	Timestamp    time.Time              `json:"timestamp"`
	RequestID    string                 `json:"requestId,omitempty"` // For request/response correlation
}

// NewMessage creates a new broadcast message
func NewMessage(msgType MessageType, data map[string]interface{}) *Message {
	return &Message{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
}

// NewTargetedMessage creates a message for specific user
func NewTargetedMessage(msgType MessageType, targetUserID string, data map[string]interface{}) *Message {
	return &Message{
		Type:         msgType,
		TargetUserID: targetUserID,
		Data:         data,
		Timestamp:    time.Now().UTC(),
	}
}

// NewErrorMessage creates an error message
func NewErrorMessage(errMsg string, requestID string) *Message {
	return &Message{
		Type: TypeError,
		Data: map[string]interface{}{
			"error": errMsg,
		},
		Timestamp: time.Now().UTC(),
		RequestID: requestID,
	}
}

// NewAckMessage creates an acknowledgment message
func NewAckMessage(requestID string, data map[string]interface{}) *Message {
	return &Message{
		Type:      TypeAck,
		Data:      data,
		Timestamp: time.Now().UTC(),
		RequestID: requestID,
	}
}
