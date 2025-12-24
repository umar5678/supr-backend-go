package homeservices

import (
	"fmt"

	"github.com/gin-gonic/gin"

	homeservicedto "github.com/umar5678/go-backend/internal/modules/homeservices/dto"
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
// @Success 200 {object} response.Response{data=[]homeservicedto.ServiceCategoryResponse}
// @Router /services/categories [get]
func (h *Handler) ListCategories(c *gin.Context) {
	categories, err := h.service.ListCategories(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, categories, "Categories retrieved successfully")
}

// GetAllCategorySlugs godoc
// @Summary Get all category slugs
// @Description Get list of all distinct category slugs for provider registration dropdown
// @Tags Provider - Profile
// @Produce json
// @Success 200 {object} response.Response{data=[]string}
// @Router /services/category-slugs [get]
func (h *Handler) GetAllCategorySlugs(c *gin.Context) {
	slugs, err := h.service.GetAllCategorySlugs(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, slugs, "Category slugs retrieved successfully")
}

// GetCategoryWithTabs godoc
// @Summary Get category with tabs
// @Description Get detailed category information with all tabs
// @Tags home-services
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.Response{data=homeservicedto.CategoryWithTabsResponse}
// @Failure 404 {object} response.Response
// @Router /services/categories/{id} [get]
func (h *Handler) GetCategoryWithTabs(c *gin.Context) {
	var id uint
	if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
		c.Error(response.BadRequest("Invalid category ID"))
		return
	}

	category, err := h.service.GetCategoryWithTabs(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, category, "Category retrieved successfully")
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
// @Success 200 {object} response.Response{data=[]homeservicedto.ServiceListResponse}
// @Router /services [get]
func (h *Handler) ListServices(c *gin.Context) {
	var query homeservicedto.ListServicesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}

	query.SetDefaults()
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
// @Success 200 {object} response.Response{data=homeservicedto.ServiceDetailResponse}
// @Failure 404 {object} response.Response
// @Router /services/{id} [get]
func (h *Handler) GetServiceDetails(c *gin.Context) {
	var id uint
	if _, err := fmt.Sscan(c.Param("id"), &id); err != nil {
		c.Error(response.BadRequest("Invalid service ID"))
		return
	}

	service, err := h.service.GetServiceDetails(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, service, "Service details retrieved successfully")
}

// ListAddOns godoc
// @Summary List add-ons
// @Description Get all available add-ons for a category
// @Tags home-services
// @Produce json
// @Param categoryId query int true "Category ID"
// @Success 200 {object} response.Response{data=[]homeservicedto.AddOnResponse}
// @Router /services/addons [get]
func (h *Handler) ListAddOns(c *gin.Context) {
	var categoryID uint
	if _, err := fmt.Sscan(c.Query("categoryId"), &categoryID); err != nil || categoryID == 0 {
		c.Error(response.BadRequest("Valid category ID is required"))
		return
	}

	addOns, err := h.service.ListAddOns(c.Request.Context(), categoryID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, addOns, "Add-ons retrieved successfully")
}

// ==================== CUSTOMER ENDPOINTS ====================

// CreateOrder godoc
// @Summary Create a service order
// @Description Book a home service
// @Tags home-services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body homeservicedto.CreateOrderRequest true "Order details"
// @Success 201 {object} response.Response{data=homeservicedto.OrderResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 422 {object} response.Response
// @Router /services/orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	var req homeservicedto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	if err := req.Validate(); err != nil {
		c.Error(err)
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
// @Success 200 {object} response.Response{data=[]homeservicedto.OrderListResponse}
// @Failure 401 {object} response.Response
// @Router /services/orders [get]
func (h *Handler) GetMyOrders(c *gin.Context) {
	var query homeservicedto.ListOrdersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}

	query.SetDefaults()
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
// @Success 200 {object} response.Response{data=homeservicedto.OrderResponse}
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

// ==================== PROVIDER ENDPOINTS ====================

// GetProviderOrders godoc
// @Summary Get provider orders
// @Description Get orders assigned to the provider
// @Tags home-services-provider
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Filter by status"
// @Success 200 {object} response.Response{data=[]homeservicedto.OrderListResponse}
// @Failure 401 {object} response.Response
// @Router /services/provider/orders [get]
func (h *Handler) GetProviderOrders(c *gin.Context) {
	var query homeservicedto.ListOrdersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}

	query.SetDefaults()
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

// ==================== ADMIN ENDPOINTS ====================

// CreateCategory godoc
// @Summary Create a category
// @Description Create a new service category (admin only)
// @Tags home-services-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body homeservicedto.CreateCategoryRequest true "Category details"
// @Success 201 {object} response.Response{data=homeservicedto.CategoryWithTabsResponse}
// @Router /services/admin/categories [post]
func (h *Handler) CreateCategory(c *gin.Context) {
	var req homeservicedto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	category, err := h.service.CreateCategory(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, category, "Category created successfully")
}

// CreateTab godoc
// @Summary Create a tab
// @Description Create a new service tab/subcategory (admin only)
// @Tags home-services-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body homeservicedto.CreateTabRequest true "Tab details"
// @Success 201 {object} response.Response{data=homeservicedto.ServiceTabResponse}
// @Router /services/admin/tabs [post]
func (h *Handler) CreateTab(c *gin.Context) {
	var req homeservicedto.CreateTabRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	tab, err := h.service.CreateTab(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, tab, "Tab created successfully")
}

// CreateService godoc
// @Summary Create a service
// @Description Create a new home service (admin only)
// @Tags home-services-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body homeservicedto.CreateServiceRequest true "Service details"
// @Success 201 {object} response.Response{data=homeservicedto.ServiceResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Router /services/admin/services [post]
func (h *Handler) CreateService(c *gin.Context) {
	var req homeservicedto.CreateServiceRequest
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

// CreateAddOn godoc
// @Summary Create an add-on
// @Description Create a new service add-on (admin only)
// @Tags home-services-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body homeservicedto.CreateAddOnRequest true "Add-on details"
// @Success 201 {object} response.Response{data=homeservicedto.AddOnResponse}
// @Router /services/admin/addons [post]
func (h *Handler) CreateAddOn(c *gin.Context) {
	var req homeservicedto.CreateAddOnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	addOn, err := h.service.CreateAddOn(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, addOn, "Add-on created successfully")
}

// UpdateService godoc
// @Summary Update a service
// @Description Update service details (admin only)
// @Tags home-services-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Service ID"
// @Param request body homeservicedto.UpdateServiceRequest true "Update details"
// @Success 200 {object} response.Response{data=homeservicedto.ServiceResponse}
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

	var req homeservicedto.UpdateServiceRequest
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

// RegisterProvider godoc
// @Summary Register as service provider
// @Description Register authenticated user as a home service provider
// @Tags home-services-provider
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body homeservicedto.RegisterProviderRequest true "Provider registration details"
// @Success 201 {object} response.Response{data=homeservicedto.ProviderProfileResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /services/provider/register [post]
func (h *Handler) RegisterProvider(c *gin.Context) {
	var req homeservicedto.RegisterProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	if err := req.Validate(); err != nil {
		c.Error(err)
		return
	}

	userID, _ := c.Get("userID")

	provider, err := h.service.RegisterProvider(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, provider, "Provider registered successfully")
}
