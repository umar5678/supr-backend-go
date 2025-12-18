package customer

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all customer home services routes
func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	orderHandler *OrderHandler,
	authMiddleware gin.HandlerFunc,
) {
	homeservices := router.Group("/homeservices")
	{
		// ==================== Public Routes (No Auth) ====================

		// Category routes
		categories := homeservices.Group("/categories")
		categories.Use(authMiddleware)
		{
			categories.GET("", handler.GetAllCategories)
			categories.GET("/:categorySlug", handler.GetCategoryDetail)
		}

		// Service routes (public)
		services := homeservices.Group("/services")
		{
			services.GET("", handler.ListServices)
			services.GET("/frequent", handler.GetFrequentServices)
			services.GET("/:slug", handler.GetService)
		}

		// Addon routes (public)
		addons := homeservices.Group("/addons")
		{
			addons.GET("", handler.ListAddons)
			addons.GET("/discounted", handler.GetDiscountedAddons)
			addons.GET("/:slug", handler.GetAddon)
		}

		// Search route (public)
		homeservices.GET("/search", handler.Search)

		// ==================== Protected Routes (Auth Required) ====================

		// Order routes (require authentication)
		orders := homeservices.Group("/orders")
		orders.Use(authMiddleware)
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("", orderHandler.ListOrders)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.GET("/:id/cancel/preview", orderHandler.GetCancellationPreview)
			orders.POST("/:id/cancel", orderHandler.CancelOrder)
			orders.POST("/:id/rate", orderHandler.RateOrder)
		}
	}
}
