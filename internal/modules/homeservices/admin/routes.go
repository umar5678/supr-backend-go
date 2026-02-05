package admin

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all admin home services routes
func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
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
			orders.GET("", handler.GetOrders)
			orders.GET("/:id", handler.GetOrderByID)
			orders.GET("/number/:orderNumber", handler.GetOrderByNumber)
			orders.GET("/:id/history", handler.GetOrderHistory)

			// Order actions
			orders.PATCH("/:id/status", handler.UpdateOrderStatus)
			orders.POST("/:id/reassign", handler.ReassignOrder)
			orders.POST("/:id/cancel", handler.CancelOrder)

			// Bulk operations
			orders.POST("/bulk/status", handler.BulkUpdateStatus)
		}

		// ==================== Analytics ====================
		analytics := homeservices.Group("/analytics")
		{
			analytics.GET("/overview", handler.GetOverviewAnalytics)
			analytics.GET("/providers", handler.GetProviderAnalytics)
			analytics.GET("/revenue", handler.GetRevenueReport)
		}

		// ==================== Dashboard ====================
		homeservices.GET("/dashboard", handler.GetDashboard)
	}
}
