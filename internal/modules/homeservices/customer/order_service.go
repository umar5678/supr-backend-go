package customer

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/customer/dto"
	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

// OrderService defines the interface for order business logic
type OrderService interface {
	// Order operations
	CreateOrder(ctx context.Context, customerID string, req dto.CreateOrderRequest) (*dto.OrderCreatedResponse, error)
	GetOrder(ctx context.Context, customerID, orderID string) (*dto.OrderResponse, error)
	ListOrders(ctx context.Context, customerID string, query dto.ListOrdersQuery) ([]dto.OrderListResponse, *response.PaginationMeta, error)

	// Cancellation
	GetCancellationPreview(ctx context.Context, customerID, orderID string) (*dto.CancellationPreviewResponse, error)
	CancelOrder(ctx context.Context, customerID, orderID string, req dto.CancelOrderRequest) (*dto.OrderResponse, error)

	// Rating
	RateOrder(ctx context.Context, customerID, orderID string, req dto.RateOrderRequest) (*dto.OrderResponse, error)
}

type orderService struct {
	orderRepo     OrderRepository
	serviceRepo   Repository // From Module 3 - for validating services/addons
	walletService WalletService
}

// NewOrderService creates a new order service instance
func NewOrderService(orderRepo OrderRepository, serviceRepo Repository, walletService WalletService) OrderService {
	return &orderService{
		orderRepo:     orderRepo,
		serviceRepo:   serviceRepo,
		walletService: walletService,
	}
}

// ==================== Create Order ====================

func (s *orderService) CreateOrder(ctx context.Context, customerID string, req dto.CreateOrderRequest) (*dto.OrderCreatedResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Check for too many active orders
	activeCount, err := s.orderRepo.CountCustomerActiveOrders(ctx, customerID)
	if err != nil {
		logger.Error("failed to count active orders", "error", err, "customerID", customerID)
		return nil, response.InternalServerError("Failed to create order", err)
	}
	if activeCount >= 5 {
		return nil, response.BadRequest("You have too many active orders. Please wait for some to complete before booking again.")
	}

	// Validate and calculate services
	servicesTotal, selectedServices, err := s.validateAndCalculateServices(ctx, req.CategorySlug, req.SelectedServices)
	if err != nil {
		return nil, err
	}

	// Validate and calculate addons
	addonsTotal, selectedAddons, err := s.validateAndCalculateAddons(ctx, req.CategorySlug, req.SelectedAddons)
	if err != nil {
		return nil, err
	}

	// Calculate pricing
	subtotal := servicesTotal + addonsTotal
	platformCommission := shared.CalculatePlatformCommission(subtotal)
	totalPrice := subtotal // Customer pays subtotal, commission is taken from provider payment

	// Validate wallet balance if paying with wallet
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

	// Create order
	// Parse PreferredTime (accept RFC3339 first, then fallback to "15:04" hour:minute format)
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

	// Save order FIRST to generate the ID
	if err := s.orderRepo.Create(ctx, order); err != nil {
		logger.Error("failed to create order", "error", err, "customerID", customerID)
		return nil, response.InternalServerError("Failed to create order", err)
	}

	// Hold funds if paying with wallet
	if req.PaymentMethod == "wallet" {
		holdID, err := s.walletService.HoldFunds(
			ctx,
			customerID,
			totalPrice,
			"service_order",
			order.ID,
			fmt.Sprintf("Hold for order booking"),
		)
		if err != nil {
			logger.Error("failed to hold wallet funds", "error", err, "customerID", customerID, "amount", totalPrice)

			s.orderRepo.Delete(ctx, order.ID)
			return nil, response.InternalServerError("Failed to process payment", err)
		}

		// Strip "hold_" prefix if present to get the actual UUID
		actualHoldID := holdID
		if len(holdID) > 5 && holdID[:5] == "hold_" {
			actualHoldID = holdID[5:] // Remove "hold_" prefix
		}

		order.WalletHoldID = &actualHoldID
		if err := s.orderRepo.Update(ctx, order); err != nil {
			// Release hold if update fails
			s.walletService.ReleaseHold(ctx, holdID)
			s.orderRepo.Delete(ctx, order.ID)

			logger.Error("failed to update order with hold ID", "error", err, "orderID", order.ID)
			return nil, response.InternalServerError("Failed to create order", err)
		}
	}

	// // Save order
	// if err := s.orderRepo.Create(ctx, order); err != nil {
	// 	// Release hold if order creation fails
	// 	if order.WalletHoldID != nil {
	// 		s.walletService.ReleaseHold(ctx, *order.WalletHoldID)
	// 	}
	// 	logger.Error("failed to create order", "error", err, "customerID", customerID)
	// 	return nil, response.InternalServerError("Failed to create order", err)
	// }

	// Create status history
	history := models.NewOrderStatusHistory(
		order.ID,
		"",
		shared.OrderStatusPending,
		&customerID,
		shared.RoleCustomer,
		"Order created",
		nil,
	)
	if err := s.orderRepo.CreateStatusHistory(ctx, history); err != nil {
		logger.Error("failed to create status history", "error", err, "orderID", order.ID)
		// Don't fail the order creation for this
	}

	// Update status to searching_provider
	order.Status = shared.OrderStatusSearchingProvider
	if err := s.orderRepo.Update(ctx, order); err != nil {
		logger.Error("failed to update order status", "error", err, "orderID", order.ID)
	}

	// Create status history for searching
	history = models.NewOrderStatusHistory(
		order.ID,
		shared.OrderStatusPending,
		shared.OrderStatusSearchingProvider,
		nil,
		shared.RoleSystem,
		"Searching for available providers",
		nil,
	)
	s.orderRepo.CreateStatusHistory(ctx, history)

	logger.Info("order created", "orderID", order.ID, "orderNumber", order.OrderNumber, "customerID", customerID)

	// TODO: Trigger provider matching in background
	// go s.providerMatchingService.FindProviders(ctx, order)

	return dto.ToOrderCreatedResponse(order), nil
}

func (s *orderService) validateAndCalculateServices(ctx context.Context, categorySlug string, services []dto.SelectedServiceRequest) (float64, models.SelectedServices, error) {
	var total float64
	var selectedServices models.SelectedServices

	for _, svc := range services {
		// Get service from database
		service, err := s.serviceRepo.GetActiveServiceBySlug(ctx, svc.ServiceSlug)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return 0, nil, response.BadRequest(fmt.Sprintf("Service '%s' not found or unavailable", svc.ServiceSlug))
			}
			return 0, nil, response.InternalServerError("Failed to validate services", err)
		}

		// Validate category matches
		if service.CategorySlug != categorySlug {
			return 0, nil, response.BadRequest(fmt.Sprintf("Service '%s' does not belong to category '%s'", svc.ServiceSlug, categorySlug))
		}

		// Check if service has a price
		if service.BasePrice == nil {
			return 0, nil, response.BadRequest(fmt.Sprintf("Service '%s' does not have a price set", svc.ServiceSlug))
		}

		// Calculate subtotal
		subtotal := *service.BasePrice * float64(svc.Quantity)
		total += subtotal

		// Add to selected services
		selectedServices = append(selectedServices, models.SelectedServiceItem{
			ServiceSlug: service.ServiceSlug,
			Title:       service.Title,
			Price:       *service.BasePrice,
			Quantity:    svc.Quantity,
		})
	}

	return shared.RoundToTwoDecimals(total), selectedServices, nil
}

func (s *orderService) validateAndCalculateAddons(ctx context.Context, categorySlug string, addons []dto.SelectedAddonRequest) (float64, models.SelectedAddons, error) {
	if len(addons) == 0 {
		return 0, nil, nil
	}

	var total float64
	var selectedAddons models.SelectedAddons

	for _, add := range addons {
		// Get addon from database
		addon, err := s.serviceRepo.GetActiveAddonBySlug(ctx, add.AddonSlug)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return 0, nil, response.BadRequest(fmt.Sprintf("Addon '%s' not found or unavailable", add.AddonSlug))
			}
			return 0, nil, response.InternalServerError("Failed to validate addons", err)
		}

		// Validate category matches
		if addon.CategorySlug != categorySlug {
			return 0, nil, response.BadRequest(fmt.Sprintf("Addon '%s' does not belong to category '%s'", add.AddonSlug, categorySlug))
		}

		// Calculate subtotal
		subtotal := addon.Price * float64(add.Quantity)
		total += subtotal

		// Add to selected addons
		selectedAddons = append(selectedAddons, models.SelectedAddonItem{
			AddonSlug: addon.AddonSlug,
			Title:     addon.Title,
			Price:     addon.Price,
			Quantity:  add.Quantity,
		})
	}

	return shared.RoundToTwoDecimals(total), selectedAddons, nil
}

// ==================== Get Order ====================

func (s *orderService) GetOrder(ctx context.Context, customerID, orderID string) (*dto.OrderResponse, error) {
	order, err := s.orderRepo.GetCustomerOrderByID(ctx, customerID, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		logger.Error("failed to get order", "error", err, "orderID", orderID, "customerID", customerID)
		return nil, response.InternalServerError("Failed to get order", err)
	}

	return dto.ToOrderResponse(order), nil
}

// ==================== List Orders ====================

func (s *orderService) ListOrders(ctx context.Context, customerID string, query dto.ListOrdersQuery) ([]dto.OrderListResponse, *response.PaginationMeta, error) {
	// Validate query
	if err := query.Validate(); err != nil {
		return nil, nil, response.BadRequest(err.Error())
	}

	// Set defaults
	query.SetDefaults()

	// Get orders
	orders, total, err := s.orderRepo.GetCustomerOrders(ctx, customerID, query)
	if err != nil {
		logger.Error("failed to list orders", "error", err, "customerID", customerID)
		return nil, nil, response.InternalServerError("Failed to list orders", err)
	}

	// Convert to response
	responses := dto.ToOrderListResponses(orders)

	// Create pagination metadata
	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

// ==================== Cancellation ====================

func (s *orderService) GetCancellationPreview(ctx context.Context, customerID, orderID string) (*dto.CancellationPreviewResponse, error) {
	order, err := s.orderRepo.GetCustomerOrderByID(ctx, customerID, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to get order", err)
	}

	// Check if order can be cancelled
	if !order.CanBeCancelled() {
		return nil, response.BadRequest(fmt.Sprintf("Order cannot be cancelled in '%s' status", order.Status))
	}

	// Calculate cancellation fee
	cancellationFee, refundAmount := shared.CalculateCancellationFee(order.Status, order.TotalPrice)

	// Determine fee percentage for display
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

func (s *orderService) CancelOrder(ctx context.Context, customerID, orderID string, req dto.CancelOrderRequest) (*dto.OrderResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Get order
	order, err := s.orderRepo.GetCustomerOrderByID(ctx, customerID, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to get order", err)
	}

	// Check if order can be cancelled
	if !order.CanBeCancelled() {
		return nil, response.BadRequest(fmt.Sprintf("Order cannot be cancelled in '%s' status", order.Status))
	}

	// Calculate cancellation fee
	cancellationFee, refundAmount := shared.CalculateCancellationFee(order.Status, order.TotalPrice)

	// Process refund/release hold
	if order.WalletHoldID != nil {
		if refundAmount > 0 {
			// Release the hold
			if err := s.walletService.ReleaseHold(ctx, *order.WalletHoldID); err != nil {
				logger.Error("failed to release wallet hold", "error", err, "holdID", *order.WalletHoldID)
				return nil, response.InternalServerError("Failed to process refund", err)
			}

			// If there's a cancellation fee, debit it
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
					// Continue with cancellation even if fee debit fails
				}
			}
		}
	}

	// Update order
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

	if err := s.orderRepo.Update(ctx, order); err != nil {
		logger.Error("failed to update order", "error", err, "orderID", order.ID)
		return nil, response.InternalServerError("Failed to cancel order", err)
	}

	// Create status history
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
	s.orderRepo.CreateStatusHistory(ctx, history)

	logger.Info("order cancelled", "orderID", order.ID, "customerID", customerID,
		"cancellationFee", cancellationFee, "refundAmount", refundAmount)

	return dto.ToOrderResponse(order), nil
}

// ==================== Rating ====================

func (s *orderService) RateOrder(ctx context.Context, customerID, orderID string, req dto.RateOrderRequest) (*dto.OrderResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Get order
	order, err := s.orderRepo.GetCustomerOrderByID(ctx, customerID, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to get order", err)
	}

	// Check if order can be rated
	if !order.CanBeRatedByCustomer() {
		if order.Status != shared.OrderStatusCompleted {
			return nil, response.BadRequest("Only completed orders can be rated")
		}
		return nil, response.BadRequest("You have already rated this order")
	}

	// Update order with rating
	now := time.Now()
	order.CustomerRating = &req.Rating
	order.CustomerReview = req.Review
	order.CustomerRatedAt = &now

	if err := s.orderRepo.Update(ctx, order); err != nil {
		logger.Error("failed to update order rating", "error", err, "orderID", order.ID)
		return nil, response.InternalServerError("Failed to submit rating", err)
	}

	// TODO: Update provider's average rating
	// if order.AssignedProviderID != nil {
	//     s.providerService.UpdateRating(ctx, *order.AssignedProviderID, req.Rating)
	// }

	logger.Info("order rated by customer", "orderID", order.ID, "customerID", customerID, "rating", req.Rating)

	return dto.ToOrderResponse(order), nil
}
