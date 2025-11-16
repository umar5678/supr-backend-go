package websocket

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients (userID -> []*Client for multi-device support)
	clients map[string][]*Client

	// Mutex for thread-safe access to clients map
	mu sync.RWMutex

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to clients
	broadcast chan *Message
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string][]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
	}
}

// Run starts the hub and listens for events
func (h *Hub) Run(ctx context.Context) {
	// Subscribe to Redis for cross-server messaging
	pubsub := cache.SubscribeChannel(ctx, "websocket:broadcast")
	defer pubsub.Close()

	// Handle Redis messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := pubsub.ReceiveMessage(ctx)
				if err != nil {
					logger.Error("redis pubsub receive error", "error", err)
					continue
				}

				var broadcastMsg Message
				if err := json.Unmarshal([]byte(msg.Payload), &broadcastMsg); err != nil {
					logger.Error("failed to unmarshal broadcast message", "error", err)
					continue
				}

				// Broadcast to local clients
				h.broadcast <- &broadcastMsg
			}
		}
	}()

	// Main hub loop
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-ctx.Done():
			logger.Info("websocket hub shutting down")
			h.closeAllConnections()
			return
		}
	}
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add client to user's device list
	h.clients[client.UserID] = append(h.clients[client.UserID], client)

	deviceCount := len(h.clients[client.UserID])

	logger.Info("websocket client registered",
		"userID", client.UserID,
		"clientID", client.ID,
		"deviceCount", deviceCount,
	)

	// Set presence in Redis
	ctx := context.Background()
	metadata := map[string]interface{}{
		"userAgent": client.UserAgent,
		"clientID":  client.ID,
	}
	cache.SetPresence(ctx, client.UserID, client.ID, metadata)

	// Broadcast user online status (only if first device)
	if deviceCount == 1 {
		h.broadcastPresence(client.UserID, true)
	}
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[client.UserID]; ok {
		// Find and remove this specific client
		for i, c := range clients {
			if c.ID == client.ID {
				// Remove from slice
				h.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
				close(c.send)
				break
			}
		}

		// If no devices left, remove user entirely
		if len(h.clients[client.UserID]) == 0 {
			delete(h.clients, client.UserID)

			// Remove presence from Redis
			ctx := context.Background()
			cache.RemovePresence(ctx, client.UserID, client.ID)

			// Broadcast user offline status
			h.broadcastPresence(client.UserID, false)

			logger.Info("websocket client unregistered - user offline",
				"userID", client.UserID,
				"clientID", client.ID,
			)
		} else {
			// Just remove this device from Redis
			ctx := context.Background()
			cache.RemovePresence(ctx, client.UserID, client.ID)

			logger.Info("websocket client unregistered - user still online",
				"userID", client.UserID,
				"clientID", client.ID,
				"remainingDevices", len(h.clients[client.UserID]),
			)
		}
	}
}

// broadcastMessage sends message to appropriate clients
func (h *Hub) broadcastMessage(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if message.TargetUserID != "" {
		// Send to specific user (all their devices)
		if clients, ok := h.clients[message.TargetUserID]; ok {
			for _, client := range clients {
				select {
				case client.send <- message:
				default:
					logger.Warn("client send buffer full",
						"userID", client.UserID,
						"clientID", client.ID,
					)
				}
			}
		}
	} else {
		// Broadcast to all connected clients
		for _, clients := range h.clients {
			for _, client := range clients {
				select {
				case client.send <- message:
				default:
					logger.Warn("client send buffer full",
						"userID", client.UserID,
						"clientID", client.ID,
					)
				}
			}
		}
	}
}

// broadcastPresence broadcasts user online/offline status
func (h *Hub) broadcastPresence(userID string, isOnline bool) {
	msgType := TypeUserOnline
	if !isOnline {
		msgType = TypeUserOffline
	}

	msg := NewMessage(msgType, map[string]interface{}{
		"userId": userID,
	})

	// Broadcast locally
	h.broadcast <- msg

	// Publish to Redis for other servers
	ctx := context.Background()
	cache.PublishMessage(ctx, "websocket:broadcast", msg)
}

// closeAllConnections closes all client connections
func (h *Hub) closeAllConnections() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, clients := range h.clients {
		for _, client := range clients {
			close(client.send)
		}
	}

	h.clients = make(map[string][]*Client)
}

// Public methods for external use

// SendToUser sends a message to a specific user (all devices)
func (h *Hub) SendToUser(userID string, msg *Message) {
	msg.TargetUserID = userID
	h.broadcast <- msg
}

// BroadcastToAll sends a message to all connected clients
func (h *Hub) BroadcastToAll(msg *Message) {
	h.broadcast <- msg
}

// GetConnectedUsers returns the number of unique connected users
func (h *Hub) GetConnectedUsers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetTotalConnections returns the total number of connections (including multi-device)
func (h *Hub) GetTotalConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	total := 0
	for _, clients := range h.clients {
		total += len(clients)
	}
	return total
}

// IsUserConnected checks if a user has any active connections
func (h *Hub) IsUserConnected(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[userID]
	return ok && len(clients) > 0
}

// GetUserConnectionCount returns the number of connections for a user
func (h *Hub) GetUserConnectionCount(userID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[userID]; ok {
		return len(clients)
	}
	return 0
}
