package customer

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/customer/dto"
	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
)

type Repository interface {
	GetActiveServiceBySlug(ctx context.Context, slug string) (*models.ServiceNew, error)
	ListActiveServices(ctx context.Context, query dto.ListServicesQuery) ([]*models.ServiceNew, int64, error)
	GetActiveServicesByCategory(ctx context.Context, categorySlug string) ([]*models.ServiceNew, error)
	CountActiveServicesByCategory(ctx context.Context, categorySlug string) (int64, error)
	GetFrequentServices(ctx context.Context, limit int) ([]*models.ServiceNew, error)

	GetActiveAddonBySlug(ctx context.Context, slug string) (*models.Addon, error)
	ListActiveAddons(ctx context.Context, query dto.ListAddonsQuery) ([]*models.Addon, int64, error)
	GetActiveAddonsByCategory(ctx context.Context, categorySlug string) ([]*models.Addon, error)
	CountActiveAddonsByCategory(ctx context.Context, categorySlug string) (int64, error)
	GetDiscountedAddons(ctx context.Context, limit int) ([]*models.Addon, error)

	GetAllActiveCategories(ctx context.Context) ([]CategoryInfo, error)

	SearchServices(ctx context.Context, query string, categorySlug string, limit int) ([]*models.ServiceNew, error)
	SearchAddons(ctx context.Context, query string, categorySlug string, limit int) ([]*models.Addon, error)

	Create(ctx context.Context, order *models.ServiceOrderNew) error
	GetByID(ctx context.Context, id string) (*models.ServiceOrderNew, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*models.ServiceOrderNew, error)
	Update(ctx context.Context, order *models.ServiceOrderNew) error
	Delete(ctx context.Context, orderID string) error

	GetCustomerOrders(ctx context.Context, customerID string, query dto.ListOrdersQuery) ([]*models.ServiceOrderNew, int64, error)
	GetCustomerOrderByID(ctx context.Context, customerID, orderID string) (*models.ServiceOrderNew, error)
	CountCustomerActiveOrders(ctx context.Context, customerID string) (int64, error)

	UpdateStatus(ctx context.Context, orderID, status string) error

	CreateStatusHistory(ctx context.Context, history *models.OrderStatusHistory) error
	GetOrderStatusHistory(ctx context.Context, orderID string) ([]models.OrderStatusHistory, error)
}

type CategoryInfo struct {
	Slug         string
	ServiceCount int64
	AddonCount   int64
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}


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

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := query.SortBy
	if query.SortBy == "popularity" {
		orderClause = "sort_order"
	}
	if query.SortDesc {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}
	db = db.Order(orderClause)

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

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := query.SortBy
	if query.SortDesc {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}
	db = db.Order(orderClause)

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

func (r *repository) GetAllActiveCategories(ctx context.Context) ([]CategoryInfo, error) {
	var categoryInfos []CategoryInfo

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

func (r *repository) Create(ctx context.Context, order *models.ServiceOrderNew) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *repository) GetByID(ctx context.Context, id string) (*models.ServiceOrderNew, error) {
	var order models.ServiceOrderNew
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *repository) GetByOrderNumber(ctx context.Context, orderNumber string) (*models.ServiceOrderNew, error) {
	var order models.ServiceOrderNew
	err := r.db.WithContext(ctx).Where("order_number = ?", orderNumber).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *repository) Update(ctx context.Context, order *models.ServiceOrderNew) error {
	return r.db.WithContext(ctx).Model(&models.ServiceOrderNew{}).Where("id = ?", order.ID).Updates(order).Error
}

func (r *repository) GetCustomerOrders(ctx context.Context, customerID string, query dto.ListOrdersQuery) ([]*models.ServiceOrderNew, int64, error) {
	var orders []*models.ServiceOrderNew
	var total int64

	db := r.db.WithContext(ctx).Model(&models.ServiceOrderNew{}).
		Where("customer_id = ?", customerID)

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if query.FromDate != "" {
		fromDate, _ := time.Parse("2006-01-02", query.FromDate)
		db = db.Where("created_at >= ?", fromDate)
	}

	if query.ToDate != "" {
		toDate, _ := time.Parse("2006-01-02", query.ToDate)
		toDate = toDate.AddDate(0, 0, 1)
		db = db.Where("created_at < ?", toDate)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderClause := query.SortBy
	if query.SortBy == "booking_date" {
		orderClause = "booking_info->>'date'"
	}
	if query.SortDesc {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}
	db = db.Order(orderClause)

	offset := query.PaginationParams.GetOffset()
	if err := db.Offset(offset).Limit(query.Limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *repository) GetCustomerOrderByID(ctx context.Context, customerID, orderID string) (*models.ServiceOrderNew, error) {
	var order models.ServiceOrderNew
	err := r.db.WithContext(ctx).
		Where("id = ? AND customer_id = ?", orderID, customerID).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *repository) CountCustomerActiveOrders(ctx context.Context, customerID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("customer_id = ? AND status IN ?", customerID, shared.ActiveOrderStatuses()).
		Count(&count).Error
	return count, err
}

func (r *repository) UpdateStatus(ctx context.Context, orderID, status string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	switch status {
	case shared.OrderStatusCompleted:
		updates["completed_at"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("id = ?", orderID).
		Updates(updates).Error
}

func (r *repository) CreateStatusHistory(ctx context.Context, history *models.OrderStatusHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *repository) GetOrderStatusHistory(ctx context.Context, orderID string) ([]models.OrderStatusHistory, error) {
	var history []models.OrderStatusHistory
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at ASC").
		Find(&history).Error
	return history, err
}

func (r *repository) Delete(ctx context.Context, orderID string) error {
	return r.db.WithContext(ctx).Where("id = ?", orderID).Delete(&models.ServiceOrderNew{}).Error
}
