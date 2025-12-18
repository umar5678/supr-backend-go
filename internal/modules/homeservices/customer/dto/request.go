package dto

import (
	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
)

// ListServicesQuery represents query parameters for listing services
type ListServicesQuery struct {
	shared.PaginationParams
	CategorySlug string   `form:"category"`
	Search       string   `form:"search" binding:"omitempty,max=100"`
	MinPrice     *float64 `form:"minPrice" binding:"omitempty,gte=0"`
	MaxPrice     *float64 `form:"maxPrice" binding:"omitempty,gte=0"`
	IsFrequent   *bool    `form:"isFrequent"`
	SortBy       string   `form:"sortBy" binding:"omitempty,oneof=title price sort_order popularity"`
	SortDesc     bool     `form:"sortDesc"`
}

// SetDefaults sets default values for query parameters
func (q *ListServicesQuery) SetDefaults() {
	q.PaginationParams.SetDefaults()
	if q.SortBy == "" {
		q.SortBy = "sort_order"
	}
}

// ListAddonsQuery represents query parameters for listing addons
type ListAddonsQuery struct {
	shared.PaginationParams
	CategorySlug string   `form:"category"`
	Search       string   `form:"search" binding:"omitempty,max=100"`
	MinPrice     *float64 `form:"minPrice" binding:"omitempty,gte=0"`
	MaxPrice     *float64 `form:"maxPrice" binding:"omitempty,gte=0"`
	HasDiscount  *bool    `form:"hasDiscount"`
	SortBy       string   `form:"sortBy" binding:"omitempty,oneof=title price sort_order"`
	SortDesc     bool     `form:"sortDesc"`
}

// SetDefaults sets default values for query parameters
func (q *ListAddonsQuery) SetDefaults() {
	q.PaginationParams.SetDefaults()
	if q.SortBy == "" {
		q.SortBy = "sort_order"
	}
}

// SearchQuery represents a general search query
type SearchQuery struct {
	shared.PaginationParams
	Query        string `form:"q" binding:"required,min=2,max=100"`
	CategorySlug string `form:"category"`
}

// SetDefaults sets default values
func (q *SearchQuery) SetDefaults() {
	q.PaginationParams.SetDefaults()
}
