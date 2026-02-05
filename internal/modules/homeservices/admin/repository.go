package admin

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/admin/dto"
	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
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

	// Order queries
	GetOrders(ctx context.Context, query dto.ListOrdersQuery) ([]*models.ServiceOrderNew, int64, error)
	GetOrderByID(ctx context.Context, id string) (*models.ServiceOrderNew, error)
	GetOrderByNumber(ctx context.Context, orderNumber string) (*models.ServiceOrderNew, error)
	UpdateOrder(ctx context.Context, order *models.ServiceOrderNew) error

	// Status history
	GetOrderStatusHistory(ctx context.Context, orderID string) ([]models.OrderStatusHistory, error)
	CreateStatusHistory(ctx context.Context, history *models.OrderStatusHistory) error

	// Analytics
	GetOrderStats(ctx context.Context, fromDate, toDate time.Time) (*OrderStats, error)
	GetOrdersByStatus(ctx context.Context, fromDate, toDate time.Time) ([]StatusStats, error)
	GetOrdersByCategory(ctx context.Context, fromDate, toDate time.Time) ([]CategoryStats, error)
	GetRevenueBreakdown(ctx context.Context, fromDate, toDate time.Time, groupBy string) ([]RevenueStats, error)
	GetProviderAnalytics(ctx context.Context, fromDate, toDate time.Time, query dto.ProviderAnalyticsQuery) ([]ProviderStats, error)
	GetPaymentMethodStats(ctx context.Context, fromDate, toDate time.Time) ([]PaymentStats, error)

	// Dashboard
	GetTodayStats(ctx context.Context) (*TodayStatsData, error)
	GetWeeklyStats(ctx context.Context) (*WeeklyStatsData, error)
	GetPendingActions(ctx context.Context) (*PendingActionsData, error)
	GetRecentOrders(ctx context.Context, limit int) ([]*models.ServiceOrderNew, error)

	// Bulk operations
	BulkUpdateStatus(ctx context.Context, orderIDs []string, status string, changedBy string, reason string) (int64, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new admin repository instance
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Stats structures
type OrderStats struct {
	TotalOrders          int64
	CompletedOrders      int64
	CancelledOrders      int64
	PendingOrders        int64
	TotalRevenue         float64
	TotalCommission      float64
	TotalProviderPayouts float64
	TotalRatings         int64
	TotalRatingSum       int64
}

type StatusStats struct {
	Status string
	Count  int64
}

type CategoryStats struct {
	CategorySlug string
	OrderCount   int64
	Revenue      float64
}

type RevenueStats struct {
	Period          string
	OrderCount      int64
	Revenue         float64
	Commission      float64
	ProviderPayouts float64
}

type ProviderStats struct {
	ProviderID      string
	CompletedOrders int64
	CancelledOrders int64
	TotalEarnings   float64
	TotalRatings    int64
	TotalRatingSum  int64
}

type PaymentStats struct {
	Method      string
	OrderCount  int64
	TotalAmount float64
}

type TodayStatsData struct {
	TotalOrders      int64
	CompletedOrders  int64
	PendingOrders    int64
	InProgressOrders int64
	Revenue          float64
	Commission       float64
}

type WeeklyStatsData struct {
	TotalOrders     int64
	CompletedOrders int64
	TotalRevenue    float64
	TotalRatings    int64
	TotalRatingSum  int64
	DailyStats      []DailyStatsData
}

type DailyStatsData struct {
	Date       string
	OrderCount int64
	Revenue    float64
}

type PendingActionsData struct {
	OrdersNeedingProvider int64
	ExpiredOrders         int64
	DisputedOrders        int64
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

// ==================== Order Queries ====================

func (r *repository) GetOrders(ctx context.Context, query dto.ListOrdersQuery) ([]*models.ServiceOrderNew, int64, error) {
	var orders []*models.ServiceOrderNew
	var total int64

	db := r.db.WithContext(ctx).Model(&models.ServiceOrderNew{})

	// Apply filters
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.CategorySlug != "" {
		db = db.Where("category_slug = ?", query.CategorySlug)
	}
	if query.CustomerID != "" {
		db = db.Where("customer_id = ?", query.CustomerID)
	}
	if query.ProviderID != "" {
		db = db.Where("assigned_provider_id = ?", query.ProviderID)
	}
	if query.OrderNumber != "" {
		db = db.Where("order_number ILIKE ?", "%"+query.OrderNumber+"%")
	}

	// Date filters (created_at)
	if query.FromDate != "" {
		fromDate, _ := time.Parse("2006-01-02", query.FromDate)
		db = db.Where("created_at >= ?", fromDate)
	}
	if query.ToDate != "" {
		toDate, _ := time.Parse("2006-01-02", query.ToDate)
		toDate = toDate.AddDate(0, 0, 1)
		db = db.Where("created_at < ?", toDate)
	}

	// Booking date filters
	if query.BookingFromDate != "" {
		db = db.Where("booking_info->>'date' >= ?", query.BookingFromDate)
	}
	if query.BookingToDate != "" {
		db = db.Where("booking_info->>'date' <= ?", query.BookingToDate)
	}

	// Price filters
	if query.MinPrice != nil {
		db = db.Where("total_price >= ?", *query.MinPrice)
	}
	if query.MaxPrice != nil {
		db = db.Where("total_price <= ?", *query.MaxPrice)
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
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

	// Apply pagination
	offset := query.PaginationParams.GetOffset()
	if err := db.Offset(offset).Limit(query.Limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *repository) GetOrderByID(ctx context.Context, id string) (*models.ServiceOrderNew, error) {
	var order models.ServiceOrderNew
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *repository) GetOrderByNumber(ctx context.Context, orderNumber string) (*models.ServiceOrderNew, error) {
	var order models.ServiceOrderNew
	err := r.db.WithContext(ctx).Where("order_number = ?", orderNumber).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *repository) UpdateOrder(ctx context.Context, order *models.ServiceOrderNew) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// ==================== Status History ====================

func (r *repository) GetOrderStatusHistory(ctx context.Context, orderID string) ([]models.OrderStatusHistory, error) {
	var history []models.OrderStatusHistory
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at ASC").
		Find(&history).Error
	return history, err
}

func (r *repository) CreateStatusHistory(ctx context.Context, history *models.OrderStatusHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// ==================== Analytics ====================

func (r *repository) GetOrderStats(ctx context.Context, fromDate, toDate time.Time) (*OrderStats, error) {
	stats := &OrderStats{}
	toDateEnd := toDate.AddDate(0, 0, 1)

	// Total orders
	if err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("created_at >= ? AND created_at < ?", fromDate, toDateEnd).
		Count(&stats.TotalOrders).Error; err != nil {
		return nil, err
	}

	// Completed orders and revenue
	row := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("created_at >= ? AND created_at < ?", fromDate, toDateEnd).
		Where("status = ?", shared.OrderStatusCompleted).
		Select("COUNT(*), COALESCE(SUM(total_price), 0), COALESCE(SUM(platform_commission), 0)").
		Row()
	if err := row.Scan(&stats.CompletedOrders, &stats.TotalRevenue, &stats.TotalCommission); err != nil {
		return nil, err
	}

	stats.TotalProviderPayouts = stats.TotalRevenue - stats.TotalCommission

	// Cancelled orders
	if err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("created_at >= ? AND created_at < ?", fromDate, toDateEnd).
		Where("status = ?", shared.OrderStatusCancelled).
		Count(&stats.CancelledOrders).Error; err != nil {
		return nil, err
	}

	// Pending orders
	if err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("created_at >= ? AND created_at < ?", fromDate, toDateEnd).
		Where("status IN ?", []string{shared.OrderStatusPending, shared.OrderStatusSearchingProvider}).
		Count(&stats.PendingOrders).Error; err != nil {
		return nil, err
	}

	// Ratings
	row = r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("created_at >= ? AND created_at < ?", fromDate, toDateEnd).
		Where("customer_rating IS NOT NULL").
		Select("COUNT(*), COALESCE(SUM(customer_rating), 0)").
		Row()
	row.Scan(&stats.TotalRatings, &stats.TotalRatingSum)

	return stats, nil
}

func (r *repository) GetOrdersByStatus(ctx context.Context, fromDate, toDate time.Time) ([]StatusStats, error) {
	var stats []StatusStats
	toDateEnd := toDate.AddDate(0, 0, 1)

	err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("created_at >= ? AND created_at < ?", fromDate, toDateEnd).
		Select("status, COUNT(*) as count").
		Group("status").
		Order("count DESC").
		Find(&stats).Error

	return stats, err
}

func (r *repository) GetOrdersByCategory(ctx context.Context, fromDate, toDate time.Time) ([]CategoryStats, error) {
	var stats []CategoryStats
	toDateEnd := toDate.AddDate(0, 0, 1)

	err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("created_at >= ? AND created_at < ?", fromDate, toDateEnd).
		Where("status = ?", shared.OrderStatusCompleted).
		Select("category_slug, COUNT(*) as order_count, COALESCE(SUM(total_price), 0) as revenue").
		Group("category_slug").
		Order("revenue DESC").
		Find(&stats).Error

	return stats, err
}

func (r *repository) GetRevenueBreakdown(ctx context.Context, fromDate, toDate time.Time, groupBy string) ([]RevenueStats, error) {
	var stats []RevenueStats
	toDateEnd := toDate.AddDate(0, 0, 1)

	var dateFormat string
	switch groupBy {
	case "week":
		dateFormat = "YYYY-IW" // ISO week
	case "month":
		dateFormat = "YYYY-MM"
	default:
		dateFormat = "YYYY-MM-DD"
	}

	err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("created_at >= ? AND created_at < ?", fromDate, toDateEnd).
		Where("status = ?", shared.OrderStatusCompleted).
		Select("TO_CHAR(completed_at, ?) as period, COUNT(*) as order_count, COALESCE(SUM(total_price), 0) as revenue, COALESCE(SUM(platform_commission), 0) as commission, COALESCE(SUM(total_price - platform_commission), 0) as provider_payouts", dateFormat).
		Group("period").
		Order("period ASC").
		Find(&stats).Error

	return stats, err
}

func (r *repository) GetProviderAnalytics(ctx context.Context, fromDate, toDate time.Time, query dto.ProviderAnalyticsQuery) ([]ProviderStats, error) {
	var stats []ProviderStats
	toDateEnd := toDate.AddDate(0, 0, 1)

	db := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("created_at >= ? AND created_at < ?", fromDate, toDateEnd).
		Where("assigned_provider_id IS NOT NULL")

	if query.CategorySlug != "" {
		db = db.Where("category_slug = ?", query.CategorySlug)
	}

	db = db.Select(`
		assigned_provider_id as provider_id,
		SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as completed_orders,
		SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) as cancelled_orders,
		COALESCE(SUM(CASE WHEN status = ? THEN total_price * 0.9 ELSE 0 END), 0) as total_earnings,
		SUM(CASE WHEN customer_rating IS NOT NULL THEN 1 ELSE 0 END) as total_ratings,
		COALESCE(SUM(customer_rating), 0) as total_rating_sum
	`, shared.OrderStatusCompleted, shared.OrderStatusCancelled, shared.OrderStatusCompleted).
		Group("assigned_provider_id")

	// Apply filters
	if query.MinOrders != nil {
		db = db.Having("SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) >= ?", shared.OrderStatusCompleted, *query.MinOrders)
	}

	// Sorting
	switch query.SortBy {
	case "earnings":
		if query.SortDesc {
			db = db.Order("total_earnings DESC")
		} else {
			db = db.Order("total_earnings ASC")
		}
	case "rating":
		if query.SortDesc {
			db = db.Order("total_rating_sum / NULLIF(total_ratings, 0) DESC NULLS LAST")
		} else {
			db = db.Order("total_rating_sum / NULLIF(total_ratings, 0) ASC NULLS LAST")
		}
	default: // completed_orders
		if query.SortDesc {
			db = db.Order("completed_orders DESC")
		} else {
			db = db.Order("completed_orders ASC")
		}
	}

	db = db.Limit(query.Limit)

	err := db.Find(&stats).Error
	return stats, err
}

func (r *repository) GetPaymentMethodStats(ctx context.Context, fromDate, toDate time.Time) ([]PaymentStats, error) {
	var stats []PaymentStats
	toDateEnd := toDate.AddDate(0, 0, 1)

	err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("created_at >= ? AND created_at < ?", fromDate, toDateEnd).
		Where("status = ?", shared.OrderStatusCompleted).
		Select("payment_info->>'method' as method, COUNT(*) as order_count, COALESCE(SUM(total_price), 0) as total_amount").
		Group("payment_info->>'method'").
		Order("total_amount DESC").
		Find(&stats).Error

	return stats, err
}

// ==================== Dashboard ====================

func (r *repository) GetTodayStats(ctx context.Context) (*TodayStatsData, error) {
	stats := &TodayStatsData{}
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.AddDate(0, 0, 1)

	// Total orders today
	r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("created_at >= ? AND created_at < ?", today, tomorrow).
		Count(&stats.TotalOrders)

	// Completed today
	row := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("completed_at >= ? AND completed_at < ?", today, tomorrow).
		Where("status = ?", shared.OrderStatusCompleted).
		Select("COUNT(*), COALESCE(SUM(total_price), 0), COALESCE(SUM(platform_commission), 0)").
		Row()
	row.Scan(&stats.CompletedOrders, &stats.Revenue, &stats.Commission)

	// Pending
	r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("status IN ?", []string{shared.OrderStatusPending, shared.OrderStatusSearchingProvider, shared.OrderStatusAssigned, shared.OrderStatusAccepted}).
		Count(&stats.PendingOrders)

	// In progress
	r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("status = ?", shared.OrderStatusInProgress).
		Count(&stats.InProgressOrders)

	return stats, nil
}

func (r *repository) GetWeeklyStats(ctx context.Context) (*WeeklyStatsData, error) {
	stats := &WeeklyStatsData{}
	today := time.Now().Truncate(24 * time.Hour)
	weekAgo := today.AddDate(0, 0, -7)
	tomorrow := today.AddDate(0, 0, 1)

	// Weekly totals
	row := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("created_at >= ? AND created_at < ?", weekAgo, tomorrow).
		Where("status = ?", shared.OrderStatusCompleted).
		Select("COUNT(*), COALESCE(SUM(total_price), 0)").
		Row()
	row.Scan(&stats.CompletedOrders, &stats.TotalRevenue)

	r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("created_at >= ? AND created_at < ?", weekAgo, tomorrow).
		Count(&stats.TotalOrders)

	// Ratings
	row = r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("created_at >= ? AND created_at < ?", weekAgo, tomorrow).
		Where("customer_rating IS NOT NULL").
		Select("COUNT(*), COALESCE(SUM(customer_rating), 0)").
		Row()
	row.Scan(&stats.TotalRatings, &stats.TotalRatingSum)

	// Daily breakdown
	rows, err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("created_at >= ? AND created_at < ?", weekAgo, tomorrow).
		Where("status = ?", shared.OrderStatusCompleted).
		Select("DATE(completed_at) as date, COUNT(*) as order_count, COALESCE(SUM(total_price), 0) as revenue").
		Group("DATE(completed_at)").
		Order("date ASC").
		Rows()
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	for rows.Next() {
		var daily DailyStatsData
		rows.Scan(&daily.Date, &daily.OrderCount, &daily.Revenue)
		stats.DailyStats = append(stats.DailyStats, daily)
	}

	return stats, nil
}

func (r *repository) GetPendingActions(ctx context.Context) (*PendingActionsData, error) {
	data := &PendingActionsData{}

	// Orders needing provider
	r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("status IN ?", []string{shared.OrderStatusPending, shared.OrderStatusSearchingProvider}).
		Where("assigned_provider_id IS NULL").
		Count(&data.OrdersNeedingProvider)

	// Expired orders
	r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("status IN ?", []string{shared.OrderStatusPending, shared.OrderStatusSearchingProvider}).
		Where("expires_at < ?", time.Now()).
		Count(&data.ExpiredOrders)

	return data, nil
}

func (r *repository) GetRecentOrders(ctx context.Context, limit int) ([]*models.ServiceOrderNew, error) {
	var orders []*models.ServiceOrderNew
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Find(&orders).Error
	return orders, err
}

// ==================== Bulk Operations ====================

func (r *repository) BulkUpdateStatus(ctx context.Context, orderIDs []string, status string, changedBy string, reason string) (int64, error) {
	result := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("id IN ?", orderIDs).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return 0, result.Error
	}

	// Create status history for each order
	for _, orderID := range orderIDs {
		history := models.NewOrderStatusHistory(
			orderID,
			"", // Unknown previous status in bulk
			status,
			&changedBy,
			shared.RoleAdmin,
			reason,
			nil,
		)
		r.CreateStatusHistory(ctx, history)
	}

	return result.RowsAffected, nil
}
