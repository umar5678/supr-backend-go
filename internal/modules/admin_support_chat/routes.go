package admin_support_chat

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/middleware"
)

func RegisterRoutes(router *gin.Engine, cfg *config.Config, service Service) {
	handler := NewHandler(service)

	chat := router.Group("/admin-support-chat")
	chat.Use(middleware.Auth(cfg))
	{
		chat.POST("/send", handler.SendMessage)
		chat.GET("/conversations", handler.GetUserConversations)
		chat.GET("/conversations/:conversationId", handler.GetConversationMessages)
		chat.POST("/:messageId/read", handler.MarkAsRead)
	}
}