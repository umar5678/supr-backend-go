package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Service interface for cache operations
type Service interface {
	// String operations
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	// JSON operations
	GetJSON(ctx context.Context, key string, dest interface{}) error
	SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Counter operations
	Increment(ctx context.Context, key string) (int64, error)
	IncrementWithExpiry(ctx context.Context, key string, ttl time.Duration) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)

	// Expiry operations
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
}

type redisService struct {
	client *redis.Client
}

// NewRedisService creates cache service using CacheClient
func NewRedisService() Service {
	return &redisService{client: CacheClient}
}

// Get retrieves string value
func (s *redisService) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

// Set stores string value with TTL
func (s *redisService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes key
func (s *redisService) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

// Exists checks if key exists
func (s *redisService) Exists(ctx context.Context, key string) (bool, error) {
	count, err := s.client.Exists(ctx, key).Result()
	return count > 0, err
}

// GetJSON retrieves and unmarshals JSON value
func (s *redisService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// SetJSON marshals and stores JSON value
func (s *redisService) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return s.client.Set(ctx, key, data, ttl).Err()
}

// Increment increments counter
func (s *redisService) Increment(ctx context.Context, key string) (int64, error) {
	return s.client.Incr(ctx, key).Result()
}

// IncrementWithExpiry increments and sets expiry if new key
func (s *redisService) IncrementWithExpiry(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	val, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if val == 1 {
		s.client.Expire(ctx, key, ttl)
	}
	return val, nil
}

// Decrement decrements counter
func (s *redisService) Decrement(ctx context.Context, key string) (int64, error) {
	return s.client.Decr(ctx, key).Result()
}

// Expire sets expiry on key
func (s *redisService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return s.client.Expire(ctx, key, ttl).Err()
}

// TTL gets remaining TTL for key
func (s *redisService) TTL(ctx context.Context, key string) (time.Duration, error) {
	return s.client.TTL(ctx, key).Result()
}

// Standalone helper functions for direct usage

// SetJSON stores object as JSON (standalone)
func SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return CacheClient.Set(ctx, key, data, ttl).Err()
}

// GetJSON retrieves object from JSON (standalone)
func GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := CacheClient.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Delete removes key (standalone)
func Delete(ctx context.Context, key string) error {
	return CacheClient.Del(ctx, key).Err()
}

// Exists checks if key exists (standalone)
func Exists(ctx context.Context, key string) (bool, error) {
	result, err := CacheClient.Exists(ctx, key).Result()
	return result > 0, err
}

// Get retrieves string value (standalone)
func Get(ctx context.Context, key string) (string, error) {
	return CacheClient.Get(ctx, key).Result()
}

// Set stores value with expiry (standalone)
func Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return CacheClient.Set(ctx, key, value, ttl).Err()
}

// Increment increments counter (standalone)
func Increment(ctx context.Context, key string) (int64, error) {
	return CacheClient.Incr(ctx, key).Result()
}
