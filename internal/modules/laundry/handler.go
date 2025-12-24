package laundry

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/modules/laundry/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetServicesWithProducts - GET /api/v1/laundry/services
// @Summary Get Services With Products
// @Description Retrieve all available laundry services with nested products, pricing information, and turnaround times
// @Tags Laundry Services
// @Produce json
// @Success 200 {array} dto.LaundryServiceResponse "Services with products"
// @Failure 500 {object} response.Response "Failed to fetch services"
// @Router /api/v1/laundry/services [get]
func (h *Handler) GetServicesWithProducts(c *gin.Context) {
	services, err := h.service.GetServiceCatalog(c)
	if err != nil {
		c.Error(response.InternalServerError("Failed to fetch service catalog", err))
		return
	}

	// Convert models to DTOs
	serviceResponses := dto.ToLaundryServiceResponses(services)
	response.Success(c, serviceResponses, "Service catalog retrieved successfully")
}

// GetServiceProducts - GET /api/v1/laundry/services/:slug/products
// @Summary Get Products for Service
// @Description Retrieve all products available for a specific laundry service, including pricing and specifications
// @Tags Laundry Services
// @Produce json
// @Param slug path string true "Service slug (e.g., wash-fold, dry-clean)"
// @Success 200 {array} dto.ProductResponse "Products for the service"
// @Failure 404 {object} response.Response "Service not found"
// @Failure 500 {object} response.Response "Failed to fetch products"
// @Router /api/v1/laundry/services/{slug}/products [get]
func (h *Handler) GetServiceProducts(c *gin.Context) {
	slug := c.Param("slug")

	if slug == "" {
		c.Error(response.BadRequest("Service slug is required"))
		return
	}

	products, err := h.service.GetServiceProducts(c, slug)
	if err != nil {
		c.Error(response.InternalServerError("Failed to fetch products for service", err))
		return
	}

	if len(products) == 0 {
		c.Error(response.NotFoundError("No products found for service " + slug))
		return
	}

	response.Success(c, products, "Products retrieved successfully")
}

// CreateOrder - POST /api/v1/laundry/orders
// @Summary Create Laundry Order
// @Description Create a new laundry order with selected products. The system will:
// 1. Validate all products exist
// 2. Calculate price based on service pricing model (weight-based or item-based)
// 3. Add express fee if selected
// 4. Find the nearest service provider facility
// 5. Schedule pickup and delivery
// @Tags Laundry Orders
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateLaundryOrderRequest true "Order creation request with product selections"
// @Success 201 {object} dto.LaundryOrderResponse "Order created successfully with calculated pricing"
// @Failure 400 {object} response.Response "Invalid request or validation failed"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Failed to create order"
// @Router /api/v1/laundry/orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(response.UnauthorizedError("User ID not found in context"))
		return
	}

	var req dto.CreateLaundryOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.Error(response.BadRequest("Validation failed: " + err.Error()))
		return
	}

	order, err := h.service.CreateOrder(c, userID.(string), &req)
	if err != nil {
		c.Error(err)
		return
	}

	// Convert to response DTO for consistent frontend data
	orderResponse, err := h.service.GetOrder(c, order.ID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, orderResponse, "Order created successfully", "LAUNDRY_ORDER_CREATED")
}

// GetOrder - GET /api/v1/laundry/orders/:id
// @Summary Get Order Details
// @Description Retrieve full details of a laundry order including pickup, delivery, items, and issues
// @Tags Laundry Orders
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} dto.LaundryOrderResponse "Order details"
// @Router /api/v1/laundry/orders/{id} [get]
func (h *Handler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.Error(response.BadRequest("Order ID is required"))
		return
	}

	order, err := h.service.GetOrder(c, orderID)
	if err != nil {
		c.Error(response.NotFoundError("Order"))
		return
	}
	response.Success(c, order, "Order retrieved successfully")
}

// GetAvailableOrders - GET /api/v1/laundry/provider/orders/available
// @Summary Get Available Orders for Provider
// @Description Get all available laundry orders that match provider's service category
// The system automatically filters orders based on the provider's registered category.
// Providers see all orders in their category and can accept any of them.
// @Tags Provider - Orders
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} dto.LaundryOrderResponse "Available orders for provider"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Failed to fetch available orders"
// @Router /api/v1/laundry/provider/orders/available [get]
func (h *Handler) GetAvailableOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(response.UnauthorizedError("User ID not found in context"))
		return
	}

	// Get provider ID from user (user ID is provider ID in this context)
	providerID := userID.(string)

	orders, err := h.service.GetAvailableOrders(c, providerID)
	if err != nil {
		c.Error(response.InternalServerError("Failed to fetch available orders", err))
		return
	}

	response.Success(c, orders, "Available orders retrieved successfully")
}

// InitiatePickup - POST /api/v1/laundry/orders/:id/pickup/start
// @Summary Report Issue on Order
// @Description Report a problem with a laundry order (e.g., missing item, damage, poor cleaning, late delivery)
// @Summary Initiate Pickup for Order
// @Description Start a pickup for a laundry order at the specified location
// @Tags Provider - Pickups
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} dto.LaundryPickupResponse "Pickup initiated"
// @Router /api/v1/laundry/orders/{id}/pickup/start [post]
func (h *Handler) InitiatePickup(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.Error(response.BadRequest("Order ID is required"))
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.Error(response.UnauthorizedError("User ID not found in context"))
		return
	}

	// TODO: Get provider ID from user (implement provider lookup)
	providerID := userID.(string)

	pickup, err := h.service.InitiatePickup(c, orderID, providerID)
	if err != nil {
		c.Error(response.InternalServerError("Failed to initiate pickup", err))
		return
	}

	response.Success(c, pickup, "Pickup initiated successfully")
}

// CompletePickup - POST /api/v1/laundry/orders/:id/pickup/complete
// @Summary Complete Pickup
// @Description Mark a pickup as completed and record pickup details
// @Tags Provider - Pickups
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Param request body dto.CompletePickupRequest true "Pickup completion details"
// @Success 200 {object} dto.LaundryPickupResponse "Pickup completed successfully"
// @Router /api/v1/laundry/provider/orders/{id}/pickup/complete [post]
func (h *Handler) CompletePickup(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.Error(response.BadRequest("Order ID is required"))
		return
	}

	var req dto.CompletePickupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.Error(response.BadRequest("Validation failed: " + err.Error()))
		return
	}

	if err := h.service.CompletePickup(c, orderID, &req); err != nil {
		c.Error(response.InternalServerError("Failed to complete pickup", err))
		return
	}

	response.Success(c, nil, "Pickup completed successfully")
}

// AddItems - POST /api/v1/laundry/orders/:id/items
// @Summary Add Items to Order
// @Description Add laundry items (garments) to an order after pickup is completed. Each item gets a unique QR code
// @Tags Provider - Items
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Param request body dto.AddLaundryItemsRequest true "Items to add"
// @Success 201 {array} dto.LaundryOrderItemResponse "Items created with QR codes"
// @Router /api/v1/laundry/provider/orders/{id}/items [post]
func (h *Handler) AddItems(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.Error(response.BadRequest("Order ID is required"))
		return
	}

	var req dto.AddLaundryItemsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.Error(response.BadRequest("Validation failed: " + err.Error()))
		return
	}

	items, err := h.service.AddItems(c, orderID, &req)
	if err != nil {
		c.Error(response.InternalServerError("Failed to add items", err))
		return
	}

	response.Success(c, items, "Items added successfully", "LAUNDRY_ITEMS_ADDED")
}

// UpdateItemStatus - PATCH /api/v1/laundry/items/:qrCode/status
// @Summary Update Item Status
// @Description Update the processing status of a laundry item (garment). Valid statuses: pending, received, washing, drying, pressing, packed, delivered
// @Tags Provider - Items
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param qrCode path string true "Item QR Code (e.g., LDY-abc12345)"
// @Param request body dto.UpdateItemStatusRequest true "New status"
// @Success 200 {object} dto.LaundryOrderItemResponse "Item status updated"
// @Router /api/v1/laundry/provider/items/{qrCode}/status [patch]
func (h *Handler) UpdateItemStatus(c *gin.Context) {
	qrCode := c.Param("qrCode")
	if qrCode == "" {
		c.Error(response.BadRequest("QR code is required"))
		return
	}

	var req dto.UpdateItemStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.Error(response.BadRequest("Validation failed: " + err.Error()))
		return
	}

	item, err := h.service.UpdateItemStatus(c, qrCode, req.Status)
	if err != nil {
		c.Error(response.InternalServerError("Failed to update item status", err))
		return
	}

	response.Success(c, item, "Item status updated successfully")
}

// InitiateDelivery - POST /api/v1/laundry/orders/:id/delivery/start
// @Summary Initiate Delivery
// @Description Mark a laundry order delivery as initiated (provider is en route to deliver)
// @Tags Provider - Deliveries
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Success 200 {object} dto.LaundryDeliveryResponse "Delivery initiated"
// @Router /api/v1/laundry/provider/orders/{id}/delivery/start [post]
func (h *Handler) InitiateDelivery(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.Error(response.BadRequest("Order ID is required"))
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.Error(response.UnauthorizedError("User ID not found in context"))
		return
	}

	// TODO: Get provider ID from user (implement provider lookup)
	providerID := userID.(string)

	delivery, err := h.service.InitiateDelivery(c, orderID, providerID)
	if err != nil {
		c.Error(response.InternalServerError("Failed to initiate delivery", err))
		return
	}

	response.Success(c, delivery, "Delivery initiated successfully")
}

// CompleteDelivery - POST /api/v1/laundry/orders/:id/delivery/complete
// @Summary Complete Delivery
// @Description Mark a delivery as completed and record delivery details
// @Tags Provider - Deliveries
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Param request body dto.CompleteDeliveryRequest true "Delivery completion details"
// @Success 200 {object} dto.LaundryDeliveryResponse "Delivery completed successfully"
// @Router /api/v1/laundry/provider/orders/{id}/delivery/complete [post]
func (h *Handler) CompleteDelivery(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.Error(response.BadRequest("Order ID is required"))
		return
	}

	var req dto.CompleteDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.Error(response.BadRequest("Validation failed: " + err.Error()))
		return
	}

	if err := h.service.CompleteDelivery(c, orderID, &req); err != nil {
		c.Error(response.InternalServerError("Failed to complete delivery", err))
		return
	}

	response.Success(c, nil, "Delivery completed successfully")
}

// ReportIssue - POST /api/v1/laundry/orders/:id/issues
// @Summary Report Issue on Order
// @Description Report a problem with a laundry order (e.g., missing item, damage, poor cleaning, late delivery)
// @Tags Laundry Issues
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID (UUID)"
// @Param request body dto.ReportIssueRequest true "Issue report"
// @Success 201 {object} dto.LaundryIssueResponse "Issue created"
// @Router /api/v1/laundry/orders/{id}/issues [post]
func (h *Handler) ReportIssue(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.Error(response.BadRequest("Order ID is required"))
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.Error(response.UnauthorizedError("User ID not found in context"))
		return
	}

	var req dto.ReportIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		c.Error(response.BadRequest("Validation failed: " + err.Error()))
		return
	}

	// Get provider ID from order (empty string, will be resolved in service)
	issue, err := h.service.ReportIssue(c, orderID, userID.(string), "", &req)
	if err != nil {
		c.Error(response.InternalServerError("Failed to report issue", err))
		return
	}

	response.Success(c, issue, "Issue reported successfully", "LAUNDRY_ISSUE_REPORTED")
}

// GetProviderPickups - GET /api/v1/laundry/provider/pickups
// @Summary Get Provider Pickups
// @Description Retrieve all pending pickups assigned to the authenticated service provider
// @Tags Provider - Pickups
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} dto.LaundryPickupResponse "List of pending pickups"
// @Router /api/v1/laundry/provider/pickups [get]
func (h *Handler) GetProviderPickups(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(response.UnauthorizedError("User ID not found in context"))
		return
	}

	// TODO: Get provider ID from user (implement provider lookup)
	pickups, err := h.service.GetProviderPickups(c, userID.(string))
	if err != nil {
		c.Error(response.InternalServerError("Failed to fetch pickups", err))
		return
	}

	response.Success(c, pickups, "Pickups retrieved successfully")
}

// GetProviderDeliveries - GET /api/v1/laundry/provider/deliveries
// @Summary Get Provider Deliveries
// @Description Retrieve all pending deliveries assigned to the authenticated service provider
// @Tags Provider - Deliveries
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} dto.LaundryDeliveryResponse "List of pending deliveries"
// @Router /api/v1/laundry/provider/deliveries [get]
func (h *Handler) GetProviderDeliveries(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(response.UnauthorizedError("User ID not found in context"))
		return
	}

	// TODO: Get provider ID from user (implement provider lookup)
	deliveries, err := h.service.GetProviderDeliveries(c, userID.(string))
	if err != nil {
		c.Error(response.InternalServerError("Failed to fetch deliveries", err))
		return
	}

	response.Success(c, deliveries, "Deliveries retrieved successfully")
}

// GetProviderIssues - GET /api/v1/laundry/provider/issues
// @Summary Get Provider Issues
// @Description Retrieve all open issues reported for orders assigned to the authenticated service provider
// @Tags Provider - Issues
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {array} dto.LaundryIssueResponse "List of open issues"
// @Router /api/v1/laundry/provider/issues [get]
func (h *Handler) GetProviderIssues(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.Error(response.UnauthorizedError("User ID not found in context"))
		return
	}

	// TODO: Get provider ID from user (implement provider lookup)
	issues, err := h.service.GetProviderIssues(c, userID.(string))
	if err != nil {
		c.Error(response.InternalServerError("Failed to fetch issues", err))
		return
	}

	response.Success(c, issues, "Issues retrieved successfully")
}

// ResolveIssue - PATCH /api/v1/laundry/issues/:id
// @Summary Resolve Issue
// @Description Resolve a customer issue (mark as resolved or rejected) and optionally process a refund
// @Tags Provider - Issues
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Issue ID (UUID)"
// @Param request body dto.ResolveIssueRequest true "Resolution details"
// @Success 200 {object} dto.LaundryIssueResponse "Issue resolved"
// @Router /api/v1/laundry/provider/issues/{id} [patch]
func (h *Handler) ResolveIssue(c *gin.Context) {
	issueID := c.Param("id")
	if issueID == "" {
		c.Error(response.BadRequest("Issue ID is required"))
		return
	}

	var req gin.H
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body: " + err.Error()))
		return
	}

	resolution := ""
	if r, exists := req["resolution"]; exists {
		resolution = r.(string)
	}

	var refundAmount *float64
	if r, exists := req["refundAmount"]; exists {
		amount := r.(float64)
		refundAmount = &amount
	}

	if err := h.service.ResolveIssue(c, issueID, resolution, refundAmount); err != nil {
		c.Error(response.InternalServerError("Failed to resolve issue", err))
		return
	}

	response.Success(c, nil, "Issue resolved successfully")
}
