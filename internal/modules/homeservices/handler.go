package homeservices

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/umar5678/go-backend/internal/modules/homeservices/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// --- Customer Endpoints ---

// ListCategories godoc
// @Summary List service categories
// @Description Get all active service categories
// @Tags home-services
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.ServiceCategoryResponse}
// @Router /services/categories [get]
func (h *Handler) ListCategories(c *gin.Context) {
	categories, err := h.service.ListCategories(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, categories, "Categories retrieved successfully")
}

// ListServices godoc
// @Summary List services
// @Description Get paginated list of services with filters
// @Tags home-services
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param categoryId query int false "Filter by category ID"
// @Param search query string false "Search in name and description"
// @Param minPrice query number false "Minimum price"
// @Param maxPrice query number false "Maximum price"
// @Param isActive query bool false "Filter by active status"
// @Success 200 {object} response.Response{data=[]dto.ServiceListResponse}
// @Router /services [get]
func (h *Handler) ListServices(c *gin.Context) {
	var query dto.ListServicesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}

	services, pagination, err := h.service.ListServices(c.Request.Context(), query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Paginated(c, services, *pagination, "Services retrieved successfully")
}

// GetServiceDetails godoc
// @Summary Get service details
// @Description Get detailed service information including options
// @Tags home-services
// @Produce json
// @Param id path int true "Service ID"
// @Success 200 {object} response.Response{data=dto.ServiceResponse}
// @Failure 404 {object} response.Response
// @Router /services/{id} [get]
func (h *Handler) GetServiceDetails(c *gin.Context) {
	id := c.Param("id")

	var serviceID uint
	if _, err := fmt.Sscan(id, &serviceID); err != nil {
		c.Error(response.BadRequest("Invalid service ID"))
		return
	}

	service, err := h.service.GetServiceDetails(c.Request.Context(), serviceID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, service, "Service details retrieved successfully")
}

// CreateOrder godoc
// @Summary Create a service order
// @Description Book a home service
// @Tags home-services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateOrderRequest true "Order details"
// @Success 201 {object} response.Response{data=dto.OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 422 {object} response.Response
// @Router /services/orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	userID, _ := c.Get("userID")

	order, err := h.service.CreateOrder(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Order created successfully")
}

// GetMyOrders godoc
// @Summary Get my orders
// @Description Get authenticated user's service orders
// @Tags home-services
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Filter by status"
// @Success 200 {object} response.Response{data=[]dto.OrderListResponse}
// @Failure 401 {object} response.Response
// @Router /services/orders [get]
func (h *Handler) GetMyOrders(c *gin.Context) {
	var query dto.ListOrdersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}

	userID, _ := c.Get("userID")

	orders, pagination, err := h.service.GetMyOrders(c.Request.Context(), userID.(string), query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Paginated(c, orders, *pagination, "Orders retrieved successfully")
}

// GetOrderDetails godoc
// @Summary Get order details
// @Description Get detailed information about a specific order
// @Tags home-services
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response{data=dto.OrderResponse}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /services/orders/{id} [get]
func (h *Handler) GetOrderDetails(c *gin.Context) {
	orderID := c.Param("id")
	userID, _ := c.Get("userID")

	order, err := h.service.GetOrderDetails(c.Request.Context(), userID.(string), orderID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, order, "Order details retrieved successfully")
}

// CancelOrder godoc
// @Summary Cancel an order
// @Description Cancel a pending or searching order
// @Tags home-services
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /services/orders/{id}/cancel [post]
func (h *Handler) CancelOrder(c *gin.Context) {
	orderID := c.Param("id")
	userID, _ := c.Get("userID")

	if err := h.service.CancelOrder(c.Request.Context(), userID.(string), orderID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Order cancelled successfully")
}

// --- Provider Endpoints ---

// GetProviderOrders godoc
// @Summary Get provider orders
// @Description Get orders assigned to the provider
// @Tags home-services-provider
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Filter by status"
// @Success 200 {object} response.Response{data=[]dto.OrderListResponse}
// @Failure 401 {object} response.Response
// @Router /services/provider/orders [get]
func (h *Handler) GetProviderOrders(c *gin.Context) {
	var query dto.ListOrdersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}

	// Get provider ID from context (set by auth middleware)
	providerID, exists := c.Get("providerID")
	if !exists {
		c.Error(response.ForbiddenError("Provider account required"))
		return
	}

	orders, pagination, err := h.service.GetProviderOrders(c.Request.Context(), providerID.(string), query)
	if err != nil {
		c.Error(err)
		return
	}

	response.Paginated(c, orders, *pagination, "Orders retrieved successfully")
}

// AcceptOrder godoc
// @Summary Accept an order
// @Description Accept a job offer
// @Tags home-services-provider
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /services/provider/orders/{id}/accept [post]
func (h *Handler) AcceptOrder(c *gin.Context) {
	orderID := c.Param("id")
	providerID, _ := c.Get("providerID")

	if err := h.service.AcceptOrder(c.Request.Context(), providerID.(string), orderID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Order accepted successfully")
}

// RejectOrder godoc
// @Summary Reject an order
// @Description Reject a job offer
// @Tags home-services-provider
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /services/provider/orders/{id}/reject [post]
func (h *Handler) RejectOrder(c *gin.Context) {
	orderID := c.Param("id")
	providerID, _ := c.Get("providerID")

	if err := h.service.RejectOrder(c.Request.Context(), providerID.(string), orderID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Order rejected successfully")
}

// StartOrder godoc
// @Summary Start an order
// @Description Mark order as in progress
// @Tags home-services-provider
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /services/provider/orders/{id}/start [post]
func (h *Handler) StartOrder(c *gin.Context) {
	orderID := c.Param("id")
	providerID, _ := c.Get("providerID")

	if err := h.service.StartOrder(c.Request.Context(), providerID.(string), orderID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Order started successfully")
}

// CompleteOrder godoc
// @Summary Complete an order
// @Description Mark order as completed and process payment
// @Tags home-services-provider
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /services/provider/orders/{id}/complete [post]
func (h *Handler) CompleteOrder(c *gin.Context) {
	orderID := c.Param("id")
	providerID, _ := c.Get("providerID")

	if err := h.service.CompleteOrder(c.Request.Context(), providerID.(string), orderID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Order completed successfully")
}

// --- Admin Endpoints ---

// CreateService godoc
// @Summary Create a service
// @Description Create a new home service (admin only)
// @Tags home-services-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateServiceRequest true "Service details"
// @Success 201 {object} response.Response{data=dto.ServiceResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /services/admin/services [post]
func (h *Handler) CreateService(c *gin.Context) {
	var req dto.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	service, err := h.service.CreateService(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, service, "Service created successfully")
}

// UpdateService godoc
// @Summary Update a service
// @Description Update service details (admin only)
// @Tags home-services-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Service ID"
// @Param request body dto.UpdateServiceRequest true "Update details"
// @Success 200 {object} response.Response{data=dto.ServiceResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /services/admin/services/{id} [put]
func (h *Handler) UpdateService(c *gin.Context) {
	id := c.Param("id")

	var serviceID uint
	if _, err := fmt.Sscan(id, &serviceID); err != nil {
		c.Error(response.BadRequest("Invalid service ID"))
		return
	}

	var req dto.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	service, err := h.service.UpdateService(c.Request.Context(), serviceID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, service, "Service updated successfully")
}
