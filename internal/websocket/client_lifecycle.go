package websocket

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"gorm.io/gorm"
)

type ClientLifecycle struct {
	sessionManager *SessionManager
	messageStore   MessageStore
	db             *gorm.DB
}

func NewClientLifecycle(sessionManager *SessionManager, messageStore MessageStore) *ClientLifecycle {
	return &ClientLifecycle{
		sessionManager: sessionManager,
		messageStore:   messageStore,
		db:             nil,
	}
}

func NewClientLifecycleWithDB(sessionManager *SessionManager, messageStore MessageStore, db *gorm.DB) *ClientLifecycle {
	return &ClientLifecycle{
		sessionManager: sessionManager,
		messageStore:   messageStore,
		db:             db,
	}
}

func (cl *ClientLifecycle) OnClientConnect(client *Client) error {
	session, err := cl.sessionManager.CreateSession(client.UserID, client.Role)
	if err != nil {
		logger.Error("failed to create session on connect", "error", err, "userId", client.UserID)
		return err
	}

	client.mu.Lock()
	client.reconnectToken = session.SessionID
	client.mu.Unlock()

	// Mark driver online when first connection is established
	if client.Role == models.RoleDriver && cl.db != nil {
		if err := cl.markDriverOnline(client.UserID); err != nil {
			logger.Warn("failed to mark driver online in database", "error", err, "driverId", client.UserID)
		}
	}

	logger.Info("client connected with new session",
		"clientId", client.ID,
		"userId", client.UserID,
		"sessionId", session.SessionID,
		"role", client.Role,
	)

	msg := NewMessage(TypeConnectionAck, map[string]interface{}{
		"sessionId": session.SessionID,
		"clientId":  client.ID,
		"role":      client.Role,
	})
	msg.RequireAck = false

	select {
	case client.send <- msg:
	default:
		logger.Warn("failed to send connection ack", "clientId", client.ID)
	}

	return nil
}

// OnClientDisconnect handles disconnection of a client
// Returns the session ID if the session should be kept alive for reconnection
func (cl *ClientLifecycle) OnClientDisconnect(client *Client) string {
	client.mu.RLock()
	sessionID := client.reconnectToken
	client.mu.RUnlock()

	if sessionID == "" {
		logger.Info("client disconnected without session", "clientId", client.ID, "userId", client.UserID)
		return ""
	}

	if err := cl.sessionManager.UpdateSession(sessionID, map[string]interface{}{
		"lastHeartbeat": time.Now(),
	}); err != nil {
		logger.Warn("failed to update session on disconnect", "error", err, "sessionId", sessionID)
	}

	// Mark driver offline when last connection is lost (will be called from hub after checking device count)
	if client.Role == models.RoleDriver && cl.db != nil {
		if err := cl.markDriverOffline(client.UserID); err != nil {
			logger.Warn("failed to mark driver offline in database", "error", err, "driverId", client.UserID)
		}
	}

	logger.Info("client disconnected, session preserved for reconnection",
		"clientId", client.ID,
		"userId", client.UserID,
		"sessionId", sessionID,
	)

	return sessionID
}

func (cl *ClientLifecycle) OnClientReconnect(client *Client, sessionID string) error {

	session, err := cl.sessionManager.ReconnectSession(sessionID, client.ID)
	if err != nil {
		logger.Error("failed to reconnect session", "error", err, "sessionId", sessionID, "userId", client.UserID)
		return err
	}

	client.mu.Lock()
	client.reconnectToken = session.SessionID
	client.mu.Unlock()

	logger.Info("client reconnected to existing session",
		"clientId", client.ID,
		"userId", client.UserID,
		"sessionId", session.SessionID,
		"reconnectionCount", session.ReconnectionCount,
	)

	return nil
}

func (cl *ClientLifecycle) GetSessionInfo(sessionID string) (*SessionState, error) {
	return cl.sessionManager.GetSession(sessionID)
}

// markDriverOnline updates driver status to 'online' in the database
func (cl *ClientLifecycle) markDriverOnline(driverUserID string) error {
	if cl.db == nil {
		return nil
	}

	result := cl.db.Model(&models.DriverProfile{}).
		Where("user_id = ?", driverUserID).
		Update("status", "online")

	if result.Error != nil {
		logger.Error("failed to update driver status to online",
			"error", result.Error,
			"driverId", driverUserID,
		)
		return result.Error
	}

	if result.RowsAffected > 0 {
		logger.Info("✅ Driver ONLINE",
			"driverId", driverUserID,
		)
	}

	return nil
}

// markDriverOffline updates driver status to 'offline' in the database
func (cl *ClientLifecycle) markDriverOffline(driverUserID string) error {
	if cl.db == nil {
		return nil
	}

	result := cl.db.Model(&models.DriverProfile{}).
		Where("user_id = ?", driverUserID).
		Update("status", "offline")

	if result.Error != nil {
		logger.Error("failed to update driver status to offline",
			"error", result.Error,
			"driverId", driverUserID,
		)
		return result.Error
	}

	if result.RowsAffected > 0 {
		logger.Info("Driver OFFLINE",
			"driverId", driverUserID,
		)
	}

	return nil
}
