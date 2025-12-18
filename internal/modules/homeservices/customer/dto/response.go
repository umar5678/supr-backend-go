package dto

import (
	"fmt"

	"github.com/umar5678/go-backend/internal/models"
)

// ==================== Category Responses ====================

// CategoryResponse represents a service category for customers
type CategoryResponse struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	Image        string `json:"image"`
	ServiceCount int    `json:"serviceCount"`
	AddonCount   int    `json:"addonCount"`
}

// CategoryListResponse represents a list of categories
type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Total      int                `json:"total"`
}

// CategoryDetailResponse represents full category details with services and addons
type CategoryDetailResponse struct {
	Slug        string            `json:"slug"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Icon        string            `json:"icon"`
	Image       string            `json:"image"`
	Services    []ServiceResponse `json:"services"`
	Addons      []AddonResponse   `json:"addons"`
}

// ==================== Service Responses ====================

// ServiceResponse represents a service for customers (full details)
type ServiceResponse struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	LongTitle          string   `json:"longTitle,omitempty"`
	Slug               string   `json:"slug"`
	CategorySlug       string   `json:"categorySlug"`
	Description        string   `json:"description"`
	LongDescription    string   `json:"longDescription,omitempty"`
	Highlights         string   `json:"highlights"`
	WhatsIncluded      []string `json:"whatsIncluded"`
	TermsAndConditions []string `json:"termsAndConditions,omitempty"`
	BannerImage        string   `json:"bannerImage,omitempty"`
	Thumbnail          string   `json:"thumbnail"`
	Duration           *int     `json:"duration,omitempty"` // in minutes
	DurationText       string   `json:"durationText,omitempty"`
	IsFrequent         bool     `json:"isFrequent"`
	Frequency          string   `json:"frequency,omitempty"`
	BasePrice          *float64 `json:"basePrice,omitempty"`
	FormattedPrice     string   `json:"formattedPrice,omitempty"`
}

// ServiceListResponse represents a service in list view (lighter)
type ServiceListResponse struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Slug           string   `json:"slug"`
	CategorySlug   string   `json:"categorySlug"`
	Description    string   `json:"description"`
	Thumbnail      string   `json:"thumbnail"`
	Duration       *int     `json:"duration,omitempty"`
	DurationText   string   `json:"durationText,omitempty"`
	IsFrequent     bool     `json:"isFrequent"`
	BasePrice      *float64 `json:"basePrice,omitempty"`
	FormattedPrice string   `json:"formattedPrice,omitempty"`
}

// ServiceDetailResponse represents full service details with related addons
type ServiceDetailResponse struct {
	Service ServiceResponse `json:"service"`
	Addons  []AddonResponse `json:"addons"`
}

// ==================== Addon Responses ====================

// AddonResponse represents an addon for customers (full details)
type AddonResponse struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Slug               string   `json:"slug"`
	CategorySlug       string   `json:"categorySlug"`
	Description        string   `json:"description"`
	WhatsIncluded      []string `json:"whatsIncluded"`
	Notes              []string `json:"notes,omitempty"`
	Image              string   `json:"image"`
	Price              float64  `json:"price"`
	FormattedPrice     string   `json:"formattedPrice"`
	StrikethroughPrice *float64 `json:"strikethroughPrice,omitempty"`
	DiscountPercentage float64  `json:"discountPercentage,omitempty"`
	HasDiscount        bool     `json:"hasDiscount"`
}

// AddonListResponse represents an addon in list view (lighter)
type AddonListResponse struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Slug               string   `json:"slug"`
	CategorySlug       string   `json:"categorySlug"`
	Image              string   `json:"image"`
	Price              float64  `json:"price"`
	FormattedPrice     string   `json:"formattedPrice"`
	StrikethroughPrice *float64 `json:"strikethroughPrice,omitempty"`
	DiscountPercentage float64  `json:"discountPercentage,omitempty"`
	HasDiscount        bool     `json:"hasDiscount"`
}

// ==================== Search Response ====================

// SearchResultItem represents a single search result
type SearchResultItem struct {
	Type           string   `json:"type"` // "service" or "addon"
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Slug           string   `json:"slug"`
	CategorySlug   string   `json:"categorySlug"`
	Description    string   `json:"description"`
	Image          string   `json:"image"`
	Price          *float64 `json:"price,omitempty"`
	FormattedPrice string   `json:"formattedPrice,omitempty"`
}

// SearchResponse represents search results
type SearchResponse struct {
	Query   string             `json:"query"`
	Results []SearchResultItem `json:"results"`
	Total   int                `json:"total"`
}

// ==================== Conversion Functions ====================

// FormatDuration formats duration in minutes to human-readable string
func FormatDuration(minutes *int) string {
	if minutes == nil {
		return ""
	}
	m := *minutes
	if m < 60 {
		return fmt.Sprintf("%d min", m)
	}
	hours := m / 60
	mins := m % 60
	if mins == 0 {
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	if hours == 1 {
		return fmt.Sprintf("1 hour %d min", mins)
	}
	return fmt.Sprintf("%d hours %d min", hours, mins)
}

// FormatPrice formats price to string with currency
func FormatPrice(price *float64) string {
	if price == nil {
		return ""
	}
	return fmt.Sprintf("$%.2f", *price)
}

// FormatPriceValue formats a price value to string
func FormatPriceValue(price float64) string {
	return fmt.Sprintf("$%.2f", price)
}

// ToServiceResponse converts model to full service response
func ToServiceResponse(service *models.ServiceNew) ServiceResponse {
	// highlights := make([]string, len(service.Highlights))
	// copy(highlights, service.Highlights)

	whatsIncluded := make([]string, len(service.WhatsIncluded))
	copy(whatsIncluded, service.WhatsIncluded)

	termsAndConditions := make([]string, len(service.TermsAndConditions))
	copy(termsAndConditions, service.TermsAndConditions)

	return ServiceResponse{
		ID:                 service.ID,
		Title:              service.Title,
		LongTitle:          service.LongTitle,
		Slug:               service.ServiceSlug,
		CategorySlug:       service.CategorySlug,
		Description:        service.Description,
		LongDescription:    service.LongDescription,
		Highlights:         service.Highlights,
		WhatsIncluded:      whatsIncluded,
		TermsAndConditions: termsAndConditions,
		BannerImage:        service.BannerImage,
		Thumbnail:          service.Thumbnail,
		Duration:           service.Duration,
		DurationText:       FormatDuration(service.Duration),
		IsFrequent:         service.IsFrequent,
		Frequency:          service.Frequency,
		BasePrice:          service.BasePrice,
		FormattedPrice:     FormatPrice(service.BasePrice),
	}
}

// ToServiceListResponse converts model to list service response
func ToServiceListResponse(service *models.ServiceNew) ServiceListResponse {
	return ServiceListResponse{
		ID:             service.ID,
		Title:          service.Title,
		Slug:           service.ServiceSlug,
		CategorySlug:   service.CategorySlug,
		Description:    service.Description,
		Thumbnail:      service.Thumbnail,
		Duration:       service.Duration,
		DurationText:   FormatDuration(service.Duration),
		IsFrequent:     service.IsFrequent,
		BasePrice:      service.BasePrice,
		FormattedPrice: FormatPrice(service.BasePrice),
	}
}

// ToServiceListResponses converts multiple models to list responses
func ToServiceListResponses(services []*models.ServiceNew) []ServiceListResponse {
	if services == nil {
		return []ServiceListResponse{}
	}
	responses := make([]ServiceListResponse, len(services))
	for i, service := range services {
		responses[i] = ToServiceListResponse(service)
	}
	return responses
}

// ToServiceResponses converts multiple models to full responses
func ToServiceResponses(services []*models.ServiceNew) []ServiceResponse {
	if services == nil {
		return []ServiceResponse{}
	}
	responses := make([]ServiceResponse, len(services))
	for i, service := range services {
		responses[i] = ToServiceResponse(service)
	}
	return responses
}

// ToAddonResponse converts model to full addon response
func ToAddonResponse(addon *models.Addon) AddonResponse {
	whatsIncluded := make([]string, len(addon.WhatsIncluded))
	copy(whatsIncluded, addon.WhatsIncluded)

	notes := make([]string, len(addon.Notes))
	copy(notes, addon.Notes)

	return AddonResponse{
		ID:                 addon.ID,
		Title:              addon.Title,
		Slug:               addon.AddonSlug,
		CategorySlug:       addon.CategorySlug,
		Description:        addon.Description,
		WhatsIncluded:      whatsIncluded,
		Notes:              notes,
		Image:              addon.Image,
		Price:              addon.Price,
		FormattedPrice:     FormatPriceValue(addon.Price),
		StrikethroughPrice: addon.StrikethroughPrice,
		DiscountPercentage: addon.DiscountPercentage(),
		HasDiscount:        addon.HasDiscount(),
	}
}

// ToAddonListResponse converts model to list addon response
func ToAddonListResponse(addon *models.Addon) AddonListResponse {
	return AddonListResponse{
		ID:                 addon.ID,
		Title:              addon.Title,
		Slug:               addon.AddonSlug,
		CategorySlug:       addon.CategorySlug,
		Image:              addon.Image,
		Price:              addon.Price,
		FormattedPrice:     FormatPriceValue(addon.Price),
		StrikethroughPrice: addon.StrikethroughPrice,
		DiscountPercentage: addon.DiscountPercentage(),
		HasDiscount:        addon.HasDiscount(),
	}
}

// ToAddonListResponses converts multiple models to list responses
func ToAddonListResponses(addons []*models.Addon) []AddonListResponse {
	if addons == nil {
		return []AddonListResponse{}
	}
	responses := make([]AddonListResponse, len(addons))
	for i, addon := range addons {
		responses[i] = ToAddonListResponse(addon)
	}
	return responses
}

// ToAddonResponses converts multiple models to full responses
func ToAddonResponses(addons []*models.Addon) []AddonResponse {
	if addons == nil {
		return []AddonResponse{}
	}
	responses := make([]AddonResponse, len(addons))
	for i, addon := range addons {
		responses[i] = ToAddonResponse(addon)
	}
	return responses
}

// ToSearchResultFromService converts service to search result
func ToSearchResultFromService(service *models.ServiceNew) SearchResultItem {
	return SearchResultItem{
		Type:           "service",
		ID:             service.ID,
		Title:          service.Title,
		Slug:           service.ServiceSlug,
		CategorySlug:   service.CategorySlug,
		Description:    service.Description,
		Image:          service.Thumbnail,
		Price:          service.BasePrice,
		FormattedPrice: FormatPrice(service.BasePrice),
	}
}

// ToSearchResultFromAddon converts addon to search result
func ToSearchResultFromAddon(addon *models.Addon) SearchResultItem {
	price := addon.Price
	return SearchResultItem{
		Type:           "addon",
		ID:             addon.ID,
		Title:          addon.Title,
		Slug:           addon.AddonSlug,
		CategorySlug:   addon.CategorySlug,
		Description:    addon.Description,
		Image:          addon.Image,
		Price:          &price,
		FormattedPrice: FormatPriceValue(addon.Price),
	}
}
