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
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/database"
	_ "github.com/umar5678/go-backend/internal/docs"
	"github.com/umar5678/go-backend/internal/middleware"
	"github.com/umar5678/go-backend/internal/modules/auth"
	"github.com/umar5678/go-backend/internal/modules/drivers"
	_ "github.com/umar5678/go-backend/internal/modules/homeservices/dto" // Alias for clarity
	"github.com/umar5678/go-backend/internal/modules/pricing"
	_ "github.com/umar5678/go-backend/internal/modules/ratings/dto"
	"github.com/umar5678/go-backend/internal/modules/riders"
	"github.com/umar5678/go-backend/internal/modules/rides"
	todos "github.com/umar5678/go-backend/internal/modules/todo"
	"github.com/umar5678/go-backend/internal/modules/tracking"
	"github.com/umar5678/go-backend/internal/modules/vehicles"
	_ "github.com/umar5678/go-backend/internal/modules/vehicles/dto"
	"github.com/umar5678/go-backend/internal/modules/wallet"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/websocket"

	// ✅ ADD THESE IMPORTS
	// "github.com/umar5678/go-backend/internal/websockets/websocketutils"

	"github.com/umar5678/go-backend/internal/websocket/handlers"
	"github.com/umar5678/go-backend/internal/websocket/websocketutils"
)

// @title supr booking server in go
// @version 1.0
// @description Production-grade Go backend API

func main() {
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

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

	// Connect to Redis (NEW)
	if err := cache.ConnectRedis(&cfg.Redis); err != nil {
		logger.Fatal("failed to connect to redis", "error", err)
	}
	defer cache.CloseRedis()

	// // ✅ Initialize WebSocket hub
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// hub := websocket.NewHub()
	// go hub.Run(ctx)

	// // Initialize WebSocket manager and server
	// wsConfig := &websocket.Config{
	// 	PersistenceEnabled: false,
	// }
	// wsManager := websocket.NewManager(wsConfig)
	// wsServer := websocket.NewServer(wsManager)

	// logger.Info("websocket hub started")

	// ✅ Initialize WebSocket manager properly
	// remove the manual context/hub creation if relying on Manager.Start()
	wsConfig := &websocket.Config{
		PersistenceEnabled: false, // Set to true if you want Redis storage
		HeartbeatInterval:  30 * time.Second,
	}

	// 1. Create Manager
	wsManager := websocket.NewManager(wsConfig)
	wsServer := websocket.NewServer(wsManager)

	// 2. ✅ CRITICAL FIX: Initialize the global utility
	websocketutils.Initialize(wsManager)

	// 3. ✅ Register Handlers (Ride, Chat, etc.)
	handlers.RegisterAllHandlers(wsManager)

	// 4. ✅ Start Manager (This starts the Hub, Heartbeats, and Metrics)
	// Do not call `go hub.Run(ctx)` manually, let the manager handle it
	if err := wsManager.Start(); err != nil {
		logger.Fatal("failed to start websocket manager", "error", err)
	}

	logger.Info("websocket system initialized successfully")

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

	// 	// Global middleware
	// 	router.Use(middleware.RequestContext(cfg.App.Version))

	// 	if os.Getenv("ENV") == "development" || gin.Mode() == gin.DebugMode {
	// 		router.Use(middleware.DevelopmentLogger())
	// 	}

	// 	router.Use(middleware.Logger())
	// 	router.Use(middleware.Recovery())
	// 	router.Use(middleware.CORS(cfg.Server.CORS))

	if os.Getenv("ENV") == "development" || gin.Mode() == gin.DebugMode {
		router.Use(middleware.DevelopmentLogger())
	}

	// Health check endpoints
	router.GET("/health", healthCheck)
	router.GET("/ready", readyCheck(db))

	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.Use(middleware.RateLimit(cfg.Server.RateLimit))

		// Riders module (initialize FIRST, before auth)
		ridersRepo := riders.NewRepository(db)
		ridersService := riders.NewService(ridersRepo)
		ridersHandler := riders.NewHandler(ridersService)

		// Auth module
		authRepo := auth.NewRepository(db)
		authService := auth.NewService(authRepo, cfg, ridersService)
		authHandler := auth.NewHandler(authService)
		authMiddleware := middleware.Auth(cfg)
		auth.RegisterRoutes(v1, authHandler, authMiddleware)

		// Todo Module
		todoRepo := todos.NewRepository(db)
		todoService := todos.NewService(*todoRepo)
		todoHandler := todos.NewHandler(todoService)
		todos.RegisterRoutes(v1, todoHandler, authMiddleware)

		// Register riders routes
		riders.RegisterRoutes(v1, ridersHandler, authMiddleware)
		// Wallet module
		walletRepo := wallet.NewRepository(db)
		walletService := wallet.NewService(walletRepo, db)
		walletHandler := wallet.NewHandler(walletService)
		wallet.RegisterRoutes(v1, walletHandler, authMiddleware)

		// Vehicle Types module
		vehiclesRepo := vehicles.NewRepository(db)
		vehiclesService := vehicles.NewService(vehiclesRepo)
		vehiclesHandler := vehicles.NewHandler(vehiclesService)
		vehicles.RegisterRoutes(v1, vehiclesHandler)

		// Drivers module
		driversRepo := drivers.NewRepository(db)
		driversService := drivers.NewService(driversRepo)
		driversHandler := drivers.NewHandler(driversService)
		drivers.RegisterRoutes(v1, driversHandler, authMiddleware)

		// tracking service
		trackingRepo := tracking.NewRepository(db)
		trackingService := tracking.NewService(trackingRepo)
		trackingHandler := tracking.NewHandler(trackingService)
		tracking.RegisterRoutes(v1, trackingHandler, authMiddleware)

		// Pricing module
		pricingRepo := pricing.NewRepository(db)
		pricingService := pricing.NewService(pricingRepo, vehiclesRepo)
		pricingHandler := pricing.NewHandler(pricingService)
		pricing.RegisterRoutes(v1, pricingHandler)

		// rides service
		ridesRepo := rides.NewRepository(db)
		ridesService := rides.NewService(
			ridesRepo,
			driversRepo,
			ridersRepo,
			pricingService,
			trackingService,
			walletService,
		)
		ridesHandler := rides.NewHandler(ridesService)
		rides.RegisterRoutes(v1, ridesHandler, authMiddleware)

		// WebSocket routes
		websocket.RegisterRoutes(router, cfg, wsServer)

		// Add other modules here...
	}

	// Swagger documentation
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	// Cancel websocket hub context
	// cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}

	logger.Info("server stopped gracefully")
}

func healthCheck(c *gin.Context) {
	// Check Redis health
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	redisStatus := "healthy"
	if err := cache.HealthCheck(ctx); err != nil {
		redisStatus = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
		"redis":  redisStatus,
	})
}

func readyCheck(db interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		redisStatus := "connected"
		if err := cache.HealthCheck(ctx); err != nil {
			redisStatus = "disconnected"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ready",
			"database": "connected",
			"redis":    redisStatus,
			"time":     time.Now().Format(time.RFC3339),
		})
	}
}

// // main.go - UPDATED VERSION
// package main

// import (
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/joho/godotenv"
// 	swaggerFiles "github.com/swaggo/files"
// 	ginSwagger "github.com/swaggo/gin-swagger"

// 	"github.com/umar5678/go-backend/internal/config"
// 	"github.com/umar5678/go-backend/internal/database"
// 	_ "github.com/umar5678/go-backend/internal/docs"
// 	"github.com/umar5678/go-backend/internal/middleware"
// 	"github.com/umar5678/go-backend/internal/modules/auth"
// 	"github.com/umar5678/go-backend/internal/modules/drivers"
// 	"github.com/umar5678/go-backend/internal/modules/homeservices"
// 	"github.com/umar5678/go-backend/internal/modules/pricing"
// 	"github.com/umar5678/go-backend/internal/modules/ratings"
// 	"github.com/umar5678/go-backend/internal/modules/riders"
// 	"github.com/umar5678/go-backend/internal/modules/rides"
// 	todos "github.com/umar5678/go-backend/internal/modules/todo"
// 	"github.com/umar5678/go-backend/internal/modules/tracking"
// 	"github.com/umar5678/go-backend/internal/modules/vehicles"
// 	_ "github.com/umar5678/go-backend/internal/modules/vehicles/dto"
// 	"github.com/umar5678/go-backend/internal/modules/wallet"
// 	"github.com/umar5678/go-backend/internal/services/cache"
// 	"github.com/umar5678/go-backend/internal/utils/logger"
// 	"github.com/umar5678/go-backend/internal/websocket"
// 	"github.com/umar5678/go-backend/internal/websocket/handlers"
// )

// // @title supr booking server in go
// // @version 1.0
// // @description Production-ready Go backend API

// // Global WebSocket manager
// var wsManager *websocket.Manager

// func main() {
// 	_ = godotenv.Load()

// 	// Load configuration
// 	cfg, err := config.LoadConfig()
// 	if err != nil {
// 		fmt.Printf("Failed to load config: %v\n", err)
// 		os.Exit(1)
// 	}

// 	if err := cfg.Validate(); err != nil {
// 		fmt.Printf("Invalid configuration: %v\n", err)
// 		os.Exit(1)
// 	}

// 	// Initialize logger
// 	if err := logger.Initialize(&cfg.Logger); err != nil {
// 		fmt.Printf("Failed to initialize logger: %v\n", err)
// 		os.Exit(1)
// 	}
// 	defer logger.Sync()

// 	logger.Info("starting application",
// 		"name", cfg.App.Name,
// 		"environment", cfg.App.Environment,
// 		"version", cfg.App.Version,
// 	)

// 	// Connect to database
// 	db, err := database.ConnectPostgres(&cfg.Database)
// 	if err != nil {
// 		logger.Fatal("failed to connect to database", "error", err)
// 	}
// 	defer database.Close(db)

// 	// Connect to Redis
// 	if err := cache.ConnectRedis(&cfg.Redis); err != nil {
// 		logger.Fatal("failed to connect to redis", "error", err)
// 	}
// 	defer cache.CloseRedis()

// 	// Initialize WebSocket Manager
// 	wsConfig := &websocket.Config{
// 		JWTSecret:          cfg.JWT.Secret,
// 		MaxConnections:     10000,
// 		MessageBufferSize:  256,
// 		HeartbeatInterval:  30 * time.Second,
// 		ConnectionTimeout:  5 * time.Minute,
// 		EnablePresence:     true,
// 		EnableMessageStore: true,
// 		PersistenceEnabled: true,
// 	}

// 	// Register ride-specific handlers
// 	handlers.RegisterAllHandlers(wsManager)

// 	// Start WebSocket manager
// 	if err := wsManager.Start(); err != nil {
// 		logger.Fatal("failed to start websocket manager", "error", err)
// 	}
// 	defer wsManager.Shutdown(30 * time.Second)

// 	wsManager = websocket.NewManager(wsConfig)
// 	wsServer := websocket.NewServer(wsManager)
// 	// Create WebSocket server

// 	// Setup Gin
// 	if cfg.App.Environment == "production" {
// 		gin.SetMode(gin.ReleaseMode)
// 	}

// 	router := gin.New()

// 	// Global middleware
// 	router.Use(middleware.RequestContext(cfg.App.Version))

// 	if os.Getenv("ENV") == "development" || gin.Mode() == gin.DebugMode {
// 		router.Use(middleware.DevelopmentLogger())
// 	}

// 	router.Use(middleware.Logger())
// 	router.Use(middleware.Recovery())
// 	router.Use(middleware.CORS(cfg.Server.CORS))

// 	// Health check endpoints
// 	router.GET("/redis-health", healthCheck)
// 	router.GET("/ready", readyCheck(db))

// 	// Register WebSocket routes
// 	websocket.RegisterRoutes(router, cfg, wsServer)

// 	// API routes
// 	v1 := router.Group("/api/v1")
// 	{
// 		v1.Use(middleware.RateLimit(cfg.Server.RateLimit))

// 		// Riders module (initialize FIRST, before auth)
// 		ridersRepo := riders.NewRepository(db)
// 		ridersService := riders.NewService(ridersRepo)
// 		ridersHandler := riders.NewHandler(ridersService)

// 		// Auth module
// 		authRepo := auth.NewRepository(db)
// 		authService := auth.NewService(authRepo, cfg, ridersService)
// 		authHandler := auth.NewHandler(authService)
// 		authMiddleware := middleware.Auth(cfg)
// 		auth.RegisterRoutes(v1, authHandler, authMiddleware)

// 		// Todo Module
// 		todoRepo := todos.NewRepository(db)
// 		todoService := todos.NewService(*todoRepo)
// 		todoHandler := todos.NewHandler(todoService)
// 		todos.RegisterRoutes(v1, todoHandler, authMiddleware)

// 		// Register riders routes
// 		riders.RegisterRoutes(v1, ridersHandler, authMiddleware)

// 		// Wallet module
// 		walletRepo := wallet.NewRepository(db)
// 		walletService := wallet.NewService(walletRepo, db)
// 		walletHandler := wallet.NewHandler(walletService)
// 		wallet.RegisterRoutes(v1, walletHandler, authMiddleware)

// 		// Vehicle Types module
// 		vehiclesRepo := vehicles.NewRepository(db)
// 		vehiclesService := vehicles.NewService(vehiclesRepo)
// 		vehiclesHandler := vehicles.NewHandler(vehiclesService)
// 		vehicles.RegisterRoutes(v1, vehiclesHandler)

// 		// Drivers module
// 		driversRepo := drivers.NewRepository(db)
// 		driversService := drivers.NewService(driversRepo)
// 		driversHandler := drivers.NewHandler(driversService)
// 		drivers.RegisterRoutes(v1, driversHandler, authMiddleware)

// 		// tracking service
// 		trackingRepo := tracking.NewRepository(db)
// 		trackingService := tracking.NewService(trackingRepo)
// 		trackingHandler := tracking.NewHandler(trackingService)
// 		tracking.RegisterRoutes(v1, trackingHandler, authMiddleware)

// 		// Home Services module
// 		homeServiceRepo := homeservices.NewRepository(db)
// 		homeServiceSvc := homeservices.NewService(homeServiceRepo, walletService, cfg)
// 		homeServiceHandler := homeservices.NewHandler(homeServiceSvc)
// 		homeservices.RegisterRoutes(v1, homeServiceHandler, authMiddleware)

// 		// Ratings module
// 		ratingsRepo := ratings.NewRepository(db)
// 		ratingsService := ratings.NewService(ratingsRepo, homeServiceRepo)
// 		ratingsHandler := ratings.NewHandler(ratingsService)
// 		ratings.RegisterRoutes(v1, ratingsHandler, authMiddleware)

// 		// Pricing module
// 		pricingRepo := pricing.NewRepository(db)
// 		pricingService := pricing.NewService(pricingRepo, vehiclesRepo)
// 		pricingHandler := pricing.NewHandler(pricingService)
// 		pricing.RegisterRoutes(v1, pricingHandler)

// 		// rides service
// 		ridesRepo := rides.NewRepository(db)
// 		ridesService := rides.NewService(
// 			ridesRepo,
// 			driversRepo,
// 			ridersRepo,
// 			pricingService,
// 			trackingService,
// 			walletService,
// 		)
// 		ridesHandler := rides.NewHandler(ridesService)
// 		rides.RegisterRoutes(v1, ridesHandler, authMiddleware)
// 	}

// 	// Swagger documentation
// 	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

// 	// Start server
// 	srv := &http.Server{
// 		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
// 		Handler:      router,
// 		ReadTimeout:  cfg.Server.ReadTimeout,
// 		WriteTimeout: cfg.Server.WriteTimeout,
// 	}

// 	// Graceful shutdown
// 	go func() {
// 		logger.Info("server starting",
// 			"host", cfg.Server.Host,
// 			"port", cfg.Server.Port,
// 		)

// 		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 			logger.Fatal("failed to start server", "error", err)
// 		}
// 	}()

// 	// Wait for interrupt signal
// 	quit := make(chan os.Signal, 1)
// 	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
// 	<-quit

// 	logger.Info("shutting down server...")

// 	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
// 	defer cancel()

// 	if err := srv.Shutdown(shutdownCtx); err != nil {
// 		logger.Error("server forced to shutdown", "error", err)
// 	}

// 	logger.Info("server stopped gracefully")
// }

// func healthCheck(c *gin.Context) {
// 	// Check Redis health
// 	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
// 	defer cancel()

// 	redisStatus := "healthy"
// 	if err := cache.HealthCheck(ctx); err != nil {
// 		redisStatus = "unhealthy"
// 	}

// 	// Check WebSocket manager
// 	wsStatus := "healthy"
// 	if wsManager == nil {
// 		wsStatus = "not_initialized"
// 	}

// 	stats := wsManager.GetStats()

// 	c.JSON(http.StatusOK, gin.H{
// 		"status":    "ok",
// 		"time":      time.Now().Format(time.RFC3339),
// 		"redis":     redisStatus,
// 		"websocket": wsStatus,
// 		"ws_stats": gin.H{
// 			"connected_users":   stats.ConnectedUsers,
// 			"total_connections": stats.TotalConnections,
// 		},
// 	})
// }

// func readyCheck(db interface{}) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
// 		defer cancel()

// 		redisStatus := "connected"
// 		if err := cache.HealthCheck(ctx); err != nil {
// 			redisStatus = "disconnected"
// 		}

// 		wsStatus := "ready"
// 		var stats websocket.Stats
// 		if wsManager != nil {
// 			stats = wsManager.GetStats()
// 		} else {
// 			wsStatus = "not_ready"
// 		}

// 		c.JSON(http.StatusOK, gin.H{
// 			"status":          "ready",
// 			"database":        "connected",
// 			"redis":           redisStatus,
// 			"websocket":       wsStatus,
// 			"connected_users": stats.ConnectedUsers,
// 			"time":            time.Now().Format(time.RFC3339),
// 		})
// 	}
// }

// package main

// import (
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/gorilla/websocket"
// 	"github.com/joho/godotenv"
// 	swaggerFiles "github.com/swaggo/files"
// 	ginSwagger "github.com/swaggo/gin-swagger"

// 	"github.com/umar5678/go-backend/internal/config"
// 	"github.com/umar5678/go-backend/internal/database"
// 	_ "github.com/umar5678/go-backend/internal/docs"
// 	"github.com/umar5678/go-backend/internal/middleware"
// 	"github.com/umar5678/go-backend/internal/modules/auth"
// 	"github.com/umar5678/go-backend/internal/modules/drivers"
// 	"github.com/umar5678/go-backend/internal/modules/homeservices"
// 	"github.com/umar5678/go-backend/internal/modules/pricing"
// 	"github.com/umar5678/go-backend/internal/modules/ratings"
// 	"github.com/umar5678/go-backend/internal/modules/riders"
// 	"github.com/umar5678/go-backend/internal/modules/rides"
// 	todos "github.com/umar5678/go-backend/internal/modules/todo"
// 	"github.com/umar5678/go-backend/internal/modules/tracking"
// 	"github.com/umar5678/go-backend/internal/modules/vehicles"
// 	_ "github.com/umar5678/go-backend/internal/modules/vehicles/dto"
// 	"github.com/umar5678/go-backend/internal/modules/wallet"
// 	"github.com/umar5678/go-backend/internal/services/cache"
// 	"github.com/umar5678/go-backend/internal/utils/logger"
// 	ws "github.com/umar5678/go-backend/internal/websocket"
// )

// // @title supr booking server in go
// // @version 1.0
// // @description Production-ready Go backend API

// // Global WebSocket hub
// var wsHub *ws.Hub

// func main() {
// 	_ = godotenv.Load()

// 	// Load configuration
// 	cfg, err := config.LoadConfig()
// 	if err != nil {
// 		fmt.Printf("Failed to load config: %v\n", err)
// 		os.Exit(1)
// 	}

// 	if err := cfg.Validate(); err != nil {
// 		fmt.Printf("Invalid configuration: %v\n", err)
// 		os.Exit(1)
// 	}

// 	// Initialize logger
// 	if err := logger.Initialize(&cfg.Logger); err != nil {
// 		fmt.Printf("Failed to initialize logger: %v\n", err)
// 		os.Exit(1)
// 	}
// 	defer logger.Sync()

// 	logger.Info("starting application",
// 		"name", cfg.App.Name,
// 		"environment", cfg.App.Environment,
// 		"version", cfg.App.Version,
// 	)

// 	// Connect to database
// 	db, err := database.ConnectPostgres(&cfg.Database)
// 	if err != nil {
// 		logger.Fatal("failed to connect to database", "error", err)
// 	}
// 	defer database.Close(db)

// 	// Connect to Redis
// 	if err := cache.ConnectRedis(&cfg.Redis); err != nil {
// 		logger.Fatal("failed to connect to redis", "error", err)
// 	}
// 	defer cache.CloseRedis()

// 	// Initialize WebSocket Manager
// 	wsConfig := &websocket.Config{
// 		JWTSecret:          cfg.JWT.Secret,
// 		MaxConnections:     10000,
// 		MessageBufferSize:  256,
// 		HeartbeatInterval:  30 * time.Second,
// 		ConnectionTimeout:  5 * time.Minute,
// 		EnablePresence:     true,
// 		EnableMessageStore: true,
// 		PersistenceEnabled: true,
// 	}
// 	wsManager := websocket.NewManager(wsConfig)

// 	// Start WebSocket manager
// 	if err := wsManager.Start(); err != nil {
// 		logger.Fatal("failed to start websocket manager", "error", err)
// 	}

// 	// Create WebSocket server
// 	wsServer := websocket.NewServer(wsManager)

// 	// Setup Gin
// 	if cfg.App.Environment == "production" {
// 		gin.SetMode(gin.ReleaseMode)
// 	}

// 	router := gin.New()

// 	// Global middleware
// 	router.Use(middleware.RequestContext(cfg.App.Version))
// 	router.Use(middleware.Logger())
// 	router.Use(middleware.Recovery())
// 	router.Use(middleware.CORS(cfg.Server.CORS))

// 	// Health check endpoints
// 	router.GET("/redis-health", healthCheck)
// 	router.GET("/ready", readyCheck(db))

// 	// API routes
// 	v1 := router.Group("/api/v1")
// 	{
// 		v1.Use(middleware.RateLimit(cfg.Server.RateLimit))

// 		// WebSocket endpoint (before auth middleware)
// 		// v1.GET("/ws", handleWebSocket(cfg))

// 		// Riders module (initialize FIRST, before auth)
// 		ridersRepo := riders.NewRepository(db)
// 		ridersService := riders.NewService(ridersRepo)
// 		ridersHandler := riders.NewHandler(ridersService)

// 		// Auth module
// 		authRepo := auth.NewRepository(db)
// 		authService := auth.NewService(authRepo, cfg, ridersService)
// 		authHandler := auth.NewHandler(authService)
// 		authMiddleware := middleware.Auth(cfg)
// 		auth.RegisterRoutes(v1, authHandler, authMiddleware)

// 		// Todo Module
// 		todoRepo := todos.NewRepository(db)
// 		todoService := todos.NewService(*todoRepo)
// 		todoHandler := todos.NewHandler(todoService)
// 		todos.RegisterRoutes(v1, todoHandler, authMiddleware)

// 		// Register riders routes
// 		riders.RegisterRoutes(v1, ridersHandler, authMiddleware)

// 		// Wallet module
// 		walletRepo := wallet.NewRepository(db)
// 		walletService := wallet.NewService(walletRepo, db)
// 		walletHandler := wallet.NewHandler(walletService)
// 		wallet.RegisterRoutes(v1, walletHandler, authMiddleware)

// 		// Vehicle Types module
// 		vehiclesRepo := vehicles.NewRepository(db)
// 		vehiclesService := vehicles.NewService(vehiclesRepo)
// 		vehiclesHandler := vehicles.NewHandler(vehiclesService)
// 		vehicles.RegisterRoutes(v1, vehiclesHandler)

// 		// Drivers module
// 		driversRepo := drivers.NewRepository(db)
// 		driversService := drivers.NewService(driversRepo)
// 		driversHandler := drivers.NewHandler(driversService)
// 		drivers.RegisterRoutes(v1, driversHandler, authMiddleware)

// 		// tracking service
// 		trackingRepo := tracking.NewRepository(db)
// 		trackingService := tracking.NewService(trackingRepo)
// 		trackingHandler := tracking.NewHandler(trackingService)
// 		tracking.RegisterRoutes(v1, trackingHandler, authMiddleware)

// 		// Home Services module
// 		homeServiceRepo := homeservices.NewRepository(db)
// 		homeServiceSvc := homeservices.NewService(homeServiceRepo, walletService, cfg)
// 		homeServiceHandler := homeservices.NewHandler(homeServiceSvc)
// 		homeservices.RegisterRoutes(v1, homeServiceHandler, authMiddleware)

// 		// Ratings module
// 		ratingsRepo := ratings.NewRepository(db)
// 		ratingsService := ratings.NewService(ratingsRepo, homeServiceRepo)
// 		ratingsHandler := ratings.NewHandler(ratingsService)
// 		ratings.RegisterRoutes(v1, ratingsHandler, authMiddleware)

// 		// Pricing module
// 		pricingRepo := pricing.NewRepository(db)
// 		pricingService := pricing.NewService(pricingRepo, vehiclesRepo)
// 		pricingHandler := pricing.NewHandler(pricingService)
// 		pricing.RegisterRoutes(v1, pricingHandler)

// 		// rides service
// 		ridesRepo := rides.NewRepository(db)
// 		ridesService := rides.NewService(
// 			ridesRepo,
// 			driversRepo,
// 			ridersRepo,
// 			pricingService,
// 			trackingService,
// 			walletService,
// 		)
// 		ridesHandler := rides.NewHandler(ridesService)
// 		rides.RegisterRoutes(v1, ridesHandler, authMiddleware)
// 	}

// 	// WebSocket stats endpoint (optional - for monitoring)
// 	v1.GET("/ws/stats", func(c *gin.Context) {
// 		c.JSON(http.StatusOK, gin.H{
// 			"connected_users":   wsHub.GetConnectedUsers(),
// 			"total_connections": wsHub.GetTotalConnections(),
// 		})
// 	})

// 	// Swagger documentation
// 	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

// 	// Start server
// 	srv := &http.Server{
// 		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
// 		Handler:      router,
// 		ReadTimeout:  cfg.Server.ReadTimeout,
// 		WriteTimeout: cfg.Server.WriteTimeout,
// 	}

// 	// Graceful shutdown
// 	go func() {
// 		logger.Info("server starting",
// 			"host", cfg.Server.Host,
// 			"port", cfg.Server.Port,
// 		)

// 		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 			logger.Fatal("failed to start server", "error", err)
// 		}
// 	}()

// 	// Wait for interrupt signal
// 	quit := make(chan os.Signal, 1)
// 	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
// 	<-quit

// 	logger.Info("shutting down server...")

// 	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
// 	defer cancel()

// 	if err := srv.Shutdown(shutdownCtx); err != nil {
// 		logger.Error("server forced to shutdown", "error", err)
// 	}

// 	logger.Info("server stopped gracefully")
// }

// // handleWebSocket upgrades HTTP connection to WebSocket
// func handleWebSocket(cfg *config.Config) gin.HandlerFunc {
// 	upgrader := websocket.Upgrader{
// 		ReadBufferSize:  1024,
// 		WriteBufferSize: 1024,
// 		CheckOrigin: func(r *http.Request) bool {
// 			// In production, you might want to validate origin
// 			// For now, allow all origins (adjust based on your needs)
// 			return true
// 		},
// 	}

// 	return func(c *gin.Context) {
// 		// Get token from query param or header
// 		token := c.Query("token")
// 		if token == "" {
// 			token = c.GetHeader("Authorization")
// 		}

// 		if token == "" {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error":   "authentication token required",
// 				"message": "provide token as query param (?token=xxx) or Authorization header",
// 			})
// 			return
// 		}

// 		// Authenticate user
// 		userID, err := ws.AuthenticateWebSocket(
// 			c.Request.Context(),
// 			token,
// 			cfg.JWT.Secret,
// 		)
// 		if err != nil {
// 			c.JSON(http.StatusUnauthorized, gin.H{
// 				"error": err.Error(),
// 			})
// 			return
// 		}

// 		// Upgrade connection
// 		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
// 		if err != nil {
// 			logger.Error("websocket upgrade failed",
// 				"error", err,
// 				"userID", userID,
// 			)
// 			return
// 		}

// 		// Create client
// 		userAgent := c.Request.UserAgent()
// 		client := ws.NewClient(wsHub, conn, userID, userAgent)

// 		logger.Info("new websocket connection",
// 			"userID", userID,
// 			"clientID", client.ID,
// 			"userAgent", userAgent,
// 		)

// 		// Register client with hub
// 		wsHub.register <- client

// 		// Start read and write pumps
// 		go client.WritePump()
// 		go client.ReadPump()
// 	}
// }

// func healthCheck(c *gin.Context) {
// 	// Check Redis health
// 	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
// 	defer cancel()

// 	redisStatus := "healthy"
// 	if err := cache.HealthCheck(ctx); err != nil {
// 		redisStatus = "unhealthy"
// 	}

// 	// Check WebSocket hub
// 	wsStatus := "healthy"
// 	if wsHub == nil {
// 		wsStatus = "not_initialized"
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"status":    "ok",
// 		"time":      time.Now().Format(time.RFC3339),
// 		"redis":     redisStatus,
// 		"websocket": wsStatus,
// 		"ws_users":  wsHub.GetConnectedUsers(),
// 	})
// }

// func readyCheck(db interface{}) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
// 		defer cancel()

// 		redisStatus := "connected"
// 		if err := cache.HealthCheck(ctx); err != nil {
// 			redisStatus = "disconnected"
// 		}

// 		wsStatus := "ready"
// 		connectedUsers := 0
// 		if wsHub != nil {
// 			connectedUsers = wsHub.GetConnectedUsers()
// 		} else {
// 			wsStatus = "not_ready"
// 		}

// 		c.JSON(http.StatusOK, gin.H{
// 			"status":          "ready",
// 			"database":        "connected",
// 			"redis":           redisStatus,
// 			"websocket":       wsStatus,
// 			"connected_users": connectedUsers,
// 			"time":            time.Now().Format(time.RFC3339),
// 		})
// 	}
// }
