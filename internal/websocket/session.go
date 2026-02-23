package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

const (
	sessionKeyPrefix        = "ws:session:"
	sessionExpiry           = 30 * time.Minute
	messageHistoryKeyPrefix = "ws:messages:"
	messageHistoryRetention = 1 * time.Hour
)

type SessionState struct {
	SessionID         string          `json:"sessionId"`
	UserID            string          `json:"userId"`
	ClientID          string          `json:"clientId"`
	Role              models.UserRole `json:"role"`
	LastMessageID     string          `json:"lastMessageId"`
	LastHeartbeat     time.Time       `json:"lastHeartbeat"`
	ConnectedAt       time.Time       `json:"connectedAt"`
	ReconnectionCount int             `json:"reconnectionCount"`
	ExpiresAt         time.Time       `json:"expiresAt"`
}

type SessionManager struct {
	ctx            context.Context
	activeSessions map[string]*SessionState
	sessionMutex   map[string]*sessionLock
	mu             *sync.RWMutex
}

type sessionLock struct {
	token string
}

func NewSessionManager(ctx context.Context) *SessionManager {
	return &SessionManager{
		ctx:            ctx,
		activeSessions: make(map[string]*SessionState),
		sessionMutex:   make(map[string]*sessionLock),
		mu:             &sync.RWMutex{},
	}
}

func (sm *SessionManager) CreateSession(userID string, role models.UserRole) (*SessionState, error) {
	sessionID := uuid.New().String()
	clientID := uuid.New().String()

	session := &SessionState{
		SessionID:     sessionID,
		UserID:        userID,
		ClientID:      clientID,
		Role:          role,
		LastHeartbeat: time.Now(),
		ConnectedAt:   time.Now(),
		ExpiresAt:     time.Now().Add(sessionExpiry),
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		logger.Error("failed to marshal session state", "error", err)
		return nil, err
	}

	key := sessionKeyPrefix + sessionID
	if err := cache.Set(sm.ctx, key, string(sessionJSON), sessionExpiry); err != nil {
		logger.Error("failed to store session in redis", "error", err)
		return nil, err
	}

	sm.mu.Lock()
	sm.activeSessions[sessionID] = session
	sm.mu.Unlock()

	logger.Info("session created",
		"sessionId", sessionID,
		"userId", userID,
		"role", role,
	)

	return session, nil
}

func (sm *SessionManager) GetSession(sessionID string) (*SessionState, error) {
	sm.mu.RLock()
	if session, ok := sm.activeSessions[sessionID]; ok {
		sm.mu.RUnlock()
		return session, nil
	}
	sm.mu.RUnlock()

	key := sessionKeyPrefix + sessionID
	val, err := cache.Get(sm.ctx, key)
	if err != nil {
		logger.Error("failed to get session from redis", "error", err)
		return nil, err
	}

	if val == "" {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	session := &SessionState{}
	if err := json.Unmarshal([]byte(val), session); err != nil {
		logger.Error("failed to unmarshal session state", "error", err)
		return nil, err
	}

	sm.mu.Lock()
	sm.activeSessions[sessionID] = session
	sm.mu.Unlock()

	return session, nil
}

func (sm *SessionManager) UpdateSession(sessionID string, updates map[string]interface{}) error {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return err
	}

	if lastMsgID, ok := updates["lastMessageId"].(string); ok {
		session.LastMessageID = lastMsgID
	}
	if lastHB, ok := updates["lastHeartbeat"].(time.Time); ok {
		session.LastHeartbeat = lastHB
	}

	session.ExpiresAt = time.Now().Add(sessionExpiry)

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return err
	}

	key := sessionKeyPrefix + sessionID
	if err := cache.Set(sm.ctx, key, string(sessionJSON), sessionExpiry); err != nil {
		logger.Error("failed to update session in redis", "error", err)
		return err
	}

	sm.mu.Lock()
	sm.activeSessions[sessionID] = session
	sm.mu.Unlock()

	return nil
}

func (sm *SessionManager) ReconnectSession(sessionID string, newClientID string) (*SessionState, error) {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to reconnect: session not found")
	}

	session.ClientID = newClientID
	session.ReconnectionCount++
	session.LastHeartbeat = time.Now()
	session.ExpiresAt = time.Now().Add(sessionExpiry)

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}

	key := sessionKeyPrefix + sessionID
	if err := cache.Set(sm.ctx, key, string(sessionJSON), sessionExpiry); err != nil {
		return nil, err
	}

	sm.mu.Lock()
	sm.activeSessions[sessionID] = session
	sm.mu.Unlock()

	logger.Info("session reconnected",
		"sessionId", sessionID,
		"userId", session.UserID,
		"newClientId", newClientID,
		"reconnectionCount", session.ReconnectionCount,
	)

	return session, nil
}

func (sm *SessionManager) DeleteSession(sessionID string) error {
	key := sessionKeyPrefix + sessionID
	if err := cache.Delete(sm.ctx, key); err != nil {
		logger.Error("failed to delete session from redis", "error", err)
		return err
	}

	sm.mu.Lock()
	delete(sm.activeSessions, sessionID)
	sm.mu.Unlock()

	logger.Info("session deleted", "sessionId", sessionID)
	return nil
}

func (sm *SessionManager) GetSessionsByUserID(userID string) []*SessionState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sessions []*SessionState
	for _, session := range sm.activeSessions {
		if session.UserID == userID {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

type ReconnectionHandler struct {
	sessionManager      *SessionManager
	messageStore        MessageStore
	connectionMonitor   *ConnectionMonitor
	maxMessagesPerBatch int
}

func NewReconnectionHandler(sessionManager *SessionManager, messageStore MessageStore, connectionMonitor *ConnectionMonitor) *ReconnectionHandler {
	return &ReconnectionHandler{
		sessionManager:      sessionManager,
		messageStore:        messageStore,
		connectionMonitor:   connectionMonitor,
		maxMessagesPerBatch: 100,
	}
}

func (rh *ReconnectionHandler) HandleReconnect(client *Client, reconnectReq *ReconnectRequest) error {
	session, err := rh.sessionManager.GetSession(reconnectReq.SessionID)
	if err != nil {
		logger.Warn("reconnection failed: session not found",
			"sessionId", reconnectReq.SessionID,
			"error", err,
		)
		client.SendMessage(NewMessage(TypeSessionExpired, map[string]interface{}{
			"reason": "Your session has expired. Please log in again.",
		}))
		return fmt.Errorf("session not found: %w", err)
	}

	if session.UserID != client.UserID {
		logger.Warn("reconnection rejected: user mismatch",
			"sessionId", reconnectReq.SessionID,
			"expected", session.UserID,
			"actual", client.UserID,
		)
		return fmt.Errorf("user mismatch for session")
	}

	if rh.connectionMonitor != nil && !rh.connectionMonitor.ValidateClientRole(session.UserID, session.Role) {
		logger.Warn("reconnection rejected: user role validation failed",
			"sessionId", reconnectReq.SessionID,
			"userId", session.UserID,
			"sessionRole", session.Role,
		)
		client.SendMessage(NewMessage(TypeSessionExpired, map[string]interface{}{
			"reason": "Your account role has changed. Please log in again.",
		}))
		return fmt.Errorf("user role validation failed")
	}

	updatedSession, err := rh.sessionManager.ReconnectSession(reconnectReq.SessionID, client.ID)
	if err != nil {
		logger.Error("failed to reconnect session", "error", err)
		return err
	}

	client.mu.Lock()
	client.reconnectToken = updatedSession.SessionID
	client.mu.Unlock()

	ackData := ReconnectAck{
		SessionID:          updatedSession.SessionID,
		ClientID:           client.ID,
		ReconnectionCount:  updatedSession.ReconnectionCount,
		HasPendingMessages: reconnectReq.LastMessageID != "" || updatedSession.LastMessageID != "",
		From:               "reconnect_handler",
	}

	ackMsg := NewMessage(TypeReconnectAck, map[string]interface{}{
		"sessionId":          ackData.SessionID,
		"clientId":           ackData.ClientID,
		"reconnectionCount":  ackData.ReconnectionCount,
		"hasPendingMessages": ackData.HasPendingMessages,
	})
	ackMsg.RequireAck = false

	client.send <- ackMsg

	logger.Info("client reconnected successfully",
		"sessionId", updatedSession.SessionID,
		"userId", updatedSession.UserID,
		"reconnectionCount", updatedSession.ReconnectionCount,
	)

	return nil
}

func (rh *ReconnectionHandler) SyncMessageHistory(client *Client, lastMessageID string) error {
	if lastMessageID == "" {
		logger.Info("no message sync needed, client has no previous messages")
		client.send <- NewMessage(TypeSyncComplete, map[string]interface{}{
			"synced": 0,
		})
		return nil
	}

	messages, err := rh.messageStore.GetMessagesAfterID(client.ctx, client.UserID, lastMessageID)
	if err != nil {
		logger.Error("failed to retrieve message history", "error", err)
		return err
	}

	if len(messages) == 0 {
		logger.Info("no pending messages for client",
			"userId", client.UserID,
			"lastMessageId", lastMessageID,
		)
		client.send <- NewMessage(TypeSyncComplete, map[string]interface{}{
			"synced": 0,
		})
		return nil
	}

	totalBatches := (len(messages) + rh.maxMessagesPerBatch - 1) / rh.maxMessagesPerBatch

	for batchIdx := 0; batchIdx < totalBatches; batchIdx++ {
		start := batchIdx * rh.maxMessagesPerBatch
		end := start + rh.maxMessagesPerBatch
		if end > len(messages) {
			end = len(messages)
		}

		batch := messages[start:end]
		batchLastID := ""
		if len(batch) > 0 {
			batchLastID = batch[len(batch)-1].MessageID
		}

		syncMsg := NewMessage(TypeMessageSync, map[string]interface{}{
			"messages":      batch,
			"batchIndex":    batchIdx,
			"totalBatches":  totalBatches,
			"lastMessageId": batchLastID,
		})
		syncMsg.RequireAck = true

		select {
		case client.send <- syncMsg:
			logger.Info("message batch sent",
				"userId", client.UserID,
				"batchIndex", batchIdx,
				"batchSize", len(batch),
				"totalBatches", totalBatches,
			)
		case <-client.ctx.Done():
			return fmt.Errorf("client context cancelled during sync")
		}

		time.Sleep(10 * time.Millisecond)
	}

	client.send <- NewMessage(TypeSyncComplete, map[string]interface{}{
		"synced": len(messages),
	})

	logger.Info("message history sync completed",
		"userId", client.UserID,
		"totalMessages", len(messages),
		"totalBatches", totalBatches,
	)

	return nil
}

func (rh *ReconnectionHandler) HandleSyncAck(client *Client, ackData map[string]interface{}) error {
	syncSessionID, ok := ackData["sessionId"].(string)
	if !ok {
		return fmt.Errorf("missing sessionId in sync ack")
	}

	lastMessageID, ok := ackData["lastMessageId"].(string)
	if !ok {
		return fmt.Errorf("missing lastMessageId in sync ack")
	}

	if err := rh.sessionManager.UpdateSession(syncSessionID, map[string]interface{}{
		"lastMessageId": lastMessageID,
	}); err != nil {
		logger.Error("failed to update session with last message id", "error", err)
		return err
	}

	logger.Info("sync acknowledgment received and processed",
		"sessionId", syncSessionID,
		"userId", client.UserID,
		"lastMessageId", lastMessageID,
	)

	return nil
}

func (rh *ReconnectionHandler) HandleReconnectMessage(client *Client, msg *Message) error {
	if msg.Data == nil {
		msg.Data = make(map[string]interface{})
	}

	switch msg.Type {
	case TypeReconnect:
		jsonData, _ := json.Marshal(msg.Data)
		var reconReq ReconnectRequest
		if err := json.Unmarshal(jsonData, &reconReq); err != nil {
			logger.Error("failed to parse reconnect request", "error", err)
			return err
		}

		if err := rh.HandleReconnect(client, &reconReq); err != nil {
			return err
		}

		return rh.SyncMessageHistory(client, reconReq.LastMessageID)

	case TypeMessageSyncAck:
		return rh.HandleSyncAck(client, msg.Data)

	default:
		return fmt.Errorf("unhandled reconnection message type: %s", msg.Type)
	}
}
