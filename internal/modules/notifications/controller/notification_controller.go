package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/modules/notifications/dto"
	"github.com/umar5678/go-backend/internal/modules/notifications/service"
	"github.com/umar5678/go-backend/internal/utils/jwt"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type RegisterPushTokenRequest struct {
	Token    string `json:"token" binding:"required"`
	DeviceID string `json:"device_id" binding:"required"`
	DeviceOS string `json:"device_os" binding:"required"`
}

type UnregisterPushTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type NotificationController struct {
	notifService service.NotificationService
	pushService  service.PushService
	upgrader     websocket.Upgrader
	cfg          *config.Config
}

func NewNotificationController(notifService service.NotificationService, pushService service.PushService, cfg *config.Config) *NotificationController {
	return &NotificationController{
		notifService: notifService,
		pushService:  pushService,
		cfg:          cfg,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

type WebSocketSubscribable interface {
	SubscribeToUser(userID uuid.UUID, subscriberID string, msgChan chan service.PushMessage) error
	UnsubscribeFromUser(userID uuid.UUID, subscriberID string) error
}

func (c *NotificationController) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	notifications := rg.Group("/notifications")
	notifications.Use(authMiddleware)
	{
		notifications.GET("", c.GetNotifications)
		notifications.GET("/unread/count", c.GetUnreadCount)
		notifications.POST("/:id/read", c.MarkAsRead)
		notifications.POST("/read-all", c.MarkAllAsRead)
		notifications.DELETE("/:id", c.DeleteNotification)

		notifications.POST("/push-token", c.RegisterPushToken)
		notifications.DELETE("/push-token", c.UnregisterPushToken)

		notifications.GET("/stats", c.GetPushStats)
	}

	rg.GET("/notifications/ws/push", c.SubscribePush)
}

// GetNotifications godoc
// @Summary Get user notifications
// @Description Get paginated list of notifications for authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /notifications [get]
// @Security BearerAuth
func (c *NotificationController) GetNotifications(ctx *gin.Context) {
	userIDStr, exists := ctx.Get("userID")
	if !exists {
		response.Unauthorized(ctx, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Unauthorized(ctx, "invalid user id")
		return
	}

	var req dto.GetNotificationsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.SendError(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	logger.Info("GetNotifications called", "userID", userID.String(), "page", req.Page, "pageSize", req.PageSize)

	result, err := c.notifService.GetUserNotifications(ctx.Request.Context(), userID, &req)
	if err != nil {
		logger.Error("failed to get notifications", "error", err, "userID", userID.String())
		response.InternalError(ctx, "Failed to get notifications")
		return
	}

	logger.Info("GetNotifications success", "userID", userID.String(), "count", len(result.Notifications), "total", result.Total)
	response.Success(ctx, result, "Notifications retrieved successfully")
}

// GetUnreadCount godoc
// @Summary Get unread notification count
// @Description Get count of unread notifications for authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /notifications/unread/count [get]
// @Security BearerAuth
func (c *NotificationController) GetUnreadCount(ctx *gin.Context) {
	userIDStr, exists := ctx.Get("userID")
	if !exists {
		response.Unauthorized(ctx, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Unauthorized(ctx, "invalid user id")
		return
	}

	count, err := c.notifService.GetUnreadCount(ctx.Request.Context(), userID)
	if err != nil {
		logger.Error("failed to get unread count", "error", err)
		response.InternalError(ctx, "Failed to get unread count")
		return
	}

	response.Success(ctx, dto.UnreadCountResponse{Count: count}, "Unread count retrieved")
}

// MarkAsRead godoc
// @Summary Mark notification as read
// @Description Mark a specific notification as read
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /notifications/{id}/read [post]
// @Security BearerAuth
func (c *NotificationController) MarkAsRead(ctx *gin.Context) {
	userIDStr, exists := ctx.Get("userID")
	if !exists {
		response.Unauthorized(ctx, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Unauthorized(ctx, "invalid user id")
		return
	}

	notificationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.SendError(ctx, http.StatusBadRequest, "invalid notification ID", nil)
		return
	}

	if err := c.notifService.MarkAsRead(ctx.Request.Context(), notificationID, userID); err != nil {
		logger.Error("failed to mark notification as read", "error", err)
		response.InternalError(ctx, "Failed to mark notification as read")
		return
	}

	response.Success(ctx, nil, "Notification marked as read")
}

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Mark all user notifications as read
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /notifications/read-all [post]
// @Security BearerAuth
func (c *NotificationController) MarkAllAsRead(ctx *gin.Context) {
	userIDStr, exists := ctx.Get("userID")
	if !exists {
		response.Unauthorized(ctx, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Unauthorized(ctx, "invalid user id")
		return
	}

	if err := c.notifService.MarkAllAsRead(ctx.Request.Context(), userID); err != nil {
		logger.Error("failed to mark all notifications as read", "error", err)
		response.InternalError(ctx, "Failed to mark all notifications as read")
		return
	}

	response.Success(ctx, nil, "All notifications marked as read")
}

// DeleteNotification godoc
// @Summary Delete notification
// @Description Delete a specific notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /notifications/{id} [delete]
// @Security BearerAuth
func (c *NotificationController) DeleteNotification(ctx *gin.Context) {
	userIDStr, exists := ctx.Get("userID")
	if !exists {
		response.Unauthorized(ctx, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Unauthorized(ctx, "invalid user id")
		return
	}

	notificationID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.SendError(ctx, http.StatusBadRequest, "invalid notification ID", nil)
		return
	}

	if err := c.notifService.DeleteNotification(ctx.Request.Context(), notificationID, userID); err != nil {
		logger.Error("failed to delete notification", "error", err)
		response.InternalError(ctx, "Failed to delete notification")
		return
	}

	response.Success(ctx, nil, "Notification deleted")
}

// RegisterPushToken godoc
// @Summary Register push token
// @Description Register a device push token for receiving notifications
// @Tags notifications
// @Accept json
// @Produce json
// @Param payload body RegisterPushTokenRequest true "Token details"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /notifications/push-token [post]
// @Security BearerAuth
func (c *NotificationController) RegisterPushToken(ctx *gin.Context) {
	userIDStr, exists := ctx.Get("userID")
	if !exists {
		response.Unauthorized(ctx, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Unauthorized(ctx, "invalid user id")
		return
	}

	var req RegisterPushTokenRequest
	if err := ctx.BindJSON(&req); err != nil {
		response.SendError(ctx, http.StatusBadRequest, "invalid request", nil)
		return
	}

	if err := c.pushService.RegisterToken(ctx.Request.Context(), userID, req.Token, req.DeviceID, req.DeviceOS); err != nil {
		logger.Error("failed to register push token", "error", err)
		response.InternalError(ctx, "Failed to register push token")
		return
	}

	response.Success(ctx, nil, "Push token registered successfully")
}

// UnregisterPushToken godoc
// @Summary Unregister push token
// @Description Unregister a device push token
// @Tags notifications
// @Accept json
// @Produce json
// @Param payload body UnregisterPushTokenRequest true "Token to unregister"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /notifications/push-token [delete]
// @Security BearerAuth
func (c *NotificationController) UnregisterPushToken(ctx *gin.Context) {
	var req UnregisterPushTokenRequest
	if err := ctx.BindJSON(&req); err != nil {
		response.SendError(ctx, http.StatusBadRequest, "invalid request", nil)
		return
	}

	if err := c.pushService.UnregisterToken(ctx.Request.Context(), req.Token); err != nil {
		logger.Error("failed to unregister push token", "error", err)
		response.InternalError(ctx, "Failed to unregister push token")
		return
	}

	response.Success(ctx, nil, "Push token unregistered")
}

// SubscribePush godoc
// @Summary Subscribe to push notifications
// @Description Establish a WebSocket connection for real-time push notifications
// @Tags notifications
// @Produce json
// @Router /notifications/ws/push [get]
// @Security BearerAuth
func (c *NotificationController) SubscribePush(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}

	claims, err := jwt.ValidateToken(token, c.cfg.JWT.Secret, c.cfg.JWT.Issuer)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	ws, err := c.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.Error("websocket upgrade error", "error", err)
		return
	}
	defer ws.Close()

	msgChan := make(chan service.PushMessage, 10)
	subscriberID := uuid.New().String()

	subscribable, ok := c.pushService.(WebSocketSubscribable)
	if !ok {
		ws.WriteMessage(websocket.TextMessage,
			[]byte(`{"error":"Push service doesn't support WebSocket"}`))
		return
	}

	if err := subscribable.SubscribeToUser(userID, subscriberID, msgChan); err != nil {
		logger.Error("subscribe failed", "error", err)
		ws.WriteMessage(websocket.TextMessage,
			[]byte(`{"error":"Failed to subscribe"}`))
		return
	}
	defer func() {
		subscribable.UnsubscribeFromUser(userID, subscriberID)
		close(msgChan)
	}()

	// Pump incoming WebSocket messages (ping/close frames) so the connection
	// is properly torn down when the client disconnects.
	go func() {
		for {
			if _, _, err := ws.ReadMessage(); err != nil {
				subscribable.UnsubscribeFromUser(userID, subscriberID)
				return
			}
		}
	}()

	for msg := range msgChan {
		if err := ws.WriteJSON(msg); err != nil {
			break
		}
	}
}

// GetPushStats godoc
// @Summary Get push statistics
// @Description Get push service statistics (admin only)
// @Tags notifications
// @Produce json
// @Success 200 {object} response.Response
// @Router /notifications/stats [get]
func (c *NotificationController) GetPushStats(ctx *gin.Context) {
	type statsProvider interface {
		Stats() map[string]interface{}
	}

	if sp, ok := c.pushService.(statsProvider); ok {
		response.Success(ctx, sp.Stats(), "Push service statistics")
		return
	}

	response.Success(ctx, gin.H{"message": "Push service running"}, "Service status")
}
