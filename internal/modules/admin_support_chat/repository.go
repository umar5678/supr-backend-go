package admin_support_chat

import (
	"context"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	SaveMessage(ctx context.Context, message *models.AdminSupportChat) error
	GetConversationMessages(ctx context.Context, conversationID string, limit, offset int) ([]*models.AdminSupportChat, int64, error)
	GetUserConversations(ctx context.Context, userID string, limit, offset int) ([]string, int64, error)
	MarkAsRead(ctx context.Context, messageID string, readAt ...interface{}) error
	GetUnreadCount(ctx context.Context, conversationID string) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) SaveMessage(ctx context.Context, message *models.AdminSupportChat) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *repository) GetConversationMessages(ctx context.Context, conversationID string, limit, offset int) ([]*models.AdminSupportChat, int64, error) {
	var messages []*models.AdminSupportChat
	var total int64

	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).
		Error

	if err != nil {
		return nil, 0, err
	}

	// Get total count
	r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Model(&models.AdminSupportChat{}).
		Count(&total)

	return messages, total, nil
}

func (r *repository) GetUserConversations(ctx context.Context, userID string, limit, offset int) ([]string, int64, error) {
	var conversationIDs []string
	var total int64

	err := r.db.WithContext(ctx).
		Model(&models.AdminSupportChat{}).
		Distinct("conversation_id").
		Where("sender_id = ?", userID).
		Order("MAX(created_at) DESC").
		Limit(limit).
		Offset(offset).
		Pluck("conversation_id", &conversationIDs).
		Error

	if err != nil {
		return nil, 0, err
	}

	// Get total count of unique conversations
	r.db.WithContext(ctx).
		Model(&models.AdminSupportChat{}).
		Distinct("conversation_id").
		Where("sender_id = ?", userID).
		Count(&total)

	return conversationIDs, total, nil
}

func (r *repository) MarkAsRead(ctx context.Context, messageID string, readAt ...interface{}) error {
	query := r.db.WithContext(ctx).
		Model(&models.AdminSupportChat{}).
		Where("id = ?", messageID)

	if len(readAt) > 0 && readAt[0] != nil {
		return query.Updates(map[string]interface{}{
			"is_read":          true,
			"read_by_admin_at": readAt[0],
		}).Error
	}

	return query.Update("is_read", true).Error
}

func (r *repository) GetUnreadCount(ctx context.Context, conversationID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Where("conversation_id = ? AND is_read = false", conversationID).
		Model(&models.AdminSupportChat{}).
		Count(&count).
		Error
	return count, err
}
