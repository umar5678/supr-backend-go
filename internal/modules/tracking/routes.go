package tracking

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	tracking := router.Group("/tracking")
	{
		// Driver location update (protected)
		tracking.POST("/location", authMiddleware, handler.UpdateLocation)

		// Location queries
		tracking.GET("/driver/:driverId", handler.GetDriverLocation)
		tracking.GET("/nearby", handler.FindNearbyDrivers)

		// // Polyline endpoints (protected)
		// tracking.GET("/polyline/ride/:rideId", authMiddleware, handler.GetRidePolyline)
		// tracking.GET("/polyline/driver/:driverId", authMiddleware, handler.GeneratePolyline)
	}
}
