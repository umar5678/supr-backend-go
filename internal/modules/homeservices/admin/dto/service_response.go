package dto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
)

// ServiceResponse represents a full service response
type ServiceResponse struct {
	ID                 string    `json:"id"`
	Title              string    `json:"title"`
	LongTitle          string    `json:"longTitle"`
	ServiceSlug        string    `json:"serviceSlug"`
	CategorySlug       string    `json:"categorySlug"`
	Description        string    `json:"description"`
	LongDescription    string    `json:"longDescription"`
	Highlights         string    `json:"highlights"`
	WhatsIncluded      []string  `json:"whatsIncluded"`
	TermsAndConditions []string  `json:"termsAndConditions"`
	BannerImage        string    `json:"bannerImage"`
	Thumbnail          string    `json:"thumbnail"`
	Duration           *int      `json:"duration"`
	IsFrequent         bool      `json:"isFrequent"`
	Frequency          string    `json:"frequency"`
	SortOrder          int       `json:"sortOrder"`
	IsActive           bool      `json:"isActive"`
	IsAvailable        bool      `json:"isAvailable"`
	BasePrice          *float64  `json:"basePrice"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// // HomeCleaningServiceResponse represents a home cleaning service response
// type HomeCleaningServiceResponse struct {
// 	ID          string    `json:"id"`
// 	Title       string    `json:"title"`
// 	ServiceSlug string    `json:"serviceSlug"`
// 	BasePrice   float64   `json:"basePrice"`
// 	CreatedAt   time.Time `json:"createdAt"`
// }

// // ToHomeCleaningServiceResponse converts model to response
// func ToHomeCleaningServiceResponse(service *models.HomeCleaningService) *HomeCleaningServiceResponse {
// 	if service == nil {
// 		return nil
// 	}
// 	return &HomeCleaningServiceResponse{
// 		ID:          service.ID,
// 		Title:       service.Title,
// 		ServiceSlug: service.ServiceSlug,
// 		BasePrice:   service.BasePrice,
// 		CreatedAt:   service.CreatedAt,
// 	}
// }

// // HomeCleaningServiceAddonResponse represents an addon in service response
// type HomeCleaningServiceAddonResponse struct {
// 	CategorySlug string               `json:"categorySlug"`
// 	Addons       []*AddonListResponse `json:"addons"`
// 	TotalCount   int                  `json:"totalCount"`
// }

// ServiceListResponse represents a service in list view (lighter)
type ServiceListResponse struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	ServiceSlug  string    `json:"serviceSlug"`
	CategorySlug string    `json:"categorySlug"`
	Thumbnail    string    `json:"thumbnail"`
	Duration     *int      `json:"duration"`
	IsActive     bool      `json:"isActive"`
	IsAvailable  bool      `json:"isAvailable"`
	BasePrice    *float64  `json:"basePrice"`
	SortOrder    int       `json:"sortOrder"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ToServiceResponse converts model to full response
func ToServiceResponse(service *models.ServiceNew) *ServiceResponse {
	if service == nil {
		return nil
	}

	// // Convert pq.StringArray to []string
	// highlights := make([]string, len(service.Highlights))
	// copy(highlights, service.Highlights)

	whatsIncluded := make([]string, len(service.WhatsIncluded))
	copy(whatsIncluded, service.WhatsIncluded)

	termsAndConditions := make([]string, len(service.TermsAndConditions))
	copy(termsAndConditions, service.TermsAndConditions)

	return &ServiceResponse{
		ID:                 service.ID,
		Title:              service.Title,
		LongTitle:          service.LongTitle,
		ServiceSlug:        service.ServiceSlug,
		CategorySlug:       service.CategorySlug,
		Description:        service.Description,
		LongDescription:    service.LongDescription,
		Highlights:         service.Highlights,
		WhatsIncluded:      whatsIncluded,
		TermsAndConditions: termsAndConditions,
		BannerImage:        service.BannerImage,
		Thumbnail:          service.Thumbnail,
		Duration:           service.Duration,
		IsFrequent:         service.IsFrequent,
		Frequency:          service.Frequency,
		SortOrder:          service.SortOrder,
		IsActive:           service.IsActive,
		IsAvailable:        service.IsAvailable,
		BasePrice:          service.BasePrice,
		CreatedAt:          service.CreatedAt,
		UpdatedAt:          service.UpdatedAt,
	}
}

// ToServiceListResponse converts model to list response
func ToServiceListResponse(service *models.ServiceNew) *ServiceListResponse {
	if service == nil {
		return nil
	}

	return &ServiceListResponse{
		ID:           service.ID,
		Title:        service.Title,
		ServiceSlug:  service.ServiceSlug,
		CategorySlug: service.CategorySlug,
		Thumbnail:    service.Thumbnail,
		Duration:     service.Duration,
		IsActive:     service.IsActive,
		IsAvailable:  service.IsAvailable,
		BasePrice:    service.BasePrice,
		SortOrder:    service.SortOrder,
		CreatedAt:    service.CreatedAt,
	}
}

// ToServiceListResponses converts multiple models to list responses
func ToServiceListResponses(services []*models.ServiceNew) []*ServiceListResponse {
	if services == nil {
		return []*ServiceListResponse{}
	}

	responses := make([]*ServiceListResponse, len(services))
	for i, service := range services {
		responses[i] = ToServiceListResponse(service)
	}
	return responses
}

// CategoryServicesResponse represents services grouped by category
type CategoryServicesResponse struct {
	CategorySlug string                 `json:"categorySlug"`
	Services     []*ServiceListResponse `json:"services"`
	Addons       []*AddonListResponse   `json:"addons"`
	TotalCount   int                    `json:"totalCount"`
}
