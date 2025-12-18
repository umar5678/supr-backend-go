package dto

import (
	"fmt"
	"regexp"
	"time"

	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
)

// ==================== Create Order Request ====================

// CustomerInfoRequest represents customer information for booking
type CustomerInfoRequest struct {
	Name    string  `json:"name" binding:"required,min=2,max=100"`
	Phone   string  `json:"phone" binding:"required"`
	Email   string  `json:"email" binding:"omitempty,email"`
	Address string  `json:"address" binding:"required,min=10,max=500"`
	Lat     float64 `json:"lat" binding:"required,latitude"`
	Lng     float64 `json:"lng" binding:"required,longitude"`
}

// Validate validates customer info
func (c *CustomerInfoRequest) Validate() error {
	// Validate phone format (basic validation)
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{10,15}$`)
	cleanPhone := regexp.MustCompile(`[\s\-\(\)]+`).ReplaceAllString(c.Phone, "")
	if !phoneRegex.MatchString(cleanPhone) {
		return fmt.Errorf("invalid phone number format")
	}
	return nil
}

// BookingInfoRequest represents booking date/time information
type BookingInfoRequest struct {
	Date           string `json:"date" binding:"required"`           // YYYY-MM-DD
	Time           string `json:"time" binding:"required"`           // HH:MM
	PreferredTime  string `json:"preferredTime" binding:"omitempty"` // optional specific time HH:MM
	QuantityOfPros int    `json:"quantityOfPros" binding:"required,min=1,max=5"`
}

// Validate validates booking info
func (b *BookingInfoRequest) Validate() error {
	// Validate date format
	_, err := time.Parse("2006-01-02", b.Date)
	if err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}

	// Validate time format
	_, err = time.Parse("15:04", b.Time)
	if err != nil {
		return fmt.Errorf("invalid time format, expected HH:MM")
	}

	// Validate booking is in the future (at least 2 hours from now)
	bookingDateTime, _ := time.Parse("2006-01-02 15:04", b.Date+" "+b.Time)
	minBookingTime := time.Now().Add(2 * time.Hour)
	if bookingDateTime.Before(minBookingTime) {
		return fmt.Errorf("booking must be at least 2 hours in the future")
	}

	// Validate booking is not too far in the future (max 30 days)
	maxBookingTime := time.Now().AddDate(0, 0, 30)
	if bookingDateTime.After(maxBookingTime) {
		return fmt.Errorf("booking cannot be more than 30 days in the future")
	}

	// Set day of week
	return nil
}

// GetDayOfWeek returns the day of week for the booking date
func (b *BookingInfoRequest) GetDayOfWeek() string {
	date, _ := time.Parse("2006-01-02", b.Date)
	return date.Weekday().String()
}

// SelectedServiceRequest represents a service selection in the order
type SelectedServiceRequest struct {
	ServiceSlug string `json:"serviceSlug" binding:"required"`
	Quantity    int    `json:"quantity" binding:"required,min=1,max=10"`
}

// SelectedAddonRequest represents an addon selection in the order
type SelectedAddonRequest struct {
	AddonSlug string `json:"addonSlug" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1,max=10"`
}

// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	CustomerInfo     CustomerInfoRequest      `json:"customerInfo" binding:"required"`
	BookingInfo      BookingInfoRequest       `json:"bookingInfo" binding:"required"`
	CategorySlug     string                   `json:"categorySlug" binding:"required"`
	SelectedServices []SelectedServiceRequest `json:"selectedServices" binding:"required,min=1,dive"`
	SelectedAddons   []SelectedAddonRequest   `json:"selectedAddons" binding:"omitempty,dive"`
	SpecialNotes     string                   `json:"specialNotes" binding:"omitempty,max=1000"`
	PaymentMethod    string                   `json:"paymentMethod" binding:"required,oneof=wallet cash"`
}

// Validate performs custom validation on the create order request
func (r *CreateOrderRequest) Validate() error {
	// Validate customer info
	if err := r.CustomerInfo.Validate(); err != nil {
		return fmt.Errorf("customerInfo: %w", err)
	}

	// Validate booking info
	if err := r.BookingInfo.Validate(); err != nil {
		return fmt.Errorf("bookingInfo: %w", err)
	}

	// Validate at least one service is selected
	if len(r.SelectedServices) == 0 {
		return fmt.Errorf("at least one service must be selected")
	}

	// Check for duplicate services
	serviceMap := make(map[string]bool)
	for _, s := range r.SelectedServices {
		if serviceMap[s.ServiceSlug] {
			return fmt.Errorf("duplicate service: %s", s.ServiceSlug)
		}
		serviceMap[s.ServiceSlug] = true
	}

	// Check for duplicate addons
	if len(r.SelectedAddons) > 0 {
		addonMap := make(map[string]bool)
		for _, a := range r.SelectedAddons {
			if addonMap[a.AddonSlug] {
				return fmt.Errorf("duplicate addon: %s", a.AddonSlug)
			}
			addonMap[a.AddonSlug] = true
		}
	}

	return nil
}

// ==================== Cancel Order Request ====================

// CancelOrderRequest represents the request to cancel an order
type CancelOrderRequest struct {
	Reason string `json:"reason" binding:"required,min=10,max=500"`
}

// Validate validates the cancel request
func (r *CancelOrderRequest) Validate() error {
	if len(r.Reason) < 10 {
		return fmt.Errorf("cancellation reason must be at least 10 characters")
	}
	return nil
}

// ==================== Rate Order Request ====================

// RateOrderRequest represents the request to rate a completed order
type RateOrderRequest struct {
	Rating int    `json:"rating" binding:"required,min=1,max=5"`
	Review string `json:"review" binding:"omitempty,max=1000"`
}

// Validate validates the rating request
func (r *RateOrderRequest) Validate() error {
	if r.Rating < 1 || r.Rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}
	return nil
}

// ==================== List Orders Query ====================

// ListOrdersQuery represents query parameters for listing orders
type ListOrdersQuery struct {
	shared.PaginationParams
	Status   string `form:"status" binding:"omitempty"`
	FromDate string `form:"fromDate" binding:"omitempty"` // YYYY-MM-DD
	ToDate   string `form:"toDate" binding:"omitempty"`   // YYYY-MM-DD
	SortBy   string `form:"sortBy" binding:"omitempty,oneof=created_at booking_date status"`
	SortDesc bool   `form:"sortDesc"`
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
	// Validate status if provided
	if q.Status != "" && !shared.IsValidOrderStatus(q.Status) {
		return fmt.Errorf("invalid status: %s", q.Status)
	}

	// Validate date formats if provided
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

	return nil
}
