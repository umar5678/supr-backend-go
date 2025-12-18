package dto

import (
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
)

// ==================== Order Responses ====================

// OrderServiceItem represents a service item in order response
type OrderServiceItem struct {
	ServiceSlug string  `json:"serviceSlug"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

// OrderAddonItem represents an addon item in order response
type OrderAddonItem struct {
	AddonSlug string  `json:"addonSlug"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
}

// OrderCustomerInfo represents customer info in order response
type OrderCustomerInfo struct {
	Name    string  `json:"name"`
	Phone   string  `json:"phone"`
	Email   string  `json:"email,omitempty"`
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

// OrderBookingInfo represents booking info in order response
type OrderBookingInfo struct {
	Day            string `json:"day"`
	Date           string `json:"date"`
	Time           string `json:"time"`
	PreferredTime  string `json:"preferredTime,omitempty"`
	FormattedDate  string `json:"formattedDate"` // e.g., "Monday, January 15, 2024"
	FormattedTime  string `json:"formattedTime"` // e.g., "2:30 PM"
	QuantityOfPros int    `json:"quantityOfPros"`
}

// OrderPricing represents pricing breakdown in order response
type OrderPricing struct {
	ServicesTotal      float64 `json:"servicesTotal"`
	AddonsTotal        float64 `json:"addonsTotal"`
	Subtotal           float64 `json:"subtotal"`
	PlatformCommission float64 `json:"platformCommission"`
	TotalPrice         float64 `json:"totalPrice"`
	FormattedTotal     string  `json:"formattedTotal"`
}

// OrderPaymentInfo represents payment info in order response
type OrderPaymentInfo struct {
	Method        string  `json:"method"`
	Status        string  `json:"status"`
	AmountPaid    float64 `json:"amountPaid"`
	TransactionID string  `json:"transactionId,omitempty"`
}

// OrderProviderInfo represents assigned provider info (minimal for customer)
type OrderProviderInfo struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Phone        string  `json:"phone,omitempty"` // Only shown after acceptance
	Rating       float64 `json:"rating"`
	TotalReviews int     `json:"totalReviews"`
	Photo        string  `json:"photo,omitempty"`
}

// OrderCancellationInfo represents cancellation details
type OrderCancellationInfo struct {
	CancelledBy     string    `json:"cancelledBy"`
	CancelledAt     time.Time `json:"cancelledAt"`
	Reason          string    `json:"reason"`
	CancellationFee float64   `json:"cancellationFee"`
	RefundAmount    float64   `json:"refundAmount"`
}

// OrderRatingInfo represents rating information
type OrderRatingInfo struct {
	CustomerRating  *int       `json:"customerRating,omitempty"`
	CustomerReview  string     `json:"customerReview,omitempty"`
	CustomerRatedAt *time.Time `json:"customerRatedAt,omitempty"`
	ProviderRating  *int       `json:"providerRating,omitempty"`
	ProviderReview  string     `json:"providerReview,omitempty"`
	ProviderRatedAt *time.Time `json:"providerRatedAt,omitempty"`
	CanRate         bool       `json:"canRate"`
}

// OrderStatusInfo represents order status with timestamps
type OrderStatusInfo struct {
	Current            string     `json:"current"`
	DisplayStatus      string     `json:"displayStatus"`
	ProviderAcceptedAt *time.Time `json:"providerAcceptedAt,omitempty"`
	ProviderStartedAt  *time.Time `json:"providerStartedAt,omitempty"`
	CompletedAt        *time.Time `json:"completedAt,omitempty"`
	CanCancel          bool       `json:"canCancel"`
}

// OrderResponse represents a full order response
type OrderResponse struct {
	ID            string                 `json:"id"`
	OrderNumber   string                 `json:"orderNumber"`
	CustomerInfo  OrderCustomerInfo      `json:"customerInfo"`
	BookingInfo   OrderBookingInfo       `json:"bookingInfo"`
	CategorySlug  string                 `json:"categorySlug"`
	CategoryTitle string                 `json:"categoryTitle"`
	Services      []OrderServiceItem     `json:"services"`
	Addons        []OrderAddonItem       `json:"addons,omitempty"`
	SpecialNotes  string                 `json:"specialNotes,omitempty"`
	Pricing       OrderPricing           `json:"pricing"`
	Payment       *OrderPaymentInfo      `json:"payment,omitempty"`
	Provider      *OrderProviderInfo     `json:"provider,omitempty"`
	Status        OrderStatusInfo        `json:"status"`
	Cancellation  *OrderCancellationInfo `json:"cancellation,omitempty"`
	Rating        *OrderRatingInfo       `json:"rating,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
}

// OrderListResponse represents an order in list view (lighter)
type OrderListResponse struct {
	ID             string           `json:"id"`
	OrderNumber    string           `json:"orderNumber"`
	CategorySlug   string           `json:"categorySlug"`
	CategoryTitle  string           `json:"categoryTitle"`
	BookingInfo    OrderBookingInfo `json:"bookingInfo"`
	TotalPrice     float64          `json:"totalPrice"`
	FormattedTotal string           `json:"formattedTotal"`
	Status         string           `json:"status"`
	DisplayStatus  string           `json:"displayStatus"`
	ServiceCount   int              `json:"serviceCount"`
	CanCancel      bool             `json:"canCancel"`
	CanRate        bool             `json:"canRate"`
	CreatedAt      time.Time        `json:"createdAt"`
}

// OrderCreatedResponse represents the response after creating an order
type OrderCreatedResponse struct {
	ID                      string           `json:"id"`
	OrderNumber             string           `json:"orderNumber"`
	Status                  string           `json:"status"`
	DisplayStatus           string           `json:"displayStatus"`
	BookingInfo             OrderBookingInfo `json:"bookingInfo"`
	TotalPrice              float64          `json:"totalPrice"`
	FormattedTotal          string           `json:"formattedTotal"`
	EstimatedAssignmentTime string           `json:"estimatedAssignmentTime"`
	Message                 string           `json:"message"`
}

// CancellationPreviewResponse represents cancellation fee preview
type CancellationPreviewResponse struct {
	OrderID         string  `json:"orderId"`
	OrderNumber     string  `json:"orderNumber"`
	CurrentStatus   string  `json:"currentStatus"`
	TotalPrice      float64 `json:"totalPrice"`
	CancellationFee float64 `json:"cancellationFee"`
	RefundAmount    float64 `json:"refundAmount"`
	FeePercentage   float64 `json:"feePercentage"`
	Message         string  `json:"message"`
}

// ==================== Conversion Functions ====================

// GetDisplayStatus returns human-readable status
func GetDisplayStatus(status string) string {
	statusMap := map[string]string{
		"pending":            "Pending",
		"searching_provider": "Finding Provider",
		"assigned":           "Provider Assigned",
		"accepted":           "Provider Confirmed",
		"in_progress":        "In Progress",
		"completed":          "Completed",
		"cancelled":          "Cancelled",
		"expired":            "Expired",
	}
	if display, ok := statusMap[status]; ok {
		return display
	}
	return status
}

// FormatOrderDate formats date for display
func FormatOrderDate(dateStr string) string {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return date.Format("Monday, January 2, 2006")
}

// FormatOrderTime formats time for display
func FormatOrderTime(timeStr string) string {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("3:04 PM")
}

// ToOrderCustomerInfo converts model to response
func ToOrderCustomerInfo(info models.CustomerInfo) OrderCustomerInfo {
	return OrderCustomerInfo{
		Name:    info.Name,
		Phone:   info.Phone,
		Email:   info.Email,
		Address: info.Address,
		Lat:     info.Lat,
		Lng:     info.Lng,
	}
}

// ToOrderBookingInfo converts model to response
func ToOrderBookingInfo(info models.BookingInfo) OrderBookingInfo {
	// Use fmt.Sprint to safely convert underlying types (string, time.Time, *string, etc.) to string
	dateStr := fmt.Sprint(info.Date)
	timeStr := fmt.Sprint(info.Time)
	preferredStr := fmt.Sprint(info.PreferredTime)

	return OrderBookingInfo{
		Day:            fmt.Sprint(info.Day),
		Date:           dateStr,
		Time:           timeStr,
		PreferredTime:  preferredStr,
		FormattedDate:  FormatOrderDate(dateStr),
		FormattedTime:  FormatOrderTime(timeStr),
		QuantityOfPros: info.QuantityOfPros,
	}
}

// ToOrderServiceItems converts model to response
func ToOrderServiceItems(services models.SelectedServices) []OrderServiceItem {
	items := make([]OrderServiceItem, len(services))
	for i, s := range services {
		items[i] = OrderServiceItem{
			ServiceSlug: s.ServiceSlug,
			Title:       s.Title,
			Price:       s.Price,
			Quantity:    s.Quantity,
			Subtotal:    s.Price * float64(s.Quantity),
		}
	}
	return items
}

// ToOrderAddonItems converts model to response
func ToOrderAddonItems(addons models.SelectedAddons) []OrderAddonItem {
	if addons == nil {
		return nil
	}
	items := make([]OrderAddonItem, len(addons))
	for i, a := range addons {
		items[i] = OrderAddonItem{
			AddonSlug: a.AddonSlug,
			Title:     a.Title,
			Price:     a.Price,
			Quantity:  a.Quantity,
			Subtotal:  a.Price * float64(a.Quantity),
		}
	}
	return items
}

// ToOrderPricing converts order pricing to response
func ToOrderPricing(order *models.ServiceOrderNew) OrderPricing {
	return OrderPricing{
		ServicesTotal:      order.ServicesTotal,
		AddonsTotal:        order.AddonsTotal,
		Subtotal:           order.Subtotal,
		PlatformCommission: order.PlatformCommission,
		TotalPrice:         order.TotalPrice,
		FormattedTotal:     FormatPriceValue(order.TotalPrice),
	}
}

// ToOrderPaymentInfo converts payment info to response
func ToOrderPaymentInfo(info *models.PaymentInfo) *OrderPaymentInfo {
	if info == nil {
		return nil
	}
	return &OrderPaymentInfo{
		Method:        info.Method,
		Status:        info.Status,
		AmountPaid:    info.AmountPaid,
		TransactionID: info.TransactionID,
	}
}

// ToOrderCancellationInfo converts cancellation info to response
func ToOrderCancellationInfo(info *models.CancellationInfo) *OrderCancellationInfo {
	if info == nil {
		return nil
	}
	return &OrderCancellationInfo{
		CancelledBy:     info.CancelledBy,
		CancelledAt:     info.CancelledAt,
		Reason:          info.Reason,
		CancellationFee: info.CancellationFee,
		RefundAmount:    info.RefundAmount,
	}
}

// ToOrderStatusInfo converts order status to response
func ToOrderStatusInfo(order *models.ServiceOrderNew) OrderStatusInfo {
	return OrderStatusInfo{
		Current:            order.Status,
		DisplayStatus:      GetDisplayStatus(order.Status),
		ProviderAcceptedAt: order.ProviderAcceptedAt,
		ProviderStartedAt:  order.ProviderStartedAt,
		CompletedAt:        order.CompletedAt,
		CanCancel:          order.CanBeCancelled(),
	}
}

// ToOrderRatingInfo converts rating info to response
func ToOrderRatingInfo(order *models.ServiceOrderNew) *OrderRatingInfo {
	if order.Status != "completed" {
		return nil
	}
	return &OrderRatingInfo{
		CustomerRating:  order.CustomerRating,
		CustomerReview:  order.CustomerReview,
		CustomerRatedAt: order.CustomerRatedAt,
		ProviderRating:  order.ProviderRating,
		ProviderReview:  order.ProviderReview,
		ProviderRatedAt: order.ProviderRatedAt,
		CanRate:         order.CanBeRatedByCustomer(),
	}
}

// GetCategoryTitle returns category title from slug
func GetCategoryTitle(slug string) string {
	configs := map[string]string{
		"pest-control": "Pest Control",
		"cleaning":     "Cleaning Services",
		"iv-therapy":   "IV Therapy",
		"massage":      "Massage Therapy",
		"handyman":     "Handyman Services",
	}
	if title, ok := configs[slug]; ok {
		return title
	}
	return slug
}

// ToOrderResponse converts model to full response
func ToOrderResponse(order *models.ServiceOrderNew) *OrderResponse {
	response := &OrderResponse{
		ID:            order.ID,
		OrderNumber:   order.OrderNumber,
		CustomerInfo:  ToOrderCustomerInfo(order.CustomerInfo),
		BookingInfo:   ToOrderBookingInfo(order.BookingInfo),
		CategorySlug:  order.CategorySlug,
		CategoryTitle: GetCategoryTitle(order.CategorySlug),
		Services:      ToOrderServiceItems(order.SelectedServices),
		Addons:        ToOrderAddonItems(order.SelectedAddons),
		SpecialNotes:  order.SpecialNotes,
		Pricing:       ToOrderPricing(order),
		Payment:       ToOrderPaymentInfo(order.PaymentInfo),
		Status:        ToOrderStatusInfo(order),
		Cancellation:  ToOrderCancellationInfo(order.CancellationInfo),
		Rating:        ToOrderRatingInfo(order),
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}

	// Add provider info if assigned (placeholder - would load from user table)
	if order.AssignedProviderID != nil {
		// In real implementation, load provider details
		response.Provider = &OrderProviderInfo{
			ID: *order.AssignedProviderID,
			// Name, Phone, Rating would be loaded from provider profile
		}
	}

	return response
}

// ToOrderListResponse converts model to list response
func ToOrderListResponse(order *models.ServiceOrderNew) OrderListResponse {
	return OrderListResponse{
		ID:             order.ID,
		OrderNumber:    order.OrderNumber,
		CategorySlug:   order.CategorySlug,
		CategoryTitle:  GetCategoryTitle(order.CategorySlug),
		BookingInfo:    ToOrderBookingInfo(order.BookingInfo),
		TotalPrice:     order.TotalPrice,
		FormattedTotal: FormatPriceValue(order.TotalPrice),
		Status:         order.Status,
		DisplayStatus:  GetDisplayStatus(order.Status),
		ServiceCount:   len(order.SelectedServices),
		CanCancel:      order.CanBeCancelled(),
		CanRate:        order.CanBeRatedByCustomer(),
		CreatedAt:      order.CreatedAt,
	}
}

// ToOrderListResponses converts multiple models to list responses
func ToOrderListResponses(orders []*models.ServiceOrderNew) []OrderListResponse {
	if orders == nil {
		return []OrderListResponse{}
	}
	responses := make([]OrderListResponse, len(orders))
	for i, order := range orders {
		responses[i] = ToOrderListResponse(order)
	}
	return responses
}

// ToOrderCreatedResponse creates response for newly created order
func ToOrderCreatedResponse(order *models.ServiceOrderNew) *OrderCreatedResponse {
	return &OrderCreatedResponse{
		ID:                      order.ID,
		OrderNumber:             order.OrderNumber,
		Status:                  order.Status,
		DisplayStatus:           GetDisplayStatus(order.Status),
		BookingInfo:             ToOrderBookingInfo(order.BookingInfo),
		TotalPrice:              order.TotalPrice,
		FormattedTotal:          FormatPriceValue(order.TotalPrice),
		EstimatedAssignmentTime: "5-15 minutes",
		Message:                 "Your booking has been created. We're finding the best provider for you.",
	}
}
