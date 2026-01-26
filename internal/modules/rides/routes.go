package rides

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	rides := router.Group("/rides")
	rides.Use(authMiddleware)
	{
		// Rider endpoints
		rides.POST("", handler.CreateRide)
		rides.GET("", handler.ListRides)
		rides.GET("/:id", handler.GetRide)
		rides.POST("/:id/cancel", handler.CancelRide)
		rides.POST("/:id/emergency", handler.TriggerSOS)                     // Emergency SOS
		rides.POST("/available-cars", handler.GetAvailableCars)              // Available cars near rider
		rides.POST("/vehicles-with-details", handler.GetVehiclesWithDetails) // Complete vehicle details with pickup & destination

		// Driver endpoints
		rides.POST("/:id/accept", handler.AcceptRide)
		rides.POST("/:id/reject", handler.RejectRide)
		rides.POST("/:id/arrived", handler.MarkArrived)
		rides.POST("/:id/start", handler.StartRide)
		rides.POST("/:id/complete", handler.CompleteRide)
	}
}
