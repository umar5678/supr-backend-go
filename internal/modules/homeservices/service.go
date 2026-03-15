package homeservices

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/models"
	homeservicedto "github.com/umar5678/go-backend/internal/modules/homeservices/dto"
	notificationsmodule "github.com/umar5678/go-backend/internal/modules/notifications"
	"github.com/umar5678/go-backend/internal/modules/wallet"
	walletdto "github.com/umar5678/go-backend/internal/modules/wallet/dto"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

const (
	ProviderOfferTimeout = 60 * time.Second
	HoldExpiryDuration   = 24 * time.Hour
	DefaultSearchRadius  = 15000
)

type Service interface {

	ListCategories(ctx context.Context) ([]*homeservicedto.ServiceCategoryResponse, error)
	GetCategoryWithTabs(ctx context.Context, id uint) (*homeservicedto.CategoryWithTabsResponse, error)
	GetAllCategorySlugs(ctx context.Context) ([]string, error)
	ListServices(ctx context.Context, query homeservicedto.ListServicesQuery) ([]*homeservicedto.ServiceListResponse, *response.PaginationMeta, error)
	GetServiceDetails(ctx context.Context, id uint) (*homeservicedto.ServiceDetailResponse, error)
	ListAddOns(ctx context.Context, categoryID uint) ([]*homeservicedto.AddOnResponse, error)

	CreateOrder(ctx context.Context, userID string, req homeservicedto.CreateOrderRequest) (*homeservicedto.OrderResponse, error)
	GetMyOrders(ctx context.Context, userID string, query homeservicedto.ListOrdersQuery) ([]*homeservicedto.OrderListResponse, *response.PaginationMeta, error)
	GetOrderDetails(ctx context.Context, userID, orderID string) (*homeservicedto.OrderResponse, error)
	CancelOrder(ctx context.Context, userID, orderID string) error

	GetProviderOrders(ctx context.Context, providerID string, query homeservicedto.ListOrdersQuery) ([]*homeservicedto.OrderListResponse, *response.PaginationMeta, error)
	RegisterProvider(ctx context.Context, userID string, req homeservicedto.RegisterProviderRequest) (*homeservicedto.ProviderProfileResponse, error)
	AcceptOrder(ctx context.Context, providerID, orderID string) error
	RejectOrder(ctx context.Context, providerID, orderID string) error
	StartOrder(ctx context.Context, providerID, orderID string) error
	CompleteOrder(ctx context.Context, providerID, orderID string) error

	FindAndNotifyNextProvider(orderID string)

	CreateCategory(ctx context.Context, req homeservicedto.CreateCategoryRequest) (*homeservicedto.CategoryWithTabsResponse, error)
	CreateTab(ctx context.Context, req homeservicedto.CreateTabRequest) (*homeservicedto.ServiceTabResponse, error)
	CreateService(ctx context.Context, req homeservicedto.CreateServiceRequest) (*homeservicedto.ServiceDetailResponse, error)
	UpdateService(ctx context.Context, id uint, req homeservicedto.UpdateServiceRequest) (*homeservicedto.ServiceDetailResponse, error)
	CreateAddOn(ctx context.Context, req homeservicedto.CreateAddOnRequest) (*homeservicedto.AddOnResponse, error)
}

type service struct {
	repo            Repository
	walletService   wallet.Service
	cfg             *config.Config
	eventProducer   notificationsmodule.EventProducer
}

func NewService(repo Repository, walletService wallet.Service, cfg *config.Config) Service {
	return NewServiceWithNotifications(repo, walletService, cfg, nil)
}

func NewServiceWithNotifications(repo Repository, walletService wallet.Service, cfg *config.Config, eventProducer notificationsmodule.EventProducer) Service {
	return &service{
		repo:          repo,
		walletService: walletService,
		cfg:           cfg,
		eventProducer: eventProducer,
	}
}

func (s *service) ListCategories(ctx context.Context) ([]*homeservicedto.ServiceCategoryResponse, error) {
	categories, err := s.repo.ListCategories(ctx)
	if err != nil {
		logger.Error("failed to list categories", "error", err)
		return nil, response.InternalServerError("Failed to fetch categories", err)
	}

	return homeservicedto.ToServiceCategoryList(categories), nil
}

func (s *service) GetAllCategorySlugs(ctx context.Context) ([]string, error) {
	slugs, err := s.repo.GetAllCategorySlugs(ctx)
	if err != nil {
		logger.Error("failed to list category slugs", "error", err)
		return nil, response.InternalServerError("Failed to fetch category slugs", err)
	}
	return slugs, nil
}

func (s *service) GetCategoryWithTabs(ctx context.Context, id uint) (*homeservicedto.CategoryWithTabsResponse, error) {
	category, err := s.repo.GetCategoryWithTabs(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Category")
		}
		logger.Error("failed to get category with tabs", "error", err, "categoryID", id)
		return nil, response.InternalServerError("Failed to fetch category", err)
	}

	return homeservicedto.ToCategoryWithTabsResponse(category), nil
}

func (s *service) ListServices(ctx context.Context, query homeservicedto.ListServicesQuery) ([]*homeservicedto.ServiceListResponse, *response.PaginationMeta, error) {
	query.SetDefaults()

	services, total, err := s.repo.ListServices(ctx, query)
	if err != nil {
		logger.Error("failed to list services", "error", err)
		return nil, nil, response.InternalServerError("Failed to fetch services", err)
	}

	responses := make([]*homeservicedto.ServiceListResponse, len(services))
	for i, svc := range services {
		responses[i] = &homeservicedto.ServiceListResponse{
			ID:                 svc.ID,
			CategoryID:         svc.CategoryID,
			TabID:              svc.TabID,
			Name:               svc.Name,
			ImageURL:           svc.ImageURL,
			BasePrice:          svc.BasePrice,
			OriginalPrice:      svc.OriginalPrice,
			DiscountPercentage: s.calculateDiscountPercentage(svc.OriginalPrice, svc.BasePrice),
			DurationMinutes:    svc.BaseDurationMinutes,
			IsActive:           svc.IsActive,
			IsFeatured:         svc.IsFeatured,
			CreatedAt:          svc.CreatedAt,
		}
	}

	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)
	return responses, &pagination, nil
}

func (s *service) GetServiceDetails(ctx context.Context, id uint) (*homeservicedto.ServiceDetailResponse, error) {
	service, err := s.repo.GetServiceWithOptions(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Service")
		}
		logger.Error("failed to get service", "error", err, "serviceID", id)
		return nil, response.InternalServerError("Failed to fetch service", err)
	}

	return homeservicedto.ToServiceDetailResponse(service), nil
}

func (s *service) ListAddOns(ctx context.Context, categoryID uint) ([]*homeservicedto.AddOnResponse, error) {
	addOns, err := s.repo.ListAddOns(ctx, categoryID)
	if err != nil {
		logger.Error("failed to list add-ons", "error", err, "categoryID", categoryID)
		return nil, response.InternalServerError("Failed to fetch add-ons", err)
	}

	responses := make([]*homeservicedto.AddOnResponse, len(addOns))
	for i, addon := range addOns {
		responses[i] = homeservicedto.ToAddOnResponse(&addon)
	}

	return responses, nil
}

func (s *service) RegisterProvider(ctx context.Context, userID string, req homeservicedto.RegisterProviderRequest) (*homeservicedto.ProviderProfileResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	_, err := s.repo.FindProviderByUserID(ctx, userID)
	if err == nil {
		return nil, response.BadRequest("User is already registered as a service provider")
	}

	providerID := uuid.New().String()
	provider := &models.ServiceProviderProfile{
		ID:              providerID,
		UserID:          userID,
		ServiceCategory: req.CategorySlug,
		ServiceType:     req.CategorySlug,
		Status:          models.SPStatusActive,
		IsVerified:      true,
		IsAvailable:     true,
	}

	if err := s.repo.CreateProvider(ctx, provider); err != nil {
		logger.Error("failed to create provider profile", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to create provider profile", err)
	}

	if req.CategorySlug != "" {
		category := &models.ProviderServiceCategory{
			ProviderID:        providerID,
			CategorySlug:      req.CategorySlug,
			ExpertiseLevel:    "beginner",
			YearsOfExperience: 0,
			IsActive:          true,
		}
		if err := s.repo.AddProviderCategory(ctx, category); err != nil {
			logger.Error("failed to add provider category", "error", err, "providerID", providerID, "category", req.CategorySlug)
		}
	}

	logger.Info("provider registered with category-based dynamic service assignment",
		"providerID", providerID,
		"userID", userID,
		"category", req.CategorySlug,
	)

	logger.Info("provider registered successfully",
		"providerID", providerID,
		"userID", userID,
		"category", req.CategorySlug,
	)

	provider, _ = s.repo.GetProviderByID(ctx, providerID)

	return &homeservicedto.ProviderProfileResponse{
		ID:              provider.ID,
		UserID:          provider.UserID,
		ServiceCategory: provider.ServiceCategory,
		Status:          string(provider.Status),
		IsVerified:      provider.IsVerified,
		Rating:          provider.Rating,
		TotalReviews:    provider.TotalReviews,
		CompletedJobs:   provider.CompletedJobs,
		IsAvailable:     provider.IsAvailable,
		Currency:        provider.Currency,
		CreatedAt:       provider.CreatedAt,
	}, nil
}

func (s *service) CreateOrder(ctx context.Context, userID string, req homeservicedto.CreateOrderRequest) (*homeservicedto.OrderResponse, error) {
	req.SetDefaults()
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	serviceDate, err := time.Parse(time.RFC3339, req.ServiceDate)
	if err != nil {
		return nil, response.BadRequest("Invalid service date format. Use RFC3339")
	}

	if serviceDate.Before(time.Now()) {
		return nil, response.BadRequest("Service date must be in the future")
	}

	var categorySlug string
	var subtotal float64
	selectedServices := models.SelectedServices{}

	for i, itemReq := range req.Items {
		svc, err := s.repo.GetServiceNewByID(ctx, itemReq.ServiceID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, response.BadRequest(fmt.Sprintf("Service with ID %s not found", itemReq.ServiceID))
			}
			return nil, response.InternalServerError("Failed to fetch service", err)
		}

		if !svc.IsActive {
			return nil, response.BadRequest(fmt.Sprintf("Service '%s' is not available", svc.Title))
		}

		if i == 0 {
			categorySlug = svc.CategorySlug
		}

		price := 0.0
		if svc.BasePrice != nil {
			price = *svc.BasePrice
		}

		selectedServices = append(selectedServices, models.SelectedServiceItem{
			ServiceSlug: svc.ServiceSlug,
			Title:       svc.Title,
			Price:       price,
			Quantity:    1,
		})

		subtotal += price
	}

	var selectedAddons models.SelectedAddons
	if len(req.AddOnIDs) > 0 {
		addOnServices, err := s.repo.GetAddOnsByIDs(ctx, req.AddOnIDs)
		if err != nil {
			return nil, response.InternalServerError("Failed to fetch add-ons", err)
		}

		for _, addon := range addOnServices {
			if !addon.IsActive {
				return nil, response.BadRequest(fmt.Sprintf("Add-on '%s' is not available", addon.Title))
			}

			selectedAddons = append(selectedAddons, models.SelectedAddonItem{
				AddonSlug: addon.Title,
				Title:     addon.Title,
				Price:     addon.Price,
				Quantity:  1,
			})

			subtotal += addon.Price
		}
	}

	platformFee := subtotal * 0.10
	totalPrice := subtotal + platformFee

	// 6. Create wallet hold using existing wallet service (non-blocking for cash model)
	var holdID *string
	holdReq := walletdto.HoldFundsRequest{
		Amount:        totalPrice,
		ReferenceType: "service_order",
		ReferenceID:   uuid.New().String(),
		HoldDuration:  int(HoldExpiryDuration.Minutes()),
	}

	holdResp, err := s.walletService.HoldFunds(ctx, userID, holdReq)
	if err != nil {
		logger.Warn("failed to create payment tracking hold", "error", err, "total", totalPrice)
		// Don't fail the order - cash payment doesn't require wallet authorization
	} else {
		holdID = &holdResp.ID
		logger.Info("payment hold created for tracking", "holdID", holdResp.ID, "amount", totalPrice)
	}

	// 7. Create order using ServiceOrderNew
	order := &models.ServiceOrderNew{
		ID:          uuid.New().String(),
		OrderNumber: s.generateOrderCode(),
		CustomerID:  userID,
		CustomerInfo: models.CustomerInfo{
			Name:    "",
			Phone:   "",
			Email:   "",
			Address: req.Address,
			Lat:     req.Latitude,
			Lng:     req.Longitude,
		},
		BookingInfo: models.BookingInfo{
			Date:           serviceDate.Format("2006-01-02"),
			Time:           serviceDate.Format("15:04"),
			QuantityOfPros: req.QuantityOfPros,
			PersonCount:    req.PersonCount,
			ToolsRequired:  req.ToolsRequired,
			Frequency:      &req.Frequency,
		},
		CategorySlug:       categorySlug,
		SelectedServices:   selectedServices,
		SelectedAddons:     selectedAddons,
		SpecialNotes:       "",
		ServicesTotal:      subtotal,
		AddonsTotal:        0,
		Subtotal:           subtotal,
		PlatformCommission: platformFee,
		TotalPrice:         totalPrice,
		PaymentInfo: &models.PaymentInfo{
			Method: "cash",
			Status: "pending",
			Total:  totalPrice,
		},
		WalletHoldID: holdID,
		Status:       "searching_provider",
	}

	if err := s.repo.CreateOrder(ctx, order); err != nil {
		// Release hold if order creation fails and hold was successful
		if holdID != nil {
			releaseReq := walletdto.ReleaseHoldRequest{HoldID: *holdID}
			s.walletService.ReleaseHold(ctx, userID, releaseReq)
		}
		logger.Error("failed to create order", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to create order", err)
	}

	logger.Info("order created", "orderID", order.ID, "userID", userID, "total", totalPrice)

	// 8. Trigger async provider search
	go s.FindAndNotifyNextProvider(order.ID)

	// Convert to DTO response using the customer DTO converter
	return homeservicedto.ToOrderResponseFromNew(order), nil
}

// Helper function to check if services use hourly pricing
func (s *service) isHourlyPricing(services []*models.Service) bool {
	for _, svc := range services {
		if svc.PricingModel == "hourly" {
			return true
		}
	}
	return false
}

// ✅ Generate unique order code
func (s *service) generateOrderCode() string {
	// Format: HS-YYYY-NNNNNN (HS = Home Service)
	year := time.Now().Year()
	random := rand.Intn(999999)
	return fmt.Sprintf("HS-%d-%06d", year, random)
}

func (s *service) GetMyOrders(ctx context.Context, userID string, query homeservicedto.ListOrdersQuery) ([]*homeservicedto.OrderListResponse, *response.PaginationMeta, error) {
	query.SetDefaults()

	orders, total, err := s.repo.ListUserOrders(ctx, userID, query)
	if err != nil {
		logger.Error("failed to list user orders", "error", err, "userID", userID)
		return nil, nil, response.InternalServerError("Failed to fetch orders", err)
	}

	responses := homeservicedto.ToOrderListResponses(orders)
	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

func (s *service) GetOrderDetails(ctx context.Context, userID, orderID string) (*homeservicedto.OrderResponse, error) {
	// Fetch ServiceOrderNew directly for rich details
	orderNew, err := s.repo.GetOrderByIDWithDetails(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		logger.Error("failed to fetch order details", "error", err, "orderID", orderID)
		return nil, response.InternalServerError("Failed to fetch order", err)
	}

	if orderNew == nil {
		logger.Error("order is nil after fetch", "orderID", orderID)
		return nil, response.NotFoundError("Order")
	}

	// Verify ownership
	if orderNew.CustomerID != userID {
		logger.Warn("unauthorized order access attempt", "orderID", orderID, "requestUserID", userID, "orderUserID", orderNew.CustomerID)
		return nil, response.ForbiddenError("You don't have access to this order")
	}

	// Convert to OrderResponse with full details from ServiceOrderNew
	return homeservicedto.ToOrderResponseFromNew(orderNew), nil
}

func (s *service) CancelOrder(ctx context.Context, CustomerID, orderID string) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.NotFoundError("Order")
		}
		return response.InternalServerError("Failed to fetch order", err)
	}

	// Verify ownership
	if order.CustomerID != CustomerID {
		return response.ForbiddenError("You don't have access to this order")
	}

	// Check if order can be cancelled
	if order.Status == "completed" || order.Status == "cancelled" {
		return response.BadRequest("Cannot cancel order in current status")
	}

	// Release wallet hold using existing wallet service
	if order.WalletHoldID != nil {
		releaseReq := walletdto.ReleaseHoldRequest{
			HoldID: *order.WalletHoldID,
		}
		if err := s.walletService.ReleaseHold(ctx, CustomerID, releaseReq); err != nil {
			logger.Error("failed to release hold on order cancellation", "error", err, "orderID", orderID)
		}
	}

	// Update order status
	if err := s.repo.UpdateOrderStatus(ctx, orderID, "cancelled"); err != nil {
		return response.InternalServerError("Failed to cancel order", err)
	}

	logger.Info("order cancelled", "orderID", orderID, "userID", CustomerID)

	return nil
}

// --- Provider - Orders ---

func (s *service) GetProviderOrders(ctx context.Context, providerID string, query homeservicedto.ListOrdersQuery) ([]*homeservicedto.OrderListResponse, *response.PaginationMeta, error) {
	query.SetDefaults()

	orders, total, err := s.repo.ListProviderOrders(ctx, providerID, query)
	if err != nil {
		logger.Error("failed to list provider orders", "error", err, "providerID", providerID)
		return nil, nil, response.InternalServerError("Failed to fetch orders", err)
	}

	responses := homeservicedto.ToOrderListResponses(orders)
	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

func (s *service) AcceptOrder(ctx context.Context, providerID, orderID string) error {
	offerKey := fmt.Sprintf("provider:%s:current_offer", providerID)

	// 1. Verify the offer is still valid
	val, err := cache.Get(ctx, offerKey)
	if err != nil || val != orderID {
		return response.ForbiddenError("Offer expired or invalid")
	}

	// 2. Assign provider to order
	if err := s.repo.AssignProviderToOrder(ctx, providerID, orderID); err != nil {
		logger.Error("failed to assign provider to order", "error", err, "providerID", providerID, "orderID", orderID)
		return response.InternalServerError("Failed to accept order", err)
	}

	// 3. Delete offer key to prevent timeout logic
	cache.Delete(ctx, offerKey)

	// 4. Update provider status
	s.repo.UpdateProviderStatus(ctx, providerID, "busy")

	logger.Info("provider accepted order", "providerID", providerID, "orderID", orderID)

	// TODO: Send notification to customer

	return nil
}

func (s *service) RejectOrder(ctx context.Context, providerID, orderID string) error {
	offerKey := fmt.Sprintf("provider:%s:current_offer", providerID)

	// Verify the offer exists
	val, err := cache.Get(ctx, offerKey)
	if err != nil || val != orderID {
		return response.ForbiddenError("No active offer for this order")
	}

	// Delete the offer
	cache.Delete(ctx, offerKey)

	logger.Info("provider rejected order", "providerID", providerID, "orderID", orderID)

	// Trigger finding next provider
	go s.FindAndNotifyNextProvider(orderID)

	return nil
}

func (s *service) StartOrder(ctx context.Context, providerID, orderID string) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return response.NotFoundError("Order")
	}

	if order.AssignedProviderID == nil || *order.AssignedProviderID != providerID {
		return response.ForbiddenError("You are not assigned to this order")
	}

	if order.Status != "accepted" {
		return response.BadRequest("Order cannot be started in current status")
	}

	if err := s.repo.UpdateOrderStatus(ctx, orderID, "in_progress"); err != nil {
		return response.InternalServerError("Failed to start order", err)
	}

	logger.Info("order started", "providerID", providerID, "orderID", orderID)

	return nil
}

func (s *service) CompleteOrder(ctx context.Context, providerID, orderID string) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		return response.NotFoundError("Order")
	}

	if order.AssignedProviderID == nil || *order.AssignedProviderID != providerID {
		return response.ForbiddenError("You are not assigned to this order")
	}

	if order.Status != "in_progress" {
		return response.BadRequest("Order must be in progress to complete")
	}

	// 1. Capture the wallet hold using existing wallet service
	if order.WalletHoldID != nil {
		captureReq := walletdto.CaptureHoldRequest{
			HoldID:      *order.WalletHoldID,
			Description: fmt.Sprintf("Payment for order %s", order.OrderNumber),
		}
		if _, err := s.walletService.CaptureHold(ctx, order.CustomerID, captureReq); err != nil {
			logger.Error("failed to capture hold", "error", err, "orderID", orderID)
			return response.InternalServerError("Payment processing failed", err)
		}
	}

	// 2. Transfer funds to provider using existing wallet service
	provider, err := s.repo.GetProviderByID(ctx, providerID)
	if err == nil && provider != nil {
		providerAmount := order.TotalPrice - order.PlatformCommission
		transferReq := walletdto.TransferFundsRequest{
			RecipientID: provider.UserID, // Use provider's UserID for wallet transfer
			Amount:      providerAmount,
			Description: fmt.Sprintf("Earnings from order %s", order.OrderNumber),
		}
		if _, err := s.walletService.TransferFunds(ctx, order.CustomerID, transferReq); err != nil {
			logger.Error("failed to transfer to provider", "error", err, "providerID", providerID)
			// Don't fail the completion, but log for manual reconciliation
		}
	}

	// 3. Update order status
	if err := s.repo.UpdateOrderStatus(ctx, orderID, "completed"); err != nil {
		return response.InternalServerError("Failed to complete order", err)
	}

	// 4. Update provider status back to available
	s.repo.UpdateProviderStatus(ctx, providerID, "available")

	logger.Info("order completed", "providerID", providerID, "orderID", orderID)

	return nil
}

func (s *service) FindAndNotifyNextProvider(orderID string) {
	ctx := context.Background()

	// 2. Fetch order
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		logger.Error("failed to fetch order for matching", "error", err, "orderID", orderID)
		return
	}

	// 3. Get service slugs from order items
	var serviceSlugs []string
	fullOrder, _ := s.repo.GetOrderByIDWithDetails(ctx, orderID)
	if fullOrder != nil {
		for _, item := range fullOrder.SelectedServices {
			serviceSlugs = append(serviceSlugs, item.ServiceSlug)
		}
	}

	if len(serviceSlugs) == 0 {
		logger.Error("no service slugs found for order", "orderID", orderID)
		return
	}

	// Get order's category slug for filtering providers
	orderCategorySlug := order.CategorySlug
	if orderCategorySlug == "" {
		logger.Error("order has no category slug", "orderID", orderID)
		return
	}

	if err := s.repo.UpdateOrderStatus(ctx, orderID, "pending"); err != nil {
		logger.Error("failed to update order status", "error", err, "orderID", orderID)
		// Release wallet hold on error
		if order.WalletHoldID != nil {
			releaseReq := walletdto.ReleaseHoldRequest{HoldID: *order.WalletHoldID}
			s.walletService.ReleaseHold(ctx, order.CustomerID, releaseReq)
		}
		return
	}

	logger.Info("order is available for providers",
		"orderID", orderID,
		"category", orderCategorySlug,
		"status", "pending")
}

// --- Admin - Service Management ---

func (s *service) CreateCategory(ctx context.Context, req homeservicedto.CreateCategoryRequest) (*homeservicedto.CategoryWithTabsResponse, error) {
	category := &models.ServiceCategory{
		Name:        req.Name,
		Description: req.Description,
		IconURL:     req.IconURL,
		BannerImage: req.BannerImage,
		Highlights:  req.Highlights,
		IsActive:    req.IsActive,
		SortOrder:   req.SortOrder,
	}

	if err := s.repo.CreateCategory(ctx, category); err != nil {
		logger.Error("failed to create category", "error", err)
		return nil, response.InternalServerError("Failed to create category", err)
	}

	logger.Info("category created", "categoryID", category.ID, "name", category.Name)

	return homeservicedto.ToCategoryWithTabsResponse(category), nil
}

func (s *service) CreateTab(ctx context.Context, req homeservicedto.CreateTabRequest) (*homeservicedto.ServiceTabResponse, error) {
	// Verify category exists
	_, err := s.repo.GetCategoryByID(ctx, req.CategoryID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.BadRequest("Category not found")
		}
		return nil, response.InternalServerError("Failed to verify category", err)
	}

	tab := &models.ServiceTab{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		IconURL:     req.IconURL,
		BannerTitle: req.BannerTitle,
		BannerDesc:  req.BannerDesc,
		BannerImage: req.BannerImage,
		IsActive:    req.IsActive,
		SortOrder:   req.SortOrder,
	}

	if err := s.repo.CreateTab(ctx, tab); err != nil {
		logger.Error("failed to create tab", "error", err)
		return nil, response.InternalServerError("Failed to create tab", err)
	}

	response := &homeservicedto.ServiceTabResponse{
		ID:            tab.ID,
		CategoryID:    tab.CategoryID,
		Name:          tab.Name,
		Description:   tab.Description,
		IconURL:       tab.IconURL,
		BannerTitle:   tab.BannerTitle,
		BannerDesc:    tab.BannerDesc,
		BannerImage:   tab.BannerImage,
		IsActive:      tab.IsActive,
		SortOrder:     tab.SortOrder,
		CreatedAt:     tab.CreatedAt,
		ServicesCount: 0, // You might want to calculate this
	}

	logger.Info("tab created", "tabID", tab.ID, "name", tab.Name)

	return response, nil
}

func (s *service) CreateService(ctx context.Context, req homeservicedto.CreateServiceRequest) (*homeservicedto.ServiceDetailResponse, error) {
	// Verify category exists
	_, err := s.repo.GetCategoryByID(ctx, req.CategoryID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.BadRequest("Category not found")
		}
		return nil, response.InternalServerError("Failed to verify category", err)
	}

	// Verify tab exists
	_, err = s.repo.GetTabByID(ctx, req.TabID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.BadRequest("Tab not found")
		}
		return nil, response.InternalServerError("Failed to verify tab", err)
	}

	service := &models.Service{
		CategoryID:          req.CategoryID,
		TabID:               req.TabID,
		Name:                req.Name,
		Description:         req.Description,
		ImageURL:            req.ImageURL,
		BasePrice:           req.BasePrice,
		OriginalPrice:       req.OriginalPrice,
		PricingModel:        req.PricingModel,
		BaseDurationMinutes: req.BaseDurationMinutes,
		MaxQuantity:         req.MaxQuantity,
		IsActive:            true,
		IsFeatured:          req.IsFeatured,
	}

	if err := s.repo.CreateService(ctx, service); err != nil {
		logger.Error("failed to create service", "error", err)
		return nil, response.InternalServerError("Failed to create service", err)
	}

	logger.Info("service created", "serviceID", service.ID, "name", service.Name)

	// Fetch the complete service with relations
	completeService, err := s.repo.GetServiceWithOptions(ctx, service.ID)
	if err != nil {
		return nil, response.InternalServerError("Failed to fetch created service", err)
	}

	return homeservicedto.ToServiceDetailResponse(completeService), nil
}

func (s *service) UpdateService(ctx context.Context, id uint, req homeservicedto.UpdateServiceRequest) (*homeservicedto.ServiceDetailResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	service, err := s.repo.GetServiceByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Service")
		}
		return nil, response.InternalServerError("Failed to fetch service", err)
	}

	// Apply updates
	if req.CategoryID != nil {
		service.CategoryID = *req.CategoryID
	}
	if req.Name != nil {
		service.Name = *req.Name
	}
	if req.Description != nil {
		service.Description = *req.Description
	}
	if req.ImageURL != nil {
		service.ImageURL = *req.ImageURL
	}
	if req.BasePrice != nil {
		service.BasePrice = *req.BasePrice
	}
	if req.PricingModel != nil {
		service.PricingModel = *req.PricingModel
	}
	if req.BaseDurationMinutes != nil {
		service.BaseDurationMinutes = *req.BaseDurationMinutes
	}
	if req.IsActive != nil {
		service.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateService(ctx, service); err != nil {
		logger.Error("failed to update service", "error", err, "serviceID", id)
		return nil, response.InternalServerError("Failed to update service", err)
	}

	logger.Info("service updated", "serviceID", id)

	// Fetch the complete service with relations
	completeService, err := s.repo.GetServiceWithOptions(ctx, id)
	if err != nil {
		return nil, response.InternalServerError("Failed to fetch updated service", err)
	}

	return homeservicedto.ToServiceDetailResponse(completeService), nil
}

func (s *service) CreateAddOn(ctx context.Context, req homeservicedto.CreateAddOnRequest) (*homeservicedto.AddOnResponse, error) {
	// Verify category exists
	_, err := s.repo.GetCategoryByID(ctx, req.CategoryID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.BadRequest("Category not found")
		}
		return nil, response.InternalServerError("Failed to verify category", err)
	}

	addOn := &models.AddOnService{
		CategoryID:      req.CategoryID,
		Title:           req.Title,
		Description:     req.Description,
		ImageURL:        req.ImageURL,
		Price:           req.Price,
		OriginalPrice:   req.OriginalPrice,
		DurationMinutes: req.DurationMinutes,
		IsActive:        req.IsActive,
		SortOrder:       req.SortOrder,
	}

	if err := s.repo.CreateAddOn(ctx, addOn); err != nil {
		logger.Error("failed to create add-on", "error", err)
		return nil, response.InternalServerError("Failed to create add-on", err)
	}

	logger.Info("add-on created", "addOnID", addOn.ID, "title", addOn.Title)

	return homeservicedto.ToAddOnResponse(addOn), nil
}

// --- Helper Functions ---

func (s *service) calculateItemPrice(svc *models.Service, selectedOptions []homeservicedto.SelectedOptionRequest) (float64, int, models.JSONBMap, error) {
	price := svc.BasePrice
	duration := svc.BaseDurationMinutes
	optionsMap := make(models.JSONBMap)

	// Build a map of option IDs for quick lookup
	optionsById := make(map[uint]*models.ServiceOption)
	for i := range svc.Options {
		optionsById[svc.Options[i].ID] = &svc.Options[i]
	}

	// Process each selected option
	for _, selOpt := range selectedOptions {
		option, exists := optionsById[selOpt.OptionID]
		if !exists {
			return 0, 0, nil, fmt.Errorf("invalid option ID: %d", selOpt.OptionID)
		}

		// Find the choice and apply price/duration modifiers
		if selOpt.ChoiceID != nil {
			choiceFound := false
			for _, choice := range option.Choices {
				if choice.ID == *selOpt.ChoiceID {
					price += choice.PriceModifier
					duration += choice.DurationModifierMinutes
					optionsMap[fmt.Sprintf("option_%d", selOpt.OptionID)] = choice.Label
					choiceFound = true
					break
				}
			}
			if !choiceFound {
				return 0, 0, nil, fmt.Errorf("invalid choice ID %d for option %d", *selOpt.ChoiceID, selOpt.OptionID)
			}
		} else if selOpt.Value != nil {
			// For text/quantity type options
			optionsMap[fmt.Sprintf("option_%d", selOpt.OptionID)] = *selOpt.Value
		}
	}

	// Validate required options are provided
	for _, option := range svc.Options {
		if option.IsRequired {
			if _, exists := optionsMap[fmt.Sprintf("option_%d", option.ID)]; !exists {
				return 0, 0, nil, fmt.Errorf("required option '%s' not provided", option.Name)
			}
		}
	}

	return price, duration, optionsMap, nil
}

func (s *service) calculateSurgeFee(lat, lon float64, serviceDate time.Time) float64 {
	// Simple time-based surge
	hour := serviceDate.Hour()
	if hour >= 17 && hour <= 21 {
		return 5.00 // Peak hours
	}

	// TODO: Query surge_zones table for location-based surge

	return 0.00
}

func (s *service) calculateDiscountPercentage(originalPrice, basePrice float64) int {
	if originalPrice > 0 && basePrice < originalPrice {
		return int(((originalPrice - basePrice) / originalPrice) * 100)
	}
	return 0
}

// Utility functions
func (s *service) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (s *service) parseCommaSeparated(str string) []string {
	if str == "" {
		return []string{}
	}
	return strings.Split(str, ",")
}

func (s *service) joinCommaSeparated(slice []string) string {
	return strings.Join(slice, ",")
}
