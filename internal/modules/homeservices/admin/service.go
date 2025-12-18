package admin

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/admin/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

// Service defines the interface for admin home services business logic
type Service interface {
	// Service operations
	CreateService(ctx context.Context, req dto.CreateServiceRequest) (*dto.ServiceResponse, error)
	GetServiceBySlug(ctx context.Context, slug string) (*dto.ServiceResponse, error)
	UpdateService(ctx context.Context, slug string, req dto.UpdateServiceRequest) (*dto.ServiceResponse, error)
	// UpdateHomeCleaningService(ctx context.Context, slug string, req dto.UpdateHomeCleaningServiceRequest) (*dto.HomeCleaningServiceResponse, error)
	UpdateServiceStatus(ctx context.Context, slug string, req dto.UpdateServiceStatusRequest) (*dto.ServiceResponse, error)
	DeleteService(ctx context.Context, slug string) error
	ListServices(ctx context.Context, query dto.ListServicesQuery) ([]*dto.ServiceListResponse, *response.PaginationMeta, error)

	// Addon operations
	CreateAddon(ctx context.Context, req dto.CreateAddonRequest) (*dto.AddonResponse, error)
	GetAddonBySlug(ctx context.Context, slug string) (*dto.AddonResponse, error)
	UpdateAddon(ctx context.Context, slug string, req dto.UpdateAddonRequest) (*dto.AddonResponse, error)
	UpdateAddonStatus(ctx context.Context, slug string, req dto.UpdateAddonStatusRequest) (*dto.AddonResponse, error)
	DeleteAddon(ctx context.Context, slug string) error
	ListAddons(ctx context.Context, query dto.ListAddonsQuery) ([]*dto.AddonListResponse, *response.PaginationMeta, error)

	// Category operations
	GetCategoryDetails(ctx context.Context, categorySlug string) (*dto.CategoryServicesResponse, error)
	GetAllCategories(ctx context.Context) ([]string, error)
}

type service struct {
	repo Repository
}

// NewService creates a new admin service instance
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// ==================== Service Operations ====================

func (s *service) CreateService(ctx context.Context, req dto.CreateServiceRequest) (*dto.ServiceResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Check if slug already exists
	exists, err := s.repo.ServiceSlugExists(ctx, req.ServiceSlug, "")
	if err != nil {
		logger.Error("failed to check service slug existence", "error", err, "slug", req.ServiceSlug)
		return nil, response.InternalServerError("Failed to create service", err)
	}
	if exists {
		return nil, response.ConflictError(fmt.Sprintf("Service with slug '%s' already exists", req.ServiceSlug))
	}

	// Create service model
	svc := &models.ServiceNew{
		Title:              req.Title,
		LongTitle:          req.LongTitle,
		ServiceSlug:        req.ServiceSlug,
		CategorySlug:       req.CategorySlug,
		Description:        req.Description,
		LongDescription:    req.LongDescription,
		Highlights:         req.Highlights,
		WhatsIncluded:      pq.StringArray(req.WhatsIncluded),
		TermsAndConditions: pq.StringArray(req.TermsAndConditions),
		BannerImage:        req.BannerImage,
		Thumbnail:          req.Thumbnail,
		Duration:           req.Duration,
		IsFrequent:         req.IsFrequent,
		Frequency:          req.Frequency,
		SortOrder:          req.SortOrder,
		IsActive:           *req.IsActive,
		IsAvailable:        *req.IsAvailable,
		BasePrice:          req.BasePrice,
	}

	// Save to database
	if err := s.repo.CreateService(ctx, svc); err != nil {
		logger.Error("failed to create service", "error", err, "slug", req.ServiceSlug)
		return nil, response.InternalServerError("Failed to create service", err)
	}

	logger.Info("service created", "serviceID", svc.ID, "slug", svc.ServiceSlug)

	return dto.ToServiceResponse(svc), nil
}

func (s *service) GetServiceBySlug(ctx context.Context, slug string) (*dto.ServiceResponse, error) {
	svc, err := s.repo.GetServiceBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Service")
		}
		logger.Error("failed to get service", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to get service", err)
	}

	return dto.ToServiceResponse(svc), nil
}

func (s *service) UpdateService(ctx context.Context, slug string, req dto.UpdateServiceRequest) (*dto.ServiceResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Get existing service
	svc, err := s.repo.GetServiceBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Service")
		}
		return nil, response.InternalServerError("Failed to get service", err)
	}

	// Update fields (only if provided)
	if req.Title != nil {
		svc.Title = *req.Title
	}
	if req.LongTitle != nil {
		svc.LongTitle = *req.LongTitle
	}
	if req.CategorySlug != nil {
		svc.CategorySlug = *req.CategorySlug
	}
	if req.Description != nil {
		svc.Description = *req.Description
	}
	if req.LongDescription != nil {
		svc.LongDescription = *req.LongDescription
	}
	if req.Highlights != nil {
		svc.Highlights = *req.Highlights
	}
	if req.WhatsIncluded != nil {
		svc.WhatsIncluded = pq.StringArray(req.WhatsIncluded)
	}
	if req.TermsAndConditions != nil {
		svc.TermsAndConditions = pq.StringArray(req.TermsAndConditions)
	}
	if req.BannerImage != nil {
		svc.BannerImage = *req.BannerImage
	}
	if req.Thumbnail != nil {
		svc.Thumbnail = *req.Thumbnail
	}
	if req.Duration != nil {
		svc.Duration = req.Duration
	}
	if req.IsFrequent != nil {
		svc.IsFrequent = *req.IsFrequent
	}
	if req.Frequency != nil {
		svc.Frequency = *req.Frequency
	}
	if req.SortOrder != nil {
		svc.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		svc.IsActive = *req.IsActive
	}
	if req.IsAvailable != nil {
		svc.IsAvailable = *req.IsAvailable
	}
	if req.BasePrice != nil {
		svc.BasePrice = req.BasePrice
	}

	// Save to database
	if err := s.repo.UpdateService(ctx, svc); err != nil {
		logger.Error("failed to update service", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to update service", err)
	}

	logger.Info("service updated", "serviceID", svc.ID, "slug", svc.ServiceSlug)

	return dto.ToServiceResponse(svc), nil
}

// func (s *service) UpdateHomeCleaningService(ctx context.Context, slug string, req dto.UpdateHomeCleaningServiceRequest) (*dto.HomeCleaningServiceResponse, error) {
// 	// Validate request
// 	if err := req.Validate(); err != nil {
// 		return nil, response.BadRequest(err.Error())
// 	}

// 	// Get existing service
// 	svc, err := s.repo.GetServiceBySlug(ctx, slug)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, response.NotFoundError("Service")
// 		}
// 		return nil, response.InternalServerError("Failed to get service", err)
// 	}

// 	// Map existing ServiceNew to HomeCleaningService model before updating
// 	hc := &models.HomeCleaningService{
// 		ID:          svc.ID,
// 		Title:       svc.Title,
// 		ServiceSlug: svc.ServiceSlug,
// 	}
// 	if svc.BasePrice != nil {
// 		hc.BasePrice = *svc.BasePrice
// 	}

// 	// Update fields (only if provided)
// 	if req.Title != nil {
// 		hc.Title = *req.Title
// 	}
// 	if req.ServiceSlug != nil {
// 		hc.ServiceSlug = *req.ServiceSlug
// 	}
// 	if req.BasePrice != nil {
// 		hc.BasePrice = *req.BasePrice
// 	}

// 	// Save to database
// 	if err := s.repo.UpdateHomeCleaningService(ctx, hc); err != nil {
// 		logger.Error("failed to update service", "error", err, "slug", slug)
// 		return nil, response.InternalServerError("Failed to update service", err)
// 	}

// 	logger.Info("service updated", "serviceID", hc.ID, "slug", hc.ServiceSlug)

// 	return dto.ToHomeCleaningServiceResponse(hc), nil
// }

func (s *service) UpdateServiceStatus(ctx context.Context, slug string, req dto.UpdateServiceStatusRequest) (*dto.ServiceResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Get existing service
	svc, err := s.repo.GetServiceBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Service")
		}
		return nil, response.InternalServerError("Failed to get service", err)
	}

	// Update status fields
	if req.IsActive != nil {
		svc.IsActive = *req.IsActive
	}
	if req.IsAvailable != nil {
		svc.IsAvailable = *req.IsAvailable
	}

	// Save to database
	if err := s.repo.UpdateService(ctx, svc); err != nil {
		logger.Error("failed to update service status", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to update service status", err)
	}

	logger.Info("service status updated", "serviceID", svc.ID, "slug", svc.ServiceSlug,
		"isActive", svc.IsActive, "isAvailable", svc.IsAvailable)

	return dto.ToServiceResponse(svc), nil
}

func (s *service) DeleteService(ctx context.Context, slug string) error {
	// Check if service exists
	svc, err := s.repo.GetServiceBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.NotFoundError("Service")
		}
		return response.InternalServerError("Failed to get service", err)
	}

	// Soft delete
	if err := s.repo.DeleteService(ctx, svc.ID); err != nil {
		logger.Error("failed to delete service", "error", err, "slug", slug)
		return response.InternalServerError("Failed to delete service", err)
	}

	logger.Info("service deleted", "serviceID", svc.ID, "slug", slug)

	return nil
}

func (s *service) ListServices(ctx context.Context, query dto.ListServicesQuery) ([]*dto.ServiceListResponse, *response.PaginationMeta, error) {
	// Set defaults
	query.SetDefaults()

	// Get services from repository
	services, total, err := s.repo.ListServices(ctx, query)
	if err != nil {
		logger.Error("failed to list services", "error", err)
		return nil, nil, response.InternalServerError("Failed to list services", err)
	}

	// Convert to response DTOs
	responses := dto.ToServiceListResponses(services)

	// Create pagination metadata
	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

// ==================== Addon Operations ====================

func (s *service) CreateAddon(ctx context.Context, req dto.CreateAddonRequest) (*dto.AddonResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Check if slug already exists
	exists, err := s.repo.AddonSlugExists(ctx, req.AddonSlug, "")
	if err != nil {
		logger.Error("failed to check addon slug existence", "error", err, "slug", req.AddonSlug)
		return nil, response.InternalServerError("Failed to create addon", err)
	}
	if exists {
		return nil, response.ConflictError(fmt.Sprintf("Addon with slug '%s' already exists", req.AddonSlug))
	}

	// Create addon model
	addon := &models.Addon{
		Title:              req.Title,
		AddonSlug:          req.AddonSlug,
		CategorySlug:       req.CategorySlug,
		Description:        req.Description,
		WhatsIncluded:      pq.StringArray(req.WhatsIncluded),
		Notes:              pq.StringArray(req.Notes),
		Image:              req.Image,
		Price:              req.Price,
		StrikethroughPrice: req.StrikethroughPrice,
		IsActive:           *req.IsActive,
		IsAvailable:        *req.IsAvailable,
		SortOrder:          req.SortOrder,
	}

	// Save to database
	if err := s.repo.CreateAddon(ctx, addon); err != nil {
		logger.Error("failed to create addon", "error", err, "slug", req.AddonSlug)
		return nil, response.InternalServerError("Failed to create addon", err)
	}

	logger.Info("addon created", "addonID", addon.ID, "slug", addon.AddonSlug)

	return dto.ToAddonResponse(addon), nil
}

func (s *service) GetAddonBySlug(ctx context.Context, slug string) (*dto.AddonResponse, error) {
	addon, err := s.repo.GetAddonBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Addon")
		}
		logger.Error("failed to get addon", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to get addon", err)
	}

	return dto.ToAddonResponse(addon), nil
}

func (s *service) UpdateAddon(ctx context.Context, slug string, req dto.UpdateAddonRequest) (*dto.AddonResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Get existing addon
	addon, err := s.repo.GetAddonBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Addon")
		}
		return nil, response.InternalServerError("Failed to get addon", err)
	}

	// Update fields (only if provided)
	if req.Title != nil {
		addon.Title = *req.Title
	}
	if req.CategorySlug != nil {
		addon.CategorySlug = *req.CategorySlug
	}
	if req.Description != nil {
		addon.Description = *req.Description
	}
	if req.WhatsIncluded != nil {
		addon.WhatsIncluded = pq.StringArray(req.WhatsIncluded)
	}
	if req.Notes != nil {
		addon.Notes = pq.StringArray(req.Notes)
	}
	if req.Image != nil {
		addon.Image = *req.Image
	}
	if req.Price != nil {
		addon.Price = *req.Price
	}
	if req.StrikethroughPrice != nil {
		// Allow setting to 0 to remove strikethrough price
		if *req.StrikethroughPrice == 0 {
			addon.StrikethroughPrice = nil
		} else {
			addon.StrikethroughPrice = req.StrikethroughPrice
		}
	}
	if req.IsActive != nil {
		addon.IsActive = *req.IsActive
	}
	if req.IsAvailable != nil {
		addon.IsAvailable = *req.IsAvailable
	}
	if req.SortOrder != nil {
		addon.SortOrder = *req.SortOrder
	}

	// Validate strikethrough price is greater than price
	if addon.StrikethroughPrice != nil && *addon.StrikethroughPrice <= addon.Price {
		return nil, response.BadRequest("strikethroughPrice must be greater than price")
	}

	// Save to database
	if err := s.repo.UpdateAddon(ctx, addon); err != nil {
		logger.Error("failed to update addon", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to update addon", err)
	}

	logger.Info("addon updated", "addonID", addon.ID, "slug", addon.AddonSlug)

	return dto.ToAddonResponse(addon), nil
}

func (s *service) UpdateAddonStatus(ctx context.Context, slug string, req dto.UpdateAddonStatusRequest) (*dto.AddonResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Get existing addon
	addon, err := s.repo.GetAddonBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Addon")
		}
		return nil, response.InternalServerError("Failed to get addon", err)
	}

	// Update status fields
	if req.IsActive != nil {
		addon.IsActive = *req.IsActive
	}
	if req.IsAvailable != nil {
		addon.IsAvailable = *req.IsAvailable
	}

	// Save to database
	if err := s.repo.UpdateAddon(ctx, addon); err != nil {
		logger.Error("failed to update addon status", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to update addon status", err)
	}

	logger.Info("addon status updated", "addonID", addon.ID, "slug", addon.AddonSlug,
		"isActive", addon.IsActive, "isAvailable", addon.IsAvailable)

	return dto.ToAddonResponse(addon), nil
}

func (s *service) DeleteAddon(ctx context.Context, slug string) error {
	// Check if addon exists
	addon, err := s.repo.GetAddonBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.NotFoundError("Addon")
		}
		return response.InternalServerError("Failed to get addon", err)
	}

	// Soft delete
	if err := s.repo.DeleteAddon(ctx, addon.ID); err != nil {
		logger.Error("failed to delete addon", "error", err, "slug", slug)
		return response.InternalServerError("Failed to delete addon", err)
	}

	logger.Info("addon deleted", "addonID", addon.ID, "slug", slug)

	return nil
}

func (s *service) ListAddons(ctx context.Context, query dto.ListAddonsQuery) ([]*dto.AddonListResponse, *response.PaginationMeta, error) {
	// Set defaults
	query.SetDefaults()

	// Get addons from repository
	addons, total, err := s.repo.ListAddons(ctx, query)
	if err != nil {
		logger.Error("failed to list addons", "error", err)
		return nil, nil, response.InternalServerError("Failed to list addons", err)
	}

	// Convert to response DTOs
	responses := dto.ToAddonListResponses(addons)

	// Create pagination metadata
	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

// ==================== Category Operations ====================

func (s *service) GetCategoryDetails(ctx context.Context, categorySlug string) (*dto.CategoryServicesResponse, error) {
	// Get services for category
	services, err := s.repo.GetServicesByCategory(ctx, categorySlug)
	if err != nil {
		logger.Error("failed to get services by category", "error", err, "category", categorySlug)
		return nil, response.InternalServerError("Failed to get category details", err)
	}

	// Get addons for category
	addons, err := s.repo.GetAddonsByCategory(ctx, categorySlug)
	if err != nil {
		logger.Error("failed to get addons by category", "error", err, "category", categorySlug)
		return nil, response.InternalServerError("Failed to get category details", err)
	}

	return &dto.CategoryServicesResponse{
		CategorySlug: categorySlug,
		Services:     dto.ToServiceListResponses(services),
		Addons:       dto.ToAddonListResponses(addons),
		TotalCount:   len(services) + len(addons),
	}, nil
}

func (s *service) GetAllCategories(ctx context.Context) ([]string, error) {
	categories, err := s.repo.GetAllCategories(ctx)
	if err != nil {
		logger.Error("failed to get all categories", "error", err)
		return nil, response.InternalServerError("Failed to get categories", err)
	}
	return categories, nil
}
