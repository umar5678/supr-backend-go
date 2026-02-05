package tracking

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	tracking := router.Group("/tracking")
	{
		tracking.POST("/location", authMiddleware, handler.UpdateLocation)

		tracking.GET("/driver/:driverId", handler.GetDriverLocation)
		tracking.GET("/nearby", handler.FindNearbyDrivers)
	}
}
