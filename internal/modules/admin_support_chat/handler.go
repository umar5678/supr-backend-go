package admin_support_chat

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/utils/response"
	"github.com/umar5678/go-backend/internal/websocket"
)

type Handler struct {
	service Service
	wsHub   *websocket.Hub
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func NewHandlerWithWebSocket(service Service, wsHub *websocket.Hub) *Handler {
	return &Handler{
		service: service,
		wsHub:   wsHub,
	}
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
// @Success 200 {object} response.Response{data=models.AdminSupportChat}
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

	// Broadcast message to conversation owner and admins in real-time if WebSocket hub is available
	if h.wsHub != nil {
		broadcastData := map[string]interface{}{
			"senderId":       userID,
			"senderRole":     role,
			"content":        req.Content,
			"metadata":       req.Metadata,
			"conversationId": conversationID,
			"messageId":      message.ID,
			"timestamp":      message.CreatedAt,
		}

		// Send to the user in the conversation (they will see the admin reply)
		msg := websocket.NewTargetedMessage(websocket.TypeChatMessage, conversationID, broadcastData)
		h.wsHub.SendToUser(conversationID, msg)

		// Also broadcast to all admins so they see the sent message
		broadcastMsg := websocket.NewMessage(websocket.TypeChatMessage, broadcastData)
		h.wsHub.BroadcastToAll(broadcastMsg)
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
// @Success 200 {object} response.Response{data=[]models.AdminSupportChat}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /admin-support-chat/conversations/:conversationId [get]
func (h *Handler) GetConversationMessages(c *gin.Context) {
	conversationID := c.Param("conversationId")
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

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	response.Paginated(c, messages, response.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, "Messages retrieved successfully")
}

// GetUserConversations godoc
// @Summary Get user conversations with latest message
// @Description Retrieve all conversations for a user with preview of latest message
// @Tags Admin Support Chat
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(50)
// @Success 200 {object} response.Response{data=[]map[string]interface{}}
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

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	response.Paginated(c, conversations, response.PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, "Conversations retrieved successfully")
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

	if err := h.service.MarkAsRead(c.Request.Context(), messageID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Message marked as read")
}
