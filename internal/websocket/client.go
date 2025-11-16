package websocket

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512 * 1024 // 512KB
)

// Client represents a WebSocket connection
type Client struct {
	ID        string          // Unique client ID
	UserID    string          // User ID from authentication
	UserAgent string          // User agent string
	hub       *Hub            // Reference to hub
	conn      *websocket.Conn // WebSocket connection
	send      chan *Message   // Buffered channel of outbound messages
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID, userAgent string) *Client {
	return &Client{
		ID:        uuid.New().String(),
		UserID:    userID,
		UserAgent: userAgent,
		hub:       hub,
		conn:      conn,
		send:      make(chan *Message, 256),
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
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

		// Handle incoming message
		c.handleIncomingMessage(&msg)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
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
				// Hub closed the channel
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

// handleIncomingMessage processes messages received from the client
func (c *Client) handleIncomingMessage(msg *Message) {
	switch msg.Type {
	case TypePing:
		c.handlePing(msg)
	case TypeTyping:
		c.handleTyping(msg)
	case TypeReadReceipt:
		c.handleReadReceipt(msg)
	default:
		logger.Warn("unknown message type received",
			"type", msg.Type,
			"userID", c.UserID,
		)
		c.sendError("Unknown message type", msg.RequestID)
	}
}

// handlePing responds with pong
func (c *Client) handlePing(msg *Message) {
	pong := NewMessage(TypePong, map[string]interface{}{
		"timestamp": msg.Timestamp,
	})
	pong.RequestID = msg.RequestID
	c.send <- pong
}

// handleTyping broadcasts typing indicator
func (c *Client) handleTyping(msg *Message) {
	receiverID, ok := msg.Data["receiverId"].(string)
	if !ok {
		c.sendError("receiverId required", msg.RequestID)
		return
	}

	isTyping, _ := msg.Data["isTyping"].(bool)

	typingMsg := NewTargetedMessage(TypeTyping, receiverID, map[string]interface{}{
		"senderId": c.UserID,
		"isTyping": isTyping,
	})

	c.hub.SendToUser(receiverID, typingMsg)

	// Send ack
	c.send <- NewAckMessage(msg.RequestID, map[string]interface{}{
		"success": true,
	})
}

// handleReadReceipt sends read receipt to sender
func (c *Client) handleReadReceipt(msg *Message) {
	senderID, ok := msg.Data["senderId"].(string)
	if !ok {
		c.sendError("senderId required", msg.RequestID)
		return
	}

	messageIDs, ok := msg.Data["messageIds"].([]interface{})
	if !ok {
		c.sendError("messageIds required", msg.RequestID)
		return
	}

	receiptMsg := NewTargetedMessage(TypeReadReceipt, senderID, map[string]interface{}{
		"readBy":     c.UserID,
		"messageIds": messageIDs,
	})

	c.hub.SendToUser(senderID, receiptMsg)

	// Send ack
	c.send <- NewAckMessage(msg.RequestID, map[string]interface{}{
		"success": true,
	})
}

// sendError sends error message to client
func (c *Client) sendError(errMsg, requestID string) {
	c.send <- NewErrorMessage(errMsg, requestID)
}
