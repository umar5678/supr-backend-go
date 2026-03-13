package admin_support_chat

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service       Service
	broadcastFunc func(map[string]interface{}) error
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service:       service,
		broadcastFunc: nil,
	}
}

// SetBroadcastFunc sets the WebSocket broadcast function for real-time messaging
func (h *Handler) SetBroadcastFunc(fn func(map[string]interface{}) error) {
	h.broadcastFunc = fn
}

type SendMessageRequest struct {
	ConversationID string                 `json:"conversationId" binding:"required"`
	Content        string                 `json:"content" binding:"required,min=1"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type GetMessagesRequest struct {
	ConversationID string `json:"conversationId" binding:"required"`
	Page           int    `form:"page" binding:"omitempty,min=1" default:"1"`
	Limit          int    `form:"limit" binding:"omitempty,min=1,max=100" default:"50"`
}

// SendMessage godoc
// @Summary Send admin support chat message
// @Description Send a message in admin support chat
// @Tags Admin Support Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SendMessageRequest true "Message details"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /admin-support-chat/send [post]
func (h *Handler) SendMessage(c *gin.Context) {
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request: " + err.Error()))
		return
	}

	// Enforce conversation ID logic:
	// - Non-admins (riders, drivers, service providers, etc.) MUST use their userID as conversationID
	// - Admins reply to existing conversations and should pass the user's ID as conversationID
	conversationID := req.ConversationID
	if role != "admin" {
		// For non-admin users, conversation is always with themselves (1-to-1 with admins)
		conversationID = userID.(string)
	} else if req.ConversationID == "" {
		// Admin must specify which user's conversation to reply to
		c.Error(response.BadRequest("Admin must specify conversationId (user ID) to reply to"))
		return
	}

	message, err := h.service.SendMessage(
		c.Request.Context(),
		conversationID,
		userID.(string),
		role.(string),
		req.Content,
		req.Metadata,
	)
	if err != nil {
		c.Error(err)
		return
	}

	// Broadcast message via WebSocket to all clients if broadcast function is configured
	if h.broadcastFunc != nil {
		broadcastData := map[string]interface{}{
			"id":             message.ID,
			"conversationId": conversationID,
			"senderId":       userID,
			"senderRole":     role,
			"content":        req.Content,
			"metadata":       req.Metadata,
			"isRead":         message.IsRead,
			"createdAt":      message.CreatedAt,
			"timestamp":      message.CreatedAt,
		}

		if err := h.broadcastFunc(broadcastData); err != nil {
			logger.Warn("failed to broadcast admin support message via websocket", "error", err, "messageId", message.ID)
			// Don't return error - message was saved successfully even if broadcast failed
		}
	}

	response.Success(c, message, "Message sent successfully")
}

// GetConversationMessages godoc
// @Summary Get conversation messages
// @Description Retrieve all messages in a conversation
// @Tags Admin Support Chat
// @Produce json
// @Security BearerAuth
// @Param conversationId query string true "Conversation ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /admin-support-chat/conversations/:conversationId [get]
func (h *Handler) GetConversationMessages(c *gin.Context) {
	conversationID := c.Param("conversationId")

	// Validate conversationId is not empty and not malformed
	if conversationID == "" || conversationID == "[object Object]" {
		c.Error(response.BadRequest("Invalid conversationId parameter"))
		return
	}

	page := 1
	limit := 50

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			page = parsed
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	messages, total, err := h.service.GetConversationMessages(c.Request.Context(), conversationID, page, limit)
	if err != nil {
		c.Error(err)
		return
	}

	pagination := response.NewPaginationMeta(total, page, limit)
	response.Paginated(c, messages, pagination, "Messages retrieved successfully")
}

// GetUserConversations godoc
// @Summary Get user conversations with latest message
// @Description Retrieve all conversations for a user with preview of latest message
// @Tags Admin Support Chat
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /admin-support-chat/conversations [get]
func (h *Handler) GetUserConversations(c *gin.Context) {
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")
	page := 1
	limit := 50

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			page = parsed
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	conversations, total, err := h.service.GetUserConversationsWithDetails(c.Request.Context(), userID.(string), role.(string), page, limit)
	if err != nil {
		c.Error(err)
		return
	}

	pagination := response.NewPaginationMeta(total, page, limit)
	response.Paginated(c, conversations, pagination, "Conversations retrieved successfully")
}

// MarkAsRead godoc
// @Summary Mark message as read
// @Description Mark a message as read by admin
// @Tags Admin Support Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param messageId path string true "Message ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /admin-support-chat/:messageId/read [post]
func (h *Handler) MarkAsRead(c *gin.Context) {
	messageID := c.Param("messageId")

	// Validate messageId is not empty and not malformed
	if messageID == "" || messageID == "[object Object]" {
		c.Error(response.BadRequest("Invalid messageId parameter"))
		return
	}

	if err := h.service.MarkAsRead(c.Request.Context(), messageID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Message marked as read")
}

// ResolveConversation godoc
// @Summary Resolve a conversation
// @Description Admin endpoint to mark a conversation as resolved
// @Tags Admin Support Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param conversationId path string true "Conversation ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /admin-support-chat/conversations/:conversationId/resolve [post]
func (h *Handler) ResolveConversation(c *gin.Context) {
	conversationID := c.Param("conversationId")

	// Validate conversationId is not empty and not malformed
	if conversationID == "" || conversationID == "[object Object]" {
		c.Error(response.BadRequest("Invalid conversationId parameter"))
		return
	}

	if err := h.service.ResolveConversation(c.Request.Context(), conversationID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Conversation resolved successfully")
}

// DeleteConversation godoc
// @Summary Delete a conversation
// @Description Admin endpoint to delete (soft delete) a conversation and all its messages
// @Tags Admin Support Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param conversationId path string true "Conversation ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /admin-support-chat/conversations/:conversationId [delete]
func (h *Handler) DeleteConversation(c *gin.Context) {
	conversationID := c.Param("conversationId")

	// Validate conversationId is not empty and not malformed
	if conversationID == "" || conversationID == "[object Object]" {
		c.Error(response.BadRequest("Invalid conversationId parameter"))
		return
	}

	if err := h.service.DeleteConversation(c.Request.Context(), conversationID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Conversation deleted successfully")
}
