package customer

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/customer/dto"
)

// Repository defines the interface for customer home services data access
type Repository interface {
	// Service operations
	GetActiveServiceBySlug(ctx context.Context, slug string) (*models.ServiceNew, error)
	ListActiveServices(ctx context.Context, query dto.ListServicesQuery) ([]*models.ServiceNew, int64, error)
	GetActiveServicesByCategory(ctx context.Context, categorySlug string) ([]*models.ServiceNew, error)
	CountActiveServicesByCategory(ctx context.Context, categorySlug string) (int64, error)
	GetFrequentServices(ctx context.Context, limit int) ([]*models.ServiceNew, error)

	// Addon operations
	GetActiveAddonBySlug(ctx context.Context, slug string) (*models.Addon, error)
	ListActiveAddons(ctx context.Context, query dto.ListAddonsQuery) ([]*models.Addon, int64, error)
	GetActiveAddonsByCategory(ctx context.Context, categorySlug string) ([]*models.Addon, error)
	CountActiveAddonsByCategory(ctx context.Context, categorySlug string) (int64, error)
	GetDiscountedAddons(ctx context.Context, limit int) ([]*models.Addon, error)

	// Category operations
	GetAllActiveCategories(ctx context.Context) ([]CategoryInfo, error)

	// Search operations
	SearchServices(ctx context.Context, query string, categorySlug string, limit int) ([]*models.ServiceNew, error)
	SearchAddons(ctx context.Context, query string, categorySlug string, limit int) ([]*models.Addon, error)
}

// CategoryInfo holds category information with counts
type CategoryInfo struct {
	Slug         string
	ServiceCount int64
	AddonCount   int64
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new customer repository instance
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// ==================== Service Operations ====================

func (r *repository) GetActiveServiceBySlug(ctx context.Context, slug string) (*models.ServiceNew, error) {
	var service models.ServiceNew
	err := r.db.WithContext(ctx).
		Where("service_slug = ? AND is_active = true AND is_available = true", slug).
		First(&service).Error
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *repository) ListActiveServices(ctx context.Context, query dto.ListServicesQuery) ([]*models.ServiceNew, int64, error) {
	var services []*models.ServiceNew
	var total int64

	db := r.db.WithContext(ctx).Model(&models.ServiceNew{}).
		Where("is_active = true AND is_available = true")

	// Apply filters
	if query.CategorySlug != "" {
		db = db.Where("category_slug = ?", query.CategorySlug)
	}

	if query.Search != "" {
		searchPattern := "%" + strings.ToLower(query.Search) + "%"
		db = db.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)
	}

	if query.MinPrice != nil {
		db = db.Where("base_price >= ?", *query.MinPrice)
	}

	if query.MaxPrice != nil {
		db = db.Where("base_price <= ?", *query.MaxPrice)
	}

	if query.IsFrequent != nil {
		db = db.Where("is_frequent = ?", *query.IsFrequent)
	}

	// Count total before pagination
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderClause := query.SortBy
	if query.SortBy == "popularity" {
		// For now, use sort_order as popularity proxy
		// In future, could track views/bookings
		orderClause = "sort_order"
	}
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

func (r *repository) GetActiveServicesByCategory(ctx context.Context, categorySlug string) ([]*models.ServiceNew, error) {
	var services []*models.ServiceNew
	err := r.db.WithContext(ctx).
		Where("category_slug = ? AND is_active = true AND is_available = true", categorySlug).
		Order("sort_order ASC, title ASC").
		Find(&services).Error
	return services, err
}

func (r *repository) CountActiveServicesByCategory(ctx context.Context, categorySlug string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ServiceNew{}).
		Where("category_slug = ? AND is_active = true AND is_available = true", categorySlug).
		Count(&count).Error
	return count, err
}

func (r *repository) GetFrequentServices(ctx context.Context, limit int) ([]*models.ServiceNew, error) {
	var services []*models.ServiceNew
	err := r.db.WithContext(ctx).
		Where("is_frequent = true AND is_active = true AND is_available = true").
		Order("sort_order ASC").
		Limit(limit).
		Find(&services).Error
	return services, err
}

// ==================== Addon Operations ====================

func (r *repository) GetActiveAddonBySlug(ctx context.Context, slug string) (*models.Addon, error) {
	var addon models.Addon
	err := r.db.WithContext(ctx).
		Where("addon_slug = ? AND is_active = true AND is_available = true", slug).
		First(&addon).Error
	if err != nil {
		return nil, err
	}
	return &addon, nil
}

func (r *repository) ListActiveAddons(ctx context.Context, query dto.ListAddonsQuery) ([]*models.Addon, int64, error) {
	var addons []*models.Addon
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Addon{}).
		Where("is_active = true AND is_available = true")

	// Apply filters
	if query.CategorySlug != "" {
		db = db.Where("category_slug = ?", query.CategorySlug)
	}

	if query.Search != "" {
		searchPattern := "%" + strings.ToLower(query.Search) + "%"
		db = db.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)
	}

	if query.MinPrice != nil {
		db = db.Where("price >= ?", *query.MinPrice)
	}

	if query.MaxPrice != nil {
		db = db.Where("price <= ?", *query.MaxPrice)
	}

	if query.HasDiscount != nil && *query.HasDiscount {
		db = db.Where("strikethrough_price IS NOT NULL AND strikethrough_price > price")
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

func (r *repository) GetActiveAddonsByCategory(ctx context.Context, categorySlug string) ([]*models.Addon, error) {
	var addons []*models.Addon
	err := r.db.WithContext(ctx).
		Where("category_slug = ? AND is_active = true AND is_available = true", categorySlug).
		Order("sort_order ASC, title ASC").
		Find(&addons).Error
	return addons, err
}

func (r *repository) CountActiveAddonsByCategory(ctx context.Context, categorySlug string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Addon{}).
		Where("category_slug = ? AND is_active = true AND is_available = true", categorySlug).
		Count(&count).Error
	return count, err
}

func (r *repository) GetDiscountedAddons(ctx context.Context, limit int) ([]*models.Addon, error) {
	var addons []*models.Addon
	err := r.db.WithContext(ctx).
		Where("is_active = true AND is_available = true AND strikethrough_price IS NOT NULL AND strikethrough_price > price").
		Order("(strikethrough_price - price) / strikethrough_price DESC"). // Order by discount percentage
		Limit(limit).
		Find(&addons).Error
	return addons, err
}

// ==================== Category Operations ====================

func (r *repository) GetAllActiveCategories(ctx context.Context) ([]CategoryInfo, error) {
	// Get unique categories with their counts
	var categoryInfos []CategoryInfo

	// Get service counts by category
	type categoryCount struct {
		CategorySlug string
		Count        int64
	}
	var serviceCounts []categoryCount
	err := r.db.WithContext(ctx).
		Model(&models.ServiceNew{}).
		Select("category_slug, COUNT(*) as count").
		Where("is_active = true AND is_available = true").
		Group("category_slug").
		Find(&serviceCounts).Error
	if err != nil {
		return nil, err
	}

	// Get addon counts by category
	var addonCounts []categoryCount
	err = r.db.WithContext(ctx).
		Model(&models.Addon{}).
		Select("category_slug, COUNT(*) as count").
		Where("is_active = true AND is_available = true").
		Group("category_slug").
		Find(&addonCounts).Error
	if err != nil {
		return nil, err
	}

	// Merge counts
	categoryMap := make(map[string]*CategoryInfo)

	for _, sc := range serviceCounts {
		if _, exists := categoryMap[sc.CategorySlug]; !exists {
			categoryMap[sc.CategorySlug] = &CategoryInfo{Slug: sc.CategorySlug}
		}
		categoryMap[sc.CategorySlug].ServiceCount = sc.Count
	}

	for _, ac := range addonCounts {
		if _, exists := categoryMap[ac.CategorySlug]; !exists {
			categoryMap[ac.CategorySlug] = &CategoryInfo{Slug: ac.CategorySlug}
		}
		categoryMap[ac.CategorySlug].AddonCount = ac.Count
	}

	for _, info := range categoryMap {
		categoryInfos = append(categoryInfos, *info)
	}

	return categoryInfos, nil
}

// ==================== Search Operations ====================

func (r *repository) SearchServices(ctx context.Context, query string, categorySlug string, limit int) ([]*models.ServiceNew, error) {
	var services []*models.ServiceNew
	searchPattern := "%" + strings.ToLower(query) + "%"

	db := r.db.WithContext(ctx).
		Where("is_active = true AND is_available = true").
		Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(long_description) LIKE ?",
			searchPattern, searchPattern, searchPattern)

	if categorySlug != "" {
		db = db.Where("category_slug = ?", categorySlug)
	}

	err := db.Order("sort_order ASC").Limit(limit).Find(&services).Error
	return services, err
}

func (r *repository) SearchAddons(ctx context.Context, query string, categorySlug string, limit int) ([]*models.Addon, error) {
	var addons []*models.Addon
	searchPattern := "%" + strings.ToLower(query) + "%"

	db := r.db.WithContext(ctx).
		Where("is_active = true AND is_available = true").
		Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)

	if categorySlug != "" {
		db = db.Where("category_slug = ?", categorySlug)
	}

	err := db.Order("sort_order ASC").Limit(limit).Find(&addons).Error
	return addons, err
}
