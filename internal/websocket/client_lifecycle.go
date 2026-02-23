package websocket

import (
	"time"

	"github.com/umar5678/go-backend/internal/utils/logger"
)

type ClientLifecycle struct {
	sessionManager *SessionManager
	messageStore   MessageStore
}

func NewClientLifecycle(sessionManager *SessionManager, messageStore MessageStore) *ClientLifecycle {
	return &ClientLifecycle{
		sessionManager: sessionManager,
		messageStore:   messageStore,
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
