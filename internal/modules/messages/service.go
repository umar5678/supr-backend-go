package messages

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
	SendMessage(ctx context.Context, rideID, senderID, senderType, content string, metadata map[string]interface{}) (*models.MessageResponse, error)
	GetMessages(ctx context.Context, rideID string, limit, offset int) ([]*models.MessageResponse, error)
	MarkAsRead(ctx context.Context, messageID, userID string) error
	DeleteMessage(ctx context.Context, messageID, userID string) error
	GetUnreadCount(ctx context.Context, rideID, userID string) (int64, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) SendMessage(ctx context.Context, rideID, senderID, senderType, content string, metadata map[string]interface{}) (*models.MessageResponse, error) {
	if rideID == "" || senderID == "" || content == "" {
		return nil, response.BadRequest("rideID, senderID, and content are required")
	}

	if senderType != "rider" && senderType != "driver" {
		return nil, response.BadRequest("senderType must be 'rider' or 'driver'")
	}

	msg := &models.RideMessage{
		ID:          uuid.New().String(),
		RideID:      rideID,
		SenderID:    senderID,
		SenderType:  senderType,
		MessageType: models.MessageTypeText,
		Content:     content,
		Metadata:    metadata,
		IsRead:      false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		logger.Error("failed to create message", "error", err, "rideID", rideID)
		return nil, response.InternalServerError("Failed to send message", err)
	}

	// Get sender name
	senderName, err := s.repo.GetSenderName(ctx, senderID)
	if err != nil {
		senderName = senderType
	}

	return &models.MessageResponse{
		ID:          msg.ID,
		RideID:      msg.RideID,
		SenderID:    msg.SenderID,
		SenderName:  senderName,
		SenderType:  msg.SenderType,
		MessageType: string(msg.MessageType),
		Content:     msg.Content,
		Metadata:    msg.Metadata,
		IsRead:      msg.IsRead,
		CreatedAt:   msg.CreatedAt,
	}, nil
}

func (s *service) GetMessages(ctx context.Context, rideID string, limit, offset int) ([]*models.MessageResponse, error) {
	if rideID == "" {
		return nil, response.BadRequest("rideID is required")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	messages, err := s.repo.GetMessages(ctx, rideID, limit, offset)
	if err != nil {
		logger.Error("failed to get messages", "error", err, "rideID", rideID)
		return nil, response.InternalServerError("Failed to fetch messages", err)
	}

	responses := make([]*models.MessageResponse, len(messages))
	for i, msg := range messages {
		senderName, _ := s.repo.GetSenderName(ctx, msg.SenderID)
		responses[i] = &models.MessageResponse{
			ID:          msg.ID,
			RideID:      msg.RideID,
			SenderID:    msg.SenderID,
			SenderName:  senderName,
			SenderType:  msg.SenderType,
			MessageType: string(msg.MessageType),
			Content:     msg.Content,
			Metadata:    msg.Metadata,
			IsRead:      msg.IsRead,
			ReadAt:      msg.ReadAt,
			CreatedAt:   msg.CreatedAt,
		}
	}

	return responses, nil
}

func (s *service) MarkAsRead(ctx context.Context, messageID, userID string) error {
	if messageID == "" || userID == "" {
		return response.BadRequest("messageID and userID are required")
	}

	now := time.Now()
	if err := s.repo.UpdateMessage(ctx, messageID, map[string]interface{}{
		"is_read": true,
		"read_at": now,
	}); err != nil {
		logger.Error("failed to mark message as read", "error", err)
		return response.InternalServerError("Failed to mark as read", err)
	}

	return nil
}

func (s *service) DeleteMessage(ctx context.Context, messageID, userID string) error {
	if messageID == "" || userID == "" {
		return response.BadRequest("messageID and userID are required")
	}

	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return response.NotFoundError("Message not found")
	}

	if msg.SenderID != userID {
		return response.ForbiddenError("You can only delete your own messages")
	}

	if time.Since(msg.CreatedAt) > 5*time.Minute {
		return response.BadRequest("Can only delete messages within 5 minutes")
	}

	if err := s.repo.DeleteMessage(ctx, messageID); err != nil {
		logger.Error("failed to delete message", "error", err)
		return response.InternalServerError("Failed to delete message", err)
	}

	return nil
}

func (s *service) GetUnreadCount(ctx context.Context, rideID, userID string) (int64, error) {
	if rideID == "" || userID == "" {
		return 0, response.BadRequest("rideID and userID are required")
	}

	count, err := s.repo.CountUnreadMessages(ctx, rideID, userID)
	if err != nil {
		logger.Error("failed to get unread count", "error", err)
		return 0, response.InternalServerError("Failed to get unread count", err)
	}

	return count, nil
}
