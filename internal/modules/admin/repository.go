package admin

// now

import (
	"context"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	FindUserByID(ctx context.Context, id string) (*models.User, error)
	ListUsers(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.User, int64, error)
	UpdateUserStatus(ctx context.Context, userID string, status models.UserStatus) error
	DeleteUser(ctx context.Context, userID string) error
	GetDashboardStats(ctx context.Context) (map[string]interface{}, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindUserByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return &user, err
}

func (r *repository) ListUsers(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	query := r.db.WithContext(ctx).Model(&models.User{})


	for key, value := range filters {
		if value != "" {
			query = query.Where(key+" = ?", value)
		}
	}


	query.Count(&total)


	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&users).Error

	return users, total, err
}

func (r *repository) UpdateUserStatus(ctx context.Context, userID string, status models.UserStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("status", status).Error
}

func (r *repository) DeleteUser(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", userID).Error
}

func (r *repository) GetDashboardStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var roleCounts []struct {
		Role  string
		Count int64
	}
	r.db.WithContext(ctx).
		Model(&models.User{}).
		Select("role, COUNT(*) as count").
		Group("role").
		Scan(&roleCounts)

	stats["usersByRole"] = roleCounts

	// Count users by status
	var statusCounts []struct {
		Status string
		Count  int64
	}
	r.db.WithContext(ctx).
		Model(&models.User{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts)

	stats["usersByStatus"] = statusCounts

	var totalUsers int64
	r.db.WithContext(ctx).Model(&models.User{}).Count(&totalUsers)
	stats["totalUsers"] = totalUsers

	return stats, nil
}
