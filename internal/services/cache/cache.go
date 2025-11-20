package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// Global Redis clients
var (
	// MainClient    *redis.Client // DB 0 - General purpose
	CacheClient   *redis.Client // DB 3 - Cache operations
	SessionClient *redis.Client // DB 4 - Sessions & presence
	PubSubClient  *redis.Client // DB 1 - WebSocket Pub/Sub
)

// ConnectRedis initializes all Redis clients
func ConnectRedis(cfg *config.RedisConfig) error {
	baseOpts := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	// Main client (DB 0)
	mainOpts := *baseOpts
	mainOpts.DB = cfg.MainDB
	MainClient = redis.NewClient(&mainOpts)

	// Cache client (DB 3)
	cacheOpts := *baseOpts
	cacheOpts.DB = cfg.CacheDB
	CacheClient = redis.NewClient(&cacheOpts)

	// Session client (DB 4)
	sessionOpts := *baseOpts
	sessionOpts.DB = cfg.SessionDB
	SessionClient = redis.NewClient(&sessionOpts)

	// PubSub client (DB 1)
	pubsubOpts := *baseOpts
	pubsubOpts.DB = cfg.PubSubDB
	PubSubClient = redis.NewClient(&pubsubOpts)

	// Test main connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := MainClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis main client connection failed: %w", err)
	}

	logger.Info("redis clients connected successfully",
		"host", cfg.Host,
		"port", cfg.Port,
		"clients", 4,
	)

	return nil
}

// CloseRedis closes all Redis connections gracefully
func CloseRedis() error {
	var errs []error

	if err := MainClient.Close(); err != nil {
		errs = append(errs, fmt.Errorf("main client: %w", err))
	}
	if err := CacheClient.Close(); err != nil {
		errs = append(errs, fmt.Errorf("cache client: %w", err))
	}
	if err := SessionClient.Close(); err != nil {
		errs = append(errs, fmt.Errorf("session client: %w", err))
	}
	if err := PubSubClient.Close(); err != nil {
		errs = append(errs, fmt.Errorf("pubsub client: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("redis close errors: %v", errs)
	}

	logger.Info("redis clients disconnected")
	return nil
}

// HealthCheck checks if Redis is healthy
func HealthCheck(ctx context.Context) error {
	if err := MainClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("main client unhealthy: %w", err)
	}
	if err := PubSubClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("pubsub client unhealthy: %w", err)
	}
	return nil
}
