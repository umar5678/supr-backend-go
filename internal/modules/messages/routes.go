package messages

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	messages := router.Group("/messages")
	messages.Use(authMiddleware)
	{
		// Message endpoints
		messages.GET("/rides/:rideId", handler.GetMessages)
		messages.GET("/rides/:rideId/unread-count", handler.GetUnreadCount)
		messages.POST("", handler.SendMessage)
		messages.POST("/:messageId/read", handler.MarkAsRead)
		messages.DELETE("/:messageId", handler.DeleteMessage)
	}
}
