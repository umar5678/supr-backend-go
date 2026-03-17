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
	"github.com/umar5678/go-backend/internal/modules/admin_support_chat"
	"github.com/umar5678/go-backend/internal/modules/auth"
	"github.com/umar5678/go-backend/internal/modules/batching"
	"github.com/umar5678/go-backend/internal/modules/documents"
	"github.com/umar5678/go-backend/internal/modules/drivers"
	"github.com/umar5678/go-backend/internal/modules/fraud"
	"github.com/umar5678/go-backend/internal/modules/homeservices"
	homeservicesAdmin "github.com/umar5678/go-backend/internal/modules/homeservices/admin"
	homeservicesCustomer "github.com/umar5678/go-backend/internal/modules/homeservices/customer"
	_ "github.com/umar5678/go-backend/internal/modules/homeservices/dto" // Alias for clarity
	homeservicesProvider "github.com/umar5678/go-backend/internal/modules/homeservices/provider"
	"github.com/umar5678/go-backend/internal/modules/laundry"
	"github.com/umar5678/go-backend/internal/modules/messages"
	"github.com/umar5678/go-backend/internal/modules/notifications"
	notificationcontroller "github.com/umar5678/go-backend/internal/modules/notifications/controller"
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

	"github.com/umar5678/go-backend/internal/websocket/handlers"
	"github.com/umar5678/go-backend/internal/websocket/websocketutils"
)

// @title supr booking server in go
// @version 1.0
// @description Go backend API

func main() {
	_ = godotenv.Load()

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		os.Exit(1)
	}

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

	db, err := database.ConnectPostgres(&cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", "error", err)
	}
	defer database.Close(db)

	if err := cache.ConnectRedis(&cfg.Redis); err != nil {
		logger.Fatal("failed to connect to redis", "error", err)
	}
	defer cache.CloseRedis()

	wsConfig := &websocket.Config{
		JWTSecret:           cfg.JWT.Secret,
		MaxConnections:      cfg.WebSocket.MaxConnections,
		MessageBufferSize:   cfg.WebSocket.MessageBufferSize,
		HeartbeatInterval:   cfg.WebSocket.PingPeriod,
		ConnectionTimeout:   cfg.WebSocket.PongWait,
		EnablePresence:      cfg.WebSocket.EnablePresence,
		EnableMessageStore:  cfg.WebSocket.EnableMessageStore,
		PersistenceEnabled:  cfg.WebSocket.PersistenceEnabled,
		PersistenceMode:     cfg.WebSocket.PersistenceMode,
		RDBSnapshotInterval: cfg.WebSocket.RDBSnapshotInterval,
		AOFSyncPolicy:       cfg.WebSocket.AOFSyncPolicy,
	}

	wsManager := websocket.NewManager(wsConfig, db)
	wsServer := websocket.NewServer(wsManager)

	handlers.RegisterAllHandlers(wsManager)

	if err := wsManager.Start(); err != nil {
		logger.Fatal("failed to start websocket manager", "error", err)
	}

	logger.Info("websocket system initialized successfully")

	notificationSystem, err := notifications.NewNotificationSystem(
		context.Background(),
		db,
		cfg.Kafka,
	)
	if err != nil {
		logger.Fatal("failed to initialize notification system", "error", err)
	}
	defer func() {
		if err := notificationSystem.Stop(); err != nil {
			logger.Error("error stopping notification system", "error", err)
		}
	}()

	if err := notificationSystem.Start(context.Background()); err != nil {
		logger.Fatal("failed to start notification system", "error", err)
	}

	logger.Info("notification system initialized and started successfully")

	orderExpirationService := homeservices.NewOrderExpirationService(db)
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := orderExpirationService.ExpireUnacceptedOrders(ctx); err != nil {
				logger.Error("order expiration job failed", "error", err)
			}
			cancel()
		}
	}()

	logger.Info("order expiration job started")

	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	router.Use(middleware.RequestContext(cfg.App.Version))
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CORS(cfg.Server.CORS))

	if os.Getenv("ENV") == "development" || gin.Mode() == gin.DebugMode {
		router.Use(middleware.DevelopmentLogger())
	}

	router.GET("/health", healthCheck)
	router.GET("/ready", readyCheck(db))

	v1 := router.Group("/api/v1")
	{
		v1.Use(middleware.RateLimit(cfg.Server.RateLimit))

		ridersRepo := riders.NewRepository(db)
		ridersService := riders.NewServiceWithNotifications(ridersRepo, notificationSystem.GetProducer())
		ridersHandler := riders.NewHandler(ridersService)

		spRepo := serviceproviders.NewRepository(db)
		spService := serviceproviders.NewService(spRepo)

		authRepo := auth.NewRepository(db)
		authService := auth.NewServiceWithNotifications(authRepo, cfg, ridersService, spService, notificationSystem.GetProducer())
		authHandler := auth.NewHandler(authService)
		authMiddleware := middleware.Auth(cfg)
		auth.RegisterRoutes(v1, authHandler, authMiddleware)

		riders.RegisterRoutes(v1, ridersHandler, authMiddleware)

		walletRepo := wallet.NewRepository(db)
		walletService := wallet.NewServiceWithNotifications(walletRepo, db, notificationSystem.GetProducer())
		walletHandler := wallet.NewHandler(walletService)
		wallet.RegisterRoutes(v1, walletHandler, authMiddleware)

		vehiclesRepo := vehicles.NewRepository(db)
		vehiclesService := vehicles.NewServiceWithNotifications(vehiclesRepo, notificationSystem.GetProducer())
		vehiclesHandler := vehicles.NewHandler(vehiclesService)
		vehicles.RegisterRoutes(v1, vehiclesHandler)

		driversRepo := drivers.NewRepository(db)
		driversService := drivers.NewServiceWithNotifications(driversRepo, walletService, db, notificationSystem.GetProducer())
		driversHandler := drivers.NewHandler(driversService)
		drivers.RegisterRoutes(v1, driversHandler, authMiddleware)

		trackingRepo := tracking.NewRepository(db)
		trackingService := tracking.NewServiceWithNotifications(trackingRepo, notificationSystem.GetProducer())
		trackingHandler := tracking.NewHandler(trackingService)
		tracking.RegisterRoutes(v1, trackingHandler, authMiddleware)

		pricingRepo := pricing.NewRepository(db)
		pricingService := pricing.NewServiceWithNotifications(pricingRepo, db, vehiclesRepo, notificationSystem.GetProducer())
		pricingHandler := pricing.NewHandler(pricingService)
		pricing.RegisterRoutes(v1, pricingHandler, authMiddleware)

		adminRepo := admin.NewRepository(db)
		adminService := admin.NewServiceWithNotifications(adminRepo, spRepo, driversRepo, notificationSystem.GetProducer())
		adminHandler := admin.NewHandler(adminService)
		admin.RegisterRoutes(v1, adminHandler, authMiddleware)

		profileRepo := profile.NewRepository(db)
		profileService := profile.NewServiceWithNotifications(profileRepo, walletService, notificationSystem.GetProducer())
		profileHandler := profile.NewHandler(profileService)
		profile.RegisterRoutes(v1, profileHandler, authMiddleware)

		promotionsRepo := promotions.NewRepository(db)
		promotionsService := promotions.NewServiceWithNotifications(promotionsRepo, notificationSystem.GetProducer())
		promotionsHandler := promotions.NewHandler(promotionsService)
		promotions.RegisterRoutes(v1, promotionsHandler, authMiddleware)

		fraudRepo := fraud.NewRepository(db)
		fraudService := fraud.NewServiceWithNotifications(fraudRepo, notificationSystem.GetProducer())
		fraudHandler := fraud.NewHandler(fraudService)
		fraud.RegisterRoutes(v1, fraudHandler, authMiddleware)

		sosRepo := sos.NewRepository(db)
		sosService := sos.NewServiceWithNotifications(sosRepo, db, notificationSystem.GetProducer())
		sosHandler := sos.NewHandler(sosService)
		sos.RegisterRoutes(v1, sosHandler, authMiddleware)

		ridePinRepo := ridepin.NewRepository(db)
		ridePinService := ridepin.NewServiceWithNotifications(ridePinRepo, notificationSystem.GetProducer())

		messagesRepo := messages.NewRepository(db)
		messagesService := messages.NewServiceWithNotifications(messagesRepo, notificationSystem.GetProducer())
		messagesHandler := messages.NewHandler(messagesService)
		messages.RegisterRoutes(v1, messagesHandler, authMiddleware)

		handlers.RegisterMessageHandlers(wsManager, messagesService)

		notifController := notificationcontroller.NewNotificationController(
			notificationSystem.GetNotificationService(),
			notificationSystem.GetPushService(),
			cfg,
		)
		notifController.RegisterRoutes(v1, authMiddleware)

		homeServicesRepo := homeservices.NewRepository(db)
		homeServicesService := homeservices.NewServiceWithNotifications(homeServicesRepo, walletService, cfg, notificationSystem.GetProducer())
		homeServicesHandler := homeservices.NewHandler(homeServicesService)
		homeservices.RegisterRoutes(v1, homeServicesHandler, authMiddleware)

		ratingsRepo := ratings.NewRepository(db)
		ratingsService := ratings.NewService(ratingsRepo, db, homeServicesRepo)
		ratingsHandler := ratings.NewHandler(ratingsService)
		ratings.RegisterRoutes(v1, ratingsHandler, authMiddleware)

		batchingService := batching.NewService(
			driversRepo,
			ratingsService,
			trackingService,
			10*time.Second,
			20,
		)

		ridesRepo := rides.NewRepository(db)
		ridesService := rides.NewServiceWithNotifications(
			ridesRepo,
			driversRepo,
			ridersRepo,
			pricingService,
			trackingService,
			walletService,
			ridePinService,
			profileService,
			sosService,
			promotionsService,
			ratingsService,
			fraudService,
			batchingService,
			adminRepo,
			notificationSystem.GetProducer(),
		)
		ridesHandler := rides.NewHandler(ridesService)
		rides.RegisterRoutes(v1, ridesHandler, authMiddleware)

		websocket.RegisterRoutes(router, cfg, wsServer)

		homeservicesAdminRepo := homeservicesAdmin.NewRepository(db)
		homeservicesAdminService := homeservicesAdmin.NewService(homeservicesAdminRepo, walletService)
		homeservicesAdminHandler := homeservicesAdmin.NewHandler(homeservicesAdminService)
		adminGroup := v1.Group("/admin")
		homeservicesAdmin.RegisterRoutes(
			adminGroup,
			homeservicesAdminHandler,
			authMiddleware,
		)

		homeservicesCustomerRepo := homeservicesCustomer.NewRepository(db)
		homeservicesCustomerService := homeservicesCustomer.NewService(homeservicesCustomerRepo, homeservicesCustomerRepo, walletService)
		homeservicesCustomerHandler := homeservicesCustomer.NewHandler(homeservicesCustomerService)

		homeservicesCustomer.RegisterRoutes(v1, homeservicesCustomerHandler, authMiddleware)

		homeservicesProviderRepo := homeservicesProvider.NewRepository(db)
		homeservicesProviderService := homeservicesProvider.NewService(
			homeservicesProviderRepo,
			walletService,
			ridePinService,
		)
		homeservicesProviderHandler := homeservicesProvider.NewHandler(homeservicesProviderService)

		homeservicesProvider.RegisterRoutes(
			v1,
			homeservicesProviderHandler,
			authMiddleware,
		)

		laundry.RegisterRoutesWithNotifications(router, db, cfg, walletService, ridePinService, notificationSystem.GetProducer())

		adminSupportRepo := admin_support_chat.NewRepository(db)
		adminSupportService := admin_support_chat.NewService(adminSupportRepo, notificationSystem.GetProducer())
		websocketutils.Initialize(wsManager, adminSupportService)

		admin_support_chat.RegisterRoutesWithBroadcast(
			v1,
			cfg,
			adminSupportService,
			websocketutils.BroadcastAdminSupportMessage,
		)

		documentsRepo := documents.NewRepository(db)
		documentsService := documents.NewService(documentsRepo, cfg, notificationSystem.GetProducer())
		documentsHandler := documents.NewHandler(documentsService)
		documents.RegisterRoutes(v1, documentsHandler, authMiddleware)

		// Add other modules here...
	}

	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		logger.Info("server starting",
			"host", cfg.Server.Host,
			"port", cfg.Server.Port,
		)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
	}

	logger.Info("server stopped gracefully")
}

func healthCheck(c *gin.Context) {

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
