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

type PushToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"type:varchar(500);uniqueIndex" json:"token"`
	DeviceID  string    `gorm:"type:varchar(255)" json:"device_id"`
	DeviceOS  string    `gorm:"type:varchar(50)" json:"device_os"` 
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (PushToken) TableName() string {
	return "push_tokens"
}

type LocalPushService struct {
	db          *gorm.DB
	notifRepo   repository.NotificationRepository
	wsNotifier  WSNotifier
	mu          sync.RWMutex
	pushQueue   map[uuid.UUID][]PushMessage
	subscribers map[uuid.UUID][]PushSubscriber
}

type PushMessage struct {
	ID        uuid.UUID              `json:"id"`
	UserID    uuid.UUID              `json:"user_id"`
	Title     string                 `json:"title"`
	Body      string                 `json:"body"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Read      bool                   `json:"read"`
}

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

func (s *LocalPushService) SetWSNotifier(ws WSNotifier) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.wsNotifier = ws
}

func (s *LocalPushService) RegisterToken(ctx context.Context, userID uuid.UUID, token, deviceID, deviceOS string) error {
	token_record := &PushToken{
		UserID:   userID,
		Token:    token,
		DeviceID: deviceID,
		DeviceOS: deviceOS,
	}

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

func (s *LocalPushService) UnregisterToken(ctx context.Context, token string) error {
	if err := s.db.WithContext(ctx).Where("token = ?", token).Delete(&PushToken{}).Error; err != nil {
		logger.Error("failed to unregister push token", "error", err)
		return fmt.Errorf("failed to unregister push token: %w", err)
	}
	return nil
}

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

	s.mu.Lock()
	s.pushQueue[userID] = append(s.pushQueue[userID], message)
	s.mu.Unlock()

	dataBytes, _ := json.Marshal(data)
	notification := &models.Notification{
		UserID:   userID,
		Channel:  models.ChannelPush,
		Status:   models.NotificationStatusPending,
		Title:    title,
		Message:  body,
		Metadata: dataBytes,
	}

	logger.Info("attempting to persist notification to database", "userID", userID, "title", title, "channel", models.ChannelPush)
	if err := s.db.WithContext(ctx).Create(notification).Error; err != nil {
		logger.Error("FAILED to persist push notification to database", "error", err, "userID", userID, "title", title)
		return fmt.Errorf("failed to persist notification: %w", err)
	}
	logger.Info("notification successfully persisted to database", "userID", userID, "title", title, "notificationID", notification.ID)

	s.BroadcastToUser(userID, message)

	if err := s.db.WithContext(ctx).Model(notification).Update("status", models.NotificationStatusSent).Error; err != nil {
		logger.Warn("failed to update notification status to sent", "error", err, "notificationID", notification.ID)
	}

	logger.Info("push notification sent (both persisted and broadcasted)", "userID", userID, "title", title)
	return nil
}

func (s *LocalPushService) GetPendingMessages(ctx context.Context, userID uuid.UUID) ([]PushMessage, error) {
	s.mu.RLock()
	messages, exists := s.pushQueue[userID]
	s.mu.RUnlock()

	if !exists {
		return []PushMessage{}, nil
	}

	var unread []PushMessage
	for _, msg := range messages {
		if !msg.Read {
			unread = append(unread, msg)
		}
	}

	return unread, nil
}

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

func (s *LocalPushService) UnsubscribeFromUser(userID uuid.UUID, subscriberID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	subscribers, exists := s.subscribers[userID]
	if !exists {
		return nil
	}

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

func (s *LocalPushService) BroadcastToUser(userID uuid.UUID, message PushMessage) {
	s.mu.RLock()
	ws := s.wsNotifier
	s.mu.RUnlock()

	if ws != nil {
		if err := ws.SendNotification(userID.String(), map[string]interface{}{
			"type":    "notification",
			"title":   message.Title,
			"body":    message.Body,
			"data":    message.Data,
			"id":      message.ID,
			"sent_at": message.Timestamp,
		}); err != nil {
			logger.Warn("ws notification delivery failed", "userID", userID, "error", err)
		} else {
			logger.Info("notification delivered via main websocket", "userID", userID)
		}
	}

	// Also send to any dedicated push subscribers (e.g. /notifications/ws/push)
	s.mu.RLock()
	subscribers, exists := s.subscribers[userID]
	s.mu.RUnlock()

	if !exists || len(subscribers) == 0 {
		return
	}

	for _, subscriber := range subscribers {
		select {
		case subscriber.Chan <- message:
		case <-time.After(1 * time.Second):
			logger.Warn("failed to send push to subscriber (timeout)", "userID", userID, "subscriberID", subscriber.ID)
		}
	}
}

func (s *LocalPushService) GetUserTokens(ctx context.Context, userID uuid.UUID) ([]PushToken, error) {
	var tokens []PushToken
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		return nil, fmt.Errorf("failed to get user tokens: %w", err)
	}
	return tokens, nil
}

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
