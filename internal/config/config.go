package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gorm.io/gorm/logger"
)

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

	cfg.App.Name = v.GetString("APP_NAME")
	cfg.App.Environment = v.GetString("APP_ENV")
	cfg.App.Version = v.GetString("APP_VERSION")
	cfg.App.Debug = v.GetBool("APP_DEBUG")

	cfg.Server.Host = v.GetString("SERVER_HOST")
	cfg.Server.Port = v.GetInt("SERVER_PORT")
	cfg.Server.ReadTimeout = v.GetDuration("SERVER_READ_TIMEOUT") * time.Second
	cfg.Server.WriteTimeout = v.GetDuration("SERVER_WRITE_TIMEOUT") * time.Second
	cfg.Server.ShutdownTimeout = v.GetDuration("SERVER_SHUTDOWN_TIMEOUT") * time.Second

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

	cfg.Server.CORS.Enabled = v.GetBool("ENABLE_CORS_MIDDLEWARE")

	cfg.Server.RateLimit.RequestsPerSecond = v.GetInt("RATE_LIMIT_REQUESTS_PER_SECOND")
	cfg.Server.RateLimit.Burst = v.GetInt("RATE_LIMIT_BURST")

	if cfg.Server.RateLimit.RequestsPerSecond == 0 {
		cfg.Server.RateLimit.RequestsPerSecond = 100
	}
	if cfg.Server.RateLimit.Burst == 0 {
		cfg.Server.RateLimit.Burst = 200
	}

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

	cfg.Redis.Host = v.GetString("REDIS_HOST")
	cfg.Redis.Port = v.GetInt("REDIS_PORT")
	cfg.Redis.Password = v.GetString("REDIS_PASSWORD")
	cfg.Redis.PoolSize = v.GetInt("REDIS_POOL_SIZE")

	cfg.Redis.MainDB = v.GetInt("REDIS_MAIN_DB")
	cfg.Redis.CacheDB = v.GetInt("REDIS_CACHE_DB")
	cfg.Redis.SessionDB = v.GetInt("REDIS_SESSION_DB")
	cfg.Redis.PubSubDB = v.GetInt("REDIS_PUBSUB_DB")

	if cfg.Redis.PoolSize == 0 {
		cfg.Redis.PoolSize = 10
	}
	if cfg.Redis.MainDB == 0 && cfg.Redis.CacheDB == 0 && cfg.Redis.SessionDB == 0 && cfg.Redis.PubSubDB == 0 {
		cfg.Redis.MainDB = 0
		cfg.Redis.CacheDB = 3
		cfg.Redis.SessionDB = 4
		cfg.Redis.PubSubDB = 1
	}

	cfg.JWT.Secret = v.GetString("JWT_SECRET")
	cfg.JWT.AccessExpiry = v.GetDuration("JWT_ACCESS_EXPIRY") * time.Hour
	cfg.JWT.RefreshExpiry = v.GetDuration("JWT_REFRESH_EXPIRY") * time.Hour
	cfg.JWT.Issuer = v.GetString("JWT_ISSUER")

	cfg.Logger.Level = v.GetString("LOG_LEVEL")
	cfg.Logger.Format = v.GetString("LOG_FORMAT")
	cfg.Logger.Output = v.GetString("LOG_OUTPUT")
	cfg.Logger.FilePath = v.GetString("LOG_FILE_PATH")

	// Load WebSocket config with defaults
	cfg.WebSocket = DefaultWebSocketConfig()

	// Override with environment variables if provided
	if readBufSize := v.GetInt("WEBSOCKET_READ_BUFFER_SIZE"); readBufSize > 0 {
		cfg.WebSocket.ReadBufferSize = readBufSize
	}
	if writeBufSize := v.GetInt("WEBSOCKET_WRITE_BUFFER_SIZE"); writeBufSize > 0 {
		cfg.WebSocket.WriteBufferSize = writeBufSize
	}
	if maxMsgSize := v.GetInt64("WEBSOCKET_MAX_MESSAGE_SIZE"); maxMsgSize > 0 {
		cfg.WebSocket.MaxMessageSize = maxMsgSize
	}
	if handshakeTimeout := v.GetDuration("WEBSOCKET_HANDSHAKE_TIMEOUT"); handshakeTimeout > 0 {
		cfg.WebSocket.HandshakeTimeout = handshakeTimeout * time.Second
	}
	if writeWait := v.GetDuration("WEBSOCKET_WRITE_WAIT"); writeWait > 0 {
		cfg.WebSocket.WriteWait = writeWait * time.Second
	}
	if pongWait := v.GetDuration("WEBSOCKET_PONG_WAIT"); pongWait > 0 {
		cfg.WebSocket.PongWait = pongWait * time.Second
	}
	if pingPeriod := v.GetDuration("WEBSOCKET_PING_PERIOD"); pingPeriod > 0 {
		cfg.WebSocket.PingPeriod = pingPeriod * time.Second
	}
	if maxConnections := v.GetInt("WEBSOCKET_MAX_CONNECTIONS"); maxConnections > 0 {
		cfg.WebSocket.MaxConnections = maxConnections
	}
	if msgBufSize := v.GetInt("WEBSOCKET_MESSAGE_BUFFER_SIZE"); msgBufSize > 0 {
		cfg.WebSocket.MessageBufferSize = msgBufSize
	}
	cfg.WebSocket.EnablePresence = v.GetBool("WEBSOCKET_ENABLE_PRESENCE")
	cfg.WebSocket.EnableMessageStore = v.GetBool("WEBSOCKET_ENABLE_MESSAGE_STORE")
	cfg.WebSocket.PersistenceEnabled = v.GetBool("WEBSOCKET_PERSISTENCE_ENABLED")
	if persistenceMode := v.GetString("WEBSOCKET_PERSISTENCE_MODE"); persistenceMode != "" {
		cfg.WebSocket.PersistenceMode = persistenceMode
	}
	if rdbInterval := v.GetDuration("WEBSOCKET_RDB_SNAPSHOT_INTERVAL"); rdbInterval > 0 {
		cfg.WebSocket.RDBSnapshotInterval = rdbInterval * time.Second
	}
	if aofSync := v.GetString("WEBSOCKET_AOF_SYNC_POLICY"); aofSync != "" {
		cfg.WebSocket.AOFSyncPolicy = aofSync
	}

	return &cfg, nil
}

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
