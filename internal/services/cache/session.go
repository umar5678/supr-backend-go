package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type PresenceData struct {
	SocketID  string                 `json:"socketId"`
	UserAgent string                 `json:"userAgent"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

func SetSession(ctx context.Context, userID string, sessionData map[string]interface{}) error {
	if redisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}

	data, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("session:%s", userID)
	return redisClient.Set(ctx, key, data, 24*time.Hour).Err()
}

func GetSession(ctx context.Context, userID string) (map[string]interface{}, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	key := fmt.Sprintf("session:%s", userID)
	data, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var session map[string]interface{}
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, err
	}

	return session, nil
}

func DeleteSession(ctx context.Context, userID string) error {
	key := fmt.Sprintf("session:%s", userID)
	return SessionClient.Del(ctx, key).Err()
}

func SessionExists(ctx context.Context, userID string) (bool, error) {
	key := fmt.Sprintf("session:%s", userID)
	count, err := SessionClient.Exists(ctx, key).Result()
	return count > 0, err
}

func RefreshSession(ctx context.Context, userID string, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", userID)
	return SessionClient.Expire(ctx, key, ttl).Err()
}

func SetPresence(ctx context.Context, userID, socketID string, metadata map[string]interface{}) error {
	key := fmt.Sprintf("presence:%s", userID)
	devicesKey := fmt.Sprintf("presence:devices:%s", userID)

	deviceData := PresenceData{
		SocketID:  socketID,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	deviceJSON, err := json.Marshal(deviceData)
	if err != nil {
		return err
	}

	if err := SessionClient.HSet(ctx, devicesKey, socketID, deviceJSON).Err(); err != nil {
		return err
	}

	if err := SessionClient.Expire(ctx, devicesKey, 5*time.Minute).Err(); err != nil {
		return err
	}

	return SessionClient.Set(ctx, key, "online", 5*time.Minute).Err()
}

func RemovePresence(ctx context.Context, userID, socketID string) error {
	devicesKey := fmt.Sprintf("presence:devices:%s", userID)
	presenceKey := fmt.Sprintf("presence:%s", userID)

	if err := SessionClient.HDel(ctx, devicesKey, socketID).Err(); err != nil {
		return err
	}

	count, err := SessionClient.HLen(ctx, devicesKey).Result()
	if err != nil {
		return err
	}

	if count == 0 {
		SessionClient.Del(ctx, presenceKey)
		SessionClient.Del(ctx, devicesKey)
	}

	return nil
}

func IsOnline(ctx context.Context, userID string) (bool, error) {
	key := fmt.Sprintf("presence:%s", userID)
	result, err := SessionClient.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result == "online", nil
}

func GetUserDevices(ctx context.Context, userID string) ([]PresenceData, error) {
	devicesKey := fmt.Sprintf("presence:devices:%s", userID)

	devices, err := SessionClient.HGetAll(ctx, devicesKey).Result()
	if err != nil {
		return nil, err
	}

	var result []PresenceData
	for _, deviceJSON := range devices {
		var device PresenceData
		if err := json.Unmarshal([]byte(deviceJSON), &device); err == nil {
			result = append(result, device)
		}
	}

	return result, nil
}

func GetDeviceCount(ctx context.Context, userID string) (int64, error) {
	devicesKey := fmt.Sprintf("presence:devices:%s", userID)
	return SessionClient.HLen(ctx, devicesKey).Result()
}

func GetAllOnlineUsers(ctx context.Context) ([]string, error) {
	keys, err := SessionClient.Keys(ctx, "presence:*").Result()
	if err != nil {
		return nil, err
	}

	var userIDs []string
	for _, key := range keys {
		if len(key) > 9 {
			userIDs = append(userIDs, key[9:])
		}
	}

	return userIDs, nil
}

func RefreshPresence(ctx context.Context, userID string, ttl time.Duration) error {
	key := fmt.Sprintf("presence:%s", userID)
	devicesKey := fmt.Sprintf("presence:devices:%s", userID)

	SessionClient.Expire(ctx, key, ttl)
	return SessionClient.Expire(ctx, devicesKey, ttl).Err()
}