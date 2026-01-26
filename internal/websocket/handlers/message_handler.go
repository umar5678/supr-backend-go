// internal/websocket/handlers/message_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/modules/messages"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/websocket"
)

// Message event types
const (
	MessageEventNew      websocket.MessageType = "message:new"
	MessageEventRead     websocket.MessageType = "message:read"
	MessageEventDelete   websocket.MessageType = "message:delete"
	MessageEventTyping   websocket.MessageType = "message:typing"
	PresenceEventOnline  websocket.MessageType = "presence:online"
	PresenceEventOffline websocket.MessageType = "presence:offline"
)

// MessageHandler handles messaging events
type MessageHandler struct {
	messageService messages.Service
	manager        *websocket.Manager
}

// Message event payloads
type MessageEventPayload struct {
	ID          string                 `json:"id"`
	RideID      string                 `json:"rideId"`
	SenderID    string                 `json:"senderId"`
	SenderName  string                 `json:"senderName"`
	SenderType  string                 `json:"senderType"`
	Content     string                 `json:"content"`
	MessageType string                 `json:"messageType"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	IsRead      bool                   `json:"isRead"`
	ReadAt      *time.Time             `json:"readAt,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// RegisterMessageHandlers registers message-related WebSocket handlers
func RegisterMessageHandlers(manager *websocket.Manager, msgService messages.Service) {
	handler := &MessageHandler{
		messageService: msgService,
		manager:        manager,
	}

	// Register handlers for different message types
	manager.RegisterHandler(websocket.MessageType("message:send"), handler.HandleSendMessage)
	manager.RegisterHandler(websocket.MessageType("message:read"), handler.HandleMarkAsRead)
	manager.RegisterHandler(websocket.MessageType("message:delete"), handler.HandleDeleteMessage)
	manager.RegisterHandler(websocket.MessageType("message:typing"), handler.HandleTyping)
	manager.RegisterHandler(websocket.MessageType("presence:online"), handler.HandlePresenceOnline)
	manager.RegisterHandler(websocket.MessageType("presence:offline"), handler.HandlePresenceOffline)

	logger.Info("message websocket handlers registered")
}

// HandleSendMessage processes new message events
func (h *MessageHandler) HandleSendMessage(client *websocket.Client, msg *websocket.Message) error {
	// Parse payload from message data
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		logger.Error("failed to marshal message data", "error", err)
		return err
	}

	var payload struct {
		RideID   string                 `json:"rideId"`
		Content  string                 `json:"content"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}

	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		logger.Error("failed to unmarshal message data", "error", err)
		return err
	}

	// Validate input
	if payload.RideID == "" || payload.Content == "" {
		return fmt.Errorf("rideId and content are required")
	}

	// Get sender info from client
	userID := client.UserID
	userType := "rider" // Default, convert from role
	if client.Role == websocket.RoleDriver {
		userType = "driver"
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Save message to database via service
	msgResp, err := h.messageService.SendMessage(ctx, payload.RideID, userID, userType, payload.Content, payload.Metadata)
	if err != nil {
		logger.Error("failed to send message", "error", err, "rideId", payload.RideID)
		return err
	}

	// Prepare event payload
	eventPayload := MessageEventPayload{
		ID:          msgResp.ID,
		RideID:      msgResp.RideID,
		SenderID:    msgResp.SenderID,
		SenderName:  msgResp.SenderName,
		SenderType:  msgResp.SenderType,
		Content:     msgResp.Content,
		MessageType: msgResp.MessageType,
		Metadata:    msgResp.Metadata,
		IsRead:      msgResp.IsRead,
		Timestamp:   msgResp.CreatedAt,
	}

	// Broadcast to all clients in the hub
	h.manager.Hub().BroadcastToAll(&websocket.Message{
		Type: websocket.MessageType(MessageEventNew),
		Data: map[string]interface{}{
			"message": eventPayload,
			"rideId":  payload.RideID,
		},
		Timestamp: time.Now(),
	})

	logger.Info("message sent and broadcasted", "rideId", payload.RideID, "messageId", msgResp.ID)
	return nil
}

// HandleMarkAsRead processes mark as read events
func (h *MessageHandler) HandleMarkAsRead(client *websocket.Client, msg *websocket.Message) error {
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return err
	}

	var payload struct {
		MessageID string `json:"messageId"`
		RideID    string `json:"rideId"`
	}

	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		return err
	}

	if payload.MessageID == "" || payload.RideID == "" {
		return fmt.Errorf("messageId and rideId are required")
	}

	// Mark as read in database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.messageService.MarkAsRead(ctx, payload.MessageID, client.UserID); err != nil {
		logger.Error("failed to mark as read", "error", err)
		return err
	}

	// Broadcast read receipt
	h.manager.Hub().BroadcastToAll(&websocket.Message{
		Type: websocket.MessageType(MessageEventRead),
		Data: map[string]interface{}{
			"messageId": payload.MessageID,
			"rideId":    payload.RideID,
			"readBy":    client.UserID,
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	})

	logger.Info("message marked as read", "messageId", payload.MessageID, "rideId", payload.RideID)
	return nil
}

// HandleDeleteMessage processes message deletion
func (h *MessageHandler) HandleDeleteMessage(client *websocket.Client, msg *websocket.Message) error {
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return err
	}

	var payload struct {
		MessageID string `json:"messageId"`
		RideID    string `json:"rideId"`
	}

	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		return err
	}

	if payload.MessageID == "" || payload.RideID == "" {
		return fmt.Errorf("messageId and rideId are required")
	}

	// Delete message (service validates ownership and time window)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.messageService.DeleteMessage(ctx, payload.MessageID, client.UserID); err != nil {
		logger.Error("failed to delete message", "error", err)
		return err
	}

	// Broadcast deletion
	h.manager.Hub().BroadcastToAll(&websocket.Message{
		Type: websocket.MessageType(MessageEventDelete),
		Data: map[string]interface{}{
			"messageId": payload.MessageID,
			"rideId":    payload.RideID,
			"deletedBy": client.UserID,
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	})

	logger.Info("message deleted", "messageId", payload.MessageID, "rideId", payload.RideID)
	return nil
}

// HandleTyping processes typing indicator events
func (h *MessageHandler) HandleTyping(client *websocket.Client, msg *websocket.Message) error {
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return err
	}

	var payload struct {
		RideID   string `json:"rideId"`
		IsTyping bool   `json:"isTyping"`
	}

	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		return err
	}

	if payload.RideID == "" {
		return fmt.Errorf("rideId is required")
	}

	// Broadcast typing indicator
	h.manager.Hub().BroadcastToAll(&websocket.Message{
		Type: websocket.MessageType(MessageEventTyping),
		Data: map[string]interface{}{
			"rideId":   payload.RideID,
			"userId":   client.UserID,
			"isTyping": payload.IsTyping,
		},
		Timestamp: time.Now(),
	})

	logger.Info("typing indicator", "rideId", payload.RideID, "userId", client.UserID, "isTyping", payload.IsTyping)
	return nil
}

// HandlePresenceOnline processes user online events
func (h *MessageHandler) HandlePresenceOnline(client *websocket.Client, msg *websocket.Message) error {
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return err
	}

	var payload struct {
		RideID string `json:"rideId"`
	}

	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		return err
	}

	if payload.RideID == "" {
		return fmt.Errorf("rideId is required")
	}

	// Broadcast online status
	h.manager.Hub().BroadcastToAll(&websocket.Message{
		Type: websocket.MessageType(PresenceEventOnline),
		Data: map[string]interface{}{
			"rideId":    payload.RideID,
			"userId":    client.UserID,
			"status":    "online",
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	})

	logger.Info("user online", "rideId", payload.RideID, "userId", client.UserID)
	return nil
}

// HandlePresenceOffline processes user offline events
func (h *MessageHandler) HandlePresenceOffline(client *websocket.Client, msg *websocket.Message) error {
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return err
	}

	var payload struct {
		RideID string `json:"rideId"`
	}

	if err := json.Unmarshal(dataBytes, &payload); err != nil {
		return err
	}

	if payload.RideID == "" {
		return fmt.Errorf("rideId is required")
	}

	// Broadcast offline status
	h.manager.Hub().BroadcastToAll(&websocket.Message{
		Type: websocket.MessageType(PresenceEventOffline),
		Data: map[string]interface{}{
			"rideId":    payload.RideID,
			"userId":    client.UserID,
			"status":    "offline",
			"timestamp": time.Now(),
		},
		Timestamp: time.Now(),
	})

	logger.Info("user offline", "rideId", payload.RideID, "userId", client.UserID)
	return nil
}
