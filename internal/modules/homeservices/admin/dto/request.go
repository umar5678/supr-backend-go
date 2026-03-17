package dto

import (
	"fmt"
	"strings"
	"time"

	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
)

type ListOrdersQuery struct {
	shared.PaginationParams

	Status       string `form:"status"`
	CategorySlug string `form:"category"`
	CustomerID   string `form:"customerId"`
	ProviderID   string `form:"providerId"`
	OrderNumber  string `form:"orderNumber"`

	FromDate        string `form:"fromDate"`       
	ToDate          string `form:"toDate"`         
	BookingFromDate string `form:"bookingFromDate"`
	BookingToDate   string `form:"bookingToDate"`  

	MinPrice *float64 `form:"minPrice" binding:"omitempty,gte=0"`
	MaxPrice *float64 `form:"maxPrice" binding:"omitempty,gte=0"`

	SortBy   string `form:"sortBy" binding:"omitempty,oneof=created_at booking_date total_price status completed_at"`
	SortDesc bool   `form:"sortDesc"`

	IncludeCustomer bool `form:"includeCustomer"`
	IncludeProvider bool `form:"includeProvider"`
}

func (q *ListOrdersQuery) SetDefaults() {
	q.PaginationParams.SetDefaults()
	if q.SortBy == "" {
		q.SortBy = "created_at"
	}
}

func (q *ListOrdersQuery) Validate() error {
	if q.Status != "" && !shared.IsValidOrderStatus(q.Status) {
		return fmt.Errorf("invalid status: %s", q.Status)
	}

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

	if q.MinPrice != nil && q.MaxPrice != nil && *q.MinPrice > *q.MaxPrice {
		return fmt.Errorf("minPrice cannot be greater than maxPrice")
	}

	return nil
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Reason string `json:"reason" binding:"omitempty,max=500"`
	Notes  string `json:"notes" binding:"omitempty,max=1000"`
}

func (r *UpdateOrderStatusRequest) Validate() error {
	if !shared.IsValidOrderStatus(r.Status) {
		return fmt.Errorf("invalid status: %s", r.Status)
	}
	return nil
}

type ReassignOrderRequest struct {
	ProviderID string `json:"providerId" binding:"required,uuid"`
	Reason     string `json:"reason" binding:"required,min=10,max=500"`
}

func (r *ReassignOrderRequest) Validate() error {
	if len(r.Reason) < 10 {
		return fmt.Errorf("reason must be at least 10 characters")
	}
	return nil
}

type AdminCancelOrderRequest struct {
	Reason        string   `json:"reason" binding:"required,min=10,max=500"`
	RefundAmount  *float64 `json:"refundAmount" binding:"omitempty,gte=0"` 
	NotifyParties bool     `json:"notifyParties"`
}

func (r *AdminCancelOrderRequest) Validate() error {
	if len(r.Reason) < 10 {
		return fmt.Errorf("reason must be at least 10 characters")
	}
	return nil
}

type AnalyticsQuery struct {
	FromDate string `form:"fromDate" binding:"required"` 
	ToDate   string `form:"toDate" binding:"required"`   
	GroupBy  string `form:"groupBy" binding:"omitempty,oneof=day week month"`
}

func (q *AnalyticsQuery) SetDefaults() {
	if q.GroupBy == "" {
		q.GroupBy = "day"
	}
}

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

	if toDate.Sub(fromDate) > 365*24*time.Hour {
		return fmt.Errorf("date range cannot exceed 1 year")
	}

	return nil
}

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

func (q *ProviderAnalyticsQuery) SetDefaults() {
	if q.SortBy == "" {
		q.SortBy = "completed_orders"
	}
	if q.Limit <= 0 {
		q.Limit = 20
	}
}

type BulkUpdateStatusRequest struct {
	OrderIDs []string `json:"orderIds" binding:"required,min=1,max=100"`
	Status   string   `json:"status" binding:"required"`
	Reason   string   `json:"reason" binding:"required,min=10,max=500"`
}

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

type ExportOrdersQuery struct {
	FromDate     string `form:"fromDate" binding:"required"`
	ToDate       string `form:"toDate" binding:"required"`
	Status       string `form:"status"`
	CategorySlug string `form:"category"`
	Format       string `form:"format" binding:"omitempty,oneof=csv xlsx json"`
}

func (q *ExportOrdersQuery) SetDefaults() {
	if q.Format == "" {
		q.Format = "csv"
	}
}

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

type CreateAddonRequest struct {
	Title              string   `json:"title" binding:"required,min=3,max=255"`
	AddonSlug          string   `json:"addonSlug" binding:"required,min=3,max=255"`
	CategorySlug       string   `json:"categorySlug" binding:"required,min=2,max=255"`
	Description        string   `json:"description" binding:"omitempty,max=2000"`
	WhatsIncluded      []string `json:"whatsIncluded" binding:"omitempty,dive,min=1,max=500"`
	Notes              []string `json:"notes" binding:"omitempty,dive,min=1,max=500"`
	Image              string   `json:"image" binding:"omitempty,url,max=500"`
	Price              float64  `json:"price" binding:"required,gt=0"`
	StrikethroughPrice *float64 `json:"strikethroughPrice" binding:"omitempty,gt=0"`
	IsActive           *bool    `json:"isActive"`
	IsAvailable        *bool    `json:"isAvailable"`
	SortOrder          int      `json:"sortOrder" binding:"omitempty,min=0"`
}

func (r *CreateAddonRequest) Validate() error {
	r.AddonSlug = strings.ToLower(strings.TrimSpace(r.AddonSlug))
	r.CategorySlug = strings.ToLower(strings.TrimSpace(r.CategorySlug))

	if !isValidSlug(r.AddonSlug) {
		return fmt.Errorf("addonSlug must contain only lowercase letters, numbers, and hyphens")
	}

	if !isValidSlug(r.CategorySlug) {
		return fmt.Errorf("categorySlug must contain only lowercase letters, numbers, and hyphens")
	}

	if r.StrikethroughPrice != nil && *r.StrikethroughPrice <= r.Price {
		return fmt.Errorf("strikethroughPrice must be greater than price")
	}

	if r.IsActive == nil {
		defaultActive := true
		r.IsActive = &defaultActive
	}
	if r.IsAvailable == nil {
		defaultAvailable := true
		r.IsAvailable = &defaultAvailable
	}

	return nil
}

type UpdateAddonRequest struct {
	Title              *string  `json:"title" binding:"omitempty,min=3,max=255"`
	CategorySlug       *string  `json:"categorySlug" binding:"omitempty,min=2,max=255"`
	Description        *string  `json:"description" binding:"omitempty,max=2000"`
	WhatsIncluded      []string `json:"whatsIncluded" binding:"omitempty,dive,min=1,max=500"`
	Notes              []string `json:"notes" binding:"omitempty,dive,min=1,max=500"`
	Image              *string  `json:"image" binding:"omitempty,max=500"`
	Price              *float64 `json:"price" binding:"omitempty,gt=0"`
	StrikethroughPrice *float64 `json:"strikethroughPrice" binding:"omitempty,gte=0"`
	IsActive           *bool    `json:"isActive"`
	IsAvailable        *bool    `json:"isAvailable"`
	SortOrder          *int     `json:"sortOrder" binding:"omitempty,min=0"`
}

func (r *UpdateAddonRequest) Validate() error {
	if r.Title == nil && r.CategorySlug == nil && r.Description == nil &&
		r.WhatsIncluded == nil && r.Notes == nil && r.Image == nil &&
		r.Price == nil && r.StrikethroughPrice == nil &&
		r.IsActive == nil && r.IsAvailable == nil && r.SortOrder == nil {
		return fmt.Errorf("at least one field must be provided for update")
	}

	if r.CategorySlug != nil {
		normalized := strings.ToLower(strings.TrimSpace(*r.CategorySlug))
		r.CategorySlug = &normalized
		if !isValidSlug(*r.CategorySlug) {
			return fmt.Errorf("categorySlug must contain only lowercase letters, numbers, and hyphens")
		}
	}

	return nil
}

func (r *UpdateAddonRequest) HasUpdates() bool {
	return r.Title != nil || r.CategorySlug != nil || r.Description != nil ||
		r.WhatsIncluded != nil || r.Notes != nil || r.Image != nil ||
		r.Price != nil || r.StrikethroughPrice != nil ||
		r.IsActive != nil || r.IsAvailable != nil || r.SortOrder != nil
}

type UpdateAddonStatusRequest struct {
	IsActive    *bool `json:"isActive"`
	IsAvailable *bool `json:"isAvailable"`
}

func (r *UpdateAddonStatusRequest) Validate() error {
	if r.IsActive == nil && r.IsAvailable == nil {
		return fmt.Errorf("at least one of isActive or isAvailable must be provided")
	}
	return nil
}

type ListAddonsQuery struct {
	shared.PaginationParams
	CategorySlug string   `form:"categorySlug"`
	Search       string   `form:"search" binding:"omitempty,max=100"`
	IsActive     *bool    `form:"isActive"`
	IsAvailable  *bool    `form:"isAvailable"`
	MinPrice     *float64 `form:"minPrice" binding:"omitempty,gte=0"`
	MaxPrice     *float64 `form:"maxPrice" binding:"omitempty,gte=0"`
	SortBy       string   `form:"sortBy" binding:"omitempty,oneof=title created_at sort_order price"`
	SortDesc     bool     `form:"sortDesc"`
}

func (q *ListAddonsQuery) SetDefaults() {
	q.PaginationParams.SetDefaults()
	if q.SortBy == "" {
		q.SortBy = "sort_order"
	}
}

type CreateServiceRequest struct {
	Title              string   `json:"title" binding:"required,min=3,max=255"`
	LongTitle          string   `json:"longTitle" binding:"omitempty,max=500"`
	ServiceSlug        string   `json:"serviceSlug" binding:"required,min=3,max=255"`
	CategorySlug       string   `json:"categorySlug" binding:"required,min=2,max=255"`
	Description        string   `json:"description" binding:"omitempty,max=2000"`
	LongDescription    string   `json:"longDescription" binding:"omitempty,max=10000"`
	Highlights         string   `json:"highlights" binding:"required,min=3,max=500"`
	WhatsIncluded      []string `json:"whatsIncluded" binding:"required,min=1,dive,min=1,max=500"`
	TermsAndConditions []string `json:"termsAndConditions" binding:"omitempty,dive,min=1,max=1000"`
	BannerImage        string   `json:"bannerImage" binding:"omitempty,url,max=500"`
	Thumbnail          string   `json:"thumbnail" binding:"omitempty,url,max=500"`
	Duration           *int     `json:"duration" binding:"omitempty,min=1,max=10"`
	IsFrequent         bool     `json:"isFrequent"`
	Frequency          string   `json:"frequency" binding:"omitempty,max=100"`
	SortOrder          int      `json:"sortOrder" binding:"omitempty,min=0"`
	IsActive           *bool    `json:"isActive"`
	IsAvailable        *bool    `json:"isAvailable"`
	BasePrice          *float64 `json:"basePrice" binding:"omitempty,gte=0"`
}

type CreateHomeCleaningServiceRequest struct {
	Title        string  `json:"title" binding:"required,min=3,max=255"`
	CategorySlug string  `json:"categorySlug" binding:"required,min=2,max=255"`
	BasePrice    float64 `json:"basePrice" binding:"required,gte=0"`
}

func (r *CreateHomeCleaningServiceRequest) Validate() error {
	r.CategorySlug = strings.ToLower(strings.TrimSpace(r.CategorySlug))
	if !isValidSlug(r.CategorySlug) {
		return fmt.Errorf("categorySlug must contain only lowercase letters, numbers, and hyphens")
	}

	return nil
}

type UpdateHomeCleaningServiceRequest struct {
	Title       *string  `json:"title" binding:"omitempty,min=3,max=255"`
	ServiceSlug *string  `json:"serviceSlug" binding:"omitempty,min=2,max=255"`
	BasePrice   *float64 `json:"basePrice" binding:"omitempty,gte=0"`
}

func (r *UpdateHomeCleaningServiceRequest) HasUpdates() bool {
	return r.Title != nil || r.ServiceSlug != nil || r.BasePrice != nil
}

func (r *UpdateHomeCleaningServiceRequest) Validate() error {
	if r.Title == nil && r.ServiceSlug == nil && r.BasePrice == nil {
		return fmt.Errorf("at least one field must be provided for update")
	}
	if r.ServiceSlug != nil {
		normalized := strings.ToLower(strings.TrimSpace(*r.ServiceSlug))
		r.ServiceSlug = &normalized
		if !isValidSlug(*r.ServiceSlug) {
			return fmt.Errorf("categorySlug must contain only lowercase letters, numbers, and hyphens")
		}
	}

	return nil
}

func (r *CreateServiceRequest) Validate() error {
	r.ServiceSlug = strings.ToLower(strings.TrimSpace(r.ServiceSlug))
	r.CategorySlug = strings.ToLower(strings.TrimSpace(r.CategorySlug))

	if !isValidSlug(r.ServiceSlug) {
		return fmt.Errorf("serviceSlug must contain only lowercase letters, numbers, and hyphens")
	}

	if !isValidSlug(r.CategorySlug) {
		return fmt.Errorf("categorySlug must contain only lowercase letters, numbers, and hyphens")
	}

	if len(r.WhatsIncluded) == 0 {
		return fmt.Errorf("whatsIncluded must have at least one item")
	}

	if r.IsActive == nil {
		defaultActive := true
		r.IsActive = &defaultActive
	}
	if r.IsAvailable == nil {
		defaultAvailable := true
		r.IsAvailable = &defaultAvailable
	}

	return nil
}

type UpdateServiceRequest struct {
	Title              *string  `json:"title" binding:"omitempty,min=3,max=255"`
	LongTitle          *string  `json:"longTitle" binding:"omitempty,max=500"`
	CategorySlug       *string  `json:"categorySlug" binding:"omitempty,min=2,max=255"`
	Description        *string  `json:"description" binding:"omitempty,max=2000"`
	LongDescription    *string  `json:"longDescription" binding:"omitempty,max=10000"`
	Highlights         *string  `json:"highlights" binding:"required,min=3,max=500"`
	WhatsIncluded      []string `json:"whatsIncluded" binding:"omitempty,min=1,dive,min=1,max=500"`
	TermsAndConditions []string `json:"termsAndConditions" binding:"omitempty,dive,min=1,max=1000"`
	BannerImage        *string  `json:"bannerImage" binding:"omitempty,max=500"`
	Thumbnail          *string  `json:"thumbnail" binding:"omitempty,max=500"`
	Duration           *int     `json:"duration" binding:"omitempty,min=1,max=1440"`
	IsFrequent         *bool    `json:"isFrequent"`
	Frequency          *string  `json:"frequency" binding:"omitempty,max=100"`
	SortOrder          *int     `json:"sortOrder" binding:"omitempty,min=0"`
	IsActive           *bool    `json:"isActive"`
	IsAvailable        *bool    `json:"isAvailable"`
	BasePrice          *float64 `json:"basePrice" binding:"omitempty,gte=0"`
}

func (r *UpdateServiceRequest) Validate() error {
	if r.Title == nil && r.LongTitle == nil && r.CategorySlug == nil &&
		r.Description == nil && r.LongDescription == nil &&
		r.Highlights == nil && r.WhatsIncluded == nil &&
		r.TermsAndConditions == nil && r.BannerImage == nil &&
		r.Thumbnail == nil && r.Duration == nil && r.IsFrequent == nil &&
		r.Frequency == nil && r.SortOrder == nil && r.IsActive == nil &&
		r.IsAvailable == nil && r.BasePrice == nil {
		return fmt.Errorf("at least one field must be provided for update")
	}

	if r.CategorySlug != nil {
		normalized := strings.ToLower(strings.TrimSpace(*r.CategorySlug))
		r.CategorySlug = &normalized
		if !isValidSlug(*r.CategorySlug) {
			return fmt.Errorf("categorySlug must contain only lowercase letters, numbers, and hyphens")
		}
	}

	return nil
}

func (r *UpdateServiceRequest) HasUpdates() bool {
	return r.Title != nil || r.LongTitle != nil || r.CategorySlug != nil ||
		r.Description != nil || r.LongDescription != nil ||
		r.Highlights != nil || r.WhatsIncluded != nil ||
		r.TermsAndConditions != nil || r.BannerImage != nil ||
		r.Thumbnail != nil || r.Duration != nil || r.IsFrequent != nil ||
		r.Frequency != nil || r.SortOrder != nil || r.IsActive != nil ||
		r.IsAvailable != nil || r.BasePrice != nil
}

type UpdateServiceStatusRequest struct {
	IsActive    *bool `json:"isActive"`
	IsAvailable *bool `json:"isAvailable"`
}

func (r *UpdateServiceStatusRequest) Validate() error {
	if r.IsActive == nil && r.IsAvailable == nil {
		return fmt.Errorf("at least one of isActive or isAvailable must be provided")
	}
	return nil
}

type ListServicesQuery struct {
	shared.PaginationParams
	CategorySlug string `form:"categorySlug"`
	Search       string `form:"search" binding:"omitempty,max=100"`
	IsActive     *bool  `form:"isActive"`
	IsAvailable  *bool  `form:"isAvailable"`
	SortBy       string `form:"sortBy" binding:"omitempty,oneof=title created_at sort_order base_price"`
	SortDesc     bool   `form:"sortDesc"`
}

func (q *ListServicesQuery) SetDefaults() {
	q.PaginationParams.SetDefaults()
	if q.SortBy == "" {
		q.SortBy = "sort_order"
	}
}

func isValidSlug(slug string) bool {
	if len(slug) == 0 {
		return false
	}
	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	if slug[0] == '-' || slug[len(slug)-1] == '-' {
		return false
	}
	return true
}
