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
	maxMessageSize = 512 * 1024
)

type UserRole string

const (
	RoleDriver UserRole = "driver"
	RoleRider  UserRole = "rider"
	RoleAdmin  UserRole = "admin"
)

type Client struct {
	ID             string
	UserID         string
	UserAgent      string
	Role           UserRole
	hub            *Hub
	manager        *Manager
	conn           *websocket.Conn
	send           chan *Message
	reconnectToken string
	lastHeartbeat  time.Time
	connectedAt    time.Time
	mu             sync.RWMutex
	pendingAcks map[string]*Message
}

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

func (c *Client) ReadPump() {
	var closeErr error

	defer func() {
		if closeErr != nil {
			logger.Warn("🔌 Connection BROKEN",
				"userID", c.UserID,
				"role", c.Role,
				"clientID", c.ID,
				"error", closeErr.Error(),
			)
		} else {
			logger.Info("🔌 Connection closed normally",
				"userID", c.UserID,
				"role", c.Role,
				"clientID", c.ID,
			)
		}

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
			closeErr = err

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket unexpected close error",
					"error", err,
					"userID", c.UserID,
				)
			}
			break
		}

		c.updateHeartbeat()

		c.handleIncomingMessage(&msg)
	}
}

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
					"role", c.Role,
					"messageType", message.Type,
				)
				return
			}

			if message.RequireAck && message.MessageID != "" {
				c.pendingAcks[message.MessageID] = message

				go c.waitForAck(message.MessageID, 10*time.Second)
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) waitForAck(messageID string, timeout time.Duration) {
	time.Sleep(timeout)

	if msg, exists := c.pendingAcks[messageID]; exists {
		logger.Warn("message acknowledgment timeout",
			"messageID", messageID,
			"userID", c.UserID,
			"role", c.Role,
			"type", msg.Type,
		)

		if msg.RetryCount < 3 {
			msg.RetryCount++
			logger.Info("retrying message delivery",
				"messageID", messageID,
				"retryCount", msg.RetryCount,
			)
			c.send <- msg
		} else {
			logger.Error("message delivery failed after retries",
				"messageID", messageID,
				"userID", c.UserID,
				"type", msg.Type,
			)
			delete(c.pendingAcks, messageID)
		}
	}
}

func (c *Client) handleIncomingMessage(msg *Message) {
	if msg.Data == nil {
		msg.Data = make(map[string]interface{})
	}
	msg.Data["senderId"] = c.UserID

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

	switch msg.Type {
	case TypePing:
		c.handlePing(msg)
	case TypeAck:
		c.handleAck(msg)
	case TypeTyping:
		c.handleTyping(msg)
	case TypeReadReceipt:
		c.handleReadReceipt(msg)
	case TypeDriverLocationUpdate:
		c.handleDriverLocation(msg)
	default:
		logger.Warn("unhandled message type",
			"type", msg.Type,
			"userID", c.UserID,
			"role", c.Role,
		)
		c.SendError("Unhandled message type", msg.RequestID)
	}
}

func (c *Client) handleDriverLocation(msg *Message) {
	if c.Role != RoleDriver {
		c.SendError("Only drivers can send location updates", msg.RequestID)
		return
	}

	latitude, latOk := msg.Data["latitude"].(float64)
	longitude, lonOk := msg.Data["longitude"].(float64)

	if !latOk || !lonOk {
		c.SendError("latitude and longitude required", msg.RequestID)
		return
	}

	logger.Info("driver location update",
		"driverID", c.UserID,
		"latitude", latitude,
		"longitude", longitude,
	)

	locationMsg := NewMessage(TypeDriverLocationUpdate, map[string]interface{}{
		"driverId":  c.UserID,
		"latitude":  latitude,
		"longitude": longitude,
		"timestamp": time.Now().UTC(),
	})

	c.hub.BroadcastToAll(locationMsg)

	c.send <- NewAckMessage(msg.RequestID, map[string]interface{}{
		"success": true,
	})
}

func (c *Client) handleAck(msg *Message) {
	messageID, ok := msg.Data["messageId"].(string)
	if !ok {
		logger.Warn("ack without messageId", "userID", c.UserID)
		return
	}

	if _, exists := c.pendingAcks[messageID]; exists {
		logger.Info("message acknowledged",
			"messageID", messageID,
			"userID", c.UserID,
			"role", c.Role,
		)
		delete(c.pendingAcks, messageID)
	}
}

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

func (c *Client) SendError(errMsg, requestID string) error {
	c.send <- NewErrorMessage(errMsg, requestID)
	return nil
}

func (c *Client) SendAck(requestID string, data map[string]interface{}) error {
	c.send <- NewAckMessage(requestID, data)
	return nil
}

func (c *Client) GenerateReconnectToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (c *Client) updateHeartbeat() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastHeartbeat = time.Now()
}

func (c *Client) GetLastHeartbeat() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastHeartbeat
}

func (c *Client) GetConnectionDuration() time.Duration {
	return time.Since(c.connectedAt)
}

func (c *Client) IsHealthy() bool {
	return time.Since(c.GetLastHeartbeat()) < pongWait
}
