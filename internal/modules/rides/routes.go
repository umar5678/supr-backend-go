package rides

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	rides := router.Group("/rides")
	rides.Use(authMiddleware)
	{
		rides.POST("", handler.CreateRide)
		rides.GET("", handler.ListRides)
		rides.GET("/:id", handler.GetRide)
		rides.POST("/:id/cancel", handler.CancelRide)
		rides.POST("/:id/emergency", handler.TriggerSOS)                    
		rides.POST("/available-cars", handler.GetAvailableCars)             
		rides.POST("/vehicles-with-details", handler.GetVehiclesWithDetails)

		rides.POST("/:id/accept", handler.AcceptRide)
		rides.POST("/:id/reject", handler.RejectRide)
		rides.POST("/:id/arrived", handler.MarkArrived)
		rides.POST("/:id/start", handler.StartRide)
		rides.POST("/:id/complete", handler.CompleteRide)
	}
}
