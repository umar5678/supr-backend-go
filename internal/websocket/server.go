package websocket

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	manager *Manager
}

func NewServer(manager *Manager) *Server {
	return &Server{
		manager: manager,
	}
}

func (s *Server) HandleConnection() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, response.UnauthorizedError("Unauthorized"))
			return
		}

		userIDStr := userID.(string)

		logger.Info("Incoming WebSocket Connection Request",
			"userID", userIDStr,
			"ip", c.ClientIP(),
		)

		reconnectToken := c.Query("reconnect_token")

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logger.Error("websocket upgrade failed", "error", err, "userID", userIDStr)
			return
		}

		client := NewClient(s.manager.hub, conn, userIDStr, c.Request.UserAgent())
		client.manager = s.manager
		client.reconnectToken = reconnectToken

		s.manager.hub.register <- client

		welcomeMsg := NewMessage(TypeSystemMessage, map[string]interface{}{
			"message":        "Connected successfully",
			"clientId":       client.ID,
			"reconnectToken": client.GenerateReconnectToken(),
			"serverTime":     time.Now().UTC(),
		})
		client.send <- welcomeMsg

		go s.deliverOfflineMessages(client)

		go client.WritePump()
		go client.ReadPump()

		logger.Info("websocket connection established",
			"userID", userIDStr,
			"clientID", client.ID,
			"userAgent", c.Request.UserAgent(),
		)
	}
}

func (s *Server) deliverOfflineMessages(client *Client) {
	if !s.manager.config.PersistenceEnabled {
		return
	}

	if s.manager.notificationStore != nil {
		notifications, err := s.manager.notificationStore.GetPending(s.manager.ctx, client.UserID)
		if err != nil {
			logger.Error("failed to fetch offline notifications", "error", err, "userID", client.UserID)
			return
		}

		if len(notifications) > 0 {
			bulkMsg := NewMessage(TypeNotificationBulk, map[string]interface{}{
				"notifications": notifications,
				"count":         len(notifications),
			})
			client.send <- bulkMsg

			logger.Info("delivered offline notifications",
				"userID", client.UserID,
				"count", len(notifications),
			)
		}
	}
}

func (s *Server) HandleHealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := s.manager.GetStats()

		c.JSON(http.StatusOK, gin.H{
			"status":            "healthy",
			"connected_users":   stats.ConnectedUsers,
			"total_connections": stats.TotalConnections,
			"timestamp":         time.Now().UTC(),
		})
	}
}

func (s *Server) HandleStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, response.ForbiddenError("forbidden"))
			return
		}

		stats := s.manager.GetStats()

		c.JSON(http.StatusOK, gin.H{
			"connected_users":          stats.ConnectedUsers,
			"total_connections":        stats.TotalConnections,
			"avg_connections_per_user": stats.AvgConnectionsPerUser,
			"timestamp":                time.Now().UTC(),
		})
	}
}

func (s *Server) HandleUserPresence() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserIDs []string `json:"userIds" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.BadRequest("invalid request | Bad request"))
			return
		}

		presence := make(map[string]interface{})
		for _, userID := range req.UserIDs {
			presence[userID] = map[string]interface{}{
				"online":          s.manager.hub.IsUserConnected(userID),
				"connectionCount": s.manager.hub.GetUserConnectionCount(userID),
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"presence": presence,
		})
	}
}

func (s *Server) HandleBroadcast() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, response.ForbiddenError("forbidden"))
			return
		}

		var req struct {
			Type MessageType            `json:"type" binding:"required"`
			Data map[string]interface{} `json:"data" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.BadRequest("invalid request"))
			return
		}

		msg := NewMessage(req.Type, req.Data)
		s.manager.hub.BroadcastToAll(msg)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "broadcast sent",
		})
	}
}

func (s *Server) HandleSendToUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID string                 `json:"userId" binding:"required"`
			Type   MessageType            `json:"type" binding:"required"`
			Data   map[string]interface{} `json:"data" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.BadRequest("invalid request"))
			return
		}

		if !s.manager.hub.IsUserConnected(req.UserID) {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "user offline",
				"online":  false,
			})
			return
		}

		msg := NewTargetedMessage(req.Type, req.UserID, req.Data)
		s.manager.hub.SendToUser(req.UserID, msg)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "message sent",
			"online":  true,
		})
	}
}
