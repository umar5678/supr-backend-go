// internal/modules/promotions/routes.go
package promotions

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	promotions := router.Group("/promotions")
	promotions.Use(authMiddleware)
	{
		// Promo codes
		promotions.POST("/promo-codes", handler.CreatePromoCode) // Admin only
		promotions.GET("/promo-codes", handler.ListPromoCodes)   // Admin only
		promotions.GET("/promo-codes/:code", handler.GetPromoCode)
		promotions.POST("/promo-codes/:id/deactivate", handler.DeactivatePromoCode) // Admin only

		// Validation and application
		promotions.POST("/validate", handler.ValidatePromoCode)
		promotions.POST("/apply", handler.ApplyPromoCode)

		// Free ride credits
		promotions.GET("/free-ride-credits", handler.GetFreeRideCredits)
		promotions.POST("/free-ride-credits", handler.AddFreeRideCredit) // Admin only
	}
}
