package pricing

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	pricing := router.Group("/pricing")
	{
		// Public endpoints (no auth required)
		pricing.POST("/estimate", handler.GetFareEstimate)
		pricing.GET("/surge", handler.GetSurgeMultiplier)
		pricing.GET("/surge/zones", handler.GetActiveSurgeZones)
	}
}
