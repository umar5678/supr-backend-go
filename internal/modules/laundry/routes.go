package laundry

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/middleware"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	repo := NewRepository(db)
	service := NewService(repo, db)
	handler := NewHandler(service)

	public := router.Group("/api/v1/laundry")
	{
		public.GET("/services", handler.GetServicesWithProducts)

		public.GET("/services/:slug/products", handler.GetServiceProducts)
	}

	customer := router.Group("/api/v1/laundry")
	customer.Use(middleware.Auth(cfg))
	{
		customer.POST("/orders", handler.CreateOrder)
		customer.GET("/orders/:id", handler.GetOrder)

		customer.POST("/orders/:id/pickup/start", handler.InitiatePickup)
		customer.POST("/orders/:id/pickup/complete", handler.CompletePickup)
		customer.POST("/orders/:id/delivery/start", handler.InitiateDelivery)
		customer.POST("/orders/:id/delivery/complete", handler.CompleteDelivery)

		customer.POST("/orders/:id/issues", handler.ReportIssue)
	}

	provider := router.Group("/api/v1/laundry/provider")
	provider.Use(middleware.Auth(cfg))
	provider.Use(middleware.RequireRole("service_provider"))
	{

		provider.GET("/orders/available", handler.GetAvailableOrders)

		provider.GET("/pickups", handler.GetProviderPickups)
		provider.GET("/deliveries", handler.GetProviderDeliveries)
		provider.GET("/issues", handler.GetProviderIssues)

		provider.POST("/orders/:id/pickup/start", handler.InitiatePickup)
		provider.POST("/orders/:id/pickup/complete", handler.CompletePickup)

		provider.POST("/orders/:id/items", handler.AddItems)
		provider.PATCH("/items/:qrCode/status", handler.UpdateItemStatus)

		provider.POST("/orders/:id/delivery/start", handler.InitiateDelivery)
		provider.POST("/orders/:id/delivery/complete", handler.CompleteDelivery)

		provider.PATCH("/issues/:id", handler.ResolveIssue)
	}
}
