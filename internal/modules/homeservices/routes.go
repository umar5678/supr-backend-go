package homeservices

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	services := router.Group("/services")
	{
		services.GET("/categories", handler.ListCategories)
		services.GET("/category-slugs", handler.GetAllCategorySlugs)
		services.GET("/categories/:id", handler.GetCategoryWithTabs)
		services.GET("", handler.ListServices)
		services.GET("/:id", handler.GetServiceDetails)
		services.GET("/addons", handler.ListAddOns)

		customer := services.Group("/orders")
		customer.Use(authMiddleware)
		customer.Use(middleware.RequireRole("rider"))
		{
			customer.POST("", handler.CreateOrder)
			customer.GET("", handler.GetMyOrders)
			customer.GET("/:id", handler.GetOrderDetails)
			customer.POST("/:id/cancel", handler.CancelOrder)
		}

		provider := services.Group("/provider")
		provider.Use(authMiddleware)
		provider.Use(middleware.RequireRole("service_provider"))
		{
			provider.POST("/register", handler.RegisterProvider)
			provider.GET("/orders", handler.GetProviderOrders)
			provider.POST("/orders/:id/accept", handler.AcceptOrder)
			provider.POST("/orders/:id/reject", handler.RejectOrder)
			provider.POST("/orders/:id/start", handler.StartOrder)
			provider.POST("/orders/:id/complete", handler.CompleteOrder)
		}

		admin := services.Group("/admin")
		admin.Use(authMiddleware)
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.POST("/categories", handler.CreateCategory)
			admin.POST("/tabs", handler.CreateTab)
			admin.POST("/addons", handler.CreateAddOn)
			admin.POST("/services", handler.CreateService)
			admin.PUT("/services/:id", handler.UpdateService)
		}
	}
}
