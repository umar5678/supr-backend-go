// internal/websocket/client.go - ENHANCED VERSION
package websocket

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512KB
)

// Client represents a WebSocket connection
type Client struct {
	ID             string
	UserID         string
	UserAgent      string
	hub            *Hub
	manager        *Manager
	conn           *websocket.Conn
	send           chan *Message
	reconnectToken string
	lastHeartbeat  time.Time
	connectedAt    time.Time
	mu             sync.RWMutex
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID, userAgent string) *Client {
	return &Client{
		ID:            uuid.New().String(),
		UserID:        userID,
		UserAgent:     userAgent,
		hub:           hub,
		conn:          conn,
		send:          make(chan *Message, 256),
		lastHeartbeat: time.Now(),
		connectedAt:   time.Now(),
	}
}

func (c *Client) Manager() *Manager {
	return c.manager
}

// ReadPump pumps messages from WebSocket to hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.updateHeartbeat()
		return nil
	})

	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("websocket read error",
					"error", err,
					"userID", c.UserID,
					"clientID", c.ID,
				)
			}
			break
		}

		// Update heartbeat on any message
		c.updateHeartbeat()

		// Handle incoming message
		c.handleIncomingMessage(&msg)
	}
}

// WritePump pumps messages from hub to WebSocket
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				logger.Error("websocket write error",
					"error", err,
					"userID", c.UserID,
					"clientID", c.ID,
				)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleIncomingMessage processes messages from client
func (c *Client) handleIncomingMessage(msg *Message) {
	// Set sender info
	if msg.Data == nil {
		msg.Data = make(map[string]interface{})
	}
	msg.Data["senderId"] = c.UserID

	// Check for registered handler in manager
	if c.manager != nil {
		if handler, exists := c.manager.GetHandler(msg.Type); exists {
			if err := handler(c, msg); err != nil {
				logger.Error("handler error",
					"error", err,
					"type", msg.Type,
					"userID", c.UserID,
				)
				c.SendError(err.Error(), msg.RequestID)
			}
			return
		}
	}

	// Fallback to default handling
	switch msg.Type {
	case TypePing:
		c.handlePing(msg)
	case TypeTyping:
		c.handleTyping(msg)
	case TypeReadReceipt:
		c.handleReadReceipt(msg)
	default:
		logger.Warn("unhandled message type",
			"type", msg.Type,
			"userID", c.UserID,
		)
		c.SendError("Unhandled message type", msg.RequestID)
	}
}

// Default message handlers

func (c *Client) handlePing(msg *Message) {
	pong := NewMessage(TypePong, map[string]interface{}{
		"timestamp": time.Now().UTC(),
	})
	pong.RequestID = msg.RequestID
	c.send <- pong
}

func (c *Client) handleTyping(msg *Message) {
	receiverID, ok := msg.Data["receiverId"].(string)
	if !ok {
		c.SendError("receiverId required", msg.RequestID)
		return
	}

	isTyping, _ := msg.Data["isTyping"].(bool)

	typingMsg := NewTargetedMessage(TypeTyping, receiverID, map[string]interface{}{
		"senderId": c.UserID,
		"isTyping": isTyping,
	})

	c.hub.SendToUser(receiverID, typingMsg)
	c.SendAck(msg.RequestID, map[string]interface{}{"success": true})
}

func (c *Client) handleReadReceipt(msg *Message) {
	senderID, ok := msg.Data["senderId"].(string)
	if !ok {
		c.SendError("senderId required", msg.RequestID)
		return
	}

	messageIDs, ok := msg.Data["messageIds"].([]interface{})
	if !ok {
		c.SendError("messageIds required", msg.RequestID)
		return
	}

	receiptMsg := NewTargetedMessage(TypeReadReceipt, senderID, map[string]interface{}{
		"readBy":     c.UserID,
		"messageIds": messageIDs,
		"readAt":     time.Now().UTC(),
	})

	c.hub.SendToUser(senderID, receiptMsg)
	c.SendAck(msg.RequestID, map[string]interface{}{"success": true})
}

// Helper methods

// SendError sends error message to client
func (c *Client) SendError(errMsg, requestID string) error {
	c.send <- NewErrorMessage(errMsg, requestID)
	return nil
}

// SendAck sends acknowledgment message
func (c *Client) SendAck(requestID string, data map[string]interface{}) error {
	c.send <- NewAckMessage(requestID, data)
	return nil
}

// GenerateReconnectToken generates a secure reconnection token
func (c *Client) GenerateReconnectToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// updateHeartbeat updates the last heartbeat timestamp
func (c *Client) updateHeartbeat() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastHeartbeat = time.Now()
}

// GetLastHeartbeat returns the last heartbeat time
func (c *Client) GetLastHeartbeat() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastHeartbeat
}

// GetConnectionDuration returns how long client has been connected
func (c *Client) GetConnectionDuration() time.Duration {
	return time.Since(c.connectedAt)
}

// IsHealthy checks if connection is healthy based on heartbeat
func (c *Client) IsHealthy() bool {
	return time.Since(c.GetLastHeartbeat()) < pongWait
}

// package websocket

// import (
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/gorilla/websocket"
// 	"github.com/umar5678/go-backend/internal/utils/logger"
// )

// const (
// 	// Time allowed to write a message to the peer
// 	writeWait = 10 * time.Second

// 	// Time allowed to read the next pong message from the peer
// 	pongWait = 60 * time.Second

// 	// Send pings to peer with this period (must be less than pongWait)
// 	pingPeriod = (pongWait * 9) / 10

// 	// Maximum message size allowed from peer
// 	maxMessageSize = 512 * 1024 // 512KB
// )

// // Client represents a WebSocket connection
// type Client struct {
// 	ID        string          // Unique client ID
// 	UserID    string          // User ID from authentication
// 	UserAgent string          // User agent string
// 	hub       *Hub            // Reference to hub
// 	conn      *websocket.Conn // WebSocket connection
// 	send      chan *Message   // Buffered channel of outbound messages
// }

// // NewClient creates a new WebSocket client
// func NewClient(hub *Hub, conn *websocket.Conn, userID, userAgent string) *Client {
// 	return &Client{
// 		ID:        uuid.New().String(),
// 		UserID:    userID,
// 		UserAgent: userAgent,
// 		hub:       hub,
// 		conn:      conn,
// 		send:      make(chan *Message, 256),
// 	}
// }

// // ReadPump pumps messages from the WebSocket connection to the hub
// func (c *Client) ReadPump() {
// 	defer func() {
// 		c.hub.unregister <- c
// 		c.conn.Close()
// 	}()

// 	c.conn.SetReadLimit(maxMessageSize)
// 	c.conn.SetReadDeadline(time.Now().Add(pongWait))
// 	c.conn.SetPongHandler(func(string) error {
// 		c.conn.SetReadDeadline(time.Now().Add(pongWait))
// 		return nil
// 	})

// 	for {
// 		var msg Message
// 		if err := c.conn.ReadJSON(&msg); err != nil {
// 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
// 				logger.Error("websocket read error",
// 					"error", err,
// 					"userID", c.UserID,
// 					"clientID", c.ID,
// 				)
// 			}
// 			break
// 		}

// 		// Handle incoming message
// 		c.handleIncomingMessage(&msg)
// 	}
// }

// // WritePump pumps messages from the hub to the WebSocket connection
// func (c *Client) WritePump() {
// 	ticker := time.NewTicker(pingPeriod)
// 	defer func() {
// 		ticker.Stop()
// 		c.conn.Close()
// 	}()

// 	for {
// 		select {
// 		case message, ok := <-c.send:
// 			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
// 			if !ok {
// 				// Hub closed the channel
// 				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
// 				return
// 			}

// 			if err := c.conn.WriteJSON(message); err != nil {
// 				logger.Error("websocket write error",
// 					"error", err,
// 					"userID", c.UserID,
// 					"clientID", c.ID,
// 				)
// 				return
// 			}

// 		case <-ticker.C:
// 			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
// 			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
// 				return
// 			}
// 		}
// 	}
// }

// // handleIncomingMessage processes messages received from the client
// func (c *Client) handleIncomingMessage(msg *Message) {
// 	switch msg.Type {
// 	case TypePing:
// 		c.handlePing(msg)
// 	case TypeTyping:
// 		c.handleTyping(msg)
// 	case TypeReadReceipt:
// 		c.handleReadReceipt(msg)
// 	default:
// 		logger.Warn("unknown message type received",
// 			"type", msg.Type,
// 			"userID", c.UserID,
// 		)
// 		c.sendError("Unknown message type", msg.RequestID)
// 	}
// }

// // handlePing responds with pong
// func (c *Client) handlePing(msg *Message) {
// 	pong := NewMessage(TypePong, map[string]interface{}{
// 		"timestamp": msg.Timestamp,
// 	})
// 	pong.RequestID = msg.RequestID
// 	c.send <- pong
// }

// // handleTyping broadcasts typing indicator
// func (c *Client) handleTyping(msg *Message) {
// 	receiverID, ok := msg.Data["receiverId"].(string)
// 	if !ok {
// 		c.sendError("receiverId required", msg.RequestID)
// 		return
// 	}

// 	isTyping, _ := msg.Data["isTyping"].(bool)

// 	typingMsg := NewTargetedMessage(TypeTyping, receiverID, map[string]interface{}{
// 		"senderId": c.UserID,
// 		"isTyping": isTyping,
// 	})

// 	c.hub.SendToUser(receiverID, typingMsg)

// 	// Send ack
// 	c.send <- NewAckMessage(msg.RequestID, map[string]interface{}{
// 		"success": true,
// 	})
// }

// // handleReadReceipt sends read receipt to sender
// func (c *Client) handleReadReceipt(msg *Message) {
// 	senderID, ok := msg.Data["senderId"].(string)
// 	if !ok {
// 		c.sendError("senderId required", msg.RequestID)
// 		return
// 	}

// 	messageIDs, ok := msg.Data["messageIds"].([]interface{})
// 	if !ok {
// 		c.sendError("messageIds required", msg.RequestID)
// 		return
// 	}

// 	receiptMsg := NewTargetedMessage(TypeReadReceipt, senderID, map[string]interface{}{
// 		"readBy":     c.UserID,
// 		"messageIds": messageIDs,
// 	})

// 	c.hub.SendToUser(senderID, receiptMsg)

// 	// Send ack
// 	c.send <- NewAckMessage(msg.RequestID, map[string]interface{}{
// 		"success": true,
// 	})
// }

// // sendError sends error message to client
// func (c *Client) sendError(errMsg, requestID string) {
// 	c.send <- NewErrorMessage(errMsg, requestID)
// }
