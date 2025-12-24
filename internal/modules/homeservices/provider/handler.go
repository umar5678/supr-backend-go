package provider

import (
	"github.com/gin-gonic/gin"

   	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/provider/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

// Handler handles HTTP requests for providers
type Handler struct {
	service Service
}

// NewHandler creates a new provider handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// getProviderIDFromContext extracts userID and converts it to providerID
// Returns empty string if user is service_provider role but doesn't have profile yet (during registration)
func (h *Handler) getProviderIDFromContext(c *gin.Context) (string, error) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	providerID, err := h.service.GetProviderIDByUserID(c.Request.Context(), userID.(string))
	if err != nil {
		// Check if user is a service provider (allow during registration process)
		if userRole == string(models.RoleServiceProvider) {
			// User is a service provider but doesn't have a profile yet
			// This is OK during registration - return empty string to indicate "not yet registered"
			return "", nil
		}
		// User is not a service provider, deny access
		return "", response.UnauthorizedError("You must be a service provider to access this endpoint")
	}
	return providerID, nil
}

// ==================== Profile Handlers ====================

// GetProfile godoc
// @Summary Get provider profile
// @Description Get current provider's profile with statistics
// @Tags Provider - Profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.ProviderProfileResponse}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /provider/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	profile, err := h.service.GetProfile(c.Request.Context(), providerID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, profile, "Profile retrieved successfully")
}

// UpdateAvailability godoc
// @Summary Update availability status
// @Description Update provider's availability status and location
// @Tags Provider - Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateAvailabilityRequest true "Availability update"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /provider/availability [patch]
func (h *Handler) UpdateAvailability(c *gin.Context) {
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req dto.UpdateAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	if err := h.service.UpdateAvailability(c.Request.Context(), providerID, req); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Availability updated successfully")
}

// ==================== Service Category Handlers ====================

// GetServiceCategories godoc
// @Summary Get service categories
// @Description Get provider's registered service categories. Returns empty list if provider is still in registration process.
// @Tags Provider - Categories
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]dto.ServiceCategoryResponse}
// @Failure 401 {object} response.Response
// @Router /provider/categories [get]
func (h *Handler) GetServiceCategories(c *gin.Context) {
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	// If providerID is empty, user is service provider but not yet registered
	if providerID == "" {
		response.Success(c, []dto.ServiceCategoryResponse{}, "No categories yet - user is in registration process")
		return
	}

	categories, err := h.service.GetServiceCategories(c.Request.Context(), providerID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, categories, "Categories retrieved successfully")
}

// AddServiceCategory godoc
// @Summary Add service category
// @Description Add a new service category to provider's profile. If provider doesn't exist yet, creates the profile during first registration.
// @Tags Provider - Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.AddServiceCategoryRequest true "Category details"
// @Success 201 {object} response.Response{data=dto.ServiceCategoryResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 409 {object} response.Response
// @Router /provider/categories [post]
func (h *Handler) AddServiceCategory(c *gin.Context) {
	userID, _ := c.Get("userID")
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	var req dto.AddServiceCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	// If providerID is empty, create the provider profile first
	if providerID == "" {
		// Create provider profile on first category registration
		// This is part of the registration flow
		providerID = h.service.CreateProviderOnFirstCategory(c.Request.Context(), userID.(string))
	}

	category, err := h.service.AddServiceCategory(c.Request.Context(), providerID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, category, "Category added successfully", "PROVIDER_CATEGORY_ADDED")
}

// UpdateServiceCategory godoc
// @Summary Update service category
// @Description Update a service category
// @Tags Provider - Categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param categorySlug path string true "Category slug"
// @Param request body dto.UpdateServiceCategoryRequest true "Update details"
// @Success 200 {object} response.Response{data=dto.ServiceCategoryResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /provider/categories/{categorySlug} [put]
func (h *Handler) UpdateServiceCategory(c *gin.Context) {
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}
	categorySlug := c.Param("categorySlug")

	var req dto.UpdateServiceCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	category, err := h.service.UpdateServiceCategory(c.Request.Context(), providerID, categorySlug, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, category, "Category updated successfully")
}

// DeleteServiceCategory godoc
// @Summary Delete service category
// @Description Remove a service category from provider's profile
// @Tags Provider - Categories
// @Produce json
// @Security BearerAuth
// @Param categorySlug path string true "Category slug"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /provider/categories/{categorySlug} [delete]
func (h *Handler) DeleteServiceCategory(c *gin.Context) {
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}
	categorySlug := c.Param("categorySlug")

	if err := h.service.DeleteServiceCategory(c.Request.Context(), providerID, categorySlug); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Category deleted successfully")
}

// ==================== Available Orders Handlers ====================

// GetAvailableOrders godoc
// @Summary Get available orders
// @Description Get orders available for the provider to accept
// @Tags Provider - Orders
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Filter by category"
// @Param date query string false "Filter by booking date (YYYY-MM-DD)"
// @Param sortBy query string false "Sort by" Enums(created_at, booking_date, distance, price)
// @Param sortDesc query bool false "Sort descending"
// @Success 200 {object} response.Response{data=[]dto.AvailableOrderResponse}
// @Failure 401 {object} response.Response
// @Router /provider/orders/available [get]
func (h *Handler) GetAvailableOrders(c *gin.Context) {
	userID, _ := c.Get("userID")

	// Convert user ID to provider ID
	providerID, err := h.service.GetProviderIDByUserID(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	var query dto.ListAvailableOrdersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters: " + err.Error()))
		return
	}

	orders, pagination, err := h.service.GetAvailableOrders(c.Request.Context(), providerID, query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Paginated(c, orders, *pagination, "Available orders retrieved successfully")
}

// GetAvailableOrderDetail godoc
// @Summary Get available order detail
// @Description Get details of an available order
// @Tags Provider - Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response{data=dto.AvailableOrderResponse}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /provider/orders/available/{id} [get]
func (h *Handler) GetAvailableOrderDetail(c *gin.Context) {
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}
	orderID := c.Param("id")

	order, err := h.service.GetAvailableOrderDetail(c.Request.Context(), providerID, orderID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Order retrieved successfully")
}

// ==================== My Orders Handlers ====================

// GetMyOrders godoc
// @Summary Get my orders
// @Description Get provider's assigned/completed orders
// @Tags Provider - Orders
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Filter by status"
// @Param fromDate query string false "From date (YYYY-MM-DD)"
// @Param toDate query string false "To date (YYYY-MM-DD)"
// @Param sortBy query string false "Sort by" Enums(created_at, booking_date, completed_at)
// @Param sortDesc query bool false "Sort descending" default(true)
// @Success 200 {object} response.Response{data=[]dto.ProviderOrderListResponse}
// @Failure 401 {object} response.Response
// @Router /provider/orders [get]
func (h *Handler) GetMyOrders(c *gin.Context) {
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	var query dto.ListMyOrdersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters: " + err.Error()))
		return
	}

	if c.Query("sortDesc") == "" {
		query.SortDesc = true
	}

	orders, pagination, err := h.service.GetMyOrders(c.Request.Context(), providerID, query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Paginated(c, orders, *pagination, "Orders retrieved successfully")
}

// GetMyOrderDetail godoc
// @Summary Get my order detail
// @Description Get details of a provider's order
// @Tags Provider - Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response{data=dto.ProviderOrderResponse}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /provider/orders/{id} [get]
func (h *Handler) GetMyOrderDetail(c *gin.Context) {
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}
	orderID := c.Param("id")

	order, err := h.service.GetMyOrderDetail(c.Request.Context(), providerID, orderID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Order retrieved successfully")
}

// ==================== Order Actions Handlers ====================

// AcceptOrder godoc
// @Summary Accept an order
// @Description Accept an available order
// @Tags Provider - Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response{data=dto.ProviderOrderResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /provider/orders/{id}/accept [post]
func (h *Handler) AcceptOrder(c *gin.Context) {
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}
	orderID := c.Param("id")

	order, err := h.service.AcceptOrder(c.Request.Context(), providerID, orderID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Order accepted successfully")
}

// RejectOrder godoc
// @Summary Reject an order
// @Description Reject an assigned order
// @Tags Provider - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body dto.RejectOrderRequest true "Rejection reason"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /provider/orders/{id}/reject [post]
func (h *Handler) RejectOrder(c *gin.Context) {
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}
	orderID := c.Param("id")

	var req dto.RejectOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	if err := h.service.RejectOrder(c.Request.Context(), providerID, orderID, req); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Order rejected successfully")
}

// StartOrder godoc
// @Summary Start an order
// @Description Mark an accepted order as in progress
// @Tags Provider - Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response{data=dto.ProviderOrderResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /provider/orders/{id}/start [post]
func (h *Handler) StartOrder(c *gin.Context) {
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}
	orderID := c.Param("id")

	order, err := h.service.StartOrder(c.Request.Context(), providerID, orderID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Order started successfully")
}

// CompleteOrder godoc
// @Summary Complete an order
// @Description Mark an in-progress order as completed
// @Tags Provider - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body dto.CompleteOrderRequest false "Completion notes"
// @Success 200 {object} response.Response{data=dto.ProviderOrderResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /provider/orders/{id}/complete [post]
func (h *Handler) CompleteOrder(c *gin.Context) {
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}
	orderID := c.Param("id")

	var req dto.CompleteOrderRequest
	c.ShouldBindJSON(&req) // Optional body

	order, err := h.service.CompleteOrder(c.Request.Context(), providerID, orderID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Order completed successfully")
}

// RateCustomer godoc
// @Summary Rate a customer
// @Description Submit rating for a completed order's customer
// @Tags Provider - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body dto.RateCustomerRequest true "Rating details"
// @Success 200 {object} response.Response{data=dto.ProviderOrderResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /provider/orders/{id}/rate [post]
func (h *Handler) RateCustomer(c *gin.Context) {
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}
	orderID := c.Param("id")

	var req dto.RateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	order, err := h.service.RateCustomer(c.Request.Context(), providerID, orderID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Rating submitted successfully")
}

// ==================== Statistics Handlers ====================

// GetStatistics godoc
// @Summary Get statistics
// @Description Get provider's performance statistics
// @Tags Provider - Statistics
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.ProviderStatistics}
// @Failure 401 {object} response.Response
// @Router /provider/statistics [get]
func (h *Handler) GetStatistics(c *gin.Context) {
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	stats, err := h.service.GetStatistics(c.Request.Context(), providerID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, stats, "Statistics retrieved successfully")
}

// GetEarnings godoc
// @Summary Get earnings
// @Description Get provider's earnings for a date range
// @Tags Provider - Statistics
// @Produce json
// @Security BearerAuth
// @Param fromDate query string true "From date (YYYY-MM-DD)"
// @Param toDate query string true "To date (YYYY-MM-DD)"
// @Param groupBy query string false "Group by" Enums(day, week, month)
// @Success 200 {object} response.Response{data=dto.EarningsSummaryResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /provider/earnings [get]
func (h *Handler) GetEarnings(c *gin.Context) {
	providerID, err := h.getProviderIDFromContext(c)
	if err != nil {
		c.Error(err)
		return
	}

	var query dto.EarningsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters: " + err.Error()))
		return
	}

	earnings, err := h.service.GetEarnings(c.Request.Context(), providerID, query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, earnings, "Earnings retrieved successfully")
}
