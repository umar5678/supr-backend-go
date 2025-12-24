package laundry

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/middleware"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// Initialize repository and service
	repo := NewRepository(db)
	service := NewService(repo, db)
	handler := NewHandler(service)

	// Public routes - Get service catalog and products
	public := router.Group("/api/v1/laundry")
	{
		// Get all services with nested products
		public.GET("/services", handler.GetServicesWithProducts)

		// Get products for specific service
		public.GET("/services/:slug/products", handler.GetServiceProducts)
	}

	// Customer routes (authenticated as customer/rider)
	customer := router.Group("/api/v1/laundry")
	customer.Use(middleware.Auth(cfg))
	{
		// Order management
		customer.POST("/orders", handler.CreateOrder)
		customer.GET("/orders/:id", handler.GetOrder)

		// Pickup & Delivery management (customer can also manage these)
		customer.POST("/orders/:id/pickup/start", handler.InitiatePickup)
		customer.POST("/orders/:id/pickup/complete", handler.CompletePickup)
		customer.POST("/orders/:id/delivery/start", handler.InitiateDelivery)
		customer.POST("/orders/:id/delivery/complete", handler.CompleteDelivery)

		// Issue reporting
		customer.POST("/orders/:id/issues", handler.ReportIssue)
	}

	// Provider routes (authenticated as service provider)
	provider := router.Group("/api/v1/laundry/provider")
	provider.Use(middleware.Auth(cfg))
	provider.Use(middleware.RequireRole("service_provider")) // Ensure user is a provider
	{
		// View available orders for provider
		provider.GET("/orders/available", handler.GetAvailableOrders)

		// View assigned work
		provider.GET("/pickups", handler.GetProviderPickups)
		provider.GET("/deliveries", handler.GetProviderDeliveries)
		provider.GET("/issues", handler.GetProviderIssues)

		// Manage pickups
		provider.POST("/orders/:id/pickup/start", handler.InitiatePickup)
		provider.POST("/orders/:id/pickup/complete", handler.CompletePickup)

		// Manage items
		provider.POST("/orders/:id/items", handler.AddItems)
		provider.PATCH("/items/:qrCode/status", handler.UpdateItemStatus)

		// Manage deliveries
		provider.POST("/orders/:id/delivery/start", handler.InitiateDelivery)
		provider.POST("/orders/:id/delivery/complete", handler.CompleteDelivery)

		// Manage issues
		provider.PATCH("/issues/:id", handler.ResolveIssue)
	}
}
