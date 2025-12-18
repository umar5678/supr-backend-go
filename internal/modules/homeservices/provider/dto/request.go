package dto

import (
	"fmt"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
)

// ==================== Profile Requests ====================

// UpdateProviderProfileRequest represents provider profile update
type UpdateProviderProfileRequest struct {
	Bio               *string  `json:"bio" binding:"omitempty,max=1000"`
	YearsOfExperience *int     `json:"yearsOfExperience" binding:"omitempty,min=0,max=50"`
	Photo             *string  `json:"photo" binding:"omitempty,url,max=500"`
	IsAvailable       *bool    `json:"isAvailable"`
	Latitude          *float64 `json:"latitude" binding:"omitempty,latitude"`
	Longitude         *float64 `json:"longitude" binding:"omitempty,longitude"`
}

// Validate validates the request
func (r *UpdateProviderProfileRequest) Validate() error {
	if r.Bio == nil && r.YearsOfExperience == nil && r.Photo == nil &&
		r.IsAvailable == nil && r.Latitude == nil && r.Longitude == nil {
		return fmt.Errorf("at least one field must be provided for update")
	}
	return nil
}

// UpdateAvailabilityRequest represents availability status update
type UpdateAvailabilityRequest struct {
	IsAvailable bool     `json:"isAvailable"`
	Latitude    *float64 `json:"latitude" binding:"omitempty,latitude"`
	Longitude   *float64 `json:"longitude" binding:"omitempty,longitude"`
}

// ==================== Service Category Requests ====================

// AddServiceCategoryRequest represents adding a service category
type AddServiceCategoryRequest struct {
	CategorySlug      string `json:"categorySlug" binding:"required,min=2,max=100"`
	ExpertiseLevel    string `json:"expertiseLevel" binding:"required,oneof=beginner intermediate expert"`
	YearsOfExperience int    `json:"yearsOfExperience" binding:"min=0,max=50"`
}

// Validate validates the request
func (r *AddServiceCategoryRequest) Validate() error {
	if !models.IsValidExpertiseLevel(r.ExpertiseLevel) {
		return fmt.Errorf("expertiseLevel must be one of: beginner, intermediate, expert")
	}
	return nil
}

// UpdateServiceCategoryRequest represents updating a service category
type UpdateServiceCategoryRequest struct {
	ExpertiseLevel    *string `json:"expertiseLevel" binding:"omitempty,oneof=beginner intermediate expert"`
	YearsOfExperience *int    `json:"yearsOfExperience" binding:"omitempty,min=0,max=50"`
	IsActive          *bool   `json:"isActive"`
}

// Validate validates the request
func (r *UpdateServiceCategoryRequest) Validate() error {
	if r.ExpertiseLevel == nil && r.YearsOfExperience == nil && r.IsActive == nil {
		return fmt.Errorf("at least one field must be provided for update")
	}
	if r.ExpertiseLevel != nil && !models.IsValidExpertiseLevel(*r.ExpertiseLevel) {
		return fmt.Errorf("expertiseLevel must be one of: beginner, intermediate, expert")
	}
	return nil
}

// ==================== Order Requests ====================

// RejectOrderRequest represents order rejection
type RejectOrderRequest struct {
	Reason string `json:"reason" binding:"required,min=10,max=500"`
}

// Validate validates the request
func (r *RejectOrderRequest) Validate() error {
	if len(r.Reason) < 10 {
		return fmt.Errorf("rejection reason must be at least 10 characters")
	}
	return nil
}

// CompleteOrderRequest represents order completion (optional notes)
type CompleteOrderRequest struct {
	Notes string `json:"notes" binding:"omitempty,max=1000"`
}

// RateCustomerRequest represents rating a customer
type RateCustomerRequest struct {
	Rating int    `json:"rating" binding:"required,min=1,max=5"`
	Review string `json:"review" binding:"omitempty,max=1000"`
}

// Validate validates the request
func (r *RateCustomerRequest) Validate() error {
	if r.Rating < 1 || r.Rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}
	return nil
}

// ==================== Query Requests ====================

// ListAvailableOrdersQuery represents query for available orders
type ListAvailableOrdersQuery struct {
	shared.PaginationParams
	CategorySlug string `form:"category"`
	Date         string `form:"date"` // YYYY-MM-DD
	SortBy       string `form:"sortBy" binding:"omitempty,oneof=created_at booking_date distance price"`
	SortDesc     bool   `form:"sortDesc"`
}

// SetDefaults sets default values
func (q *ListAvailableOrdersQuery) SetDefaults() {
	q.PaginationParams.SetDefaults()
	if q.SortBy == "" {
		q.SortBy = "created_at"
	}
}

// ListMyOrdersQuery represents query for provider's orders
type ListMyOrdersQuery struct {
	shared.PaginationParams
	Status   string `form:"status"`
	FromDate string `form:"fromDate"` // YYYY-MM-DD
	ToDate   string `form:"toDate"`   // YYYY-MM-DD
	SortBy   string `form:"sortBy" binding:"omitempty,oneof=created_at booking_date completed_at"`
	SortDesc bool   `form:"sortDesc"`
}

// SetDefaults sets default values
func (q *ListMyOrdersQuery) SetDefaults() {
	q.PaginationParams.SetDefaults()
	if q.SortBy == "" {
		q.SortBy = "created_at"
	}
}

// Validate validates the query
func (q *ListMyOrdersQuery) Validate() error {
	if q.Status != "" && !shared.IsValidOrderStatus(q.Status) {
		return fmt.Errorf("invalid status: %s", q.Status)
	}
	return nil
}

// EarningsQuery represents query for earnings
type EarningsQuery struct {
	FromDate string `form:"fromDate" binding:"required"` // YYYY-MM-DD
	ToDate   string `form:"toDate" binding:"required"`   // YYYY-MM-DD
	GroupBy  string `form:"groupBy" binding:"omitempty,oneof=day week month"`
}

// SetDefaults sets default values
func (q *EarningsQuery) SetDefaults() {
	if q.GroupBy == "" {
		q.GroupBy = "day"
	}
}
