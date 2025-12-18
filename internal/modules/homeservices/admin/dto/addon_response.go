package dto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
)

// AddonResponse represents a full addon response
type AddonResponse struct {
	ID                 string    `json:"id"`
	Title              string    `json:"title"`
	AddonSlug          string    `json:"addonSlug"`
	CategorySlug       string    `json:"categorySlug"`
	Description        string    `json:"description"`
	WhatsIncluded      []string  `json:"whatsIncluded"`
	Notes              []string  `json:"notes"`
	Image              string    `json:"image"`
	Price              float64   `json:"price"`
	StrikethroughPrice *float64  `json:"strikethroughPrice"`
	DiscountPercentage float64   `json:"discountPercentage"`
	IsActive           bool      `json:"isActive"`
	IsAvailable        bool      `json:"isAvailable"`
	SortOrder          int       `json:"sortOrder"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// AddonListResponse represents an addon in list view (lighter)
type AddonListResponse struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	AddonSlug          string   `json:"addonSlug"`
	CategorySlug       string   `json:"categorySlug"`
	Image              string   `json:"image"`
	Price              float64  `json:"price"`
	StrikethroughPrice *float64 `json:"strikethroughPrice"`
	DiscountPercentage float64  `json:"discountPercentage"`
	IsActive           bool     `json:"isActive"`
	IsAvailable        bool     `json:"isAvailable"`
	SortOrder          int      `json:"sortOrder"`
}

// ToAddonResponse converts model to full response
func ToAddonResponse(addon *models.Addon) *AddonResponse {
	if addon == nil {
		return nil
	}

	// Convert pq.StringArray to []string
	whatsIncluded := make([]string, len(addon.WhatsIncluded))
	copy(whatsIncluded, addon.WhatsIncluded)

	notes := make([]string, len(addon.Notes))
	copy(notes, addon.Notes)

	return &AddonResponse{
		ID:                 addon.ID,
		Title:              addon.Title,
		AddonSlug:          addon.AddonSlug,
		CategorySlug:       addon.CategorySlug,
		Description:        addon.Description,
		WhatsIncluded:      whatsIncluded,
		Notes:              notes,
		Image:              addon.Image,
		Price:              addon.Price,
		StrikethroughPrice: addon.StrikethroughPrice,
		DiscountPercentage: addon.DiscountPercentage(),
		IsActive:           addon.IsActive,
		IsAvailable:        addon.IsAvailable,
		SortOrder:          addon.SortOrder,
		CreatedAt:          addon.CreatedAt,
		UpdatedAt:          addon.UpdatedAt,
	}
}

// ToAddonListResponse converts model to list response
func ToAddonListResponse(addon *models.Addon) *AddonListResponse {
	if addon == nil {
		return nil
	}

	return &AddonListResponse{
		ID:                 addon.ID,
		Title:              addon.Title,
		AddonSlug:          addon.AddonSlug,
		CategorySlug:       addon.CategorySlug,
		Image:              addon.Image,
		Price:              addon.Price,
		StrikethroughPrice: addon.StrikethroughPrice,
		DiscountPercentage: addon.DiscountPercentage(),
		IsActive:           addon.IsActive,
		IsAvailable:        addon.IsAvailable,
		SortOrder:          addon.SortOrder,
	}
}

// ToAddonListResponses converts multiple models to list responses
func ToAddonListResponses(addons []*models.Addon) []*AddonListResponse {
	if addons == nil {
		return []*AddonListResponse{}
	}

	responses := make([]*AddonListResponse, len(addons))
	for i, addon := range addons {
		responses[i] = ToAddonListResponse(addon)
	}
	return responses
}
