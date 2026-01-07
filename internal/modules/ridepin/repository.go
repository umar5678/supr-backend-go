// internal/modules/ridepin/repository.go
package ridepin

import (
	"context"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	VerifyRidePIN(ctx context.Context, userID, pin string) (bool, error)
	UpdateRidePIN(ctx context.Context, userID, newPIN string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	return &user, err
}

func (r *repository) VerifyRidePIN(ctx context.Context, userID, pin string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND ride_pin = ?", userID, pin).
		Count(&count).Error

	return count > 0, err
}

func (r *repository) UpdateRidePIN(ctx context.Context, userID, newPIN string) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("ride_pin", newPIN).Error
}
