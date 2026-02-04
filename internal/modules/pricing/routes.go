package pricing

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	pricing := router.Group("/pricing")
	{
		// Public endpoints (no auth required)
		pricing.POST("/estimate", handler.GetFareEstimate)
		pricing.GET("/surge", handler.GetSurgeMultiplier)
		pricing.GET("/surge/zones", handler.GetActiveSurgeZones)

		// Admin endpoints (admin role required)
		pricing.POST("/surge/zones", middleware.RequireAdmin(), handler.CreateSurgeZone)

		// Enhanced surge pricing endpoints
		pricing.GET("/surge-rules", handler.GetSurgePricingRules)
		pricing.POST("/surge-rules", handler.CreateSurgePricingRule)
		pricing.POST("/calculate-surge", handler.CalculateSurge)
		pricing.GET("/demand", handler.GetCurrentDemand)
		pricing.POST("/calculate-eta", handler.CalculateETA)
	}
}
