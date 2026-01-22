package messages

import (
    "context"
    "github.com/umar5678/go-backend/internal/models"
    "gorm.io/gorm"
)

type Repository interface {
    CreateMessage(ctx context.Context, msg *models.RideMessage) error
    GetMessages(ctx context.Context, rideID string, limit, offset int) ([]*models.RideMessage, error)
    GetMessageByID(ctx context.Context, messageID string) (*models.RideMessage, error)
    UpdateMessage(ctx context.Context, messageID string, updates map[string]interface{}) error
    DeleteMessage(ctx context.Context, messageID string) error
    CountUnreadMessages(ctx context.Context, rideID, userID string) (int64, error)
    GetSenderName(ctx context.Context, userID string) (string, error)
}

type repository struct {
    db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
    return &repository{db: db}
}

func (r *repository) CreateMessage(ctx context.Context, msg *models.RideMessage) error {
    return r.db.WithContext(ctx).Create(msg).Error
}

func (r *repository) GetMessages(ctx context.Context, rideID string, limit, offset int) ([]*models.RideMessage, error) {
    var messages []*models.RideMessage
    err := r.db.WithContext(ctx).
        Where("ride_id = ?", rideID).
        Where("deleted_at IS NULL").
        Order("created_at DESC").
        Limit(limit).
        Offset(offset).
        Find(&messages).Error
    return messages, err
}

func (r *repository) GetMessageByID(ctx context.Context, messageID string) (*models.RideMessage, error) {
    var msg *models.RideMessage
    err := r.db.WithContext(ctx).
        Where("id = ?", messageID).
        Where("deleted_at IS NULL").
        First(&msg).Error
    return msg, err
}

func (r *repository) UpdateMessage(ctx context.Context, messageID string, updates map[string]interface{}) error {
    return r.db.WithContext(ctx).
        Model(&models.RideMessage{}).
        Where("id = ?", messageID).
        Updates(updates).Error
}

func (r *repository) DeleteMessage(ctx context.Context, messageID string) error {
    return r.db.WithContext(ctx).
        Model(&models.RideMessage{}).
        Where("id = ?", messageID).
        Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

func (r *repository) CountUnreadMessages(ctx context.Context, rideID, userID string) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&models.RideMessage{}).
        Where("ride_id = ?", rideID).
        Where("sender_id != ?", userID).
        Where("is_read = ?", false).
        Where("deleted_at IS NULL").
        Count(&count).Error
    return count, err
}

func (r *repository) GetSenderName(ctx context.Context, userID string) (string, error) {
    var user struct {
        Name string
    }
    err := r.db.WithContext(ctx).
        Table("users").
        Where("id = ?", userID).
        Select("name").
        First(&user).Error
    return user.Name, err
}