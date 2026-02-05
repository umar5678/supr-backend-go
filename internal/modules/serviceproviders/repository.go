package serviceproviders

import (
	"context"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	CreateProfile(ctx context.Context, profile *models.ServiceProviderProfile) error
	FindByUserID(ctx context.Context, userID string) (*models.ServiceProviderProfile, error)
	FindByID(ctx context.Context, id string) (*models.ServiceProviderProfile, error)
	Update(ctx context.Context, profile *models.ServiceProviderProfile) error
	List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.ServiceProviderProfile, int64, error)
	UpdateStatus(ctx context.Context, id string, status models.ServiceProviderStatus) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateProfile(ctx context.Context, profile *models.ServiceProviderProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *repository) FindByUserID(ctx context.Context, userID string) (*models.ServiceProviderProfile, error) {
	var profile models.ServiceProviderProfile
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("user_id = ?", userID).
		First(&profile).Error
	return &profile, err
}

func (r *repository) FindByID(ctx context.Context, id string) (*models.ServiceProviderProfile, error) {
	var profile models.ServiceProviderProfile
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("id = ?", id).
		First(&profile).Error
	return &profile, err
}

func (r *repository) Update(ctx context.Context, profile *models.ServiceProviderProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

func (r *repository) List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.ServiceProviderProfile, int64, error) {
	var profiles []*models.ServiceProviderProfile
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ServiceProviderProfile{}).Preload("User")

	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&profiles).Error

	return profiles, total, err
}

func (r *repository) UpdateStatus(ctx context.Context, id string, status models.ServiceProviderStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.ServiceProviderProfile{}).
		Where("id = ?", id).
		Update("status", status).Error
}
