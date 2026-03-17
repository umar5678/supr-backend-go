package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/utils/logger"
)

type ClientReconnectionStrategy struct {
	maxRetries             int
	initialBackoffTime     time.Duration
	maxBackoffTime         time.Duration
	heartbeatCheckInterval time.Duration
}

func NewClientReconnectionStrategy() *ClientReconnectionStrategy {
	return &ClientReconnectionStrategy{
		maxRetries:             5,
		initialBackoffTime:     time.Second,
		maxBackoffTime:         30 * time.Second,
		heartbeatCheckInterval: 30 * time.Second,
	}
}

func (crs *ClientReconnectionStrategy) GenerateReconnectPayload(sessionID string, lastMessageID string) map[string]interface{} {
	return map[string]interface{}{
		"sessionId":     sessionID,
		"lastMessageId": lastMessageID,
		"timestamp":     time.Now().UTC().Unix(),
	}
}

func (crs *ClientReconnectionStrategy) CalculateBackoffDuration(attemptNumber int) time.Duration {
	baseBackoff := crs.initialBackoffTime * (1 << uint(attemptNumber-1))
	if baseBackoff > crs.maxBackoffTime {
		baseBackoff = crs.maxBackoffTime
	}
	return baseBackoff
}

type MessageBuffer struct {
	messages []*Message
	maxSize  int
}

func NewMessageBuffer(maxSize int) *MessageBuffer {
	return &MessageBuffer{
		messages: make([]*Message, 0, maxSize),
		maxSize:  maxSize,
	}
}

func (mb *MessageBuffer) Add(msg *Message) error {
	if len(mb.messages) >= mb.maxSize {
		mb.messages = mb.messages[1:]
	}
	mb.messages = append(mb.messages, msg)
	return nil
}

func (mb *MessageBuffer) GetAll() []*Message {
	messages := make([]*Message, len(mb.messages))
	copy(messages, mb.messages)
	mb.messages = mb.messages[:0]
	return messages
}

func (mb *MessageBuffer) Size() int {
	return len(mb.messages)
}

type ClientConnectionValidator struct {
	minSessionDuration time.Duration
	maxSessionIdletime time.Duration
}

func NewClientConnectionValidator() *ClientConnectionValidator {
	return &ClientConnectionValidator{
		minSessionDuration: 5 * time.Second,
		maxSessionIdletime: 5 * time.Minute,
	}
}

func (ccv *ClientConnectionValidator) ValidateSessionForReconnect(session *SessionState) (bool, string) {
	if session == nil {
		return false, "session not found"
	}

	if time.Now().After(session.ExpiresAt) {
		return false, "session expired"
	}

	timeSinceActivity := time.Since(session.LastHeartbeat)
	if timeSinceActivity > ccv.maxSessionIdletime {
		return false, fmt.Sprintf("session idle for %v", timeSinceActivity)
	}

	return true, ""
}

func (ccv *ClientConnectionValidator) ValidateReconnectRequest(req *ReconnectRequest) error {
	if req.SessionID == "" {
		return fmt.Errorf("session ID required for reconnection")
	}

	if len(req.SessionID) < 10 {
		return fmt.Errorf("invalid session ID format")
	}

	return nil
}

type ReconnectionAuditLog struct {
	events []ReconnectionEvent
	maxLog int
}

type ReconnectionEvent struct {
	Timestamp         time.Time
	UserID            string
	SessionID         string
	OldClientID       string
	NewClientID       string
	ReconnectionCount int
	Success           bool
	ErrorReason       string
}

func NewReconnectionAuditLog(maxLog int) *ReconnectionAuditLog {
	return &ReconnectionAuditLog{
		events: make([]ReconnectionEvent, 0, maxLog),
		maxLog: maxLog,
	}
}

func (ral *ReconnectionAuditLog) LogReconnection(userID, sessionID, oldClientID, newClientID string, count int, success bool, err string) {
	event := ReconnectionEvent{
		Timestamp:         time.Now(),
		UserID:            userID,
		SessionID:         sessionID,
		OldClientID:       oldClientID,
		NewClientID:       newClientID,
		ReconnectionCount: count,
		Success:           success,
		ErrorReason:       err,
	}

	if len(ral.events) >= ral.maxLog {
		ral.events = ral.events[1:]
	}
	ral.events = append(ral.events, event)

	if success {
		logger.Info("reconnection successful",
			"userId", userID,
			"sessionId", sessionID,
			"oldClientId", oldClientID,
			"newClientId", newClientID,
			"reconnectionCount", count,
		)
	} else {
		logger.Error("reconnection failed",
			"userId", userID,
			"sessionId", sessionID,
			"reason", err,
		)
	}
}

func (ral *ReconnectionAuditLog) GetRecentEvents(limit int) []ReconnectionEvent {
	if limit > len(ral.events) {
		limit = len(ral.events)
	}
	result := make([]ReconnectionEvent, limit)
	copy(result, ral.events[len(ral.events)-limit:])
	return result
}

func SerializeReconnectMessage(sessionID, lastMessageID string) (*Message, error) {
	msg := NewMessage(TypeReconnect, map[string]interface{}{
		"sessionId":     sessionID,
		"lastMessageId": lastMessageID,
	})
	msg.RequireAck = true
	return msg, nil
}

func ParseReconnectAck(msg *Message) (*ReconnectAck, error) {
	if msg.Type != TypeReconnectAck {
		return nil, fmt.Errorf("expected TypeReconnectAck, got %s", msg.Type)
	}

	dataJSON, err := json.Marshal(msg.Data)
	if err != nil {
		return nil, err
	}

	var ack ReconnectAck
	if err := json.Unmarshal(dataJSON, &ack); err != nil {
		return nil, err
	}

	return &ack, nil
}

type HealthCheckPayload struct {
	SessionID       string
	ClientID        string
	LastHeartbeat   time.Time
	MessagesPending int
	Status          string
}

func CreateHealthCheckMessage(sessionID, clientID string, messagesPending int) *Message {
	msg := NewMessage(TypePing, map[string]interface{}{
		"sessionId":       sessionID,
		"clientId":        clientID,
		"lastHeartbeat":   time.Now(),
		"messagesPending": messagesPending,
		"status":          "healthy",
	})
	msg.RequireAck = true
	return msg
}
