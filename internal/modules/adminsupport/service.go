package adminsupport

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

type Service interface {
	SendMessage(ctx context.Context, conversationID string, senderID string, senderRole string, content string) (*models.AdminSupportChat, error)
	GetConversation(ctx context.Context, conversationID string, limit, offset int) ([]*models.AdminSupportChat, error)
	ReplyToMessage(ctx context.Context, conversationID string, parentMessageID string, senderID string, senderRole string, content string) (*models.AdminSupportChat, error)
	MarkAsRead(ctx context.Context, messageID string) error
	MarkConversationAsRead(ctx context.Context, conversationID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) SendMessage(ctx context.Context, conversationID string, senderID string, senderRole string, content string) (*models.AdminSupportChat, error) {
	if conversationID == "" {
		return nil, fmt.Errorf("conversation ID is required")
	}
	if senderID == "" {
		return nil, fmt.Errorf("sender ID is required")
	}
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	msg := &models.AdminSupportChat{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		SenderID:       senderID,
		SenderRole:     senderRole,
		Content:        content,
	}

	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		logger.Error("failed to create admin support chat message", "error", err, "conversationId", conversationID)
		return nil, err
	}

	logger.Info("admin support chat message created",
		"messageId", msg.ID,
		"conversationId", conversationID,
		"senderId", senderID,
		"senderRole", senderRole,
	)

	return msg, nil
}

func (s *service) GetConversation(ctx context.Context, conversationID string, limit, offset int) ([]*models.AdminSupportChat, error) {
	if limit == 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	return s.repo.GetConversation(ctx, conversationID, limit, offset)
}

func (s *service) ReplyToMessage(ctx context.Context, conversationID string, parentMessageID string, senderID string, senderRole string, content string) (*models.AdminSupportChat, error) {
	// Verify parent message exists
	if parentMessageID != "" {
		parent, err := s.repo.GetMessageByID(ctx, parentMessageID)
		if err != nil {
			logger.Error("parent message not found", "error", err, "parentMessageId", parentMessageID)
			return nil, fmt.Errorf("parent message not found")
		}
		// Ensure parent message is in the same conversation
		if parent.ConversationID != conversationID {
			return nil, fmt.Errorf("parent message does not belong to this conversation")
		}
	}

	// Create reply with same conversation ID and parent message reference
	msg := &models.AdminSupportChat{
		ID:              uuid.New().String(),
		ConversationID:  conversationID,
		ParentMessageID: &parentMessageID,
		SenderID:        senderID,
		SenderRole:      senderRole,
		Content:         content,
	}

	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		logger.Error("failed to create admin support chat reply", "error", err, "conversationId", conversationID)
		return nil, err
	}

	logger.Info("admin support chat reply created",
		"messageId", msg.ID,
		"conversationId", conversationID,
		"parentMessageId", parentMessageID,
		"senderId", senderID,
	)

	return msg, nil
}

func (s *service) MarkAsRead(ctx context.Context, messageID string) error {
	return s.repo.MarkAsRead(ctx, messageID)
}

func (s *service) MarkConversationAsRead(ctx context.Context, conversationID string) error {
	return s.repo.MarkConversationAsRead(ctx, conversationID)
}
