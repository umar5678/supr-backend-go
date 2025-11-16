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
	"github.com/umar5678/go-backend/internal/modules/homeservices"
	"github.com/umar5678/go-backend/internal/modules/pricing"
	"github.com/umar5678/go-backend/internal/modules/ratings"
	"github.com/umar5678/go-backend/internal/modules/riders"
	"github.com/umar5678/go-backend/internal/modules/rides"
	todos "github.com/umar5678/go-backend/internal/modules/todo"
	"github.com/umar5678/go-backend/internal/modules/tracking"
	"github.com/umar5678/go-backend/internal/modules/vehicles"
	_ "github.com/umar5678/go-backend/internal/modules/vehicles/dto"
	"github.com/umar5678/go-backend/internal/modules/wallet"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// @title supr booking server in go
// @version 1.0
// @description Production-ready Go backend API

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

	// Connect to Redis (FIXED)
	if err := cache.ConnectRedis(&cfg.Redis); err != nil { // ✅ Actually call InitRedis
		logger.Fatal("failed to connect to redis", "error", err)
	}
	defer cache.CloseRedis()

	// Setup Gin
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.RequestContext(cfg.App.Version))
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS(cfg.Server.CORS))

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

		// Home Services module
		homeServiceRepo := homeservices.NewRepository(db)
		homeServiceSvc := homeservices.NewService(homeServiceRepo, walletService, cfg)
		homeServiceHandler := homeservices.NewHandler(homeServiceSvc)
		homeservices.RegisterRoutes(v1, homeServiceHandler, authMiddleware)

		// Ratings module
		ratingsRepo := ratings.NewRepository(db)
		ratingsService := ratings.NewService(ratingsRepo, homeServiceRepo)
		ratingsHandler := ratings.NewHandler(ratingsService)
		ratings.RegisterRoutes(v1, ratingsHandler, authMiddleware)

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

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
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
