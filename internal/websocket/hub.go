package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

type Hub struct {
	clients map[string][]*Client
	drivers map[string][]*Client
	riders map[string][]*Client
	adminClients      []*Client
	safetyTeamClients []*Client
	mu sync.RWMutex
	register chan *Client
	unregister chan *Client
	broadcast chan *Message
}

func NewHub() *Hub {
	return &Hub{
		clients:           make(map[string][]*Client),
		drivers:           make(map[string][]*Client),
		riders:            make(map[string][]*Client),
		adminClients:      make([]*Client, 0),
		safetyTeamClients: make([]*Client, 0),
		register:          make(chan *Client),
		unregister:        make(chan *Client),
		broadcast:         make(chan *Message, 256),
	}
}

func (h *Hub) Run(ctx context.Context) {
	pubsub := cache.SubscribeChannel(ctx, "websocket:broadcast")
	defer pubsub.Close()
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

				logger.Info("received message from redis",
					"type", broadcastMsg.Type,
					"targetUser", broadcastMsg.TargetUserID,
				)
				h.broadcast <- &broadcastMsg
			}
		}
	}()

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

func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.UserID] = append(h.clients[client.UserID], client)
	switch client.Role {
	case RoleDriver:
		h.drivers[client.UserID] = append(h.drivers[client.UserID], client)
		logger.Info("driver registered",
			"driverID", client.UserID,
			"clientID", client.ID,
			"totalDrivers", len(h.drivers),
		)
	case RoleRider:
		h.riders[client.UserID] = append(h.riders[client.UserID], client)
		logger.Info("rider registered",
			"riderID", client.UserID,
			"clientID", client.ID,

			"totalRiders", len(h.riders),
		)
	}

	deviceCount := len(h.clients[client.UserID])

	logger.Info("WebSocket client registered",
		"userID", client.UserID,
		"clientID", client.ID,
		"deviceCount", deviceCount,
		"userAgent", client.UserAgent,
		"totalUsers", len(h.clients),
		"totalConnections", h.getTotalConnectionsUnsafe(),
	)

	ctx := context.Background()
	metadata := map[string]interface{}{
		"userAgent": client.UserAgent,
		"clientID":  client.ID,
		"role":      string(client.Role),
	}

	logger.Debug("Setting Redis presence",
		"userID", client.UserID,
		"clientID", client.ID,
		"role", client.Role,
		"deviceCount", deviceCount,
		"totalClients", len(h.clients),
	)

	cache.SetPresence(ctx, client.UserID, client.ID, metadata)

	if deviceCount == 1 {
		logger.Info("USER ONLINE (First Connection)",
			"userID", client.UserID,
			"role", client.Role,
			"userAgent", client.UserAgent,
		)
		h.broadcastPresence(client.UserID, true)
	} else {
		logger.Info("New Device Connected (User already online)",
			"userID", client.UserID,
			"deviceCount", deviceCount,
		)
	}

	ackMsg := NewMessage(TypeConnectionAck, map[string]interface{}{
		"userId":    client.UserID,
		"clientId":  client.ID,
		"role":      string(client.Role),
		"timestamp": time.Now().UTC(),
	})
	client.send <- ackMsg
}

func (h *Hub) DebugInfo() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userConnections := make(map[string]int)
	for userID, clients := range h.clients {
		userConnections[userID] = len(clients)
	}

	info := map[string]interface{}{
		"total_users":       len(h.clients),
		"total_connections": h.getTotalConnectionsUnsafe(),
		"user_connections":  userConnections,
	}

	logger.Debug("Hub debug info", "info", info)

	return info
}

func (h *Hub) getTotalConnectionsUnsafe() int {
	total := 0
	for _, clients := range h.clients {
		total += len(clients)
	}
	return total
}

func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[client.UserID]; ok {
		logger.Debug("Finding client to unregister",
			"userID", client.UserID,
			"clientID", client.ID,
			"currentDeviceCount", len(clients),
		)

		for i, c := range clients {
			if c.ID == client.ID {
				h.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
				close(c.send)

				logger.Debug("Client found and removed from slice",
					"userID", client.UserID,
					"clientID", client.ID,
					"remainingDevices", len(h.clients[client.UserID]),
				)
				break
			}
		}

		switch client.Role {
		case RoleDriver:
			h.removeFromRoleMap(h.drivers, client)
		case RoleRider:
			h.removeFromRoleMap(h.riders, client)
		}

		if len(h.clients[client.UserID]) == 0 {
			delete(h.clients, client.UserID)

			logger.Info("User going offline (last device disconnected)",
				"userID", client.UserID,
				"clientID", client.ID,
			)

			ctx := context.Background()
			cache.RemovePresence(ctx, client.UserID, client.ID)

			h.broadcastPresence(client.UserID, false)

			logger.Info("WebSocket client unregistered - user offline",
				"userID", client.UserID,
				"clientID", client.ID,
				"role", client.Role,
				"totalUsers", len(h.clients),
				"totalConnections", h.getTotalConnectionsUnsafe(),
			)
		} else {
			ctx := context.Background()
			cache.RemovePresence(ctx, client.UserID, client.ID)

			logger.Info("WebSocket client unregistered - user still online",
				"userID", client.UserID,
				"clientID", client.ID,
				"role", client.Role,
				"remainingDevices", len(h.clients[client.UserID]),
			)
		}
	} else {
		logger.Warn("Attempted to unregister unknown client",
			"userID", client.UserID,
			"clientID", client.ID,
		)
	}
}

func (h *Hub) removeFromRoleMap(roleMap map[string][]*Client, client *Client) {
	if clients, ok := roleMap[client.UserID]; ok {
		for i, c := range clients {
			if c.ID == client.ID {
				roleMap[client.UserID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}
		if len(roleMap[client.UserID]) == 0 {
			delete(roleMap, client.UserID)
		}
	}
}

func (h *Hub) broadcastMessage(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	logger.Info("Broadcasting message",
		"type", message.Type,
		"targetUserID", message.TargetUserID,
		"total_users", len(h.clients),
		"total_connections", h.getTotalConnectionsUnsafe(),
		"payload", message.Data,
	)

	if message.TargetUserID != "" {
		if clients, ok := h.clients[message.TargetUserID]; ok {
			logger.Info("Sending to specific user",
				"targetUserID", message.TargetUserID,
				"device_count", len(clients),
				"messageType", message.Type,
			)

			successCount := 0
			for _, client := range clients {
				select {
				case client.send <- message:
					successCount++
					logger.Debug("Message queued to client",
						"userID", client.UserID,
						"clientID", client.ID,
						"role", client.Role,
						"type", message.Type,
						"queueSize", len(client.send),
					)
				default:
					logger.Warn("Client send buffer full - message dropped",
						"userID", client.UserID,
						"clientID", client.ID,
						"role", client.Role,
						"type", message.Type,
						"bufferSize", cap(client.send),
					)
				}
			}

			logger.Info("Message delivery summary",
				"targetUserID", message.TargetUserID,
				"messageType", message.Type,
				"totalDevices", len(clients),
				"successfulDeliveries", successCount,
				"failedDeliveries", len(clients)-successCount,
			)
		} else {
			logger.Warn("Target user not connected",
				"targetUserID", message.TargetUserID,
				"messageType", message.Type,
				"online_users", len(h.clients),
				"online_user_ids", h.getOnlineUserIDsUnsafe(),
			)
		}
	} else {
		logger.Info("Broadcasting to ALL users",
			"messageType", message.Type,
			"total_users", len(h.clients),
			"total_connections", h.getTotalConnectionsUnsafe(),
		)

		totalDevices := 0
		successCount := 0

		for userID, clients := range h.clients {
			for _, client := range clients {
				totalDevices++
				select {
				case client.send <- message:
					successCount++
					logger.Debug("Broadcast message queued",
						"userID", userID,
						"clientID", client.ID,
						"type", message.Type,
					)
				default:
					logger.Warn("Broadcast client send buffer full",
						"userID", userID,
						"clientID", client.ID,
						"type", message.Type,
					)
				}
			}
		}

		logger.Info("Broadcast delivery summary",
			"messageType", message.Type,
			"totalDevices", totalDevices,
			"successfulDeliveries", successCount,
			"failedDeliveries", totalDevices-successCount,
		)
	}
}

func (h *Hub) getOnlineUserIDsUnsafe() []string {
	userIDs := make([]string, 0, len(h.clients))
	for userID := range h.clients {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

func (h *Hub) broadcastPresence(userID string, isOnline bool) {
	msgType := TypeUserOnline
	status := "online"
	if !isOnline {
		msgType = TypeUserOffline
		status = "offline"
	}

	logger.Info("Broadcasting user presence",
		"userID", userID,
		"status", status,
		"messageType", msgType,
	)

	msg := NewMessage(msgType, map[string]interface{}{
		"userId": userID,
		"status": status,
	})

	h.broadcast <- msg

	ctx := context.Background()
	logger.Debug("Publishing presence to Redis",
		"userID", userID,
		"status", status,
	)
	cache.PublishMessage(ctx, "websocket:broadcast", msg)
}

func (h *Hub) closeAllConnections() {
	h.mu.Lock()
	defer h.mu.Unlock()

	logger.Info("Closing all WebSocket connections",
		"total_users", len(h.clients),
		"total_connections", h.getTotalConnectionsUnsafe(),
	)

	for userID, clients := range h.clients {
		for _, client := range clients {
			close(client.send)
			logger.Debug("Closed client connection",
				"userID", userID,
				"clientID", client.ID,
			)
		}
	}

	h.clients = make(map[string][]*Client)
	logger.Info("All connections closed")
}

func (h *Hub) SendToUser(userID string, msg *Message) {
	logger.Info("SendToUser called",
		"userID", userID,
		"messageType", msg.Type,
		"payload", msg.Data,
	)

	msg.TargetUserID = userID
	h.broadcast <- msg

	logger.Debug("Message queued for broadcast",
		"userID", userID,
		"messageType", msg.Type,
	)
	ctx := context.Background()
	cache.PublishMessage(ctx, "websocket:broadcast", msg)
}

func (h *Hub) SendToDriver(driverID string, msg *Message) {
	h.mu.RLock()
	clients, exists := h.drivers[driverID]
	h.mu.RUnlock()

	if !exists || len(clients) == 0 {
		logger.Warn("driver not connected",
			"driverID", driverID,
			"messageType", msg.Type,
		)
		return
	}

	logger.Info("sending message to driver",
		"driverID", driverID,
		"deviceCount", len(clients),
		"messageType", msg.Type,
		"messageID", msg.MessageID,
	)

	msg.TargetUserID = driverID
	h.broadcast <- msg

	ctx := context.Background()
	cache.PublishMessage(ctx, "websocket:broadcast", msg)
}

func (h *Hub) SendToRider(riderID string, msg *Message) {
	h.mu.RLock()
	clients, exists := h.riders[riderID]
	h.mu.RUnlock()

	if !exists || len(clients) == 0 {
		logger.Warn("rider not connected",
			"riderID", riderID,
			"messageType", msg.Type,
		)
		return
	}

	logger.Info("sending message to rider",
		"riderID", riderID,
		"deviceCount", len(clients),
		"messageType", msg.Type,
		"messageID", msg.MessageID,
	)

	msg.TargetUserID = riderID
	h.broadcast <- msg

	ctx := context.Background()
	cache.PublishMessage(ctx, "websocket:broadcast", msg)
}

func (h *Hub) BroadcastToAllDrivers(msg *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	logger.Info("broadcasting to all drivers",
		"driverCount", len(h.drivers),
		"messageType", msg.Type,
	)

	for _, clients := range h.drivers {
		for _, client := range clients {
			select {
			case client.send <- msg:
			default:
				logger.Warn("driver send buffer full",
					"driverID", client.UserID,
					"clientID", client.ID,
				)
			}
		}
	}
}

func (h *Hub) GetConnectedDrivers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.drivers)
}

func (h *Hub) GetConnectedRiders() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.riders)
}

func (h *Hub) BroadcastToAll(msg *Message) {
	logger.Info("BroadcastToAll called",
		"messageType", msg.Type,
		"payload", msg.Data,
		"currentConnections", h.GetTotalConnections(),
	)

	h.broadcast <- msg

	logger.Debug("Broadcast message queued",
		"messageType", msg.Type,
	)
}

func (h *Hub) IsDriverOnline(driverID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.drivers[driverID]
	return ok && len(clients) > 0
}

func (h *Hub) GetConnectedUsers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := len(h.clients)

	logger.Debug("GetConnectedUsers called", "count", count)
	return count
}

func (h *Hub) GetTotalConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	total := h.getTotalConnectionsUnsafe()
	logger.Debug("GetTotalConnections called", "count", total)
	return total
}

func (h *Hub) IsUserConnected(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[userID]
	connected := ok && len(clients) > 0

	logger.Debug("IsUserConnected called",
		"userID", userID,
		"connected", connected,
		"deviceCount", len(clients),
	)

	return connected
}

func (h *Hub) GetUserConnectionCount(userID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	if clients, ok := h.clients[userID]; ok {
		count = len(clients)
	}

	logger.Debug("GetUserConnectionCount called",
		"userID", userID,
		"count", count,
	)

	return count
}

func (h *Hub) BroadcastToRole(role string, msg *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	logger.Info("broadcasting to role",
		"role", role,
		"messageType", msg.Type,
	)

	var targetClients []*Client

	switch role {
	case string(RoleAdmin):
		targetClients = h.adminClients
	case string(RoleDriver):

		for _, clients := range h.drivers {
			targetClients = append(targetClients, clients...)
		}
	case string(RoleRider):

		for _, clients := range h.riders {
			targetClients = append(targetClients, clients...)
		}
	default:
		for _, clients := range h.clients {
			for _, client := range clients {
				if string(client.Role) == role {
					targetClients = append(targetClients, client)
				}
			}
		}
	}

	sentCount := 0
	for _, client := range targetClients {
		select {
		case client.send <- msg:
			sentCount++
			logger.Debug("message sent to role user",
				"role", role,
				"userID", client.UserID,
				"clientID", client.ID,
				"messageType", msg.Type,
			)
		default:
			logger.Warn("role send buffer full",
				"role", role,
				"userID", client.UserID,
				"clientID", client.ID,
			)
		}
	}

	logger.Info("broadcast to role completed",
		"role", role,
		"messageType", msg.Type,
		"totalTargets", len(targetClients),
		"sentToDevices", sentCount,
	)

	// Publish to Redis for multi-server setup
	ctx := context.Background()
	cache.PublishMessage(ctx, "websocket:broadcast", msg)
}
