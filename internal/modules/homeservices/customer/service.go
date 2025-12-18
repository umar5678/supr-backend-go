package customer

import (
	"context"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/modules/homeservices/customer/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

// CategoryConfig holds static category configuration
// In production, this could come from database or config file
type CategoryConfig struct {
	Slug        string
	Title       string
	Description string
	Icon        string
	Image       string
	SortOrder   int
}

// GetCategoryConfigs returns category configurations
// TODO: Move to database or configuration file in production
func GetCategoryConfigs() map[string]CategoryConfig {
	return map[string]CategoryConfig{
		"pest-control": {
			Slug:        "pest-control",
			Title:       "Pest Control",
			Description: "Professional pest control services for your home and office",
			Icon:        "pest-icon",
			Image:       "https://example.com/images/pest-control.jpg",
			SortOrder:   1,
		},
		"cleaning": {
			Slug:        "cleaning",
			Title:       "Cleaning Services",
			Description: "Professional cleaning services for homes and offices",
			Icon:        "cleaning-icon",
			Image:       "https://example.com/images/cleaning.jpg",
			SortOrder:   2,
		},
		"iv-therapy": {
			Slug:        "iv-therapy",
			Title:       "IV Therapy",
			Description: "Professional IV therapy and wellness treatments",
			Icon:        "iv-icon",
			Image:       "https://example.com/images/iv-therapy.jpg",
			SortOrder:   3,
		},
		"massage": {
			Slug:        "massage",
			Title:       "Massage Therapy",
			Description: "Professional massage and relaxation services",
			Icon:        "massage-icon",
			Image:       "https://example.com/images/massage.jpg",
			SortOrder:   4,
		},
		"handyman": {
			Slug:        "handyman",
			Title:       "Handyman Services",
			Description: "Professional handyman services for repairs and maintenance",
			Icon:        "handyman-icon",
			Image:       "https://example.com/images/handyman.jpg",
			SortOrder:   5,
		},
	}
}

// Service defines the interface for customer home services business logic
type Service interface {
	// Category operations
	GetAllCategories(ctx context.Context) (*dto.CategoryListResponse, error)
	GetCategoryDetail(ctx context.Context, categorySlug string) (*dto.CategoryDetailResponse, error)

	// Service operations
	GetServiceBySlug(ctx context.Context, slug string) (*dto.ServiceDetailResponse, error)
	ListServices(ctx context.Context, query dto.ListServicesQuery) ([]dto.ServiceListResponse, *response.PaginationMeta, error)
	GetFrequentServices(ctx context.Context, limit int) ([]dto.ServiceListResponse, error)

	// Addon operations
	GetAddonBySlug(ctx context.Context, slug string) (*dto.AddonResponse, error)
	ListAddons(ctx context.Context, query dto.ListAddonsQuery) ([]dto.AddonListResponse, *response.PaginationMeta, error)
	GetDiscountedAddons(ctx context.Context, limit int) ([]dto.AddonListResponse, error)

	// Search operations
	Search(ctx context.Context, query dto.SearchQuery) (*dto.SearchResponse, error)
}

type service struct {
	repo Repository
}

// NewService creates a new customer service instance
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ==================== Category Operations ====================

func (s *service) GetAllCategories(ctx context.Context) (*dto.CategoryListResponse, error) {
	// Get category counts from database
	categoryInfos, err := s.repo.GetAllActiveCategories(ctx)
	if err != nil {
		logger.Error("failed to get categories", "error", err)
		return nil, response.InternalServerError("Failed to get categories", err)
	}

	// Get static category configs
	configs := GetCategoryConfigs()

	// Build response
	var categories []dto.CategoryResponse
	for _, info := range categoryInfos {
		config, exists := configs[info.Slug]
		if !exists {
			// Use default values if config not found
			config = CategoryConfig{
				Slug:  info.Slug,
				Title: info.Slug, // Fallback to slug as title
			}
		}

		categories = append(categories, dto.CategoryResponse{
			Slug:         info.Slug,
			Title:        config.Title,
			Description:  config.Description,
			Icon:         config.Icon,
			Image:        config.Image,
			ServiceCount: int(info.ServiceCount),
			AddonCount:   int(info.AddonCount),
		})
	}

	return &dto.CategoryListResponse{
		Categories: categories,
		Total:      len(categories),
	}, nil
}

func (s *service) GetCategoryDetail(ctx context.Context, categorySlug string) (*dto.CategoryDetailResponse, error) {
	// Get services for this category
	services, err := s.repo.GetActiveServicesByCategory(ctx, categorySlug)
	if err != nil {
		logger.Error("failed to get services by category", "error", err, "category", categorySlug)
		return nil, response.InternalServerError("Failed to get category details", err)
	}

	// Get addons for this category
	addons, err := s.repo.GetActiveAddonsByCategory(ctx, categorySlug)
	if err != nil {
		logger.Error("failed to get addons by category", "error", err, "category", categorySlug)
		return nil, response.InternalServerError("Failed to get category details", err)
	}

	// Check if category exists (has any services or addons)
	if len(services) == 0 && len(addons) == 0 {
		return nil, response.NotFoundError("Category")
	}

	// Get category config
	configs := GetCategoryConfigs()
	config, exists := configs[categorySlug]
	if !exists {
		config = CategoryConfig{
			Slug:  categorySlug,
			Title: categorySlug,
		}
	}

	return &dto.CategoryDetailResponse{
		Slug:        categorySlug,
		Title:       config.Title,
		Description: config.Description,
		Icon:        config.Icon,
		Image:       config.Image,
		Services:    dto.ToServiceResponses(services),
		Addons:      dto.ToAddonResponses(addons),
	}, nil
}

// ==================== Service Operations ====================

func (s *service) GetServiceBySlug(ctx context.Context, slug string) (*dto.ServiceDetailResponse, error) {
	// Get service
	svc, err := s.repo.GetActiveServiceBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Service")
		}
		logger.Error("failed to get service", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to get service", err)
	}

	// Get related addons (same category)
	addons, err := s.repo.GetActiveAddonsByCategory(ctx, svc.CategorySlug)
	if err != nil {
		logger.Error("failed to get related addons", "error", err, "category", svc.CategorySlug)
		// Don't fail the request, just return empty addons
		addons = []*models.Addon{}
	}

	return &dto.ServiceDetailResponse{
		Service: dto.ToServiceResponse(svc),
		Addons:  dto.ToAddonResponses(addons),
	}, nil
}

func (s *service) ListServices(ctx context.Context, query dto.ListServicesQuery) ([]dto.ServiceListResponse, *response.PaginationMeta, error) {
	// Set defaults
	query.SetDefaults()

	// Get services
	services, total, err := s.repo.ListActiveServices(ctx, query)
	if err != nil {
		logger.Error("failed to list services", "error", err)
		return nil, nil, response.InternalServerError("Failed to list services", err)
	}

	// Convert to response
	responses := dto.ToServiceListResponses(services)

	// Create pagination metadata
	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

func (s *service) GetFrequentServices(ctx context.Context, limit int) ([]dto.ServiceListResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	services, err := s.repo.GetFrequentServices(ctx, limit)
	if err != nil {
		logger.Error("failed to get frequent services", "error", err)
		return nil, response.InternalServerError("Failed to get frequent services", err)
	}

	return dto.ToServiceListResponses(services), nil
}

// ==================== Addon Operations ====================

func (s *service) GetAddonBySlug(ctx context.Context, slug string) (*dto.AddonResponse, error) {
	addon, err := s.repo.GetActiveAddonBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Addon")
		}
		logger.Error("failed to get addon", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to get addon", err)
	}

	result := dto.ToAddonResponse(addon)
	return &result, nil
}

func (s *service) ListAddons(ctx context.Context, query dto.ListAddonsQuery) ([]dto.AddonListResponse, *response.PaginationMeta, error) {
	// Set defaults
	query.SetDefaults()

	// Get addons
	addons, total, err := s.repo.ListActiveAddons(ctx, query)
	if err != nil {
		logger.Error("failed to list addons", "error", err)
		return nil, nil, response.InternalServerError("Failed to list addons", err)
	}

	// Convert to response
	responses := dto.ToAddonListResponses(addons)

	// Create pagination metadata
	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

func (s *service) GetDiscountedAddons(ctx context.Context, limit int) ([]dto.AddonListResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	addons, err := s.repo.GetDiscountedAddons(ctx, limit)
	if err != nil {
		logger.Error("failed to get discounted addons", "error", err)
		return nil, response.InternalServerError("Failed to get discounted addons", err)
	}

	return dto.ToAddonListResponses(addons), nil
}

// ==================== Search Operations ====================

func (s *service) Search(ctx context.Context, query dto.SearchQuery) (*dto.SearchResponse, error) {
	query.SetDefaults()

	// Limit results per type
	limitPerType := query.Limit / 2
	if limitPerType < 5 {
		limitPerType = 5
	}

	// Search services
	services, err := s.repo.SearchServices(ctx, query.Query, query.CategorySlug, limitPerType)
	if err != nil {
		logger.Error("failed to search services", "error", err, "query", query.Query)
		return nil, response.InternalServerError("Failed to search", err)
	}

	// Search addons
	addons, err := s.repo.SearchAddons(ctx, query.Query, query.CategorySlug, limitPerType)
	if err != nil {
		logger.Error("failed to search addons", "error", err, "query", query.Query)
		return nil, response.InternalServerError("Failed to search", err)
	}

	// Build combined results
	var results []dto.SearchResultItem

	// Add services to results
	for _, svc := range services {
		results = append(results, dto.ToSearchResultFromService(svc))
	}

	// Add addons to results
	for _, addon := range addons {
		results = append(results, dto.ToSearchResultFromAddon(addon))
	}

	return &dto.SearchResponse{
		Query:   query.Query,
		Results: results,
		Total:   len(results),
	}, nil
}
