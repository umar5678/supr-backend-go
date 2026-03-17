package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"google.golang.org/api/option"
	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/notifications/repository"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

type FCMPushService struct {
	db          *gorm.DB
	notifRepo   repository.NotificationRepository
	fcmClient   *messaging.Client
	mu          sync.RWMutex
	subscribers map[uuid.UUID][]PushSubscriber
	pushQueue   map[uuid.UUID][]PushMessage // kept for Stats() compatibility
}

func NewFCMPushService(
	db *gorm.DB,
	notifRepo repository.NotificationRepository,
	credentialsFile string,
) (*FCMPushService, error) {
	ctx := context.Background()

	var opt option.ClientOption
	if credentialsFile != "" {
		// Load from file path
		opt = option.WithCredentialsFile(credentialsFile)
	} else {
		return nil, fmt.Errorf("firebase credentials file or JSON must be provided")
	}

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %w", err)
	}

	fcmClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get FCM client: %w", err)
	}

	logger.Info("FCM push service initialized successfully")

	return &FCMPushService{
		db:          db,
		notifRepo:   notifRepo,
		fcmClient:   fcmClient,
		subscribers: make(map[uuid.UUID][]PushSubscriber),
		pushQueue:   make(map[uuid.UUID][]PushMessage),
	}, nil
}

// NewFCMPushServiceFromJSON creates a new FCM push service from JSON credentials string
func NewFCMPushServiceFromJSON(
	db *gorm.DB,
	notifRepo repository.NotificationRepository,
	credentialsJSON string,
) (*FCMPushService, error) {
	ctx := context.Background()

	if credentialsJSON == "" {
		return nil, fmt.Errorf("firebase credentials JSON must be provided")
	}

	opt := option.WithCredentialsJSON([]byte(credentialsJSON))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app from JSON: %w", err)
	}

	fcmClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get FCM client: %w", err)
	}

	logger.Info("FCM push service initialized successfully from JSON credentials")

	return &FCMPushService{
		db:          db,
		notifRepo:   notifRepo,
		fcmClient:   fcmClient,
		subscribers: make(map[uuid.UUID][]PushSubscriber),
		pushQueue:   make(map[uuid.UUID][]PushMessage),
	}, nil
}

func (s *FCMPushService) RegisterToken(
	ctx context.Context,
	userID uuid.UUID,
	token, deviceID, deviceOS string,
) error {
	record := &PushToken{
		UserID:   userID,
		Token:    token,
		DeviceID: deviceID,
		DeviceOS: deviceOS,
	}

	result := s.db.WithContext(ctx).
		Where("token = ?", token).
		Assign(record).
		FirstOrCreate(record)

	if result.Error != nil {
		return fmt.Errorf("failed to register push token: %w", result.Error)
	}

	logger.Info("push token registered",
		"userID", userID,
		"deviceOS", deviceOS,
	)
	return nil
}

func (s *FCMPushService) UnregisterToken(ctx context.Context, token string) error {
	if err := s.db.WithContext(ctx).
		Where("token = ?", token).
		Delete(&PushToken{}).Error; err != nil {
		return fmt.Errorf("failed to unregister push token: %w", err)
	}
	return nil
}

func (s *FCMPushService) SendPush(
	ctx context.Context,
	userID uuid.UUID,
	title, body string,
	data map[string]interface{},
) error {
	// 1. Persist notification to database
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
		logger.Error("failed to persist notification",
			"error", err, "userID", userID,
		)
		return fmt.Errorf("failed to persist notification: %w", err)
	}

	// 2. Send via FCM to ALL registered devices for this user
	fcmErr := s.sendViaFCM(ctx, userID, title, body, data)

	// 3. Broadcast to WebSocket subscribers (foreground bonus)
	message := PushMessage{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		Body:      body,
		Data:      data,
		Timestamp: time.Now(),
		Read:      false,
	}
	s.broadcastToUser(userID, message)

	// 4. Update notification status
	if fcmErr != nil {
		logger.Error("FCM send failed, notification persisted",
			"error", fcmErr, "userID", userID,
		)
		s.db.WithContext(ctx).
			Model(notification).
			Update("status", models.NotificationStatusFailed)
		return fcmErr
	}

	s.db.WithContext(ctx).
		Model(notification).
		Update("status", models.NotificationStatusSent)

	logger.Info("push notification sent via FCM",
		"userID", userID, "title", title,
	)
	return nil
}

// ---- The critical missing piece: actually calling FCM ----

func (s *FCMPushService) sendViaFCM(
	ctx context.Context,
	userID uuid.UUID,
	title, body string,
	data map[string]interface{},
) error {
	// Fetch all registered device tokens for this user
	var tokens []PushToken
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&tokens).Error; err != nil {
		return fmt.Errorf("failed to fetch tokens: %w", err)
	}

	if len(tokens) == 0 {
		logger.Debug("no push tokens registered for user",
			"userID", userID,
		)
		return nil // Not an error — user just hasn't registered
	}

	// Convert data map to string map (FCM requirement)
	stringData := make(map[string]string)
	for k, v := range data {
		stringData[k] = fmt.Sprintf("%v", v)
	}

	// Build FCM token list
	registrationTokens := make([]string, 0, len(tokens))
	for _, t := range tokens {
		registrationTokens = append(registrationTokens, t.Token)
	}

	// Send to multiple devices at once
	msg := &messaging.MulticastMessage{
		Tokens: registrationTokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: stringData,
		// Critical for background delivery on Android
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				ChannelID:   "high_importance_channel",
				Priority:    messaging.PriorityHigh,
				Sound:       "default",
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
			},
		},
		// Critical for background delivery on iOS
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority":  "10",
				"apns-push-type": "alert",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: title,
						Body:  body,
					},
					Sound:            "default",
					ContentAvailable: true,
					MutableContent:   true,
				},
			},
		},
	}

	batchResp, err := s.fcmClient.SendEachForMulticast(ctx, msg)
	if err != nil {
		return fmt.Errorf("FCM multicast send failed: %w", err)
	}

	// Clean up invalid tokens
	if batchResp.FailureCount > 0 {
		s.cleanupFailedTokens(ctx, registrationTokens, batchResp.Responses)
	}

	logger.Info("FCM multicast result",
		"userID", userID,
		"success", batchResp.SuccessCount,
		"failure", batchResp.FailureCount,
	)

	return nil
}

func (s *FCMPushService) cleanupFailedTokens(
	ctx context.Context,
	tokens []string,
	responses []*messaging.SendResponse,
) {
	for i, resp := range responses {
		if resp.Error != nil {
			// Remove tokens that are permanently invalid
			if messaging.IsUnregistered(resp.Error) ||
				messaging.IsInvalidArgument(resp.Error) {
				logger.Warn("removing invalid FCM token",
					"token", tokens[i][:20]+"...",
					"error", resp.Error,
				)
				s.db.Where("token = ?", tokens[i]).
					Delete(&PushToken{})
			}
		}
	}
}

// ---- WebSocket subscriber methods (same as your LocalPushService) ----

func (s *FCMPushService) SubscribeToUser(
	userID uuid.UUID,
	subscriberID string,
	msgChan chan PushMessage,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.subscribers[userID] = append(s.subscribers[userID], PushSubscriber{
		ID:   subscriberID,
		Chan: msgChan,
	})
	return nil
}

func (s *FCMPushService) UnsubscribeFromUser(
	userID uuid.UUID,
	subscriberID string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	subs := s.subscribers[userID]
	var retained []PushSubscriber
	for _, sub := range subs {
		if sub.ID != subscriberID {
			retained = append(retained, sub)
		}
	}

	if len(retained) == 0 {
		delete(s.subscribers, userID)
	} else {
		s.subscribers[userID] = retained
	}
	return nil
}

func (s *FCMPushService) broadcastToUser(
	userID uuid.UUID,
	message PushMessage,
) {
	s.mu.RLock()
	subs, exists := s.subscribers[userID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	for _, sub := range subs {
		select {
		case sub.Chan <- message:
		case <-time.After(1 * time.Second):
			logger.Warn("websocket broadcast timeout",
				"userID", userID,
				"subscriberID", sub.ID,
			)
		}
	}
}

func (s *FCMPushService) GetPendingMessages(
	ctx context.Context,
	userID uuid.UUID,
) ([]PushMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	msgs := s.pushQueue[userID]
	var unread []PushMessage
	for _, m := range msgs {
		if !m.Read {
			unread = append(unread, m)
		}
	}
	return unread, nil
}

func (s *FCMPushService) GetUserTokens(
	ctx context.Context,
	userID uuid.UUID,
) ([]PushToken, error) {
	var tokens []PushToken
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&tokens).Error
	return tokens, err
}

func (s *FCMPushService) Stats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalSubs := 0
	for _, subs := range s.subscribers {
		totalSubs += len(subs)
	}

	return map[string]interface{}{
		"total_subscribers": totalSubs,
		"provider":          "fcm",
	}
}
