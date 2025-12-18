package admin

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all admin home services routes
func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	orderHandler *OrderHandler,
	adminAuthMiddleware gin.HandlerFunc,
) {
	// All routes require admin authentication
	homeservices := router.Group("/homeservices")
	homeservices.Use(adminAuthMiddleware)
	{
		// ==================== Service Management ====================
		services := homeservices.Group("/services")
		{
			services.POST("", handler.CreateService)
			// services.POST("", handler.UpdateHomeCleaningService)
			services.GET("", handler.ListServices)
			services.GET("/:slug", handler.GetService)
			services.PUT("/:slug", handler.UpdateService)
			services.PATCH("/:slug/status", handler.UpdateServiceStatus)
			services.DELETE("/:slug", handler.DeleteService)
		}

		// ==================== Addon Management ====================
		addons := homeservices.Group("/addons")
		{
			addons.POST("", handler.CreateAddon)
			addons.GET("", handler.ListAddons)
			addons.GET("/:slug", handler.GetAddon)
			addons.PUT("/:slug", handler.UpdateAddon)
			addons.PATCH("/:slug/status", handler.UpdateAddonStatus)
			addons.DELETE("/:slug", handler.DeleteAddon)
		}

		// ==================== Category Management ====================
		categories := homeservices.Group("/categories")
		{
			categories.GET("", handler.GetAllCategories)
			categories.GET("/:categorySlug", handler.GetCategoryDetails)
		}

		// ==================== Order Management ====================
		orders := homeservices.Group("/orders")
		{
			// List and search
			orders.GET("", orderHandler.GetOrders)
			orders.GET("/:id", orderHandler.GetOrderByID)
			orders.GET("/number/:orderNumber", orderHandler.GetOrderByNumber)
			orders.GET("/:id/history", orderHandler.GetOrderHistory)

			// Order actions
			orders.PATCH("/:id/status", orderHandler.UpdateOrderStatus)
			orders.POST("/:id/reassign", orderHandler.ReassignOrder)
			orders.POST("/:id/cancel", orderHandler.CancelOrder)

			// Bulk operations
			orders.POST("/bulk/status", orderHandler.BulkUpdateStatus)
		}

		// ==================== Analytics ====================
		analytics := homeservices.Group("/analytics")
		{
			analytics.GET("/overview", orderHandler.GetOverviewAnalytics)
			analytics.GET("/providers", orderHandler.GetProviderAnalytics)
			analytics.GET("/revenue", orderHandler.GetRevenueReport)
		}

		// ==================== Dashboard ====================
		homeservices.GET("/dashboard", orderHandler.GetDashboard)
	}
}
