package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/notifications/repository"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// PushToken represents a device push token
type PushToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"type:varchar(500);uniqueIndex" json:"token"`
	DeviceID  string    `gorm:"type:varchar(255)" json:"device_id"`
	DeviceOS  string    `gorm:"type:varchar(50)" json:"device_os"` // ios, android, web
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (PushToken) TableName() string {
	return "push_tokens"
}

// LocalPushService implements push notifications locally without external services
type LocalPushService struct {
	db        *gorm.DB
	notifRepo repository.NotificationRepository
	mu        sync.RWMutex
	// In-memory push queue for real-time delivery
	pushQueue map[uuid.UUID][]PushMessage
	// Connected WebSocket clients for real-time delivery
	subscribers map[uuid.UUID][]PushSubscriber
}

// PushMessage represents a message to be pushed
type PushMessage struct {
	ID        uuid.UUID              `json:"id"`
	UserID    uuid.UUID              `json:"user_id"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Read      bool                   `json:"read"`
}

// PushSubscriber represents a WebSocket subscriber
type PushSubscriber struct {
	ID   string
	Chan chan PushMessage
}

func NewLocalPushService(db *gorm.DB, notifRepo repository.NotificationRepository) *LocalPushService {
	return &LocalPushService{
		db:          db,
		notifRepo:   notifRepo,
		pushQueue:   make(map[uuid.UUID][]PushMessage),
		subscribers: make(map[uuid.UUID][]PushSubscriber),
	}
}

// RegisterToken registers or updates a push token for a user
func (s *LocalPushService) RegisterToken(ctx context.Context, userID uuid.UUID, token, deviceID, deviceOS string) error {
	token_record := &PushToken{
		UserID:   userID,
		Token:    token,
		DeviceID: deviceID,
		DeviceOS: deviceOS,
	}

	// Upsert: update if exists, create if not
	result := s.db.WithContext(ctx).
		Where("token = ?", token).
		Assign(token_record).
		FirstOrCreate(token_record)

	if result.Error != nil {
		logger.Error("failed to register push token", "error", result.Error, "userID", userID)
		return fmt.Errorf("failed to register push token: %w", result.Error)
	}

	logger.Info("push token registered", "userID", userID, "deviceOS", deviceOS)
	return nil
}

// UnregisterToken removes a push token
func (s *LocalPushService) UnregisterToken(ctx context.Context, token string) error {
	if err := s.db.WithContext(ctx).Where("token = ?", token).Delete(&PushToken{}).Error; err != nil {
		logger.Error("failed to unregister push token", "error", err)
		return fmt.Errorf("failed to unregister push token: %w", err)
	}
	return nil
}

// SendPush sends a push notification to a user
func (s *LocalPushService) SendPush(ctx context.Context, userID uuid.UUID, title, body string, data map[string]interface{}) error {
	message := PushMessage{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		Body:      body,
		Data:      data,
		Timestamp: time.Now(),
		Read:      false,
	}

	// Queue message in memory
	s.mu.Lock()
	s.pushQueue[userID] = append(s.pushQueue[userID], message)
	s.mu.Unlock()

	// Store in database for persistence
	dataBytes, _ := json.Marshal(data)
	notification := &models.Notification{
		UserID:   userID,
		Channel:  models.ChannelPush,
		Status:   models.NotificationStatusPending,
		Title:    title,
		Message:  body,
		Metadata: dataBytes,
	}

	if err := s.db.WithContext(ctx).Create(notification).Error; err != nil {
		logger.Error("failed to persist push notification", "error", err, "userID", userID)
		return fmt.Errorf("failed to persist notification: %w", err)
	}

	// Broadcast to real-time subscribers
	s.BroadcastToUser(userID, message)

	logger.Info("push notification sent", "userID", userID, "title", title)
	return nil
}

// GetPendingMessages retrieves pending push messages for a user
func (s *LocalPushService) GetPendingMessages(ctx context.Context, userID uuid.UUID) ([]PushMessage, error) {
	s.mu.RLock()
	messages, exists := s.pushQueue[userID]
	s.mu.RUnlock()

	if !exists {
		return []PushMessage{}, nil
	}

	// Filter unread messages
	var unread []PushMessage
	for _, msg := range messages {
		if !msg.Read {
			unread = append(unread, msg)
		}
	}

	return unread, nil
}

// MarkMessageAsRead marks a push message as read
func (s *LocalPushService) MarkMessageAsRead(ctx context.Context, userID uuid.UUID, messageID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	messages, exists := s.pushQueue[userID]
	if !exists {
		return nil
	}

	for i := range messages {
		if messages[i].ID == messageID {
			messages[i].Read = true
			break
		}
	}

	return nil
}

// ClearOldMessages clears messages older than retention period
func (s *LocalPushService) ClearOldMessages(ctx context.Context, retentionHours int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	retentionTime := now.Add(-time.Duration(retentionHours) * time.Hour)

	for userID := range s.pushQueue {
		var retained []PushMessage
		for _, msg := range s.pushQueue[userID] {
			if msg.Timestamp.After(retentionTime) {
				retained = append(retained, msg)
			}
		}
		s.pushQueue[userID] = retained
	}

	return nil
}

// SubscribeToUser subscribes a WebSocket connection to user push notifications
func (s *LocalPushService) SubscribeToUser(userID uuid.UUID, subscriberID string, msgChan chan PushMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	subscriber := PushSubscriber{
		ID:   subscriberID,
		Chan: msgChan,
	}

	s.subscribers[userID] = append(s.subscribers[userID], subscriber)
	logger.Info("push subscriber connected", "userID", userID, "subscriberID", subscriberID)

	return nil
}

// UnsubscribeFromUser removes a WebSocket subscriber
func (s *LocalPushService) UnsubscribeFromUser(userID uuid.UUID, subscriberID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	subscribers, exists := s.subscribers[userID]
	if !exists {
		return nil
	}

	// Remove subscriber
	var retained []PushSubscriber
	for _, sub := range subscribers {
		if sub.ID != subscriberID {
			retained = append(retained, sub)
		}
	}

	if len(retained) == 0 {
		delete(s.subscribers, userID)
	} else {
		s.subscribers[userID] = retained
	}

	logger.Info("push subscriber disconnected", "userID", userID, "subscriberID", subscriberID)
	return nil
}

// BroadcastToUser broadcasts a message to all subscribers of a user
func (s *LocalPushService) BroadcastToUser(userID uuid.UUID, message PushMessage) {
	s.mu.RLock()
	subscribers, exists := s.subscribers[userID]
	s.mu.RUnlock()

	if !exists || len(subscribers) == 0 {
		return
	}

	// Send to all subscribers non-blocking
	for _, subscriber := range subscribers {
		select {
		case subscriber.Chan <- message:
		case <-time.After(1 * time.Second):
			logger.Warn("failed to send push to subscriber (timeout)", "userID", userID, "subscriberID", subscriber.ID)
		}
	}
}

// GetUserTokens retrieves all push tokens for a user
func (s *LocalPushService) GetUserTokens(ctx context.Context, userID uuid.UUID) ([]PushToken, error) {
	var tokens []PushToken
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("failed to get user tokens: %w", err)
	}
	return tokens, nil
}

// Stats returns push service statistics
func (s *LocalPushService) Stats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalMessages := 0
	for _, messages := range s.pushQueue {
		totalMessages += len(messages)
	}

	totalSubscribers := 0
	for _, subscribers := range s.subscribers {
		totalSubscribers += len(subscribers)
	}

	return map[string]interface{}{
		"total_messages":    totalMessages,
		"total_subscribers": totalSubscribers,
		"queued_users":      len(s.pushQueue),
	}
}
