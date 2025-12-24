package homeservices

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	homeServiceDto "github.com/umar5678/go-backend/internal/modules/homeservices/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

type Repository interface {
	// Service Categories & Tabs
	ListCategories(ctx context.Context) ([]models.ServiceCategory, error)
	GetCategoryByID(ctx context.Context, id uint) (*models.ServiceCategory, error)
	GetCategoryWithTabs(ctx context.Context, id uint) (*models.ServiceCategory, error)
	CreateCategory(ctx context.Context, category *models.ServiceCategory) error
	GetAllCategorySlugs(ctx context.Context) ([]string, error)
	ListServices(ctx context.Context, query homeServiceDto.ListServicesQuery) ([]*models.Service, int64, error)
	GetServiceByID(ctx context.Context, id uint) (*models.Service, error)
	GetServiceWithOptions(ctx context.Context, id uint) (*models.Service, error)
	GetServicesByIDs(ctx context.Context, ids []uint) ([]*models.Service, error)
	GetServicesByUUIDs(ctx context.Context, ids []string) ([]*models.ServiceNew, error)
	GetServiceNewByID(ctx context.Context, id string) (*models.ServiceNew, error)

	ListTabs(ctx context.Context, categoryID uint) ([]models.ServiceTab, error)
	GetTabByID(ctx context.Context, id uint) (*models.ServiceTab, error)
	CreateTab(ctx context.Context, tab *models.ServiceTab) error

	// Add-ons
	ListAddOns(ctx context.Context, categoryID uint) ([]models.AddOnService, error)
	GetAddOnByID(ctx context.Context, id uint) (*models.AddOnService, error)
	GetAddOnsByIDs(ctx context.Context, ids []uint) ([]*models.AddOnService, error)
	CreateAddOn(ctx context.Context, addon *models.AddOnService) error

	// Orders
	CreateOrder(ctx context.Context, order *models.ServiceOrder) error
	GetOrderByID(ctx context.Context, id string) (*models.ServiceOrder, error)
	GetOrderByIDWithDetails(ctx context.Context, id string) (*models.ServiceOrder, error)
	ListUserOrders(ctx context.Context, userID string, query homeServiceDto.ListOrdersQuery) ([]*models.ServiceOrder, int64, error)
	ListProviderOrders(ctx context.Context, providerID string, query homeServiceDto.ListOrdersQuery) ([]*models.ServiceOrder, int64, error)
	UpdateOrderStatus(ctx context.Context, orderID, status string) error
	AssignProviderToOrder(ctx context.Context, providerID, orderID string) error

	// Provider Registeration
	FindProviderByUserID(ctx context.Context, userID string) (*models.ServiceProviderProfile, error)

	// Provider Matching
	FindNearestAvailableProviders(ctx context.Context, serviceIDs []uint, lat, lon float64, radiusMeters int) ([]models.ServiceProvider, error)
	GetProviderByID(ctx context.Context, providerID string) (*models.ServiceProviderProfile, error)
	UpdateProviderStatus(ctx context.Context, providerID, status string) error
	UpdateProviderLocation(ctx context.Context, providerID string, lat, lon float64) error

	// Admin - Service Management
	CreateService(ctx context.Context, service *models.Service) error
	UpdateService(ctx context.Context, service *models.Service) error
	CreateServiceOption(ctx context.Context, option *models.ServiceOption) error
	CreateOptionChoice(ctx context.Context, choice *models.ServiceOptionChoice) error

	// Provider Management
	CreateProvider(ctx context.Context, provider *models.ServiceProviderProfile) error
	AssignServiceToProvider(ctx context.Context, providerID string, serviceID string) error
	RemoveServiceFromProvider(ctx context.Context, providerID string, serviceID string) error
	AddProviderCategory(ctx context.Context, category *models.ProviderServiceCategory) error
	GetProviderCategory(ctx context.Context, providerID string, categorySlug string) (*models.ProviderServiceCategory, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// ==================== CATEGORIES & TABS ====================

func (r *repository) ListCategories(ctx context.Context) ([]models.ServiceCategory, error) {
	var categories []models.ServiceCategory
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC, name ASC").
		Preload("Tabs", "is_active = ?", true).
		Find(&categories).Error
	return categories, err
}

func (r *repository) GetCategoryByID(ctx context.Context, id uint) (*models.ServiceCategory, error) {
	var category models.ServiceCategory
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error
	return &category, err
}

func (r *repository) GetCategoryWithTabs(ctx context.Context, id uint) (*models.ServiceCategory, error) {
	var category models.ServiceCategory
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Preload("Tabs", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_active = ?", true).Order("sort_order ASC, name ASC")
		}).
		First(&category).Error
	return &category, err
}

func (r *repository) CreateCategory(ctx context.Context, category *models.ServiceCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *repository) GetAllCategorySlugs(ctx context.Context) ([]string, error) {
	var slugs []string

	// Get distinct category slugs from all sources (ServiceNew, services, laundry products, etc.)
	// Using UNION to combine results from multiple tables
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT DISTINCT category_slug FROM (
				-- From ServiceNew table
				SELECT DISTINCT category_slug FROM services WHERE category_slug IS NOT NULL AND category_slug != ''
				UNION
				-- From services table
				SELECT DISTINCT category_slug FROM services WHERE category_slug IS NOT NULL AND category_slug != ''
				UNION
				-- From laundry_service_products table
				SELECT DISTINCT category_slug FROM laundry_service_products WHERE category_slug IS NOT NULL AND category_slug != ''
				UNION
				-- From laundry_service_catalog table
				SELECT DISTINCT category_slug FROM laundry_service_catalog WHERE category_slug IS NOT NULL AND category_slug != ''
			) AS all_slugs
			ORDER BY category_slug ASC
		`).
		Scan(&slugs).Error

	return slugs, err
}

func (r *repository) ListTabs(ctx context.Context, categoryID uint) ([]models.ServiceTab, error) {
	var tabs []models.ServiceTab
	err := r.db.WithContext(ctx).
		Where("category_id = ? AND is_active = ?", categoryID, true).
		Order("sort_order ASC, name ASC").
		Find(&tabs).Error
	return tabs, err
}

func (r *repository) GetTabByID(ctx context.Context, id uint) (*models.ServiceTab, error) {
	var tab models.ServiceTab
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tab).Error
	return &tab, err
}

func (r *repository) CreateTab(ctx context.Context, tab *models.ServiceTab) error {
	return r.db.WithContext(ctx).Create(tab).Error
}

// ==================== SERVICES ====================

func (r *repository) ListServices(ctx context.Context, query homeServiceDto.ListServicesQuery) ([]*models.Service, int64, error) {
	var services []*models.Service
	var total int64

	db := r.db.WithContext(ctx).Model(&models.Service{})

	// Apply filters
	if query.CategoryID != nil {
		db = db.Where("category_id = ?", *query.CategoryID)
	}

	if query.TabID != nil {
		db = db.Where("tab_id = ?", *query.TabID)
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

	if query.IsFeatured != nil {
		db = db.Where("is_featured = ?", *query.IsFeatured)
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	err := db.Offset(query.GetOffset()).
		Limit(query.Limit).
		Preload("Category").
		Preload("Tab").
		Order("is_featured DESC, sort_order ASC, name ASC").
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
		Preload("Tab").
		Preload("Options.Choices").
		Where("id = ?", id).
		First(&service).Error
	return &service, err
}

func (r *repository) CreateService(ctx context.Context, service *models.Service) error {
	return r.db.WithContext(ctx).Create(service).Error
}

func (r *repository) UpdateService(ctx context.Context, service *models.Service) error {
	return r.db.WithContext(ctx).Save(service).Error
}

func (r *repository) GetServicesByIDs(ctx context.Context, ids []uint) ([]*models.Service, error) {
	var services []*models.Service
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Preload("Options.Choices").
		Find(&services).Error
	return services, err
}

func (r *repository) GetServicesByUUIDs(ctx context.Context, ids []string) ([]*models.ServiceNew, error) {
	var services []*models.ServiceNew
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&services).Error
	return services, err
}

func (r *repository) GetServiceNewByID(ctx context.Context, id string) (*models.ServiceNew, error) {
	var service models.ServiceNew
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&service).Error
	return &service, err
}

// ==================== ADD-ONS ====================

func (r *repository) ListAddOns(ctx context.Context, categoryID uint) ([]models.AddOnService, error) {
	var addOns []models.AddOnService
	err := r.db.WithContext(ctx).
		Where("category_id = ? AND is_active = ?", categoryID, true).
		Order("sort_order ASC, title ASC").
		Find(&addOns).Error
	return addOns, err
}

func (r *repository) GetAddOnByID(ctx context.Context, id uint) (*models.AddOnService, error) {
	var addOn models.AddOnService
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&addOn).Error
	return &addOn, err
}

func (r *repository) GetAddOnsByIDs(ctx context.Context, ids []uint) ([]*models.AddOnService, error) {
	var addOns []*models.AddOnService
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&addOns).Error
	return addOns, err
}

func (r *repository) CreateAddOn(ctx context.Context, addon *models.AddOnService) error {
	return r.db.WithContext(ctx).Create(addon).Error
}

// ==================== ORDER METHODS ====================

func (r *repository) CreateOrder(ctx context.Context, order *models.ServiceOrder) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create order
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// Create order items
		if len(order.Items) > 0 {
			for i := range order.Items {
				order.Items[i].OrderID = order.ID
			}
			if err := tx.Create(&order.Items).Error; err != nil {
				return err
			}
		}

		// Create order add-ons
		if len(order.AddOns) > 0 {
			for i := range order.AddOns {
				order.AddOns[i].OrderID = order.ID
			}
			if err := tx.Create(&order.AddOns).Error; err != nil {
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
		Preload("AddOns").
		Preload("Provider").
		// Preload("Provider.User").
		Where("id = ?", id).
		First(&order).Error
	return &order, err
}

func (r *repository) ListUserOrders(ctx context.Context, userID string, query homeServiceDto.ListOrdersQuery) ([]*models.ServiceOrder, int64, error) {
	var ordersNew []*models.ServiceOrderNew
	var total int64

	// Query ServiceOrderNew which has customer_id field
	db := r.db.WithContext(ctx).Model(&models.ServiceOrderNew{}).Where("customer_id = ?", userID)

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
		Find(&ordersNew).Error

	if err != nil {
		return nil, 0, err
	}

	// Convert ServiceOrderNew to ServiceOrder for API compatibility
	orders := make([]*models.ServiceOrder, len(ordersNew))
	for i, orderNew := range ordersNew {
		orders[i] = convertServiceOrderNewToServiceOrder(orderNew)
	}

	return orders, total, nil
}

func (r *repository) ListProviderOrders(ctx context.Context, providerID string, query homeServiceDto.ListOrdersQuery) ([]*models.ServiceOrder, int64, error) {
	var ordersNew []*models.ServiceOrderNew
	var total int64

	// Query ServiceOrderNew which has assigned_provider_id field
	db := r.db.WithContext(ctx).Model(&models.ServiceOrderNew{}).Where("assigned_provider_id = ?", providerID)

	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch with pagination
	err := db.Order("created_at ASC").
		Offset(query.GetOffset()).
		Limit(query.Limit).
		Find(&ordersNew).Error

	if err != nil {
		return nil, 0, err
	}

	// Convert ServiceOrderNew to ServiceOrder for API compatibility
	orders := make([]*models.ServiceOrder, len(ordersNew))
	for i, orderNew := range ordersNew {
		orders[i] = convertServiceOrderNewToServiceOrder(orderNew)
	}

	return orders, total, nil
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

// ==================== PROVIDER MATCHING ====================

func (r *repository) FindNearestAvailableProviders(ctx context.Context, serviceIDs []uint, lat, lon float64, radiusMeters int) ([]models.ServiceProvider, error) {
	var providers []models.ServiceProvider

	// Query to find providers by their service category
	// Providers are matched based on their registered service category, not explicit service assignments
	// This way, providers automatically get access to all services in their category, including future ones
	err := r.db.WithContext(ctx).
		Table("service_providers").
		Select("DISTINCT service_providers.*").
		Joins("JOIN services ON services.category_slug = service_providers.service_category").
		Where("service_providers.status = ?", "available").
		Where("service_providers.is_verified = ?", true).
		Where("services.id IN ?", serviceIDs).
		Where("ST_DWithin(service_providers.location::geography, ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography, ?)",
			lon, lat, radiusMeters).
		Group("service_providers.id").
		Order(fmt.Sprintf("ST_Distance(service_providers.location::geography, ST_SetSRID(ST_MakePoint(%f, %f), 4326)::geography)", lon, lat)).
		Limit(10). // Top 10 nearest
		Scan(&providers).Error

	return providers, err
}

func (r *repository) GetProviderByID(ctx context.Context, providerID string) (*models.ServiceProviderProfile, error) {
	var provider models.ServiceProviderProfile
	err := r.db.WithContext(ctx).
		Where("id = ?", providerID).
		First(&provider).Error
	return &provider, err
}

func (r *repository) GetProviderWithUser(ctx context.Context, providerID string) (*models.ServiceProvider, error) {
	var provider models.ServiceProvider
	err := r.db.WithContext(ctx).
		Where("id = ?", providerID).
		First(&provider).Error
	return &provider, err
}

func (r *repository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("id = ?", userID).
		First(&user).Error
	return &user, err
}

// func (r *repository) GetProviderByID(ctx context.Context, providerID string) (*models.ServiceProvider, error) {
// 	var provider models.ServiceProvider
// 	err := r.db.WithContext(ctx).
// 		Preload("User").
// 		Where("id = ?", providerID).
// 		First(&provider).Error
// 	return &provider, err
// }

// Find provider by user ID
func (r *repository) FindProviderByUserID(ctx context.Context, userID string) (*models.ServiceProviderProfile, error) {
	var provider models.ServiceProviderProfile
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
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

func (r *repository) CreateServiceOption(ctx context.Context, option *models.ServiceOption) error {
	return r.db.WithContext(ctx).Create(option).Error
}

func (r *repository) CreateOptionChoice(ctx context.Context, choice *models.ServiceOptionChoice) error {
	return r.db.WithContext(ctx).Create(choice).Error
}

// --- Provider Management ---

func (r *repository) CreateProvider(ctx context.Context, provider *models.ServiceProviderProfile) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

func (r *repository) AssignServiceToProvider(ctx context.Context, providerID string, serviceID string) error {
	logger.Info("attempting to assign service to provider", "providerID", providerID, "serviceID", serviceID)

	result := r.db.WithContext(ctx).Exec(`
        INSERT INTO provider_qualified_services (provider_id, service_id)
        VALUES (?, ?)
        ON CONFLICT DO NOTHING
    `, providerID, serviceID)

	if result.Error != nil {
		logger.Error("failed to assign service to provider", "error", result.Error, "providerID", providerID, "serviceID", serviceID)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Warn("service assignment returned 0 rows affected (likely already exists)", "providerID", providerID, "serviceID", serviceID)
	} else {
		logger.Info("service assigned successfully", "providerID", providerID, "serviceID", serviceID, "rowsAffected", result.RowsAffected)
	}

	return nil
}

func (r *repository) RemoveServiceFromProvider(ctx context.Context, providerID string, serviceID string) error {
	return r.db.WithContext(ctx).Exec(`
        DELETE FROM provider_qualified_services
        WHERE provider_id = ? AND service_id = ?
    `, providerID, serviceID).Error
}

func (r *repository) AddProviderCategory(ctx context.Context, category *models.ProviderServiceCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *repository) GetProviderCategory(ctx context.Context, providerID string, categorySlug string) (*models.ProviderServiceCategory, error) {
	var category models.ProviderServiceCategory
	err := r.db.WithContext(ctx).
		Where("provider_id = ? AND category_slug = ?", providerID, categorySlug).
		First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// convertServiceOrderNewToServiceOrder converts ServiceOrderNew to ServiceOrder for API compatibility
func convertServiceOrderNewToServiceOrder(orderNew *models.ServiceOrderNew) *models.ServiceOrder {
	if orderNew == nil {
		return nil
	}

	return &models.ServiceOrder{
		ID:           orderNew.ID,
		Code:         orderNew.OrderNumber,
		UserID:       orderNew.CustomerID,
		Status:       orderNew.Status,
		Address:      orderNew.CustomerInfo.Address,
		Latitude:     orderNew.CustomerInfo.Lat,
		Longitude:    orderNew.CustomerInfo.Lng,
		ServiceDate:  orderNew.CreatedAt,
		CategorySlug: orderNew.CategorySlug,
		Subtotal:     orderNew.ServicesTotal,
		Total:        orderNew.TotalPrice,
		PlatformFee:  orderNew.PlatformCommission,
		CreatedAt:    orderNew.CreatedAt,
		AcceptedAt:   orderNew.ProviderAcceptedAt,
		StartedAt:    orderNew.ProviderStartedAt,
		CompletedAt:  orderNew.ProviderCompletedAt,
		CancelledAt:  orderNew.CompletedAt, // Map completed time as fallback
	}
}
