package riders

import (
	"context"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, profile *models.RiderProfile) error
	FindByID(ctx context.Context, id string) (*models.RiderProfile, error)
	FindByUserID(ctx context.Context, userID string) (*models.RiderProfile, error)
	Update(ctx context.Context, profile *models.RiderProfile) error
	Delete(ctx context.Context, id string) error

	// Statistics
	IncrementTotalRides(ctx context.Context, userID string) error
	UpdateRating(ctx context.Context, userID string, newRating float64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, profile *models.RiderProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*models.RiderProfile, error) {
	var profile models.RiderProfile
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Wallet").
		Where("id = ?", id).
		First(&profile).Error
	return &profile, err
}

func (r *repository) FindByUserID(ctx context.Context, userID string) (*models.RiderProfile, error) {
	var profile models.RiderProfile
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Wallet").
		Where("user_id = ?", userID).
		First(&profile).Error
	return &profile, err
}

func (r *repository) Update(ctx context.Context, profile *models.RiderProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.RiderProfile{}, "id = ?", id).Error
}

func (r *repository) IncrementTotalRides(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).
		Model(&models.RiderProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("total_rides", gorm.Expr("total_rides + ?", 1)).Error
}

func (r *repository) UpdateRating(ctx context.Context, userID string, newRating float64) error {
	return r.db.WithContext(ctx).
		Model(&models.RiderProfile{}).
		Where("user_id = ?", userID).
		Update("rating", newRating).Error
}
