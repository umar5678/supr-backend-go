package websocket

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type TestHandler struct {
	manager *Manager
}

func NewTestHandler(manager *Manager) *TestHandler {
	return &TestHandler{
		manager: manager,
	}
}

type TestMessageRequest struct {
	UserID  string                 `json:"userId" binding:"required"`
	Type    MessageType            `json:"type" binding:"required"`
	Data    map[string]interface{} `json:"data"`
	DelayMS int                    `json:"delayMs"`
}

type TestBroadcastRequest struct {
	Type    MessageType            `json:"type" binding:"required"`
	Data    map[string]interface{} `json:"data"`
	DelayMS int                    `json:"delayMs"`
}

func (h *TestHandler) SendTestMessage(c *gin.Context) {
	var req TestMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	msg := NewTargetedMessage(req.Type, req.UserID, req.Data)

	if req.DelayMS > 0 {
		go func() {
			time.Sleep(time.Duration(req.DelayMS) * time.Millisecond)
			h.manager.Hub().SendToUser(req.UserID, msg)
		}()
	} else {
		h.manager.Hub().SendToUser(req.UserID, msg)
	}

	logger.Info("test message sent",
		"userID", req.UserID,
		"type", req.Type,
		"data", req.Data,
	)

	response.Success(c, gin.H{
		"message":   "Test message sent",
		"userID":    req.UserID,
		"type":      req.Type,
		"timestamp": time.Now().UTC(),
	}, "Test message sent successfully")
}

func (h *TestHandler) SendTestBroadcast(c *gin.Context) {
	var req TestBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	msg := NewMessage(req.Type, req.Data)

	if req.DelayMS > 0 { 
		go func() {
			time.Sleep(time.Duration(req.DelayMS) * time.Millisecond)
			h.manager.Hub().BroadcastToAll(msg)
		}()
	} else {
		h.manager.Hub().BroadcastToAll(msg)
	}

	logger.Info("test broadcast sent",
		"type", req.Type,
		"data", req.Data,
	)

	response.Success(c, gin.H{
		"message":   "Test broadcast sent",
		"type":      req.Type,
		"timestamp": time.Now().UTC(),
	}, "Test broadcast sent successfully")
}

func (h *TestHandler) GetConnectionStats(c *gin.Context) {
	stats := h.manager.GetStats()

	response.Success(c, gin.H{
		"connected_users":          stats.ConnectedUsers,
		"total_connections":        stats.TotalConnections,
		"avg_connections_per_user": stats.AvgConnectionsPerUser,
		"timestamp":                time.Now().UTC(),
	}, "Connection stats retrieved")
}

func (h *TestHandler) TestPresence(c *gin.Context) {
	userIDs := c.QueryArray("userIds")
	if len(userIDs) == 0 {
		c.Error(response.BadRequest("userIds query parameter required"))
		return
	}

	presence := make(map[string]bool)
	for _, userID := range userIDs {
		presence[userID] = h.manager.Hub().IsUserConnected(userID)
	}

	response.Success(c, gin.H{
		"presence":  presence,
		"timestamp": time.Now().UTC(),
	}, "Presence check completed")
}

func (h *TestHandler) DebugConnections(c *gin.Context) {
	debugInfo := h.manager.Hub().DebugInfo()

	ctx := context.Background()
	onlineUsers, _ := cache.GetAllOnlineUsers(ctx)

	response.Success(c, gin.H{
		"hub_debug":          debugInfo,
		"redis_online_users": onlineUsers,
		"timestamp":          time.Now().UTC(),
	}, "Debug information retrieved")
}

func (h *TestHandler) SendTestMessageDirect(c *gin.Context) {
	var req TestMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	msg := NewTargetedMessage(req.Type, req.UserID, req.Data)

	logger.Info("sending test message directly",
		"targetUserID", req.UserID,
		"type", req.Type,
		"data", req.Data,
	)

	h.manager.Hub().SendToUser(req.UserID, msg)

	response.Success(c, gin.H{
		"message":   "Test message sent directly via WebSocket",
		"userID":    req.UserID,
		"type":      req.Type,
		"timestamp": time.Now().UTC(),
	}, "Test message sent successfully")
}
