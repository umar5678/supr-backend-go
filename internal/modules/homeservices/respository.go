package homeservices

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/dto"
)

type Repository interface {
	// Service Catalog
	ListCategories(ctx context.Context) ([]models.ServiceCategory, error)
	GetCategoryByID(ctx context.Context, id uint) (*models.ServiceCategory, error)
	ListServices(ctx context.Context, query dto.ListServicesQuery) ([]*models.Service, int64, error)
	GetServiceByID(ctx context.Context, id uint) (*models.Service, error)
	GetServiceWithOptions(ctx context.Context, id uint) (*models.Service, error)

	// Orders
	CreateOrder(ctx context.Context, order *models.ServiceOrder, lat, lon float64) error
	GetOrderByID(ctx context.Context, id string) (*models.ServiceOrder, error)
	GetOrderByIDWithDetails(ctx context.Context, id string) (*models.ServiceOrder, error)
	ListUserOrders(ctx context.Context, userID string, query dto.ListOrdersQuery) ([]*models.ServiceOrder, int64, error)
	ListProviderOrders(ctx context.Context, providerID string, query dto.ListOrdersQuery) ([]*models.ServiceOrder, int64, error)
	UpdateOrderStatus(ctx context.Context, orderID, status string) error
	AssignProviderToOrder(ctx context.Context, providerID, orderID string) error

	// Provider Matching
	FindNearestAvailableProviders(ctx context.Context, serviceIDs []uint, lat, lon float64, radiusMeters int) ([]models.ServiceProvider, error)
	GetProviderByID(ctx context.Context, providerID string) (*models.ServiceProvider, error)
	UpdateProviderStatus(ctx context.Context, providerID, status string) error
	UpdateProviderLocation(ctx context.Context, providerID string, lat, lon float64) error

	// Admin - Service Management
	CreateService(ctx context.Context, service *models.Service) error
	UpdateService(ctx context.Context, service *models.Service) error
	CreateServiceOption(ctx context.Context, option *models.ServiceOption) error
	CreateOptionChoice(ctx context.Context, choice *models.ServiceOptionChoice) error

	// Provider Management
	CreateProvider(ctx context.Context, provider *models.ServiceProvider, lat, lon float64) error
	AssignServiceToProvider(ctx context.Context, providerID string, serviceID uint) error
	RemoveServiceFromProvider(ctx context.Context, providerID string, serviceID uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// --- Service Catalog Methods ---

func (r *repository) ListCategories(ctx context.Context) ([]models.ServiceCategory, error) {
	var categories []models.ServiceCategory
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC").
		Find(&categories).Error
	return categories, err
}

func (r *repository) GetCategoryByID(ctx context.Context, id uint) (*models.ServiceCategory, error) {
	var category models.ServiceCategory
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error
	return &category, err
}

func (r *repository) ListServices(ctx context.Context, query dto.ListServicesQuery) ([]*models.Service, int64, error) {
	var services []*models.Service
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Service{})

	// Apply filters
	if query.CategoryID != nil {
		db = db.Where("category_id = ?", *query.CategoryID)
	}

	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		db = db.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	if query.MinPrice != nil {
		db = db.Where("base_price >= ?", *query.MinPrice)
	}

	if query.MaxPrice != nil {
		db = db.Where("base_price <= ?", *query.MaxPrice)
	}

	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	} else {
		db = db.Where("is_active = ?", true) // Default to active only
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	err := db.Offset(query.GetOffset()).
		Limit(query.Limit).
		Preload("Category").
		Find(&services).Error

	return services, total, err
}

func (r *repository) GetServiceByID(ctx context.Context, id uint) (*models.Service, error) {
	var service models.Service
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&service).Error
	return &service, err
}

func (r *repository) GetServiceWithOptions(ctx context.Context, id uint) (*models.Service, error) {
	var service models.Service
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Options.Choices").
		Where("id = ?", id).
		First(&service).Error
	return &service, err
}

// --- Order Methods ---

func (r *repository) CreateOrder(ctx context.Context, order *models.ServiceOrder, lat, lon float64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create order with location using raw SQL for PostGIS
		result := tx.Exec(`
			INSERT INTO service_orders (
				id, code, user_id, status, address, location, service_date, 
				frequency, subtotal, discount, surge_fee, platform_fee, total, 
				coupon_code, wallet_hold_id, notes, created_at
			) VALUES (
				?, ?, ?, ?, ?, ST_SetSRID(ST_MakePoint(?, ?), 4326), ?, 
				?, ?, ?, ?, ?, ?, ?, ?, ?, NOW()
			)`,
			order.ID, order.Code, order.UserID, order.Status, order.Address,
			lon, lat, order.ServiceDate, order.Frequency, order.Subtotal,
			order.Discount, order.SurgeFee, order.PlatformFee, order.Total,
			order.CouponCode, order.WalletHoldID, order.Notes,
		)

		if result.Error != nil {
			return result.Error
		}

		// Create order items
		if len(order.Items) > 0 {
			if err := tx.Create(&order.Items).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *repository) GetOrderByID(ctx context.Context, id string) (*models.ServiceOrder, error) {
	var order models.ServiceOrder
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&order).Error
	return &order, err
}

func (r *repository) GetOrderByIDWithDetails(ctx context.Context, id string) (*models.ServiceOrder, error) {
	var order models.ServiceOrder
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Provider.User").
		Where("id = ?", id).
		First(&order).Error
	return &order, err
}

func (r *repository) ListUserOrders(ctx context.Context, userID string, query dto.ListOrdersQuery) ([]*models.ServiceOrder, int64, error) {
	var orders []*models.ServiceOrder
	var total int64

	db := r.db.WithContext(ctx).Model(&models.ServiceOrder{}).Where("user_id = ?", userID)

	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	err := db.Order("created_at DESC").
		Offset(query.GetOffset()).
		Limit(query.Limit).
		Preload("Items").
		Find(&orders).Error

	return orders, total, err
}

func (r *repository) ListProviderOrders(ctx context.Context, providerID string, query dto.ListOrdersQuery) ([]*models.ServiceOrder, int64, error) {
	var orders []*models.ServiceOrder
	var total int64

	db := r.db.WithContext(ctx).Model(&models.ServiceOrder{}).Where("provider_id = ?", providerID)

	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	err := db.Order("service_date ASC").
		Offset(query.GetOffset()).
		Limit(query.Limit).
		Preload("Items").
		Find(&orders).Error

	return orders, total, err
}

func (r *repository) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	// Set timestamp based on status
	switch status {
	case "accepted":
		updates["accepted_at"] = gorm.Expr("NOW()")
	case "in_progress":
		updates["started_at"] = gorm.Expr("NOW()")
	case "completed":
		updates["completed_at"] = gorm.Expr("NOW()")
	case "cancelled":
		updates["cancelled_at"] = gorm.Expr("NOW()")
	}

	return r.db.WithContext(ctx).
		Model(&models.ServiceOrder{}).
		Where("id = ?", orderID).
		Updates(updates).Error
}

func (r *repository) AssignProviderToOrder(ctx context.Context, providerID, orderID string) error {
	return r.db.WithContext(ctx).
		Model(&models.ServiceOrder{}).
		Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"provider_id": providerID,
			"status":      "accepted",
			"accepted_at": gorm.Expr("NOW()"),
		}).Error
}

// --- Provider Matching Methods ---

func (r *repository) FindNearestAvailableProviders(ctx context.Context, serviceIDs []uint, lat, lon float64, radiusMeters int) ([]models.ServiceProvider, error) {
	var providers []models.ServiceProvider

	// Complex query using PostGIS for geospatial search
	err := r.db.WithContext(ctx).
		Table("service_providers").
		Select("DISTINCT service_providers.*").
		Joins("JOIN provider_qualified_services pqs ON pqs.provider_id = service_providers.id").
		Where("service_providers.status = ?", "available").
		Where("service_providers.is_verified = ?", true).
		Where("pqs.service_id IN ?", serviceIDs).
		Where("ST_DWithin(service_providers.location::geography, ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography, ?)",
			lon, lat, radiusMeters).
		Group("service_providers.id").
		Having("COUNT(DISTINCT pqs.service_id) = ?", len(serviceIDs)). // Must be qualified for ALL services
		Order(fmt.Sprintf("ST_Distance(service_providers.location::geography, ST_SetSRID(ST_MakePoint(%f, %f), 4326)::geography)", lon, lat)).
		Limit(10). // Top 10 nearest
		Scan(&providers).Error

	return providers, err
}

func (r *repository) GetProviderByID(ctx context.Context, providerID string) (*models.ServiceProvider, error) {
	var provider models.ServiceProvider
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("id = ?", providerID).
		First(&provider).Error
	return &provider, err
}

func (r *repository) UpdateProviderStatus(ctx context.Context, providerID, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.ServiceProvider{}).
		Where("id = ?", providerID).
		Updates(map[string]interface{}{
			"status":      status,
			"last_active": gorm.Expr("NOW()"),
		}).Error
}

func (r *repository) UpdateProviderLocation(ctx context.Context, providerID string, lat, lon float64) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE service_providers 
		SET location = ST_SetSRID(ST_MakePoint(?, ?), 4326),
		    last_active = NOW()
		WHERE id = ?
	`, lon, lat, providerID).Error
}

// --- Admin - Service Management ---

func (r *repository) CreateService(ctx context.Context, service *models.Service) error {
	return r.db.WithContext(ctx).Create(service).Error
}

func (r *repository) UpdateService(ctx context.Context, service *models.Service) error {
	return r.db.WithContext(ctx).Save(service).Error
}

func (r *repository) CreateServiceOption(ctx context.Context, option *models.ServiceOption) error {
	return r.db.WithContext(ctx).Create(option).Error
}

func (r *repository) CreateOptionChoice(ctx context.Context, choice *models.ServiceOptionChoice) error {
	return r.db.WithContext(ctx).Create(choice).Error
}

// --- Provider Management ---

func (r *repository) CreateProvider(ctx context.Context, provider *models.ServiceProvider, lat, lon float64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create provider with location
		result := tx.Exec(`
			INSERT INTO service_providers (
				id, user_id, photo, rating, status, location, is_verified, 
				total_jobs, completed_jobs, created_at, updated_at
			) VALUES (
				?, ?, ?, ?, ?, ST_SetSRID(ST_MakePoint(?, ?), 4326), ?, 
				?, ?, NOW(), NOW()
			)`,
			provider.ID, provider.UserID, provider.Photo, provider.Rating,
			provider.Status, lon, lat, provider.IsVerified,
			provider.TotalJobs, provider.CompletedJobs,
		)
		return result.Error
	})
}

func (r *repository) AssignServiceToProvider(ctx context.Context, providerID string, serviceID uint) error {
	return r.db.WithContext(ctx).Exec(`
		INSERT INTO provider_qualified_services (provider_id, service_id)
		VALUES (?, ?)
		ON CONFLICT DO NOTHING
	`, providerID, serviceID).Error
}

func (r *repository) RemoveServiceFromProvider(ctx context.Context, providerID string, serviceID uint) error {
	return r.db.WithContext(ctx).Exec(`
		DELETE FROM provider_qualified_services
		WHERE provider_id = ? AND service_id = ?
	`, providerID, serviceID).Error
}
