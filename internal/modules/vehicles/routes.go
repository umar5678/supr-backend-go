package vehicles

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	vehicles := router.Group("/vehicles")
	{
		vehicles.GET("/types", handler.GetAllVehicleTypes)
		vehicles.GET("/types/active", handler.GetActiveVehicleTypes)
		vehicles.GET("/types/:id", handler.GetVehicleTypeByID)
	}
}
