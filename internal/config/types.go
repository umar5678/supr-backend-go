package config

import (
	"time"

	"gorm.io/gorm/logger"
)

// Config is the top-level configuration struct.
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Upload   UploadConfig
	Logger   LoggerConfig
}

// AppConfig holds application-level settings.
type AppConfig struct {
	Name        string
	Environment string
	Version     string
	Debug       bool
}

// ServerConfig holds server settings.
type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	CORS            CORSConfig
	RateLimit       RateLimitConfig
}

// CORSConfig holds CORS settings.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	Enabled          bool // NEW: Toggle CORS middleware on/off
}

// RateLimitConfig holds rate limiting settings.
type RateLimitConfig struct {
	RequestsPerSecond int
	Burst             int
}

// DatabaseConfig holds database connection settings.
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

// RedisConfig holds Redis settings - MULTIPLE CLIENTS
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	PoolSize int

	// Database numbers for different purposes
	MainDB    int // DB 0 - General purpose
	CacheDB   int // DB 3 - Caching
	SessionDB int // DB 4 - Sessions & presence
	PubSubDB  int // DB 1 - WebSocket Pub/Sub
}

// JWTConfig holds JWT settings.
type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

// UploadConfig holds file upload settings.
type UploadConfig struct {
	Provider  string
	MaxSize   int64
	S3        S3Config
	LocalPath string
}

// S3Config holds S3-specific settings.
type S3Config struct {
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
}

// LoggerConfig holds logging settings.
type LoggerConfig struct {
	Level    string
	Format   string
	Output   string
	FilePath string
}
