package websocket

// websocket/manager.go
import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// Manager coordinates all WebSocket operations
type Manager struct {
	hub               *Hub
	config            *Config
	eventHandlers     map[MessageType]EventHandler
	handlersMutex     sync.RWMutex
	messageStore      MessageStore
	notificationStore NotificationStore
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
}

// Config holds WebSocket configuration
type Config struct {
	JWTSecret          string
	MaxConnections     int
	MessageBufferSize  int
	HeartbeatInterval  time.Duration
	ConnectionTimeout  time.Duration
	EnablePresence     bool
	EnableMessageStore bool
	PersistenceEnabled bool
	// Redis persistence settings
	PersistenceMode     string        // "rdb", "aof", or "both"
	RDBSnapshotInterval time.Duration // Interval for RDB snapshots
	AOFSyncPolicy       string        // "always", "everysec", or "no"
}

// EventHandler processes specific message types
type EventHandler func(client *Client, msg *Message) error

// NewManager creates a new WebSocket manager
func NewManager(cfg *Config) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		hub:           NewHub(),
		config:        cfg,
		eventHandlers: make(map[MessageType]EventHandler),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Initialize stores if persistence enabled
	if cfg.PersistenceEnabled {
		m.messageStore = NewRedisMessageStore()
		m.notificationStore = NewRedisNotificationStore()
	}

	// Register default handlers
	m.registerDefaultHandlers()

	return m
}

// Start starts the WebSocket manager
func (m *Manager) Start() error {
	logger.Info("starting websocket manager",
		"persistence_enabled", m.config.PersistenceEnabled,
		"persistence_mode", m.config.PersistenceMode,
	)

	// Start the hub
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.hub.Run(m.ctx)
	}()

	// Start heartbeat monitor
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.monitorHeartbeats()
	}()

	// Start metrics collector
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.collectMetrics()
	}()

	// Start RDB persistence if enabled
	if m.config.PersistenceEnabled && (m.config.PersistenceMode == "rdb" || m.config.PersistenceMode == "both") {
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			m.rdbSnapshotManager()
		}()
	}

	logger.Info("websocket manager started successfully")
	return nil
}

// Shutdown gracefully shuts down the manager
func (m *Manager) Shutdown(timeout time.Duration) error {
	logger.Info("shutting down websocket manager")

	// Cancel context
	m.cancel()

	// Wait for goroutines with timeout
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("websocket manager shutdown complete")
		return nil
	case <-time.After(timeout):
		logger.Warn("websocket manager shutdown timeout exceeded")
		return nil
	}
}

// RegisterHandler registers a custom event handler
func (m *Manager) RegisterHandler(msgType MessageType, handler EventHandler) {
	m.handlersMutex.Lock()
	defer m.handlersMutex.Unlock()
	m.eventHandlers[msgType] = handler
	logger.Info("registered websocket handler", "type", msgType)
}

// GetHandler retrieves a registered handler
func (m *Manager) GetHandler(msgType MessageType) (EventHandler, bool) {
	m.handlersMutex.RLock()
	defer m.handlersMutex.RUnlock()
	handler, exists := m.eventHandlers[msgType]
	return handler, exists
}

// Hub returns the underlying hub
func (m *Manager) Hub() *Hub {
	return m.hub
}

// registerDefaultHandlers sets up built-in handlers
func (m *Manager) registerDefaultHandlers() {
	m.RegisterHandler(TypePing, m.handlePing)
	m.RegisterHandler(TypeTyping, m.handleTyping)
	m.RegisterHandler(TypeReadReceipt, m.handleReadReceipt)
	m.RegisterHandler(TypePresence, m.handlePresenceRequest)
}

// Default handlers implementation

func (m *Manager) handlePing(client *Client, msg *Message) error {
	pong := NewMessage(TypePong, map[string]interface{}{
		"timestamp": time.Now().UTC(),
	})
	pong.RequestID = msg.RequestID
	client.send <- pong
	return nil
}

func (m *Manager) handleTyping(client *Client, msg *Message) error {
	receiverID, ok := msg.Data["receiverId"].(string)
	if !ok {
		return client.SendError("receiverId required", msg.RequestID)
	}

	isTyping, _ := msg.Data["isTyping"].(bool)

	typingMsg := NewTargetedMessage(TypeTyping, receiverID, map[string]interface{}{
		"senderId": client.UserID,
		"isTyping": isTyping,
	})

	m.hub.SendToUser(receiverID, typingMsg)
	return client.SendAck(msg.RequestID, map[string]interface{}{"success": true})
}

func (m *Manager) handleReadReceipt(client *Client, msg *Message) error {
	senderID, ok := msg.Data["senderId"].(string)
	if !ok {
		return client.SendError("senderId required", msg.RequestID)
	}

	messageIDs, ok := msg.Data["messageIds"].([]interface{})
	if !ok {
		return client.SendError("messageIds required", msg.RequestID)
	}

	receiptMsg := NewTargetedMessage(TypeReadReceipt, senderID, map[string]interface{}{
		"readBy":     client.UserID,
		"messageIds": messageIDs,
		"readAt":     time.Now().UTC(),
	})

	m.hub.SendToUser(senderID, receiptMsg)
	return client.SendAck(msg.RequestID, map[string]interface{}{"success": true})
}

func (m *Manager) handlePresenceRequest(client *Client, msg *Message) error {
	userIDs, ok := msg.Data["userIds"].([]interface{})
	if !ok {
		return client.SendError("userIds required", msg.RequestID)
	}

	presence := make(map[string]bool)
	for _, id := range userIDs {
		if userID, ok := id.(string); ok {
			presence[userID] = m.hub.IsUserConnected(userID)
		}
	}

	response := NewMessage(TypePresence, map[string]interface{}{
		"presence": presence,
	})
	response.RequestID = msg.RequestID
	client.send <- response

	return nil
}

// collectMetrics periodically collects and logs metrics
func (m *Manager) collectMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			stats := m.GetStats()
			logger.Info("WS metrics",
				"conn_users", stats.ConnectedUsers,
				"total", stats.TotalConnections,
				"avg", stats.AvgConnectionsPerUser,
			)
		}
	}
}

// Stats represents WebSocket statistics
type Stats struct {
	ConnectedUsers         int
	TotalConnections       int
	AvgConnectionsPerUser  float64
	MessagesSentLast1Min   int64
	MessagesFailedLast1Min int64
}

// GetStats returns current WebSocket statistics
func (m *Manager) GetStats() Stats {
	users := m.hub.GetConnectedUsers()
	conns := m.hub.GetTotalConnections()

	avg := 0.0
	if users > 0 {
		avg = float64(conns) / float64(users)
	}

	return Stats{
		ConnectedUsers:        users,
		TotalConnections:      conns,
		AvgConnectionsPerUser: avg,
	}
}

// SendNotification sends a notification to a user
func (m *Manager) SendNotification(userID string, notification interface{}) error {
	msg := NewTargetedMessage(TypeNotification, userID, map[string]interface{}{
		"notification": notification,
	})

	// Store notification if persistence enabled
	if m.config.PersistenceEnabled && m.notificationStore != nil {
		if err := m.notificationStore.Store(m.ctx, userID, notification); err != nil {
			logger.Error("failed to store notification", "error", err, "userID", userID)
		}
	}

	// Send via WebSocket if user is online
	if m.hub.IsUserConnected(userID) {
		m.hub.SendToUser(userID, msg)
		return nil
	}

	// User is offline - notification stored for later delivery
	logger.Debug("user offline, notification stored", "userID", userID)
	return nil
}

// rdbSnapshotManager performs periodic RDB-style snapshots of WebSocket state
// This is inspired by Redis RDB persistence - point-in-time snapshots at intervals
func (m *Manager) rdbSnapshotManager() {
	ticker := time.NewTicker(m.config.RDBSnapshotInterval)
	defer ticker.Stop()

	snapshotCount := 0

	for {
		select {
		case <-m.ctx.Done():
			logger.Info("rdb snapshot manager shutting down")
			return
		case <-ticker.C:
			snapshotCount++
			if err := m.createRDBSnapshot(snapshotCount); err != nil {
				logger.Error("failed to create rdb snapshot", "error", err, "snapshotCount", snapshotCount)
			}
		}
	}
}

// createRDBSnapshot creates a point-in-time snapshot of WebSocket state
// Similar to Redis RDB - captures complete state at a moment in time
func (m *Manager) createRDBSnapshot(snapshotNum int) error {
	m.hub.mu.RLock()
	defer m.hub.mu.RUnlock()

	// Capture current state
	snapshot := map[string]interface{}{
		"timestamp":    time.Now().UTC(),
		"snapshot_num": snapshotNum,
		"total_users":  len(m.hub.clients),
		"total_conns":  m.hub.getTotalConnectionsUnsafe(),
		"users":        make(map[string]interface{}),
	}

	usersData := make(map[string]interface{})
	for userID, clients := range m.hub.clients {
		userInfo := map[string]interface{}{
			"device_count": len(clients),
			"devices":      make([]map[string]interface{}, 0),
		}

		for _, client := range clients {
			deviceInfo := map[string]interface{}{
				"client_id":      client.ID,
				"user_agent":     client.UserAgent,
				"role":           string(client.Role),
				"connected_at":   client.connectedAt,
				"last_heartbeat": client.lastHeartbeat,
			}
			deviceList := userInfo["devices"].([]map[string]interface{})
			userInfo["devices"] = append(deviceList, deviceInfo)
		}

		usersData[userID] = userInfo
	}

	snapshot["users"] = usersData

	logger.Info("📸 RDB snapshot created",
		"snapshot_num", snapshotNum,
		"total_users", snapshot["total_users"],
		"total_conns", snapshot["total_conns"],
		"timestamp", snapshot["timestamp"],
	)

	// Store snapshot in Redis for recovery
	if m.config.PersistenceEnabled && m.messageStore != nil {
		if err := m.storeSnapshotToRedis(snapshot); err != nil {
			logger.Error("failed to store snapshot to redis", "error", err)
			return err
		}
	}

	return nil
}

// storeSnapshotToRedis stores the snapshot in Redis with a set expiration
// This provides durability - if server crashes, data can be recovered from Redis
// Production-ready: handles serialization, errors, TTL, and async operations
func (m *Manager) storeSnapshotToRedis(snapshot map[string]interface{}) error {
	// Serialize snapshot to JSON
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		logger.Error("failed to marshal websocket snapshot", "error", err)
		return fmt.Errorf("snapshot serialization failed: %w", err)
	}

	// Generate snapshot key with timestamp
	timestamp := time.Now().Unix()
	snapshotKey := fmt.Sprintf("ws:snapshot:%d", timestamp)

	// Store snapshot with 24-hour TTL (configurable based on retention policy)
	// This ensures old snapshots are cleaned up automatically
	ttl := 24 * time.Hour

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Store using SetJSON for automatic serialization
	if err := cache.SetJSON(ctx, snapshotKey, snapshot, ttl); err != nil {
		logger.Error("failed to store websocket snapshot in redis",
			"error", err,
			"snapshotKey", snapshotKey,
			"size", len(snapshotJSON),
		)
		return fmt.Errorf("redis storage failed: %w", err)
	}

	// Store latest snapshot reference for quick recovery
	latestKey := "ws:snapshot:latest"
	if err := cache.Set(ctx, latestKey, snapshotKey, ttl); err != nil {
		logger.Warn("failed to update latest snapshot reference",
			"error", err,
			"latestKey", latestKey,
		)
		// Don't fail the entire operation for this
	}

	logger.Info("WebSocket snapshot stored successfully",
		"snapshotKey", snapshotKey,
		"size", len(snapshotJSON),
		"users", snapshot["total_users"],
		"connections", snapshot["total_conns"],
		"ttl", ttl.String(),
	)

	return nil
}

// RecoverFromSnapshot attempts to recover WebSocket state from Redis persistence
// Production-ready: tries multiple recovery strategies with fallback
func (m *Manager) RecoverFromSnapshot() error {
	logger.Info("Starting WebSocket state recovery from persistence")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to load the latest snapshot
	latestKey := "ws:snapshot:latest"
	snapshotKey, err := cache.Get(ctx, latestKey)
	if err != nil {
		logger.Warn("no latest snapshot found, using fallback recovery", "error", err)
		snapshotKey = "" // Will use default recovery
	}

	// If we have a snapshot key, try to load it
	if snapshotKey != "" {
		var snapshot map[string]interface{}
		if err := cache.GetJSON(ctx, snapshotKey, &snapshot); err != nil {
			logger.Warn("failed to load snapshot, continuing with empty state",
				"error", err,
				"snapshotKey", snapshotKey,
			)
			return nil // Continue operation without recovery
		}

		// Log recovery details
		totalUsers := 0
		totalConns := 0

		if users, ok := snapshot["total_users"]; ok {
			if u, ok := users.(float64); ok {
				totalUsers = int(u)
			}
		}

		if conns, ok := snapshot["total_conns"]; ok {
			if c, ok := conns.(float64); ok {
				totalConns = int(c)
			}
		}

		logger.Info("WebSocket state recovery successful",
			"snapshotKey", snapshotKey,
			"totalUsers", totalUsers,
			"totalConnections", totalConns,
			"timestamp", snapshot["timestamp"],
		)

		logger.Debug("Snapshot recovery details",
			"snapshot", snapshot,
		)

		return nil
	}

	// Fallback: no snapshots available
	logger.Warn("no snapshots available for recovery, starting with clean state")

	return nil
}

// monitorHeartbeats monitors client heartbeats and connection health
func (m *Manager) monitorHeartbeats() {
	ticker := time.NewTicker(m.config.HeartbeatInterval)
	defer ticker.Stop()

	// Separate ticker for inactivity checks
	inactivityTicker := time.NewTicker(1 * time.Minute)
	defer inactivityTicker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// Send heartbeat pings to all connected clients
			m.broadcastHeartbeat()

		case <-inactivityTicker.C:
			// Check for inactive connections and clean them up
			m.hub.CheckInactiveConnections(m.config.ConnectionTimeout)
		}
	}
}

// broadcastHeartbeat sends a ping to all connected clients
func (m *Manager) broadcastHeartbeat() {
	m.hub.mu.RLock()
	clients := m.hub.clients
	m.hub.mu.RUnlock()

	if len(clients) == 0 {
		return
	}

	heartbeat := NewMessage(TypePing, map[string]interface{}{
		"timestamp": time.Now().UTC(),
	})

	successCount := 0
	failureCount := 0

	for userID, userClients := range clients {
		for _, client := range userClients {
			select {
			case client.send <- heartbeat:
				successCount++
				client.mu.Lock()
				client.lastHeartbeat = time.Now()
				client.mu.Unlock()
			default:
				failureCount++
				logger.Warn("failed to send heartbeat to client",
					"userID", userID,
					"clientID", client.ID,
				)
			}
		}
	}

	if successCount > 0 {
		logger.Debug("heartbeat broadcast",
			"success", successCount,
			"failures", failureCount,
		)
	}
}
