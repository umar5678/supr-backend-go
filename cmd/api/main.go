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
	"github.com/umar5678/go-backend/internal/modules/admin"
	"github.com/umar5678/go-backend/internal/modules/auth"
	"github.com/umar5678/go-backend/internal/modules/batching"
	"github.com/umar5678/go-backend/internal/modules/drivers"
	"github.com/umar5678/go-backend/internal/modules/fraud"
	"github.com/umar5678/go-backend/internal/modules/homeservices"
	homeservicesAdmin "github.com/umar5678/go-backend/internal/modules/homeservices/admin"
	homeservicesCustomer "github.com/umar5678/go-backend/internal/modules/homeservices/customer"
	_ "github.com/umar5678/go-backend/internal/modules/homeservices/dto" // Alias for clarity
	homeservicesProvider "github.com/umar5678/go-backend/internal/modules/homeservices/provider"
	"github.com/umar5678/go-backend/internal/modules/laundry"
	"github.com/umar5678/go-backend/internal/modules/messages"
	"github.com/umar5678/go-backend/internal/modules/pricing"
	"github.com/umar5678/go-backend/internal/modules/profile"
	"github.com/umar5678/go-backend/internal/modules/promotions"
	"github.com/umar5678/go-backend/internal/modules/ratings"
	"github.com/umar5678/go-backend/internal/modules/ridepin"
	"github.com/umar5678/go-backend/internal/modules/riders"
	"github.com/umar5678/go-backend/internal/modules/rides"
	"github.com/umar5678/go-backend/internal/modules/serviceproviders"
	"github.com/umar5678/go-backend/internal/modules/sos"

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

	// 1. Create Managerccccc
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

		// Use mock wallet service for now - TODO: implement proper wallet service adapter
		mockWalletService := homeservicesCustomer.NewMockWalletService()

		// Riders module (initialize FIRST, before auth)
		ridersRepo := riders.NewRepository(db)
		ridersService := riders.NewService(ridersRepo)
		ridersHandler := riders.NewHandler(ridersService)
		// Service Providers module
		spRepo := serviceproviders.NewRepository(db)
		spService := serviceproviders.NewService(spRepo)

		// Auth module
		authRepo := auth.NewRepository(db)
		authService := auth.NewService(authRepo, cfg, ridersService, spService)
		authHandler := auth.NewHandler(authService)
		authMiddleware := middleware.Auth(cfg)
		auth.RegisterRoutes(v1, authHandler, authMiddleware)

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
		driversService := drivers.NewService(driversRepo, walletService, db)
		driversHandler := drivers.NewHandler(driversService)
		drivers.RegisterRoutes(v1, driversHandler, authMiddleware)

		// tracking service
		trackingRepo := tracking.NewRepository(db)
		trackingService := tracking.NewService(trackingRepo)
		trackingHandler := tracking.NewHandler(trackingService)
		tracking.RegisterRoutes(v1, trackingHandler, authMiddleware)

		// Pricing Module
		pricingRepo := pricing.NewRepository(db)
		pricingService := pricing.NewService(pricingRepo, db, vehiclesRepo)
		pricingHandler := pricing.NewHandler(pricingService)
		pricing.RegisterRoutes(v1, pricingHandler, authMiddleware)

		// Admin module (initialize early for rides service)
		adminRepo := admin.NewRepository(db)
		adminService := admin.NewService(adminRepo, spRepo)
		adminHandler := admin.NewHandler(adminService)
		admin.RegisterRoutes(v1, adminHandler, authMiddleware)

		// Profile Module (initialize before rides service)
		profileRepo := profile.NewRepository(db)
		profileService := profile.NewService(profileRepo)
		profileHandler := profile.NewHandler(profileService)
		profile.RegisterRoutes(v1, profileHandler, authMiddleware)

		// Promotions Module (initialize before rides service)
		promotionsRepo := promotions.NewRepository(db)
		promotionsService := promotions.NewService(promotionsRepo)
		promotionsHandler := promotions.NewHandler(promotionsService)
		promotions.RegisterRoutes(v1, promotionsHandler, authMiddleware)

		// Fraud Detection Module (initialize before rides service)
		fraudRepo := fraud.NewRepository(db)
		fraudService := fraud.NewService(fraudRepo)
		fraudHandler := fraud.NewHandler(fraudService)
		fraud.RegisterRoutes(v1, fraudHandler, authMiddleware)

		// SOS Module (initialize before rides service)
		sosRepo := sos.NewRepository(db)
		sosService := sos.NewService(sosRepo, db)
		sosHandler := sos.NewHandler(sosService)
		sos.RegisterRoutes(v1, sosHandler, authMiddleware)

		// RidePin Module (initialize before rides service)
		ridePinRepo := ridepin.NewRepository(db)
		ridePinService := ridepin.NewService(ridePinRepo)

		// Messages Module (initialize before rides service)
		messagesRepo := messages.NewRepository(db)
		messagesService := messages.NewService(messagesRepo)
		messagesHandler := messages.NewHandler(messagesService)
		messages.RegisterRoutes(v1, messagesHandler, authMiddleware)

		// Register WebSocket message handlers (requires messagesService)
		handlers.RegisterMessageHandlers(wsManager, messagesService)

		// Home Services module (initialize before ratings)
		homeServicesRepo := homeservices.NewRepository(db)
		homeServicesService := homeservices.NewService(homeServicesRepo, walletService, cfg)
		homeServicesHandler := homeservices.NewHandler(homeServicesService)
		homeservices.RegisterRoutes(v1, homeServicesHandler, authMiddleware)

		// Ratings Module (initialize before rides service)
		ratingsRepo := ratings.NewRepository(db)
		ratingsService := ratings.NewService(ratingsRepo, db, homeServicesRepo)
		ratingsHandler := ratings.NewHandler(ratingsService)
		ratings.RegisterRoutes(v1, ratingsHandler, authMiddleware)

		// Batching Module (initialize before rides service)
		// 10-second collection window, max 20 requests per batch
		batchingService := batching.NewService(
			driversRepo,
			ratingsService,
			trackingService,
			10*time.Second, // Batch window
			20,             // Max batch size
		)

		// rides service
		ridesRepo := rides.NewRepository(db)
		ridesService := rides.NewService(
			ridesRepo,
			driversRepo,
			ridersRepo,
			pricingService,
			trackingService,
			walletService,
			ridePinService,    // ✅ RidePin service
			profileService,    // ✅ NOW DEFINED
			sosService,        // ✅ ADDED: SOS service for emergency features
			promotionsService, // ✅ NOW DEFINED
			ratingsService,    // ✅ Ratings service
			fraudService,      // ✅ NOW DEFINED
			batchingService,   // ✅ ADDED: Request batching service
			adminRepo,
		)
		ridesHandler := rides.NewHandler(ridesService)
		rides.RegisterRoutes(v1, ridesHandler, authMiddleware)

		// WebSocket routes
		websocket.RegisterRoutes(router, cfg, wsServer)

		// Admin Home Services
		homeservicesAdminRepo := homeservicesAdmin.NewRepository(db)
		homeservicesAdminService := homeservicesAdmin.NewService(homeservicesAdminRepo)
		homeservicesAdminHandler := homeservicesAdmin.NewHandler(homeservicesAdminService)

		// Order management (Module 6)
		homeservicesAdminOrderRepo := homeservicesAdmin.NewOrderRepository(db)
		homeservicesAdminOrderService := homeservicesAdmin.NewOrderService(
			homeservicesAdminOrderRepo,
			mockWalletService,
		)
		homeservicesAdminOrderHandler := homeservicesAdmin.NewOrderHandler(homeservicesAdminOrderService)
		adminGroup := v1.Group("/admin")
		homeservicesAdmin.RegisterRoutes(
			adminGroup,
			homeservicesAdminHandler,
			homeservicesAdminOrderHandler,
			authMiddleware,
		)

		// Customer Home Services
		homeservicesCustomerRepo := homeservicesCustomer.NewRepository(db)
		homeservicesCustomerService := homeservicesCustomer.NewService(homeservicesCustomerRepo)
		homeservicesCustomerHandler := homeservicesCustomer.NewHandler(homeservicesCustomerService)

		// Customer Order Management
		homeservicesOrderRepo := homeservicesCustomer.NewOrderRepository(db)
		homeservicesOrderService := homeservicesCustomer.NewOrderService(homeservicesOrderRepo, homeservicesCustomerRepo, mockWalletService)
		homeservicesOrderHandler := homeservicesCustomer.NewOrderHandler(homeservicesOrderService)

		homeservicesCustomer.RegisterRoutes(v1, homeservicesCustomerHandler, homeservicesOrderHandler, authMiddleware)

		// Provider Home Services
		homeservicesProviderRepo := homeservicesProvider.NewRepository(db)
		homeservicesProviderService := homeservicesProvider.NewService(
			homeservicesProviderRepo,
			mockWalletService,
		)
		homeservicesProviderHandler := homeservicesProvider.NewHandler(homeservicesProviderService)

		homeservicesProvider.RegisterRoutes(
			v1,
			homeservicesProviderHandler,
			authMiddleware,
		)

		// Laundry Service module
		laundry.RegisterRoutes(router, db, cfg)

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
