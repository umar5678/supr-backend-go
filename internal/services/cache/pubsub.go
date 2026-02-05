package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func PublishMessage(ctx context.Context, channel string, data interface{}) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return PubSubClient.Publish(ctx, channel, payload).Err()
}

func SubscribeChannel(ctx context.Context, channel string) *redis.PubSub {
	return PubSubClient.Subscribe(ctx, channel)
}

func SubscribeMultiple(ctx context.Context, channels ...string) *redis.PubSub {
	return PubSubClient.Subscribe(ctx, channels...)
}

func PublishToUser(ctx context.Context, userID string, data interface{}) error {
	channel := fmt.Sprintf("user:%s", userID)
	return PublishMessage(ctx, channel, data)
}

func SubscribeToUser(ctx context.Context, userID string) *redis.PubSub {
	channel := fmt.Sprintf("user:%s", userID)
	return SubscribeChannel(ctx, channel)
}

func BroadcastToChannel(ctx context.Context, channel string, data interface{}) error {
	return PublishMessage(ctx, fmt.Sprintf("broadcast:%s", channel), data)
}