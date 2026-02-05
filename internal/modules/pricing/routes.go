package pricing

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/middleware"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	pricing := router.Group("/pricing")
	{
		pricing.POST("/estimate", handler.GetFareEstimate)
		pricing.GET("/surge", handler.GetSurgeMultiplier)
		pricing.GET("/surge/zones", handler.GetActiveSurgeZones)

		pricing.POST("/surge/zones", authMiddleware, middleware.RequireAdmin(), handler.CreateSurgeZone)

		pricing.GET("/surge-rules", handler.GetSurgePricingRules)
		pricing.POST("/surge-rules", handler.CreateSurgePricingRule)
		pricing.POST("/calculate-surge", handler.CalculateSurge)
		pricing.GET("/demand", handler.GetCurrentDemand)
		pricing.POST("/calculate-eta", handler.CalculateETA)
	}
}
