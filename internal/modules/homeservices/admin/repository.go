package admin

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/admin/dto"
)

// Repository defines the interface for admin home services data access
type Repository interface {
	// Service operations
	CreateService(ctx context.Context, service *models.ServiceNew) error
	GetServiceByID(ctx context.Context, id string) (*models.ServiceNew, error)
	GetServiceBySlug(ctx context.Context, slug string) (*models.ServiceNew, error)
	UpdateService(ctx context.Context, service *models.ServiceNew) error
	// UpdateHomeCleaningService(ctx context.Context, service *models.HomeCleaningService) error
	DeleteService(ctx context.Context, id string) error
	ListServices(ctx context.Context, query dto.ListServicesQuery) ([]*models.ServiceNew, int64, error)
	ServiceSlugExists(ctx context.Context, slug string, excludeID string) (bool, error)

	// Addon operations
	CreateAddon(ctx context.Context, addon *models.Addon) error
	GetAddonByID(ctx context.Context, id string) (*models.Addon, error)
	GetAddonBySlug(ctx context.Context, slug string) (*models.Addon, error)
	UpdateAddon(ctx context.Context, addon *models.Addon) error
	DeleteAddon(ctx context.Context, id string) error
	ListAddons(ctx context.Context, query dto.ListAddonsQuery) ([]*models.Addon, int64, error)
	AddonSlugExists(ctx context.Context, slug string, excludeID string) (bool, error)

	// Category operations
	GetServicesByCategory(ctx context.Context, categorySlug string) ([]*models.ServiceNew, error)
	GetAddonsByCategory(ctx context.Context, categorySlug string) ([]*models.Addon, error)
	GetAllCategories(ctx context.Context) ([]string, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new admin repository instance
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// ==================== Service Operations ====================

func (r *repository) CreateService(ctx context.Context, service *models.ServiceNew) error {
	return r.db.WithContext(ctx).Create(service).Error
}

func (r *repository) GetServiceByID(ctx context.Context, id string) (*models.ServiceNew, error) {
	var service models.ServiceNew
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&service).Error
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *repository) GetServiceBySlug(ctx context.Context, slug string) (*models.ServiceNew, error) {
	var service models.ServiceNew
	err := r.db.WithContext(ctx).Where("service_slug = ?", slug).First(&service).Error
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *repository) UpdateService(ctx context.Context, service *models.ServiceNew) error {
	return r.db.WithContext(ctx).Save(service).Error
}

// func (r *repository) UpdateHomeCleaningService(ctx context.Context, service *models.HomeCleaningService) error {
// 	return r.db.WithContext(ctx).Save(service).Error
// }

func (r *repository) DeleteService(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.ServiceNew{}, "id = ?", id).Error
}

func (r *repository) ListServices(ctx context.Context, query dto.ListServicesQuery) ([]*models.ServiceNew, int64, error) {
	var services []*models.ServiceNew
	var total int64

	db := r.db.WithContext(ctx).Model(&models.ServiceNew{})

	// Apply filters
	if query.CategorySlug != "" {
		db = db.Where("category_slug = ?", query.CategorySlug)
	}

	if query.Search != "" {
		searchPattern := "%" + strings.ToLower(query.Search) + "%"
		db = db.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)
	}

	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}

	if query.IsAvailable != nil {
		db = db.Where("is_available = ?", *query.IsAvailable)
	}

	// Count total before pagination
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderClause := query.SortBy
	if query.SortDesc {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}
	db = db.Order(orderClause)

	// Apply pagination
	offset := query.PaginationParams.GetOffset()
	if err := db.Offset(offset).Limit(query.Limit).Find(&services).Error; err != nil {
		return nil, 0, err
	}

	return services, total, nil
}

func (r *repository) ServiceSlugExists(ctx context.Context, slug string, excludeID string) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&models.ServiceNew{}).Where("service_slug = ?", slug)

	if excludeID != "" {
		db = db.Where("id != ?", excludeID)
	}

	if err := db.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ==================== Addon Operations ====================

func (r *repository) CreateAddon(ctx context.Context, addon *models.Addon) error {
	return r.db.WithContext(ctx).Create(addon).Error
}

func (r *repository) GetAddonByID(ctx context.Context, id string) (*models.Addon, error) {
	var addon models.Addon
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&addon).Error
	if err != nil {
		return nil, err
	}
	return &addon, nil
}

func (r *repository) GetAddonBySlug(ctx context.Context, slug string) (*models.Addon, error) {
	var addon models.Addon
	err := r.db.WithContext(ctx).Where("addon_slug = ?", slug).First(&addon).Error
	if err != nil {
		return nil, err
	}
	return &addon, nil
}

func (r *repository) UpdateAddon(ctx context.Context, addon *models.Addon) error {
	return r.db.WithContext(ctx).Save(addon).Error
}

func (r *repository) DeleteAddon(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.Addon{}, "id = ?", id).Error
}

func (r *repository) ListAddons(ctx context.Context, query dto.ListAddonsQuery) ([]*models.Addon, int64, error) {
	var addons []*models.Addon
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Addon{})

	// Apply filters
	if query.CategorySlug != "" {
		db = db.Where("category_slug = ?", query.CategorySlug)
	}

	if query.Search != "" {
		searchPattern := "%" + strings.ToLower(query.Search) + "%"
		db = db.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)
	}

	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}

	if query.IsAvailable != nil {
		db = db.Where("is_available = ?", *query.IsAvailable)
	}

	if query.MinPrice != nil {
		db = db.Where("price >= ?", *query.MinPrice)
	}

	if query.MaxPrice != nil {
		db = db.Where("price <= ?", *query.MaxPrice)
	}

	// Count total before pagination
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderClause := query.SortBy
	if query.SortDesc {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}
	db = db.Order(orderClause)

	// Apply pagination
	offset := query.PaginationParams.GetOffset()
	if err := db.Offset(offset).Limit(query.Limit).Find(&addons).Error; err != nil {
		return nil, 0, err
	}

	return addons, total, nil
}

func (r *repository) AddonSlugExists(ctx context.Context, slug string, excludeID string) (bool, error) {
	var count int64
	db := r.db.WithContext(ctx).Model(&models.Addon{}).Where("addon_slug = ?", slug)

	if excludeID != "" {
		db = db.Where("id != ?", excludeID)
	}

	if err := db.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ==================== Category Operations ====================

func (r *repository) GetServicesByCategory(ctx context.Context, categorySlug string) ([]*models.ServiceNew, error) {
	var services []*models.ServiceNew
	err := r.db.WithContext(ctx).
		Where("category_slug = ?", categorySlug).
		Order("sort_order ASC, title ASC").
		Find(&services).Error
	return services, err
}

func (r *repository) GetAddonsByCategory(ctx context.Context, categorySlug string) ([]*models.Addon, error) {
	var addons []*models.Addon
	err := r.db.WithContext(ctx).
		Where("category_slug = ?", categorySlug).
		Order("sort_order ASC, title ASC").
		Find(&addons).Error
	return addons, err
}

func (r *repository) GetAllCategories(ctx context.Context) ([]string, error) {
	var categories []string

	// Get unique categories from services
	var serviceCategories []string
	if err := r.db.WithContext(ctx).
		Model(&models.ServiceNew{}).
		Distinct("category_slug").
		Pluck("category_slug", &serviceCategories).Error; err != nil {
		return nil, err
	}

	// Get unique categories from addons
	var addonCategories []string
	if err := r.db.WithContext(ctx).
		Model(&models.Addon{}).
		Distinct("category_slug").
		Pluck("category_slug", &addonCategories).Error; err != nil {
		return nil, err
	}

	// Merge and deduplicate
	categoryMap := make(map[string]bool)
	for _, c := range serviceCategories {
		categoryMap[c] = true
	}
	for _, c := range addonCategories {
		categoryMap[c] = true
	}

	for c := range categoryMap {
		categories = append(categories, c)
	}

	return categories, nil
}
