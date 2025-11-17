package config

// base config, app-root level ,. envs, etc. /internal.config/config.go

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gorm.io/gorm/logger"
)

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config

	// App Config
	cfg.App.Name = v.GetString("APP_NAME")
	cfg.App.Environment = v.GetString("APP_ENV")
	cfg.App.Version = v.GetString("APP_VERSION")
	cfg.App.Debug = v.GetBool("APP_DEBUG")

	// Server Config
	cfg.Server.Host = v.GetString("SERVER_HOST")
	cfg.Server.Port = v.GetInt("SERVER_PORT")
	cfg.Server.ReadTimeout = v.GetDuration("SERVER_READ_TIMEOUT") * time.Second
	cfg.Server.WriteTimeout = v.GetDuration("SERVER_WRITE_TIMEOUT") * time.Second
	cfg.Server.ShutdownTimeout = v.GetDuration("SERVER_SHUTDOWN_TIMEOUT") * time.Second

	// CORS Config
	originsStr := v.GetString("CORS_ALLOWED_ORIGINS")
	if originsStr != "" {
		cfg.Server.CORS.AllowedOrigins = strings.Split(originsStr, ",")
	} else {
		cfg.Server.CORS.AllowedOrigins = []string{"*"}
	}

	methodsStr := v.GetString("CORS_ALLOWED_METHODS")
	if methodsStr != "" {
		cfg.Server.CORS.AllowedMethods = strings.Split(methodsStr, ",")
	} else {
		cfg.Server.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}

	headersStr := v.GetString("CORS_ALLOWED_HEADERS")
	if headersStr != "" {
		cfg.Server.CORS.AllowedHeaders = strings.Split(headersStr, ",")
	} else {
		cfg.Server.CORS.AllowedHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	}

	cfg.Server.CORS.AllowCredentials = v.GetBool("CORS_ALLOW_CREDENTIALS")

	// Rate Limiting
	cfg.Server.RateLimit.RequestsPerSecond = v.GetInt("RATE_LIMIT_REQUESTS_PER_SECOND")
	cfg.Server.RateLimit.Burst = v.GetInt("RATE_LIMIT_BURST")

	if cfg.Server.RateLimit.RequestsPerSecond == 0 {
		cfg.Server.RateLimit.RequestsPerSecond = 100
	}
	if cfg.Server.RateLimit.Burst == 0 {
		cfg.Server.RateLimit.Burst = 200
	}

	// Database Config
	cfg.Database.Host = v.GetString("DB_HOST")
	cfg.Database.Port = v.GetInt("DB_PORT")
	cfg.Database.User = v.GetString("DB_USER")
	cfg.Database.Password = v.GetString("DB_PASSWORD")
	cfg.Database.Name = v.GetString("DB_NAME")
	cfg.Database.SSLMode = v.GetString("DB_SSLMODE")
	cfg.Database.MaxOpenConns = v.GetInt("DB_MAX_OPEN_CONNS")
	cfg.Database.MaxIdleConns = v.GetInt("DB_MAX_IDLE_CONNS")
	cfg.Database.MaxLifetime = v.GetDuration("DB_MAX_LIFETIME") * time.Minute
	cfg.Database.LogLevel = logger.LogLevel(v.GetInt("DB_LOG_LEVEL"))

	// Redis Config (NEW)
	cfg.Redis.Host = v.GetString("REDIS_HOST")
	cfg.Redis.Port = v.GetInt("REDIS_PORT")
	cfg.Redis.Password = v.GetString("REDIS_PASSWORD")
	cfg.Redis.PoolSize = v.GetInt("REDIS_POOL_SIZE")

	// Database numbers
	cfg.Redis.MainDB = v.GetInt("REDIS_MAIN_DB")
	cfg.Redis.CacheDB = v.GetInt("REDIS_CACHE_DB")
	cfg.Redis.SessionDB = v.GetInt("REDIS_SESSION_DB")
	cfg.Redis.PubSubDB = v.GetInt("REDIS_PUBSUB_DB")

	// Defaults for Redis DBs
	if cfg.Redis.PoolSize == 0 {
		cfg.Redis.PoolSize = 10
	}
	if cfg.Redis.MainDB == 0 && cfg.Redis.CacheDB == 0 && cfg.Redis.SessionDB == 0 && cfg.Redis.PubSubDB == 0 {
		cfg.Redis.MainDB = 0
		cfg.Redis.CacheDB = 3
		cfg.Redis.SessionDB = 4
		cfg.Redis.PubSubDB = 1
	}

	// JWT Config
	cfg.JWT.Secret = v.GetString("JWT_SECRET")
	cfg.JWT.AccessExpiry = v.GetDuration("JWT_ACCESS_EXPIRY") * time.Hour
	cfg.JWT.RefreshExpiry = v.GetDuration("JWT_REFRESH_EXPIRY") * time.Hour
	cfg.JWT.Issuer = v.GetString("JWT_ISSUER")

	// Logger Config
	cfg.Logger.Level = v.GetString("LOG_LEVEL")
	cfg.Logger.Format = v.GetString("LOG_FORMAT")
	cfg.Logger.Output = v.GetString("LOG_OUTPUT")
	cfg.Logger.FilePath = v.GetString("LOG_FILE_PATH")

	return &cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("APP_NAME is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Redis.Host == "" {
		return fmt.Errorf("REDIS_HOST is required")
	}
	return nil
}
