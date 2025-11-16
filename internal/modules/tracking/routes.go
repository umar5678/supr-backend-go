package tracking

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	tracking := router.Group("/tracking")
	{
		// Driver location update (protected)
		tracking.POST("/location", authMiddleware, handler.UpdateLocation)

		// Public/internal endpoints
		tracking.GET("/driver/:driverId", handler.GetDriverLocation)
		tracking.GET("/nearby", handler.FindNearbyDrivers)
	}
}
