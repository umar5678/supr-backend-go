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

// ==================== Order Management ====================

// GetOrders godoc
// @Summary List all orders
// @Description Get paginated list of all orders with filters
// @Tags Admin - Orders
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Filter by status"
// @Param category query string false "Filter by category"
// @Param customerId query string false "Filter by customer ID"
// @Param providerId query string false "Filter by provider ID"
// @Param orderNumber query string false "Search by order number"
// @Param fromDate query string false "Filter from date (YYYY-MM-DD)"
// @Param toDate query string false "Filter to date (YYYY-MM-DD)"
// @Param bookingFromDate query string false "Filter by booking from date"
// @Param bookingToDate query string false "Filter by booking to date"
// @Param minPrice query number false "Minimum price"
// @Param maxPrice query number false "Maximum price"
// @Param sortBy query string false "Sort by field" Enums(created_at, booking_date, total_price, status, completed_at)
// @Param sortDesc query bool false "Sort descending" default(true)
// @Success 200 {object} response.Response{data=[]dto.AdminOrderListResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /admin/homeservices/orders [get]
func (h *Handler) GetOrders(c *gin.Context) {
	var query dto.ListOrdersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters: " + err.Error()))
		return
	}

	// Default to descending
	if c.Query("sortDesc") == "" {
		query.SortDesc = true
	}

	orders, pagination, err := h.service.GetOrders(c.Request.Context(), query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Paginated(c, orders, *pagination, "Orders retrieved successfully")
}

// GetOrderByID godoc
// @Summary Get order by ID
// @Description Get detailed order information by ID
// @Tags Admin - Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response{data=dto.AdminOrderDetailResponse}
// @Failure 404 {object} response.Response
// @Router /admin/homeservices/orders/{id} [get]
func (h *Handler) GetOrderByID(c *gin.Context) {
	orderID := c.Param("id")

	order, err := h.service.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Order retrieved successfully")
}

// GetOrderByNumber godoc
// @Summary Get order by order number
// @Description Get detailed order information by order number
// @Tags Admin - Orders
// @Produce json
// @Security BearerAuth
// @Param orderNumber path string true "Order Number (e.g., HS-2024-000001)"
// @Success 200 {object} response.Response{data=dto.AdminOrderDetailResponse}
// @Failure 404 {object} response.Response
// @Router /admin/homeservices/orders/number/{orderNumber} [get]
func (h *Handler) GetOrderByNumber(c *gin.Context) {
	orderNumber := c.Param("orderNumber")

	order, err := h.service.GetOrderByNumber(c.Request.Context(), orderNumber)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Order retrieved successfully")
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Update the status of an order
// @Tags Admin - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body dto.UpdateOrderStatusRequest true "Status update"
// @Success 200 {object} response.Response{data=dto.AdminOrderDetailResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /admin/homeservices/orders/{id}/status [patch]
func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	adminID, _ := c.Get("userID")

	var req dto.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	order, err := h.service.UpdateOrderStatus(c.Request.Context(), orderID, req, adminID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Order status updated successfully")
}

// ReassignOrder godoc
// @Summary Reassign order to another provider
// @Description Reassign an order to a different service provider
// @Tags Admin - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body dto.ReassignOrderRequest true "Reassignment details"
// @Success 200 {object} response.Response{data=dto.AdminOrderDetailResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /admin/homeservices/orders/{id}/reassign [post]
func (h *Handler) ReassignOrder(c *gin.Context) {
	orderID := c.Param("id")
	adminID, _ := c.Get("userID")

	var req dto.ReassignOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	order, err := h.service.ReassignOrder(c.Request.Context(), orderID, req, adminID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Order reassigned successfully")
}

// CancelOrder godoc
// @Summary Cancel order (Admin)
// @Description Cancel an order with admin privileges
// @Tags Admin - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body dto.AdminCancelOrderRequest true "Cancellation details"
// @Success 200 {object} response.Response{data=dto.AdminOrderDetailResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /admin/homeservices/orders/{id}/cancel [post]
func (h *Handler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")
	adminID, _ := c.Get("userID")

	var req dto.AdminCancelOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	order, err := h.service.CancelOrder(c.Request.Context(), orderID, req, adminID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Order cancelled successfully")
}

// GetOrderHistory godoc
// @Summary Get order status history
// @Description Get the complete status change history for an order
// @Tags Admin - Orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response{data=dto.AdminOrderDetailResponse}
// @Failure 404 {object} response.Response
// @Router /admin/homeservices/orders/{id}/history [get]
func (h *Handler) GetOrderHistory(c *gin.Context) {
	orderID := c.Param("id")

	// GetOrderByID already includes history
	order, err := h.service.GetOrderByID(c.Request.Context(), orderID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order.StatusHistory, "Order history retrieved successfully")
}

// ==================== Bulk Operations ====================

// BulkUpdateStatus godoc
// @Summary Bulk update order status
// @Description Update status of multiple orders at once
// @Tags Admin - Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.BulkUpdateStatusRequest true "Bulk update request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /admin/homeservices/orders/bulk/status [post]
func (h *Handler) BulkUpdateStatus(c *gin.Context) {
	adminID, _ := c.Get("userID")

	var req dto.BulkUpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	affected, err := h.service.BulkUpdateStatus(c.Request.Context(), req, adminID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, map[string]interface{}{
		"affected":  affected,
		"requested": len(req.OrderIDs),
	}, "Bulk update completed successfully")
}

// ==================== Analytics ====================

// GetOverviewAnalytics godoc
// @Summary Get overview analytics
// @Description Get comprehensive analytics overview for orders
// @Tags Admin - Analytics
// @Produce json
// @Security BearerAuth
// @Param fromDate query string true "From date (YYYY-MM-DD)"
// @Param toDate query string true "To date (YYYY-MM-DD)"
// @Param groupBy query string false "Group by" Enums(day, week, month)
// @Success 200 {object} response.Response{data=dto.OverviewAnalyticsResponse}
// @Failure 400 {object} response.Response
// @Router /admin/homeservices/analytics/overview [get]
func (h *Handler) GetOverviewAnalytics(c *gin.Context) {
	var query dto.AnalyticsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters: " + err.Error()))
		return
	}

	analytics, err := h.service.GetOverviewAnalytics(c.Request.Context(), query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, analytics, "Analytics retrieved successfully")
}

// GetProviderAnalytics godoc
// @Summary Get provider analytics
// @Description Get analytics for service providers
// @Tags Admin - Analytics
// @Produce json
// @Security BearerAuth
// @Param fromDate query string true "From date (YYYY-MM-DD)"
// @Param toDate query string true "To date (YYYY-MM-DD)"
// @Param category query string false "Filter by category"
// @Param minOrders query int false "Minimum completed orders"
// @Param minRating query number false "Minimum rating"
// @Param sortBy query string false "Sort by" Enums(completed_orders, earnings, rating)
// @Param sortDesc query bool false "Sort descending" default(true)
// @Param limit query int false "Number of providers" default(20)
// @Success 200 {object} response.Response{data=dto.ProviderAnalyticsResponse}
// @Failure 400 {object} response.Response
// @Router /admin/homeservices/analytics/providers [get]
func (h *Handler) GetProviderAnalytics(c *gin.Context) {
	var query dto.ProviderAnalyticsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters: " + err.Error()))
		return
	}

	if c.Query("sortDesc") == "" {
		query.SortDesc = true
	}

	analytics, err := h.service.GetProviderAnalytics(c.Request.Context(), query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, analytics, "Provider analytics retrieved successfully")
}

// GetRevenueReport godoc
// @Summary Get revenue report
// @Description Get detailed revenue report
// @Tags Admin - Analytics
// @Produce json
// @Security BearerAuth
// @Param fromDate query string true "From date (YYYY-MM-DD)"
// @Param toDate query string true "To date (YYYY-MM-DD)"
// @Param groupBy query string false "Group by" Enums(day, week, month)
// @Success 200 {object} response.Response{data=dto.RevenueReportResponse}
// @Failure 400 {object} response.Response
// @Router /admin/homeservices/analytics/revenue [get]
func (h *Handler) GetRevenueReport(c *gin.Context) {
	var query dto.AnalyticsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters: " + err.Error()))
		return
	}

	report, err := h.service.GetRevenueReport(c.Request.Context(), query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, report, "Revenue report retrieved successfully")
}

// ==================== Dashboard ====================

// GetDashboard godoc
// @Summary Get admin dashboard
// @Description Get dashboard overview with key metrics and recent activity
// @Tags Admin - Dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dto.DashboardResponse}
// @Router /admin/homeservices/dashboard [get]
func (h *Handler) GetDashboard(c *gin.Context) {
	dashboard, err := h.service.GetDashboard(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, dashboard, "Dashboard retrieved successfully")
}
