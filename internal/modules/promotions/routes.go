// internal/modules/promotions/routes.go
package promotions

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	promotions := router.Group("/promotions")
	promotions.Use(authMiddleware)
	{
		promotions.POST("/promo-codes", handler.CreatePromoCode, middleware.RequireAdmin()) 
		promotions.GET("/promo-codes", handler.ListPromoCodes, middleware.RequireAdmin())   
		promotions.GET("/promo-codes/:code", handler.GetPromoCode)
		promotions.POST("/promo-codes/:id/deactivate", handler.DeactivatePromoCode, middleware.RequireAdmin()) 
		promotions.POST("/validate", handler.ValidatePromoCode)
		promotions.POST("/apply", handler.ApplyPromoCode)

		promotions.GET("/free-ride-credits", handler.GetFreeRideCredits)
		promotions.POST("/free-ride-credits", handler.AddFreeRideCredit, middleware.RequireAdmin())
	}
}
