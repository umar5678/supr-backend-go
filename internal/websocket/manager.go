package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/umar5678/go-backend/internal/utils/logger"
)

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

type Config struct {
	JWTSecret          string
	MaxConnections     int
	MessageBufferSize  int
	HeartbeatInterval  time.Duration
	ConnectionTimeout  time.Duration
	EnablePresence     bool
	EnableMessageStore bool
	PersistenceEnabled bool
}

type EventHandler func(client *Client, msg *Message) error

func NewManager(cfg *Config) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		hub:           NewHub(),
		config:        cfg,
		eventHandlers: make(map[MessageType]EventHandler),
		ctx:           ctx,
		cancel:        cancel,
	}

	if cfg.PersistenceEnabled {
		m.messageStore = NewRedisMessageStore()
		m.notificationStore = NewRedisNotificationStore()
	}

	m.registerDefaultHandlers()

	return m
}

func (m *Manager) Start() error {
	logger.Info("starting websocket manager")

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.hub.Run(m.ctx)
	}()

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.monitorHeartbeats()
	}()

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.collectMetrics()
	}()

	logger.Info("websocket manager started successfully")
	return nil
}

func (m *Manager) Shutdown(timeout time.Duration) error {
	logger.Info("shutting down websocket manager")

	m.cancel()

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

func (m *Manager) RegisterHandler(msgType MessageType, handler EventHandler) {
	m.handlersMutex.Lock()
	defer m.handlersMutex.Unlock()
	m.eventHandlers[msgType] = handler
	logger.Info("registered websocket handler", "type", msgType)
}

func (m *Manager) GetHandler(msgType MessageType) (EventHandler, bool) {
	m.handlersMutex.RLock()
	defer m.handlersMutex.RUnlock()
	handler, exists := m.eventHandlers[msgType]
	return handler, exists
}

func (m *Manager) Hub() *Hub {
	return m.hub
}

func (m *Manager) registerDefaultHandlers() {
	m.RegisterHandler(TypePing, m.handlePing)
	m.RegisterHandler(TypeTyping, m.handleTyping)
	m.RegisterHandler(TypeReadReceipt, m.handleReadReceipt)
	m.RegisterHandler(TypePresence, m.handlePresenceRequest)
}

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

func (m *Manager) monitorHeartbeats() {
	ticker := time.NewTicker(m.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
			case <-m.ctx.Done():
				return
			case <-ticker.C:
		}
	}
}

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

type Stats struct {
	ConnectedUsers         int
	TotalConnections       int
	AvgConnectionsPerUser  float64
	MessagesSentLast1Min   int64
	MessagesFailedLast1Min int64
}

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

func (m *Manager) SendNotification(userID string, notification interface{}) error {
	msg := NewTargetedMessage(TypeNotification, userID, map[string]interface{}{
		"notification": notification,
	})

	if m.config.PersistenceEnabled && m.notificationStore != nil {
		if err := m.notificationStore.Store(m.ctx, userID, notification); err != nil {
			logger.Error("failed to store notification", "error", err, "userID", userID)
		}
	}

	if m.hub.IsUserConnected(userID) {
		m.hub.SendToUser(userID, msg)
		return nil
	}

	logger.Debug("user offline, notification stored", "userID", userID)
	return nil
}

func (m *Manager) BroadcastNotification(userIDs []string, notification interface{}) error {
	for _, userID := range userIDs {
		if err := m.SendNotification(userID, notification); err != nil {
			logger.Error("failed to send notification", "error", err, "userID", userID)
		}
	}
	return nil
}
