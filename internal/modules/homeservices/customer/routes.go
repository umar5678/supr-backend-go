package customer

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	authMiddleware gin.HandlerFunc,
) {
	homeservices := router.Group("/homeservices")
	{

		categories := homeservices.Group("/categories")
		categories.Use(authMiddleware)
		{
			categories.GET("", handler.GetAllCategories)
			categories.GET("/:categorySlug", handler.GetCategoryDetail)
		}

		services := homeservices.Group("/services")
		{
			services.GET("", handler.ListServices)
			services.GET("/frequent", handler.GetFrequentServices)
			services.GET("/:slug", handler.GetService)
		}

		addons := homeservices.Group("/addons")
		{
			addons.GET("", handler.ListAddons)
			addons.GET("/discounted", handler.GetDiscountedAddons)
			addons.GET("/:slug", handler.GetAddon)
		}

		homeservices.GET("/search", handler.Search)

		orders := homeservices.Group("/orders")
		orders.Use(authMiddleware)
		{
			orders.POST("", handler.CreateOrder)
			orders.GET("", handler.ListOrders)
			orders.GET("/:id", handler.GetOrder)
			orders.GET("/:id/cancel/preview", handler.GetCancellationPreview)
			orders.POST("/:id/cancel", handler.CancelOrder)
			orders.POST("/:id/rate", handler.RateOrder)
		}
	}
}
