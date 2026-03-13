package config

import (
	"time"

	"gorm.io/gorm/logger"
)

type Config struct {
	App       AppConfig
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Upload    UploadConfig
	Logger    LoggerConfig
	WebSocket WebSocketConfig
}

type AppConfig struct {
	Name        string
	Environment string
	Version     string
	Debug       bool
}

type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	CORS            CORSConfig
	RateLimit       RateLimitConfig
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	Enabled          bool
}

type RateLimitConfig struct {
	RequestsPerSecond int
	Burst             int
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
	LogLevel     logger.LogLevel
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	PoolSize int

	MainDB    int
	CacheDB   int
	SessionDB int
	PubSubDB  int
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

type UploadConfig struct {
	Provider  string
	MaxSize   int64
	S3        S3Config
	ImageKit  ImageKitConfig
	LocalPath string
}

type S3Config struct {
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
}

type ImageKitConfig struct {
	PublicKey        string
	PrivateKey       string
	URLEndpoint      string
	DocumentsFolder  string
	BannersFolder    string
	DocumentsMaxSize int64
	BannersMaxSize   int64
}

type LoggerConfig struct {
	Level    string
	Format   string
	Output   string
	FilePath string
}
