package laundry

import (
	"context"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"gorm.io/gorm"
)

// =====================================================
// Repository Interface
// =====================================================

type Repository interface {
	// Catalog
	GetServiceCatalog(ctx context.Context) ([]*models.LaundryServiceCatalog, error)
	GetServiceBySlug(ctx context.Context, slug string) (*models.LaundryServiceCatalog, error)

	// Provider Management
	FindProviderByUserIDAndCategory(ctx context.Context, userID, category string) (*models.ServiceProviderProfile, error)
	GetProviderByID(ctx context.Context, providerID string) (*models.ServiceProviderProfile, error)
	CreateProvider(ctx context.Context, provider *models.ServiceProviderProfile) error
	AddProviderService(ctx context.Context, providerID, serviceSlug string) error
	GetProviderServices(ctx context.Context, providerID string) ([]string, error)
	GetAvailableOrdersByCategory(ctx context.Context, category string, serviceSlugs []string) ([]*models.LaundryOrder, error)

	// Pickups & Deliveries (handled by provider)
	CreatePickup(ctx context.Context, pickup *models.LaundryPickup) error
	GetPickupByOrder(ctx context.Context, orderID string) (*models.LaundryPickup, error)
	UpdatePickupStatus(ctx context.Context, orderID, status string, pickedUpAt *time.Time) error
	GetPickupsByProvider(ctx context.Context, providerID string, statuses []string) ([]*models.LaundryPickup, error)

	CreateDelivery(ctx context.Context, delivery *models.LaundryDelivery) error
	GetDeliveryByOrder(ctx context.Context, orderID string) (*models.LaundryDelivery, error)
	UpdateDeliveryStatus(ctx context.Context, orderID, status string, deliveredAt *time.Time) error
	GetDeliveriesByProvider(ctx context.Context, providerID string, statuses []string) ([]*models.LaundryDelivery, error)

	// Items
	CreateItems(ctx context.Context, items []*models.LaundryOrderItem) error
	GetOrderItems(ctx context.Context, orderID string) ([]*models.LaundryOrderItem, error)
	UpdateItemStatus(ctx context.Context, qrCode, status string) error
	GetItemByQRCode(ctx context.Context, qrCode string) (*models.LaundryOrderItem, error)

	// Issues
	CreateIssue(ctx context.Context, issue *models.LaundryIssue) error
	GetIssuesByProvider(ctx context.Context, providerID string, statuses []string) ([]*models.LaundryIssue, error)
	GetIssuesByOrder(ctx context.Context, orderID string) ([]*models.LaundryIssue, error)
	UpdateIssueStatus(ctx context.Context, issueID, status string, resolution *string, refundAmount *float64) error

	// Services & Products
	GetServicesWithProducts(ctx context.Context) ([]*models.LaundryServiceCatalog, error)
	GetServiceProducts(ctx context.Context, serviceSlug string) ([]*models.LaundryServiceProduct, error)
	GetProductBySlug(ctx context.Context, serviceSlug, productSlug string) (*models.LaundryServiceProduct, error)
}

// =====================================================
// Repository Implementation
// =====================================================

type repository struct {
	db *gorm.DB
}

// =====================================================
// Services with Products Methods
// =====================================================

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetServicesWithProducts(ctx context.Context) ([]*models.LaundryServiceCatalog, error) {
	var services []*models.LaundryServiceCatalog
	err := r.db.WithContext(ctx).
		Preload("Products", "is_active = ?", true, func(db *gorm.DB) *gorm.DB {
			return db.Order("display_order ASC")
		}).
		Where("is_active = ?", true).
		Order("display_order ASC").
		Find(&services).Error
	return services, err
}

func (r *repository) GetServiceProducts(ctx context.Context, serviceSlug string) ([]*models.LaundryServiceProduct, error) {
	var products []*models.LaundryServiceProduct
	err := r.db.WithContext(ctx).
		Where("service_slug = ? AND is_active = ?", serviceSlug, true).
		Order("display_order ASC").
		Find(&products).Error
	return products, err
}

func (r *repository) GetProductBySlug(ctx context.Context, serviceSlug, productSlug string) (*models.LaundryServiceProduct, error) {
	var product models.LaundryServiceProduct

	// First try to find by slug
	err := r.db.WithContext(ctx).
		Where("service_slug = ? AND slug = ? AND is_active = ?", serviceSlug, productSlug, true).
		First(&product).Error

	// If not found by slug, try by ID (UUID)
	if err != nil {
		err = r.db.WithContext(ctx).
			Where("id = ? AND service_slug = ? AND is_active = ?", productSlug, serviceSlug, true).
			First(&product).Error
	}

	// If product ID is empty, it means not found
	if product.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}

	return &product, err
}

// =====================================================
// Catalog Methods
// =====================================================

func (r *repository) GetServiceCatalog(ctx context.Context) ([]*models.LaundryServiceCatalog, error) {
	var services []*models.LaundryServiceCatalog
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("display_order ASC").
		Find(&services).Error
	return services, err
}

func (r *repository) GetServiceBySlug(ctx context.Context, slug string) (*models.LaundryServiceCatalog, error) {
	var service models.LaundryServiceCatalog
	err := r.db.WithContext(ctx).
		Where("slug = ? AND is_active = ?", slug, true).
		First(&service).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &service, err
}

// =====================================================
// Pickup Methods
// =====================================================

func (r *repository) CreatePickup(ctx context.Context, pickup *models.LaundryPickup) error {
	return r.db.WithContext(ctx).Create(pickup).Error
}

func (r *repository) GetPickupByOrder(ctx context.Context, orderID string) (*models.LaundryPickup, error) {
	var pickup models.LaundryPickup
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		First(&pickup).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &pickup, err
}

func (r *repository) UpdatePickupStatus(ctx context.Context, orderID, status string, pickedUpAt *time.Time) error {
	updates := map[string]interface{}{"status": status}
	if pickedUpAt != nil {
		updates["picked_up_at"] = pickedUpAt
	}
	if status == "arrived" {
		updates["arrived_at"] = time.Now()
	}
	return r.db.WithContext(ctx).
		Model(&models.LaundryPickup{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}

func (r *repository) GetPickupsByProvider(ctx context.Context, providerID string, statuses []string) ([]*models.LaundryPickup, error) {
	var pickups []*models.LaundryPickup
	query := r.db.WithContext(ctx).
		Where("provider_id = ?", providerID)

	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}

	err := query.Order("scheduled_at ASC").Find(&pickups).Error
	return pickups, err
}

// =====================================================
// Delivery Methods
// =====================================================

func (r *repository) CreateDelivery(ctx context.Context, delivery *models.LaundryDelivery) error {
	return r.db.WithContext(ctx).Create(delivery).Error
}

func (r *repository) GetDeliveryByOrder(ctx context.Context, orderID string) (*models.LaundryDelivery, error) {
	var delivery models.LaundryDelivery
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		First(&delivery).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &delivery, err
}

func (r *repository) UpdateDeliveryStatus(ctx context.Context, orderID, status string, deliveredAt *time.Time) error {
	updates := map[string]interface{}{"status": status}
	if deliveredAt != nil {
		updates["delivered_at"] = deliveredAt
	}
	if status == "arrived" {
		updates["arrived_at"] = time.Now()
	}
	return r.db.WithContext(ctx).
		Model(&models.LaundryDelivery{}).
		Where("order_id = ?", orderID).
		Updates(updates).Error
}

func (r *repository) GetDeliveriesByProvider(ctx context.Context, providerID string, statuses []string) ([]*models.LaundryDelivery, error) {
	var deliveries []*models.LaundryDelivery
	query := r.db.WithContext(ctx).
		Where("provider_id = ?", providerID)

	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}

	err := query.Order("scheduled_at ASC").Find(&deliveries).Error
	return deliveries, err
}

// =====================================================
// Item Methods
// =====================================================

func (r *repository) CreateItems(ctx context.Context, items []*models.LaundryOrderItem) error {
	return r.db.WithContext(ctx).Create(&items).Error
}

func (r *repository) GetOrderItems(ctx context.Context, orderID string) ([]*models.LaundryOrderItem, error) {
	var items []*models.LaundryOrderItem
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at ASC").
		Find(&items).Error
	return items, err
}

func (r *repository) UpdateItemStatus(ctx context.Context, qrCode, status string) error {
	updates := map[string]interface{}{"status": status}

	// Set timestamps based on status
	now := time.Now()
	switch status {
	case "received":
		updates["received_at"] = now
	case "packed":
		updates["packed_at"] = now
	case "delivered":
		updates["delivered_at"] = now
	}

	return r.db.WithContext(ctx).
		Model(&models.LaundryOrderItem{}).
		Where("qr_code = ?", qrCode).
		Updates(updates).Error
}

func (r *repository) GetItemByQRCode(ctx context.Context, qrCode string) (*models.LaundryOrderItem, error) {
	var item models.LaundryOrderItem
	err := r.db.WithContext(ctx).
		Where("qr_code = ?", qrCode).
		First(&item).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &item, err
}

// =====================================================
// Issue Methods
// =====================================================

func (r *repository) CreateIssue(ctx context.Context, issue *models.LaundryIssue) error {
	return r.db.WithContext(ctx).Create(issue).Error
}

func (r *repository) GetIssuesByProvider(ctx context.Context, providerID string, statuses []string) ([]*models.LaundryIssue, error) {
	var issues []*models.LaundryIssue
	query := r.db.WithContext(ctx).
		Where("provider_id = ?", providerID)

	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	} else {
		query = query.Where("status IN ?", []string{"open", "investigating"})
	}

	err := query.Order("priority DESC, created_at DESC").Find(&issues).Error
	return issues, err
}

func (r *repository) GetIssuesByOrder(ctx context.Context, orderID string) ([]*models.LaundryIssue, error) {
	var issues []*models.LaundryIssue
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at DESC").
		Find(&issues).Error
	return issues, err
}

func (r *repository) UpdateIssueStatus(ctx context.Context, issueID, status string, resolution *string, refundAmount *float64) error {
	updates := map[string]interface{}{"status": status}
	if resolution != nil {
		updates["resolution"] = resolution
	}
	if refundAmount != nil {
		updates["refund_amount"] = refundAmount
	}
	if status == "resolved" {
		updates["resolved_at"] = time.Now()
	}
	return r.db.WithContext(ctx).
		Model(&models.LaundryIssue{}).
		Where("id = ?", issueID).
		Updates(updates).Error
}

// =====================================================
// Provider Management Methods
// =====================================================

func (r *repository) FindProviderByUserIDAndCategory(ctx context.Context, userID, category string) (*models.ServiceProviderProfile, error) {
	var provider *models.ServiceProviderProfile
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND service_category = ?", userID, category).
		First(&provider).Error
	if err != nil {
		return nil, err
	}
	return provider, nil
}

func (r *repository) GetProviderByID(ctx context.Context, providerID string) (*models.ServiceProviderProfile, error) {
	var provider *models.ServiceProviderProfile
	err := r.db.WithContext(ctx).
		Where("id = ?", providerID).
		First(&provider).Error
	return provider, err
}

func (r *repository) CreateProvider(ctx context.Context, provider *models.ServiceProviderProfile) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

func (r *repository) AddProviderService(ctx context.Context, providerID, serviceSlug string) error {
	// Store the provider-service relationship
	// Using a JSON array in the ServiceType or a separate table would be ideal
	// For now, we'll just return nil as the service slug is inherent to orders filtered by category
	return nil
}

func (r *repository) GetProviderServices(ctx context.Context, providerID string) ([]string, error) {
	// For laundry, we don't explicitly assign services. Instead, we return ALL services
	// in the laundry category, so the provider automatically gets all current and future services
	var slugs []string
	err := r.db.WithContext(ctx).
		Model(&models.LaundryServiceCatalog{}).
		Where("category_slug = ? AND is_active = ?", "laundry", true).
		Pluck("slug", &slugs).Error

	if err != nil {
		logger.Error("GetProviderServices: failed to fetch laundry services",
			"error", err,
			"providerID", providerID,
		)
		// Return empty slice on error (provider would get no orders)
		return []string{}, nil
	}

	if len(slugs) == 0 {
		logger.Warn("GetProviderServices: no active laundry services found",
			"providerID", providerID,
		)
	}

	return slugs, nil
}

func (r *repository) GetAvailableOrdersByCategory(ctx context.Context, category string, serviceSlugs []string) ([]*models.LaundryOrder, error) {
	var orders []*models.LaundryOrder
	// Get unassigned orders in the laundry category with pending or ready status
	err := r.db.WithContext(ctx).
		Where("category_slug = ? AND provider_id IS NULL AND status IN ?", category, []string{"pending", "ready"}).
		Order("created_at ASC").
		Find(&orders).Error
	return orders, err
}
