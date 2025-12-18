package dto

import (
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
)

// ==================== List Orders Query ====================

// ListOrdersQuery represents query parameters for listing orders
type ListOrdersQuery struct {
	shared.PaginationParams

	// Filters
	Status       string `form:"status"`
	CategorySlug string `form:"category"`
	CustomerID   string `form:"customerId"`
	ProviderID   string `form:"providerId"`
	OrderNumber  string `form:"orderNumber"`

	// Date filters
	FromDate        string `form:"fromDate"`        // YYYY-MM-DD (created_at)
	ToDate          string `form:"toDate"`          // YYYY-MM-DD (created_at)
	BookingFromDate string `form:"bookingFromDate"` // YYYY-MM-DD (booking date)
	BookingToDate   string `form:"bookingToDate"`   // YYYY-MM-DD (booking date)

	// Price filters
	MinPrice *float64 `form:"minPrice" binding:"omitempty,gte=0"`
	MaxPrice *float64 `form:"maxPrice" binding:"omitempty,gte=0"`

	// Sorting
	SortBy   string `form:"sortBy" binding:"omitempty,oneof=created_at booking_date total_price status completed_at"`
	SortDesc bool   `form:"sortDesc"`

	// Include related data
	IncludeCustomer bool `form:"includeCustomer"`
	IncludeProvider bool `form:"includeProvider"`
}

// SetDefaults sets default values
func (q *ListOrdersQuery) SetDefaults() {
	q.PaginationParams.SetDefaults()
	if q.SortBy == "" {
		q.SortBy = "created_at"
	}
}

// Validate validates the query parameters
func (q *ListOrdersQuery) Validate() error {
	// Validate status
	if q.Status != "" && !shared.IsValidOrderStatus(q.Status) {
		return fmt.Errorf("invalid status: %s", q.Status)
	}

	// Validate date formats
	if q.FromDate != "" {
		if _, err := time.Parse("2006-01-02", q.FromDate); err != nil {
			return fmt.Errorf("invalid fromDate format, expected YYYY-MM-DD")
		}
	}
	if q.ToDate != "" {
		if _, err := time.Parse("2006-01-02", q.ToDate); err != nil {
			return fmt.Errorf("invalid toDate format, expected YYYY-MM-DD")
		}
	}
	if q.BookingFromDate != "" {
		if _, err := time.Parse("2006-01-02", q.BookingFromDate); err != nil {
			return fmt.Errorf("invalid bookingFromDate format, expected YYYY-MM-DD")
		}
	}
	if q.BookingToDate != "" {
		if _, err := time.Parse("2006-01-02", q.BookingToDate); err != nil {
			return fmt.Errorf("invalid bookingToDate format, expected YYYY-MM-DD")
		}
	}

	// Validate price range
	if q.MinPrice != nil && q.MaxPrice != nil && *q.MinPrice > *q.MaxPrice {
		return fmt.Errorf("minPrice cannot be greater than maxPrice")
	}

	return nil
}

// ==================== Update Order Status ====================

// UpdateOrderStatusRequest represents status update request
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Reason string `json:"reason" binding:"omitempty,max=500"`
	Notes  string `json:"notes" binding:"omitempty,max=1000"`
}

// Validate validates the request
func (r *UpdateOrderStatusRequest) Validate() error {
	if !shared.IsValidOrderStatus(r.Status) {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	return nil
}

// ==================== Reassign Order ====================

// ReassignOrderRequest represents order reassignment request
type ReassignOrderRequest struct {
	ProviderID string `json:"providerId" binding:"required,uuid"`
	Reason     string `json:"reason" binding:"required,min=10,max=500"`
}

// Validate validates the request
func (r *ReassignOrderRequest) Validate() error {
	if len(r.Reason) < 10 {
		return fmt.Errorf("reason must be at least 10 characters")
	}
	return nil
}

// ==================== Cancel Order ====================

// AdminCancelOrderRequest represents admin cancellation request
type AdminCancelOrderRequest struct {
	Reason        string   `json:"reason" binding:"required,min=10,max=500"`
	RefundAmount  *float64 `json:"refundAmount" binding:"omitempty,gte=0"` // Optional custom refund
	NotifyParties bool     `json:"notifyParties"`
}

// Validate validates the request
func (r *AdminCancelOrderRequest) Validate() error {
	if len(r.Reason) < 10 {
		return fmt.Errorf("reason must be at least 10 characters")
	}
	return nil
}

// ==================== Analytics Query ====================

// AnalyticsQuery represents query for analytics data
type AnalyticsQuery struct {
	FromDate string `form:"fromDate" binding:"required"` // YYYY-MM-DD
	ToDate   string `form:"toDate" binding:"required"`   // YYYY-MM-DD
	GroupBy  string `form:"groupBy" binding:"omitempty,oneof=day week month"`
}

// SetDefaults sets default values
func (q *AnalyticsQuery) SetDefaults() {
	if q.GroupBy == "" {
		q.GroupBy = "day"
	}
}

// Validate validates the query
func (q *AnalyticsQuery) Validate() error {
	if _, err := time.Parse("2006-01-02", q.FromDate); err != nil {
		return fmt.Errorf("invalid fromDate format")
	}
	if _, err := time.Parse("2006-01-02", q.ToDate); err != nil {
		return fmt.Errorf("invalid toDate format")
	}

	fromDate, _ := time.Parse("2006-01-02", q.FromDate)
	toDate, _ := time.Parse("2006-01-02", q.ToDate)

	if fromDate.After(toDate) {
		return fmt.Errorf("fromDate cannot be after toDate")
	}

	// Limit to 1 year max
	if toDate.Sub(fromDate) > 365*24*time.Hour {
		return fmt.Errorf("date range cannot exceed 1 year")
	}

	return nil
}

// ProviderAnalyticsQuery represents query for provider analytics
type ProviderAnalyticsQuery struct {
	FromDate     string   `form:"fromDate" binding:"required"`
	ToDate       string   `form:"toDate" binding:"required"`
	CategorySlug string   `form:"category"`
	MinOrders    *int     `form:"minOrders"`
	MinRating    *float64 `form:"minRating"`
	SortBy       string   `form:"sortBy" binding:"omitempty,oneof=completed_orders earnings rating"`
	SortDesc     bool     `form:"sortDesc"`
	Limit        int      `form:"limit" binding:"omitempty,min=1,max=100"`
}

// SetDefaults sets default values
func (q *ProviderAnalyticsQuery) SetDefaults() {
	if q.SortBy == "" {
		q.SortBy = "completed_orders"
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
}

// ==================== Bulk Actions ====================

// BulkUpdateStatusRequest represents bulk status update
type BulkUpdateStatusRequest struct {
	OrderIDs []string `json:"orderIds" binding:"required,min=1,max=100"`
	Status   string   `json:"status" binding:"required"`
	Reason   string   `json:"reason" binding:"required,min=10,max=500"`
}

// Validate validates the request
func (r *BulkUpdateStatusRequest) Validate() error {
	if !shared.IsValidOrderStatus(r.Status) {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	if len(r.OrderIDs) == 0 {
		return fmt.Errorf("at least one order ID required")
	}
	if len(r.OrderIDs) > 100 {
		return fmt.Errorf("cannot update more than 100 orders at once")
	}
	return nil
}

// ==================== Export Query ====================

// ExportOrdersQuery represents query for exporting orders
type ExportOrdersQuery struct {
	FromDate     string `form:"fromDate" binding:"required"`
	ToDate       string `form:"toDate" binding:"required"`
	Status       string `form:"status"`
	CategorySlug string `form:"category"`
	Format       string `form:"format" binding:"omitempty,oneof=csv xlsx json"`
}

// SetDefaults sets default values
func (q *ExportOrdersQuery) SetDefaults() {
	if q.Format == "" {
		q.Format = "csv"
	}
}

// Validate validates the query
func (q *ExportOrdersQuery) Validate() error {
	if _, err := time.Parse("2006-01-02", q.FromDate); err != nil {
		return fmt.Errorf("invalid fromDate format")
	}
	if _, err := time.Parse("2006-01-02", q.ToDate); err != nil {
		return fmt.Errorf("invalid toDate format")
	}
	if q.Status != "" && !shared.IsValidOrderStatus(q.Status) {
		return fmt.Errorf("invalid status: %s", q.Status)
	}
	return nil
}
