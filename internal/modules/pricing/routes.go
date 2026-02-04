package pricing

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	pricing := router.Group("/pricing")
	{
		// Public endpoints (no auth required)
		pricing.POST("/estimate", handler.GetFareEstimate)
		pricing.GET("/surge", handler.GetSurgeMultiplier)
		pricing.GET("/surge/zones", handler.GetActiveSurgeZones)

		// Admin endpoints (auth required)
		pricing.POST("/surge/zones", authMiddleware, handler.CreateSurgeZone)

		// Enhanced surge pricing endpoints
		pricing.GET("/surge-rules", handler.GetSurgePricingRules)
		pricing.POST("/surge-rules", handler.CreateSurgePricingRule)
		pricing.POST("/calculate-surge", handler.CalculateSurge)
		pricing.GET("/demand", handler.GetCurrentDemand)
		pricing.POST("/calculate-eta", handler.CalculateETA)
	}
}
