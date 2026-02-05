package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

type Service interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)

	GetJSON(ctx context.Context, key string, dest interface{}) error
	SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	Increment(ctx context.Context, key string) (int64, error)
	IncrementWithExpiry(ctx context.Context, key string, ttl time.Duration) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)

	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
}

type redisService struct {
	client *redis.Client
}

var (
	MainClient  *redis.Client
	redisClient *redis.Client
)

func NewRedisService(cfg *config.RedisConfig) (Service, error) {
	options := &redis.Options{
		Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		DisableIndentity: true,
	}

	client := redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	MainClient = client
	redisClient = client

	logger.Info("redis connected successfully",
		"host", cfg.Host,
		"port", cfg.Port,
	)

	return &redisService{client: client}, nil
}

func (s *redisService) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

func (s *redisService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return s.client.Set(ctx, key, value, ttl).Err()
}

func (s *redisService) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

func (s *redisService) Exists(ctx context.Context, key string) (bool, error) {
	count, err := s.client.Exists(ctx, key).Result()
	return count > 0, err
}

func (s *redisService) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (s *redisService) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return s.client.Set(ctx, key, data, ttl).Err()
}

func (s *redisService) Increment(ctx context.Context, key string) (int64, error) {
	return s.client.Incr(ctx, key).Result()
}

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

func (s *redisService) Decrement(ctx context.Context, key string) (int64, error) {
	return s.client.Decr(ctx, key).Result()
}

func (s *redisService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return s.client.Expire(ctx, key, ttl).Err()
}

func (s *redisService) TTL(ctx context.Context, key string) (time.Duration, error) {
	return s.client.TTL(ctx, key).Result()
}

func SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return CacheClient.Set(ctx, key, data, ttl).Err()
}

func GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := CacheClient.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func Delete(ctx context.Context, key string) error {
	return CacheClient.Del(ctx, key).Err()
}

func Exists(ctx context.Context, key string) (bool, error) {
	result, err := CacheClient.Exists(ctx, key).Result()
	return result > 0, err
}

func Get(ctx context.Context, key string) (string, error) {
	return CacheClient.Get(ctx, key).Result()
}

func Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return CacheClient.Set(ctx, key, value, ttl).Err()
}

func Increment(ctx context.Context, key string) (int64, error) {
	return CacheClient.Incr(ctx, key).Result()
}