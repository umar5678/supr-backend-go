package ratings

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	ratings := router.Group("/ratings")
	ratings.Use(authMiddleware)
	{
		ratings.POST("", handler.CreateRating)
		// Public routes
		ratings.GET("/driver/:driverId/stats", handler.GetDriverRatingStats)
		ratings.GET("/driver/:driverId/breakdown", handler.GetDriverRatingBreakdown)

		// Protected routes
		ratings.Use(authMiddleware)
		ratings.POST("/driver", handler.RateDriver)
		ratings.POST("/rider", handler.RateRider)
		ratings.GET("/rider/:riderId/stats", handler.GetRiderRatingStats)
		ratings.GET("/rider/:riderId/breakdown", handler.GetRiderRatingBreakdown)
	}
}
