package provider

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, providerAuthMiddleware gin.HandlerFunc) {
	provider := router.Group("/provider")
	provider.Use(providerAuthMiddleware)
	{
		provider.GET("/profile", handler.GetProfile)
		provider.PATCH("/availability", handler.UpdateAvailability)

		categories := provider.Group("/categories")
		{
			categories.GET("", handler.GetServiceCategories)
			categories.POST("", handler.AddServiceCategory)
			categories.PUT("/:categorySlug", handler.UpdateServiceCategory)
			categories.DELETE("/:categorySlug", handler.DeleteServiceCategory)
		}

		orders := provider.Group("/orders")
		{
			orders.GET("/available", handler.GetAvailableOrders)
			orders.GET("/available/:id", handler.GetAvailableOrderDetail)

			orders.GET("", handler.GetMyOrders)
			orders.GET("/:id", handler.GetMyOrderDetail)

			orders.POST("/:id/accept", handler.AcceptOrder)
			orders.POST("/:id/reject", handler.RejectOrder)
			orders.POST("/:id/start", handler.StartOrder)
			orders.POST("/:id/complete", handler.CompleteOrder)
			orders.POST("/:id/rate", handler.RateCustomer)
		}

		provider.GET("/statistics", handler.GetStatistics)
		provider.GET("/earnings", handler.GetEarnings)
	}
}
