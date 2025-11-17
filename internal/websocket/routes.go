// internal/websocket/routes.go - UPDATED
package websocket

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/middleware"
)

// RegisterRoutes sets up WebSocket routes
func RegisterRoutes(router *gin.Engine, cfg *config.Config, server *Server) {
	ws := router.Group("/ws")
	{
		// WebSocket connection endpoint (uses WebSocket-specific auth)
		ws.GET("/connect", AuthMiddleware(cfg.JWT.Secret), server.HandleConnection())

		// Health check (public)
		ws.GET("/health", server.HandleHealthCheck())

		// Stats endpoint (admin only)
		ws.GET("/stats",
			middleware.Auth(cfg),
			middleware.RequireRole("admin"),
			server.HandleStats(),
		)

		// User presence check (requires auth)
		ws.POST("/presence",
			middleware.Auth(cfg),
			server.HandleUserPresence(),
		)

		// Send message to user (requires auth)
		ws.POST("/send",
			middleware.Auth(cfg),
			server.HandleSendToUser(),
		)

		// Broadcast message (admin only)
		ws.POST("/broadcast",
			middleware.Auth(cfg),
			middleware.RequireRole("admin"),
			server.HandleBroadcast(),
		)
	}

	// Test endpoints (only in development)
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
