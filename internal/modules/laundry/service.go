package laundry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/laundry/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"gorm.io/gorm"
)

type Service interface {
	// Catalog
	GetServiceCatalog(ctx context.Context) ([]*models.LaundryServiceCatalog, error)
	GetServicesWithProducts(ctx context.Context) ([]*dto.LaundryServiceDTO, error)
	GetServiceProducts(ctx context.Context, serviceSlug string) ([]*models.LaundryServiceProduct, error)

	// Orders
	CreateOrder(ctx context.Context, customerID string, req *dto.CreateLaundryOrderRequest) (*models.LaundryOrder, error)
	GetOrder(ctx context.Context, orderID string) (*dto.LaundryOrderResponse, error)
	GetOrderWithDetails(ctx context.Context, orderID string) (*models.LaundryOrder, error)
	GetAvailableOrders(ctx context.Context, providerID string) ([]*models.LaundryOrder, error)

	// Pickups
	InitiatePickup(ctx context.Context, orderID string, providerID string) (*models.LaundryPickup, error)
	CompletePickup(ctx context.Context, orderID string, req *dto.CompletePickupRequest) error
	GetProviderPickups(ctx context.Context, providerID string) ([]*models.LaundryPickup, error)

	// Items
	AddItems(ctx context.Context, orderID string, req *dto.AddLaundryItemsRequest) ([]*models.LaundryOrderItem, error)
	UpdateItemStatus(ctx context.Context, qrCode, status string) (*models.LaundryOrderItem, error)
	GetOrderItems(ctx context.Context, orderID string) ([]*models.LaundryOrderItem, error)

	// Deliveries
	InitiateDelivery(ctx context.Context, orderID string, providerID string) (*models.LaundryDelivery, error)
	CompleteDelivery(ctx context.Context, orderID string, req *dto.CompleteDeliveryRequest) error
	GetProviderDeliveries(ctx context.Context, providerID string) ([]*models.LaundryDelivery, error)

	// Issues
	ReportIssue(ctx context.Context, orderID, customerID, providerID string, req *dto.ReportIssueRequest) (*models.LaundryIssue, error)
	GetProviderIssues(ctx context.Context, providerID string) ([]*models.LaundryIssue, error)
	ResolveIssue(ctx context.Context, issueID string, resolution string, refundAmount *float64) error
}

type service struct {
	repo Repository
	db   *gorm.DB
}

func NewService(repo Repository, db *gorm.DB) Service {
	return &service{repo: repo, db: db}
}

// =====================================================
// Catalog
// =====================================================

func (s *service) GetServiceCatalog(ctx context.Context) ([]*models.LaundryServiceCatalog, error) {
	return s.repo.GetServiceCatalog(ctx)
}

func (s *service) GetServicesWithProducts(ctx context.Context) ([]*dto.LaundryServiceDTO, error) {
	services, err := s.repo.GetServicesWithProducts(ctx)
	if err != nil {
		logger.Error("GetServicesWithProducts: failed to fetch services", "error", err)
		return nil, fmt.Errorf("failed to fetch services: %w", err)
	}

	// Convert to DTO
	result := make([]*dto.LaundryServiceDTO, 0, len(services))
	for _, service := range services {
		serviceDTO := &dto.LaundryServiceDTO{
			ID:              service.ID,
			Slug:            service.Slug,
			Title:           service.Title,
			Description:     service.Description,
			ColorCode:       service.ColorCode,
			BasePrice:       service.BasePrice,
			PricingUnit:     service.PricingUnit,
			TurnaroundHours: service.TurnaroundHours,
			ExpressFee:      service.ExpressFee,
			ExpressHours:    service.ExpressHours,
			CategorySlug:    service.CategorySlug,
			IsActive:        service.IsActive,
			ProductCount:    len(service.Products),
			Products:        make([]dto.ProductResponse, 0, len(service.Products)),
		}

		for _, product := range service.Products {
			productDTO := dto.ProductResponse{
				ID:                  product.ID,
				Name:                product.Name,
				Slug:                product.Slug,
				Description:         product.Description,
				IconURL:             product.IconURL,
				Price:               product.Price,
				PricingUnit:         product.PricingUnit,
				TypicalWeight:       product.TypicalWeight,
				RequiresSpecialCare: product.RequiresSpecialCare,
				SpecialCareFee:      product.SpecialCareFee,
				CategorySlug:        product.CategorySlug,
			}
			serviceDTO.Products = append(serviceDTO.Products, productDTO)
		}

		result = append(result, serviceDTO)
	}

	return result, nil
}

func (s *service) GetServiceProducts(ctx context.Context, serviceSlug string) ([]*models.LaundryServiceProduct, error) {
	return s.repo.GetServiceProducts(ctx, serviceSlug)
}

// =====================================================
// Orders
// =====================================================

func (s *service) CreateOrder(ctx context.Context, customerID string, req *dto.CreateLaundryOrderRequest) (*models.LaundryOrder, error) {
	// Validate request
	if req == nil {
		logger.Error("CreateOrder: request is nil", "customerID", customerID)
		return nil, errors.New("request is required")
	}

	// Validate all required fields
	if err := req.Validate(); err != nil {
		logger.Error("CreateOrder: request validation failed",
			"error", err,
			"customerID", customerID,
			"address", req.Address,
		)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get service catalog to verify service exists
	service, err := s.repo.GetServiceBySlug(ctx, req.ServiceSlug)
	if err != nil {
		logger.Error("CreateOrder: service not found",
			"error", err,
			"serviceSlug", req.ServiceSlug,
		)
		return nil, fmt.Errorf("service not found: %w", err)
	}

	// Calculate total price from products
	totalPrice := 0.0

	for _, item := range req.Items {
		// Get product details
		product, err := s.repo.GetProductBySlug(ctx, req.ServiceSlug, item.ProductSlug)
		if err != nil {
			logger.Error("CreateOrder: product not found",
				"error", err,
				"productSlug", item.ProductSlug,
			)
			return nil, fmt.Errorf("product '%s' not found", item.ProductSlug)
		}

		// Calculate price based on pricing unit
		itemPrice := 0.0

		if service.PricingUnit == "kg" {
			// For weight-based services, only use product price (no base price per kg)
			if product.Price != nil {
				itemPrice = *product.Price * float64(item.Quantity)
			}
		} else {
			// Item-based pricing: base_price + product_price per item
			price := service.BasePrice
			if product.Price != nil {
				price += *product.Price // Add product price to base price
			}
			itemPrice = price * float64(item.Quantity)
		}

		// Add special care fee if required
		if product.RequiresSpecialCare {
			itemPrice += product.SpecialCareFee * float64(item.Quantity)
		}

		totalPrice += itemPrice
	}

	// Add express fee if requested
	if req.IsExpress {
		totalPrice += service.ExpressFee
	}

	// Add tip if provided
	if req.Tip != nil && *req.Tip > 0 {
		totalPrice += *req.Tip
	}

	logger.Info("CreateOrder: calculated pricing",
		"customerID", customerID,
		"totalPrice", totalPrice,
		"isExpress", req.IsExpress,
	)

	// In a real system, this could be based on availability, location, ratings, etc.
	logger.Info("CreateOrder: preparing order creation",
		"customerID", customerID,
		"serviceSlug", req.ServiceSlug,
		"totalPrice", totalPrice,
	)

	// Create service order
	orderID := uuid.New().String()
	now := time.Now()
	order := &models.LaundryOrder{
		ID:           orderID,
		OrderNumber:  fmt.Sprintf("LDY-%d", time.Now().Unix()),
		UserID:       &customerID,
		CategorySlug: "laundry",
		Status:       "pending",
		Address:      req.Address,
		Latitude:     req.Lat,
		Longitude:    req.Lng,
		ServiceDate:  nil, // Will be set when pickup is created
		Total:        totalPrice,
		Tip:          req.Tip,       // Store the tip
		IsExpress:    req.IsExpress, // Store the express flag
		ProviderID:   nil,           // Will be assigned when provider accepts
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.db.WithContext(ctx).Create(order).Error; err != nil {
		logger.Error("CreateOrder: failed to create order in database",
			"error", err,
			"customerID", customerID,
			"orderID", orderID,
			"total", totalPrice,
		)
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	logger.Info("CreateOrder: order created successfully",
		"orderID", orderID,
		"customerID", customerID,
		"totalPrice", totalPrice,
	)

	// Create order items with pricing
	items := make([]*models.LaundryOrderItem, len(req.Items))
	for i, item := range req.Items {
		product, _ := s.repo.GetProductBySlug(ctx, req.ServiceSlug, item.ProductSlug)

		// Calculate item price (use same logic as total calculation)
		itemPrice := 0.0
		if service.PricingUnit == "kg" {
			// For weight-based services, only use product price (no base price per kg)
			if product.Price != nil {
				itemPrice = *product.Price * float64(item.Quantity)
			}
		} else {
			// Item-based pricing: base_price + product_price per item
			price := service.BasePrice
			if product.Price != nil {
				price += *product.Price
			}
			itemPrice = price * float64(item.Quantity)
		}

		if product.RequiresSpecialCare {
			itemPrice += product.SpecialCareFee * float64(item.Quantity)
		}

		items[i] = &models.LaundryOrderItem{
			OrderID:     orderID,
			ServiceSlug: req.ServiceSlug,
			ProductSlug: item.ProductSlug,
			ItemType:    item.ProductSlug, // Use product slug as item type
			Quantity:    item.Quantity,
			Weight:      item.Weight,
			Status:      "pending",
			Price:       itemPrice,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	if err := s.repo.CreateItems(ctx, items); err != nil {
		logger.Error("CreateOrder: failed to create order items",
			"error", err,
			"orderID", orderID,
			"itemCount", len(items),
		)
		return nil, fmt.Errorf("failed to create order items: %w", err)
	}

	logger.Info("CreateOrder: order items created",
		"orderID", orderID,
		"itemCount", len(items),
	)

	// Create pickup event - will be assigned to provider when they accept
	// Parse pickup time - handle format "11:00 AM - 12:00 PM" by extracting just the start time
	startTime := req.PickupTime
	if strings.Contains(startTime, "-") {
		// Extract the start time from the range
		parts := strings.Split(startTime, "-")
		startTime = strings.TrimSpace(parts[0])
	}

	// Parse the date and time
	pickupDateTime, err := time.Parse("2006-01-02 3:04 PM", fmt.Sprintf("%s %s", req.PickupDate, startTime))
	if err != nil {
		// Try alternative format without spaces
		pickupDateTime, err = time.Parse("2006-01-0215:04", fmt.Sprintf("%s%s", req.PickupDate, startTime))
		if err != nil {
			logger.Info("CreateOrder: failed to parse pickup datetime, using default",
				"error", err,
				"providedDate", req.PickupDate,
				"providedTime", req.PickupTime,
			)
			// Fallback to current time + 2 hours
			pickupDateTime = time.Now().Add(2 * time.Hour)
		}
	}

	// For now, create pickup without provider assignment - will be assigned when provider accepts
	pickup := &models.LaundryPickup{
		OrderID:     orderID,
		ProviderID:  nil, // Will be assigned when provider accepts
		ScheduledAt: pickupDateTime,
		Status:      "scheduled",
		Notes:       req.SpecialNotes,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreatePickup(ctx, pickup); err != nil {
		logger.Error("CreateOrder: failed to create pickup",
			"error", err,
			"orderID", orderID,
			"pickupDateTime", pickupDateTime,
		)
		return nil, fmt.Errorf("failed to create pickup: %w", err)
	}

	logger.Info("CreateOrder: pickup scheduled",
		"orderID", orderID,
		"pickupDateTime", pickupDateTime,
	)

	// Create delivery event (scheduled for turnaround time after pickup)
	turnaroundHours := service.TurnaroundHours
	if req.IsExpress {
		turnaroundHours = service.ExpressHours
	}
	turnaroundDuration := time.Duration(turnaroundHours) * time.Hour

	deliveryDateTime := pickupDateTime.Add(turnaroundDuration)

	delivery := &models.LaundryDelivery{
		OrderID:     orderID,
		ProviderID:  nil, // Will be assigned when provider accepts
		ScheduledAt: deliveryDateTime,
		Status:      "scheduled",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateDelivery(ctx, delivery); err != nil {
		logger.Error("CreateOrder: failed to create delivery",
			"error", err,
			"orderID", orderID,
			"deliveryDateTime", deliveryDateTime,
		)
		return nil, fmt.Errorf("failed to create delivery: %w", err)
	}

	logger.Info("CreateOrder: delivery scheduled",
		"orderID", orderID,
		"deliveryDateTime", deliveryDateTime,
		"isExpress", req.IsExpress,
		"turnaroundHours", turnaroundHours,
	)

	logger.Info("CreateOrder: order creation completed successfully",
		"orderID", orderID,
		"customerID", customerID,
		"orderNumber", order.OrderNumber,
		"status", order.Status,
	)

	return order, nil
}

func (s *service) GetOrder(ctx context.Context, orderID string) (*dto.LaundryOrderResponse, error) {
	order, err := s.GetOrderWithDetails(ctx, orderID)
	if err != nil {
		return nil, err
	}

	pickup, _ := s.repo.GetPickupByOrder(ctx, orderID)
	delivery, _ := s.repo.GetDeliveryByOrder(ctx, orderID)
	items, _ := s.repo.GetOrderItems(ctx, orderID)

	// Convert items to DTO
	itemDTOs := make([]dto.LaundryOrderItemDTO, 0, len(items))
	for _, item := range items {
		itemDTO := dto.LaundryOrderItemDTO{
			ID:               item.ID,
			OrderID:          item.OrderID,
			ProductSlug:      item.ProductSlug,
			ItemType:         item.ItemType,
			Quantity:         item.Quantity,
			Weight:           item.Weight,
			QRCode:           item.QRCode,
			Status:           item.Status,
			HasIssue:         item.HasIssue,
			IssueDescription: item.IssueDescription,
			Price:            item.Price,
			ReceivedAt:       item.ReceivedAt,
			PackedAt:         item.PackedAt,
			DeliveredAt:      item.DeliveredAt,
			CreatedAt:        item.CreatedAt,
		}
		itemDTOs = append(itemDTOs, itemDTO)
	}

	// Convert pickup to DTO
	var pickupDTO *dto.LaundryPickupDTO
	if pickup != nil {
		pickupDTO = &dto.LaundryPickupDTO{
			ID:          pickup.ID,
			OrderID:     pickup.OrderID,
			ProviderID:  pickup.ProviderID,
			ScheduledAt: pickup.ScheduledAt,
			ArrivedAt:   pickup.ArrivedAt,
			PickedUpAt:  pickup.PickedUpAt,
			Status:      pickup.Status,
			BagCount:    pickup.BagCount,
			Notes:       pickup.Notes,
			PhotoURL:    pickup.PhotoURL,
			CreatedAt:   pickup.CreatedAt,
		}
	}

	// Convert delivery to DTO
	var deliveryDTO *dto.LaundryDeliveryDTO
	if delivery != nil {
		deliveryDTO = &dto.LaundryDeliveryDTO{
			ID:                 delivery.ID,
			OrderID:            delivery.OrderID,
			ProviderID:         delivery.ProviderID,
			ScheduledAt:        delivery.ScheduledAt,
			ArrivedAt:          delivery.ArrivedAt,
			DeliveredAt:        delivery.DeliveredAt,
			Status:             delivery.Status,
			RecipientName:      delivery.RecipientName,
			RecipientSignature: delivery.RecipientSignature,
			Notes:              delivery.Notes,
			PhotoURL:           delivery.PhotoURL,
			RescheduleCount:    delivery.RescheduleCount,
			CreatedAt:          delivery.CreatedAt,
		}
	}

	customerID := ""
	if order.UserID != nil {
		customerID = *order.UserID
	}

	providerID := ""
	if order.ProviderID != nil {
		providerID = *order.ProviderID
	}

	response := &dto.LaundryOrderResponse{
		ID:          order.ID,
		OrderNumber: order.OrderNumber,
		CustomerID:  customerID,
		ProviderID:  providerID,
		ServiceSlug: order.CategorySlug,
		Status:      order.Status,
		TotalPrice:  order.Total,
		Tip:         order.Tip,
		IsExpress:   order.IsExpress,
		Address:     order.Address,
		Lat:         order.Latitude,
		Lng:         order.Longitude,
		Items:       itemDTOs,
		Pickup:      pickupDTO,
		Delivery:    deliveryDTO,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}

	return response, nil
}

func (s *service) GetOrderWithDetails(ctx context.Context, orderID string) (*models.LaundryOrder, error) {
	var order models.LaundryOrder
	if err := s.db.WithContext(ctx).First(&order, "id = ?", orderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("failed to fetch order: %w", err)
	}
	return &order, nil
}

// GetAvailableOrders gets all available laundry orders for a provider to accept
// Available orders are those that match provider's category and are not yet assigned
func (s *service) GetAvailableOrders(ctx context.Context, providerID string) ([]*models.LaundryOrder, error) {
	// Get provider's service category
	provider, err := s.repo.GetProviderByID(ctx, providerID)
	if err != nil {
		logger.Error("GetAvailableOrders: failed to get provider",
			"error", err,
			"providerID", providerID,
		)
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	if provider.ServiceCategory == "" {
		logger.Warn("GetAvailableOrders: provider has no service category",
			"providerID", providerID,
		)
		return nil, errors.New("provider not registered with a service category")
	}

	// Get all active services for this category
	serviceSlugs, err := s.repo.GetProviderServices(ctx, providerID)
	if err != nil {
		logger.Error("GetAvailableOrders: failed to get provider services",
			"error", err,
			"providerID", providerID,
			"category", provider.ServiceCategory,
		)
		return nil, fmt.Errorf("failed to get available services: %w", err)
	}

	if len(serviceSlugs) == 0 {
		logger.Warn("GetAvailableOrders: no active services for provider's category",
			"providerID", providerID,
			"category", provider.ServiceCategory,
		)
		return []*models.LaundryOrder{}, nil
	}

	// Get available orders matching provider's category
	orders, err := s.repo.GetAvailableOrdersByCategory(ctx, provider.ServiceCategory, serviceSlugs)
	if err != nil {
		logger.Error("GetAvailableOrders: failed to get available orders",
			"error", err,
			"providerID", providerID,
			"category", provider.ServiceCategory,
		)
		return nil, fmt.Errorf("failed to get available orders: %w", err)
	}

	logger.Info("GetAvailableOrders: fetched available orders",
		"providerID", providerID,
		"category", provider.ServiceCategory,
		"orderCount", len(orders),
	)

	return orders, nil
}

// =====================================================
// Pickups
// =====================================================

func (s *service) InitiatePickup(ctx context.Context, orderID string, providerID string) (*models.LaundryPickup, error) {
	pickup, err := s.repo.GetPickupByOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pickup: %w", err)
	}
	if pickup == nil {
		return nil, errors.New("pickup not found for this order")
	}

	// If pickup is not yet assigned, assign it to this provider
	if pickup.ProviderID == nil {
		pickup.ProviderID = &providerID
	} else if *pickup.ProviderID != providerID {
		// If already assigned to someone else, deny access
		return nil, errors.New("unauthorized: you are not assigned to this pickup")
	}

	if err := s.repo.UpdatePickupStatus(ctx, orderID, "en_route", nil); err != nil {
		return nil, fmt.Errorf("failed to update pickup status: %w", err)
	}

	pickup.Status = "en_route"
	pickup.UpdatedAt = time.Now()
	return pickup, nil
}

func (s *service) CompletePickup(ctx context.Context, orderID string, req *dto.CompletePickupRequest) error {
	if req == nil {
		return errors.New("request is required")
	}

	now := time.Now()
	if err := s.repo.UpdatePickupStatus(ctx, orderID, "completed", &now); err != nil {
		return fmt.Errorf("failed to complete pickup: %w", err)
	}

	// Update order status to "pickup_completed"
	if err := s.db.WithContext(ctx).
		Model(&models.LaundryOrder{}).
		Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"status":     "pickup_completed",
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

func (s *service) GetProviderPickups(ctx context.Context, providerID string) ([]*models.LaundryPickup, error) {
	// Get pickups that are scheduled or in_route (not yet completed)
	return s.repo.GetPickupsByProvider(ctx, providerID, []string{"scheduled", "en_route", "arrived"})
}

// =====================================================
// Items
// =====================================================

func (s *service) AddItems(ctx context.Context, orderID string, req *dto.AddLaundryItemsRequest) ([]*models.LaundryOrderItem, error) {
	if req == nil || len(req.Items) == 0 {
		return nil, errors.New("at least one item is required")
	}

	// Verify order exists
	_, err := s.GetOrderWithDetails(ctx, orderID)
	if err != nil {
		return nil, err
	}

	items := make([]*models.LaundryOrderItem, len(req.Items))
	now := time.Now()

	for i, itemReq := range req.Items {
		item := &models.LaundryOrderItem{
			ID:          uuid.New().String(),
			OrderID:     orderID,
			ServiceSlug: itemReq.ServiceSlug,
			ProductSlug: itemReq.ProductSlug,
			ItemType:    itemReq.ItemType,
			Quantity:    itemReq.Quantity,
			Weight:      itemReq.Weight,
			Price:       itemReq.Price,
			Status:      "pending",
			HasIssue:    false,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		// QR code will be auto-generated by BeforeCreate hook
		items[i] = item
	}

	if err := s.repo.CreateItems(ctx, items); err != nil {
		return nil, fmt.Errorf("failed to create items: %w", err)
	}

	// Update order status to "processing"
	s.db.WithContext(ctx).
		Model(&models.LaundryOrder{}).
		Where("id = ?", orderID).
		Update("status", "processing")

	return items, nil
}

func (s *service) UpdateItemStatus(ctx context.Context, qrCode, status string) (*models.LaundryOrderItem, error) {
	if qrCode == "" || status == "" {
		return nil, errors.New("qr_code and status are required")
	}

	// Validate status
	validStatuses := map[string]bool{
		"pending":   true,
		"received":  true,
		"washing":   true,
		"drying":    true,
		"pressing":  true,
		"packed":    true,
		"delivered": true,
	}

	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid status: %s (valid: pending, received, washing, drying, pressing, packed, delivered)", status)
	}

	if err := s.repo.UpdateItemStatus(ctx, qrCode, status); err != nil {
		return nil, fmt.Errorf("failed to update item status: %w", err)
	}

	item, err := s.repo.GetItemByQRCode(ctx, qrCode)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated item: %w", err)
	}

	return item, nil
}

func (s *service) GetOrderItems(ctx context.Context, orderID string) ([]*models.LaundryOrderItem, error) {
	return s.repo.GetOrderItems(ctx, orderID)
}

// =====================================================
// Deliveries
// =====================================================

func (s *service) InitiateDelivery(ctx context.Context, orderID string, providerID string) (*models.LaundryDelivery, error) {
	delivery, err := s.repo.GetDeliveryByOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}
	if delivery == nil {
		return nil, errors.New("delivery not found for this order")
	}

	// If delivery is not yet assigned, assign it to this provider
	if delivery.ProviderID == nil {
		delivery.ProviderID = &providerID
	} else if *delivery.ProviderID != providerID {
		// If already assigned to someone else, deny access
		return nil, errors.New("unauthorized: you are not assigned to this delivery")
	}

	if err := s.repo.UpdateDeliveryStatus(ctx, orderID, "en_route", nil); err != nil {
		return nil, fmt.Errorf("failed to update delivery status: %w", err)
	}

	delivery.Status = "en_route"
	delivery.UpdatedAt = time.Now()
	return delivery, nil
}

func (s *service) CompleteDelivery(ctx context.Context, orderID string, req *dto.CompleteDeliveryRequest) error {
	if req == nil {
		return errors.New("request is required")
	}

	now := time.Now()
	if err := s.repo.UpdateDeliveryStatus(ctx, orderID, "completed", &now); err != nil {
		return fmt.Errorf("failed to complete delivery: %w", err)
	}

	// Update all items to "delivered"
	s.db.WithContext(ctx).
		Model(&models.LaundryOrderItem{}).
		Where("order_id = ?", orderID).
		Updates(map[string]interface{}{
			"status":       "delivered",
			"delivered_at": now,
			"updated_at":   now,
		})

	// Update order status to "completed"
	if err := s.db.WithContext(ctx).
		Model(&models.LaundryOrder{}).
		Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"status":     "completed",
			"updated_at": now,
		}).Error; err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

func (s *service) GetProviderDeliveries(ctx context.Context, providerID string) ([]*models.LaundryDelivery, error) {
	// Get deliveries that are scheduled or in_route (not yet completed)
	return s.repo.GetDeliveriesByProvider(ctx, providerID, []string{"scheduled", "en_route", "arrived"})
}

// =====================================================
// Issues
// =====================================================

func (s *service) ReportIssue(ctx context.Context, orderID, userID, providerID string, req *dto.ReportIssueRequest) (*models.LaundryIssue, error) {
	if req == nil {
		return nil, errors.New("request is required")
	}

	// Verify order exists and customer owns it
	order, err := s.GetOrderWithDetails(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != &userID {
		return nil, errors.New("unauthorized: this order does not belong to you")
	}

	// Get provider ID from order if not provided
	if providerID == "" && order.ProviderID != nil {
		providerID = *order.ProviderID
	}

	priority := "medium"
	if req.Priority != "" {
		priority = req.Priority
	}

	issue := &models.LaundryIssue{
		ID:          uuid.New().String(),
		OrderID:     orderID,
		CustomerID:  userID,
		ProviderID:  providerID,
		IssueType:   req.IssueType,
		Description: req.Description,
		Priority:    priority,
		Status:      "open",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateIssue(ctx, issue); err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return issue, nil
}

func (s *service) GetProviderIssues(ctx context.Context, providerID string) ([]*models.LaundryIssue, error) {
	// Get all open issues for provider
	return s.repo.GetIssuesByProvider(ctx, providerID, []string{})
}

func (s *service) ResolveIssue(ctx context.Context, issueID string, resolution string, refundAmount *float64) error {
	if issueID == "" {
		return errors.New("issue_id is required")
	}

	// Update issue status
	now := time.Now()
	updates := map[string]interface{}{
		"status":      "resolved",
		"resolution":  resolution,
		"resolved_at": now,
		"updated_at":  now,
	}

	if refundAmount != nil {
		updates["refund_amount"] = *refundAmount
	}

	if err := s.repo.UpdateIssueStatus(ctx, issueID, "resolved", &resolution, refundAmount); err != nil {
		return fmt.Errorf("failed to resolve issue: %w", err)
	}

	return nil
}

// =====================================================
// Helpers
// =====================================================
