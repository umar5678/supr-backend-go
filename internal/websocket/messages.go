package websocket

import "time"

type MessageType string

const (
	TypeNotification     MessageType = "notification"
	TypeNotificationRead MessageType = "notification_read"
	TypeNotificationBulk MessageType = "notification_bulk"

	TypeChatMessage     MessageType = "chat_message"
	TypeChatMessageSent MessageType = "chat_message_sent"
	TypeChatEdit        MessageType = "chat_edit"
	TypeChatDelete      MessageType = "chat_delete"
	TypeTyping          MessageType = "typing"
	TypeReadReceipt     MessageType = "read_receipt"

	TypeUserOnline  MessageType = "user_online"
	TypeUserOffline MessageType = "user_offline"
	TypePresence    MessageType = "presence"

	TypeRideRequest          MessageType = "ride_request"           
	TypeRideRequestAccepted  MessageType = "ride_request_accepted"  
	TypeRideRequestRejected  MessageType = "ride_request_rejected"  
	TypeRideStatusUpdate     MessageType = "ride_status_update"     
	TypeRideDriverArriving   MessageType = "ride_driver_arriving"   
	TypeRideDriverArrived    MessageType = "ride_driver_arrived"    
	TypeRideStarted          MessageType = "ride_started"           
	TypeRideCompleted        MessageType = "ride_completed"         
	TypeRideCancelled        MessageType = "ride_cancelled"         
	TypeDriverLocationUpdate MessageType = "driver_location_update" 
	TypeRatingPrompt         MessageType = "rating_prompt"          

	TypeSOSAlert     = "sos_alert"
	TypeSOSResolved  = "sos_resolved"
	TypeSOSEscalated = "sos_escalated"

	TypeSystemMessage MessageType = "system"
	TypeError         MessageType = "error"
	TypePing          MessageType = "ping"
	TypePong          MessageType = "pong"
	TypeAck           MessageType = "ack"
	TypeConnectionAck MessageType = "connection_ack"
)

type Message struct {
	Type         MessageType            `json:"type"`
	TargetUserID string                 `json:"targetUserId,omitempty"` 
	Data         map[string]interface{} `json:"data"`
	Timestamp    time.Time              `json:"timestamp"`
	RequestID    string                 `json:"requestId,omitempty"`
	RequireAck bool   `json:"requireAck,omitempty"` 
	RetryCount int    `json:"-"`                    
	MessageID  string `json:"messageId,omitempty"`  
}

func NewMessage(msgType MessageType, data map[string]interface{}) *Message {
	return &Message{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
}

func NewTargetedMessage(msgType MessageType, targetUserID string, data map[string]interface{}) *Message {
	return &Message{
		Type:         msgType,
		TargetUserID: targetUserID,
		Data:         data,
		Timestamp:    time.Now().UTC(),
	}
}

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

func NewAckMessage(requestID string, data map[string]interface{}) *Message {
	return &Message{
		Type:      TypeAck,
		Data:      data,
		Timestamp: time.Now().UTC(),
		RequestID: requestID,
	}
}
