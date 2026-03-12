package adminsupport

import (
	"context"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	CreateMessage(ctx context.Context, msg *models.AdminSupportChat) error
	GetConversation(ctx context.Context, conversationID string, limit, offset int) ([]*models.AdminSupportChat, error)
	GetConversationCount(ctx context.Context, conversationID string) (int64, error)
	MarkAsRead(ctx context.Context, messageID string) error
	MarkConversationAsRead(ctx context.Context, conversationID string) error
	GetMessageByID(ctx context.Context, messageID string) (*models.AdminSupportChat, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateMessage(ctx context.Context, msg *models.AdminSupportChat) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *repository) GetConversation(ctx context.Context, conversationID string, limit, offset int) ([]*models.AdminSupportChat, error) {
	var messages []*models.AdminSupportChat
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", conversationID).
		Where("deleted_at IS NULL").
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

func (r *repository) GetConversationCount(ctx context.Context, conversationID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.AdminSupportChat{}).
		Where("conversation_id = ?", conversationID).
		Where("deleted_at IS NULL").
		Count(&count).Error
	return count, err
}

func (r *repository) MarkAsRead(ctx context.Context, messageID string) error {
	now := gorm.Expr("NOW()")
	return r.db.WithContext(ctx).
		Model(&models.AdminSupportChat{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"is_read":          true,
			"read_by_admin_at": now,
		}).Error
}

func (r *repository) MarkConversationAsRead(ctx context.Context, conversationID string) error {
	now := gorm.Expr("NOW()")
	return r.db.WithContext(ctx).
		Model(&models.AdminSupportChat{}).
		Where("conversation_id = ?", conversationID).
		Where("is_read = ?", false).
		Updates(map[string]interface{}{
			"is_read":          true,
			"read_by_admin_at": now,
		}).Error
}

func (r *repository) GetMessageByID(ctx context.Context, messageID string) (*models.AdminSupportChat, error) {
	var message *models.AdminSupportChat
	err := r.db.WithContext(ctx).
		Where("id = ?", messageID).
		First(&message).Error
	return message, err
}
