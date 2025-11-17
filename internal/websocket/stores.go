// internal/websocket/stores.go
package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/services/cache"
)

// MessageStore handles persistent message storage
type MessageStore interface {
	Store(ctx context.Context, userID string, message *Message) error
	GetPending(ctx context.Context, userID string, limit int) ([]*Message, error)
	MarkDelivered(ctx context.Context, userID string, messageIDs []string) error
	DeleteOld(ctx context.Context, olderThan time.Duration) error
}

// NotificationStore handles notification persistence
type NotificationStore interface {
	Store(ctx context.Context, userID string, notification interface{}) error
	GetPending(ctx context.Context, userID string) ([]interface{}, error)
	MarkRead(ctx context.Context, userID string, notificationIDs []string) error
	Clear(ctx context.Context, userID string) error
}

// RedisMessageStore implements MessageStore using Redis
type RedisMessageStore struct{}

func NewRedisMessageStore() *RedisMessageStore {
	return &RedisMessageStore{}
}

func (s *RedisMessageStore) Store(ctx context.Context, userID string, message *Message) error {
	key := fmt.Sprintf("messages:pending:%s", userID)

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Store in Redis list with expiry
	score := float64(time.Now().Unix())
	if err := cache.MainClient.ZAdd(ctx, key, cache.Z{
		Score:  score,
		Member: data,
	}).Err(); err != nil {
		return fmt.Errorf("failed to store message: %w", err)
	}

	// Set expiry on the key (7 days)
	cache.MainClient.Expire(ctx, key, 7*24*time.Hour)

	return nil
}

func (s *RedisMessageStore) GetPending(ctx context.Context, userID string, limit int) ([]*Message, error) {
	key := fmt.Sprintf("messages:pending:%s", userID)

	// Get messages from sorted set
	results, err := cache.MainClient.ZRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending messages: %w", err)
	}

	messages := make([]*Message, 0, len(results))
	for _, data := range results {
		var msg Message
		if err := json.Unmarshal([]byte(data), &msg); err == nil {
			messages = append(messages, &msg)
		}
	}

	return messages, nil
}

func (s *RedisMessageStore) MarkDelivered(ctx context.Context, userID string, messageIDs []string) error {
	key := fmt.Sprintf("messages:pending:%s", userID)

	// Remove delivered messages
	for _, msgID := range messageIDs {
		// This is simplified - in production, you'd store message IDs as keys
		cache.MainClient.ZRem(ctx, key, msgID)
	}

	return nil
}

func (s *RedisMessageStore) DeleteOld(ctx context.Context, olderThan time.Duration) error {
	// Scan for all pending message keys and clean old ones
	threshold := time.Now().Add(-olderThan).Unix()

	pattern := "messages:pending:*"
	iter := cache.MainClient.Scan(ctx, 0, pattern, 100).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()
		cache.MainClient.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%d", threshold))
	}

	return iter.Err()
}

// RedisNotificationStore implements NotificationStore using Redis
type RedisNotificationStore struct{}

func NewRedisNotificationStore() *RedisNotificationStore {
	return &RedisNotificationStore{}
}

func (s *RedisNotificationStore) Store(ctx context.Context, userID string, notification interface{}) error {
	key := fmt.Sprintf("notifications:pending:%s", userID)

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Add to list
	if err := cache.MainClient.LPush(ctx, key, data).Err(); err != nil {
		return fmt.Errorf("failed to store notification: %w", err)
	}

	// Limit list size to prevent memory issues
	cache.MainClient.LTrim(ctx, key, 0, 999) // Keep last 1000 notifications

	// Set expiry (30 days)
	cache.MainClient.Expire(ctx, key, 30*24*time.Hour)

	return nil
}

func (s *RedisNotificationStore) GetPending(ctx context.Context, userID string) ([]interface{}, error) {
	key := fmt.Sprintf("notifications:pending:%s", userID)

	// Get all notifications
	results, err := cache.MainClient.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	notifications := make([]interface{}, 0, len(results))
	for _, data := range results {
		var notif map[string]interface{}
		if err := json.Unmarshal([]byte(data), &notif); err == nil {
			notifications = append(notifications, notif)
		}
	}

	return notifications, nil
}

func (s *RedisNotificationStore) MarkRead(ctx context.Context, userID string, notificationIDs []string) error {
	// This is simplified - in production, you'd track read status separately
	key := fmt.Sprintf("notifications:read:%s", userID)

	for _, id := range notificationIDs {
		cache.MainClient.SAdd(ctx, key, id)
	}

	cache.MainClient.Expire(ctx, key, 30*24*time.Hour)
	return nil
}

func (s *RedisNotificationStore) Clear(ctx context.Context, userID string) error {
	key := fmt.Sprintf("notifications:pending:%s", userID)
	return cache.MainClient.Del(ctx, key).Err()
}
