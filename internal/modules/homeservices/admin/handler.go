package admin

import (
	"github.com/gin-gonic/gin"

	"github.com/umar5678/go-backend/internal/modules/homeservices/admin/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

// Handler handles HTTP requests for admin home services
type Handler struct {
	service Service
}

// NewHandler creates a new admin handler instance
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ==================== Service Handlers ====================

// CreateService godoc
// @Summary Create a new service
// @Description Create a new home service offering
// @Tags Admin - Home Services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateServiceRequest true "Service details"
// @Success 201 {object} response.Response{data=dto.ServiceResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/homeservices/services [post]
func (h *Handler) CreateService(c *gin.Context) {
	var req dto.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	svc, err := h.service.CreateService(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, svc, "Service created successfully")
}

// GetService godoc
// @Summary Get service by slug
// @Description Get detailed service information
// @Tags Admin - Home Services
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Service slug"
// @Success 200 {object} response.Response{data=dto.ServiceResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/homeservices/services/{slug} [get]
func (h *Handler) GetService(c *gin.Context) {
	slug := c.Param("slug")

	svc, err := h.service.GetServiceBySlug(c.Request.Context(), slug)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, svc, "Service retrieved successfully")
}

// UpdateService godoc
// @Summary Update service
// @Description Update service details
// @Tags Admin - Home Services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Service slug"
// @Param request body dto.UpdateServiceRequest true "Update details"
// @Success 200 {object} response.Response{data=dto.ServiceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/homeservices/services/{slug} [put]
func (h *Handler) UpdateService(c *gin.Context) {
	slug := c.Param("slug")
	var req dto.UpdateServiceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	svc, err := h.service.UpdateService(c.Request.Context(), slug, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, svc, "Service updated successfully")
}

// // UpdateHomeCleaningService godoc
// // @Summary Update HomeCleaning service
// // @Description Update Home Cleaning service details
// // @Tags Admin - Home Services
// // @Accept json
// // @Produce json
// // @Security BearerAuth
// // @Param slug path string true "Service slug"
// // @Param request body dto.UpdateServiceRequest true "Update details"
// // @Success 200 {object} response.Response{data=dto.ServiceResponse}
// // @Failure 400 {object} response.Response
// // @Failure 404 {object} response.Response
// // @Failure 500 {object} response.Response
// // @Router /admin/homeservices/services/{slug} [put]
// func (h *Handler) UpdateHomeCleaningService(c *gin.Context) {
// 	slug := c.Param("slug")
// 	var req dto.UpdateHomeCleaningServiceRequest

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
// 		return
// 	}

// 	svc, err := h.service.UpdateHomeCleaningService(c.Request.Context(), slug, req)
// 	if err != nil {
// 		c.Error(err)
// 		return
// 	}

// 	response.Success(c, svc, "Service updated successfully")
// }

// UpdateServiceStatus godoc
// @Summary Update service status
// @Description Update service active/available status
// @Tags Admin - Home Services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Service slug"
// @Param request body dto.UpdateServiceStatusRequest true "Status update"
// @Success 200 {object} response.Response{data=dto.ServiceResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/homeservices/services/{slug}/status [patch]
func (h *Handler) UpdateServiceStatus(c *gin.Context) {
	slug := c.Param("slug")
	var req dto.UpdateServiceStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	svc, err := h.service.UpdateServiceStatus(c.Request.Context(), slug, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, svc, "Service status updated successfully")
}

// DeleteService godoc
// @Summary Delete service
// @Description Soft delete a service
// @Tags Admin - Home Services
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Service slug"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/homeservices/services/{slug} [delete]
func (h *Handler) DeleteService(c *gin.Context) {
	slug := c.Param("slug")

	if err := h.service.DeleteService(c.Request.Context(), slug); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Service deleted successfully")
}

// ListServices godoc
// @Summary List services
// @Description Get paginated list of services with filters
// @Tags Admin - Home Services
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param categorySlug query string false "Filter by category"
// @Param search query string false "Search in title and description"
// @Param isActive query bool false "Filter by active status"
// @Param isAvailable query bool false "Filter by available status"
// @Param sortBy query string false "Sort by field" Enums(title, created_at, sort_order, base_price)
// @Param sortDesc query bool false "Sort descending"
// @Success 200 {object} response.Response{data=[]dto.ServiceListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/homeservices/services [get]
func (h *Handler) ListServices(c *gin.Context) {
	var query dto.ListServicesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters: " + err.Error()))
		return
	}

	services, pagination, err := h.service.ListServices(c.Request.Context(), query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Paginated(c, services, *pagination, "Services retrieved successfully")
}

// ==================== Addon Handlers ====================

// CreateAddon godoc
// @Summary Create a new addon
// @Description Create a new service addon
// @Tags Admin - Home Services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateAddonRequest true "Addon details"
// @Success 201 {object} response.Response{data=dto.AddonResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/homeservices/addons [post]
func (h *Handler) CreateAddon(c *gin.Context) {
	var req dto.CreateAddonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	addon, err := h.service.CreateAddon(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, addon, "Addon created successfully")
}

// GetAddon godoc
// @Summary Get addon by slug
// @Description Get detailed addon information
// @Tags Admin - Home Services
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Addon slug"
// @Success 200 {object} response.Response{data=dto.AddonResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/homeservices/addons/{slug} [get]
func (h *Handler) GetAddon(c *gin.Context) {
	slug := c.Param("slug")

	addon, err := h.service.GetAddonBySlug(c.Request.Context(), slug)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, addon, "Addon retrieved successfully")
}

// UpdateAddon godoc
// @Summary Update addon
// @Description Update addon details
// @Tags Admin - Home Services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Addon slug"
// @Param request body dto.UpdateAddonRequest true "Update details"
// @Success 200 {object} response.Response{data=dto.AddonResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/homeservices/addons/{slug} [put]
func (h *Handler) UpdateAddon(c *gin.Context) {
	slug := c.Param("slug")
	var req dto.UpdateAddonRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	addon, err := h.service.UpdateAddon(c.Request.Context(), slug, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, addon, "Addon updated successfully")
}

// UpdateAddonStatus godoc
// @Summary Update addon status
// @Description Update addon active/available status
// @Tags Admin - Home Services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Addon slug"
// @Param request body dto.UpdateAddonStatusRequest true "Status update"
// @Success 200 {object} response.Response{data=dto.AddonResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/homeservices/addons/{slug}/status [patch]
func (h *Handler) UpdateAddonStatus(c *gin.Context) {
	slug := c.Param("slug")
	var req dto.UpdateAddonStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	addon, err := h.service.UpdateAddonStatus(c.Request.Context(), slug, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, addon, "Addon status updated successfully")
}

// DeleteAddon godoc
// @Summary Delete addon
// @Description Soft delete an addon
// @Tags Admin - Home Services
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Addon slug"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/homeservices/addons/{slug} [delete]
func (h *Handler) DeleteAddon(c *gin.Context) {
	slug := c.Param("slug")

	if err := h.service.DeleteAddon(c.Request.Context(), slug); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Addon deleted successfully")
}

// ListAddons godoc
// @Summary List addons
// @Description Get paginated list of addons with filters
// @Tags Admin - Home Services
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param categorySlug query string false "Filter by category"
// @Param search query string false "Search in title and description"
// @Param isActive query bool false "Filter by active status"
// @Param isAvailable query bool false "Filter by available status"
// @Param minPrice query number false "Minimum price"
// @Param maxPrice query number false "Maximum price"
// @Param sortBy query string false "Sort by field" Enums(title, created_at, sort_order, price)
// @Param sortDesc query bool false "Sort descending"
// @Success 200 {object} response.Response{data=[]dto.AddonListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/homeservices/addons [get]
func (h *Handler) ListAddons(c *gin.Context) {
	var query dto.ListAddonsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters: " + err.Error()))
		return
	}

	addons, pagination, err := h.service.ListAddons(c.Request.Context(), query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Paginated(c, addons, *pagination, "Addons retrieved successfully")
}

// ==================== Category Handlers ====================

// GetCategoryDetails godoc
// @Summary Get category details
// @Description Get all services and addons for a category
// @Tags Admin - Home Services
// @Produce json
// @Security BearerAuth
// @Param categorySlug path string true "Category slug"
// @Success 200 {object} response.Response{data=dto.CategoryServicesResponse}
// @Failure 500 {object} response.Response
// @Router /admin/homeservices/categories/{categorySlug} [get]
func (h *Handler) GetCategoryDetails(c *gin.Context) {
	categorySlug := c.Param("categorySlug")

	details, err := h.service.GetCategoryDetails(c.Request.Context(), categorySlug)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, details, "Category details retrieved successfully")
}

// GetAllCategories godoc
// @Summary Get all categories
// @Description Get list of all unique category slugs
// @Tags Admin - Home Services
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]string}
// @Failure 500 {object} response.Response
// @Router /admin/homeservices/categories [get]
func (h *Handler) GetAllCategories(c *gin.Context) {
	categories, err := h.service.GetAllCategories(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, categories, "Categories retrieved successfully")
}
