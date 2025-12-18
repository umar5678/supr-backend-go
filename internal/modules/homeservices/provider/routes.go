package provider

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all provider routes
func RegisterRoutes(router *gin.RouterGroup, handler *Handler, providerAuthMiddleware gin.HandlerFunc) {
	provider := router.Group("/provider")
	provider.Use(providerAuthMiddleware)
	{
		// Profile routes
		provider.GET("/profile", handler.GetProfile)
		provider.PATCH("/availability", handler.UpdateAvailability)

		// Service categories
		categories := provider.Group("/categories")
		{
			categories.GET("", handler.GetServiceCategories)
			categories.POST("", handler.AddServiceCategory)
			categories.PUT("/:categorySlug", handler.UpdateServiceCategory)
			categories.DELETE("/:categorySlug", handler.DeleteServiceCategory)
		}

		// Order routes
		orders := provider.Group("/orders")
		{
			// Available orders
			orders.GET("/available", handler.GetAvailableOrders)
			orders.GET("/available/:id", handler.GetAvailableOrderDetail)

			// My orders
			orders.GET("", handler.GetMyOrders)
			orders.GET("/:id", handler.GetMyOrderDetail)

			// Order actions
			orders.POST("/:id/accept", handler.AcceptOrder)
			orders.POST("/:id/reject", handler.RejectOrder)
			orders.POST("/:id/start", handler.StartOrder)
			orders.POST("/:id/complete", handler.CompleteOrder)
			orders.POST("/:id/rate", handler.RateCustomer)
		}

		// Statistics routes
		provider.GET("/statistics", handler.GetStatistics)
		provider.GET("/earnings", handler.GetEarnings)
	}
}
