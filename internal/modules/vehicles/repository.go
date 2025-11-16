package vehicles

import (
	"context"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context) ([]*models.VehicleType, error)
	FindActive(ctx context.Context) ([]*models.VehicleType, error)
	FindByID(ctx context.Context, id string) (*models.VehicleType, error)
	FindByName(ctx context.Context, name string) (*models.VehicleType, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context) ([]*models.VehicleType, error) {
	var vehicleTypes []*models.VehicleType
	err := r.db.WithContext(ctx).
		Order("base_fare ASC").
		Find(&vehicleTypes).Error
	return vehicleTypes, err
}

func (r *repository) FindActive(ctx context.Context) ([]*models.VehicleType, error) {
	var vehicleTypes []*models.VehicleType
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("base_fare ASC").
		Find(&vehicleTypes).Error
	return vehicleTypes, err
}

func (r *repository) FindByID(ctx context.Context, id string) (*models.VehicleType, error) {
	var vehicleType models.VehicleType
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&vehicleType).Error
	return &vehicleType, err
}

func (r *repository) FindByName(ctx context.Context, name string) (*models.VehicleType, error) {
	var vehicleType models.VehicleType
	err := r.db.WithContext(ctx).
		Where("name = ?", name).
		First(&vehicleType).Error
	return &vehicleType, err
}
