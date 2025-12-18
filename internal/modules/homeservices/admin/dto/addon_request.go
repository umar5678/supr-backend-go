package dto

import (
	"fmt"
	"strings"

	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
)

// CreateAddonRequest represents the request to create a new addon
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

// Validate performs custom validation
func (r *CreateAddonRequest) Validate() error {
	// Normalize slugs
	r.AddonSlug = strings.ToLower(strings.TrimSpace(r.AddonSlug))
	r.CategorySlug = strings.ToLower(strings.TrimSpace(r.CategorySlug))

	// Validate slug format
	if !isValidSlug(r.AddonSlug) {
		return fmt.Errorf("addonSlug must contain only lowercase letters, numbers, and hyphens")
	}

	if !isValidSlug(r.CategorySlug) {
		return fmt.Errorf("categorySlug must contain only lowercase letters, numbers, and hyphens")
	}

	// Validate strikethrough price is greater than price
	if r.StrikethroughPrice != nil && *r.StrikethroughPrice <= r.Price {
		return fmt.Errorf("strikethroughPrice must be greater than price")
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

// UpdateAddonRequest represents the request to update an addon
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

// Validate performs custom validation
func (r *UpdateAddonRequest) Validate() error {
	// Check if at least one field is provided
	if r.Title == nil && r.CategorySlug == nil && r.Description == nil &&
		r.WhatsIncluded == nil && r.Notes == nil && r.Image == nil &&
		r.Price == nil && r.StrikethroughPrice == nil &&
		r.IsActive == nil && r.IsAvailable == nil && r.SortOrder == nil {
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
func (r *UpdateAddonRequest) HasUpdates() bool {
	return r.Title != nil || r.CategorySlug != nil || r.Description != nil ||
		r.WhatsIncluded != nil || r.Notes != nil || r.Image != nil ||
		r.Price != nil || r.StrikethroughPrice != nil ||
		r.IsActive != nil || r.IsAvailable != nil || r.SortOrder != nil
}

// UpdateAddonStatusRequest represents status update request
type UpdateAddonStatusRequest struct {
	IsActive    *bool `json:"isActive"`
	IsAvailable *bool `json:"isAvailable"`
}

// Validate performs validation
func (r *UpdateAddonStatusRequest) Validate() error {
	if r.IsActive == nil && r.IsAvailable == nil {
		return fmt.Errorf("at least one of isActive or isAvailable must be provided")
	}
	return nil
}

// ListAddonsQuery represents query parameters for listing addons
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

// SetDefaults sets default values
func (q *ListAddonsQuery) SetDefaults() {
	q.PaginationParams.SetDefaults()
	if q.SortBy == "" {
		q.SortBy = "sort_order"
	}
}
