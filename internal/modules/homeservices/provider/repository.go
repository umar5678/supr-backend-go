package provider

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/provider/dto"
	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// Repository defines the interface for provider data access
type Repository interface {
	// Provider profile
	GetProvider(ctx context.Context, providerID string) (*models.ServiceProviderProfile, error)
	GetProviderByUserID(ctx context.Context, userID string) (*models.ServiceProviderProfile, error)

	// Service categories
	GetProviderCategories(ctx context.Context, providerID string) ([]*models.ProviderServiceCategory, error)
	GetProviderCategory(ctx context.Context, providerID, categorySlug string) (*models.ProviderServiceCategory, error)
	AddProviderCategory(ctx context.Context, category *models.ProviderServiceCategory) error
	UpdateProviderCategory(ctx context.Context, category *models.ProviderServiceCategory) error
	DeleteProviderCategory(ctx context.Context, providerID, categorySlug string) error
	GetProviderCategorySlugs(ctx context.Context, providerID string) ([]string, error)

	// Available orders
	GetAvailableOrders(ctx context.Context, categorySlugs []string, query dto.ListAvailableOrdersQuery) ([]*models.ServiceOrderNew, int64, error)
	GetAvailableOrderByID(ctx context.Context, orderID string, categorySlugs []string) (*models.ServiceOrderNew, error)

	// Provider orders
	GetProviderOrders(ctx context.Context, providerID string, query dto.ListMyOrdersQuery) ([]*models.ServiceOrderNew, int64, error)
	GetProviderOrderByID(ctx context.Context, providerID, orderID string) (*models.ServiceOrderNew, error)
	CountProviderActiveOrders(ctx context.Context, providerID string) (int64, error)

	// Order operations
	GetOrderByID(ctx context.Context, orderID string) (*models.ServiceOrderNew, error)
	UpdateOrder(ctx context.Context, order *models.ServiceOrderNew) error
	AssignOrderToProvider(ctx context.Context, orderID, providerID string) error

	// Statistics
	GetProviderStatistics(ctx context.Context, providerID string) (*ProviderStats, error)
	GetProviderEarnings(ctx context.Context, providerID string, fromDate, toDate time.Time) (*EarningsData, error)
	GetCategoryEarnings(ctx context.Context, providerID string, fromDate, toDate time.Time) ([]CategoryEarningsData, error)

	// Status history
	CreateStatusHistory(ctx context.Context, history *models.OrderStatusHistory) error
}

// ProviderStats holds provider statistics data
type ProviderStats struct {
	TotalCompletedJobs   int
	TotalEarnings        float64
	TotalRatings         int
	TotalRatingSum       int
	TotalAccepted        int
	TotalRejected        int
	AvgResponseMinutes   int
	ActiveOrders         int
	TodayCompletedOrders int
	TodayEarnings        float64
}

// EarningsData holds earnings data
type EarningsData struct {
	TotalEarnings  float64
	TotalOrders    int
	DailyBreakdown []DailyEarnings
}

// DailyEarnings holds daily earnings
type DailyEarnings struct {
	Date       string
	Earnings   float64
	OrderCount int
}

// CategoryEarningsData holds category earnings
type CategoryEarningsData struct {
	CategorySlug string
	Earnings     float64
	OrderCount   int
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new provider repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// ==================== Provider Profile ====================

func (r *repository) GetProvider(ctx context.Context, providerID string) (*models.ServiceProviderProfile, error) {
	var provider models.ServiceProviderProfile
	err := r.db.WithContext(ctx).
		Where("id = ?", providerID).
		Preload("User").
		First(&provider).Error
	return &provider, err
}

func (r *repository) GetProviderByUserID(ctx context.Context, userID string) (*models.ServiceProviderProfile, error) {
	var provider models.ServiceProviderProfile
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("User").
		First(&provider).Error
	return &provider, err
}

// ==================== Service Categories ====================

func (r *repository) GetProviderCategories(ctx context.Context, providerID string) ([]*models.ProviderServiceCategory, error) {
	var categories []*models.ProviderServiceCategory
	err := r.db.WithContext(ctx).
		Where("provider_id = ?", providerID).
		Order("created_at ASC").
		Find(&categories).Error
	return categories, err
}

func (r *repository) GetProviderCategory(ctx context.Context, providerID, categorySlug string) (*models.ProviderServiceCategory, error) {
	var category models.ProviderServiceCategory
	err := r.db.WithContext(ctx).
		Where("provider_id = ? AND category_slug = ?", providerID, categorySlug).
		First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *repository) AddProviderCategory(ctx context.Context, category *models.ProviderServiceCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *repository) UpdateProviderCategory(ctx context.Context, category *models.ProviderServiceCategory) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *repository) DeleteProviderCategory(ctx context.Context, providerID, categorySlug string) error {
	return r.db.WithContext(ctx).
		Where("provider_id = ? AND category_slug = ?", providerID, categorySlug).
		Delete(&models.ProviderServiceCategory{}).Error
}

func (r *repository) GetProviderCategorySlugs(ctx context.Context, providerID string) ([]string, error) {
	// Method 1: Try to get from provider_service_categories table
	var slugs []string
	err := r.db.WithContext(ctx).
		Model(&models.ProviderServiceCategory{}).
		Where("provider_id = ? AND is_active = true", providerID).
		Pluck("category_slug", &slugs).Error

	if err != nil {
		logger.Error("failed to query provider_service_categories", "error", err, "providerID", providerID)
		return nil, err
	}

	logger.Info("queried provider_service_categories", "providerID", providerID, "foundCategories", slugs, "count", len(slugs))

	// Method 2: If no categories found, derive from provider's qualified services
	if len(slugs) == 0 {
		logger.Info("no provider_service_categories found, deriving from qualified services", "providerID", providerID)

		// First, log what services the provider is qualified for
		var qualifiedServices []map[string]interface{}
		r.db.WithContext(ctx).
			Table("provider_qualified_services pqs").
			Joins("JOIN services s ON pqs.service_id = s.id").
			Where("pqs.provider_id = ?", providerID).
			Select("pqs.service_id, s.category_slug, s.title").
			Scan(&qualifiedServices)
		logger.Info("provider qualified services", "providerID", providerID, "qualifiedServices", qualifiedServices)

		err = r.db.WithContext(ctx).
			Table("provider_qualified_services pqs").
			Joins("JOIN services s ON pqs.service_id = s.id").
			Where("pqs.provider_id = ?", providerID).
			Distinct("s.category_slug").
			Pluck("s.category_slug", &slugs).Error

		if err != nil {
			logger.Error("failed to derive categories from services", "error", err, "providerID", providerID)
			return nil, err
		}

		logger.Info("derived provider categories from services", "providerID", providerID, "derivedCategories", slugs, "count", len(slugs))
	}

	return slugs, nil
}

// ==================== Available Orders ====================

func (r *repository) GetAvailableOrders(ctx context.Context, categorySlugs []string, query dto.ListAvailableOrdersQuery) ([]*models.ServiceOrderNew, int64, error) {
	var orders []*models.ServiceOrderNew
	var total int64

	db := r.db.WithContext(ctx).Model(&models.ServiceOrderNew{}).
		Where("status IN ?", []string{shared.OrderStatusPending, shared.OrderStatusSearchingProvider}).
		Where("category_slug IN ?", categorySlugs).
		Where("assigned_provider_id IS NULL")

	logger.Info("GetAvailableOrders query filters", "categorySlugs", categorySlugs, "statusFilters", []string{shared.OrderStatusPending, shared.OrderStatusSearchingProvider})

	// Filter by specific category if provided
	if query.CategorySlug != "" {
		db = db.Where("category_slug = ?", query.CategorySlug)
	}

	// Filter by booking date
	if query.Date != "" {
		db = db.Where("booking_info->>'date' = ?", query.Date)
	}

	// Exclude expired orders
	db = db.Where("expires_at IS NULL OR expires_at > ?", time.Now())

	// Count total
	if err := db.Count(&total).Error; err != nil {
		logger.Error("failed to count available orders", "error", err)
		return nil, 0, err
	}

	logger.Info("counted available orders", "total", total, "categorySlugs", categorySlugs)

	// Debug: if total is 0, let's log what orders exist in the table
	if total == 0 {
		// Log all orders in DB with status and category
		var allOrders []map[string]interface{}
		r.db.WithContext(ctx).Model(&models.ServiceOrderNew{}).
			Select("id, category_slug, status, assigned_provider_id").
			Limit(20).
			Scan(&allOrders)
		logger.Warn("no matching orders found; all orders in DB", "allOrders", allOrders)

		// Log all available orders (no category filter)
		var availableOrders []map[string]interface{}
		r.db.WithContext(ctx).Model(&models.ServiceOrderNew{}).
			Where("status IN ?", []string{shared.OrderStatusPending, shared.OrderStatusSearchingProvider}).
			Where("assigned_provider_id IS NULL").
			Select("id, category_slug, status").
			Limit(20).
			Scan(&availableOrders)
		logger.Warn("available orders without category filter", "availableOrders", availableOrders)
	}

	// Sorting
	orderClause := query.SortBy
	if query.SortBy == "booking_date" {
		orderClause = "booking_info->>'date'"
	} else if query.SortBy == "price" {
		orderClause = "total_price"
	}
	if query.SortDesc {
		orderClause += " DESC"
	} else {
		orderClause += " ASC"
	}
	db = db.Order(orderClause)

	// Pagination
	offset := query.PaginationParams.GetOffset()
	if err := db.Offset(offset).Limit(query.Limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *repository) GetAvailableOrderByID(ctx context.Context, orderID string, categorySlugs []string) (*models.ServiceOrderNew, error) {
	var order models.ServiceOrderNew
	err := r.db.WithContext(ctx).
		Where("id = ?", orderID).
		Where("status IN ?", []string{shared.OrderStatusPending, shared.OrderStatusSearchingProvider}).
		Where("category_slug IN ?", categorySlugs).
		Where("assigned_provider_id IS NULL").
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// ==================== Provider Orders ====================

func (r *repository) GetProviderOrders(ctx context.Context, providerID string, query dto.ListMyOrdersQuery) ([]*models.ServiceOrderNew, int64, error) {
	var orders []*models.ServiceOrderNew
	var total int64

	db := r.db.WithContext(ctx).Model(&models.ServiceOrderNew{}).
		Where("assigned_provider_id = ?", providerID)

	// Filter by status
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	// Filter by date range
	if query.FromDate != "" {
		fromDate, _ := time.Parse("2006-01-02", query.FromDate)
		db = db.Where("created_at >= ?", fromDate)
	}
	if query.ToDate != "" {
		toDate, _ := time.Parse("2006-01-02", query.ToDate)
		toDate = toDate.AddDate(0, 0, 1)
		db = db.Where("created_at < ?", toDate)
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting
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

	// Pagination
	offset := query.PaginationParams.GetOffset()
	if err := db.Offset(offset).Limit(query.Limit).Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *repository) GetProviderOrderByID(ctx context.Context, providerID, orderID string) (*models.ServiceOrderNew, error) {
	var order models.ServiceOrderNew
	err := r.db.WithContext(ctx).
		Where("id = ? AND assigned_provider_id = ?", orderID, providerID).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *repository) CountProviderActiveOrders(ctx context.Context, providerID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("assigned_provider_id = ?", providerID).
		Where("status IN ?", shared.ActiveOrderStatuses()).
		Count(&count).Error
	return count, err
}

// ==================== Order Operations ====================

func (r *repository) GetOrderByID(ctx context.Context, orderID string) (*models.ServiceOrderNew, error) {
	var order models.ServiceOrderNew
	err := r.db.WithContext(ctx).Where("id = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *repository) UpdateOrder(ctx context.Context, order *models.ServiceOrderNew) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *repository) AssignOrderToProvider(ctx context.Context, orderID, providerID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"assigned_provider_id": providerID,
			"status":               shared.OrderStatusAssigned,
			"updated_at":           now,
		}).Error
}

// ==================== Statistics ====================

func (r *repository) GetProviderStatistics(ctx context.Context, providerID string) (*ProviderStats, error) {
	stats := &ProviderStats{}

	// Get completed jobs and earnings
	err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("assigned_provider_id = ? AND status = ?", providerID, shared.OrderStatusCompleted).
		Select("COUNT(*) as total_completed_jobs, COALESCE(SUM(total_price * 0.9), 0) as total_earnings").
		Row().Scan(&stats.TotalCompletedJobs, &stats.TotalEarnings)
	if err != nil {
		return nil, err
	}

	// Get ratings
	err = r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("assigned_provider_id = ? AND customer_rating IS NOT NULL", providerID).
		Select("COUNT(*) as total_ratings, COALESCE(SUM(customer_rating), 0) as total_rating_sum").
		Row().Scan(&stats.TotalRatings, &stats.TotalRatingSum)
	if err != nil {
		return nil, err
	}

	// Get active orders count
	activeCount, err := r.CountProviderActiveOrders(ctx, providerID)
	if err != nil {
		return nil, err
	}
	stats.ActiveOrders = int(activeCount)

	// Get today's stats
	today := time.Now().Truncate(24 * time.Hour)
	err = r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("assigned_provider_id = ? AND status = ? AND completed_at >= ?",
			providerID, shared.OrderStatusCompleted, today).
		Select("COUNT(*) as today_completed_orders, COALESCE(SUM(total_price * 0.9), 0) as today_earnings").
		Row().Scan(&stats.TodayCompletedOrders, &stats.TodayEarnings)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (r *repository) GetProviderEarnings(ctx context.Context, providerID string, fromDate, toDate time.Time) (*EarningsData, error) {
	earnings := &EarningsData{}

	// Get totals
	err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("assigned_provider_id = ? AND status = ?", providerID, shared.OrderStatusCompleted).
		Where("completed_at >= ? AND completed_at < ?", fromDate, toDate.AddDate(0, 0, 1)).
		Select("COALESCE(SUM(total_price * 0.9), 0) as total_earnings, COUNT(*) as total_orders").
		Row().Scan(&earnings.TotalEarnings, &earnings.TotalOrders)
	if err != nil {
		return nil, err
	}

	// Get daily breakdown
	rows, err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("assigned_provider_id = ? AND status = ?", providerID, shared.OrderStatusCompleted).
		Where("completed_at >= ? AND completed_at < ?", fromDate, toDate.AddDate(0, 0, 1)).
		Select("DATE(completed_at) as date, COALESCE(SUM(total_price * 0.9), 0) as earnings, COUNT(*) as order_count").
		Group("DATE(completed_at)").
		Order("date ASC").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var daily DailyEarnings
		if err := rows.Scan(&daily.Date, &daily.Earnings, &daily.OrderCount); err != nil {
			return nil, err
		}
		earnings.DailyBreakdown = append(earnings.DailyBreakdown, daily)
	}

	return earnings, nil
}

func (r *repository) GetCategoryEarnings(ctx context.Context, providerID string, fromDate, toDate time.Time) ([]CategoryEarningsData, error) {
	var categoryEarnings []CategoryEarningsData

	err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("assigned_provider_id = ? AND status = ?", providerID, shared.OrderStatusCompleted).
		Where("completed_at >= ? AND completed_at < ?", fromDate, toDate.AddDate(0, 0, 1)).
		Select("category_slug, COALESCE(SUM(total_price * 0.9), 0) as earnings, COUNT(*) as order_count").
		Group("category_slug").
		Order("earnings DESC").
		Find(&categoryEarnings).Error

	return categoryEarnings, err
}

func (r *repository) CreateStatusHistory(ctx context.Context, history *models.OrderStatusHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}
