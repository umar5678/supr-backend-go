package provider

import (
	"context"
	"sort"
	"time"

	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/provider/dto"
	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

type Repository interface {

	GetProvider(ctx context.Context, providerID string) (*models.ServiceProviderProfile, error)
	GetProviderByUserID(ctx context.Context, userID string) (*models.ServiceProviderProfile, error)
	CreateProvider(ctx context.Context, provider *models.ServiceProviderProfile) error

	GetProviderCategories(ctx context.Context, providerID string) ([]*models.ProviderServiceCategory, error)
	GetProviderCategory(ctx context.Context, providerID, categorySlug string) (*models.ProviderServiceCategory, error)
	AddProviderCategory(ctx context.Context, category *models.ProviderServiceCategory) error
	UpdateProviderCategory(ctx context.Context, category *models.ProviderServiceCategory) error
	DeleteProviderCategory(ctx context.Context, providerID, categorySlug string) error
	GetProviderCategorySlugs(ctx context.Context, providerID string) ([]string, error)

	GetAvailableOrders(ctx context.Context, categorySlugs []string, query dto.ListAvailableOrdersQuery) ([]*models.ServiceOrderNew, int64, error)
	GetAvailableOrderByID(ctx context.Context, orderID string, categorySlugs []string) (*models.ServiceOrderNew, error)

	GetProviderOrders(ctx context.Context, providerID string, query dto.ListMyOrdersQuery) ([]*models.ServiceOrderNew, int64, error)
	GetProviderOrderByID(ctx context.Context, providerID, orderID string) (*models.ServiceOrderNew, error)
	CountProviderActiveOrders(ctx context.Context, providerID string) (int64, error)

	GetOrderByID(ctx context.Context, orderID string) (*models.ServiceOrderNew, error)
	UpdateOrder(ctx context.Context, order *models.ServiceOrderNew) error
	AssignOrderToProvider(ctx context.Context, orderID, providerID string) error

	GetProviderStatistics(ctx context.Context, providerID string) (*ProviderStats, error)
	GetProviderEarnings(ctx context.Context, providerID string, fromDate, toDate time.Time) (*EarningsData, error)
	GetCategoryEarnings(ctx context.Context, providerID string, fromDate, toDate time.Time) ([]CategoryEarningsData, error)

	CreateStatusHistory(ctx context.Context, history *models.OrderStatusHistory) error
}

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

type EarningsData struct {
	TotalEarnings  float64
	TotalOrders    int
	DailyBreakdown []DailyEarnings
}

type DailyEarnings struct {
	Date       string
	Earnings   float64
	OrderCount int
}

type CategoryEarningsData struct {
	CategorySlug string
	Earnings     float64
	OrderCount   int
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) AssignOrderToProvider(ctx context.Context, orderID, providerID string) error {
	now := time.Now()

	result := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"assigned_provider_id": providerID,
			"status":               shared.OrderStatusAssigned,
			"updated_at":           now,
		})

	if result.Error == nil && result.RowsAffected > 0 {
		logger.Info("assigned service order to provider", "orderID", orderID, "providerID", providerID)
		return nil
	}

	result = r.db.WithContext(ctx).
		Model(&models.LaundryOrder{}).
		Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"provider_id": providerID,
			"status":      shared.OrderStatusAssigned,
			"updated_at":  now,
		})

	if result.Error != nil {
		logger.Error("failed to assign order to provider", "error", result.Error, "orderID", orderID, "providerID", providerID)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Warn("order not found in either table", "orderID", orderID)
		return gorm.ErrRecordNotFound
	}

	logger.Info("assigned laundry order to provider", "orderID", orderID, "providerID", providerID)
	return nil
}

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

func (r *repository) CreateProvider(ctx context.Context, provider *models.ServiceProviderProfile) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

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

	if len(slugs) == 0 {
		logger.Info("no provider_service_categories found, deriving from qualified services", "providerID", providerID)

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

func (r *repository) convertLaundryOrderToServiceOrder(ctx context.Context, orderID string) (*models.ServiceOrderNew, error) {
	var laundryOrder models.LaundryOrder
	err := r.db.WithContext(ctx).Where("id = ?", orderID).First(&laundryOrder).Error
	if err != nil {
		return nil, err
	}

	var customer models.User
	customerName := ""
	if laundryOrder.UserID != nil {
		if err := r.db.WithContext(ctx).Where("id = ?", *laundryOrder.UserID).First(&customer).Error; err == nil {
			customerName = customer.Name
		}
	}

	var items []*models.LaundryOrderItem
	r.db.WithContext(ctx).Where("order_id = ?", laundryOrder.ID).Find(&items)

	selectedServices := make(models.SelectedServices, 0)
	for _, item := range items {
		selectedServices = append(selectedServices, models.SelectedServiceItem{
			ServiceSlug: item.ServiceSlug,
			Title:       item.ItemType,
			Price:       item.Price,
			Quantity:    item.Quantity,
		})
	}

	bookingDate := ""
	if laundryOrder.ServiceDate != nil {
		bookingDate = laundryOrder.ServiceDate.Format("2006-01-02")
	}

	customerID := ""
	if laundryOrder.UserID != nil {
		customerID = *laundryOrder.UserID
	}

	totalPrice := laundryOrder.Total
	if laundryOrder.Tip != nil && *laundryOrder.Tip > 0 {
		totalPrice = laundryOrder.Total + *laundryOrder.Tip
	}

	order := &models.ServiceOrderNew{
		ID:                 laundryOrder.ID,
		OrderNumber:        laundryOrder.OrderNumber,
		CustomerID:         customerID,
		CategorySlug:       laundryOrder.CategorySlug,
		ServicesTotal:      laundryOrder.Total, 
		Subtotal:           laundryOrder.Total, 
		TotalPrice:         totalPrice,         
		PlatformCommission: 0,                  
		AddonsTotal:        0,                  
		Status:             laundryOrder.Status,
		AssignedProviderID: laundryOrder.ProviderID,
		CreatedAt:          laundryOrder.CreatedAt,
		UpdatedAt:          laundryOrder.UpdatedAt,
		CustomerInfo: models.CustomerInfo{
			Name:    customerName,
			Address: laundryOrder.Address,
			Lat:     laundryOrder.Latitude,
			Lng:     laundryOrder.Longitude,
		},
		BookingInfo: models.BookingInfo{
			Date: bookingDate,
			Time: "",
		},
		SelectedServices: selectedServices,
	}

	return order, nil
}

func (r *repository) GetAvailableOrders(ctx context.Context, categorySlugs []string, query dto.ListAvailableOrdersQuery) ([]*models.ServiceOrderNew, int64, error) {
	var allOrders []*models.ServiceOrderNew
	var total int64

	var serviceOrders []*models.ServiceOrderNew
	db := r.db.WithContext(ctx).Model(&models.ServiceOrderNew{}).
		Where("status IN ?", []string{shared.OrderStatusPending, shared.OrderStatusSearchingProvider}).
		Where("category_slug IN ?", categorySlugs).
		Where("assigned_provider_id IS NULL").
		Where("expires_at IS NULL OR expires_at > ?", time.Now())

	logger.Info("GetAvailableOrders query filters", "categorySlugs", categorySlugs, "statusFilters", []string{shared.OrderStatusPending, shared.OrderStatusSearchingProvider})

	if query.CategorySlug != "" {
		db = db.Where("category_slug = ?", query.CategorySlug)
	}

	if query.Date != "" {
		db = db.Where("booking_info->>'date' = ?", query.Date)
	}

	if err := db.Order("created_at DESC").Find(&serviceOrders).Error; err != nil {
		logger.Error("failed to fetch service orders", "error", err)
	}

	// Query LaundryOrder
	var laundryOrders []*models.LaundryOrder
	laundryDb := r.db.WithContext(ctx).Model(&models.LaundryOrder{}).
		Where("status IN ?", []string{"pending", "searching_provider"}).
		Where("category_slug IN ?", categorySlugs).
		Where("provider_id IS NULL")

	if query.CategorySlug != "" {
		laundryDb = laundryDb.Where("category_slug = ?", query.CategorySlug)
	}

	if err := laundryDb.Order("created_at DESC").Find(&laundryOrders).Error; err != nil {
		logger.Error("failed to fetch laundry orders", "error", err)
	}

	logger.Info("fetched orders from both tables", "serviceOrders", len(serviceOrders), "laundryOrders", len(laundryOrders))

	for _, laundryOrder := range laundryOrders {

		var customer models.User
		customerName := ""
		if laundryOrder.UserID != nil {
			if err := r.db.WithContext(ctx).Where("id = ?", *laundryOrder.UserID).First(&customer).Error; err == nil {
				customerName = customer.Name
			}
		}

		var items []*models.LaundryOrderItem
		r.db.WithContext(ctx).Where("order_id = ?", laundryOrder.ID).Find(&items)

		selectedServices := make(models.SelectedServices, 0)
		for _, item := range items {
			selectedServices = append(selectedServices, models.SelectedServiceItem{
				ServiceSlug: item.ServiceSlug,
				Title:       item.ItemType,
				Price:       item.Price,
				Quantity:    item.Quantity,
			})
		}

		bookingDate := ""
		if laundryOrder.ServiceDate != nil {
			bookingDate = laundryOrder.ServiceDate.Format("2006-01-02")
		}

		customerID := ""
		if laundryOrder.UserID != nil {
			customerID = *laundryOrder.UserID
		}

		totalPrice := laundryOrder.Total
		if laundryOrder.Tip != nil && *laundryOrder.Tip > 0 {
			totalPrice = laundryOrder.Total + *laundryOrder.Tip
		}

		serviceOrder := &models.ServiceOrderNew{
			ID:                 laundryOrder.ID,
			OrderNumber:        laundryOrder.OrderNumber,
			CustomerID:         customerID,
			CategorySlug:       laundryOrder.CategorySlug,
			ServicesTotal:      laundryOrder.Total,
			Subtotal:           laundryOrder.Total,
			TotalPrice:         totalPrice,        
			PlatformCommission: 0,                 
			AddonsTotal:        0,                 
			Status:             laundryOrder.Status,
			AssignedProviderID: laundryOrder.ProviderID,
			CreatedAt:          laundryOrder.CreatedAt,
			UpdatedAt:          laundryOrder.UpdatedAt,
			CustomerInfo: models.CustomerInfo{
				Name:    customerName,
				Address: laundryOrder.Address,
				Lat:     laundryOrder.Latitude,
				Lng:     laundryOrder.Longitude,
			},
			BookingInfo: models.BookingInfo{
				Date: bookingDate,
				Time: "",
			},
			SelectedServices: selectedServices,
		}

		allOrders = append(allOrders, serviceOrder)
	}

	allOrders = append(allOrders, serviceOrders...)

	total = int64(len(allOrders))

	orderClause := query.SortBy
	if orderClause == "booking_date" {
		orderClause = "created_at" 
	} else if orderClause == "price" {}

	if query.SortDesc {
	}

	offset := query.PaginationParams.GetOffset()
	limit := query.Limit

	start := offset
	end := offset + limit
	if start > len(allOrders) {
		start = len(allOrders)
	}
	if end > len(allOrders) {
		end = len(allOrders)
	}

	paginatedOrders := allOrders[start:end]

	logger.Info("returning paginated orders", "offset", offset, "limit", limit, "total", total, "returned", len(paginatedOrders))

	return paginatedOrders, total, nil
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

	if err == nil {
		logger.Info("found available service order", "orderID", orderID)
		return &order, nil
	}

	if err == gorm.ErrRecordNotFound {
		var laundryOrder models.LaundryOrder
		err = r.db.WithContext(ctx).
			Where("id = ?", orderID).
			Where("status IN ?", []string{"pending", "searching_provider"}).
			Where("category_slug IN ?", categorySlugs).
			Where("provider_id IS NULL").
			First(&laundryOrder).Error

		if err == nil {
			logger.Info("found available laundry order", "orderID", orderID)
			return r.convertLaundryOrderToServiceOrder(ctx, orderID)
		}
	}

	logger.Warn("available order not found", "orderID", orderID, "categorySlugs", categorySlugs)
	return nil, err
}

func (r *repository) GetProviderOrders(ctx context.Context, providerID string, query dto.ListMyOrdersQuery) ([]*models.ServiceOrderNew, int64, error) {
	var allOrders []*models.ServiceOrderNew
	var total int64

	var serviceOrders []*models.ServiceOrderNew
	db := r.db.WithContext(ctx).Model(&models.ServiceOrderNew{}).
		Where("assigned_provider_id = ?", providerID)
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

	if err := db.Order("created_at DESC").Find(&serviceOrders).Error; err != nil {
		return nil, 0, err
	}

	var laundryOrders []*models.LaundryOrder
	laundryDB := r.db.WithContext(ctx).
		Where("provider_id = ?", providerID)

	if query.Status != "" {
		laundryDB = laundryDB.Where("status = ?", query.Status)
	}

	if query.FromDate != "" {
		fromDate, _ := time.Parse("2006-01-02", query.FromDate)
		laundryDB = laundryDB.Where("created_at >= ?", fromDate)
	}
	if query.ToDate != "" {
		toDate, _ := time.Parse("2006-01-02", query.ToDate)
		toDate = toDate.AddDate(0, 0, 1)
		laundryDB = laundryDB.Where("created_at < ?", toDate)
	}

	if err := laundryDB.Order("created_at DESC").Find(&laundryOrders).Error; err != nil {
		return nil, 0, err
	}

	for _, laundryOrder := range laundryOrders {
		var customer models.User
		customerName := ""
		if laundryOrder.UserID != nil {
			if err := r.db.WithContext(ctx).Where("id = ?", *laundryOrder.UserID).First(&customer).Error; err == nil {
				customerName = customer.Name
			}
		}

		var items []*models.LaundryOrderItem
		r.db.WithContext(ctx).Where("order_id = ?", laundryOrder.ID).Find(&items)

		selectedServices := make(models.SelectedServices, 0)
		for _, item := range items {
			selectedServices = append(selectedServices, models.SelectedServiceItem{
				ServiceSlug: item.ServiceSlug,
				Title:       item.ItemType,
				Price:       item.Price,
				Quantity:    item.Quantity,
			})
		}

		bookingDate := ""
		if laundryOrder.ServiceDate != nil {
			bookingDate = laundryOrder.ServiceDate.Format("2006-01-02")
		}

		customerID := ""
		if laundryOrder.UserID != nil {
			customerID = *laundryOrder.UserID
		}

		totalPrice := laundryOrder.Total
		if laundryOrder.Tip != nil && *laundryOrder.Tip > 0 {
			totalPrice = laundryOrder.Total + *laundryOrder.Tip
		}

		serviceOrder := &models.ServiceOrderNew{
			ID:                 laundryOrder.ID,
			OrderNumber:        laundryOrder.OrderNumber,
			CustomerID:         customerID,
			CategorySlug:       laundryOrder.CategorySlug,
			ServicesTotal:      laundryOrder.Total, 
			Subtotal:           laundryOrder.Total, 
			TotalPrice:         totalPrice,         
			PlatformCommission: 0,                  
			AddonsTotal:        0,                  
			Status:             laundryOrder.Status,
			AssignedProviderID: laundryOrder.ProviderID,
			CreatedAt:          laundryOrder.CreatedAt,
			UpdatedAt:          laundryOrder.UpdatedAt,
			CustomerInfo: models.CustomerInfo{
				Name:    customerName,
				Address: laundryOrder.Address,
				Lat:     laundryOrder.Latitude,
				Lng:     laundryOrder.Longitude,
			},

			BookingInfo: models.BookingInfo{
				Date: bookingDate,
				Time: "",
			},
			SelectedServices: selectedServices,
		}

		allOrders = append(allOrders, serviceOrder)
	}
	allOrders = append(allOrders, serviceOrders...)

	sortField := "created_at"
	if query.SortBy == "booking_date" {
		sortField = "booking_info.date"
	}

	if query.SortDesc {
		sort.Slice(allOrders, func(i, j int) bool {
			if sortField == "created_at" {
				return allOrders[i].CreatedAt.After(allOrders[j].CreatedAt)
			}
			return false
		})
	} else {
		sort.Slice(allOrders, func(i, j int) bool {
			if sortField == "created_at" {
				return allOrders[i].CreatedAt.Before(allOrders[j].CreatedAt)
			}
			return false
		})
	}

	total = int64(len(allOrders))

	offset := query.PaginationParams.GetOffset()
	limit := query.Limit

	var paginatedOrders []*models.ServiceOrderNew
	if offset < int(total) {
		end := offset + limit
		if end > int(total) {
			end = int(total)
		}
		paginatedOrders = allOrders[offset:end]
	}

	return paginatedOrders, total, nil
}

func (r *repository) GetProviderOrderByID(ctx context.Context, providerID, orderID string) (*models.ServiceOrderNew, error) {
	var order models.ServiceOrderNew
	err := r.db.WithContext(ctx).
		Where("id = ? AND assigned_provider_id = ?", orderID, providerID).
		First(&order).Error

	if err == nil {
		return &order, nil
	}

	if err == gorm.ErrRecordNotFound {
		var laundryOrder models.LaundryOrder
		err = r.db.WithContext(ctx).
			Where("id = ? AND provider_id = ?", orderID, providerID).
			First(&laundryOrder).Error

		if err == nil {
			return r.convertLaundryOrderToServiceOrder(ctx, orderID)
		}
	}

	return nil, err
}

func (r *repository) CountProviderActiveOrders(ctx context.Context, providerID string) (int64, error) {
	var serviceOrderCount int64
	var laundryOrderCount int64

	err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("assigned_provider_id = ?", providerID).
		Where("status IN ?", shared.ActiveOrderStatuses()).
		Count(&serviceOrderCount).Error
	if err != nil {
		logger.Error("failed to count active service orders", "error", err, "providerID", providerID)
		return 0, err
	}

	err = r.db.WithContext(ctx).
		Model(&models.LaundryOrder{}).
		Where("provider_id = ?", providerID).
		Where("status NOT IN ?", []string{"completed", "cancelled"}).
		Count(&laundryOrderCount).Error
	if err != nil {
		logger.Error("failed to count active laundry orders", "error", err, "providerID", providerID)
		return 0, err
	}

	total := serviceOrderCount + laundryOrderCount
	logger.Info("counted active orders", "providerID", providerID, "serviceOrders", serviceOrderCount, "laundryOrders", laundryOrderCount, "total", total)

	return total, nil
}

func (r *repository) GetOrderByID(ctx context.Context, orderID string) (*models.ServiceOrderNew, error) {
	var order models.ServiceOrderNew

	err := r.db.WithContext(ctx).Where("id = ?", orderID).First(&order).Error
	if err == nil {
		return &order, nil
	}

	if err == gorm.ErrRecordNotFound {
		return r.convertLaundryOrderToServiceOrder(ctx, orderID)
	}

	return nil, err
}

func (r *repository) UpdateOrder(ctx context.Context, order *models.ServiceOrderNew) error {
	var laundryOrder models.LaundryOrder
	err := r.db.WithContext(ctx).Where("id = ?", order.ID).First(&laundryOrder).Error

	if err == nil {
		updates := map[string]interface{}{
			"status":     order.Status,
			"updated_at": time.Now(),
		}
		if order.AssignedProviderID != nil {
			updates["provider_id"] = *order.AssignedProviderID
		}

		result := r.db.WithContext(ctx).
			Model(&models.LaundryOrder{}).
			Where("id = ?", order.ID).
			Updates(updates)

		if result.Error != nil {
			logger.Error("failed to update laundry order", "error", result.Error, "orderID", order.ID)
			return result.Error
		}

		return nil
	} else if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Save(order).Error
	}

	return err
}

func (r *repository) GetProviderStatistics(ctx context.Context, providerID string) (*ProviderStats, error) {
	stats := &ProviderStats{}

	var serviceCompletedCount int64
	var serviceEarnings float64
	err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("assigned_provider_id = ? AND status = ?", providerID, shared.OrderStatusCompleted).
		Select("COUNT(*) as total_completed_jobs, COALESCE(SUM(total_price * 0.9), 0) as total_earnings").
		Row().Scan(&serviceCompletedCount, &serviceEarnings)
	if err != nil {
		return nil, err
	}

	var laundryCompletedCount int64
	var laundryEarnings float64
	err = r.db.WithContext(ctx).
		Model(&models.LaundryOrder{}).
		Where("provider_id = ? AND status = ?", providerID, shared.OrderStatusCompleted).
		Select("COUNT(*) as total_completed_jobs, COALESCE(SUM(total * 0.9), 0) as total_earnings").
		Row().Scan(&laundryCompletedCount, &laundryEarnings)
	if err != nil {
		return nil, err
	}

	stats.TotalCompletedJobs = int(serviceCompletedCount + laundryCompletedCount)
	stats.TotalEarnings = serviceEarnings + laundryEarnings

	err = r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("assigned_provider_id = ? AND customer_rating IS NOT NULL", providerID).
		Select("COUNT(*) as total_ratings, COALESCE(SUM(customer_rating), 0) as total_rating_sum").
		Row().Scan(&stats.TotalRatings, &stats.TotalRatingSum)
	if err != nil {
		return nil, err
	}

	activeCount, err := r.CountProviderActiveOrders(ctx, providerID)
	if err != nil {
		return nil, err
	}
	stats.ActiveOrders = int(activeCount)

	today := time.Now().Truncate(24 * time.Hour)
	var todayServiceCompleted int64
	var todayServiceEarnings float64
	err = r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("assigned_provider_id = ? AND status = ? AND completed_at >= ?",
			providerID, shared.OrderStatusCompleted, today).
		Select("COUNT(*) as today_completed_orders, COALESCE(SUM(total_price * 0.9), 0) as today_earnings").
		Row().Scan(&todayServiceCompleted, &todayServiceEarnings)
	if err != nil {
		return nil, err
	}

	var todayLaundryCompleted int64
	var todayLaundryEarnings float64
	err = r.db.WithContext(ctx).
		Model(&models.LaundryOrder{}).
		Where("provider_id = ? AND status = ? AND updated_at >= ?",
			providerID, shared.OrderStatusCompleted, today).
		Select("COUNT(*) as today_completed_orders, COALESCE(SUM(total * 0.9), 0) as today_earnings").
		Row().Scan(&todayLaundryCompleted, &todayLaundryEarnings)
	if err != nil {
		return nil, err
	}

	stats.TodayCompletedOrders = int(todayServiceCompleted + todayLaundryCompleted)
	stats.TodayEarnings = todayServiceEarnings + todayLaundryEarnings

	return stats, nil
}

func (r *repository) GetProviderEarnings(ctx context.Context, providerID string, fromDate, toDate time.Time) (*EarningsData, error) {
	earnings := &EarningsData{}

	err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("assigned_provider_id = ? AND status = ?", providerID, shared.OrderStatusCompleted).
		Where("completed_at >= ? AND completed_at < ?", fromDate, toDate.AddDate(0, 0, 1)).
		Select("COALESCE(SUM(total_price * 0.9), 0) as total_earnings, COUNT(*) as total_orders").
		Row().Scan(&earnings.TotalEarnings, &earnings.TotalOrders)
	if err != nil {
		return nil, err
	}

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
