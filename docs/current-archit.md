
this is project directory structure: 
go-backend

 ┣ cmd
 ┃ ┣ api
 ┃ ┃ ┗ main.go
 ┃ ┣ migrate
 ┃ ┗ worker
 ┣ deployments
 ┃ ┣ docker
 ┃ ┗ kubernetes
 ┣ docs
 ┣ internal
 ┃ ┣ config
 ┃ ┃ ┣ config.go
 ┃ ┃ ┣ database.go
 ┃ ┃ ┗ types.go
 ┃ ┣ database
 ┃ ┃ ┗ postgres.go
 ┃ ┣ docs
 ┃ ┃ ┣ docs.go
 ┃ ┃ ┣ swagger.json
 ┃ ┃ ┗ swagger.yaml
 ┃ ┣ middleware
 ┃ ┃ ┣ auth.go
 ┃ ┃ ┣ cors.go
 ┃ ┃ ┣ logger.go
 ┃ ┃ ┣ rate_limiter.go
 ┃ ┃ ┣ recovery.go
 ┃ ┃ ┗ request_context.go
 ┃ ┣ models
 ┃ ┃ ┗ user.go
 ┃ ┣ modules
 ┃ ┃ ┣ auth
 ┃ ┃ ┃ ┣ dto
 ┃ ┃ ┃ ┃ ┣ request.go
 ┃ ┃ ┃ ┃ ┗ response.go

 ┃ ┃ ┃ ┣ handler.go

 ┃ ┃ ┃ ┣ repository.go

 ┃ ┃ ┃ ┣ routes.go

 ┃ ┃ ┃ ┗ service.go

 ┃ ┃ ┣ notifications

 ┃ ┃ ┣ posts

 ┃ ┃ ┗ users

 ┃ ┃ ┃ ┗ dto

 ┃ ┣ repository

 ┃ ┣ services

 ┃ ┃ ┣ cache

 ┃ ┃ ┃ ┗ redis.go

 ┃ ┃ ┣ email

 ┃ ┃ ┣ queue

 ┃ ┃ ┗ upload

 ┃ ┣ utils

 ┃ ┃ ┣ helpers

 ┃ ┃ ┣ jwt

 ┃ ┃ ┃ ┗ jwt.go

 ┃ ┃ ┣ logger

 ┃ ┃ ┃ ┗ logger.go

 ┃ ┃ ┣ password

 ┃ ┃ ┃ ┗ hash.go

 ┃ ┃ ┣ response

 ┃ ┃ ┃ ┣ api_response.old

 ┃ ┃ ┃ ┣ error.go

 ┃ ┃ ┃ ┣ pagination.go

 ┃ ┃ ┃ ┗ response.go

 ┃ ┃ ┗ validator

 ┃ ┗ websocket

 ┃ ┃ ┣ events

 ┃ ┃ ┣ handlers

 ┃ ┃ ┣ middleware

 ┃ ┃ ┣ client.go

 ┃ ┃ ┗ hub.go

 ┣ migrations

 ┃ ┣ 000001_create_user_table.down.sql

 ┃ ┗ 000001_create_user_table.up.sql

 ┣ pkg

 ┃ ┣ constants

 ┃ ┣ errors

 ┃ ┗ pagination

 ┣ scripts

 ┣ test

 ┃ ┣ e2e

 ┃ ┗ integration

 ┣ tmp

 ┃ ┗ main.exe

 ┣ .air.toml

 ┣ .env

 ┣ .gitignore

 ┣ go.mod

 ┣ go.sum

 ┣ Makefile

 ┗ README.md

### core files 

cmd/api/main.go: 

```go
package main

  

import (

    "context"

    "fmt"

    "net/http"

    "os"

    "os/signal"

    "syscall"

    "time"

  

    "github.com/gin-gonic/gin"

    "github.com/joho/godotenv" // Add this import

    swaggerFiles "github.com/swaggo/files"

    ginSwagger "github.com/swaggo/gin-swagger"

  

    "github.com/umar5678/go-backend/internal/config"

    "github.com/umar5678/go-backend/internal/database"

  

    _ "github.com/umar5678/go-backend/internal/docs" // Import generated docs

    "github.com/umar5678/go-backend/internal/middleware"

    "github.com/umar5678/go-backend/internal/modules/auth"

  

    // "github.com/umar5678/go-backend/internal/services/cache"

    "github.com/umar5678/go-backend/internal/utils/logger"

)

  

// @title Your Project API

// @version 1.0

// @description Production-grade Go backend API

// @termsOfService http://swagger.io/terms/

  

// @contact.name API Support

// @contact.email support@yourproject.com

  

// @license.name MIT

// @license.url https://opensource.org/licenses/MIT

  

// @host localhost:8080

// @BasePath /api/v1

  

// @securityDefinitions.apikey BearerAuth

// @in header

// @name Authorization

// @description Type "Bearer" followed by a space and JWT token.

  

func main() {

    // Load .env file if it exists (ignore error for production where env vars are set externally)

    _ = godotenv.Load()

  

    // Load configuration

    cfg, err := config.LoadConfig()

    if err != nil {

        fmt.Printf("Failed to load config: %v\n", err)

        os.Exit(1)

    }

  

    // Validate configuration

    if err := cfg.Validate(); err != nil {

        fmt.Printf("Invalid configuration: %v\n", err)

        os.Exit(1)

    }

  

    // Initialize logger

    if err := logger.Initialize(&cfg.Logger); err != nil {

        fmt.Printf("Failed to initialize logger: %v\n", err)

        os.Exit(1)

    }

    defer logger.Sync()

  

    logger.Info("starting application",

        "name", cfg.App.Name,

        "environment", cfg.App.Environment,

        "version", cfg.App.Version,

    )

  

    // Connect to database

    db, err := database.ConnectPostgres(&cfg.Database)

    if err != nil {

        logger.Fatal("failed to connect to database", "error", err)

    }

    defer database.Close(db)

  

    // Connect to Redis

    // cacheService, err := cache.NewRedisService(&cfg.Redis)

    // if err != nil {

    //  logger.Fatal("failed to connect to redis", "error", err)

    // }

  

    // Setup Gin

    if cfg.App.Environment == "production" {

        gin.SetMode(gin.ReleaseMode)

    }

  

    router := gin.New()

  

    // Global middleware

    router.Use(middleware.RequestContext(cfg.App.Version))

    router.Use(middleware.Logger())

    router.Use(middleware.Recovery())

    router.Use(middleware.ErrorHandler())

    router.Use(middleware.CORS(cfg.Server.CORS))

  

    // Health check endpoints

    router.GET("/health", healthCheck)

    router.GET("/ready", readyCheck(db))

  

    // API routes

    v1 := router.Group("/api/v1")

    {

        // Rate limiter for API routes

        v1.Use(middleware.RateLimit(cfg.Server.RateLimit))

  

        // Auth module

        authRepo := auth.NewRepository(db)

        authService := auth.NewService(authRepo, cfg)

        authHandler := auth.NewHandler(authService)

        authMiddleware := middleware.Auth(cfg)

        auth.RegisterRoutes(v1, authHandler, authMiddleware)

  

        // Add other modules here...

    }

  

    // Swagger documentation

    router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

  

    // Start server

    srv := &http.Server{

        Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),

        Handler:      router,

        ReadTimeout:  cfg.Server.ReadTimeout,

        WriteTimeout: cfg.Server.WriteTimeout,

    }

  

    // Graceful shutdown

    go func() {

        logger.Info("server starting",

            "host", cfg.Server.Host,

            "port", cfg.Server.Port,

        )

  

        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {

            logger.Fatal("failed to start server", "error", err)

        }

    }()

  

    // Wait for interrupt signal

    quit := make(chan os.Signal, 1)

    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    <-quit

  

    logger.Info("shutting down server...")

  

    ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)

    defer cancel()

  

    if err := srv.Shutdown(ctx); err != nil {

        logger.Error("server forced to shutdown", "error", err)

    }

  

    logger.Info("server stopped gracefully")

}

  

func healthCheck(c *gin.Context) {

    c.JSON(http.StatusOK, gin.H{

        "status": "ok",

        "time":   time.Now().Format(time.RFC3339),

    })

}

  

func readyCheck(db interface{}) gin.HandlerFunc {

    return func(c *gin.Context) {

        // Add actual health checks here

        c.JSON(http.StatusOK, gin.H{

            "status":   "ready",

            "database": "connected",

            "time":     time.Now().Format(time.RFC3339),

        })

    }

}
```


internal/config/config.go: 

```go
package config

  

import (

    "fmt"

    "time"

  

    "github.com/spf13/viper"

    "gorm.io/gorm/logger" // For LogLevel if loading from env

)

  

// LoadConfig loads configuration from environment variables and config files

func LoadConfig() (*Config, error) {

    v := viper.New()

  

    // Set config file paths

    v.SetConfigName("config")

    v.SetConfigType("yaml")

    v.AddConfigPath(".")

    v.AddConfigPath("./config")

  

    // Auto-bind environment variables

    v.AutomaticEnv()

  

    // Read config file (optional)

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

    cfg.Database.LogLevel = logger.LogLevel(v.GetInt("DB_LOG_LEVEL")) // Optional: Add this for env support (defaults to 0/Silent if missing)

  

    // Redis Config

    cfg.Redis.Host = v.GetString("REDIS_HOST")

    cfg.Redis.Port = v.GetInt("REDIS_PORT")

    cfg.Redis.Password = v.GetString("REDIS_PASSWORD")

    cfg.Redis.DB = v.GetInt("REDIS_DB")

    cfg.Redis.PoolSize = v.GetInt("REDIS_POOL_SIZE")

  

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

    return nil

}
```

internal/config/database/go

```go
package config

  

import (

    "fmt"

    "time"

  

    "gorm.io/driver/postgres"

    "gorm.io/gorm"

    "gorm.io/gorm/logger"

)

  

// ConnectDatabase creates and configures database connection

func ConnectDatabase(cfg DatabaseConfig) (*gorm.DB, error) {

    dsn := fmt.Sprintf(

        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",

        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode, // Updated Port to %d, Name

    )

  

    // Configure GORM

    gormConfig := &gorm.Config{

        Logger: logger.Default.LogMode(cfg.LogLevel),

        NowFunc: func() time.Time {

            return time.Now().UTC()

        },

        PrepareStmt:            true,

        SkipDefaultTransaction: true,

    }

  

    db, err := gorm.Open(postgres.Open(dsn), gormConfig)

    if err != nil {

        return nil, fmt.Errorf("failed to connect to database: %w", err)

    }

  

    // Get underlying SQL DB

    sqlDB, err := db.DB()

    if err != nil {

        return nil, fmt.Errorf("failed to get database instance: %w", err)

    }

  

    // Connection pool settings

    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

    sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)

  

    // Test connection

    if err := sqlDB.Ping(); err != nil {

        return nil, fmt.Errorf("failed to ping database: %w", err)

    }

  

    return db, nil

}

  

// CloseDatabase gracefully closes database connection

func CloseDatabase(db *gorm.DB) error {

    sqlDB, err := db.DB()

    if err != nil {

        return err

    }

    return sqlDB.Close()

}

  

// DefaultDatabaseConfig returns default database configuration

func DefaultDatabaseConfig() DatabaseConfig {

    return DatabaseConfig{

        Host:         "localhost",

        Port:         5432, // Default PostgreSQL port

        User:         "postgres",

        Password:     "postgres",

        Name:         "postgres",

        SSLMode:      "disable",

        MaxOpenConns: 25,

        MaxIdleConns: 5,

        MaxLifetime:  5 * time.Minute,

        LogLevel:     logger.Warn,

    }

}
```

internal/config/types.go

```go
package config

  

import (

    "time"

  

    "gorm.io/gorm/logger"

)

  

// Config is the top-level configuration struct.

type Config struct {

    App      AppConfig

    Server   ServerConfig

    Database DatabaseConfig

    Redis    RedisConfig

    JWT      JWTConfig

    Upload   UploadConfig

    Logger   LoggerConfig

}

  

// AppConfig holds application-level settings.

type AppConfig struct {

    Name        string

    Environment string

    Version     string

    Debug       bool

}

  

// ServerConfig holds server settings.

type ServerConfig struct {

    Host            string

    Port            int

    ReadTimeout     time.Duration

    WriteTimeout    time.Duration

    ShutdownTimeout time.Duration

    CORS            CORSConfig

    RateLimit       RateLimitConfig

}

  

// CORSConfig holds CORS settings.

type CORSConfig struct {

    AllowedOrigins   []string

    AllowedMethods   []string

    AllowedHeaders   []string

    AllowCredentials bool

}

  

// RateLimitConfig holds rate limiting settings.

type RateLimitConfig struct {

    RequestsPerSecond int

    Burst             int

}

  

// DatabaseConfig holds database connection settings.

type DatabaseConfig struct {

    Host         string

    Port         int

    User         string

    Password     string

    Name         string // Standardized from DBName/Name

    SSLMode      string

    MaxOpenConns int

    MaxIdleConns int

    MaxLifetime  time.Duration

    LogLevel     logger.LogLevel

}

  

// RedisConfig holds Redis settings.

type RedisConfig struct {

    Host     string

    Port     int

    Password string

    DB       int

    PoolSize int

}

  

// JWTConfig holds JWT settings.

type JWTConfig struct {

    Secret        string

    AccessExpiry  time.Duration

    RefreshExpiry time.Duration

    Issuer        string

}

  

// UploadConfig holds file upload settings.

type UploadConfig struct {

    Provider  string // "s3", "local"

    MaxSize   int64

    S3        S3Config

    LocalPath string

}

  

// S3Config holds S3-specific settings.

type S3Config struct {

    Bucket    string

    Region    string

    AccessKey string

    SecretKey string

}

  

// LoggerConfig holds logging settings.

type LoggerConfig struct {

    Level    string // "debug", "info", "warn", "error"

    Format   string // "json", "text"

    Output   string // "stdout", "file"

    FilePath string

}
```
internal/database/postgres.go


```go
package database

  

import (

    "fmt"

    "time"

  

    "github.com/umar5678/go-backend/internal/config"

    "github.com/umar5678/go-backend/internal/utils/logger"

  

    "gorm.io/driver/postgres"

    "gorm.io/gorm"

    gormlogger "gorm.io/gorm/logger"

)

  

func ConnectPostgres(cfg *config.DatabaseConfig) (*gorm.DB, error) {

    dsn := fmt.Sprintf(

        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",

        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,

    )

  

    // Configure GORM logger

    gormLogger := gormlogger.Default.LogMode(gormlogger.Silent)

  

    // if cfg.LogQueries  {

    //  gormLogger = gormlogger.Default.LogMode(gormlogger.Info)

    // }

  

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{

        Logger: gormLogger,

        NowFunc: func() time.Time {

            return time.Now().UTC()

        },

    })

    if err != nil {

        return nil, fmt.Errorf("failed to connect to database: %w", err)

    }

  

    sqlDB, err := db.DB()

    if err != nil {

        return nil, fmt.Errorf("failed to get database instance: %w", err)

    }

  

    // Connection pool settings

    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

    sqlDB.SetConnMaxLifetime(cfg.MaxLifetime)

  

    // Test connection

    if err := sqlDB.Ping(); err != nil {

        return nil, fmt.Errorf("failed to ping database: %w", err)

    }

  

    logger.Info("database connected successfully",

        "host", cfg.Host,

        "database", cfg.Name,

    )

  

    return db, nil

}

  

// Close closes the database connection

func Close(db *gorm.DB) error {

    sqlDB, err := db.DB()

    if err != nil {

        return err

    }

    return sqlDB.Close()

}
```
internal/middleware/auth.go

```go
package middleware

  

import (

    "strings"

  

    "github.com/umar5678/go-backend/internal/config"

    "github.com/umar5678/go-backend/internal/utils/jwt"

    "github.com/umar5678/go-backend/internal/utils/response"

  

    "github.com/gin-gonic/gin"

)

  

func Auth(cfg *config.Config) gin.HandlerFunc {

    return func(c *gin.Context) {

        // Extract token from header

        authHeader := c.GetHeader("Authorization")

        if authHeader == "" {

            c.Error(response.UnauthorizedError("Authorization header required"))

            c.Abort()

            return

        }

  

        // Check Bearer prefix

        parts := strings.SplitN(authHeader, " ", 2)

        if len(parts) != 2 || parts[0] != "Bearer" {

            c.Error(response.UnauthorizedError("Invalid authorization header format"))

            c.Abort()

            return

        }

  

        tokenString := parts[1]

  

        // Verify token

        claims, err := jwt.Verify(tokenString, cfg.JWT.Secret)

        if err != nil {

            c.Error(response.UnauthorizedError("Invalid or expired token"))

            c.Abort()

            return

        }

  

        // Set user info in context

        c.Set("userID", claims.UserID)

        c.Set("role", claims.Role)

  

        c.Next()

    }

}

  

// RequireRole checks if user has required role

func RequireRole(roles ...string) gin.HandlerFunc {

    return func(c *gin.Context) {

        userRole, exists := c.Get("role")

        if !exists {

            c.Error(response.ForbiddenError("Insufficient permissions"))

            c.Abort()

            return

        }

  

        roleStr := userRole.(string)

        for _, role := range roles {

            if roleStr == role {

                c.Next()

                return

            }

        }

  

        c.Error(response.ForbiddenError("Insufficient permissions"))

        c.Abort()

    }

}
```
internal/middleware/cors.go

```go
package middleware

  

import (

    "github.com/gin-gonic/gin"

    "github.com/umar5678/go-backend/internal/config"

)

  

func CORS(cfg config.CORSConfig) gin.HandlerFunc {

    return func(c *gin.Context) {

        // Set CORS headers

        origin := c.GetHeader("Origin")

        if isAllowedOrigin(origin, cfg.AllowedOrigins) {

            c.Header("Access-Control-Allow-Origin", origin)

        }

  

        c.Header("Access-Control-Allow-Methods", joinStrings(cfg.AllowedMethods, ", "))

        c.Header("Access-Control-Allow-Headers", joinStrings(cfg.AllowedHeaders, ", "))

  

        if cfg.AllowCredentials {

            c.Header("Access-Control-Allow-Credentials", "true")

        }

  

        // Handle preflight

        if c.Request.Method == "OPTIONS" {

            c.AbortWithStatus(204)

            return

        }

  

        c.Next()

    }

}

  

func isAllowedOrigin(origin string, allowed []string) bool {

    for _, o := range allowed {

        if o == "*" || o == origin {

            return true

        }

    }

    return false

}

  

func joinStrings(strs []string, sep string) string {

    result := ""

    for i, s := range strs {

        if i > 0 {

            result += sep

        }

        result += s

    }

    return result

}
```
internal/middleware/logger.go

```go
package middleware

  

import (

    "time"

  

    "github.com/gin-gonic/gin"

    "github.com/umar5678/go-backend/internal/utils/logger"

)

  

func Logger() gin.HandlerFunc {

    return func(c *gin.Context) {

        start := time.Now()

        path := c.Request.URL.Path

        query := c.Request.URL.RawQuery

  

        c.Next()

  

        duration := time.Since(start)

        statusCode := c.Writer.Status()

  

        requestID, _ := c.Get("requestID")

  

        logger.Info("request completed",

            "requestID", requestID,

            "method", c.Request.Method,

            "path", path,

            "query", query,

            "status", statusCode,

            "duration", duration.Milliseconds(),

            "ip", c.ClientIP(),

            "userAgent", c.Request.UserAgent(),

        )

    }

}
```
internal/middleware/rate_limiter.go

```go
package middleware

  

import (

    "fmt"

  

    "github.com/umar5678/go-backend/internal/config"

    "github.com/umar5678/go-backend/internal/utils/response"

  

    "github.com/gin-gonic/gin"

    "golang.org/x/time/rate"

)

  

var limiters = make(map[string]*rate.Limiter)

  

func RateLimit(cfg config.RateLimitConfig) gin.HandlerFunc {

    return func(c *gin.Context) {

        ip := c.ClientIP()

  

        limiter, exists := limiters[ip]

        if !exists {

            limiter = rate.NewLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.Burst)

            limiters[ip] = limiter

        }

  

        if !limiter.Allow() {

            c.Error(response.TooManyRequests("Rate limit exceeded"))

            c.Abort()

            return

        }

  

        c.Next()

    }

}

  

// RateLimitByKey creates a rate limiter based on custom key

func RateLimitByKey(keyFunc func(*gin.Context) string, requestsPerSecond int, burst int) gin.HandlerFunc {

    limiters := make(map[string]*rate.Limiter)

  

    return func(c *gin.Context) {

        key := keyFunc(c)

  

        limiter, exists := limiters[key]

        if !exists {

            limiter = rate.NewLimiter(rate.Limit(requestsPerSecond), burst)

            limiters[key] = limiter

        }

  

        if !limiter.Allow() {

            c.Error(response.TooManyRequests(fmt.Sprintf("Rate limit exceeded for key: %s", key)))

            c.Abort()

            return

        }

  

        c.Next()

    }

}
```
internal/middleware/recovery.go

```go
package middleware

  

import (

    "github.com/umar5678/go-backend/internal/utils/logger"

    "github.com/umar5678/go-backend/internal/utils/response"

  

    "github.com/gin-gonic/gin"

)

  

// Recovery handles panics and converts them to proper error responses

func Recovery() gin.HandlerFunc {

    return func(c *gin.Context) {

        defer func() {

            if err := recover(); err != nil {

                requestID, _ := c.Get("requestID")

  

                logger.Error("panic recovered",

                    "requestID", requestID,

                    "error", err,

                    "path", c.Request.URL.Path,

                    "method", c.Request.Method,

                )

  

                // Check if it's an AppError

                if appErr, ok := err.(*response.AppError); ok {

                    response.SendError(c, appErr.StatusCode, appErr.Message, appErr.Errors, appErr.Code)

                    return

                }

  

                // Generic error

                response.InternalError(c, "An unexpected error occurred")

                c.Abort()

            }

        }()

  

        c.Next()

  

        // Handle errors set in context

        if len(c.Errors) > 0 {

            err := c.Errors.Last()

  

            requestID, _ := c.Get("requestID")

            logger.Error("request error",

                "requestID", requestID,

                "error", err.Error(),

                "path", c.Request.URL.Path,

            )

  

            // Check if it's an AppError

            if appErr, ok := err.Err.(*response.AppError); ok {

                if appErr.Internal != nil {

                    logger.Error("internal error details",

                        "requestID", requestID,

                        "internal", appErr.Internal.Error(),

                    )

                }

                response.SendError(c, appErr.StatusCode, appErr.Message, appErr.Errors, appErr.Code)

                return

            }

  

            // Default error response

            response.InternalError(c, err.Error())

        }

    }

}

  

// ErrorHandler is a simpler error handling middleware

func ErrorHandler() gin.HandlerFunc {

    return func(c *gin.Context) {

        c.Next()

  

        // Check if response already sent

        if c.Writer.Written() {

            return

        }

  

        // Check for errors in context

        if len(c.Errors) > 0 {

            err := c.Errors.Last().Err

  

            // Handle AppError

            if appErr, ok := err.(*response.AppError); ok {

                response.SendError(c, appErr.StatusCode, appErr.Message, appErr.Errors, appErr.Code)

                return

            }

  

            // Generic error

            response.InternalError(c, "An error occurred")

        }

    }

}
```
internal/middleware/request_context.go

```go
package middleware

  

import (

    "time"

  

    "github.com/gin-gonic/gin"

    "github.com/google/uuid"

)

  

// RequestContext adds request ID and timing to context

func RequestContext(version string) gin.HandlerFunc {

    return func(c *gin.Context) {

        // Generate or extract request ID

        requestID := c.GetHeader("X-Request-ID")

        if requestID == "" {

            requestID = uuid.New().String()

        }

  

        // Add to context

        c.Set("requestID", requestID)

        c.Set("startTime", time.Now())

        c.Set("version", version)

  

        // Add to response headers

        c.Header("X-Request-ID", requestID)

  

        c.Next()

    }

}
```
internal/models/user.go

```go
package models

  

import (

    "time"

  

    "gorm.io/gorm"

)

  

type User struct {

    ID          string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`

    Name        string         `gorm:"type:varchar(255);not null" json:"name"`

    Email       string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`

    Password    string         `gorm:"type:varchar(255);not null" json:"-"`

    Role        string         `gorm:"type:varchar(50);default:'user'" json:"role"`

    Status      string         `gorm:"type:varchar(50);default:'active'" json:"status"`

    LastLoginAt *time.Time     `json:"lastLoginAt,omitempty"`

    CreatedAt   time.Time      `gorm:"autoCreateTime" json:"createdAt"`

    UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`

    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

}

  

func (User) TableName() string {

    return "users"

}
```
internal/services/cache/redis.go

```go
package cache

  

import (

    "context"

    "fmt"

    "time"

  

    "github.com/umar5678/go-backend/internal/config"

    "github.com/umar5678/go-backend/internal/utils/logger"

  

    "github.com/redis/go-redis/v9"

)

  

type Service interface {

    Get(ctx context.Context, key string) (string, error)

    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

    Delete(ctx context.Context, key string) error

    Exists(ctx context.Context, key string) (bool, error)

    Increment(ctx context.Context, key string) (int64, error)

    Expire(ctx context.Context, key string, ttl time.Duration) error

}

  

type redisService struct {

    client *redis.Client

}

  

func NewRedisService(cfg *config.RedisConfig) (Service, error) {

    client := redis.NewClient(&redis.Options{

        Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),

        Password: cfg.Password,

        DB:       cfg.DB,

        PoolSize: cfg.PoolSize,

    })

  

    // Test connection

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

    defer cancel()

  

    if err := client.Ping(ctx).Err(); err != nil {

        return nil, fmt.Errorf("failed to connect to Redis: %w", err)

    }

  

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

  

func (s *redisService) Increment(ctx context.Context, key string) (int64, error) {

    return s.client.Incr(ctx, key).Result()

}

  

func (s *redisService) Expire(ctx context.Context, key string, ttl time.Duration) error {

    return s.client.Expire(ctx, key, ttl).Err()

}

  

func (s *redisService) Close() error {

    return s.client.Close()

}
```
internal/utils/jwt/jwt.go

```go
package jwt

  

import (

    "errors"

    "time"

  

    "github.com/golang-jwt/jwt/v5"

)

  

type Claims struct {

    UserID string `json:"userId"`

    Email  string `json:"email"`

    Role   string `json:"role"`

    jwt.RegisteredClaims

}

  

func Generate(userID, email, role, secret string, expiry time.Duration) (string, error) {

    claims := Claims{

        UserID: userID,

        Email:  email,

        Role:   role,

        RegisteredClaims: jwt.RegisteredClaims{

            ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),

            IssuedAt:  jwt.NewNumericDate(time.Now()),

        },

    }

  

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    return token.SignedString([]byte(secret))

}

  

func Verify(tokenString, secret string) (*Claims, error) {

    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {

        return []byte(secret), nil

    })

  

    if err != nil {

        return nil, err

    }

  

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {

        return claims, nil

    }

  

    return nil, errors.New("invalid token")

}

  

// package jwt

  

// import (

//  "errors"

//  "time"

  

//  "github.com/golang-jwt/jwt/v5"

// )

  

// type Claims struct {

//  UserID string `json:"userId"`

//  Role   string `json:"role"`

//  jwt.RegisteredClaims

// }

  

// // GenerateAccessToken creates a new access token

// func GenerateAccessToken(userID, secret string, expiry time.Duration) (string, error) {

//  claims := Claims{

//      UserID: userID,

//      RegisteredClaims: jwt.RegisteredClaims{

//          ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),

//          IssuedAt:  jwt.NewNumericDate(time.Now()),

//          NotBefore: jwt.NewNumericDate(time.Now()),

//      },

//  }

  

//  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

//  return token.SignedString([]byte(secret))

// }

  

// // GenerateRefreshToken creates a new refresh token

// func GenerateRefreshToken(userID, secret string, expiry time.Duration) (string, error) {

//  claims := Claims{

//      UserID: userID,

//      RegisteredClaims: jwt.RegisteredClaims{

//          ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),

//          IssuedAt:  jwt.NewNumericDate(time.Now()),

//      },

//  }

  

//  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

//  return token.SignedString([]byte(secret))

// }

  

// // VerifyToken validates and parses a JWT token

// func VerifyToken(tokenString, secret string) (*Claims, error) {

//  token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {

//      return []byte(secret), nil

//  })

  

//  if err != nil {

//      return nil, err

//  }

  

//  if claims, ok := token.Claims.(*Claims); ok && token.Valid {

//      return claims, nil

//  }

  

//  return nil, errors.New("invalid token")

// }
```

internal/utils/logger/logger.go

```go
package logger

  

import (

    "github.com/umar5678/go-backend/internal/config"

    "go.uber.org/zap"

    "go.uber.org/zap/zapcore"

)

  

var log *zap.Logger

  

// Initialize sets up the logger

func Initialize(cfg *config.LoggerConfig) error {

    var zapConfig zap.Config

  

    if cfg.Format == "json" {

        zapConfig = zap.NewProductionConfig()

    } else {

        zapConfig = zap.NewDevelopmentConfig()

    }

  

    // Set log level

    level, err := zapcore.ParseLevel(cfg.Level)

    if err != nil {

        level = zapcore.InfoLevel

    }

    zapConfig.Level = zap.NewAtomicLevelAt(level)

  

    // Configure output

    if cfg.Output == "file" && cfg.FilePath != "" {

        zapConfig.OutputPaths = []string{cfg.FilePath}

        zapConfig.ErrorOutputPaths = []string{cfg.FilePath}

    } else {

        zapConfig.OutputPaths = []string{"stdout"}

        zapConfig.ErrorOutputPaths = []string{"stderr"}

    }

  

    logger, err := zapConfig.Build()

    if err != nil {

        return err

    }

  

    log = logger

    return nil

}

  

// Get returns the logger instance

func Get() *zap.Logger {

    if log == nil {

        // Fallback logger

        log, _ = zap.NewProduction()

    }

    return log

}

  

// Helper functions for common log levels

func Debug(msg string, fields ...interface{}) {

    Get().Sugar().Debugw(msg, fields...)

}

  

func Info(msg string, fields ...interface{}) {

    Get().Sugar().Infow(msg, fields...)

}

  

func Warn(msg string, fields ...interface{}) {

    Get().Sugar().Warnw(msg, fields...)

}

  

func Error(msg string, fields ...interface{}) {

    Get().Sugar().Errorw(msg, fields...)

}

  

func Fatal(msg string, fields ...interface{}) {

    Get().Sugar().Fatalw(msg, fields...)

}

  

// Sync flushes any buffered log entries

func Sync() {

    if log != nil {

        log.Sync()

    }

}
```
internal/utils/password/hash.go

```go
package password

  

import (

    "crypto/rand"

    "encoding/base64"

    "errors"

    "fmt"

  

    "golang.org/x/crypto/bcrypt"

)

  

const (

    // MinPasswordLength defines minimum password length

    MinPasswordLength = 8

  

    // MaxPasswordLength defines maximum password length

    MaxPasswordLength = 128

  

    // DefaultCost is bcrypt default cost (14 = ~1 second on modern hardware)

    DefaultCost = 14

)

  

// Hash generates bcrypt hash from password

func Hash(password string) (string, error) {

    // Validate password length

    if len(password) < MinPasswordLength {

        return "", fmt.Errorf("password must be at least %d characters", MinPasswordLength)

    }

    if len(password) > MaxPasswordLength {

        return "", fmt.Errorf("password must not exceed %d characters", MaxPasswordLength)

    }

  

    bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)

    if err != nil {

        return "", fmt.Errorf("failed to hash password: %w", err)

    }

  

    return string(bytes), nil

}

  

// HashWithCost generates bcrypt hash with custom cost

func HashWithCost(password string, cost int) (string, error) {

    if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {

        return "", fmt.Errorf("cost must be between %d and %d", bcrypt.MinCost, bcrypt.MaxCost)

    }

  

    bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)

    if err != nil {

        return "", fmt.Errorf("failed to hash password: %w", err)

    }

  

    return string(bytes), nil

}

  

// Verify checks if password matches hash

func Verify(password, hash string) bool {

    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

    return err == nil

}

  

// VerifyWithError checks password and returns detailed error

func VerifyWithError(password, hash string) error {

    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

    if err != nil {

        if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {

            return errors.New("invalid password")

        }

        return fmt.Errorf("password verification failed: %w", err)

    }

    return nil

}

  

// Strength checks password strength

type PasswordStrength int

  

const (

    StrengthWeak PasswordStrength = iota

    StrengthFair

    StrengthGood

    StrengthStrong

)

  

// CheckStrength evaluates password strength

func CheckStrength(password string) PasswordStrength {

    length := len(password)

  

    hasUpper := false

    hasLower := false

    hasDigit := false

    hasSpecial := false

  

    for _, char := range password {

        switch {

        case char >= 'A' && char <= 'Z':

            hasUpper = true

        case char >= 'a' && char <= 'z':

            hasLower = true

        case char >= '0' && char <= '9':

            hasDigit = true

        default:

            hasSpecial = true

        }

    }

  

    score := 0

    if length >= 8 {

        score++

    }

    if length >= 12 {

        score++

    }

    if hasUpper && hasLower {

        score++

    }

    if hasDigit {

        score++

    }

    if hasSpecial {

        score++

    }

  

    switch {

    case score <= 2:

        return StrengthWeak

    case score == 3:

        return StrengthFair

    case score == 4:

        return StrengthGood

    default:

        return StrengthStrong

    }

}

  

// ValidatePassword performs comprehensive password validation

func ValidatePassword(password string) error {

    if len(password) < MinPasswordLength {

        return fmt.Errorf("password must be at least %d characters long", MinPasswordLength)

    }

  

    if len(password) > MaxPasswordLength {

        return fmt.Errorf("password must not exceed %d characters", MaxPasswordLength)

    }

  

    strength := CheckStrength(password)

    if strength == StrengthWeak {

        return errors.New("password is too weak, must contain uppercase, lowercase, numbers, and special characters")

    }

  

    return nil

}

  

// GenerateRandomPassword generates cryptographically secure random password

func GenerateRandomPassword(length int) (string, error) {

    if length < MinPasswordLength {

        length = MinPasswordLength

    }

    if length > MaxPasswordLength {

        length = MaxPasswordLength

    }

  

    bytes := make([]byte, length)

    if _, err := rand.Read(bytes); err != nil {

        return "", fmt.Errorf("failed to generate random password: %w", err)

    }

  

    return base64.URLEncoding.EncodeToString(bytes)[:length], nil

}

  

// NeedsRehash checks if password hash needs to be updated

func NeedsRehash(hash string) bool {

    cost, err := bcrypt.Cost([]byte(hash))

    if err != nil {

        return true

    }

    return cost < DefaultCost

}
```

internal/utils/response/error.go: 

```go
package response

  

import (

    "fmt"

    "net/http"

  

    "github.com/gin-gonic/gin"

)

  

// ErrorDetail represents validation or field-specific error

type ErrorDetail struct {

    Field   string `json:"field,omitempty"`

    Message string `json:"message"`

    Code    string `json:"code,omitempty"`

}

  

// AppError represents application error with context

type AppError struct {

    StatusCode int

    Message    string

    Code       string

    Errors     []ErrorDetail

    Internal   error // Not exposed to client, used for logging

}

  

// Error implements error interface

func (e *AppError) Error() string {

    if e.Internal != nil {

        return fmt.Sprintf("%s: %v", e.Message, e.Internal)

    }

    return e.Message

}

  

// NewAppError creates new application error

func NewAppError(statusCode int, message, code string, errors []ErrorDetail, internal error) *AppError {

    return &AppError{

        StatusCode: statusCode,

        Message:    message,

        Code:       code,

        Errors:     errors,

        Internal:   internal,

    }

}

  

// ToResponse converts AppError to a Response (for easier integration)

func (e *AppError) ToResponse(c *gin.Context) {

    SendError(c, e.StatusCode, e.Message, e.Errors, e.Code) // Uses SendError func from response.go

}

  

// Common error constructors

  

// BadRequest creates 400 error

func BadRequest(message string, errors ...ErrorDetail) *AppError {

    if message == "" {

        message = "Bad request"

    }

    return NewAppError(http.StatusBadRequest, message, "BAD_REQUEST", errors, nil)

}

  

// UnauthorizedError creates 401 error

func UnauthorizedError(message string) *AppError {

    if message == "" {

        message = "Unauthorized"

    }

    return NewAppError(http.StatusUnauthorized, message, "UNAUTHORIZED", nil, nil)

}

  

// ForbiddenError creates 403 error

func ForbiddenError(message string) *AppError {

    if message == "" {

        message = "Forbidden"

    }

    return NewAppError(http.StatusForbidden, message, "FORBIDDEN", nil, nil)

}

  

// NotFoundError creates 404 error

func NotFoundError(resource string) *AppError {

    message := "Resource not found"

    if resource != "" {

        message = fmt.Sprintf("%s not found", resource)

    }

    return NewAppError(http.StatusNotFound, message, "NOT_FOUND", nil, nil)

}

  

// ConflictError creates 409 error

func ConflictError(message string) *AppError {

    if message == "" {

        message = "Resource conflict"

    }

    return NewAppError(http.StatusConflict, message, "CONFLICT", nil, nil)

}

  

// NewValidationAppError creates 422 error

func NewValidationAppError(message string, errors []ErrorDetail) *AppError {

    if message == "" {

        message = "Validation failed"

    }

    return NewAppError(http.StatusUnprocessableEntity, message, "VALIDATION_ERROR", errors, nil)

}

  

// InternalServerError creates 500 error

func InternalServerError(message string, internal error) *AppError {

    if message == "" {

        message = "Internal server error"

    }

    return NewAppError(http.StatusInternalServerError, message, "INTERNAL_ERROR", nil, internal)

}

  

// TooManyRequests creates 429 error

func TooManyRequests(message string) *AppError {

    if message == "" {

        message = "Too many requests"

    }

    return NewAppError(http.StatusTooManyRequests, message, "RATE_LIMIT_EXCEEDED", nil, nil)

}

  

// ServiceUnavailable creates 503 error

func ServiceUnavailable(message string) *AppError {

    if message == "" {

        message = "Service unavailable"

    }

    return NewAppError(http.StatusServiceUnavailable, message, "SERVICE_UNAVAILABLE", nil, nil)

}

  

// NewValidationErrorDetail creates single validation error

func NewValidationErrorDetail(field, message string) ErrorDetail {

    return ErrorDetail{

        Field:   field,

        Message: message,

        Code:    "INVALID_FIELD",

    }

}

  

// NewErrorDetail creates generic error

func NewErrorDetail(message, code string) ErrorDetail {

    return ErrorDetail{

        Message: message,

        Code:    code,

    }

}
```


internal/utils/response/pagination.go

```go
package response

  

import "github.com/gin-gonic/gin"

  

// PaginationMeta represents pagination metadata

type PaginationMeta struct {

    Total      int64 `json:"total"`

    Page       int   `json:"page"`

    Limit      int   `json:"limit"`

    TotalPages int   `json:"totalPages"`

    HasNext    bool  `json:"hasNext"`

    HasPrev    bool  `json:"hasPrev"`

}

  

// NewPaginationMeta creates pagination metadata

func NewPaginationMeta(total int64, page, limit int) PaginationMeta {

    totalPages := int(total) / limit

    if int(total)%limit > 0 {

        totalPages++

    }

  

    return PaginationMeta{

        Total:      total,

        Page:       page,

        Limit:      limit,

        TotalPages: totalPages,

        HasNext:    page < totalPages,

        HasPrev:    page > 1,

    }

}

  

// Paginated sends a paginated response

func Paginated(c *gin.Context, data interface{}, pagination PaginationMeta, message string) {

    resp := Response{

        Success: true,

        Message: message,

        Data:    data,

        Meta:    extractMeta(c), // From response.go

    }

  

    // Add pagination to meta (or as separate field)

    c.Set("pagination", pagination)

  

    c.JSON(200, gin.H{

        "success":    resp.Success,

        "message":    resp.Message,

        "data":       resp.Data,

        "meta":       resp.Meta,

        "pagination": pagination,

    })

}
```

internal/utils/response/response.go

```go
package response

  

import (

    "time"

  

    "github.com/gin-gonic/gin"

)

  

// Response represents the standard API response structure

type Response struct {

    Success bool          `json:"success"`

    Message string        `json:"message"`

    Data    interface{}   `json:"data,omitempty"`

    Errors  []ErrorDetail `json:"errors,omitempty"` // Use ErrorDetail

    Meta    Meta          `json:"meta"`

    Code    string        `json:"code,omitempty"`

}

  

// Meta contains request metadata

type Meta struct {

    RequestID  string  `json:"requestId"`

    Timestamp  string  `json:"timestamp"`

    DurationMs float64 `json:"durationMs,omitempty"`

    Path       string  `json:"path"`

    Method     string  `json:"method"`

    Version    string  `json:"version,omitempty"`

}

  

// Success sends a successful response

func Success(c *gin.Context, data interface{}, message string, code ...string) {

    statusCode := 200

    if c.Request.Method == "POST" {

        statusCode = 201

    }

  

    resp := Response{

        Success: true,

        Message: message,

        Data:    data,

        Meta:    extractMeta(c),

    }

  

    if len(code) > 0 {

        resp.Code = code[0]

    }

  

    c.JSON(statusCode, resp)

}

  

// SendError sends an error response

func SendError(c *gin.Context, statusCode int, message string, errors []ErrorDetail, code ...string) {

    resp := Response{

        Success: false,

        Message: message,

        Errors:  errors,

        Meta:    extractMeta(c),

    }

  

    if len(code) > 0 {

        resp.Code = code[0]

    }

  

    c.JSON(statusCode, resp)

}

  

// ValidationError sends a validation error response

func ValidationError(c *gin.Context, errors []ErrorDetail) {

    SendError(c, 422, "Validation failed", errors, "VALIDATION_ERROR")

}

  

// NotFound sends a 404 response

func NotFound(c *gin.Context, message string) {

    SendError(c, 404, message, nil, "NOT_FOUND")

}

  

// Unauthorized sends a 401 response

func Unauthorized(c *gin.Context, message string) {

    SendError(c, 401, message, nil, "UNAUTHORIZED")

}

  

// Forbidden sends a 403 response

func Forbidden(c *gin.Context, message string) {

    SendError(c, 403, message, nil, "FORBIDDEN")

}

  

// InternalError sends a 500 response

func InternalError(c *gin.Context, message string) {

    SendError(c, 500, message, nil, "INTERNAL_ERROR")

}

  

// extractMeta extracts metadata from the context

func extractMeta(c *gin.Context) Meta {

    requestID, _ := c.Get("requestID")

    startTime, exists := c.Get("startTime")

  

    meta := Meta{

        RequestID: requestID.(string),

        Timestamp: time.Now().UTC().Format(time.RFC3339),

        Path:      c.Request.URL.Path,

        Method:    c.Request.Method,

    }

  

    if exists {

        duration := time.Since(startTime.(time.Time))

        meta.DurationMs = float64(duration.Microseconds()) / 1000

    }

  

    version, exists := c.Get("version")

    if exists {

        meta.Version = version.(string)

    }

  

    return meta

}
```


.air.toml

```toml
# Air configuration for hot reload

root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  # Paths to watch
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main.exe ./cmd/api"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "migrations", "docs"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_error = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = true

[screen]
  clear_on_rebuild = true
  keep_scroll = true
```

.env: 

```env
# Application

APP_NAME=go-backend

APP_ENV=development

APP_VERSION=1.0.0

APP_DEBUG=true

  

# Server

SERVER_HOST=0.0.0.0

SERVER_PORT=8080

SERVER_READ_TIMEOUT=30

SERVER_WRITE_TIMEOUT=30

SERVER_SHUTDOWN_TIMEOUT=5

  

# Database

DB_HOST=localhost

DB_PORT=5432

DB_USER=go_backend_admin

DB_PASSWORD=goPass

DB_NAME=go_backend

DB_SSLMODE=disable

DB_MAX_OPEN_CONNS=25

DB_MAX_IDLE_CONNS=5

DB_MAX_LIFETIME=5

  

# Redis

REDIS_HOST=localhost

REDIS_PORT=6379

REDIS_PASSWORD=

REDIS_DB=0

REDIS_POOL_SIZE=10

  

# JWT

JWT_SECRET=your-secret-key-change-this-in-production

JWT_ACCESS_EXPIRY=15m

JWT_REFRESH_EXPIRY=168h

JWT_ISSUER=your-project

  

# Logger

LOG_LEVEL=info

LOG_FORMAT=json

LOG_OUTPUT=stdout

LOG_FILE_PATH=./logs/app.log

  

# CORS

CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001

CORS_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS

CORS_ALLOWED_HEADERS=Origin,Content-Type,Accept,Authorization

CORS_ALLOW_CREDENTIALS=true

  

# Rate Limiting

RATE_LIMIT_REQUESTS_PER_SECOND=10

RATE_LIMIT_BURST=20

  

# File Upload

UPLOAD_PROVIDER=s3

UPLOAD_MAX_SIZE=10485760

S3_BUCKET=your-bucket

S3_REGION=us-east-1

S3_ACCESS_KEY=

S3_SECRET_KEY=
```

.env.example

```env
# Application

APP_NAME=go-backend

APP_ENV=development

APP_VERSION=1.0.0

APP_DEBUG=true

  

# Server

SERVER_HOST=0.0.0.0

SERVER_PORT=8080

SERVER_READ_TIMEOUT=30

SERVER_WRITE_TIMEOUT=30

SERVER_SHUTDOWN_TIMEOUT=5

  

# Database

DB_HOST=localhost

DB_PORT=5432

DB_USER=go_backend_admin

DB_PASSWORD=goPass

DB_NAME=go_backend

DB_SSLMODE=disable

DB_MAX_OPEN_CONNS=25

DB_MAX_IDLE_CONNS=5

DB_MAX_LIFETIME=5

  

# Redis

REDIS_HOST=localhost

REDIS_PORT=6379

REDIS_PASSWORD=

REDIS_DB=0

REDIS_POOL_SIZE=10

  

# JWT

JWT_SECRET=your-secret-key-change-this-in-production

JWT_ACCESS_EXPIRY=15m

JWT_REFRESH_EXPIRY=168h

JWT_ISSUER=your-project

  

# Logger

LOG_LEVEL=info

LOG_FORMAT=json

LOG_OUTPUT=stdout

LOG_FILE_PATH=./logs/app.log

  

# CORS

CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001

CORS_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS

CORS_ALLOWED_HEADERS=Origin,Content-Type,Accept,Authorization

CORS_ALLOW_CREDENTIALS=true

  

# Rate Limiting

RATE_LIMIT_REQUESTS_PER_SECOND=10

RATE_LIMIT_BURST=20

  

# File Upload

UPLOAD_PROVIDER=s3

UPLOAD_MAX_SIZE=10485760

S3_BUCKET=your-bucket

S3_REGION=us-east-1

S3_ACCESS_KEY=

S3_SECRET_KEY=
```

go.mod: 

```mod
module github.com/umar5678/go-backend

  

go 1.25.3

  

require (

    github.com/gin-gonic/gin v1.11.0

    github.com/golang-jwt/jwt/v5 v5.3.0

    github.com/google/uuid v1.6.0

    github.com/gorilla/websocket v1.5.3

    github.com/redis/go-redis/v9 v9.16.0

    github.com/spf13/viper v1.21.0

    github.com/swaggo/files v1.0.1

    github.com/swaggo/gin-swagger v1.6.1

    go.uber.org/zap v1.27.0

    golang.org/x/crypto v0.44.0

    golang.org/x/time v0.14.0

    gorm.io/driver/postgres v1.6.0

    gorm.io/gorm v1.31.1

)

  

require (

    github.com/KyleBanks/depth v1.2.1 // indirect

    github.com/PuerkitoBio/purell v1.1.1 // indirect

    github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect

    github.com/bytedance/sonic v1.14.0 // indirect

    github.com/bytedance/sonic/loader v0.3.0 // indirect

    github.com/cespare/xxhash/v2 v2.3.0 // indirect

    github.com/cloudwego/base64x v0.1.6 // indirect

    github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect

    github.com/fsnotify/fsnotify v1.9.0 // indirect

    github.com/gabriel-vasile/mimetype v1.4.8 // indirect

    github.com/gin-contrib/sse v1.1.0 // indirect

    github.com/go-openapi/jsonpointer v0.19.5 // indirect

    github.com/go-openapi/jsonreference v0.19.6 // indirect

    github.com/go-openapi/spec v0.20.4 // indirect

    github.com/go-openapi/swag v0.19.15 // indirect

    github.com/go-playground/locales v0.14.1 // indirect

    github.com/go-playground/universal-translator v0.18.1 // indirect

    github.com/go-playground/validator/v10 v10.27.0 // indirect

    github.com/go-viper/mapstructure/v2 v2.4.0 // indirect

    github.com/goccy/go-json v0.10.2 // indirect

    github.com/goccy/go-yaml v1.18.0 // indirect

    github.com/jackc/pgpassfile v1.0.0 // indirect

    github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect

    github.com/jackc/pgx/v5 v5.7.6 // indirect

    github.com/jackc/puddle/v2 v2.2.2 // indirect

    github.com/jinzhu/inflection v1.0.0 // indirect

    github.com/jinzhu/now v1.1.5 // indirect

    github.com/joho/godotenv v1.5.1 // direct

    github.com/josharian/intern v1.0.0 // indirect

    github.com/json-iterator/go v1.1.12 // indirect

    github.com/klauspost/cpuid/v2 v2.3.0 // indirect

    github.com/leodido/go-urn v1.4.0 // indirect

    github.com/mailru/easyjson v0.7.6 // indirect

    github.com/mattn/go-isatty v0.0.20 // indirect

    github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect

    github.com/modern-go/reflect2 v1.0.2 // indirect

    github.com/pelletier/go-toml/v2 v2.2.4 // indirect

    github.com/quic-go/qpack v0.5.1 // indirect

    github.com/quic-go/quic-go v0.54.0 // indirect

    github.com/sagikazarmark/locafero v0.11.0 // indirect

    github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect

    github.com/spf13/afero v1.15.0 // indirect

    github.com/spf13/cast v1.10.0 // indirect

    github.com/spf13/pflag v1.0.10 // indirect

    github.com/subosito/gotenv v1.6.0 // indirect

    github.com/swaggo/swag v1.8.12 // indirect

    github.com/twitchyliquid64/golang-asm v0.15.1 // indirect

    github.com/ugorji/go/codec v1.3.0 // indirect

    go.uber.org/mock v0.5.0 // indirect

    go.uber.org/multierr v1.10.0 // indirect

    go.yaml.in/yaml/v3 v3.0.4 // indirect

    golang.org/x/arch v0.20.0 // indirect

    golang.org/x/mod v0.29.0 // indirect

    golang.org/x/net v0.46.0 // indirect

    golang.org/x/sync v0.18.0 // indirect

    golang.org/x/sys v0.38.0 // indirect

    golang.org/x/text v0.31.0 // indirect

    golang.org/x/tools v0.38.0 // indirect

    google.golang.org/protobuf v1.36.9 // indirect

    gopkg.in/yaml.v2 v2.4.0 // indirect

)
```

Makefile

```md
# Makefile for Go project

  

.PHONY: help build run test clean migrate-up migrate-down swagger docker-build docker-run

  

# Variables

APP_NAME := go-backend  # Change if your project name differs

VERSION := $(shell git describe --tags --always --dirty)

BUILD_DIR := ./bin

MAIN_PATH := ./cmd/api

MIGRATION_PATH := ./migrations

  

# Go parameters

GOCMD := go

GOBUILD := $(GOCMD) build

GOTEST := $(GOCMD) test

GOMOD := $(GOCMD) mod

GORUN := $(GOCMD) run

  

help: ## Show this help message

    @echo "Usage: make [target]"

    @echo ""

    @echo "Available targets:"

    @grep -E '^[a-zA-Z_-]+:.*?## .*$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $1, $2}'

  

install: ## Install dependencies

    $(GOMOD) download

    $(GOMOD) tidy

    @echo "Installing tools..."

    go install github.com/swaggo/swag/cmd/swag@latest

    go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

    go install github.com/air-verse/air@latest  # For hot reload

  

build: ## Build the application

    @echo "Building $(APP_NAME)..."

    $(GOBUILD) -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

    @echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

  

run: ## Run the application

    $(GORUN) $(MAIN_PATH)/main.go

  

dev: ## Run in development mode with hot reload (requires air)

    air

  

# test: ## Run tests

#   $(GOTEST) -v -race -coverprofile=coverage.out ./...

#   $(GOCMD) tool cover -html=coverage.out -o coverage.html

  

# test-integration: ## Run integration tests

#   $(GOTEST) -v -tags=integration ./test/integration/...

  

# lint: ## Run linter

#   golangci-lint run

  

swagger: ## Generate Swagger documentation

    swag init -g cmd/api/main.go -o internal/docs

    @echo "Swagger docs generated"

  

migrate-create: ## Create a new migration (use name=migration_name)

    migrate create -ext sql -dir $(MIGRATION_PATH) -seq $(name)

  

migrate-up: ## Run database migrations

    migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" up

  

migrate-down: ## Rollback database migrations

    migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" down

  

migrate-force: ## Force migration version (use version=N)

    migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" force $(version)

  

clean: ## Clean build artifacts

    rm -rf $(BUILD_DIR)

    rm -rf coverage.out coverage.html

    @echo "Clean complete"

  

fmt: ## Format code

    go fmt ./...

    goimports -w .

  

vet: ## Run go vet

    go vet ./...

  

mod-update: ## Update dependencies

    $(GOMOD) get -u ./...

    $(GOMOD) tidy

  

security-scan: ## Run security scan

    gosec ./...

  

bench: ## Run benchmarks

    $(GOTEST) -bench=. -benchmem ./...

  

coverage: ## Generate coverage report

    $(GOTEST) -coverprofile=coverage.out ./...

    $(GOCMD) tool cover -func=coverage.out

  

all: clean install swagger build test ## Run all build steps

  

.DEFAULT_GOAL := help
```


### Modules

#### Auth module: 

internal/modules/auth/dto/request.go

```go
package dto

  

import (

    "errors"

    "regexp"

)

  

type RegisterRequest struct {

    Name     string `json:"name" binding:"required,min=2,max=255"`

    Email    string `json:"email" binding:"required,email"`

    Password string `json:"password" binding:"required,min=8"`

}

  

func (r *RegisterRequest) Validate() error {

    if !isValidEmail(r.Email) {

        return errors.New("invalid email format")

    }

    if len(r.Password) < 8 {

        return errors.New("password must be at least 8 characters")

    }

    return nil

}

  

type LoginRequest struct {

    Email    string `json:"email" binding:"required,email"`

    Password string `json:"password" binding:"required"`

}

  

func (r *LoginRequest) Validate() error {

    if !isValidEmail(r.Email) {

        return errors.New("invalid email format")

    }

    return nil

}

  

type RefreshRequest struct {

    RefreshToken string `json:"refreshToken" binding:"required"`

}

  

func isValidEmail(email string) bool {

    re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

    return re.MatchString(email)

}
```


internal/modules/auth/dto/response.go

```go
package dto

  

import (

    "time"

  

    "github.com/umar5678/go-backend/internal/models"

)

  

type AuthResponse struct {

    User         *UserResponse `json:"user"`

    AccessToken  string        `json:"accessToken"`

    RefreshToken string        `json:"refreshToken"`

    ExpiresIn    int           `json:"expiresIn"`

}

  

type UserResponse struct {

    ID        string    `json:"id"`

    Name      string    `json:"name"`

    Email     string    `json:"email"`

    Role      string    `json:"role"`

    Status    string    `json:"status"`

    CreatedAt time.Time `json:"createdAt"`

    UpdatedAt time.Time `json:"updatedAt"`

}

  

func ToUserResponse(user *models.User) *UserResponse {

    return &UserResponse{

        ID:        user.ID,

        Name:      user.Name,

        Email:     user.Email,

        Role:      user.Role,

        Status:    user.Status,

        CreatedAt: user.CreatedAt,

        UpdatedAt: user.UpdatedAt,

    }

}
```


internal/modules/auth/handler.go

```go
package auth

  

import (

    "github.com/umar5678/go-backend/internal/modules/auth/dto"

    "github.com/umar5678/go-backend/internal/utils/response"

  

    "github.com/gin-gonic/gin"

)

  

type Handler struct {

    service Service

}

  

func NewHandler(service Service) *Handler {

    return &Handler{service: service}

}

  

// Register godoc

// @Summary Register new user

// @Tags auth

// @Accept json

// @Produce json

// @Param request body dto.RegisterRequest true "Register"

// @Success 201 {object} response.Response{data=dto.AuthResponse}

// @Router /auth/register [post]

func (h *Handler) Register(c *gin.Context) {

    var req dto.RegisterRequest

    if err := c.ShouldBindJSON(&req); err != nil {

        c.Error(response.BadRequest("Invalid request"))

        return

    }

  

    result, err := h.service.Register(c.Request.Context(), req)

    if err != nil {

        c.Error(err)

        return

    }

  

    response.Success(c, result, "Registration successful")

}

  

// Login godoc

// @Summary User login

// @Tags auth

// @Accept json

// @Produce json

// @Param request body dto.LoginRequest true "Login"

// @Success 200 {object} response.Response{data=dto.AuthResponse}

// @Router /auth/login [post]

func (h *Handler) Login(c *gin.Context) {

    var req dto.LoginRequest

    if err := c.ShouldBindJSON(&req); err != nil {

        c.Error(response.BadRequest("Invalid request"))

        return

    }

  

    result, err := h.service.Login(c.Request.Context(), req)

    if err != nil {

        c.Error(err)

        return

    }

  

    response.Success(c, result, "Login successful")

}

  

// GetProfile godoc

// @Summary Get user profile

// @Tags auth

// @Security BearerAuth

// @Produce json

// @Success 200 {object} response.Response{data=dto.UserResponse}

// @Router /auth/profile [get]

func (h *Handler) GetProfile(c *gin.Context) {

    userID, _ := c.Get("userID")

  

    user, err := h.service.GetProfile(c.Request.Context(), userID.(string))

    if err != nil {

        c.Error(err)

        return

    }

  

    response.Success(c, user, "Profile retrieved")

}

  

// RefreshToken godoc

// @Summary Refresh access token

// @Tags auth

// @Accept json

// @Produce json

// @Param request body dto.RefreshRequest true "Refresh"

// @Success 200 {object} response.Response{data=dto.AuthResponse}

// @Router /auth/refresh [post]

func (h *Handler) RefreshToken(c *gin.Context) {

    var req dto.RefreshRequest

    if err := c.ShouldBindJSON(&req); err != nil {

        c.Error(response.BadRequest("Invalid request"))

        return

    }

  

    result, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)

    if err != nil {

        c.Error(err)

        return

    }

  

    response.Success(c, result, "Token refreshed")

}
```


internal/modules/auth/repository.go

```go
package auth

  

import (

    "context"

  

    "github.com/umar5678/go-backend/internal/models"

  

    "gorm.io/gorm"

)

  

type Repository interface {

    Create(ctx context.Context, user *models.User) error

    FindByEmail(ctx context.Context, email string) (*models.User, error)

    FindByID(ctx context.Context, id string) (*models.User, error)

    UpdateLastLogin(ctx context.Context, id string) error

}

  

type repository struct {

    db *gorm.DB

}

  

func NewRepository(db *gorm.DB) Repository {

    return &repository{db: db}

}

  

func (r *repository) Create(ctx context.Context, user *models.User) error {

    return r.db.WithContext(ctx).Create(user).Error

}

  

func (r *repository) FindByEmail(ctx context.Context, email string) (*models.User, error) {

    var user models.User

    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error

    return &user, err

}

  

func (r *repository) FindByID(ctx context.Context, id string) (*models.User, error) {

    var user models.User

    err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error

    return &user, err

}

  

func (r *repository) UpdateLastLogin(ctx context.Context, id string) error {

    return r.db.WithContext(ctx).Model(&models.User{}).

        Where("id = ?", id).

        Update("last_login_at", gorm.Expr("NOW()")).Error

}
```


internal/modules/auth/routes.go

```go
package auth

  

import (

    "github.com/gin-gonic/gin"

)

  

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {

    auth := router.Group("/auth")

    {

        auth.POST("/register", handler.Register)

        auth.POST("/login", handler.Login)

        auth.POST("/refresh", handler.RefreshToken)

  

        // Protected

        auth.Use(authMiddleware)

        auth.GET("/profile", handler.GetProfile)

    }

}
```


internal/modules/auth/service.go

```go
package auth

  

import (

    "context"

    "errors"

    "time"

  

    "github.com/umar5678/go-backend/internal/config"

    "github.com/umar5678/go-backend/internal/models"

    "github.com/umar5678/go-backend/internal/modules/auth/dto"

    "github.com/umar5678/go-backend/internal/utils/jwt"

    "github.com/umar5678/go-backend/internal/utils/password"

    "github.com/umar5678/go-backend/internal/utils/response"

  

    "gorm.io/gorm"

)

  

type Service interface {

    Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)

    Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)

    RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)

    GetProfile(ctx context.Context, userID string) (*dto.UserResponse, error)

    Logout(ctx context.Context, userID string) error

}

  

type service struct {

    repo Repository

    cfg  *config.Config

}

  

func NewService(repo Repository, cfg *config.Config) Service {

    return &service{repo: repo, cfg: cfg}

}

  

func (s *service) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {

    if err := req.Validate(); err != nil {

        return nil, response.BadRequest(err.Error())

    }

  

    // Check if email exists

    _, err := s.repo.FindByEmail(ctx, req.Email)

    if err == nil {

        return nil, response.ConflictError("Email already registered")

    }

  

    // Hash password

    hashedPassword, err := password.Hash(req.Password)

    if err != nil {

        return nil, response.InternalServerError("Failed to process password", err)

    }

  

    // Create user

    user := &models.User{

        Name:     req.Name,

        Email:    req.Email,

        Password: hashedPassword,

        Role:     "user",

        Status:   "active",

    }

  

    if err := s.repo.Create(ctx, user); err != nil {

        return nil, response.InternalServerError("Failed to create user", err)

    }

  

    // Generate tokens

    accessToken, _ := jwt.Generate(user.ID, user.Email, user.Role, s.cfg.JWT.Secret, 15*time.Minute)

    refreshToken, _ := jwt.Generate(user.ID, user.Email, user.Role, s.cfg.JWT.Secret, 7*24*time.Hour)

  

    return &dto.AuthResponse{

        User:         dto.ToUserResponse(user),

        AccessToken:  accessToken,

        RefreshToken: refreshToken,

        ExpiresIn:    900, // 15 minutes

    }, nil

}

  

func (s *service) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {

    if err := req.Validate(); err != nil {

        return nil, response.BadRequest(err.Error())

    }

  

    // Find user

    user, err := s.repo.FindByEmail(ctx, req.Email)

    if err != nil {

        if errors.Is(err, gorm.ErrRecordNotFound) {

            return nil, response.UnauthorizedError("Invalid credentials")

        }

        return nil, response.InternalServerError("Failed to find user", err)

    }

  

    // Verify password

    if !password.Verify(req.Password, user.Password) {

        return nil, response.UnauthorizedError("Invalid credentials")

    }

  

    // Update last login

    s.repo.UpdateLastLogin(ctx, user.ID)

  

    // Generate tokens

    accessToken, _ := jwt.Generate(user.ID, user.Email, user.Role, s.cfg.JWT.Secret, 15*time.Minute)

    refreshToken, _ := jwt.Generate(user.ID, user.Email, user.Role, s.cfg.JWT.Secret, 7*24*time.Hour)

  

    return &dto.AuthResponse{

        User:         dto.ToUserResponse(user),

        AccessToken:  accessToken,

        RefreshToken: refreshToken,

        ExpiresIn:    900,

    }, nil

}

  

func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {

    claims, err := jwt.Verify(refreshToken, s.cfg.JWT.Secret)

    if err != nil {

        return nil, response.UnauthorizedError("Invalid refresh token")

    }

  

    user, err := s.repo.FindByID(ctx, claims.UserID)

    if err != nil {

        return nil, response.UnauthorizedError("User not found")

    }

  

    accessToken, _ := jwt.Generate(user.ID, user.Email, user.Role, s.cfg.JWT.Secret, 15*time.Minute)

    newRefreshToken, _ := jwt.Generate(user.ID, user.Email, user.Role, s.cfg.JWT.Secret, 7*24*time.Hour)

  

    return &dto.AuthResponse{

        User:         dto.ToUserResponse(user),

        AccessToken:  accessToken,

        RefreshToken: newRefreshToken,

        ExpiresIn:    900,

    }, nil

}

  

func (s *service) GetProfile(ctx context.Context, userID string) (*dto.UserResponse, error) {

    user, err := s.repo.FindByID(ctx, userID)

    if err != nil {

        return nil, response.NotFoundError("User")

    }

    return dto.ToUserResponse(user), nil

}

  

func (s *service) Logout(ctx context.Context, userID string) error {

    // Implement token blacklist if needed

    return nil

}
```


migrations/000001_create_user_table.down.sql
empty
migrations/000001_create_user_table.up.sql
empty
