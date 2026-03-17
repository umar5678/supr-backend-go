package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/admin/dto"
	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
	"github.com/umar5678/go-backend/internal/modules/wallet"
	walletdto "github.com/umar5678/go-backend/internal/modules/wallet/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
	CreateService(ctx context.Context, req dto.CreateServiceRequest) (*dto.ServiceResponse, error)
	GetServiceBySlug(ctx context.Context, slug string) (*dto.ServiceResponse, error)
	UpdateService(ctx context.Context, slug string, req dto.UpdateServiceRequest) (*dto.ServiceResponse, error)
	UpdateServiceStatus(ctx context.Context, slug string, req dto.UpdateServiceStatusRequest) (*dto.ServiceResponse, error)
	DeleteService(ctx context.Context, slug string) error
	ListServices(ctx context.Context, query dto.ListServicesQuery) ([]*dto.ServiceListResponse, *response.PaginationMeta, error)

	CreateAddon(ctx context.Context, req dto.CreateAddonRequest) (*dto.AddonResponse, error)
	GetAddonBySlug(ctx context.Context, slug string) (*dto.AddonResponse, error)
	UpdateAddon(ctx context.Context, slug string, req dto.UpdateAddonRequest) (*dto.AddonResponse, error)
	UpdateAddonStatus(ctx context.Context, slug string, req dto.UpdateAddonStatusRequest) (*dto.AddonResponse, error)
	DeleteAddon(ctx context.Context, slug string) error
	ListAddons(ctx context.Context, query dto.ListAddonsQuery) ([]*dto.AddonListResponse, *response.PaginationMeta, error)

	GetCategoryDetails(ctx context.Context, categorySlug string) (*dto.CategoryServicesResponse, error)
	GetAllCategories(ctx context.Context) ([]string, error)

	GetOrders(ctx context.Context, query dto.ListOrdersQuery) ([]dto.AdminOrderListResponse, *response.PaginationMeta, error)
	GetOrderByID(ctx context.Context, orderID string) (*dto.AdminOrderDetailResponse, error)
	GetOrderByNumber(ctx context.Context, orderNumber string) (*dto.AdminOrderDetailResponse, error)
	UpdateOrderStatus(ctx context.Context, orderID string, req dto.UpdateOrderStatusRequest, adminID string) (*dto.AdminOrderDetailResponse, error)
	ReassignOrder(ctx context.Context, orderID string, req dto.ReassignOrderRequest, adminID string) (*dto.AdminOrderDetailResponse, error)
	CancelOrder(ctx context.Context, orderID string, req dto.AdminCancelOrderRequest, adminID string) (*dto.AdminOrderDetailResponse, error)

	BulkUpdateStatus(ctx context.Context, req dto.BulkUpdateStatusRequest, adminID string) (int64, error)

	GetOverviewAnalytics(ctx context.Context, query dto.AnalyticsQuery) (*dto.OverviewAnalyticsResponse, error)
	GetProviderAnalytics(ctx context.Context, query dto.ProviderAnalyticsQuery) (*dto.ProviderAnalyticsResponse, error)
	GetRevenueReport(ctx context.Context, query dto.AnalyticsQuery) (*dto.RevenueReportResponse, error)

	GetDashboard(ctx context.Context) (*dto.DashboardResponse, error)
}

type service struct {
	repo          Repository
	walletService wallet.Service
}

func NewService(repo Repository, walletService wallet.Service) Service {
	return &service{
		repo:          repo,
		walletService: walletService,
	}
}

func (s *service) CreateService(ctx context.Context, req dto.CreateServiceRequest) (*dto.ServiceResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	exists, err := s.repo.ServiceSlugExists(ctx, req.ServiceSlug, "")
	if err != nil {
		logger.Error("failed to check service slug existence", "error", err, "slug", req.ServiceSlug)
		return nil, response.InternalServerError("Failed to create service", err)
	}
	if exists {
		return nil, response.ConflictError(fmt.Sprintf("Service with slug '%s' already exists", req.ServiceSlug))
	}

	svc := &models.ServiceNew{
		Title:              req.Title,
		LongTitle:          req.LongTitle,
		ServiceSlug:        req.ServiceSlug,
		CategorySlug:       req.CategorySlug,
		Description:        req.Description,
		LongDescription:    req.LongDescription,
		Highlights:         req.Highlights,
		WhatsIncluded:      pq.StringArray(req.WhatsIncluded),
		TermsAndConditions: pq.StringArray(req.TermsAndConditions),
		BannerImage:        req.BannerImage,
		Thumbnail:          req.Thumbnail,
		Duration:           req.Duration,
		IsFrequent:         req.IsFrequent,
		Frequency:          req.Frequency,
		SortOrder:          req.SortOrder,
		IsActive:           *req.IsActive,
		IsAvailable:        *req.IsAvailable,
		BasePrice:          req.BasePrice,
	}

	if err := s.repo.CreateService(ctx, svc); err != nil {
		logger.Error("failed to create service", "error", err, "slug", req.ServiceSlug)
		return nil, response.InternalServerError("Failed to create service", err)
	}

	logger.Info("service created", "serviceID", svc.ID, "slug", svc.ServiceSlug)

	return dto.ToServiceResponse(svc), nil
}

func (s *service) GetServiceBySlug(ctx context.Context, slug string) (*dto.ServiceResponse, error) {
	svc, err := s.repo.GetServiceBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Service")
		}
		logger.Error("failed to get service", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to get service", err)
	}

	return dto.ToServiceResponse(svc), nil
}

func (s *service) UpdateService(ctx context.Context, slug string, req dto.UpdateServiceRequest) (*dto.ServiceResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	svc, err := s.repo.GetServiceBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Service")
		}
		return nil, response.InternalServerError("Failed to get service", err)
	}

	if req.Title != nil {
		svc.Title = *req.Title
	}
	if req.LongTitle != nil {
		svc.LongTitle = *req.LongTitle
	}
	if req.CategorySlug != nil {
		svc.CategorySlug = *req.CategorySlug
	}
	if req.Description != nil {
		svc.Description = *req.Description
	}
	if req.LongDescription != nil {
		svc.LongDescription = *req.LongDescription
	}
	if req.Highlights != nil {
		svc.Highlights = *req.Highlights
	}
	if req.WhatsIncluded != nil {
		svc.WhatsIncluded = pq.StringArray(req.WhatsIncluded)
	}
	if req.TermsAndConditions != nil {
		svc.TermsAndConditions = pq.StringArray(req.TermsAndConditions)
	}
	if req.BannerImage != nil {
		svc.BannerImage = *req.BannerImage
	}
	if req.Thumbnail != nil {
		svc.Thumbnail = *req.Thumbnail
	}
	if req.Duration != nil {
		svc.Duration = req.Duration
	}
	if req.IsFrequent != nil {
		svc.IsFrequent = *req.IsFrequent
	}
	if req.Frequency != nil {
		svc.Frequency = *req.Frequency
	}
	if req.SortOrder != nil {
		svc.SortOrder = *req.SortOrder
	}
	if req.IsActive != nil {
		svc.IsActive = *req.IsActive
	}
	if req.IsAvailable != nil {
		svc.IsAvailable = *req.IsAvailable
	}
	if req.BasePrice != nil {
		svc.BasePrice = req.BasePrice
	}

	if err := s.repo.UpdateService(ctx, svc); err != nil {
		logger.Error("failed to update service", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to update service", err)
	}

	logger.Info("service updated", "serviceID", svc.ID, "slug", svc.ServiceSlug)

	return dto.ToServiceResponse(svc), nil
}

func (s *service) UpdateServiceStatus(ctx context.Context, slug string, req dto.UpdateServiceStatusRequest) (*dto.ServiceResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	svc, err := s.repo.GetServiceBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Service")
		}
		return nil, response.InternalServerError("Failed to get service", err)
	}

	if req.IsActive != nil {
		svc.IsActive = *req.IsActive
	}
	if req.IsAvailable != nil {
		svc.IsAvailable = *req.IsAvailable
	}

	if err := s.repo.UpdateService(ctx, svc); err != nil {
		logger.Error("failed to update service status", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to update service status", err)
	}

	logger.Info("service status updated", "serviceID", svc.ID, "slug", svc.ServiceSlug,
		"isActive", svc.IsActive, "isAvailable", svc.IsAvailable)

	return dto.ToServiceResponse(svc), nil
}

func (s *service) DeleteService(ctx context.Context, slug string) error {
	svc, err := s.repo.GetServiceBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.NotFoundError("Service")
		}
		return response.InternalServerError("Failed to get service", err)
	}

	if err := s.repo.DeleteService(ctx, svc.ID); err != nil {
		logger.Error("failed to delete service", "error", err, "slug", slug)
		return response.InternalServerError("Failed to delete service", err)
	}

	logger.Info("service deleted", "serviceID", svc.ID, "slug", slug)

	return nil
}

func (s *service) ListServices(ctx context.Context, query dto.ListServicesQuery) ([]*dto.ServiceListResponse, *response.PaginationMeta, error) {
	query.SetDefaults()

	services, total, err := s.repo.ListServices(ctx, query)
	if err != nil {
		logger.Error("failed to list services", "error", err)
		return nil, nil, response.InternalServerError("Failed to list services", err)
	}

	responses := dto.ToServiceListResponses(services)

	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

func (s *service) CreateAddon(ctx context.Context, req dto.CreateAddonRequest) (*dto.AddonResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	exists, err := s.repo.AddonSlugExists(ctx, req.AddonSlug, "")
	if err != nil {
		logger.Error("failed to check addon slug existence", "error", err, "slug", req.AddonSlug)
		return nil, response.InternalServerError("Failed to create addon", err)
	}
	if exists {
		return nil, response.ConflictError(fmt.Sprintf("Addon with slug '%s' already exists", req.AddonSlug))
	}

	addon := &models.Addon{
		Title:              req.Title,
		AddonSlug:          req.AddonSlug,
		CategorySlug:       req.CategorySlug,
		Description:        req.Description,
		WhatsIncluded:      pq.StringArray(req.WhatsIncluded),
		Notes:              pq.StringArray(req.Notes),
		Image:              req.Image,
		Price:              req.Price,
		StrikethroughPrice: req.StrikethroughPrice,
		IsActive:           *req.IsActive,
		IsAvailable:        *req.IsAvailable,
		SortOrder:          req.SortOrder,
	}

	if err := s.repo.CreateAddon(ctx, addon); err != nil {
		logger.Error("failed to create addon", "error", err, "slug", req.AddonSlug)
		return nil, response.InternalServerError("Failed to create addon", err)
	}

	logger.Info("addon created", "addonID", addon.ID, "slug", addon.AddonSlug)

	return dto.ToAddonResponse(addon), nil
}

func (s *service) GetAddonBySlug(ctx context.Context, slug string) (*dto.AddonResponse, error) {
	addon, err := s.repo.GetAddonBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Addon")
		}
		logger.Error("failed to get addon", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to get addon", err)
	}

	return dto.ToAddonResponse(addon), nil
}

func (s *service) UpdateAddon(ctx context.Context, slug string, req dto.UpdateAddonRequest) (*dto.AddonResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	addon, err := s.repo.GetAddonBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Addon")
		}
		return nil, response.InternalServerError("Failed to get addon", err)
	}

	if req.Title != nil {
		addon.Title = *req.Title
	}
	if req.CategorySlug != nil {
		addon.CategorySlug = *req.CategorySlug
	}
	if req.Description != nil {
		addon.Description = *req.Description
	}
	if req.WhatsIncluded != nil {
		addon.WhatsIncluded = pq.StringArray(req.WhatsIncluded)
	}
	if req.Notes != nil {
		addon.Notes = pq.StringArray(req.Notes)
	}
	if req.Image != nil {
		addon.Image = *req.Image
	}
	if req.Price != nil {
		addon.Price = *req.Price
	}
	if req.StrikethroughPrice != nil {
		if *req.StrikethroughPrice == 0 {
			addon.StrikethroughPrice = nil
		} else {
			addon.StrikethroughPrice = req.StrikethroughPrice
		}
	}
	if req.IsActive != nil {
		addon.IsActive = *req.IsActive
	}
	if req.IsAvailable != nil {
		addon.IsAvailable = *req.IsAvailable
	}
	if req.SortOrder != nil {
		addon.SortOrder = *req.SortOrder
	}

	if addon.StrikethroughPrice != nil && *addon.StrikethroughPrice <= addon.Price {
		return nil, response.BadRequest("strikethroughPrice must be greater than price")
	}

	if err := s.repo.UpdateAddon(ctx, addon); err != nil {
		logger.Error("failed to update addon", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to update addon", err)
	}

	logger.Info("addon updated", "addonID", addon.ID, "slug", addon.AddonSlug)

	return dto.ToAddonResponse(addon), nil
}

func (s *service) UpdateAddonStatus(ctx context.Context, slug string, req dto.UpdateAddonStatusRequest) (*dto.AddonResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	addon, err := s.repo.GetAddonBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Addon")
		}
		return nil, response.InternalServerError("Failed to get addon", err)
	}

	if req.IsActive != nil {
		addon.IsActive = *req.IsActive
	}
	if req.IsAvailable != nil {
		addon.IsAvailable = *req.IsAvailable
	}

	if err := s.repo.UpdateAddon(ctx, addon); err != nil {
		logger.Error("failed to update addon status", "error", err, "slug", slug)
		return nil, response.InternalServerError("Failed to update addon status", err)
	}

	logger.Info("addon status updated", "addonID", addon.ID, "slug", addon.AddonSlug,
		"isActive", addon.IsActive, "isAvailable", addon.IsAvailable)

	return dto.ToAddonResponse(addon), nil
}

func (s *service) DeleteAddon(ctx context.Context, slug string) error {
	addon, err := s.repo.GetAddonBySlug(ctx, slug)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return response.NotFoundError("Addon")
		}
		return response.InternalServerError("Failed to get addon", err)
	}

	if err := s.repo.DeleteAddon(ctx, addon.ID); err != nil {
		logger.Error("failed to delete addon", "error", err, "slug", slug)
		return response.InternalServerError("Failed to delete addon", err)
	}

	logger.Info("addon deleted", "addonID", addon.ID, "slug", slug)

	return nil
}

func (s *service) ListAddons(ctx context.Context, query dto.ListAddonsQuery) ([]*dto.AddonListResponse, *response.PaginationMeta, error) {
	query.SetDefaults()

	addons, total, err := s.repo.ListAddons(ctx, query)
	if err != nil {
		logger.Error("failed to list addons", "error", err)
		return nil, nil, response.InternalServerError("Failed to list addons", err)
	}

	responses := dto.ToAddonListResponses(addons)

	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

func (s *service) GetCategoryDetails(ctx context.Context, categorySlug string) (*dto.CategoryServicesResponse, error) {
	services, err := s.repo.GetServicesByCategory(ctx, categorySlug)
	if err != nil {
		logger.Error("failed to get services by category", "error", err, "category", categorySlug)
		return nil, response.InternalServerError("Failed to get category details", err)
	}

	addons, err := s.repo.GetAddonsByCategory(ctx, categorySlug)
	if err != nil {
		logger.Error("failed to get addons by category", "error", err, "category", categorySlug)
		return nil, response.InternalServerError("Failed to get category details", err)
	}

	return &dto.CategoryServicesResponse{
		CategorySlug: categorySlug,
		Services:     dto.ToServiceListResponses(services),
		Addons:       dto.ToAddonListResponses(addons),
		TotalCount:   len(services) + len(addons),
	}, nil
}

func (s *service) GetAllCategories(ctx context.Context) ([]string, error) {
	categories, err := s.repo.GetAllCategories(ctx)
	if err != nil {
		logger.Error("failed to get all categories", "error", err)
		return nil, response.InternalServerError("Failed to get categories", err)
	}
	return categories, nil
}

func (s *service) GetOrders(ctx context.Context, query dto.ListOrdersQuery) ([]dto.AdminOrderListResponse, *response.PaginationMeta, error) {
	if err := query.Validate(); err != nil {
		return nil, nil, response.BadRequest(err.Error())
	}

	query.SetDefaults()

	orders, total, err := s.repo.GetOrders(ctx, query)
	if err != nil {
		logger.Error("failed to get orders", "error", err)
		return nil, nil, response.InternalServerError("Failed to get orders", err)
	}

	responses := dto.ToAdminOrderListResponses(orders)

	pagination := response.NewPaginationMeta(total, query.Page, query.Limit)

	return responses, &pagination, nil
}

func (s *service) GetOrderByID(ctx context.Context, orderID string) (*dto.AdminOrderDetailResponse, error) {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		logger.Error("failed to get order", "error", err, "orderID", orderID)
		return nil, response.InternalServerError("Failed to get order", err)
	}

	history, err := s.repo.GetOrderStatusHistory(ctx, orderID)
	if err != nil {
		logger.Error("failed to get order history", "error", err, "orderID", orderID)
		history = []models.OrderStatusHistory{}
	}

	response := dto.ToAdminOrderDetailResponse(order, history)

	if response.Provider != nil && response.Provider.ID != "" {
		provider, err := s.repo.GetUserByID(ctx, response.Provider.ID)
		if err == nil && provider != nil {
			response.Provider.Name = provider.Name
			if provider.Email != nil {
				response.Provider.Email = *provider.Email
			}
			if provider.Phone != nil {
				response.Provider.Phone = *provider.Phone
			}
			if provider.ProfilePhotoURL != nil {
				response.Provider.Photo = *provider.ProfilePhotoURL
			}
		} else {
			logger.Warn("failed to fetch provider details", "error", err, "providerID", response.Provider.ID)
		}
	}

	return response, nil
}

func (s *service) GetOrderByNumber(ctx context.Context, orderNumber string) (*dto.AdminOrderDetailResponse, error) {
	order, err := s.repo.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		logger.Error("failed to get order", "error", err, "orderNumber", orderNumber)
		return nil, response.InternalServerError("Failed to get order", err)
	}

	history, err := s.repo.GetOrderStatusHistory(ctx, order.ID)
	if err != nil {
		logger.Error("failed to get order history", "error", err, "orderID", order.ID)
		history = []models.OrderStatusHistory{}
	}

	return dto.ToAdminOrderDetailResponse(order, history), nil
}

func (s *service) UpdateOrderStatus(ctx context.Context, orderID string, req dto.UpdateOrderStatusRequest, adminID string) (*dto.AdminOrderDetailResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to get order", err)
	}

	if !shared.CanTransition(order.Status, req.Status) {
		return nil, response.BadRequest(fmt.Sprintf("Cannot transition from '%s' to '%s'", order.Status, req.Status))
	}

	previousStatus := order.Status

	order.Status = req.Status
	now := time.Now()

	switch req.Status {
	case shared.OrderStatusAccepted:
		if order.ProviderAcceptedAt == nil {
			order.ProviderAcceptedAt = &now
		}
	case shared.OrderStatusInProgress:
		if order.ProviderStartedAt == nil {
			order.ProviderStartedAt = &now
		}
	case shared.OrderStatusCompleted:
		if order.CompletedAt == nil {
			order.CompletedAt = &now
		}
		if order.ProviderCompletedAt == nil {
			order.ProviderCompletedAt = &now
		}
		if order.PaymentInfo != nil {
			order.PaymentInfo.Status = shared.PaymentStatusCompleted
			order.PaymentInfo.AmountPaid = order.TotalPrice
		}
	}

	if err := s.repo.UpdateOrder(ctx, order); err != nil {
		logger.Error("failed to update order status", "error", err, "orderID", orderID)
		return nil, response.InternalServerError("Failed to update order status", err)
	}

	notes := req.Reason
	if notes == "" {
		notes = fmt.Sprintf("Status changed by admin from %s to %s", previousStatus, req.Status)
	}
	if req.Notes != "" {
		notes += ". Notes: " + req.Notes
	}

	history := models.NewOrderStatusHistory(
		order.ID,
		previousStatus,
		req.Status,
		&adminID,
		shared.RoleAdmin,
		notes,
		nil,
	)
	s.repo.CreateStatusHistory(ctx, history)

	logger.Info("order status updated by admin",
		"orderID", orderID,
		"adminID", adminID,
		"fromStatus", previousStatus,
		"toStatus", req.Status,
	)

	return s.GetOrderByID(ctx, orderID)
}

func (s *service) ReassignOrder(ctx context.Context, orderID string, req dto.ReassignOrderRequest, adminID string) (*dto.AdminOrderDetailResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to get order", err)
	}

	reassignableStatuses := []string{
		shared.OrderStatusAssigned,
		shared.OrderStatusAccepted,
	}
	canReassign := false
	for _, status := range reassignableStatuses {
		if order.Status == status {
			canReassign = true
			break
		}
	}
	if !canReassign {
		return nil, response.BadRequest(fmt.Sprintf("Cannot reassign order in '%s' status", order.Status))
	}

	oldProviderID := order.AssignedProviderID

	order.AssignedProviderID = &req.ProviderID
	order.Status = shared.OrderStatusAssigned
	order.ProviderAcceptedAt = nil

	if err := s.repo.UpdateOrder(ctx, order); err != nil {
		logger.Error("failed to reassign order", "error", err, "orderID", orderID)
		return nil, response.InternalServerError("Failed to reassign order", err)
	}

	metadata := models.StatusHistoryMetadata{
		"newProviderId": req.ProviderID,
	}
	if oldProviderID != nil {
		metadata["oldProviderId"] = *oldProviderID
	}

	history := models.NewOrderStatusHistory(
		order.ID,
		order.Status,
		shared.OrderStatusAssigned,
		&adminID,
		shared.RoleAdmin,
		fmt.Sprintf("Order reassigned by admin. Reason: %s", req.Reason),
		metadata,
	)
	s.repo.CreateStatusHistory(ctx, history)

	logger.Info("order reassigned by admin",
		"orderID", orderID,
		"adminID", adminID,
		"newProviderID", req.ProviderID,
	)

	return s.GetOrderByID(ctx, orderID)
}

func (s *service) CancelOrder(ctx context.Context, orderID string, req dto.AdminCancelOrderRequest, adminID string) (*dto.AdminOrderDetailResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, response.NotFoundError("Order")
		}
		return nil, response.InternalServerError("Failed to get order", err)
	}

	if order.Status == shared.OrderStatusCancelled {
		return nil, response.BadRequest("Order is already cancelled")
	}

	if order.Status == shared.OrderStatusCompleted {
		return nil, response.BadRequest("Cannot cancel a completed order. Use refund instead.")
	}

	var refundAmount float64
	var cancellationFee float64

	if req.RefundAmount != nil {
		refundAmount = *req.RefundAmount
		if refundAmount > order.TotalPrice {
			return nil, response.BadRequest("Refund amount cannot exceed order total")
		}
		cancellationFee = order.TotalPrice - refundAmount
	} else {
		cancellationFee, refundAmount = shared.CalculateCancellationFee(order.Status, order.TotalPrice)
	}
	previousStatus := order.Status

	if order.WalletHoldID != nil && refundAmount > 0 {
		releaseReq := walletdto.ReleaseHoldRequest{HoldID: *order.WalletHoldID}
		if err := s.walletService.ReleaseHold(ctx, order.CustomerID, releaseReq); err != nil {
			logger.Error("failed to release wallet hold", "error", err, "holdID", *order.WalletHoldID)
		}

		if cancellationFee > 0 {
			metadata := map[string]interface{}{
				"order_id":     order.ID,
				"order_number": order.OrderNumber,
			}
			if _, err := s.walletService.DebitWallet(
				ctx,
				order.CustomerID,
				cancellationFee,
				"admin_cancellation_fee",
				order.ID,
				fmt.Sprintf("Cancellation fee for order %s (cancelled by admin)", order.OrderNumber),
				metadata,
			); err != nil {
				logger.Error("failed to debit cancellation fee", "error", err, "orderID", orderID)
			}
		}
	}

	now := time.Now()
	order.Status = shared.OrderStatusCancelled
	order.CancellationInfo = &models.CancellationInfo{
		CancelledBy:     shared.CancelledByAdmin,
		CancelledAt:     now,
		Reason:          req.Reason,
		CancellationFee: cancellationFee,
		RefundAmount:    refundAmount,
	}

	if order.PaymentInfo != nil {
		if refundAmount > 0 {
			order.PaymentInfo.Status = shared.PaymentStatusRefunded
		} else {
			order.PaymentInfo.Status = shared.PaymentStatusCompleted
		}
	}

	if err := s.repo.UpdateOrder(ctx, order); err != nil {
		logger.Error("failed to cancel order", "error", err, "orderID", orderID)
		return nil, response.InternalServerError("Failed to cancel order", err)
	}

	history := models.NewOrderStatusHistory(
		order.ID,
		previousStatus,
		shared.OrderStatusCancelled,
		&adminID,
		shared.RoleAdmin,
		fmt.Sprintf("Cancelled by admin. Reason: %s", req.Reason),
		models.StatusHistoryMetadata{
			"cancellationFee": cancellationFee,
			"refundAmount":    refundAmount,
			"notifyParties":   req.NotifyParties,
		},
	)
	s.repo.CreateStatusHistory(ctx, history)

	logger.Info("order cancelled by admin",
		"orderID", orderID,
		"adminID", adminID,
		"cancellationFee", cancellationFee,
		"refundAmount", refundAmount,
	)

	return s.GetOrderByID(ctx, orderID)
}

func (s *service) BulkUpdateStatus(ctx context.Context, req dto.BulkUpdateStatusRequest, adminID string) (int64, error) {
	if err := req.Validate(); err != nil {
		return 0, response.BadRequest(err.Error())
	}

	affected, err := s.repo.BulkUpdateStatus(ctx, req.OrderIDs, req.Status, adminID, req.Reason)
	if err != nil {
		logger.Error("failed to bulk update orders", "error", err)
		return 0, response.InternalServerError("Failed to update orders", err)
	}

	logger.Info("bulk status update completed",
		"adminID", adminID,
		"status", req.Status,
		"requested", len(req.OrderIDs),
		"affected", affected,
	)

	return affected, nil
}

func (s *service) GetOverviewAnalytics(ctx context.Context, query dto.AnalyticsQuery) (*dto.OverviewAnalyticsResponse, error) {
	if err := query.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	query.SetDefaults()

	fromDate, _ := time.Parse("2006-01-02", query.FromDate)
	toDate, _ := time.Parse("2006-01-02", query.ToDate)

	stats, err := s.repo.GetOrderStats(ctx, fromDate, toDate)
	if err != nil {
		logger.Error("failed to get order stats", "error", err)
		return nil, response.InternalServerError("Failed to get analytics", err)
	}

	statusStats, err := s.repo.GetOrdersByStatus(ctx, fromDate, toDate)
	if err != nil {
		logger.Error("failed to get status stats", "error", err)
		return nil, response.InternalServerError("Failed to get analytics", err)
	}

	categoryStats, err := s.repo.GetOrdersByCategory(ctx, fromDate, toDate)
	if err != nil {
		logger.Error("failed to get category stats", "error", err)
		return nil, response.InternalServerError("Failed to get analytics", err)
	}

	revenueBreakdown, err := s.repo.GetRevenueBreakdown(ctx, fromDate, toDate, query.GroupBy)
	if err != nil {
		logger.Error("failed to get revenue breakdown", "error", err)
		return nil, response.InternalServerError("Failed to get analytics", err)
	}

	var completionRate, cancellationRate, averageRating, averageOrderValue float64

	if stats.TotalOrders > 0 {
		completionRate = float64(stats.CompletedOrders) / float64(stats.TotalOrders) * 100
		cancellationRate = float64(stats.CancelledOrders) / float64(stats.TotalOrders) * 100
	}

	if stats.TotalRatings > 0 {
		averageRating = float64(stats.TotalRatingSum) / float64(stats.TotalRatings)
	}

	if stats.CompletedOrders > 0 {
		averageOrderValue = stats.TotalRevenue / float64(stats.CompletedOrders)
	}

	response := &dto.OverviewAnalyticsResponse{
		Period: dto.AnalyticsPeriod{
			FromDate: query.FromDate,
			ToDate:   query.ToDate,
			GroupBy:  query.GroupBy,
		},
		Summary: dto.AnalyticsSummary{
			TotalOrders:          int(stats.TotalOrders),
			CompletedOrders:      int(stats.CompletedOrders),
			CancelledOrders:      int(stats.CancelledOrders),
			PendingOrders:        int(stats.PendingOrders),
			TotalRevenue:         stats.TotalRevenue,
			TotalCommission:      stats.TotalCommission,
			TotalProviderPayouts: stats.TotalProviderPayouts,
			AverageOrderValue:    averageOrderValue,
			CompletionRate:       completionRate,
			CancellationRate:     cancellationRate,
			AverageRating:        averageRating,
		},
	}

	for _, ss := range statusStats {
		percentage := 0.0
		if stats.TotalOrders > 0 {
			percentage = float64(ss.Count) / float64(stats.TotalOrders) * 100
		}
		response.OrdersByStatus = append(response.OrdersByStatus, dto.StatusCount{
			Status:      ss.Status,
			DisplayName: dto.GetDisplayStatus(ss.Status),
			Count:       int(ss.Count),
			Percentage:  percentage,
		})
	}

	for _, cs := range categoryStats {
		percentage := 0.0
		if stats.TotalRevenue > 0 {
			percentage = cs.Revenue / stats.TotalRevenue * 100
		}
		response.OrdersByCategory = append(response.OrdersByCategory, dto.CategoryCount{
			CategorySlug:  cs.CategorySlug,
			CategoryTitle: dto.GetCategoryTitle(cs.CategorySlug),
			OrderCount:    int(cs.OrderCount),
			Revenue:       cs.Revenue,
			Percentage:    percentage,
		})
	}

	for _, rb := range revenueBreakdown {
		avgValue := 0.0
		if rb.OrderCount > 0 {
			avgValue = rb.Revenue / float64(rb.OrderCount)
		}
		response.RevenueBreakdown = append(response.RevenueBreakdown, dto.RevenueBreakdown{
			Period:            rb.Period,
			OrderCount:        int(rb.OrderCount),
			Revenue:           rb.Revenue,
			Commission:        rb.Commission,
			ProviderPayouts:   rb.ProviderPayouts,
			AverageOrderValue: avgValue,
		})
	}

	periodDuration := toDate.Sub(fromDate)
	previousFromDate := fromDate.Add(-periodDuration)
	previousToDate := fromDate.AddDate(0, 0, -1)

	previousStats, _ := s.repo.GetOrderStats(ctx, previousFromDate, previousToDate)
	if previousStats != nil {
		response.Trends = s.calculateTrends(stats, previousStats)
	}

	return response, nil
}

func (s *service) calculateTrends(current, previous *OrderStats) dto.AnalyticsTrends {
	trends := dto.AnalyticsTrends{}

	trends.OrdersChange = s.calculateTrendChange(
		float64(current.CompletedOrders),
		float64(previous.CompletedOrders),
	)

	trends.RevenueChange = s.calculateTrendChange(
		current.TotalRevenue,
		previous.TotalRevenue,
	)

	var currentRate, previousRate float64
	if current.TotalOrders > 0 {
		currentRate = float64(current.CompletedOrders) / float64(current.TotalOrders) * 100
	}
	if previous.TotalOrders > 0 {
		previousRate = float64(previous.CompletedOrders) / float64(previous.TotalOrders) * 100
	}
	trends.CompletionChange = s.calculateTrendChange(currentRate, previousRate)

	return trends
}

func (s *service) calculateTrendChange(current, previous float64) dto.TrendChange {
	change := dto.TrendChange{
		CurrentValue:  current,
		PreviousValue: previous,
	}

	change.Change = current - previous

	if previous > 0 {
		change.ChangePercent = (change.Change / previous) * 100
	}

	if change.Change > 0 {
		change.Trend = "up"
	} else if change.Change < 0 {
		change.Trend = "down"
	} else {
		change.Trend = "stable"
	}

	return change
}

func (s *service) GetProviderAnalytics(ctx context.Context, query dto.ProviderAnalyticsQuery) (*dto.ProviderAnalyticsResponse, error) {
	fromDate, err := time.Parse("2006-01-02", query.FromDate)
	if err != nil {
		return nil, response.BadRequest("Invalid fromDate format")
	}
	toDate, err := time.Parse("2006-01-02", query.ToDate)
	if err != nil {
		return nil, response.BadRequest("Invalid toDate format")
	}

	query.SetDefaults()

	stats, err := s.repo.GetProviderAnalytics(ctx, fromDate, toDate, query)
	if err != nil {
		logger.Error("failed to get provider analytics", "error", err)
		return nil, response.InternalServerError("Failed to get analytics", err)
	}

	providers := make([]dto.ProviderAnalyticsItem, len(stats))
	for i, ps := range stats {
		var avgRating float64
		if ps.TotalRatings > 0 {
			avgRating = float64(ps.TotalRatingSum) / float64(ps.TotalRatings)
		}

		var completionRate float64
		totalOrders := ps.CompletedOrders + ps.CancelledOrders
		if totalOrders > 0 {
			completionRate = float64(ps.CompletedOrders) / float64(totalOrders) * 100
		}
		providerName := "Provider"
		providerPhoto := ""
		if provider, err := s.repo.GetUserByID(ctx, ps.ProviderID); err == nil && provider != nil {
			providerName = provider.Name
			if provider.ProfilePhotoURL != nil {
				providerPhoto = *provider.ProfilePhotoURL
			}
		} else {
			logger.Warn("failed to fetch provider details for analytics", "error", err, "providerID", ps.ProviderID)
		}

		providers[i] = dto.ProviderAnalyticsItem{
			ProviderID:      ps.ProviderID,
			ProviderName:    providerName,
			Photo:           providerPhoto,
			CompletedOrders: int(ps.CompletedOrders),
			CancelledOrders: int(ps.CancelledOrders),
			TotalEarnings:   ps.TotalEarnings,
			AverageRating:   avgRating,
			TotalRatings:    int(ps.TotalRatings),
			CompletionRate:  completionRate,
		}
	}

	return &dto.ProviderAnalyticsResponse{
		Period: dto.AnalyticsPeriod{
			FromDate: query.FromDate,
			ToDate:   query.ToDate,
		},
		Providers: providers,
		Total:     len(providers),
	}, nil
}

func (s *service) GetRevenueReport(ctx context.Context, query dto.AnalyticsQuery) (*dto.RevenueReportResponse, error) {
	if err := query.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	query.SetDefaults()

	fromDate, _ := time.Parse("2006-01-02", query.FromDate)
	toDate, _ := time.Parse("2006-01-02", query.ToDate)

	stats, err := s.repo.GetOrderStats(ctx, fromDate, toDate)
	if err != nil {
		logger.Error("failed to get order stats", "error", err)
		return nil, response.InternalServerError("Failed to get revenue report", err)
	}

	revenueBreakdown, err := s.repo.GetRevenueBreakdown(ctx, fromDate, toDate, query.GroupBy)
	if err != nil {
		logger.Error("failed to get revenue breakdown", "error", err)
		return nil, response.InternalServerError("Failed to get revenue report", err)
	}

	categoryStats, err := s.repo.GetOrdersByCategory(ctx, fromDate, toDate)
	if err != nil {
		logger.Error("failed to get category stats", "error", err)
		return nil, response.InternalServerError("Failed to get revenue report", err)
	}

	paymentStats, err := s.repo.GetPaymentMethodStats(ctx, fromDate, toDate)
	if err != nil {
		logger.Error("failed to get payment stats", "error", err)
		return nil, response.InternalServerError("Failed to get revenue report", err)
	}

	totalRefunds, err := s.repo.GetTotalRefunds(ctx, fromDate, toDate)
	if err != nil {
		logger.Error("failed to get total refunds", "error", err)
		totalRefunds = 0
	}

	response := &dto.RevenueReportResponse{
		Period: dto.AnalyticsPeriod{
			FromDate: query.FromDate,
			ToDate:   query.ToDate,
			GroupBy:  query.GroupBy,
		},
		TotalRevenue:     stats.TotalRevenue,
		TotalCommission:  stats.TotalCommission,
		TotalPayouts:     stats.TotalProviderPayouts,
		TotalRefunds:     totalRefunds,
		NetRevenue:       stats.TotalCommission - totalRefunds,
		FormattedRevenue: dto.FormatPrice(stats.TotalRevenue),
	}

	for _, rb := range revenueBreakdown {
		avgValue := 0.0
		if rb.OrderCount > 0 {
			avgValue = rb.Revenue / float64(rb.OrderCount)
		}
		response.Breakdown = append(response.Breakdown, dto.RevenueBreakdown{
			Period:            rb.Period,
			OrderCount:        int(rb.OrderCount),
			Revenue:           rb.Revenue,
			Commission:        rb.Commission,
			ProviderPayouts:   rb.ProviderPayouts,
			AverageOrderValue: avgValue,
		})
	}

	for _, cs := range categoryStats {
		percentage := 0.0
		if stats.TotalRevenue > 0 {
			percentage = cs.Revenue / stats.TotalRevenue * 100
		}
		commission := cs.Revenue * shared.PlatformCommissionRate
		response.ByCategory = append(response.ByCategory, dto.CategoryRevenue{
			CategorySlug:  cs.CategorySlug,
			CategoryTitle: dto.GetCategoryTitle(cs.CategorySlug),
			Revenue:       cs.Revenue,
			Commission:    commission,
			OrderCount:    int(cs.OrderCount),
			Percentage:    percentage,
		})
	}

	for _, ps := range paymentStats {
		percentage := 0.0
		if stats.TotalRevenue > 0 {
			percentage = ps.TotalAmount / stats.TotalRevenue * 100
		}
		response.ByPaymentMethod = append(response.ByPaymentMethod, dto.PaymentMethodStats{
			Method:      ps.Method,
			OrderCount:  int(ps.OrderCount),
			TotalAmount: ps.TotalAmount,
			Percentage:  percentage,
		})
	}

	return response, nil
}

func (s *service) GetDashboard(ctx context.Context) (*dto.DashboardResponse, error) {
	todayData, err := s.repo.GetTodayStats(ctx)
	if err != nil {
		logger.Error("failed to get today stats", "error", err)
		return nil, response.InternalServerError("Failed to get dashboard", err)
	}

	weeklyData, err := s.repo.GetWeeklyStats(ctx)
	if err != nil {
		logger.Error("failed to get weekly stats", "error", err)
		return nil, response.InternalServerError("Failed to get dashboard", err)
	}

	pendingData, err := s.repo.GetPendingActions(ctx)
	if err != nil {
		logger.Error("failed to get pending actions", "error", err)
		return nil, response.InternalServerError("Failed to get dashboard", err)
	}

	recentOrders, err := s.repo.GetRecentOrders(ctx, 10)
	if err != nil {
		logger.Error("failed to get recent orders", "error", err)
		return nil, response.InternalServerError("Failed to get dashboard", err)
	}

	dashboard := &dto.DashboardResponse{
		Today: dto.TodayStats{
			TotalOrders:      int(todayData.TotalOrders),
			CompletedOrders:  int(todayData.CompletedOrders),
			PendingOrders:    int(todayData.PendingOrders),
			InProgressOrders: int(todayData.InProgressOrders),
			Revenue:          todayData.Revenue,
			Commission:       todayData.Commission,
		},
		RecentOrders: dto.ToAdminOrderListResponses(recentOrders),
		PendingActions: dto.PendingActions{
			OrdersNeedingProvider: int(pendingData.OrdersNeedingProvider),
			ExpiredOrders:         int(pendingData.ExpiredOrders),
		},
	}

	var avgRating float64
	if weeklyData.TotalRatings > 0 {
		avgRating = float64(weeklyData.TotalRatingSum) / float64(weeklyData.TotalRatings)
	}

	dashboard.WeeklyStats = dto.WeeklyStats{
		TotalOrders:     int(weeklyData.TotalOrders),
		CompletedOrders: int(weeklyData.CompletedOrders),
		TotalRevenue:    weeklyData.TotalRevenue,
		AverageRating:   avgRating,
	}

	for _, ds := range weeklyData.DailyStats {
		date, _ := time.Parse("2006-01-02", ds.Date)
		dashboard.WeeklyStats.DailyBreakdown = append(dashboard.WeeklyStats.DailyBreakdown, dto.DailyStats{
			Date:       ds.Date,
			DayName:    date.Weekday().String(),
			OrderCount: int(ds.OrderCount),
			Revenue:    ds.Revenue,
		})
	}

	return dashboard, nil
}
