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
	cfg.JWT.AccessExpiry = v.GetDuration("JWT_ACCESS_EXPIRY")
	cfg.JWT.RefreshExpiry = v.GetDuration("JWT_REFRESH_EXPIRY")
	cfg.JWT.Issuer = v.GetString("JWT_ISSUER")

	// Set default JWT expiry times if not provided
	if cfg.JWT.AccessExpiry == 0 {
		cfg.JWT.AccessExpiry = 24 * time.Hour
	}
	if cfg.JWT.RefreshExpiry == 0 {
		cfg.JWT.RefreshExpiry = 7 * 24 * time.Hour
	}

	// Load Upload configuration
	cfg.Upload.Provider = v.GetString("UPLOAD_PROVIDER")
	cfg.Upload.MaxSize = int64(v.GetInt("UPLOAD_MAX_SIZE"))
	cfg.Upload.S3.Bucket = v.GetString("S3_BUCKET")
	cfg.Upload.S3.Region = v.GetString("S3_REGION")
	cfg.Upload.S3.AccessKey = v.GetString("S3_ACCESS_KEY")
	cfg.Upload.S3.SecretKey = v.GetString("S3_SECRET_KEY")

	// Load ImageKit configuration
	cfg.Upload.ImageKit.PublicKey = v.GetString("IMAGEKIT_PUBLIC_KEY")
	cfg.Upload.ImageKit.PrivateKey = v.GetString("IMAGEKIT_PRIVATE_KEY")
	cfg.Upload.ImageKit.URLEndpoint = v.GetString("IMAGEKIT_URL_ENDPOINT")
	cfg.Upload.ImageKit.DocumentsFolder = v.GetString("IMAGEKIT_DOCUMENTS_FOLDER")
	cfg.Upload.ImageKit.BannersFolder = v.GetString("IMAGEKIT_BANNERS_FOLDER")
	cfg.Upload.ImageKit.DocumentsMaxSize = int64(v.GetInt("IMAGEKIT_DOCUMENTS_MAX_SIZE"))
	cfg.Upload.ImageKit.BannersMaxSize = int64(v.GetInt("IMAGEKIT_BANNERS_MAX_SIZE"))

	cfg.Logger.Level = v.GetString("LOG_LEVEL")
	cfg.Logger.Format = v.GetString("LOG_FORMAT")
	cfg.Logger.Output = v.GetString("LOG_OUTPUT")
	cfg.Logger.FilePath = v.GetString("LOG_FILE_PATH")

	// Load Kafka configuration
	brokerStr := v.GetString("KAFKA_BROKERS")
	if brokerStr != "" {
		cfg.Kafka.Brokers = strings.Split(brokerStr, ",")
	} else {
		cfg.Kafka.Brokers = []string{"localhost:9092"}
	}

	cfg.Kafka.SASL.Enabled = v.GetBool("KAFKA_SASL_ENABLED")
	cfg.Kafka.SASL.Mechanism = v.GetString("KAFKA_SASL_MECHANISM")
	cfg.Kafka.SASL.Username = v.GetString("KAFKA_SASL_USERNAME")
	cfg.Kafka.SASL.Password = v.GetString("KAFKA_SASL_PASSWORD")

	cfg.Kafka.Producer.Retries = v.GetInt("KAFKA_PRODUCER_RETRIES")
	cfg.Kafka.Producer.BatchSize = v.GetInt("KAFKA_PRODUCER_BATCH_SIZE")
	cfg.Kafka.Producer.FlushInterval = v.GetDuration("KAFKA_PRODUCER_FLUSH_INTERVAL") * time.Second
	cfg.Kafka.Producer.Timeout = v.GetDuration("KAFKA_PRODUCER_TIMEOUT") * time.Second
	cfg.Kafka.Producer.Compression = v.GetString("KAFKA_PRODUCER_COMPRESSION")

	if cfg.Kafka.Producer.Retries == 0 {
		cfg.Kafka.Producer.Retries = 3
	}
	if cfg.Kafka.Producer.BatchSize == 0 {
		cfg.Kafka.Producer.BatchSize = 100
	}
	if cfg.Kafka.Producer.FlushInterval == 0 {
		cfg.Kafka.Producer.FlushInterval = 1 * time.Second
	}
	if cfg.Kafka.Producer.Timeout == 0 {
		cfg.Kafka.Producer.Timeout = 10 * time.Second
	}

	cfg.Kafka.Consumer.GroupID = v.GetString("KAFKA_CONSUMER_GROUP_ID")
	cfg.Kafka.Consumer.MaxConcurrent = v.GetInt("KAFKA_CONSUMER_MAX_CONCURRENT")
	cfg.Kafka.Consumer.SessionTimeout = v.GetDuration("KAFKA_CONSUMER_SESSION_TIMEOUT") * time.Second
	cfg.Kafka.Consumer.HeartbeatInterval = v.GetDuration("KAFKA_CONSUMER_HEARTBEAT_INTERVAL") * time.Second
	cfg.Kafka.Consumer.CommitInterval = v.GetDuration("KAFKA_CONSUMER_COMMIT_INTERVAL") * time.Second
	cfg.Kafka.Consumer.MaxRetries = v.GetInt("KAFKA_CONSUMER_MAX_RETRIES")
	cfg.Kafka.Consumer.RetryBackoff = v.GetDuration("KAFKA_CONSUMER_RETRY_BACKOFF") * time.Second

	if cfg.Kafka.Consumer.GroupID == "" {
		cfg.Kafka.Consumer.GroupID = "notification-consumer"
	}
	if cfg.Kafka.Consumer.MaxConcurrent == 0 {
		cfg.Kafka.Consumer.MaxConcurrent = 10
	}
	if cfg.Kafka.Consumer.SessionTimeout == 0 {
		cfg.Kafka.Consumer.SessionTimeout = 10 * time.Second
	}
	if cfg.Kafka.Consumer.HeartbeatInterval == 0 {
		cfg.Kafka.Consumer.HeartbeatInterval = 3 * time.Second
	}
	if cfg.Kafka.Consumer.CommitInterval == 0 {
		cfg.Kafka.Consumer.CommitInterval = 1 * time.Second
	}

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
	// HS256 requires at least 32 bytes (256 bits) for strong security
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long for HS256 security")
	}
	if c.JWT.AccessExpiry <= 0 {
		return fmt.Errorf("JWT_ACCESS_EXPIRY must be greater than 0")
	}
	if c.JWT.RefreshExpiry <= 0 {
		return fmt.Errorf("JWT_REFRESH_EXPIRY must be greater than 0")
	}
	if c.JWT.AccessExpiry >= c.JWT.RefreshExpiry {
		return fmt.Errorf("JWT_ACCESS_EXPIRY must be less than JWT_REFRESH_EXPIRY")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Redis.Host == "" {
		return fmt.Errorf("REDIS_HOST is required")
	}
	return nil
}
