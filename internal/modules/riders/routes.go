package riders

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	riders := router.Group("/riders")
	riders.Use(authMiddleware)
	{
		riders.GET("/profile", handler.GetProfile)
		riders.PUT("/profile", handler.UpdateProfile)
		riders.GET("/stats", handler.GetStats)
	}
}
