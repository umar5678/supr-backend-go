package ratings

import (
	"context"

	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
)

type Repository interface {
	Create(ctx context.Context, rating *models.Rating) error
	FindByOrderID(ctx context.Context, orderID string) (*models.Rating, error)
	GetProviderRatings(ctx context.Context, providerID string, limit int) ([]models.Rating, error)
	GetProviderAverageRating(ctx context.Context, providerID string) (float64, error)
	UpdateProviderRating(ctx context.Context, providerID string, newAverage float64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, rating *models.Rating) error {
	return r.db.WithContext(ctx).Create(rating).Error
}

func (r *repository) FindByOrderID(ctx context.Context, orderID string) (*models.Rating, error) {
	var rating models.Rating
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&rating).Error
	return &rating, err
}

func (r *repository) GetProviderRatings(ctx context.Context, providerID string, limit int) ([]models.Rating, error) {
	var ratings []models.Rating
	err := r.db.WithContext(ctx).
		Where("provider_id = ?", providerID).
		Order("created_at DESC").
		Limit(limit).
		Find(&ratings).Error
	return ratings, err
}

func (r *repository) GetProviderAverageRating(ctx context.Context, providerID string) (float64, error) {
	var avg float64
	err := r.db.WithContext(ctx).
		Model(&models.Rating{}).
		Where("provider_id = ?", providerID).
		Select("COALESCE(AVG(score), 5.0)").
		Scan(&avg).Error
	return avg, err
}

func (r *repository) UpdateProviderRating(ctx context.Context, providerID string, newAverage float64) error {
	return r.db.WithContext(ctx).
		Model(&models.ServiceProvider{}).
		Where("id = ?", providerID).
		Update("rating", newAverage).Error
}
