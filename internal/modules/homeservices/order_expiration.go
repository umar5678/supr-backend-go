package homeservices

import (
	"context"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"gorm.io/gorm"
)

// OrderExpirationService handles automatic expiration of orders
type OrderExpirationService struct {
	db *gorm.DB
}

// NewOrderExpirationService creates a new instance
func NewOrderExpirationService(db *gorm.DB) *OrderExpirationService {
	return &OrderExpirationService{db: db}
}

// ExpireUnacceptedOrders closes orders that have expired without provider acceptance
// Should be called periodically (e.g., every minute)
func (s *OrderExpirationService) ExpireUnacceptedOrders(ctx context.Context) error {
	logger.Info("Starting order expiration job")

	// Expire ServiceOrderNew orders
	result := s.db.WithContext(ctx).
		Model(&models.ServiceOrderNew{}).
		Where("status IN ?", []string{shared.OrderStatusPending, shared.OrderStatusSearchingProvider}).
		Where("expires_at IS NOT NULL AND expires_at <= ?", time.Now()).
		Update("status", shared.OrderStatusCancelled)

	if result.Error != nil {
		logger.Error("failed to expire service orders", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected > 0 {
		logger.Info("expired service orders", "count", result.RowsAffected)
	}

	// Expire LaundryOrder orders
	result = s.db.WithContext(ctx).
		Model(&models.LaundryOrder{}).
		Where("status IN ?", []string{"pending", "searching_provider"}).
		Where("expires_at IS NOT NULL AND expires_at <= ?", time.Now()).
		Update("status", "cancelled")

	if result.Error != nil {
		logger.Error("failed to expire laundry orders", "error", result.Error)
		return result.Error
	}

	if result.RowsAffected > 0 {
		logger.Info("expired laundry orders", "count", result.RowsAffected)
	}

	return nil
}
