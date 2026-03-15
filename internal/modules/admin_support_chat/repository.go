package admin_support_chat

import (
	"context"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	SaveMessage(ctx context.Context, message *models.AdminSupportChat) error
	GetConversationMessages(ctx context.Context, conversationID string, limit, offset int) ([]*models.AdminSupportChat, int64, error)
	GetUserConversations(ctx context.Context, userID string, limit, offset int) ([]string, int64, error)
	GetAllConversations(ctx context.Context, limit, offset int) ([]string, int64, error)
	GetLatestMessage(ctx context.Context, conversationID string) (*models.AdminSupportChat, error)
	MarkAsRead(ctx context.Context, messageID string, readAt ...interface{}) error
	GetUnreadCount(ctx context.Context, conversationID string) (int64, error)
	IsConversationResolved(ctx context.Context, conversationID string) (bool, *time.Time, error)
	ResolveConversation(ctx context.Context, conversationID string) error
	DeleteConversation(ctx context.Context, conversationID string) error
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

	baseQuery := r.db.WithContext(ctx).
		Model(&models.AdminSupportChat{}).
		Where("conversation_id = ?", conversationID)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*models.AdminSupportChat{}, 0, nil
	}

	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error

	return messages, total, err
}

func (r *repository) GetUserConversations(ctx context.Context, userID string, limit, offset int) ([]string, int64, error) {
	var conversationIDs []string
	var total int64

	err := r.db.WithContext(ctx).
		Model(&models.AdminSupportChat{}).
		Where("conversation_id = ?", userID).
		Select("conversation_id").
		Group("conversation_id").
		Order("MAX(created_at) DESC").
		Limit(limit).
		Offset(offset).
		Pluck("conversation_id", &conversationIDs).
		Error

	if err != nil {
		return nil, 0, err
	}

	r.db.WithContext(ctx).
		Model(&models.AdminSupportChat{}).
		Where("conversation_id = ?", userID).
		Distinct("conversation_id").
		Count(&total)

	return conversationIDs, total, nil
}

func (r *repository) GetAllConversations(ctx context.Context, limit, offset int) ([]string, int64, error) {
	var conversationIDs []string
	var total int64

	err := r.db.WithContext(ctx).
		Model(&models.AdminSupportChat{}).
		Select("conversation_id").
		Group("conversation_id").
		Order("MAX(created_at) DESC").
		Limit(limit).
		Offset(offset).
		Pluck("conversation_id", &conversationIDs).
		Error

	if err != nil {
		return nil, 0, err
	}

	r.db.WithContext(ctx).
		Model(&models.AdminSupportChat{}).
		Distinct("conversation_id").
		Count(&total)

	return conversationIDs, total, nil
}

func (r *repository) GetLatestMessage(ctx context.Context, conversationID string) (*models.AdminSupportChat, error) {
	var message models.AdminSupportChat
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *repository) IsConversationResolved(ctx context.Context, conversationID string) (bool, *time.Time, error) {
	var message models.AdminSupportChat
	err := r.db.WithContext(ctx).
		Where("conversation_id = ? AND is_resolved = ?", conversationID, true).
		Order("resolved_at DESC").
		First(&message).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		return false, nil, err
	}

	return true, message.ResolvedAt, nil
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
		Model(&models.AdminSupportChat{}).
		Where("conversation_id = ? AND is_read = ? AND sender_role != ?", conversationID, false, "admin").
		Count(&count).Error
	return count, err
}

func (r *repository) ResolveConversation(ctx context.Context, conversationID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.AdminSupportChat{}).
		Where("conversation_id = ?", conversationID).
		Updates(map[string]interface{}{
			"is_resolved": true,
			"resolved_at": now,
		}).Error
}

func (r *repository) DeleteConversation(ctx context.Context, conversationID string) error {
	return r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Delete(&models.AdminSupportChat{}).Error
}