package homeServiceDto

import (
	"errors"
	"fmt"
)

// CreateOrderRequest represents a customer's service booking request
type CreateOrderRequest struct {
	Items          []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive"`
	AddOnIDs       []uint                   `json:"addOnIds" binding:"omitempty"`
	Address        string                   `json:"address" binding:"required,min=5,max=500"`
	Latitude       float64                  `json:"latitude" binding:"required,latitude"`
	Longitude      float64                  `json:"longitude" binding:"required,longitude"`
	ServiceDate    string                   `json:"serviceDate" binding:"required"` // RFC3339 format
	Frequency      string                   `json:"frequency" binding:"omitempty,oneof=once daily weekly monthly"`
	QuantityOfPros int                      `json:"quantityOfPros" binding:"required,min=1,max=10"`   // ✅ NEW: Number of professionals needed
	HoursOfService float64                  `json:"hoursOfService" binding:"required,min=0.5,max=24"` // ✅ NEW: Hours of service required
	Notes          *string                  `json:"notes" binding:"omitempty,max=500"`
	CouponCode     *string                  `json:"couponCode" binding:"omitempty,max=50"`
}

type CreateOrderItemRequest struct {
	ServiceID       string                  `json:"serviceId" binding:"required,min=1"`
	SelectedOptions []SelectedOptionRequest `json:"selectedOptions" binding:"omitempty,dive"`
}

type SelectedOptionRequest struct {
	OptionID uint    `json:"optionId" binding:"required,min=1"`
	ChoiceID *uint   `json:"choiceId" binding:"omitempty,min=1"` // For select_single/select_multiple
	Value    *string `json:"value" binding:"omitempty"`          // For text/quantity types
}

func (r *CreateOrderRequest) Validate() error {
	if len(r.Items) == 0 {
		return errors.New("at least one service item is required")
	}

	// Validate quantity of professionals
	if r.QuantityOfPros < 1 {
		return errors.New("at least 1 professional is required")
	}
	if r.QuantityOfPros > 10 {
		return errors.New("maximum 10 professionals allowed per order")
	}

	//  Validate hours of service
	if r.HoursOfService < 0.5 {
		return errors.New("minimum service duration is 0.5 hours")
	}
	if r.HoursOfService > 24 {
		return errors.New("maximum service duration is 24 hours")
	}

	//  Validate hours increment (must be in 0.5 hour increments)
	if r.HoursOfService != float64(int(r.HoursOfService*2))/2 {
		return errors.New("hours must be in 0.5 hour increments (e.g., 1.5, 2, 2.5)")
	}

	if r.Frequency == "" {
		r.Frequency = "once"
	}
	return nil
}

func (r *CreateOrderRequest) SetDefaults() {
	if r.Frequency == "" {
		r.Frequency = "once"
	}
	if r.QuantityOfPros == 0 {
		r.QuantityOfPros = 1 // Default to 1 professional
	}
	if r.HoursOfService == 0 {
		r.HoursOfService = 1.0 // Default to 1 hour
	}
}

// UpdateOrderStatusRequest for provider actions
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=accepted rejected in_progress completed"`
}

// ListServicesQuery - Updated with tab filtering
type ListServicesQuery struct {
	Page       int      `form:"page" binding:"omitempty,min=1"`
	Limit      int      `form:"limit" binding:"omitempty,min=1,max=100"`
	CategoryID *uint    `form:"categoryId" binding:"omitempty,min=1"`
	TabID      *uint    `form:"tabId" binding:"omitempty,min=1"`
	Search     string   `form:"search" binding:"omitempty,max=100"`
	MinPrice   *float64 `form:"minPrice" binding:"omitempty,gte=0"`
	MaxPrice   *float64 `form:"maxPrice" binding:"omitempty,gte=0"`
	IsActive   *bool    `form:"isActive"`
	IsFeatured *bool    `form:"isFeatured"`
}

func (q *ListServicesQuery) SetDefaults() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
}

func (q *ListServicesQuery) GetOffset() int {
	return (q.Page - 1) * q.Limit
}

// ListOrdersQuery for order history
type ListOrdersQuery struct {
	Page   int     `form:"page" binding:"omitempty,min=1"`
	Limit  int     `form:"limit" binding:"omitempty,min=1,max=100"`
	Status *string `form:"status"`
}

func (q *ListOrdersQuery) SetDefaults() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
}

func (q *ListOrdersQuery) GetOffset() int {
	return (q.Page - 1) * q.Limit
}

// Admin Requests
type CreateCategoryRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=150"`
	Description string   `json:"description" binding:"required,max=1000"`
	IconURL     string   `json:"iconUrl" binding:"omitempty,url"`
	BannerImage string   `json:"bannerImage" binding:"omitempty,url"`
	Highlights  []string `json:"highlights" binding:"omitempty"`
	IsActive    bool     `json:"isActive"`
	SortOrder   int      `json:"sortOrder" binding:"omitempty"`
}

type CreateTabRequest struct {
	CategoryID  uint   `json:"categoryId" binding:"required,min=1"`
	Name        string `json:"name" binding:"required,min=2,max=150"`
	Description string `json:"description" binding:"omitempty,max=500"`
	IconURL     string `json:"iconUrl" binding:"omitempty,url"`
	BannerTitle string `json:"bannerTitle" binding:"omitempty,max=150"`
	BannerDesc  string `json:"bannerDescription" binding:"omitempty,max=500"`
	BannerImage string `json:"bannerImage" binding:"omitempty,url"`
	IsActive    bool   `json:"isActive"`
	SortOrder   int    `json:"sortOrder" binding:"omitempty"`
}
type CreateServiceRequest struct {
	CategoryID          uint    `json:"categoryId" binding:"required,min=1"`
	TabID               uint    `json:"tabId" binding:"required,min=1"`
	Name                string  `json:"name" binding:"required,min=3,max=150"`
	Description         string  `json:"description" binding:"required,max=2000"`
	ImageURL            string  `json:"imageUrl" binding:"omitempty,url"`
	BasePrice           float64 `json:"basePrice" binding:"required,gt=0"`
	OriginalPrice       float64 `json:"originalPrice" binding:"omitempty,gt=0"`
	PricingModel        string  `json:"pricingModel" binding:"required,oneof=fixed hourly per_unit"`
	BaseDurationMinutes int     `json:"baseDurationMinutes" binding:"required,min=1"`
	MaxQuantity         int     `json:"maxQuantity" binding:"omitempty,min=1"`
	IsFeatured          bool    `json:"isFeatured"`
}

type CreateAddOnRequest struct {
	CategoryID      uint    `json:"categoryId" binding:"required,min=1"`
	Title           string  `json:"title" binding:"required,min=2,max=150"`
	Description     string  `json:"description" binding:"omitempty,max=500"`
	ImageURL        string  `json:"imageUrl" binding:"omitempty,url"`
	Price           float64 `json:"price" binding:"required,gt=0"`
	OriginalPrice   float64 `json:"originalPrice" binding:"omitempty,gt=0"`
	DurationMinutes int     `json:"durationMinutes" binding:"omitempty,min=0"`
	IsActive        bool    `json:"isActive"`
	SortOrder       int     `json:"sortOrder" binding:"omitempty"`
}

type UpdateServiceRequest struct {
	CategoryID          *uint    `json:"categoryId" binding:"omitempty,min=1"`
	Name                *string  `json:"name" binding:"omitempty,min=3,max=150"`
	Description         *string  `json:"description" binding:"omitempty,max=2000"`
	ImageURL            *string  `json:"imageUrl" binding:"omitempty,url"`
	BasePrice           *float64 `json:"basePrice" binding:"omitempty,gt=0"`
	PricingModel        *string  `json:"pricingModel" binding:"omitempty,oneof=fixed hourly per_unit"`
	BaseDurationMinutes *int     `json:"baseDurationMinutes" binding:"omitempty,min=1"`
	IsActive            *bool    `json:"isActive"`
}

func (r *UpdateServiceRequest) Validate() error {
	hasUpdate := r.CategoryID != nil || r.Name != nil || r.Description != nil ||
		r.ImageURL != nil || r.BasePrice != nil || r.PricingModel != nil ||
		r.BaseDurationMinutes != nil || r.IsActive != nil

	if !hasUpdate {
		return fmt.Errorf("at least one field must be provided for update")
	}
	return nil
}

type CreateServiceOptionRequest struct {
	ServiceID  uint   `json:"serviceId" binding:"required,min=1"`
	Name       string `json:"name" binding:"required,min=2,max=100"`
	Type       string `json:"type" binding:"required,oneof=select_single select_multiple quantity text"`
	IsRequired bool   `json:"isRequired"`
}

type CreateOptionChoiceRequest struct {
	OptionID                uint    `json:"optionId" binding:"required,min=1"`
	Label                   string  `json:"label" binding:"required,min=1,max=100"`
	PriceModifier           float64 `json:"priceModifier"`
	DurationModifierMinutes int     `json:"durationModifierMinutes"`
}

// RegisterProviderRequest for new provider registration
// Note: Providers are automatically assigned ALL services in their registered category
// This enables dynamic service assignment - future services are automatically available
type RegisterProviderRequest struct {
	CategorySlug string  `json:"categorySlug" binding:"required,min=2,max=100"`
	Latitude     float64 `json:"latitude" binding:"required,latitude"`
	Longitude    float64 `json:"longitude" binding:"required,longitude"`
	Photo        *string `json:"photo" binding:"omitempty,url"`
	BusinessName *string `json:"businessName" binding:"omitempty,min=2,max=255"`
	Description  *string `json:"description" binding:"omitempty,max=1000"`
}

func (r *RegisterProviderRequest) Validate() error {
	if r.CategorySlug == "" {
		return errors.New("category slug is required")
	}
	return nil
}
