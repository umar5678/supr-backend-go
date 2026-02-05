package websocket

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/middleware"
)

func RegisterRoutes(router *gin.Engine, cfg *config.Config, server *Server) {
	ws := router.Group("/ws")
	{
		ws.GET("/connect", AuthMiddleware(cfg.JWT.Secret), server.HandleConnection())

		ws.GET("/health", server.HandleHealthCheck())

		ws.GET("/stats",
			middleware.Auth(cfg),
			middleware.RequireRole("admin"),
			server.HandleStats(),
		)

		ws.POST("/presence",
			middleware.Auth(cfg),
			server.HandleUserPresence(),
		)

		ws.POST("/send",
			middleware.Auth(cfg),
			server.HandleSendToUser(),
		)

		ws.POST("/broadcast",
			middleware.Auth(cfg),
			middleware.RequireRole("admin"),
			server.HandleBroadcast(),
		)
	}

	if cfg.App.Environment == "development" {
		testHandler := NewTestHandler(server.manager)
		test := router.Group("/test/websocket")
		{
			test.POST("/send", testHandler.SendTestMessage)
			test.POST("/broadcast", testHandler.SendTestBroadcast)
			test.GET("/stats", testHandler.GetConnectionStats)
			test.GET("/presence", testHandler.TestPresence)
			test.GET("/debug", testHandler.DebugConnections)
			test.POST("/send-direct", testHandler.SendTestMessageDirect)
		}
	}
}
