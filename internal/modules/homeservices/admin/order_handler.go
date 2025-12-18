package admin

import (
	"github.com/gin-gonic/gin"

	"github.com/umar5678/go-backend/internal/modules/homeservices/admin/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

// OrderHandler handles HTTP requests for admin order management
type OrderHandler struct {
	service OrderService
}

// NewOrderHandler creates a new admin order handler
func NewOrderHandler(service OrderService) *OrderHandler {
	return &OrderHandler{service: service}
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
func (h *OrderHandler) GetOrders(c *gin.Context) {
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
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
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
func (h *OrderHandler) GetOrderByNumber(c *gin.Context) {
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
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
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
func (h *OrderHandler) ReassignOrder(c *gin.Context) {
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
func (h *OrderHandler) CancelOrder(c *gin.Context) {
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
func (h *OrderHandler) GetOrderHistory(c *gin.Context) {
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
func (h *OrderHandler) BulkUpdateStatus(c *gin.Context) {
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
func (h *OrderHandler) GetOverviewAnalytics(c *gin.Context) {
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
func (h *OrderHandler) GetProviderAnalytics(c *gin.Context) {
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
func (h *OrderHandler) GetRevenueReport(c *gin.Context) {
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
func (h *OrderHandler) GetDashboard(c *gin.Context) {
	dashboard, err := h.service.GetDashboard(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, dashboard, "Dashboard retrieved successfully")
}
