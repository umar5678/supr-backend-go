package dto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
)

// LaundryOrderResponse represents the complete order response
type LaundryOrderResponse struct {
	ID          string                `json:"id"`
	OrderNumber string                `json:"orderNumber"`
	CustomerID  string                `json:"customerId"`
	ProviderID  string                `json:"providerId"`
	ServiceSlug string                `json:"serviceSlug"`
	Status      string                `json:"status"`
	TotalPrice  float64               `json:"totalPrice"`
	Tip         *float64              `json:"tip,omitempty"`
	IsExpress   bool                  `json:"isExpress"`
	Address     string                `json:"address"`
	Lat         float64               `json:"lat"`
	Lng         float64               `json:"lng"`
	Items       []LaundryOrderItemDTO `json:"items"`
	Pickup      *LaundryPickupDTO     `json:"pickup,omitempty"`
	Delivery    *LaundryDeliveryDTO   `json:"delivery,omitempty"`
	CreatedAt   time.Time             `json:"createdAt"`
	UpdatedAt   time.Time             `json:"updatedAt"`
}

// LaundryOrderItemDTO represents an order item
type LaundryOrderItemDTO struct {
	ID               string     `json:"id"`
	OrderID          string     `json:"orderId"`
	ProductSlug      string     `json:"productSlug"`
	ItemType         string     `json:"itemType"`
	Quantity         int        `json:"quantity"`
	Weight           *float64   `json:"weight,omitempty"`
	QRCode           string     `json:"qrCode"`
	Status           string     `json:"status"`
	HasIssue         bool       `json:"hasIssue"`
	IssueDescription *string    `json:"issueDescription,omitempty"`
	Price            float64    `json:"price"`
	ReceivedAt       *time.Time `json:"receivedAt,omitempty"`
	PackedAt         *time.Time `json:"packedAt,omitempty"`
	DeliveredAt      *time.Time `json:"deliveredAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
}

// LaundryPickupDTO represents pickup information
type LaundryPickupDTO struct {
	ID          string     `json:"id"`
	OrderID     string     `json:"orderId"`
	ProviderID  *string    `json:"providerId,omitempty"`
	ScheduledAt time.Time  `json:"scheduledAt"`
	ArrivedAt   *time.Time `json:"arrivedAt,omitempty"`
	PickedUpAt  *time.Time `json:"pickedUpAt,omitempty"`
	Status      string     `json:"status"`
	BagCount    int        `json:"bagCount"`
	Notes       string     `json:"notes"`
	PhotoURL    *string    `json:"photoUrl,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// LaundryDeliveryDTO represents delivery information
type LaundryDeliveryDTO struct {
	ID                 string     `json:"id"`
	OrderID            string     `json:"orderId"`
	ProviderID         *string    `json:"providerId,omitempty"`
	ScheduledAt        time.Time  `json:"scheduledAt"`
	ArrivedAt          *time.Time `json:"arrivedAt,omitempty"`
	DeliveredAt        *time.Time `json:"deliveredAt,omitempty"`
	Status             string     `json:"status"`
	RecipientName      *string    `json:"recipientName,omitempty"`
	RecipientSignature *string    `json:"recipientSignature,omitempty"`
	Notes              string     `json:"notes"`
	PhotoURL           *string    `json:"photoUrl,omitempty"`
	RescheduleCount    int        `json:"rescheduleCount"`
	CreatedAt          time.Time  `json:"createdAt"`
}

// LaundryIssueDTO represents an issue/complaint
type LaundryIssueDTO struct {
	ID               string     `json:"id"`
	OrderID          string     `json:"orderId"`
	CustomerID       string     `json:"customerId"`
	ProviderID       string     `json:"providerId"`
	IssueType        string     `json:"issueType"`
	Description      string     `json:"description"`
	Priority         string     `json:"priority"`
	Status           string     `json:"status"`
	Resolution       *string    `json:"resolution,omitempty"`
	RefundAmount     *float64   `json:"refundAmount,omitempty"`
	CompensationType *string    `json:"compensationType,omitempty"`
	ResolvedAt       *time.Time `json:"resolvedAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// LaundryServiceDTO represents a service in catalog
type LaundryServiceDTO struct {
	ID              string            `json:"id"`
	Slug            string            `json:"slug"`
	Title           string            `json:"title"`
	Description     string            `json:"description"`
	ColorCode       string            `json:"colorCode"`
	BasePrice       float64           `json:"basePrice"`
	PricingUnit     string            `json:"pricingUnit"`
	TurnaroundHours int               `json:"turnaroundHours"`
	ExpressFee      float64           `json:"expressFee"`
	ExpressHours    int               `json:"expressHours"`
	CategorySlug    string            `json:"categorySlug"`
	IsActive        bool              `json:"isActive"`
	ProductCount    int               `json:"productCount"`
	Products        []ProductResponse `json:"products"`
}

// OrderStatusSummary represents order status summary
type OrderStatusSummary struct {
	OrderID           string  `json:"orderId"`
	TotalItems        int     `json:"totalItems"`
	ItemsPending      int     `json:"itemsPending"`
	ItemsReceived     int     `json:"itemsReceived"`
	ItemsWashing      int     `json:"itemsWashing"`
	ItemsDrying       int     `json:"itemsDrying"`
	ItemsPressing     int     `json:"itemsPressing"`
	ItemsPacked       int     `json:"itemsPacked"`
	ItemsDelivered    int     `json:"itemsDelivered"`
	ItemsWithIssues   int     `json:"itemsWithIssues"`
	CompletionPercent float64 `json:"completionPercent"`
}

// PriceEstimateResponse represents price calculation
type PriceEstimateResponse struct {
	ServiceSlug    string               `json:"serviceSlug"`
	Items          []ItemPriceBreakdown `json:"items"`
	SubTotal       float64              `json:"subTotal"`
	ExpressFee     float64              `json:"expressFee"`
	TotalPrice     float64              `json:"totalPrice"`
	TotalWeight    float64              `json:"totalWeight"`
	EstimatedHours int                  `json:"estimatedHours"`
}

type ItemPriceBreakdown struct {
	ProductSlug string  `json:"productSlug"`
	ProductName string  `json:"productName"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	Weight      float64 `json:"weight"`
	ItemTotal   float64 `json:"itemTotal"`
}

// StandardResponse is a generic API response
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// PaginatedResponse represents paginated results
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalItems int         `json:"totalItems"`
	TotalPages int         `json:"totalPages"`
}

// ProviderOrdersResponse represents provider's order list
type ProviderOrdersResponse struct {
	TodayOrders    []LaundryOrderResponse `json:"todayOrders"`
	ActiveOrders   []LaundryOrderResponse `json:"activeOrders"`
	TotalCount     int                    `json:"totalCount"`
	PendingPickup  int                    `json:"pendingPickup"`
	InProgress     int                    `json:"inProgress"`
	ReadyToDeliver int                    `json:"readyToDeliver"`
}

// GetServicesWithProductsResponse - API response for services with products
type GetServicesWithProductsResponse struct {
	ID              string            `json:"id"`
	Slug            string            `json:"slug"`
	Title           string            `json:"title"`
	Description     string            `json:"description"`
	ColorCode       string            `json:"colorCode"`
	BasePrice       float64           `json:"basePrice"`
	PricingUnit     string            `json:"pricingUnit"`
	TurnaroundHours int               `json:"turnaroundHours"`
	ExpressFee      float64           `json:"expressFee"`
	ExpressHours    int               `json:"expressHours"`
	CategorySlug    string            `json:"categorySlug"`
	ProductCount    int               `json:"productCount"`
	Products        []ProductResponse `json:"products"`
}

type ProductResponse struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Slug                string   `json:"slug"`
	Description         string   `json:"description"`
	IconURL             *string  `json:"iconUrl,omitempty"`
	Price               *float64 `json:"price,omitempty"`
	PricingUnit         *string  `json:"pricingUnit,omitempty"`
	TypicalWeight       *float64 `json:"typicalWeight,omitempty"`
	RequiresSpecialCare bool     `json:"requiresSpecialCare"`
	SpecialCareFee      float64  `json:"specialCareFee"`
	CategorySlug        string   `json:"categorySlug"`
}

// LaundryPickupResponse represents a pickup event response
type LaundryPickupResponse struct {
	ID          string     `json:"id"`
	OrderID     string     `json:"orderId"`
	ProviderID  *string    `json:"providerId,omitempty"`
	ScheduledAt time.Time  `json:"scheduledAt"`
	PickedUpAt  *time.Time `json:"pickedUpAt,omitempty"`
	Status      string     `json:"status"`
	BagCount    int        `json:"bagCount,omitempty"`
	Notes       string     `json:"notes,omitempty"`
	PhotoURL    *string    `json:"photoUrl,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// LaundryDeliveryResponse represents a delivery event response
type LaundryDeliveryResponse struct {
	ID                 string     `json:"id"`
	OrderID            string     `json:"orderId"`
	ProviderID         *string    `json:"providerId,omitempty"`
	ScheduledAt        time.Time  `json:"scheduledAt"`
	DeliveredAt        *time.Time `json:"deliveredAt,omitempty"`
	Status             string     `json:"status"`
	RecipientName      string     `json:"recipientName,omitempty"`
	RecipientSignature *string    `json:"recipientSignature,omitempty"`
	Notes              string     `json:"notes,omitempty"`
	PhotoURL           *string    `json:"photoUrl,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// LaundryOrderItemResponse represents an item in a laundry order
type LaundryOrderItemResponse struct {
	ID          string    `json:"id"`
	OrderID     string    `json:"orderId"`
	QRCode      string    `json:"qrCode"`
	ItemType    string    `json:"itemType"`
	Quantity    int       `json:"quantity"`
	ServiceSlug string    `json:"serviceSlug"`
	Weight      *float64  `json:"weight,omitempty"`
	Price       float64   `json:"price"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// LaundryIssueResponse represents an issue reported on an order
type LaundryIssueResponse struct {
	ID           string     `json:"id"`
	OrderID      string     `json:"orderId"`
	IssueType    string     `json:"issueType"`
	Description  string     `json:"description"`
	Priority     string     `json:"priority"`
	Status       string     `json:"status"`
	Resolution   *string    `json:"resolution,omitempty"`
	RefundAmount *float64   `json:"refundAmount,omitempty"`
	ResolvedAt   *time.Time `json:"resolvedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// LaundryServiceResponse represents a service in the catalog
type LaundryServiceResponse struct {
	ID              string  `json:"id"`
	Slug            string  `json:"slug"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	ColorCode       string  `json:"colorCode"`
	BasePrice       float64 `json:"basePrice"`
	PricingUnit     string  `json:"pricingUnit"`
	TurnaroundHours int     `json:"turnaroundHours"`
	ExpressFee      float64 `json:"expressFee"`
	ExpressHours    int     `json:"expressHours"`
	CategorySlug    string  `json:"categorySlug"`
}

// ToLaundryOrderResponse converts a LaundryOrder model to response DTO
func ToLaundryOrderResponse(order *models.LaundryOrder) *LaundryOrderResponse {
	return &LaundryOrderResponse{
		ID:          order.ID,
		OrderNumber: order.OrderNumber,
		Status:      order.Status,
		TotalPrice:  order.Total,
		CreatedAt:   order.CreatedAt,
	}
}

// ToLaundryPickupResponse converts a LaundryPickup model to response DTO
func ToLaundryPickupResponse(pickup *models.LaundryPickup) *LaundryPickupResponse {
	return &LaundryPickupResponse{
		ID:          pickup.ID,
		OrderID:     pickup.OrderID,
		ProviderID:  pickup.ProviderID,
		ScheduledAt: pickup.ScheduledAt,
		PickedUpAt:  pickup.PickedUpAt,
		Status:      pickup.Status,
		BagCount:    pickup.BagCount,
		Notes:       pickup.Notes,
		PhotoURL:    pickup.PhotoURL,
		CreatedAt:   pickup.CreatedAt,
		UpdatedAt:   pickup.UpdatedAt,
	}
}

// ToLaundryDeliveryResponse converts a LaundryDelivery model to response DTO
func ToLaundryDeliveryResponse(delivery *models.LaundryDelivery) *LaundryDeliveryResponse {
	var recipientName string
	if delivery.RecipientName != nil {
		recipientName = *delivery.RecipientName
	}

	return &LaundryDeliveryResponse{
		ID:                 delivery.ID,
		OrderID:            delivery.OrderID,
		ProviderID:         delivery.ProviderID,
		ScheduledAt:        delivery.ScheduledAt,
		DeliveredAt:        delivery.DeliveredAt,
		Status:             delivery.Status,
		RecipientName:      recipientName,
		RecipientSignature: delivery.RecipientSignature,
		Notes:              delivery.Notes,
		PhotoURL:           delivery.PhotoURL,
		CreatedAt:          delivery.CreatedAt,
		UpdatedAt:          delivery.UpdatedAt,
	}
}

// ToLaundryOrderItemResponse converts a LaundryOrderItem model to response DTO
func ToLaundryOrderItemResponse(item *models.LaundryOrderItem) *LaundryOrderItemResponse {
	return &LaundryOrderItemResponse{
		ID:          item.ID,
		OrderID:     item.OrderID,
		QRCode:      item.QRCode,
		ItemType:    item.ItemType,
		Quantity:    item.Quantity,
		ServiceSlug: item.ServiceSlug,
		Weight:      item.Weight,
		Price:       item.Price,
		Status:      item.Status,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

// ToLaundryIssueResponse converts a LaundryIssue model to response DTO
func ToLaundryIssueResponse(issue *models.LaundryIssue) *LaundryIssueResponse {
	return &LaundryIssueResponse{
		ID:           issue.ID,
		OrderID:      issue.OrderID,
		IssueType:    issue.IssueType,
		Description:  issue.Description,
		Priority:     issue.Priority,
		Status:       issue.Status,
		Resolution:   issue.Resolution,
		RefundAmount: issue.RefundAmount,
		ResolvedAt:   issue.ResolvedAt,
		CreatedAt:    issue.CreatedAt,
		UpdatedAt:    issue.UpdatedAt,
	}
}

// ToLaundryServiceResponse converts a LaundryServiceCatalog model to response DTO
func ToLaundryServiceResponse(service *models.LaundryServiceCatalog) *LaundryServiceResponse {
	return &LaundryServiceResponse{
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
	}
}

// ToLaundryServiceResponses converts multiple LaundryServiceCatalog models to response DTOs
func ToLaundryServiceResponses(services []*models.LaundryServiceCatalog) []*LaundryServiceResponse {
	responses := make([]*LaundryServiceResponse, len(services))
	for i, service := range services {
		responses[i] = ToLaundryServiceResponse(service)
	}
	return responses
}

// ToLaundryOrderItemResponses converts multiple LaundryOrderItem models to response DTOs
func ToLaundryOrderItemResponses(items []*models.LaundryOrderItem) []*LaundryOrderItemResponse {
	responses := make([]*LaundryOrderItemResponse, len(items))
	for i, item := range items {
		responses[i] = ToLaundryOrderItemResponse(item)
	}
	return responses
}

// ToLaundryPickupResponses converts multiple LaundryPickup models to response DTOs
func ToLaundryPickupResponses(pickups []*models.LaundryPickup) []*LaundryPickupResponse {
	responses := make([]*LaundryPickupResponse, len(pickups))
	for i, pickup := range pickups {
		responses[i] = ToLaundryPickupResponse(pickup)
	}
	return responses
}

// ToLaundryDeliveryResponses converts multiple LaundryDelivery models to response DTOs
func ToLaundryDeliveryResponses(deliveries []*models.LaundryDelivery) []*LaundryDeliveryResponse {
	responses := make([]*LaundryDeliveryResponse, len(deliveries))
	for i, delivery := range deliveries {
		responses[i] = ToLaundryDeliveryResponse(delivery)
	}
	return responses
}

// ToLaundryIssueResponses converts multiple LaundryIssue models to response DTOs
func ToLaundryIssueResponses(issues []*models.LaundryIssue) []*LaundryIssueResponse {
	responses := make([]*LaundryIssueResponse, len(issues))
	for i, issue := range issues {
		responses[i] = ToLaundryIssueResponse(issue)
	}
	return responses
}
