package customer

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/customer/dto"
	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
)

// OrderRepository defines the interface for order data access
type OrderRepository interface {
	// Order CRUD
	Create(ctx context.Context, order *models.ServiceOrderNew) error
	GetByID(ctx context.Context, id string) (*models.ServiceOrderNew, error)
	GetByOrderNumber(ctx context.Context, orderNumber string) (*models.ServiceOrderNew, error)
	Update(ctx context.Context, order *models.ServiceOrderNew) error
	Delete(ctx context.Context, orderID string) error

	// Customer orders
	GetCustomerOrders(ctx context.Context, customerID string, query dto.ListOrdersQuery) ([]*models.ServiceOrderNew, int64, error)
	GetCustomerOrderByID(ctx context.Context, customerID, orderID string) (*models.ServiceOrderNew, error)
	CountCustomerActiveOrders(ctx context.Context, customerID string) (int64, error)

	// Status operations
	UpdateStatus(ctx context.Context, orderID, status string) error

	// Status history
	CreateStatusHistory(ctx context.Context, history *models.OrderStatusHistory) error
	GetOrderStatusHistory(ctx context.Context, orderID string) ([]models.OrderStatusHistory, error)
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new order repository instance
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *models.ServiceOrderNew) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) GetByID(ctx context.Context, id string) (*models.ServiceOrderNew, error) {
	var order models.ServiceOrderNew
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (*models.ServiceOrderNew, error) {
	var order models.ServiceOrderNew
	err := r.db.WithContext(ctx).Where("order_number = ?", orderNumber).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) Update(ctx context.Context, order *models.ServiceOrderNew) error {
	return r.db.WithContext(ctx).Model(&models.ServiceOrderNew{}).Where("id = ?", order.ID).Updates(order).Error
}

func (r *orderRepository) GetCustomerOrders(ctx context.Context, customerID string, query dto.ListOrdersQuery) ([]*models.ServiceOrderNew, int64, error) {
	var orders []*models.ServiceOrderNew
	var total int64

	db := r.db.WithContext(ctx).Model(&models.ServiceOrderNew{}).
		Where("customer_id = ?", customerID)

	// Apply filters
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	if query.FromDate != "" {
		fromDate, _ := time.Parse("2006-01-02", query.FromDate)
		db = db.Where("created_at >= ?", fromDate)
	}

	if query.ToDate != "" {
		toDate, _ := time.Parse("2006-01-02", query.ToDate)
		// Add 1 day to include the end date
		toDate = toDate.AddDate(0, 0, 1)
		db = db.Where("created_at < ?", toDate)
	}

	// Count total before pagination
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

func (r *orderRepository) GetCustomerOrderByID(ctx context.Context, customerID, orderID string) (*models.ServiceOrderNew, error) {
	var order models.ServiceOrderNew
	err := r.db.WithContext(ctx).
		Where("id = ? AND customer_id = ?", orderID, customerID).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) CountCustomerActiveOrders(ctx context.Context, customerID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("customer_id = ? AND status IN ?", customerID, shared.ActiveOrderStatuses()).
		Count(&count).Error
	return count, err
}

func (r *orderRepository) UpdateStatus(ctx context.Context, orderID, status string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	// Add timestamp based on status
	switch status {
	case shared.OrderStatusCompleted:
		updates["completed_at"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("id = ?", orderID).
		Updates(updates).Error
}

func (r *orderRepository) CreateStatusHistory(ctx context.Context, history *models.OrderStatusHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

func (r *orderRepository) GetOrderStatusHistory(ctx context.Context, orderID string) ([]models.OrderStatusHistory, error) {
	var history []models.OrderStatusHistory
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at ASC").
		Find(&history).Error
	return history, err
}

func (r *orderRepository) Delete(ctx context.Context, orderID string) error {
	return r.db.WithContext(ctx).Where("id = ?", orderID).Delete(&models.ServiceOrderNew{}).Error
}
