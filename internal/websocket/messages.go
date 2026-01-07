package websocket

// websocket/messages.go
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

	// ✅ RIDE-SPECIFIC EVENTS
	TypeRideRequest          MessageType = "ride_request"           // New ride request to driver
	TypeRideRequestAccepted  MessageType = "ride_request_accepted"  // Driver accepted
	TypeRideRequestRejected  MessageType = "ride_request_rejected"  // Driver rejected
	TypeRideStatusUpdate     MessageType = "ride_status_update"     // Status changed
	TypeRideDriverArriving   MessageType = "ride_driver_arriving"   // Driver approaching
	TypeRideDriverArrived    MessageType = "ride_driver_arrived"    // Driver at pickup
	TypeRideStarted          MessageType = "ride_started"           // Ride in progress
	TypeRideCompleted        MessageType = "ride_completed"         // Ride finished
	TypeRideCancelled        MessageType = "ride_cancelled"         // Ride cancelled
	TypeDriverLocationUpdate MessageType = "driver_location_update" // Driver location
	TypeRatingPrompt         MessageType = "rating_prompt"          // Prompt for ride rating

	// NEW: SOS types
	TypeSOSAlert     = "sos_alert"
	TypeSOSResolved  = "sos_resolved"
	TypeSOSEscalated = "sos_escalated"

	// System
	TypeSystemMessage MessageType = "system"
	TypeError         MessageType = "error"
	TypePing          MessageType = "ping"
	TypePong          MessageType = "pong"
	TypeAck           MessageType = "ack"
	TypeConnectionAck MessageType = "connection_ack" // ✅ NEW - Connection confirmation
)

// Message represents a WebSocket message
type Message struct {
	Type         MessageType            `json:"type"`
	TargetUserID string                 `json:"targetUserId,omitempty"` // For targeted messages
	Data         map[string]interface{} `json:"data"`
	Timestamp    time.Time              `json:"timestamp"`
	RequestID    string                 `json:"requestId,omitempty"` // For request/response correlation
	// ✅ NEW - Delivery tracking
	RequireAck bool   `json:"requireAck,omitempty"` // Message needs acknowledgment
	RetryCount int    `json:"-"`                    // Internal retry counter
	MessageID  string `json:"messageId,omitempty"`  // Unique message ID
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
