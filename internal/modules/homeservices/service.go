package homeservices

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/config"
	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/dto"
	"github.com/umar5678/go-backend/internal/modules/wallet"
	walletDTO "github.com/umar5678/go-backend/internal/modules/wallet/dto"
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
	ListCategories(ctx context.Context) ([]*dto.ServiceCategoryResponse, error)
	ListServices(ctx context.Context, query dto.ListServicesQuery) ([]*dto.ServiceListResponse, *response.PaginationMeta, error)
	GetServiceDetails(ctx context.Context, id uint) (*dto.ServiceResponse, error)

	// Customer - Orders
	CreateOrder(ctx context.Context, userID string, req dto.CreateOrderRequest) (*dto.OrderResponse, error)
	GetMyOrders(ctx context.Context, userID string, query dto.ListOrdersQuery) ([]*dto.OrderListResponse, *response.PaginationMeta, error)
	GetOrderDetails(ctx context.Context, userID, orderID string) (*dto.OrderResponse, error)
	CancelOrder(ctx context.Context, userID, orderID string) error

	// Provider - Orders
	GetProviderOrders(ctx context.Context, providerID string, query dto.ListOrdersQuery) ([]*dto.OrderListResponse, *response.PaginationMeta, error)
	AcceptOrder(ctx context.Context, providerID, orderID string) error
	RejectOrder(ctx context.Context, providerID, orderID string) error
	StartOrder(ctx context.Context, providerID, orderID string) error
	CompleteOrder(ctx context.Context, providerID, orderID string) error

	// Provider Matching (async)
	FindAndNotifyNextProvider(orderID string)

	// Admin
	CreateService(ctx context.Context, req dto.CreateServiceRequest) (*dto.ServiceResponse, error)
	UpdateService(ctx context.Context, id uint, req dto.UpdateServiceRequest) (*dto.ServiceResponse, error)
}

type service struct {
	repo          Repository
	walletService wallet.Service
	cfg           *config.Config
}

// Updated NewService - removed cache.Service parameter
func NewService(repo Repository, walletService wallet.Service, cfg *config.Config) Service {
	return &service{
		repo:          repo,
		walletService: walletService,
		cfg:           cfg,
	}
}

// --- Customer - Service Catalog ---

func (s *service) ListCategories(ctx context.Context) ([]*dto.ServiceCategoryResponse, error) {
	categories, err := s.repo.ListCategories(ctx)
	if err != nil {
		logger.Error("failed to list categories", "error", err)
		return nil, response.InternalServerError("Failed to fetch categories", err)
	}

	return dto.ToServiceCategoryList(categories), nil
}

func (s *service) ListServices(ctx context.Context, query dto.ListServicesQuery) ([]*dto.ServiceListResponse, *response.PaginationMeta, error) {
	query.SetDefaults()

	services, total, err := s.repo.ListServices(ctx, query)
	if err != nil {
		logger.Error("failed to list services", "error", err)
		return nil, nil, response.InternalServerError("Failed to fetch services", err)
	}

	responses := dto.ToServiceListResponses(services)
	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

func (s *service) GetServiceDetails(ctx context.Context, id uint) (*dto.ServiceResponse, error) {
	service, err := s.repo.GetServiceWithOptions(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Service")
		}
		logger.Error("failed to get service", "error", err, "serviceID", id)
		return nil, response.InternalServerError("Failed to fetch service", err)
	}

	return dto.ToServiceResponse(service), nil
}

// --- Customer - Orders ---

func (s *service) CreateOrder(ctx context.Context, userID string, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	// 1. Validate and set defaults
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}
	req.SetDefaults()

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
	subtotal := 0.0
	totalDuration := 0

	for _, itemReq := range req.Items {
		// Fetch service from DB for security (never trust client pricing)
		svc, err := s.repo.GetServiceWithOptions(ctx, itemReq.ServiceID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, response.BadRequest(fmt.Sprintf("Service with ID %d not found", itemReq.ServiceID))
			}
			return nil, response.InternalServerError("Failed to fetch service", err)
		}

		if !svc.IsActive {
			return nil, response.BadRequest(fmt.Sprintf("Service '%s' is not available", svc.Name))
		}

		// Calculate price and duration based on selected options
		price, duration, selectedOpts, err := s.calculateItemPrice(svc, itemReq.SelectedOptions)
		if err != nil {
			return nil, response.BadRequest(err.Error())
		}

		items = append(items, models.OrderItem{
			ServiceID:       svc.ID,
			ServiceName:     svc.Name,
			BasePrice:       svc.BasePrice,
			CalculatedPrice: price,
			DurationMinutes: duration,
			SelectedOptions: selectedOpts,
		})

		subtotal += price
		totalDuration += duration
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

	holdReq := walletDTO.HoldFundsRequest{
		Amount:        total,
		ReferenceType: "service_order",
		ReferenceID:   orderCode,
		HoldDuration:  holdDurationMinutes,
	}

	holdResp, err := s.walletService.HoldFunds(ctx, userID, holdReq)
	if err != nil {
		// This returns "Insufficient funds" if balance is low
		return nil, err
	}

	// 6. Create order
	order := &models.ServiceOrder{
		ID:           uuid.New().String(),
		Code:         orderCode,
		UserID:       userID,
		Status:       "searching_provider",
		Address:      req.Address,
		ServiceDate:  serviceDate,
		Frequency:    req.Frequency,
		Notes:        req.Notes,
		Subtotal:     subtotal,
		Discount:     discount,
		SurgeFee:     surgeFee,
		PlatformFee:  platformFee,
		Total:        total,
		CouponCode:   req.CouponCode,
		WalletHoldID: &holdResp.ID,
		Items:        items,
	}

	if err := s.repo.CreateOrder(ctx, order, req.Latitude, req.Longitude); err != nil {
		// Release hold if order creation fails
		releaseReq := walletDTO.ReleaseHoldRequest{HoldID: holdResp.ID}
		s.walletService.ReleaseHold(ctx, userID, releaseReq)
		logger.Error("failed to create order", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to create order", err)
	}

	logger.Info("order created", "orderID", order.ID, "userID", userID, "total", total)

	// 7. Trigger async provider search
	go s.FindAndNotifyNextProvider(order.ID)

	return dto.ToOrderResponse(order), nil
}

func (s *service) GetMyOrders(ctx context.Context, userID string, query dto.ListOrdersQuery) ([]*dto.OrderListResponse, *response.PaginationMeta, error) {
	query.SetDefaults()

	orders, total, err := s.repo.ListUserOrders(ctx, userID, query)
	if err != nil {
		logger.Error("failed to list user orders", "error", err, "userID", userID)
		return nil, nil, response.InternalServerError("Failed to fetch orders", err)
	}

	responses := dto.ToOrderListResponses(orders)
	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

func (s *service) GetOrderDetails(ctx context.Context, userID, orderID string) (*dto.OrderResponse, error) {
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

	return dto.ToOrderResponse(order), nil
}

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
		releaseReq := walletDTO.ReleaseHoldRequest{
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

func (s *service) GetProviderOrders(ctx context.Context, providerID string, query dto.ListOrdersQuery) ([]*dto.OrderListResponse, *response.PaginationMeta, error) {
	query.SetDefaults()

	orders, total, err := s.repo.ListProviderOrders(ctx, providerID, query)
	if err != nil {
		logger.Error("failed to list provider orders", "error", err, "providerID", providerID)
		return nil, nil, response.InternalServerError("Failed to fetch orders", err)
	}

	responses := dto.ToOrderListResponses(orders)
	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

func (s *service) AcceptOrder(ctx context.Context, providerID, orderID string) error {
	offerKey := fmt.Sprintf("provider:%s:current_offer", providerID)

	// 1. Verify the offer is still valid - Using direct cache function
	val, err := cache.Get(ctx, offerKey)
	if err != nil || val != orderID {
		return response.ForbiddenError("Offer expired or invalid")
	}

	// 2. Assign provider to order
	if err := s.repo.AssignProviderToOrder(ctx, providerID, orderID); err != nil {
		logger.Error("failed to assign provider to order", "error", err, "providerID", providerID, "orderID", orderID)
		return response.InternalServerError("Failed to accept order", err)
	}

	// 3. Delete offer key to prevent timeout logic - Using direct cache function
	cache.Delete(ctx, offerKey)

	// 4. Update provider status
	s.repo.UpdateProviderStatus(ctx, providerID, "busy")

	logger.Info("provider accepted order", "providerID", providerID, "orderID", orderID)

	// TODO: Send notification to customer

	return nil
}

func (s *service) RejectOrder(ctx context.Context, providerID, orderID string) error {
	offerKey := fmt.Sprintf("provider:%s:current_offer", providerID)

	// Verify the offer exists - Using direct cache function
	val, err := cache.Get(ctx, offerKey)
	if err != nil || val != orderID {
		return response.ForbiddenError("No active offer for this order")
	}

	// Delete the offer - Using direct cache function
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
		captureReq := walletDTO.CaptureHoldRequest{
			HoldID:      *order.WalletHoldID,
			Description: fmt.Sprintf("Payment for order %s", order.Code),
		}
		if _, err := s.walletService.CaptureHold(ctx, order.UserID, captureReq); err != nil {
			logger.Error("failed to capture hold", "error", err, "orderID", orderID)
			return response.InternalServerError("Payment processing failed", err)
		}
	}

	// 2. Transfer funds to provider using existing wallet service
	provider, _ := s.repo.GetProviderByID(ctx, providerID)
	if provider != nil {
		providerAmount := order.Total - order.PlatformFee
		transferReq := walletDTO.TransferFundsRequest{
			RecipientID: provider.UserID,
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

	// TODO: Send notification to customer for rating

	return nil
}

// --- Provider Matching Logic ---

func (s *service) FindAndNotifyNextProvider(orderID string) {
	ctx := context.Background()

	// 1. Get rejected provider IDs from Redis - Using direct cache function
	rejectedKey := fmt.Sprintf("order:%s:rejected_providers", orderID)
	rejectedIDsStr, _ := cache.Get(ctx, rejectedKey)
	var rejectedIDs []string
	if rejectedIDsStr != "" {
		rejectedIDs = parseCommaSeparated(rejectedIDsStr)
	}

	// 2. Fetch order
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		logger.Error("failed to fetch order for matching", "error", err, "orderID", orderID)
		return
	}

	// 3. Get service IDs from order items
	var serviceIDs []uint
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

	// 4. Get lat/lon from order (requires raw query since it's stored in PostGIS)
	var lat, lon float64
	s.repo.(*repository).db.Raw(`
		SELECT ST_Y(location::geometry) as lat, ST_X(location::geometry) as lon
		FROM service_orders WHERE id = ?
	`, orderID).Scan(&lat).Scan(&lon)

	// 5. Find nearest qualified providers
	providers, err := s.repo.FindNearestAvailableProviders(ctx, serviceIDs, lat, lon, DefaultSearchRadius)
	if err != nil || len(providers) == 0 {
		logger.Warn("no providers found", "orderID", orderID, "serviceIDs", serviceIDs)
		s.repo.UpdateOrderStatus(ctx, orderID, "no_provider_available")
		// Release wallet hold
		if order.WalletHoldID != nil {
			releaseReq := walletDTO.ReleaseHoldRequest{HoldID: *order.WalletHoldID}
			s.walletService.ReleaseHold(ctx, order.UserID, releaseReq)
		}
		return
	}

	// 6. Find next provider who hasn't been tried
	var nextProvider *models.ServiceProvider
	for i := range providers {
		if !contains(rejectedIDs, providers[i].ID) {
			nextProvider = &providers[i]
			break
		}
	}

	if nextProvider == nil {
		logger.Warn("all providers exhausted", "orderID", orderID)
		s.repo.UpdateOrderStatus(ctx, orderID, "no_provider_available")
		if order.WalletHoldID != nil {
			releaseReq := walletDTO.ReleaseHoldRequest{HoldID: *order.WalletHoldID}
			s.walletService.ReleaseHold(ctx, order.UserID, releaseReq)
		}
		return
	}

	// 7. Offer job to provider - Using direct cache functions
	offerKey := fmt.Sprintf("provider:%s:current_offer", nextProvider.ID)
	cache.Set(ctx, offerKey, orderID, ProviderOfferTimeout)

	// Mark provider as tried
	rejectedIDs = append(rejectedIDs, nextProvider.ID)
	cache.Set(ctx, rejectedKey, joinCommaSeparated(rejectedIDs), 24*time.Hour)

	logger.Info("offering order to provider", "orderID", orderID, "providerID", nextProvider.ID)

	// TODO: Send push notification to provider

	// 8. Schedule timeout check
	time.AfterFunc(ProviderOfferTimeout, func() {
		s.checkOfferTimeout(orderID, nextProvider.ID)
	})
}

func (s *service) checkOfferTimeout(orderID, providerID string) {
	ctx := context.Background()
	offerKey := fmt.Sprintf("provider:%s:current_offer", providerID)

	// Using direct cache function
	val, err := cache.Get(ctx, offerKey)
	if err == nil && val == orderID {
		// Provider timed out, find next one
		logger.Info("provider offer timed out", "orderID", orderID, "providerID", providerID)
		s.FindAndNotifyNextProvider(orderID)
	}
}

// --- Admin - Service Management ---

func (s *service) CreateService(ctx context.Context, req dto.CreateServiceRequest) (*dto.ServiceResponse, error) {
	// Verify category exists
	_, err := s.repo.GetCategoryByID(ctx, req.CategoryID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.BadRequest("Category not found")
		}
		return nil, response.InternalServerError("Failed to verify category", err)
	}

	service := &models.Service{
		CategoryID:          req.CategoryID,
		Name:                req.Name,
		Description:         req.Description,
		ImageURL:            req.ImageURL,
		BasePrice:           req.BasePrice,
		PricingModel:        req.PricingModel,
		BaseDurationMinutes: req.BaseDurationMinutes,
		IsActive:            true,
	}

	if err := s.repo.CreateService(ctx, service); err != nil {
		logger.Error("failed to create service", "error", err)
		return nil, response.InternalServerError("Failed to create service", err)
	}

	logger.Info("service created", "serviceID", service.ID, "name", service.Name)

	return dto.ToServiceResponse(service), nil
}

func (s *service) UpdateService(ctx context.Context, id uint, req dto.UpdateServiceRequest) (*dto.ServiceResponse, error) {
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

	return dto.ToServiceResponse(service), nil
}

// --- Helper Functions ---

func (s *service) calculateItemPrice(svc *models.Service, selectedOptions []dto.SelectedOptionRequest) (float64, int, models.JSONBMap, error) {
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

func (s *service) generateOrderCode() string {
	return fmt.Sprintf("HS%d", time.Now().Unix())
}

// Utility functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func parseCommaSeparated(s string) []string {
	if s == "" {
		return []string{}
	}
	return []string{s}
}

func joinCommaSeparated(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	return slice[len(slice)-1]
}
