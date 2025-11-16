package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// PresenceData represents user presence information
type PresenceData struct {
	SocketID  string                 `json:"socketId"`
	UserAgent string                 `json:"userAgent"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// Session Management Functions

// SetSession stores session token for user
func SetSession(ctx context.Context, userID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", userID)
	return SessionClient.Set(ctx, key, token, ttl).Err()
}

// GetSession retrieves session token
func GetSession(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("session:%s", userID)
	return SessionClient.Get(ctx, key).Result()
}

// DeleteSession removes user session
func DeleteSession(ctx context.Context, userID string) error {
	key := fmt.Sprintf("session:%s", userID)
	return SessionClient.Del(ctx, key).Err()
}

// SessionExists checks if session exists
func SessionExists(ctx context.Context, userID string) (bool, error) {
	key := fmt.Sprintf("session:%s", userID)
	count, err := SessionClient.Exists(ctx, key).Result()
	return count > 0, err
}

// RefreshSession extends session TTL
func RefreshSession(ctx context.Context, userID string, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", userID)
	return SessionClient.Expire(ctx, key, ttl).Err()
}

// Presence Management Functions (Multi-Device Support)

// SetPresence marks user as online with device info
func SetPresence(ctx context.Context, userID, socketID string, metadata map[string]interface{}) error {
	key := fmt.Sprintf("presence:%s", userID)
	devicesKey := fmt.Sprintf("presence:devices:%s", userID)

	// Store device info
	deviceData := PresenceData{
		SocketID:  socketID,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	deviceJSON, err := json.Marshal(deviceData)
	if err != nil {
		return err
	}

	// Add device to hash (stores all devices for this user)
	if err := SessionClient.HSet(ctx, devicesKey, socketID, deviceJSON).Err(); err != nil {
		return err
	}

	// Set expiry on devices hash
	if err := SessionClient.Expire(ctx, devicesKey, 5*time.Minute).Err(); err != nil {
		return err
	}

	// Mark user as online
	return SessionClient.Set(ctx, key, "online", 5*time.Minute).Err()
}

// RemovePresence removes device from user's presence
func RemovePresence(ctx context.Context, userID, socketID string) error {
	devicesKey := fmt.Sprintf("presence:devices:%s", userID)
	presenceKey := fmt.Sprintf("presence:%s", userID)

	// Remove this specific device
	if err := SessionClient.HDel(ctx, devicesKey, socketID).Err(); err != nil {
		return err
	}

	// Check if any devices left
	count, err := SessionClient.HLen(ctx, devicesKey).Result()
	if err != nil {
		return err
	}

	// If no devices left, mark as offline
	if count == 0 {
		SessionClient.Del(ctx, presenceKey)
		SessionClient.Del(ctx, devicesKey)
	}

	return nil
}

// IsOnline checks if user has any active connections
func IsOnline(ctx context.Context, userID string) (bool, error) {
	key := fmt.Sprintf("presence:%s", userID)
	result, err := SessionClient.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result == "online", nil
}

// GetUserDevices returns all active devices for user
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

// GetDeviceCount returns number of active devices for user
func GetDeviceCount(ctx context.Context, userID string) (int64, error) {
	devicesKey := fmt.Sprintf("presence:devices:%s", userID)
	return SessionClient.HLen(ctx, devicesKey).Result()
}

// GetAllOnlineUsers returns list of all online user IDs
func GetAllOnlineUsers(ctx context.Context) ([]string, error) {
	keys, err := SessionClient.Keys(ctx, "presence:*").Result()
	if err != nil {
		return nil, err
	}

	var userIDs []string
	for _, key := range keys {
		// Extract userID from key (format: "presence:userId")
		if len(key) > 9 {
			userIDs = append(userIDs, key[9:])
		}
	}

	return userIDs, nil
}

// RefreshPresence extends presence TTL
func RefreshPresence(ctx context.Context, userID string, ttl time.Duration) error {
	key := fmt.Sprintf("presence:%s", userID)
	devicesKey := fmt.Sprintf("presence:devices:%s", userID)

	SessionClient.Expire(ctx, key, ttl)
	return SessionClient.Expire(ctx, devicesKey, ttl).Err()
}
