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
	"github.com/umar5678/go-backend/internal/modules/wallet"
	walletdto "github.com/umar5678/go-backend/internal/modules/wallet/dto"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

const (
	ProviderOfferTimeout = 60 * time.Second
	HoldExpiryDuration   = 24 * time.Hour
	DefaultSearchRadius  = 15000 // 15km in meters
)

type Service interface {
	// Customer - Service Catalog
	ListCategories(ctx context.Context) ([]*homeservicedto.ServiceCategoryResponse, error)
	GetCategoryWithTabs(ctx context.Context, id uint) (*homeservicedto.CategoryWithTabsResponse, error)
	GetAllCategorySlugs(ctx context.Context) ([]string, error)
	ListServices(ctx context.Context, query homeservicedto.ListServicesQuery) ([]*homeservicedto.ServiceListResponse, *response.PaginationMeta, error)
	GetServiceDetails(ctx context.Context, id uint) (*homeservicedto.ServiceDetailResponse, error)
	ListAddOns(ctx context.Context, categoryID uint) ([]*homeservicedto.AddOnResponse, error)

	// Customer - Orders
	CreateOrder(ctx context.Context, userID string, req homeservicedto.CreateOrderRequest) (*homeservicedto.OrderResponse, error)
	GetMyOrders(ctx context.Context, userID string, query homeservicedto.ListOrdersQuery) ([]*homeservicedto.OrderListResponse, *response.PaginationMeta, error)
	GetOrderDetails(ctx context.Context, userID, orderID string) (*homeservicedto.OrderResponse, error)
	CancelOrder(ctx context.Context, userID, orderID string) error

	// Provider - Orders
	GetProviderOrders(ctx context.Context, providerID string, query homeservicedto.ListOrdersQuery) ([]*homeservicedto.OrderListResponse, *response.PaginationMeta, error)
	RegisterProvider(ctx context.Context, userID string, req homeservicedto.RegisterProviderRequest) (*homeservicedto.ProviderProfileResponse, error)
	AcceptOrder(ctx context.Context, providerID, orderID string) error
	RejectOrder(ctx context.Context, providerID, orderID string) error
	StartOrder(ctx context.Context, providerID, orderID string) error
	CompleteOrder(ctx context.Context, providerID, orderID string) error

	// Provider Matching (async)
	FindAndNotifyNextProvider(orderID string)

	// Admin
	CreateCategory(ctx context.Context, req homeservicedto.CreateCategoryRequest) (*homeservicedto.CategoryWithTabsResponse, error)
	CreateTab(ctx context.Context, req homeservicedto.CreateTabRequest) (*homeservicedto.ServiceTabResponse, error)
	CreateService(ctx context.Context, req homeservicedto.CreateServiceRequest) (*homeservicedto.ServiceDetailResponse, error)
	UpdateService(ctx context.Context, id uint, req homeservicedto.UpdateServiceRequest) (*homeservicedto.ServiceDetailResponse, error)
	CreateAddOn(ctx context.Context, req homeservicedto.CreateAddOnRequest) (*homeservicedto.AddOnResponse, error)
}

type service struct {
	repo          Repository
	walletService wallet.Service
	cfg           *config.Config
}

func NewService(repo Repository, walletService wallet.Service, cfg *config.Config) Service {
	return &service{
		repo:          repo,
		walletService: walletService,
		cfg:           cfg,
	}
}

// --- Customer - Service Catalog ---

func (s *service) ListCategories(ctx context.Context) ([]*homeservicedto.ServiceCategoryResponse, error) {
	categories, err := s.repo.ListCategories(ctx)
	if err != nil {
		logger.Error("failed to list categories", "error", err)
		return nil, response.InternalServerError("Failed to fetch categories", err)
	}

	return homeservicedto.ToServiceCategoryList(categories), nil
}

// GetAllCategorySlugs returns all distinct category slugs from services table for dropdown
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

// RegisterProvider registers a new service provider
func (s *service) RegisterProvider(ctx context.Context, userID string, req homeservicedto.RegisterProviderRequest) (*homeservicedto.ProviderProfileResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// 1. Check if user is already a provider
	_, err := s.repo.FindProviderByUserID(ctx, userID)
	if err == nil {
		return nil, response.BadRequest("User is already registered as a service provider")
	}

	// 2. Create provider profile with category-based assignment
	// Provider is automatically assigned ALL services in their registered category
	providerID := uuid.New().String()
	provider := &models.ServiceProviderProfile{
		ID:              providerID,
		UserID:          userID,
		ServiceCategory: req.CategorySlug,
		ServiceType:     req.CategorySlug,
		Status:          models.SPStatusActive,
		IsVerified:      true, // Auto-approve for now
		IsAvailable:     true,
	}

	if err := s.repo.CreateProvider(ctx, provider); err != nil {
		logger.Error("failed to create provider profile", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to create provider profile", err)
	}

	// 3. Register provider category selection (so provider receives notifications for this category)
	if req.CategorySlug != "" {
		category := &models.ProviderServiceCategory{
			ProviderID:        providerID,
			CategorySlug:      req.CategorySlug,
			ExpertiseLevel:    "beginner",
			YearsOfExperience: 0,
			IsActive:          true,
		}
		if err := s.repo.AddProviderCategory(ctx, category); err != nil {
			// log but don't block registration
			logger.Error("failed to add provider category", "error", err, "providerID", providerID, "category", req.CategorySlug)
		}
	}

	// 4. Dynamic service assignment
	// Providers automatically get ALL services in their registered category
	// This ensures they get access to services added in the future without manual updates
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

	// 5. Fetch complete profile for response
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

// --- Customer - Orders ---
func (s *service) CreateOrder(ctx context.Context, userID string, req homeservicedto.CreateOrderRequest) (*homeservicedto.OrderResponse, error) {
	req.SetDefaults()
	// 1. Validate and set defaults
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// 2. Parse service date
	serviceDate, err := time.Parse(time.RFC3339, req.ServiceDate)
	if err != nil {
		return nil, response.BadRequest("Invalid service date format. Use RFC3339")
	}

	// Ensure service date is in the future
	if serviceDate.Before(time.Now()) {
		return nil, response.BadRequest("Service date must be in the future")
	}

	// 3. Calculate pricing for each item
	var items []models.OrderItem
	var addOns []models.OrderAddOn
	var categorySlug string // Track category from first service
	var subtotal float64
	var totalDuration int

	// Process main service items
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

		// Extract category slug from first service
		if i == 0 {
			categorySlug = svc.CategorySlug
		}

		// For the new ServiceNew model, calculate price from BasePrice
		// The new model doesn't have options like the old Service model
		price := 0.0
		duration := 0
		if svc.BasePrice != nil {
			price = *svc.BasePrice
		}
		if svc.Duration != nil {
			duration = *svc.Duration
		}

		items = append(items, models.OrderItem{
			ServiceID:       svc.ID,
			ServiceName:     svc.Title,
			BasePrice:       price,
			CalculatedPrice: price,
			DurationMinutes: duration,
			SelectedOptions: nil,
		})

		subtotal += price
		totalDuration += duration
	}

	// Process add-ons
	if len(req.AddOnIDs) > 0 {
		addOnServices, err := s.repo.GetAddOnsByIDs(ctx, req.AddOnIDs)
		if err != nil {
			return nil, response.InternalServerError("Failed to fetch add-ons", err)
		}

		for _, addon := range addOnServices {
			if !addon.IsActive {
				return nil, response.BadRequest(fmt.Sprintf("Add-on '%s' is not available", addon.Title))
			}

			addOns = append(addOns, models.OrderAddOn{
				AddOnID: addon.ID,
				Title:   addon.Title,
				Price:   addon.Price,
			})

			subtotal += addon.Price
			totalDuration += addon.DurationMinutes
		}
	}

	// 4. Calculate fees
	surgeFee := s.calculateSurgeFee(req.Latitude, req.Longitude, serviceDate)
	platformFee := subtotal * 0.10 // 10% platform fee
	discount := 0.0

	// TODO: Apply coupon if provided

	total := subtotal + surgeFee + platformFee - discount

	// 5. Create wallet hold using existing wallet service
	orderCode := s.generateOrderCode()
	holdDurationMinutes := int(HoldExpiryDuration.Minutes())

	holdReq := walletdto.HoldFundsRequest{
		Amount:        total,
		ReferenceType: "service_order",
		ReferenceID:   orderCode,
		HoldDuration:  holdDurationMinutes,
	}

	holdResp, err := s.walletService.HoldFunds(ctx, userID, holdReq)
	if err != nil {
		return nil, err
	}

	// 6. Create order
	order := &models.ServiceOrder{
		ID:             uuid.New().String(),
		Code:           orderCode,
		UserID:         userID,
		Status:         "searching_provider",
		Address:        req.Address,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		ServiceDate:    serviceDate,
		Frequency:      req.Frequency,
		QuantityOfPros: req.QuantityOfPros, // ✅ NEW
		HoursOfService: req.HoursOfService, // ✅ NEW
		CategorySlug:   categorySlug,       // ✅ Set category slug from first service
		Notes:          req.Notes,
		Subtotal:       subtotal,
		Discount:       discount,
		SurgeFee:       surgeFee,
		PlatformFee:    platformFee,
		Total:          total,
		CouponCode:     req.CouponCode,
		// Note: Using WalletHold field as temporary storage
		WalletHold: total,
		Items:      items,
		AddOns:     addOns,
	}

	if err := s.repo.CreateOrder(ctx, order); err != nil {
		// Release hold if order creation fails
		releaseReq := walletdto.ReleaseHoldRequest{HoldID: holdResp.ID}
		s.walletService.ReleaseHold(ctx, userID, releaseReq)
		logger.Error("failed to create order", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to create order", err)
	}

	logger.Info("order created", "orderID", order.ID, "userID", userID, "total", total)

	// 7. Trigger async provider search
	go s.FindAndNotifyNextProvider(order.ID)

	return homeservicedto.ToOrderResponse(order), nil
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

/* DEPRECATED: Old helper functions for legacy Service model - no longer used
// ✅ Calculate order pricing based on services, add-ons, options, quantity, and hours
func (s *service) calculateOrderPricing(
	services []*models.Service,
	addOns []*models.AddOnService,
	req homeservicedto.CreateOrderRequest,
) (subtotal float64, totalDuration int) {
	subtotal = 0.0
	totalDuration = 0

	// Calculate service prices with selected options
	for _, service := range services {
		itemPrice := service.BasePrice
		itemDuration := service.BaseDurationMinutes

		// Find matching request item for options
		var reqItem *homeservicedto.CreateOrderItemRequest
		for j := range req.Items {
			if req.Items[j].ServiceID == service.ID {
				reqItem = &req.Items[j]
				break
			}
		}

		// Apply option modifiers if present
		if reqItem != nil && len(reqItem.SelectedOptions) > 0 {
			optionPrice, optionDuration := s.calculateOptionModifiers(service, reqItem.SelectedOptions)
			itemPrice += optionPrice
			itemDuration += optionDuration
		}

		subtotal += itemPrice
		totalDuration += itemDuration

		logger.Debug("calculated service price",
			"serviceID", service.ID,
			"serviceName", service.Name,
			"basePrice", service.BasePrice,
			"finalPrice", itemPrice,
			"duration", itemDuration,
		)
	}

	// Add add-on prices
	for _, addon := range addOns {
		subtotal += addon.Price
		totalDuration += addon.DurationMinutes

		logger.Debug("added addon price",
			"addonID", addon.ID,
			"title", addon.Title,
			"price", addon.Price,
		)
	}

	return subtotal, totalDuration
}

// ✅ Calculate price and duration modifiers from selected options
func (s *service) calculateOptionModifiers(
	service *models.Service,
	selectedOptions []homeservicedto.SelectedOptionRequest,
) (float64, int) {
	priceModifier := 0.0
	durationModifier := 0

	for _, selectedOpt := range selectedOptions {
		// Find the option in service
		var option *models.ServiceOption
		for i := range service.Options {
			if service.Options[i].ID == selectedOpt.OptionID {
				option = &service.Options[i]
				break
			}
		}

		if option == nil {
			continue
		}

		// If a choice is selected, apply its modifiers
		if selectedOpt.ChoiceID != nil {
			for _, choice := range option.Choices {
				if choice.ID == *selectedOpt.ChoiceID {
					priceModifier += choice.PriceModifier
					durationModifier += choice.DurationModifierMinutes
					break
				}
			}
		}
	}

	return priceModifier, durationModifier
}

// ✅ Check if any service uses hourly pricing
func (s *service) hasHourlyPricing(services []*models.Service) bool {
	for _, svc := range services {
		if svc.PricingModel == "hourly" {
			return true
		}
	}
	return false
}


// ✅ Build order items from request
func (s *service) buildOrderItems(
	reqItems []homeservicedto.CreateOrderItemRequest,
	services []*models.Service,
) []models.OrderItem {
	items := make([]models.OrderItem, len(reqItems))

	for i, reqItem := range reqItems {
		// Find matching service
		var service *models.Service
		for _, svc := range services {
			if svc.ID == reqItem.ServiceID {
				service = svc
				break
			}
		}

		if service == nil {
			continue
		}

		// Calculate final price with options
		finalPrice := service.BasePrice
		duration := service.BaseDurationMinutes

		if len(reqItem.SelectedOptions) > 0 {
			optPrice, optDuration := s.calculateOptionModifiers(service, reqItem.SelectedOptions)
			finalPrice += optPrice
			duration += optDuration
		}

		// Convert selected options to JSON
		selectedOptionsJSON := make(map[string]interface{})
		for _, opt := range reqItem.SelectedOptions {
			selectedOptionsJSON[fmt.Sprintf("option_%d", opt.OptionID)] = map[string]interface{}{
				"optionId": opt.OptionID,
				"choiceId": opt.ChoiceID,
				"value":    opt.Value,
			}
		}

		items[i] = models.OrderItem{
			ServiceID:       service.ID,
			ServiceName:     service.Name,
			BasePrice:       service.BasePrice,
			CalculatedPrice: finalPrice,
			DurationMinutes: duration,
			SelectedOptions: selectedOptionsJSON,
		}
	}

	return items
}

// ✅ Build order add-ons from request
func (s *service) buildOrderAddOns(addOnIDs []uint, addOns []*models.AddOnService) []models.OrderAddOn {
	orderAddOns := make([]models.OrderAddOn, len(addOnIDs))

	for i, addonID := range addOnIDs {
		// Find matching add-on
		var addon *models.AddOnService
		for _, a := range addOns {
			if a.ID == addonID {
				addon = a
				break
			}
		}

		if addon == nil {
			continue
		}

		orderAddOns[i] = models.OrderAddOn{
			AddOnID: addon.ID,
			Title:   addon.Title,
			Price:   addon.Price,
		}
	}

	return orderAddOns
}
*/

// func (s *service) CreateOrder(ctx context.Context, userID string, req homeservicedto.CreateOrderRequest) (*homeservicedto.OrderResponse, error) {
// 	// 1. Validate and set defaults
// 	if err := req.Validate(); err != nil {
// 		return nil, response.BadRequest(err.Error())
// 	}
// 	req.SetDefaults()

// 	// 2. Parse service date
// 	serviceDate, err := time.Parse(time.RFC3339, req.ServiceDate)
// 	if err != nil {
// 		return nil, response.BadRequest("Invalid service date format. Use RFC3339")
// 	}

// 	// Ensure service date is in the future
// 	if serviceDate.Before(time.Now()) {
// 		return nil, response.BadRequest("Service date must be in the future")
// 	}

// 	// 3. Calculate pricing for each item
// 	var items []models.OrderItem
// 	var addOns []models.OrderAddOn
// 	subtotal := 0.0
// 	totalDuration := 0

// 	// Process main service items
// 	for _, itemReq := range req.Items {
// 		svc, err := s.repo.GetServiceWithOptions(ctx, itemReq.ServiceID)
// 		if err != nil {
// 			if err == gorm.ErrRecordNotFound {
// 				return nil, response.BadRequest(fmt.Sprintf("Service with ID %d not found", itemReq.ServiceID))
// 			}
// 			return nil, response.InternalServerError("Failed to fetch service", err)
// 		}

// 		if !svc.IsActive {
// 			return nil, response.BadRequest(fmt.Sprintf("Service '%s' is not available", svc.Name))
// 		}

// 		// Calculate price and duration based on selected options
// 		price, duration, selectedOpts, err := s.calculateItemPrice(svc, itemReq.SelectedOptions)
// 		if err != nil {
// 			return nil, response.BadRequest(err.Error())
// 		}

// 		items = append(items, models.OrderItem{
// 			ServiceID:       svc.ID,
// 			ServiceName:     svc.Name,
// 			BasePrice:       svc.BasePrice,
// 			CalculatedPrice: price,
// 			DurationMinutes: duration,
// 			SelectedOptions: selectedOpts,
// 		})

// 		subtotal += price
// 		totalDuration += duration
// 	}

// 	// Process add-ons
// 	if len(req.AddOnIDs) > 0 {
// 		addOnServices, err := s.repo.GetAddOnsByIDs(ctx, req.AddOnIDs)
// 		if err != nil {
// 			return nil, response.InternalServerError("Failed to fetch add-ons", err)
// 		}

// 		for _, addon := range addOnServices {
// 			if !addon.IsActive {
// 				return nil, response.BadRequest(fmt.Sprintf("Add-on '%s' is not available", addon.Title))
// 			}

// 			addOns = append(addOns, models.OrderAddOn{
// 				AddOnID: addon.ID,
// 				Title:   addon.Title,
// 				Price:   addon.Price,
// 			})

// 			subtotal += addon.Price
// 			totalDuration += addon.DurationMinutes
// 		}
// 	}

// 	// 4. Calculate fees
// 	surgeFee := s.calculateSurgeFee(req.Latitude, req.Longitude, serviceDate)
// 	platformFee := subtotal * 0.10 // 10% platform fee
// 	discount := 0.0

// 	// TODO: Apply coupon if provided

// 	total := subtotal + surgeFee + platformFee - discount

// 	// 5. Create wallet hold using existing wallet service
// 	orderCode := s.generateOrderCode()
// 	holdDurationMinutes := int(HoldExpiryDuration.Minutes())

// 	holdReq := wallethomeservicedto.HoldFundsRequest{
// 		Amount:        total,
// 		ReferenceType: "service_order",
// 		ReferenceID:   orderCode,
// 		HoldDuration:  holdDurationMinutes,
// 	}

// 	holdResp, err := s.walletService.HoldFunds(ctx, userID, holdReq)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 6. Create order
// 	order := &models.ServiceOrder{
// 		ID:          uuid.New().String(),
// 		Code:        orderCode,
// 		UserID:      userID,
// 		Status:      "searching_provider",
// 		Address:     req.Address,
// 		ServiceDate: serviceDate,
// 		Frequency:   req.Frequency,
// 		Notes:       req.Notes,
// 		Subtotal:    subtotal,
// 		Discount:    discount,
// 		SurgeFee:    surgeFee,
// 		PlatformFee: platformFee,
// 		Total:       total,
// 		CouponCode:  req.CouponCode,
// 		WalletHoldID:  &holdResp.ID,
// 		Items:       items,
// 		AddOns:      addOns,
// 	}

// 	if err := s.repo.CreateOrder(ctx, order); err != nil {
// 		// Release hold if order creation fails
// 		releaseReq := wallethomeservicedto.ReleaseHoldRequest{HoldID: holdResp.ID}
// 		s.walletService.ReleaseHold(ctx, userID, releaseReq)
// 		logger.Error("failed to create order", "error", err, "userID", userID)
// 		return nil, response.InternalServerError("Failed to create order", err)
// 	}

// 	logger.Info("order created", "orderID", order.ID, "userID", userID, "total", total)

// 	// 7. Trigger async provider search
// 	go s.FindAndNotifyNextProvider(order.ID)

// 	return homeservicedto.ToOrderResponse(order), nil
// }

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
	order, err := s.repo.GetOrderByIDWithDetails(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to fetch order", err)
	}

	// Verify ownership
	if order.UserID != userID {
		return nil, response.ForbiddenError("You don't have access to this order")
	}

	return homeservicedto.ToOrderResponse(order), nil
}

// func (s *service) GetOrderDetails(ctx context.Context, userID, orderID string) (*homeservicedto.OrderResponse, error) {
// 	order, err := s.repo.GetOrderByIDWithDetails(ctx, orderID)
// 	if err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, response.NotFoundError("Order")
// 		}
// 		return nil, response.InternalServerError("Failed to fetch order", err)
// 	}

// 	// Verify ownership
// 	if order.UserID != userID {
// 		return nil, response.ForbiddenError("You don't have access to this order")
// 	}

// 	return homeservicedto.ToOrderResponse(order), nil
// }

func (s *service) CancelOrder(ctx context.Context, userID, orderID string) error {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.NotFoundError("Order")
		}
		return response.InternalServerError("Failed to fetch order", err)
	}

	// Verify ownership
	if order.UserID != userID {
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
		if err := s.walletService.ReleaseHold(ctx, userID, releaseReq); err != nil {
			logger.Error("failed to release hold on order cancellation", "error", err, "orderID", orderID)
		}
	}

	// Update order status
	if err := s.repo.UpdateOrderStatus(ctx, orderID, "cancelled"); err != nil {
		return response.InternalServerError("Failed to cancel order", err)
	}

	logger.Info("order cancelled", "orderID", orderID, "userID", userID)

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

	if order.ProviderID == nil || *order.ProviderID != providerID {
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

	if order.ProviderID == nil || *order.ProviderID != providerID {
		return response.ForbiddenError("You are not assigned to this order")
	}

	if order.Status != "in_progress" {
		return response.BadRequest("Order must be in progress to complete")
	}

	// 1. Capture the wallet hold using existing wallet service
	if order.WalletHoldID != nil {
		captureReq := walletdto.CaptureHoldRequest{
			HoldID:      *order.WalletHoldID,
			Description: fmt.Sprintf("Payment for order %s", order.Code),
		}
		if _, err := s.walletService.CaptureHold(ctx, order.UserID, captureReq); err != nil {
			logger.Error("failed to capture hold", "error", err, "orderID", orderID)
			return response.InternalServerError("Payment processing failed", err)
		}
	}

	// 2. Transfer funds to provider using existing wallet service
	provider, err := s.repo.GetProviderByID(ctx, providerID)
	if err == nil && provider != nil {
		providerAmount := order.Total - order.PlatformFee
		transferReq := walletdto.TransferFundsRequest{
			RecipientID: provider.UserID, // Use provider's UserID for wallet transfer
			Amount:      providerAmount,
			Description: fmt.Sprintf("Earnings from order %s", order.Code),
		}
		if _, err := s.walletService.TransferFunds(ctx, order.UserID, transferReq); err != nil {
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

// func (s *service) CompleteOrder(ctx context.Context, providerID, orderID string) error {
// 	order, err := s.repo.GetOrderByID(ctx, orderID)
// 	if err != nil {
// 		return response.NotFoundError("Order")
// 	}

// 	if order.ProviderID == nil || *order.ProviderID != providerID {
// 		return response.ForbiddenError("You are not assigned to this order")
// 	}

// 	if order.Status != "in_progress" {
// 		return response.BadRequest("Order must be in progress to complete")
// 	}

// 	// 1. Capture the wallet hold using existing wallet service
// 	if order.WalletHoldID != nil {
// 		captureReq := wallethomeservicedto.CaptureHoldRequest{
// 			HoldID:      *order.WalletHoldID,
// 			Description: fmt.Sprintf("Payment for order %s", order.Code),
// 		}
// 		if _, err := s.walletService.CaptureHold(ctx, order.UserID, captureReq); err != nil {
// 			logger.Error("failed to capture hold", "error", err, "orderID", orderID)
// 			return response.InternalServerError("Payment processing failed", err)
// 		}
// 	}

// 	// 2. Transfer funds to provider using existing wallet service
// 	provider, _ := s.repo.GetProviderByID(ctx, providerID)
// 	if provider != nil {
// 		providerAmount := order.Total - order.PlatformFee
// 		transferReq := wallethomeservicedto.TransferFundsRequest{
// 			RecipientID: provider.UserID,
// 			Amount:      providerAmount,
// 			Description: fmt.Sprintf("Earnings from order %s", order.Code),
// 		}
// 		if _, err := s.walletService.TransferFunds(ctx, order.UserID, transferReq); err != nil {
// 			logger.Error("failed to transfer to provider", "error", err, "providerID", providerID)
// 			// Don't fail the completion, but log for manual reconciliation
// 		}
// 	}

// 	// 3. Update order status
// 	if err := s.repo.UpdateOrderStatus(ctx, orderID, "completed"); err != nil {
// 		return response.InternalServerError("Failed to complete order", err)
// 	}

// 	// 4. Update provider status back to available
// 	s.repo.UpdateProviderStatus(ctx, providerID, "available")

// 	logger.Info("order completed", "providerID", providerID, "orderID", orderID)

// 	// TODO: Send notification to customer for rating

// 	return nil
// }

// --- Provider Matching Logic ---

func (s *service) FindAndNotifyNextProvider(orderID string) {
	ctx := context.Background()

	// 2. Fetch order
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		logger.Error("failed to fetch order for matching", "error", err, "orderID", orderID)
		return
	}

	// 3. Get service IDs from order items
	var serviceIDs []string
	fullOrder, _ := s.repo.GetOrderByIDWithDetails(ctx, orderID)
	if fullOrder != nil {
		for _, item := range fullOrder.Items {
			serviceIDs = append(serviceIDs, item.ServiceID)
		}
	}

	if len(serviceIDs) == 0 {
		logger.Error("no service IDs found for order", "orderID", orderID)
		return
	}

	// Get order's category slug for filtering providers
	orderCategorySlug := order.CategorySlug
	if orderCategorySlug == "" {
		logger.Error("order has no category slug", "orderID", orderID)
		return
	}

	// 5. Find providers with this category (category-based matching)
	// The provider matching is now handled through category slugs
	// Providers register with categories and see orders with matching categories
	// This is managed through the provider app's GetAvailableOrders endpoint

	// For now, just update order status to indicate it's available for providers
	if err := s.repo.UpdateOrderStatus(ctx, orderID, "pending"); err != nil {
		logger.Error("failed to update order status", "error", err, "orderID", orderID)
		// Release wallet hold on error
		if order.WalletHoldID != nil {
			releaseReq := walletdto.ReleaseHoldRequest{HoldID: *order.WalletHoldID}
			s.walletService.ReleaseHold(ctx, order.UserID, releaseReq)
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
