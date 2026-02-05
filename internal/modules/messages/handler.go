package messages

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetMessages godoc
// @Summary Get messages for a ride
// @Tags messages
// @Security BearerAuth
// @Produce json
// @Param rideId path string true "Ride ID"
// @Param limit query int false "Limit (default: 50)"
// @Param offset query int false "Offset (default: 0)"
// @Success 200 {object} response.Response{data=dto.MessageResponse}
// @Router /messages/rides/{rideId} [get]
func (h *Handler) GetMessages(c *gin.Context) {
	userID, _ := c.Get("userID")
	rideID := c.Param("rideId")

	if rideID == "" {
		c.Error(response.BadRequest("Ride ID is required"))
		return
	}

	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	messages, err := h.service.GetMessages(c.Request.Context(), rideID, limit, offset)
	if err != nil {
		logger.Error("failed to get messages", "error", err, "rideID", rideID, "userID", userID)
		c.Error(err)
		return
	}

	response.Success(c, messages, "Messages retrieved successfully")
}

// SendMessage godoc
// @Summary Send a message
// @Tags messages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body SendMessageRequest true "Message data"
// @Success 201 {object} response.Response{data=dto.MessageResponse}
// @Router /messages [post]
func (h *Handler) SendMessage(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	if req.RideID == "" || req.Content == "" {
		c.Error(response.BadRequest("RideID and content are required"))
		return
	}

	senderType := "rider"
	if val, exists := c.Get("userRole"); exists {
		if role, ok := val.(string); ok && role == "driver" {
			senderType = "driver"
		}
	}

	msgResp, err := h.service.SendMessage(c.Request.Context(), req.RideID, userID.(string), senderType, req.Content, req.Metadata)
	if err != nil {
		logger.Error("failed to send message", "error", err, "userID", userID)
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Data:    msgResp,
		Message: "Message sent successfully",
	})
}

// MarkAsRead godoc
// @Summary Mark message as read
// @Tags messages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param messageId path string true "Message ID"
// @Success 200 {object} response.Response
// @Router /messages/{messageId}/read [post]
func (h *Handler) MarkAsRead(c *gin.Context) {
	userID, _ := c.Get("userID")
	messageID := c.Param("messageId")

	if messageID == "" {
		c.Error(response.BadRequest("Message ID is required"))
		return
	}

	if err := h.service.MarkAsRead(c.Request.Context(), messageID, userID.(string)); err != nil {
		logger.Error("failed to mark as read", "error", err, "messageID", messageID)
		c.Error(err)
		return
	}

	response.Success(c, nil, "Message marked as read")
}

// DeleteMessage godoc
// @Summary Delete a message
// @Tags messages
// @Security BearerAuth
// @Produce json
// @Param messageId path string true "Message ID"
// @Success 200 {object} response.Response
// @Router /messages/{messageId} [delete]
func (h *Handler) DeleteMessage(c *gin.Context) {
	userID, _ := c.Get("userID")
	messageID := c.Param("messageId")

	if messageID == "" {
		c.Error(response.BadRequest("Message ID is required"))
		return
	}

	if err := h.service.DeleteMessage(c.Request.Context(), messageID, userID.(string)); err != nil {
		logger.Error("failed to delete message", "error", err, "messageID", messageID)
		c.Error(err)
		return
	}

	response.Success(c, nil, "Message deleted successfully")
}

// GetUnreadCount godoc
// @Summary Get unread message count for a ride
// @Tags messages
// @Security BearerAuth
// @Produce json
// @Param rideId path string true "Ride ID"
// @Success 200 {object} response.Response{data=int}
// @Router /messages/rides/{rideId}/unread-count [get]
func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID, _ := c.Get("userID")
	rideID := c.Param("rideId")

	if rideID == "" {
		c.Error(response.BadRequest("Ride ID is required"))
		return
	}

	count, err := h.service.GetUnreadCount(c.Request.Context(), rideID, userID.(string))
	if err != nil {
		logger.Error("failed to get unread count", "error", err)
		c.Error(err)
		return
	}

	response.Success(c, map[string]interface{}{
		"rideId":      rideID,
		"unreadCount": count,
	}, "Unread count retrieved successfully")
}

// DTOs
type SendMessageRequest struct {
	RideID   string                 `json:"rideId" binding:"required"`
	Content  string                 `json:"content" binding:"required"`
	Metadata map[string]interface{} `json:"metadata"`
}
