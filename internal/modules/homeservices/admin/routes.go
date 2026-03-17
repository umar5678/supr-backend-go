package admin

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	adminAuthMiddleware gin.HandlerFunc,
) {
	homeservices := router.Group("/homeservices")
	homeservices.Use(adminAuthMiddleware)
	{
		services := homeservices.Group("/services")
		{
			services.POST("", handler.CreateService)
			services.GET("", handler.ListServices)
			services.GET("/:slug", handler.GetService)
			services.PUT("/:slug", handler.UpdateService)
			services.PATCH("/:slug/status", handler.UpdateServiceStatus)
			services.DELETE("/:slug", handler.DeleteService)
		}

		addons := homeservices.Group("/addons")
		{
			addons.POST("", handler.CreateAddon)
			addons.GET("", handler.ListAddons)
			addons.GET("/:slug", handler.GetAddon)
			addons.PUT("/:slug", handler.UpdateAddon)
			addons.PATCH("/:slug/status", handler.UpdateAddonStatus)
			addons.DELETE("/:slug", handler.DeleteAddon)
		}

		categories := homeservices.Group("/categories")
		{
			categories.GET("", handler.GetAllCategories)
			categories.GET("/:categorySlug", handler.GetCategoryDetails)
		}

		orders := homeservices.Group("/orders")
		{
			orders.GET("", handler.GetOrders)
			orders.GET("/:id", handler.GetOrderByID)
			orders.GET("/number/:orderNumber", handler.GetOrderByNumber)
			orders.GET("/:id/history", handler.GetOrderHistory)

			orders.PATCH("/:id/status", handler.UpdateOrderStatus)
			orders.POST("/:id/reassign", handler.ReassignOrder)
			orders.POST("/:id/cancel", handler.CancelOrder)

			orders.POST("/bulk/status", handler.BulkUpdateStatus)
		}

		analytics := homeservices.Group("/analytics")
		{
			analytics.GET("/overview", handler.GetOverviewAnalytics)
			analytics.GET("/providers", handler.GetProviderAnalytics)
			analytics.GET("/revenue", handler.GetRevenueReport)
		}

		homeservices.GET("/dashboard", handler.GetDashboard)
	}
}
