package customer

import (
	"context"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/modules/homeservices/customer/dto"
	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type CategoryConfig struct {
	Slug        string
	Title       string
	Description string
	Icon        string
	Image       string
	SortOrder   int
}

func GetCategoryConfigs() map[string]CategoryConfig {
	return map[string]CategoryConfig{
		"pest-control": {
			Slug:        "pest-control",
			Title:       "Pest Control",
			Description: "Professional pest control services for your home and office",
			Icon:        "pest-icon",
			Image:       "https://example.com/images/pest-control.jpg",
			SortOrder:   1,
		},
		"cleaning": {
			Slug:        "cleaning",
			Title:       "Cleaning Services",
			Description: "Professional cleaning services for homes and offices",
			Icon:        "cleaning-icon",
			Image:       "https://example.com/images/cleaning.jpg",
			SortOrder:   2,
		},
		"iv-therapy": {
			Slug:        "iv-therapy",
			Title:       "IV Therapy",
			Description: "Professional IV therapy and wellness treatments",
			Icon:        "iv-icon",
			Image:       "https://example.com/images/iv-therapy.jpg",
			SortOrder:   3,
		},
		"massage": {
			Slug:        "massage",
			Title:       "Massage Therapy",
			Description: "Professional massage and relaxation services",
			Icon:        "massage-icon",
			Image:       "https://example.com/images/massage.jpg",
			SortOrder:   4,
		},
		"handyman": {
			Slug:        "handyman",
			Title:       "Handyman Services",
			Description: "Professional handyman services for repairs and maintenance",
			Icon:        "handyman-icon",
			Image:       "https://example.com/images/handyman.jpg",
			SortOrder:   5,
		},
	}
}

type Service interface {
	GetAllCategories(ctx context.Context) (*dto.CategoryListResponse, error)
	GetCategoryDetail(ctx context.Context, categorySlug string) (*dto.CategoryDetailResponse, error)

	GetServiceBySlug(ctx context.Context, slug string) (*dto.ServiceDetailResponse, error)
	ListServices(ctx context.Context, query dto.ListServicesQuery) ([]dto.ServiceListResponse, *response.PaginationMeta, error)
	GetFrequentServices(ctx context.Context, limit int) ([]dto.ServiceListResponse, error)

	GetAddonBySlug(ctx context.Context, slug string) (*dto.AddonResponse, error)
	ListAddons(ctx context.Context, query dto.ListAddonsQuery) ([]dto.AddonListResponse, *response.PaginationMeta, error)
	GetDiscountedAddons(ctx context.Context, limit int) ([]dto.AddonListResponse, error)

	Search(ctx context.Context, query dto.SearchQuery) (*dto.SearchResponse, error)

	CreateOrder(ctx context.Context, customerID string, req dto.CreateOrderRequest) (*dto.OrderCreatedResponse, error)
	GetOrder(ctx context.Context, customerID, orderID string) (*dto.OrderResponse, error)
	ListOrders(ctx context.Context, customerID string, query dto.ListOrdersQuery) ([]dto.OrderListResponse, *response.PaginationMeta, error)

	GetCancellationPreview(ctx context.Context, customerID, orderID string) (*dto.CancellationPreviewResponse, error)
	CancelOrder(ctx context.Context, customerID, orderID string, req dto.CancelOrderRequest) (*dto.OrderResponse, error)

	RateOrder(ctx context.Context, customerID, orderID string, req dto.RateOrderRequest) (*dto.OrderResponse, error)
}

type service struct {
	repo          Repository
	serviceRepo   Repository
	walletService WalletService
}

func NewService(repo Repository, serviceRepo Repository, walletService WalletService) Service {
	return &service{
		repo:          repo,
		serviceRepo:   serviceRepo,
		walletService: walletService,
	}
}

func (s *service) GetAllCategories(ctx context.Context) (*dto.CategoryListResponse, error) {

	categoryInfos, err := s.repo.GetAllActiveCategories(ctx)
	if err != nil {
		logger.Error("failed to get categories", "error", err)
		return nil, response.InternalServerError("Failed to get categories", err)
	}

	configs := GetCategoryConfigs()

	var categories []dto.CategoryResponse
	for _, info := range categoryInfos {
		config, exists := configs[info.Slug]
		if !exists {

			config = CategoryConfig{
				Slug:  info.Slug,
				Title: info.Slug,
			}
		}

		categories = append(categories, dto.CategoryResponse{
			Slug:         info.Slug,
			Title:        config.Title,
			Description:  config.Description,
			Icon:         config.Icon,
			Image:        config.Image,
			ServiceCount: int(info.ServiceCount),
			AddonCount:   int(info.AddonCount),
		})
	}

	return &dto.CategoryListResponse{
		Categories: categories,
		Total:      len(categories),
	}, nil
}

func (s *service) GetCategoryDetail(ctx context.Context, categorySlug string) (*dto.CategoryDetailResponse, error) {

	services, err := s.repo.GetActiveServicesByCategory(ctx, categorySlug)
	if err != nil {
		logger.Error("failed to get services by category", "error", err, "category", categorySlug)
		return nil, response.InternalServerError("Failed to get category details", err)
	}

	addons, err := s.repo.GetActiveAddonsByCategory(ctx, categorySlug)
	if err != nil {
		logger.Error("failed to get addons by category", "error", err, "category", categorySlug)
		return nil, response.InternalServerError("Failed to get category details", err)
	}

	if len(services) == 0 && len(addons) == 0 {
		return nil, response.NotFoundError("Category")
	}

	configs := GetCategoryConfigs()
	config, exists := configs[categorySlug]
	if !exists {
		config = CategoryConfig{
			Slug:  categorySlug,
			Title: categorySlug,
		}
	}

	return &dto.CategoryDetailResponse{
		Slug:        categorySlug,
		Title:       config.Title,
		Description: config.Description,
		Icon:        config.Icon,
		Image:       config.Image,
		Services:    dto.ToServiceResponses(services),
		Addons:      dto.ToAddonResponses(addons),
	}, nil
}

func (s *service) GetServiceBySlug(ctx context.Context, slug string) (*dto.ServiceDetailResponse, error) {

	svc, err := s.repo.GetActiveServiceBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Service")
		}
		logger.Error("failed to get service", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to get service", err)
	}
	addons, err := s.repo.GetActiveAddonsByCategory(ctx, svc.CategorySlug)
	if err != nil {
		logger.Error("failed to get related addons", "error", err, "category", svc.CategorySlug)
		addons = []*models.Addon{}
	}

	return &dto.ServiceDetailResponse{
		Service: dto.ToServiceResponse(svc),
		Addons:  dto.ToAddonResponses(addons),
	}, nil
}

func (s *service) ListServices(ctx context.Context, query dto.ListServicesQuery) ([]dto.ServiceListResponse, *response.PaginationMeta, error) {

	query.SetDefaults()

	services, total, err := s.repo.ListActiveServices(ctx, query)
	if err != nil {
		logger.Error("failed to list services", "error", err)
		return nil, nil, response.InternalServerError("Failed to list services", err)
	}

	responses := dto.ToServiceListResponses(services)

	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

func (s *service) GetFrequentServices(ctx context.Context, limit int) ([]dto.ServiceListResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	services, err := s.repo.GetFrequentServices(ctx, limit)
	if err != nil {
		logger.Error("failed to get frequent services", "error", err)
		return nil, response.InternalServerError("Failed to get frequent services", err)
	}

	return dto.ToServiceListResponses(services), nil
}

func (s *service) GetAddonBySlug(ctx context.Context, slug string) (*dto.AddonResponse, error) {
	addon, err := s.repo.GetActiveAddonBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Addon")
		}
		logger.Error("failed to get addon", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to get addon", err)
	}

	result := dto.ToAddonResponse(addon)
	return &result, nil
}

func (s *service) ListAddons(ctx context.Context, query dto.ListAddonsQuery) ([]dto.AddonListResponse, *response.PaginationMeta, error) {
	query.SetDefaults()
	addons, total, err := s.repo.ListActiveAddons(ctx, query)
	if err != nil {
		logger.Error("failed to list addons", "error", err)
		return nil, nil, response.InternalServerError("Failed to list addons", err)
	}

	responses := dto.ToAddonListResponses(addons)

	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

func (s *service) GetDiscountedAddons(ctx context.Context, limit int) ([]dto.AddonListResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	addons, err := s.repo.GetDiscountedAddons(ctx, limit)
	if err != nil {
		logger.Error("failed to get discounted addons", "error", err)
		return nil, response.InternalServerError("Failed to get discounted addons", err)
	}

	return dto.ToAddonListResponses(addons), nil
}

func (s *service) Search(ctx context.Context, query dto.SearchQuery) (*dto.SearchResponse, error) {
	query.SetDefaults()

	limitPerType := query.Limit / 2
	if limitPerType < 5 {
		limitPerType = 5
	}

	services, err := s.repo.SearchServices(ctx, query.Query, query.CategorySlug, limitPerType)
	if err != nil {
		logger.Error("failed to search services", "error", err, "query", query.Query)
		return nil, response.InternalServerError("Failed to search", err)
	}

	addons, err := s.repo.SearchAddons(ctx, query.Query, query.CategorySlug, limitPerType)
	if err != nil {
		logger.Error("failed to search addons", "error", err, "query", query.Query)
		return nil, response.InternalServerError("Failed to search", err)
	}

	var results []dto.SearchResultItem

	for _, svc := range services {
		results = append(results, dto.ToSearchResultFromService(svc))
	}

	for _, addon := range addons {
		results = append(results, dto.ToSearchResultFromAddon(addon))
	}

	return &dto.SearchResponse{
		Query:   query.Query,
		Results: results,
		Total:   len(results),
	}, nil
}

func (s *service) CreateOrder(ctx context.Context, customerID string, req dto.CreateOrderRequest) (*dto.OrderCreatedResponse, error) {

	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	activeCount, err := s.repo.CountCustomerActiveOrders(ctx, customerID)
	if err != nil {
		logger.Error("failed to count active orders", "error", err, "customerID", customerID)
		return nil, response.InternalServerError("Failed to create order", err)
	}
	if activeCount >= 5 {
		return nil, response.BadRequest("You have too many active orders. Please wait for some to complete before booking again.")
	}

	servicesTotal, selectedServices, err := s.validateAndCalculateServices(ctx, req.CategorySlug, req.SelectedServices)
	if err != nil {
		return nil, err
	}

	addonsTotal, selectedAddons, err := s.validateAndCalculateAddons(ctx, req.CategorySlug, req.SelectedAddons)
	if err != nil {
		return nil, err
	}

	subtotal := servicesTotal + addonsTotal
	platformCommission := shared.CalculatePlatformCommission(subtotal)
	totalPrice := subtotal

	if req.PaymentMethod == "wallet" {
		balance, err := s.walletService.GetBalance(ctx, customerID)
		if err != nil {
			logger.Error("failed to get wallet balance", "error", err, "customerID", customerID)
			return nil, response.InternalServerError("Failed to check wallet balance", err)
		}
		if balance < totalPrice {
			return nil, response.BadRequest(fmt.Sprintf("Insufficient wallet balance. Required: $%.2f, Available: $%.2f", totalPrice, balance))
		}
	}

	var preferredTime time.Time
	if req.BookingInfo.PreferredTime != "" {
		pt, err := time.Parse(time.RFC3339, req.BookingInfo.PreferredTime)
		if err != nil {
			pt, err = time.Parse("15:04", req.BookingInfo.PreferredTime)
			if err != nil {
				return nil, response.BadRequest("invalid preferred time format; expected RFC3339 or 15:04")
			}
		}
		preferredTime = pt
	}

	order := &models.ServiceOrderNew{
		OrderNumber: shared.GenerateOrderNumber(),
		CustomerID:  customerID,
		CustomerInfo: models.CustomerInfo{
			Name:    req.CustomerInfo.Name,
			Phone:   req.CustomerInfo.Phone,
			Email:   req.CustomerInfo.Email,
			Address: req.CustomerInfo.Address,
			Lat:     req.CustomerInfo.Lat,
			Lng:     req.CustomerInfo.Lng,
		},
		BookingInfo: models.BookingInfo{
			Day:            req.BookingInfo.GetDayOfWeek(),
			Date:           req.BookingInfo.Date,
			Time:           req.BookingInfo.Time,
			PreferredTime:  preferredTime,
			QuantityOfPros: req.BookingInfo.QuantityOfPros,
		},
		CategorySlug:       req.CategorySlug,
		SelectedServices:   selectedServices,
		SelectedAddons:     selectedAddons,
		SpecialNotes:       req.SpecialNotes,
		ServicesTotal:      servicesTotal,
		AddonsTotal:        addonsTotal,
		Subtotal:           subtotal,
		PlatformCommission: platformCommission,
		TotalPrice:         totalPrice,
		PaymentInfo: &models.PaymentInfo{
			Method: req.PaymentMethod,
			Status: shared.PaymentStatusPending,
			Total:  totalPrice,
		},
		Status:    shared.OrderStatusPending,
		ExpiresAt: shared.TimePtr(shared.CalculateOrderExpiration()),
	}

	if err := s.repo.Create(ctx, order); err != nil {
		logger.Error("failed to create order", "error", err, "customerID", customerID)
		return nil, response.InternalServerError("Failed to create order", err)
	}

	if req.PaymentMethod == "wallet" {
		holdID, err := s.walletService.HoldFunds(
			ctx,
			customerID,
			totalPrice,
			"service_order",
			order.ID,
			"Hold for order booking",
		)
		if err != nil {
			logger.Error("failed to hold wallet funds", "error", err, "customerID", customerID, "amount", totalPrice)

			s.repo.Delete(ctx, order.ID)
			return nil, response.InternalServerError("Failed to process payment", err)
		}

		actualHoldID := holdID
		if len(holdID) > 5 && holdID[:5] == "hold_" {
			actualHoldID = holdID[5:]
		}

		order.WalletHoldID = &actualHoldID
		if err := s.repo.Update(ctx, order); err != nil {

			s.walletService.ReleaseHold(ctx, holdID)
			s.repo.Delete(ctx, order.ID)

			logger.Error("failed to update order with hold ID", "error", err, "orderID", order.ID)
			return nil, response.InternalServerError("Failed to create order", err)
		}
	}

	history := models.NewOrderStatusHistory(
		order.ID,
		"",
		shared.OrderStatusPending,
		&customerID,
		shared.RoleCustomer,
		"Order created",
		nil,
	)
	if err := s.repo.CreateStatusHistory(ctx, history); err != nil {
		logger.Error("failed to create status history", "error", err, "orderID", order.ID)
	}

	order.Status = shared.OrderStatusSearchingProvider
	if err := s.repo.Update(ctx, order); err != nil {
		logger.Error("failed to update order status", "error", err, "orderID", order.ID)
	}

	history = models.NewOrderStatusHistory(
		order.ID,
		shared.OrderStatusPending,
		shared.OrderStatusSearchingProvider,
		nil,
		shared.RoleSystem,
		"Searching for available providers",
		nil,
	)
	s.repo.CreateStatusHistory(ctx, history)

	logger.Info("order created", "orderID", order.ID, "orderNumber", order.OrderNumber, "customerID", customerID)

	return dto.ToOrderCreatedResponse(order), nil
}

func (s *service) validateAndCalculateServices(ctx context.Context, categorySlug string, services []dto.SelectedServiceRequest) (float64, models.SelectedServices, error) {
	var total float64
	var selectedServices models.SelectedServices

	for _, svc := range services {

		service, err := s.serviceRepo.GetActiveServiceBySlug(ctx, svc.ServiceSlug)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return 0, nil, response.BadRequest(fmt.Sprintf("Service '%s' not found or unavailable", svc.ServiceSlug))
			}
			return 0, nil, response.InternalServerError("Failed to validate services", err)
		}

		if service.CategorySlug != categorySlug {
			return 0, nil, response.BadRequest(fmt.Sprintf("Service '%s' does not belong to category '%s'", svc.ServiceSlug, categorySlug))
		}

		if service.BasePrice == nil {
			return 0, nil, response.BadRequest(fmt.Sprintf("Service '%s' does not have a price set", svc.ServiceSlug))
		}

		subtotal := *service.BasePrice * float64(svc.Quantity)
		total += subtotal

		selectedServices = append(selectedServices, models.SelectedServiceItem{
			ServiceSlug: service.ServiceSlug,
			Title:       service.Title,
			Price:       *service.BasePrice,
			Quantity:    svc.Quantity,
		})
	}

	return shared.RoundToTwoDecimals(total), selectedServices, nil
}

func (s *service) validateAndCalculateAddons(ctx context.Context, categorySlug string, addons []dto.SelectedAddonRequest) (float64, models.SelectedAddons, error) {
	if len(addons) == 0 {
		return 0, nil, nil
	}

	var total float64
	var selectedAddons models.SelectedAddons

	for _, add := range addons {
		addon, err := s.serviceRepo.GetActiveAddonBySlug(ctx, add.AddonSlug)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return 0, nil, response.BadRequest(fmt.Sprintf("Addon '%s' not found or unavailable", add.AddonSlug))
			}
			return 0, nil, response.InternalServerError("Failed to validate addons", err)
		}

		if addon.CategorySlug != categorySlug {
			return 0, nil, response.BadRequest(fmt.Sprintf("Addon '%s' does not belong to category '%s'", add.AddonSlug, categorySlug))
		}

		subtotal := addon.Price * float64(add.Quantity)
		total += subtotal

		selectedAddons = append(selectedAddons, models.SelectedAddonItem{
			AddonSlug: addon.AddonSlug,
			Title:     addon.Title,
			Price:     addon.Price,
			Quantity:  add.Quantity,
		})
	}

	return shared.RoundToTwoDecimals(total), selectedAddons, nil
}

func (s *service) GetOrder(ctx context.Context, customerID, orderID string) (*dto.OrderResponse, error) {
	order, err := s.repo.GetCustomerOrderByID(ctx, customerID, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		logger.Error("failed to get order", "error", err, "orderID", orderID, "customerID", customerID)
		return nil, response.InternalServerError("Failed to get order", err)
	}

	return dto.ToOrderResponse(order), nil
}

func (s *service) ListOrders(ctx context.Context, customerID string, query dto.ListOrdersQuery) ([]dto.OrderListResponse, *response.PaginationMeta, error) {
	if err := query.Validate(); err != nil {
		return nil, nil, response.BadRequest(err.Error())
	}

	query.SetDefaults()

	orders, total, err := s.repo.GetCustomerOrders(ctx, customerID, query)
	if err != nil {
		logger.Error("failed to list orders", "error", err, "customerID", customerID)
		return nil, nil, response.InternalServerError("Failed to list orders", err)
	}

	responses := dto.ToOrderListResponses(orders)

	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

func (s *service) GetCancellationPreview(ctx context.Context, customerID, orderID string) (*dto.CancellationPreviewResponse, error) {
	order, err := s.repo.GetCustomerOrderByID(ctx, customerID, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to get order", err)
	}

	if !order.CanBeCancelled() {
		return nil, response.BadRequest(fmt.Sprintf("Order cannot be cancelled in '%s' status", order.Status))
	}

	cancellationFee, refundAmount := shared.CalculateCancellationFee(order.Status, order.TotalPrice)
	var feePercentage float64
	switch order.Status {
	case shared.OrderStatusPending, shared.OrderStatusSearchingProvider:
		feePercentage = shared.CancellationFeeBeforeAcceptance * 100
	case shared.OrderStatusAssigned, shared.OrderStatusAccepted:
		feePercentage = shared.CancellationFeeAfterAcceptance * 100
	}

	message := fmt.Sprintf("Cancellation fee of %.0f%% will be applied.", feePercentage)
	if refundAmount > 0 {
		message += fmt.Sprintf(" You will receive a refund of $%.2f.", refundAmount)
	}

	return &dto.CancellationPreviewResponse{
		OrderID:         order.ID,
		OrderNumber:     order.OrderNumber,
		CurrentStatus:   order.Status,
		TotalPrice:      order.TotalPrice,
		CancellationFee: cancellationFee,
		RefundAmount:    refundAmount,
		FeePercentage:   feePercentage,
		Message:         message,
	}, nil
}

func (s *service) CancelOrder(ctx context.Context, customerID, orderID string, req dto.CancelOrderRequest) (*dto.OrderResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	order, err := s.repo.GetCustomerOrderByID(ctx, customerID, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to get order", err)
	}

	if !order.CanBeCancelled() {
		return nil, response.BadRequest(fmt.Sprintf("Order cannot be cancelled in '%s' status", order.Status))
	}

	cancellationFee, refundAmount := shared.CalculateCancellationFee(order.Status, order.TotalPrice)

	if order.WalletHoldID != nil {
		if refundAmount > 0 {
			if err := s.walletService.ReleaseHold(ctx, *order.WalletHoldID); err != nil {
				logger.Error("failed to release wallet hold", "error", err, "holdID", *order.WalletHoldID)
				return nil, response.InternalServerError("Failed to process refund", err)
			}

			if cancellationFee > 0 {
				if err := s.walletService.Debit(
					ctx,
					customerID,
					cancellationFee,
					"cancellation_fee",
					order.ID,
					fmt.Sprintf("Cancellation fee for order %s", order.OrderNumber),
				); err != nil {
					logger.Error("failed to debit cancellation fee", "error", err, "orderID", order.ID)
				}
			}
		}
	}

	previousStatus := order.Status
	order.Status = shared.OrderStatusCancelled
	order.CancellationInfo = &models.CancellationInfo{
		CancelledBy:     shared.CancelledByCustomer,
		CancelledAt:     time.Now(),
		Reason:          req.Reason,
		CancellationFee: cancellationFee,
		RefundAmount:    refundAmount,
	}

	if order.PaymentInfo != nil {
		order.PaymentInfo.Status = shared.PaymentStatusRefunded
	}

	if err := s.repo.Update(ctx, order); err != nil {
		logger.Error("failed to update order", "error", err, "orderID", order.ID)
		return nil, response.InternalServerError("Failed to cancel order", err)
	}

	history := models.NewOrderStatusHistory(
		order.ID,
		previousStatus,
		shared.OrderStatusCancelled,
		&customerID,
		shared.RoleCustomer,
		fmt.Sprintf("Cancelled by customer: %s", req.Reason),
		models.StatusHistoryMetadata{
			"cancellationFee": cancellationFee,
			"refundAmount":    refundAmount,
		},
	)
	s.repo.CreateStatusHistory(ctx, history)

	logger.Info("order cancelled", "orderID", order.ID, "customerID", customerID,
		"cancellationFee", cancellationFee, "refundAmount", refundAmount)

	return dto.ToOrderResponse(order), nil
}

func (s *service) RateOrder(ctx context.Context, customerID, orderID string, req dto.RateOrderRequest) (*dto.OrderResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	order, err := s.repo.GetCustomerOrderByID(ctx, customerID, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to get order", err)
	}

	if !order.CanBeRatedByCustomer() {
		if order.Status != shared.OrderStatusCompleted {
			return nil, response.BadRequest("Only completed orders can be rated")
		}
		return nil, response.BadRequest("You have already rated this order")
	}

	now := time.Now()
	order.CustomerRating = &req.Rating
	order.CustomerReview = req.Review
	order.CustomerRatedAt = &now

	if err := s.repo.Update(ctx, order); err != nil {
		logger.Error("failed to update order rating", "error", err, "orderID", order.ID)
		return nil, response.InternalServerError("Failed to submit rating", err)
	}

	logger.Info("order rated by customer", "orderID", order.ID, "customerID", customerID, "rating", req.Rating)

	return dto.ToOrderResponse(order), nil
}
