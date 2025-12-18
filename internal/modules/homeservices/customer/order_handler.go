package customer

import (
	"github.com/gin-gonic/gin"

	"github.com/umar5678/go-backend/internal/modules/homeservices/customer/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

// OrderHandler handles HTTP requests for customer orders
type OrderHandler struct {
	service OrderService
}

// NewOrderHandler creates a new order handler instance
func NewOrderHandler(service OrderService) *OrderHandler {
	return &OrderHandler{service: service}
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
func (h *OrderHandler) CreateOrder(c *gin.Context) {
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
func (h *OrderHandler) GetOrder(c *gin.Context) {
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
func (h *OrderHandler) ListOrders(c *gin.Context) {
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
func (h *OrderHandler) GetCancellationPreview(c *gin.Context) {
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
func (h *OrderHandler) CancelOrder(c *gin.Context) {
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
func (h *OrderHandler) RateOrder(c *gin.Context) {
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
