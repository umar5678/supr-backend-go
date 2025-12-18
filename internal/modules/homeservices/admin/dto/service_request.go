package dto

import (
	"fmt"
	"strings"

	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
)

// CreateServiceRequest represents the request to create a new service
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
	Duration           *int     `json:"duration" binding:"omitempty,min=1,max=1440"` // max 24 hours in minutes
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

// Validate performs custom validation
func (r *CreateHomeCleaningServiceRequest) Validate() error {
	// Normalize category slug
	r.CategorySlug = strings.ToLower(strings.TrimSpace(r.CategorySlug))
	// Validate slug format (alphanumeric and hyphens only)
	if !isValidSlug(r.CategorySlug) {
		return fmt.Errorf("categorySlug must contain only lowercase letters, numbers, and hyphens")
	}

	return nil
}

// UpdateHomeCleaningServiceRequest represents the request to update a home cleaning service
type UpdateHomeCleaningServiceRequest struct {
	Title       *string  `json:"title" binding:"omitempty,min=3,max=255"`
	ServiceSlug *string  `json:"serviceSlug" binding:"omitempty,min=2,max=255"`
	BasePrice   *float64 `json:"basePrice" binding:"omitempty,gte=0"`
}

// HasUpdates checks if there are any updates
func (r *UpdateHomeCleaningServiceRequest) HasUpdates() bool {
	return r.Title != nil || r.ServiceSlug != nil || r.BasePrice != nil
}

// Validate performs custom validation
func (r *UpdateHomeCleaningServiceRequest) Validate() error {
	// Check if at least one field is provided
	if r.Title == nil && r.ServiceSlug == nil && r.BasePrice == nil {
		return fmt.Errorf("at least one field must be provided for update")
	}
	// Normalize Service slug if provided
	if r.ServiceSlug != nil {
		normalized := strings.ToLower(strings.TrimSpace(*r.ServiceSlug))
		r.ServiceSlug = &normalized
		if !isValidSlug(*r.ServiceSlug) {
			return fmt.Errorf("categorySlug must contain only lowercase letters, numbers, and hyphens")
		}
	}

	return nil
}

// Validate performs custom validation
func (r *CreateServiceRequest) Validate() error {
	// Normalize slug
	r.ServiceSlug = strings.ToLower(strings.TrimSpace(r.ServiceSlug))
	r.CategorySlug = strings.ToLower(strings.TrimSpace(r.CategorySlug))

	// Validate slug format (alphanumeric and hyphens only)
	if !isValidSlug(r.ServiceSlug) {
		return fmt.Errorf("serviceSlug must contain only lowercase letters, numbers, and hyphens")
	}

	if !isValidSlug(r.CategorySlug) {
		return fmt.Errorf("categorySlug must contain only lowercase letters, numbers, and hyphens")
	}

	// Validate WhatsIncluded has at least one item
	if len(r.WhatsIncluded) == 0 {
		return fmt.Errorf("whatsIncluded must have at least one item")
	}

	// Set defaults
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

// UpdateServiceRequest represents the request to update a service
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

// Validate performs custom validation
func (r *UpdateServiceRequest) Validate() error {
	// Check if at least one field is provided
	if r.Title == nil && r.LongTitle == nil && r.CategorySlug == nil &&
		r.Description == nil && r.LongDescription == nil &&
		r.Highlights == nil && r.WhatsIncluded == nil &&
		r.TermsAndConditions == nil && r.BannerImage == nil &&
		r.Thumbnail == nil && r.Duration == nil && r.IsFrequent == nil &&
		r.Frequency == nil && r.SortOrder == nil && r.IsActive == nil &&
		r.IsAvailable == nil && r.BasePrice == nil {
		return fmt.Errorf("at least one field must be provided for update")
	}

	// Normalize category slug if provided
	if r.CategorySlug != nil {
		normalized := strings.ToLower(strings.TrimSpace(*r.CategorySlug))
		r.CategorySlug = &normalized
		if !isValidSlug(*r.CategorySlug) {
			return fmt.Errorf("categorySlug must contain only lowercase letters, numbers, and hyphens")
		}
	}

	return nil
}

// HasUpdates checks if there are any updates
func (r *UpdateServiceRequest) HasUpdates() bool {
	return r.Title != nil || r.LongTitle != nil || r.CategorySlug != nil ||
		r.Description != nil || r.LongDescription != nil ||
		r.Highlights != nil || r.WhatsIncluded != nil ||
		r.TermsAndConditions != nil || r.BannerImage != nil ||
		r.Thumbnail != nil || r.Duration != nil || r.IsFrequent != nil ||
		r.Frequency != nil || r.SortOrder != nil || r.IsActive != nil ||
		r.IsAvailable != nil || r.BasePrice != nil
}

// UpdateServiceStatusRequest represents status update request
type UpdateServiceStatusRequest struct {
	IsActive    *bool `json:"isActive"`
	IsAvailable *bool `json:"isAvailable"`
}

// Validate performs validation
func (r *UpdateServiceStatusRequest) Validate() error {
	if r.IsActive == nil && r.IsAvailable == nil {
		return fmt.Errorf("at least one of isActive or isAvailable must be provided")
	}
	return nil
}

// ListServicesQuery represents query parameters for listing services
type ListServicesQuery struct {
	shared.PaginationParams
	CategorySlug string `form:"categorySlug"`
	Search       string `form:"search" binding:"omitempty,max=100"`
	IsActive     *bool  `form:"isActive"`
	IsAvailable  *bool  `form:"isAvailable"`
	SortBy       string `form:"sortBy" binding:"omitempty,oneof=title created_at sort_order base_price"`
	SortDesc     bool   `form:"sortDesc"`
}

// SetDefaults sets default values
func (q *ListServicesQuery) SetDefaults() {
	q.PaginationParams.SetDefaults()
	if q.SortBy == "" {
		q.SortBy = "sort_order"
	}
}

// Helper function to validate slug format
func isValidSlug(slug string) bool {
	if len(slug) == 0 {
		return false
	}
	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	// Should not start or end with hyphen
	if slug[0] == '-' || slug[len(slug)-1] == '-' {
		return false
	}
	return true
}
