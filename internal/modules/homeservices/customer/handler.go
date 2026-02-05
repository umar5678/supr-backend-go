package customer

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/umar5678/go-backend/internal/modules/homeservices/customer/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

// NewHandler creates a new customer handler instance
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ==================== Category Handlers ====================

// GetAllCategories godoc
// @Summary Get all service categories
// @Description Get list of all available service categories with counts
// @Tags Home Services - Customer
// @Produce json
// @Success 200 {object} response.Response{data=dto.CategoryListResponse}
// @Failure 500 {object} response.Response
// @Router /homeservices/categories [get]
func (h *Handler) GetAllCategories(c *gin.Context) {
	categories, err := h.service.GetAllCategories(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, categories, "Categories retrieved successfully")
}

// GetCategoryDetail godoc
// @Summary Get category details
// @Description Get detailed information about a category including all services and addons
// @Tags Home Services - Customer
// @Produce json
// @Param categorySlug path string true "Category slug"
// @Success 200 {object} response.Response{data=dto.CategoryDetailResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /homeservices/categories/{categorySlug} [get]
func (h *Handler) GetCategoryDetail(c *gin.Context) {
	categorySlug := c.Param("categorySlug")

	details, err := h.service.GetCategoryDetail(c.Request.Context(), categorySlug)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, details, "Category details retrieved successfully")
}

// ==================== Service Handlers ====================

// ListServices godoc
// @Summary List services
// @Description Get paginated list of available services with filters
// @Tags Home Services - Customer
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Filter by category slug"
// @Param search query string false "Search in title and description"
// @Param minPrice query number false "Minimum price"
// @Param maxPrice query number false "Maximum price"
// @Param isFrequent query bool false "Filter by frequent services only"
// @Param sortBy query string false "Sort by field" Enums(title, price, sort_order, popularity)
// @Param sortDesc query bool false "Sort descending"
// @Success 200 {object} response.Response{data=[]dto.ServiceListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /homeservices/services [get]
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

// GetService godoc
// @Summary Get service details
// @Description Get detailed information about a service including related addons
// @Tags Home Services - Customer
// @Produce json
// @Param slug path string true "Service slug"
// @Success 200 {object} response.Response{data=dto.ServiceDetailResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /homeservices/services/{slug} [get]
func (h *Handler) GetService(c *gin.Context) {
	slug := c.Param("slug")

	service, err := h.service.GetServiceBySlug(c.Request.Context(), slug)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, service, "Service retrieved successfully")
}

// GetFrequentServices godoc
// @Summary Get frequent services
// @Description Get list of frequently booked services
// @Tags Home Services - Customer
// @Produce json
// @Param limit query int false "Number of services to return" default(10)
// @Success 200 {object} response.Response{data=[]dto.ServiceListResponse}
// @Failure 500 {object} response.Response
// @Router /homeservices/services/frequent [get]
func (h *Handler) GetFrequentServices(c *gin.Context) {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	services, err := h.service.GetFrequentServices(c.Request.Context(), limit)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, services, "Frequent services retrieved successfully")
}

// ==================== Addon Handlers ====================

// ListAddons godoc
// @Summary List addons
// @Description Get paginated list of available addons with filters
// @Tags Home Services - Customer
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param category query string false "Filter by category slug"
// @Param search query string false "Search in title and description"
// @Param minPrice query number false "Minimum price"
// @Param maxPrice query number false "Maximum price"
// @Param hasDiscount query bool false "Filter by discounted addons only"
// @Param sortBy query string false "Sort by field" Enums(title, price, sort_order)
// @Param sortDesc query bool false "Sort descending"
// @Success 200 {object} response.Response{data=[]dto.AddonListResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /homeservices/addons [get]
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

// GetAddon godoc
// @Summary Get addon details
// @Description Get detailed information about an addon
// @Tags Home Services - Customer
// @Produce json
// @Param slug path string true "Addon slug"
// @Success 200 {object} response.Response{data=dto.AddonResponse}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /homeservices/addons/{slug} [get]
func (h *Handler) GetAddon(c *gin.Context) {
	slug := c.Param("slug")

	addon, err := h.service.GetAddonBySlug(c.Request.Context(), slug)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, addon, "Addon retrieved successfully")
}

// GetDiscountedAddons godoc
// @Summary Get discounted addons
// @Description Get list of addons with active discounts
// @Tags Home Services - Customer
// @Produce json
// @Param limit query int false "Number of addons to return" default(10)
// @Success 200 {object} response.Response{data=[]dto.AddonListResponse}
// @Failure 500 {object} response.Response
// @Router /homeservices/addons/discounted [get]
func (h *Handler) GetDiscountedAddons(c *gin.Context) {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	addons, err := h.service.GetDiscountedAddons(c.Request.Context(), limit)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, addons, "Discounted addons retrieved successfully")
}

// ==================== Search Handler ====================

// Search godoc
// @Summary Search services and addons
// @Description Search across services and addons by keyword
// @Tags Home Services - Customer
// @Produce json
// @Param q query string true "Search query (min 2 characters)"
// @Param category query string false "Filter by category slug"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Response{data=dto.SearchResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /homeservices/search [get]
func (h *Handler) Search(c *gin.Context) {
	var query dto.SearchQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters: " + err.Error()))
		return
	}

	results, err := h.service.Search(c.Request.Context(), query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, results, "Search completed successfully")
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new service booking order
// @Tags Home Services - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateOrderRequest true "Order details"
// @Success 201 {object} response.Response{data=dto.OrderCreatedResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /homeservices/orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	// Get customer ID from context (set by auth middleware)
	customerID, exists := c.Get("userID")
	if !exists {
		c.Error(response.UnauthorizedError("User not authenticated"))
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), customerID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Order created successfully")
}

// GetOrder godoc
// @Summary Get order details
// @Description Get detailed information about a specific order
// @Tags Home Services - Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response{data=dto.OrderResponse}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /homeservices/orders/{id} [get]
func (h *Handler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	customerID, exists := c.Get("userID")
	if !exists {
		c.Error(response.UnauthorizedError("User not authenticated"))
		return
	}

	order, err := h.service.GetOrder(c.Request.Context(), customerID.(string), orderID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Order retrieved successfully")
}

// ListOrders godoc
// @Summary List customer orders
// @Description Get paginated list of customer's orders with filters
// @Tags Home Services - Orders
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Filter by status"
// @Param fromDate query string false "Filter from date (YYYY-MM-DD)"
// @Param toDate query string false "Filter to date (YYYY-MM-DD)"
// @Param sortBy query string false "Sort by field" Enums(created_at, booking_date, status)
// @Param sortDesc query bool false "Sort descending" default(true)
// @Success 200 {object} response.Response{data=[]dto.OrderListResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /homeservices/orders [get]
func (h *Handler) ListOrders(c *gin.Context) {
	var query dto.ListOrdersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters: " + err.Error()))
		return
	}

	// Default to descending for orders
	if c.Query("sortDesc") == "" {
		query.SortDesc = true
	}

	customerID, exists := c.Get("userID")
	if !exists {
		c.Error(response.UnauthorizedError("User not authenticated"))
		return
	}

	orders, pagination, err := h.service.ListOrders(c.Request.Context(), customerID.(string), query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Paginated(c, orders, *pagination, "Orders retrieved successfully")
}

// GetCancellationPreview godoc
// @Summary Preview order cancellation
// @Description Get cancellation fee preview before cancelling an order
// @Tags Home Services - Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response{data=dto.CancellationPreviewResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /homeservices/orders/{id}/cancel/preview [get]
func (h *Handler) GetCancellationPreview(c *gin.Context) {
	orderID := c.Param("id")

	customerID, exists := c.Get("userID")
	if !exists {
		c.Error(response.UnauthorizedError("User not authenticated"))
		return
	}

	preview, err := h.service.GetCancellationPreview(c.Request.Context(), customerID.(string), orderID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, preview, "Cancellation preview retrieved successfully")
}

// CancelOrder godoc
// @Summary Cancel an order
// @Description Cancel an existing order with a reason
// @Tags Home Services - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body dto.CancelOrderRequest true "Cancellation details"
// @Success 200 {object} response.Response{data=dto.OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /homeservices/orders/{id}/cancel [post]
func (h *Handler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")

	var req dto.CancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	customerID, exists := c.Get("userID")
	if !exists {
		c.Error(response.UnauthorizedError("User not authenticated"))
		return
	}

	order, err := h.service.CancelOrder(c.Request.Context(), customerID.(string), orderID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Order cancelled successfully")
}

// RateOrder godoc
// @Summary Rate a completed order
// @Description Submit rating and review for a completed order
// @Tags Home Services - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body dto.RateOrderRequest true "Rating details"
// @Success 200 {object} response.Response{data=dto.OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /homeservices/orders/{id}/rate [post]
func (h *Handler) RateOrder(c *gin.Context) {
	orderID := c.Param("id")

	var req dto.RateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	customerID, exists := c.Get("userID")
	if !exists {
		c.Error(response.UnauthorizedError("User not authenticated"))
		return
	}

	order, err := h.service.RateOrder(c.Request.Context(), customerID.(string), orderID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Rating submitted successfully")
}
