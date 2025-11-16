package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// PublishMessage publishes JSON message to channel
func PublishMessage(ctx context.Context, channel string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return PubSubClient.Publish(ctx, channel, payload).Err()
}

// SubscribeChannel subscribes to Redis channel
func SubscribeChannel(ctx context.Context, channel string) *redis.PubSub {
	return PubSubClient.Subscribe(ctx, channel)
}

// SubscribeMultiple subscribes to multiple channels
func SubscribeMultiple(ctx context.Context, channels ...string) *redis.PubSub {
	return PubSubClient.Subscribe(ctx, channels...)
}

// PublishToUser publishes message to user-specific channel
func PublishToUser(ctx context.Context, userID string, data interface{}) error {
	channel := fmt.Sprintf("user:%s", userID)
	return PublishMessage(ctx, channel, data)
}

// SubscribeToUser subscribes to user-specific channel
func SubscribeToUser(ctx context.Context, userID string) *redis.PubSub {
	channel := fmt.Sprintf("user:%s", userID)
	return SubscribeChannel(ctx, channel)
}

// BroadcastToChannel publishes to broadcast channel
func BroadcastToChannel(ctx context.Context, channel string, data interface{}) error {
	return PublishMessage(ctx, fmt.Sprintf("broadcast:%s", channel), data)
}
