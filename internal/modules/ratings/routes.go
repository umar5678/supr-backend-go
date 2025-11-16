package ratings

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	ratings := router.Group("/ratings")
	ratings.Use(authMiddleware)
	{
		ratings.POST("", handler.CreateRating)
	}
}
