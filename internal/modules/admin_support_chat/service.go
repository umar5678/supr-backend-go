package admin_support_chat

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
	SendMessage(ctx context.Context, conversationID, senderID, senderRole, content string, metadata map[string]interface{}) (*models.AdminSupportChat, error)
	GetConversationMessages(ctx context.Context, conversationID string, page, limit int) ([]*models.AdminSupportChat, int64, error)
	GetUserConversations(ctx context.Context, userID string, page, limit int) ([]string, int64, error)
	GetUserConversationsWithDetails(ctx context.Context, userID, role string, page, limit int) ([]map[string]interface{}, int64, error)
	MarkAsRead(ctx context.Context, messageID string) error
	GetUnreadCount(ctx context.Context, conversationID string) (int64, error)
	GetConversation(ctx context.Context, conversationID string, limit, offset int) ([]*models.AdminSupportChat, error)
	ReplyToMessage(ctx context.Context, conversationID, parentMessageID, senderID, senderRole, content string, metadata map[string]interface{}) (*models.AdminSupportChat, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) SendMessage(ctx context.Context, conversationID, senderID, senderRole, content string, metadata map[string]interface{}) (*models.AdminSupportChat, error) {
	if conversationID == "" || senderID == "" || content == "" {
		return nil, response.BadRequest("conversationId, senderId, and content are required")
	}

	message := &models.AdminSupportChat{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		SenderID:       senderID,
		SenderRole:     senderRole,
		Content:        content,
		Metadata: models.AdminSupportChatMetadata{
			MessageID: uuid.New().String(),
			Extra:     metadata,
		},
		IsRead: false,
	}

	if err := s.repo.SaveMessage(ctx, message); err != nil {
		logger.Error("failed to save admin support chat message", "error", err, "conversationId", conversationID, "senderId", senderID)
		return nil, response.InternalServerError("Failed to save message", err)
	}

	logger.Info("admin support chat message saved",
		"messageId", message.ID,
		"conversationId", conversationID,
		"senderId", senderID,
		"senderRole", senderRole,
	)

	return message, nil
}

func (s *service) GetConversationMessages(ctx context.Context, conversationID string, page, limit int) ([]*models.AdminSupportChat, int64, error) {
	if conversationID == "" {
		return nil, 0, response.BadRequest("conversationId is required")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	messages, total, err := s.repo.GetConversationMessages(ctx, conversationID, limit, offset)
	if err != nil {
		logger.Error("failed to get conversation messages", "error", err, "conversationId", conversationID)
		return nil, 0, response.InternalServerError("Failed to retrieve messages", err)
	}

	return messages, total, nil
}

func (s *service) GetUserConversations(ctx context.Context, userID string, page, limit int) ([]string, int64, error) {
	if userID == "" {
		return nil, 0, response.BadRequest("userId is required")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	conversationIDs, total, err := s.repo.GetUserConversations(ctx, userID, limit, offset)
	if err != nil {
		logger.Error("failed to get user conversations", "error", err, "userId", userID)
		return nil, 0, response.InternalServerError("Failed to retrieve conversations", err)
	}

	return conversationIDs, total, nil
}

func (s *service) MarkAsRead(ctx context.Context, messageID string) error {
	if messageID == "" {
		return response.BadRequest("messageId is required")
	}

	if err := s.repo.MarkAsRead(ctx, messageID, time.Now().UTC()); err != nil {
		logger.Error("failed to mark message as read", "error", err, "messageId", messageID)
		return response.InternalServerError("Failed to mark message as read", err)
	}

	return nil
}

func (s *service) GetUnreadCount(ctx context.Context, conversationID string) (int64, error) {
	if conversationID == "" {
		return 0, response.BadRequest("conversationId is required")
	}

	count, err := s.repo.GetUnreadCount(ctx, conversationID)
	if err != nil {
		logger.Error("failed to get unread count", "error", err, "conversationId", conversationID)
		return 0, response.InternalServerError("Failed to get unread count", err)
	}

	return count, nil
}

func (s *service) GetConversation(ctx context.Context, conversationID string, limit, offset int) ([]*models.AdminSupportChat, error) {
	if conversationID == "" {
		return nil, response.BadRequest("conversationId is required")
	}

	if limit < 1 || limit > 100 {
		limit = 50
	}

	messages, _, err := s.repo.GetConversationMessages(ctx, conversationID, limit, offset)
	if err != nil {
		logger.Error("failed to get conversation messages", "error", err, "conversationId", conversationID)
		return nil, response.InternalServerError("Failed to retrieve messages", err)
	}

	return messages, nil
}

func (s *service) ReplyToMessage(ctx context.Context, conversationID, parentMessageID, senderID, senderRole, content string, metadata map[string]interface{}) (*models.AdminSupportChat, error) {
	if conversationID == "" || senderID == "" || content == "" {
		return nil, response.BadRequest("conversationId, senderId, and content are required")
	}

	message := &models.AdminSupportChat{
		ID:              uuid.New().String(),
		ConversationID:  conversationID,
		ParentMessageID: &parentMessageID,
		SenderID:        senderID,
		SenderRole:      senderRole,
		Content:         content,
		Metadata: models.AdminSupportChatMetadata{
			MessageID: uuid.New().String(),
			Extra:     metadata,
		},
		IsRead: false,
	}

	if err := s.repo.SaveMessage(ctx, message); err != nil {
		logger.Error("failed to save reply message", "error", err, "conversationId", conversationID, "senderId", senderID)
		return nil, response.InternalServerError("Failed to save reply", err)
	}

	logger.Info("admin support chat reply saved",
		"messageId", message.ID,
		"conversationId", conversationID,
		"parentMessageId", parentMessageID,
		"senderId", senderID,
		"senderRole", senderRole,
	)

	return message, nil
}

func (s *service) GetUserConversationsWithDetails(ctx context.Context, userID, role string, page, limit int) ([]map[string]interface{}, int64, error) {
	if userID == "" {
		return nil, 0, response.BadRequest("userId is required")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit

	// Get conversation IDs
	conversationIDs, total, err := s.repo.GetUserConversations(ctx, userID, limit, offset)
	if err != nil {
		logger.Error("failed to get user conversations", "error", err, "userId", userID)
		return nil, 0, response.InternalServerError("Failed to retrieve conversations", err)
	}

	// Build conversation details with latest message and metadata
	conversations := make([]map[string]interface{}, 0, len(conversationIDs))

	for _, conversationID := range conversationIDs {
		// Get latest message in conversation
		messages, msgTotal, err := s.repo.GetConversationMessages(ctx, conversationID, 1, 0)
		if err != nil || len(messages) == 0 {
			continue
		}

		latestMsg := messages[0]

		// Get unread count for this conversation
		unreadCount, _ := s.repo.GetUnreadCount(ctx, conversationID)

		conv := map[string]interface{}{
			"conversationId": conversationID,
			"lastMessage": map[string]interface{}{
				"content":    latestMsg.Content,
				"senderRole": latestMsg.SenderRole,
				"timestamp":  latestMsg.CreatedAt,
				"isRead":     latestMsg.IsRead,
			},
			"unreadCount":    unreadCount,
			"totalMessages":  msgTotal,
			"preview":        latestMsg.Content,
			"participantId":  conversationID, // For non-admins, conversationID is the other participant's ID
		}

		conversations = append(conversations, conv)
	}

	return conversations, total, nil
}
