package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/provider/dto"
	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type WalletService interface {
	Credit(ctx context.Context, userID string, amount float64, transactionType, referenceID, description string) error
	CaptureHold(ctx context.Context, holdID string, amount float64, description string) error
}

type Service interface {

	GetProviderIDByUserID(ctx context.Context, userID string) (string, error)
	CreateProviderOnFirstCategory(ctx context.Context, userID string) string

	GetProfile(ctx context.Context, providerID string) (*dto.ProviderProfileResponse, error)
	UpdateAvailability(ctx context.Context, providerID string, req dto.UpdateAvailabilityRequest) error

	GetServiceCategories(ctx context.Context, providerID string) ([]dto.ServiceCategoryResponse, error)
	AddServiceCategory(ctx context.Context, providerID string, req dto.AddServiceCategoryRequest) (*dto.ServiceCategoryResponse, error)
	UpdateServiceCategory(ctx context.Context, providerID, categorySlug string, req dto.UpdateServiceCategoryRequest) (*dto.ServiceCategoryResponse, error)
	DeleteServiceCategory(ctx context.Context, providerID, categorySlug string) error

	GetAvailableOrders(ctx context.Context, providerID string, query dto.ListAvailableOrdersQuery) ([]dto.AvailableOrderResponse, *response.PaginationMeta, error)
	GetAvailableOrderDetail(ctx context.Context, providerID, orderID string) (*dto.AvailableOrderResponse, error)

	GetMyOrders(ctx context.Context, providerID string, query dto.ListMyOrdersQuery) ([]dto.ProviderOrderListResponse, *response.PaginationMeta, error)
	GetMyOrderDetail(ctx context.Context, providerID, orderID string) (*dto.ProviderOrderResponse, error)
	AcceptOrder(ctx context.Context, providerID, orderID string) (*dto.ProviderOrderResponse, error)
	RejectOrder(ctx context.Context, providerID, orderID string, req dto.RejectOrderRequest) error
	StartOrder(ctx context.Context, providerID, orderID string) (*dto.ProviderOrderResponse, error)
	CompleteOrder(ctx context.Context, providerID, orderID string, req dto.CompleteOrderRequest) (*dto.ProviderOrderResponse, error)
	RateCustomer(ctx context.Context, providerID, orderID string, req dto.RateCustomerRequest) (*dto.ProviderOrderResponse, error)

	GetStatistics(ctx context.Context, providerID string) (*dto.ProviderStatistics, error)
	GetEarnings(ctx context.Context, providerID string, query dto.EarningsQuery) (*dto.EarningsSummaryResponse, error)
}

type service struct {
	repo          Repository
	walletService WalletService
}

func NewService(repo Repository, walletService WalletService) Service {
	return &service{
		repo:          repo,
		walletService: walletService,
	}
}

func (s *service) GetProviderIDByUserID(ctx context.Context, userID string) (string, error) {
	provider, err := s.repo.GetProviderByUserID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", response.NotFoundError("Provider not found for this user")
		}
		logger.Error("failed to get provider by user ID", "error", err, "userID", userID)
		return "", response.InternalServerError("Failed to retrieve provider", err)
	}
	return provider.ID, nil
}

func (s *service) CreateProviderOnFirstCategory(ctx context.Context, userID string) string {
	providerID := uuid.New().String()

	provider := &models.ServiceProviderProfile{
		ID:              providerID,
		UserID:          userID,
		IsVerified:      false,
		IsAvailable:     false,
		ServiceType:     "service_provider",
		ServiceCategory: "general",
		Status:          models.SPStatusPendingApproval,
	}

	if err := s.repo.CreateProvider(ctx, provider); err != nil {
		logger.Error("failed to create provider profile on first category registration",
			"error", err,
			"userID", userID,
			"providerID", providerID,
		)
	}

	logger.Info("provider profile created on first category registration",
		"providerID", providerID,
		"userID", userID,
	)

	return providerID
}

func (s *service) GetProfile(ctx context.Context, providerID string) (*dto.ProviderProfileResponse, error) {
	provider, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Provider not found")
		}
		logger.Error("failed to get provider", "error", err, "providerID", providerID)
		return nil, response.InternalServerError("Failed to get profile", err)
	}

	categories, err := s.repo.GetProviderCategories(ctx, providerID)
	if err != nil {
		logger.Error("failed to get provider categories", "error", err, "providerID", providerID)
		return nil, response.InternalServerError("Failed to get profile", err)
	}

	stats, err := s.GetStatistics(ctx, providerID)
	if err != nil {
		logger.Error("failed to get provider statistics", "error", err, "providerID", providerID)
		stats = &dto.ProviderStatistics{}
	}

	email := ""
	phone := ""
	if provider.User.Email != nil {
		email = *provider.User.Email
	}
	if provider.User.Phone != nil {
		phone = *provider.User.Phone
	}

	profile := &dto.ProviderProfileResponse{
		ID:                provider.ID,
		UserID:            provider.UserID,
		Name:              provider.User.Name,
		Email:             email,
		Phone:             phone,
		IsVerified:        provider.IsVerified,
		IsAvailable:       provider.IsAvailable,
		ServiceCategories: dto.ToServiceCategoryResponses(categories),
		Statistics:        *stats,
		CreatedAt:         provider.CreatedAt,
	}

	return profile, nil
}

func (s *service) UpdateAvailability(ctx context.Context, providerID string, req dto.UpdateAvailabilityRequest) error {
	logger.Info("provider availability updated", "providerID", providerID, "isAvailable", req.IsAvailable)
	return nil
}

func (s *service) GetServiceCategories(ctx context.Context, providerID string) ([]dto.ServiceCategoryResponse, error) {
	categories, err := s.repo.GetProviderCategories(ctx, providerID)
	if err != nil {
		logger.Error("failed to get provider categories", "error", err, "providerID", providerID)
		return nil, response.InternalServerError("Failed to get service categories", err)
	}

	return dto.ToServiceCategoryResponses(categories), nil
}

func (s *service) AddServiceCategory(ctx context.Context, providerID string, req dto.AddServiceCategoryRequest) (*dto.ServiceCategoryResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	existing, err := s.repo.GetProviderCategory(ctx, providerID, req.CategorySlug)
	if err == nil && existing != nil {
		return nil, response.ConflictError(fmt.Sprintf("You already have category '%s' registered", req.CategorySlug))
	}

	category := &models.ProviderServiceCategory{
		ProviderID:        providerID,
		CategorySlug:      req.CategorySlug,
		ExpertiseLevel:    req.ExpertiseLevel,
		YearsOfExperience: req.YearsOfExperience,
		IsActive:          true,
	}

	if err := s.repo.AddProviderCategory(ctx, category); err != nil {
		logger.Error("failed to add provider category", "error", err, "providerID", providerID)
		return nil, response.InternalServerError("Failed to add service category", err)
	}

	logger.Info("provider category added", "providerID", providerID, "category", req.CategorySlug)

	result := dto.ToServiceCategoryResponse(category)
	return &result, nil
}

func (s *service) UpdateServiceCategory(ctx context.Context, providerID, categorySlug string, req dto.UpdateServiceCategoryRequest) (*dto.ServiceCategoryResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	category, err := s.repo.GetProviderCategory(ctx, providerID, categorySlug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Service category")
		}
		return nil, response.InternalServerError("Failed to get service category", err)
	}

	if req.ExpertiseLevel != nil {
		category.ExpertiseLevel = *req.ExpertiseLevel
	}
	if req.YearsOfExperience != nil {
		category.YearsOfExperience = *req.YearsOfExperience
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateProviderCategory(ctx, category); err != nil {
		logger.Error("failed to update provider category", "error", err, "providerID", providerID)
		return nil, response.InternalServerError("Failed to update service category", err)
	}

	logger.Info("provider category updated", "providerID", providerID, "category", categorySlug)

	result := dto.ToServiceCategoryResponse(category)
	return &result, nil
}

func (s *service) DeleteServiceCategory(ctx context.Context, providerID, categorySlug string) error {
	_, err := s.repo.GetProviderCategory(ctx, providerID, categorySlug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.NotFoundError("Service category")
		}
		return response.InternalServerError("Failed to get service category", err)
	}

	if err := s.repo.DeleteProviderCategory(ctx, providerID, categorySlug); err != nil {
		logger.Error("failed to delete provider category", "error", err, "providerID", providerID)
		return response.InternalServerError("Failed to delete service category", err)
	}

	logger.Info("provider category deleted", "providerID", providerID, "category", categorySlug)
	return nil
}

func (s *service) GetAvailableOrders(ctx context.Context, providerID string, query dto.ListAvailableOrdersQuery) ([]dto.AvailableOrderResponse, *response.PaginationMeta, error) {
	categorySlugs, err := s.repo.GetProviderCategorySlugs(ctx, providerID)
	if err != nil {
		logger.Error("failed to get provider categories", "error", err, "providerID", providerID)
		return nil, nil, response.InternalServerError("Failed to get available orders", err)
	}

	logger.Info("fetched provider category slugs", "providerID", providerID, "categories", categorySlugs)

	if provider, perr := s.repo.GetProvider(ctx, providerID); perr == nil && provider != nil {
		logger.Info("fetched provider profile", "providerID", providerID, "serviceType", provider.ServiceType, "serviceCategory", provider.ServiceCategory)

		addIfMissing := func(slice []string, v string) []string {
			if v == "" {
				return slice
			}
			for _, s := range slice {
				if s == v {
					return slice
				}
			}
			return append(slice, v)
		}

		categorySlugs = addIfMissing(categorySlugs, provider.ServiceType)
		categorySlugs = addIfMissing(categorySlugs, provider.ServiceCategory)
		logger.Info("merged profile categories with registered categories", "providerID", providerID, "mergedCategories", categorySlugs)
	} else if perr != nil && perr != gorm.ErrRecordNotFound {
		logger.Error("failed to get provider profile", "error", perr, "providerID", providerID)
		return nil, nil, response.InternalServerError("Failed to get available orders", perr)
	}

	if len(categorySlugs) == 0 {
		logger.Warn("provider has no active categories", "providerID", providerID)

		return []dto.AvailableOrderResponse{}, &response.PaginationMeta{
			Total:      0,
			Page:       1,
			Limit:      query.Limit,
			TotalPages: 0,
		}, nil
	}

	query.SetDefaults()

	orders, total, err := s.repo.GetAvailableOrders(ctx, categorySlugs, query)
	if err != nil {
		logger.Error("failed to get available orders", "error", err, "providerID", providerID)
		return nil, nil, response.InternalServerError("Failed to get available orders", err)
	}

	responses := make([]dto.AvailableOrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = dto.ToAvailableOrderResponse(order, nil)
	}

	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)
	return responses, &pagination, nil
}

func (s *service) GetAvailableOrderDetail(ctx context.Context, providerID, orderID string) (*dto.AvailableOrderResponse, error) {
	categorySlugs, err := s.repo.GetProviderCategorySlugs(ctx, providerID)
	if err != nil {
		return nil, response.InternalServerError("Failed to get order", err)
	}

	if len(categorySlugs) == 0 {
		return nil, response.ForbiddenError("You have no active service categories")
	}

	order, err := s.repo.GetAvailableOrderByID(ctx, orderID, categorySlugs)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to get order", err)
	}

	result := dto.ToAvailableOrderResponse(order, nil)
	return &result, nil
}

func (s *service) GetMyOrders(ctx context.Context, providerID string, query dto.ListMyOrdersQuery) ([]dto.ProviderOrderListResponse, *response.PaginationMeta, error) {
	if err := query.Validate(); err != nil {
		return nil, nil, response.BadRequest(err.Error())
	}

	query.SetDefaults()

	orders, total, err := s.repo.GetProviderOrders(ctx, providerID, query)
	if err != nil {
		logger.Error("failed to get provider orders", "error", err, "providerID", providerID)
		return nil, nil, response.InternalServerError("Failed to get orders", err)
	}

	responses := dto.ToProviderOrderListResponses(orders)
	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

func (s *service) GetMyOrderDetail(ctx context.Context, providerID, orderID string) (*dto.ProviderOrderResponse, error) {
	order, err := s.repo.GetProviderOrderByID(ctx, providerID, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to get order", err)
	}

	return dto.ToProviderOrderResponse(order), nil
}

func (s *service) AcceptOrder(ctx context.Context, providerID, orderID string) (*dto.ProviderOrderResponse, error) {
	activeCount, err := s.repo.CountProviderActiveOrders(ctx, providerID)
	if err != nil {
		return nil, response.InternalServerError("Failed to accept order", err)
	}
	if activeCount >= 5 {
		return nil, response.BadRequest("You have too many active orders. Complete some orders before accepting new ones.")
	}

	categorySlugs, err := s.repo.GetProviderCategorySlugs(ctx, providerID)
	if err != nil {
		return nil, response.InternalServerError("Failed to accept order", err)
	}
	order, err := s.repo.GetAvailableOrderByID(ctx, orderID, categorySlugs)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order not available")
		}
		return nil, response.InternalServerError("Failed to accept order", err)
	}
	now := time.Now()
	previousStatus := order.Status
	order.AssignedProviderID = &providerID
	order.Status = shared.OrderStatusAccepted
	order.ProviderAcceptedAt = &now

	if err := s.repo.UpdateOrder(ctx, order); err != nil {
		logger.Error("failed to update order", "error", err, "orderID", orderID)
		return nil, response.InternalServerError("Failed to accept order", err)
	}

	if order.WalletHoldID != nil {
		if err := s.walletService.CaptureHold(
			ctx,
			*order.WalletHoldID,
			order.TotalPrice,
			fmt.Sprintf("Payment for order %s", order.OrderNumber),
		); err != nil {
			logger.Error("failed to capture wallet hold", "error", err, "orderID", orderID)
		}
	}

	history := models.NewOrderStatusHistory(
		order.ID,
		previousStatus,
		shared.OrderStatusAccepted,
		&providerID,
		shared.RoleProvider,
		"Order accepted by provider",
		nil,
	)
	s.repo.CreateStatusHistory(ctx, history)

	logger.Info("order accepted", "orderID", orderID, "providerID", providerID)

	return dto.ToProviderOrderResponse(order), nil
}

func (s *service) RejectOrder(ctx context.Context, providerID, orderID string, req dto.RejectOrderRequest) error {

	if err := req.Validate(); err != nil {
		return response.BadRequest(err.Error())
	}

	order, err := s.repo.GetProviderOrderByID(ctx, providerID, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.NotFoundError("Order")
		}
		return response.InternalServerError("Failed to reject order", err)
	}

	if order.Status != shared.OrderStatusAssigned {
		return response.BadRequest(fmt.Sprintf("Cannot reject order in '%s' status", order.Status))
	}

	previousStatus := order.Status
	order.AssignedProviderID = nil
	order.Status = shared.OrderStatusSearchingProvider

	if err := s.repo.UpdateOrder(ctx, order); err != nil {
		return response.InternalServerError("Failed to reject order", err)
	}

	history := models.NewOrderStatusHistory(
		order.ID,
		previousStatus,
		shared.OrderStatusSearchingProvider,
		&providerID,
		shared.RoleProvider,
		fmt.Sprintf("Rejected by provider: %s", req.Reason),
		nil,
	)
	s.repo.CreateStatusHistory(ctx, history)

	logger.Info("order rejected", "orderID", orderID, "providerID", providerID, "reason", req.Reason)

	return nil
}

func (s *service) StartOrder(ctx context.Context, providerID, orderID string) (*dto.ProviderOrderResponse, error) {
	order, err := s.repo.GetProviderOrderByID(ctx, providerID, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to start order", err)
	}

	if order.Status != shared.OrderStatusAccepted {
		return nil, response.BadRequest(fmt.Sprintf("Cannot start order in '%s' status", order.Status))
	}

	now := time.Now()
	previousStatus := order.Status
	order.Status = shared.OrderStatusInProgress
	order.ProviderStartedAt = &now

	if err := s.repo.UpdateOrder(ctx, order); err != nil {
		return nil, response.InternalServerError("Failed to start order", err)
	}

	history := models.NewOrderStatusHistory(
		order.ID,
		previousStatus,
		shared.OrderStatusInProgress,
		&providerID,
		shared.RoleProvider,
		"Service started",
		nil,
	)
	s.repo.CreateStatusHistory(ctx, history)

	logger.Info("order started", "orderID", orderID, "providerID", providerID)

	return dto.ToProviderOrderResponse(order), nil
}

func (s *service) CompleteOrder(ctx context.Context, providerID, orderID string, req dto.CompleteOrderRequest) (*dto.ProviderOrderResponse, error) {
	order, err := s.repo.GetProviderOrderByID(ctx, providerID, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to complete order", err)
	}

	if order.Status != shared.OrderStatusInProgress {
		return nil, response.BadRequest(fmt.Sprintf("Cannot complete order in '%s' status", order.Status))
	}

	providerPayout := dto.CalculateProviderPayout(order.TotalPrice)

	if err := s.walletService.Credit(
		ctx,
		providerID,
		providerPayout,
		"service_payment",
		order.ID,
		fmt.Sprintf("Payment for order %s", order.OrderNumber),
	); err != nil {
		logger.Error("failed to credit provider wallet", "error", err, "orderID", orderID)
		return nil, response.InternalServerError("Failed to process payment", err)
	}
	now := time.Now()
	previousStatus := order.Status
	order.Status = shared.OrderStatusCompleted
	order.ProviderCompletedAt = &now
	order.CompletedAt = &now

	if order.PaymentInfo != nil {
		order.PaymentInfo.Status = shared.PaymentStatusCompleted
		order.PaymentInfo.AmountPaid = order.TotalPrice
	}

	if err := s.repo.UpdateOrder(ctx, order); err != nil {
		return nil, response.InternalServerError("Failed to complete order", err)
	}

	category, err := s.repo.GetProviderCategory(ctx, providerID, order.CategorySlug)
	if err == nil && category != nil {
		category.IncrementCompletedJobs(providerPayout)
		s.repo.UpdateProviderCategory(ctx, category)
	}

	metadata := models.StatusHistoryMetadata{
		"providerPayout": providerPayout,
	}
	if req.Notes != "" {
		metadata["completionNotes"] = req.Notes
	}
	history := models.NewOrderStatusHistory(
		order.ID,
		previousStatus,
		shared.OrderStatusCompleted,
		&providerID,
		shared.RoleProvider,
		"Service completed",
		metadata,
	)
	s.repo.CreateStatusHistory(ctx, history)

	logger.Info("order completed", "orderID", orderID, "providerID", providerID, "payout", providerPayout)

	return dto.ToProviderOrderResponse(order), nil
}

func (s *service) RateCustomer(ctx context.Context, providerID, orderID string, req dto.RateCustomerRequest) (*dto.ProviderOrderResponse, error) {

	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	order, err := s.repo.GetProviderOrderByID(ctx, providerID, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to rate customer", err)
	}

	if order.Status != shared.OrderStatusCompleted {
		return nil, response.BadRequest("Only completed orders can be rated")
	}

	if order.ProviderRating != nil {
		return nil, response.BadRequest("You have already rated this customer")
	}

	now := time.Now()
	order.ProviderRating = &req.Rating
	order.ProviderReview = req.Review
	order.ProviderRatedAt = &now

	if err := s.repo.UpdateOrder(ctx, order); err != nil {
		return nil, response.InternalServerError("Failed to submit rating", err)
	}

	logger.Info("customer rated", "orderID", orderID, "providerID", providerID, "rating", req.Rating)

	return dto.ToProviderOrderResponse(order), nil
}

func (s *service) GetStatistics(ctx context.Context, providerID string) (*dto.ProviderStatistics, error) {
	stats, err := s.repo.GetProviderStatistics(ctx, providerID)
	if err != nil {
		return nil, response.InternalServerError("Failed to get statistics", err)
	}

	var overallRating float64
	if stats.TotalRatings > 0 {
		overallRating = float64(stats.TotalRatingSum) / float64(stats.TotalRatings)
	}

	var acceptanceRate float64
	totalOffers := stats.TotalAccepted + stats.TotalRejected
	if totalOffers > 0 {
		acceptanceRate = float64(stats.TotalAccepted) / float64(totalOffers) * 100
	}

	return &dto.ProviderStatistics{
		TotalCompletedJobs:   stats.TotalCompletedJobs,
		TotalEarnings:        stats.TotalEarnings,
		OverallRating:        overallRating,
		TotalRatings:         stats.TotalRatings,
		AcceptanceRate:       acceptanceRate,
		AverageResponseTime:  stats.AvgResponseMinutes,
		TotalActiveOrders:    stats.ActiveOrders,
		TodayCompletedOrders: stats.TodayCompletedOrders,
		TodayEarnings:        stats.TodayEarnings,
	}, nil
}

func (s *service) GetEarnings(ctx context.Context, providerID string, query dto.EarningsQuery) (*dto.EarningsSummaryResponse, error) {
	query.SetDefaults()

	fromDate, err := time.Parse("2006-01-02", query.FromDate)
	if err != nil {
		return nil, response.BadRequest("Invalid fromDate format")
	}
	toDate, err := time.Parse("2006-01-02", query.ToDate)
	if err != nil {
		return nil, response.BadRequest("Invalid toDate format")
	}

	earningsData, err := s.repo.GetProviderEarnings(ctx, providerID, fromDate, toDate)
	if err != nil {
		return nil, response.InternalServerError("Failed to get earnings", err)
	}

	categoryData, err := s.repo.GetCategoryEarnings(ctx, providerID, fromDate, toDate)
	if err != nil {
		return nil, response.InternalServerError("Failed to get earnings", err)
	}

	var breakdown []dto.EarningsBreakdown
	for _, d := range earningsData.DailyBreakdown {
		breakdown = append(breakdown, dto.EarningsBreakdown{
			Period:            d.Date,
			Earnings:          d.Earnings,
			OrderCount:        d.OrderCount,
			FormattedEarnings: dto.FormatPrice(d.Earnings),
		})
	}

	var categoryEarnings []dto.CategoryEarnings
	for _, c := range categoryData {
		percentage := 0.0
		if earningsData.TotalEarnings > 0 {
			percentage = (c.Earnings / earningsData.TotalEarnings) * 100
		}
		categoryEarnings = append(categoryEarnings, dto.CategoryEarnings{
			CategorySlug:  c.CategorySlug,
			CategoryTitle: dto.GetCategoryTitle(c.CategorySlug),
			Earnings:      c.Earnings,
			OrderCount:    c.OrderCount,
			Percentage:    percentage,
		})
	}

	var averagePerOrder float64
	if earningsData.TotalOrders > 0 {
		averagePerOrder = earningsData.TotalEarnings / float64(earningsData.TotalOrders)
	}

	return &dto.EarningsSummaryResponse{
		TotalEarnings:   earningsData.TotalEarnings,
		TotalOrders:     earningsData.TotalOrders,
		AveragePerOrder: averagePerOrder,
		FormattedTotal:  dto.FormatPrice(earningsData.TotalEarnings),
		Period: dto.EarningsPeriod{
			FromDate: query.FromDate,
			ToDate:   query.ToDate,
		},
		Breakdown:  breakdown,
		ByCategory: categoryEarnings,
	}, nil
}
