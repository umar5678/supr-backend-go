package dto

import (
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
)

// ==================== Category Responses ====================

type CategoryResponse struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	Image        string `json:"image"`
	ServiceCount int    `json:"serviceCount"`
	AddonCount   int    `json:"addonCount"`
}

type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Total      int                `json:"total"`
}

type CategoryDetailResponse struct {
	Slug        string            `json:"slug"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Icon        string            `json:"icon"`
	Image       string            `json:"image"`
	Services    []ServiceResponse `json:"services"`
	Addons      []AddonResponse   `json:"addons"`
}

// ==================== Service Responses ====================

type ServiceResponse struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	LongTitle          string   `json:"longTitle,omitempty"`
	Slug               string   `json:"slug"`
	CategorySlug       string   `json:"categorySlug"`
	Description        string   `json:"description"`
	LongDescription    string   `json:"longDescription,omitempty"`
	Highlights         string   `json:"highlights"`
	WhatsIncluded      []string `json:"whatsIncluded"`
	TermsAndConditions []string `json:"termsAndConditions,omitempty"`
	BannerImage        string   `json:"bannerImage,omitempty"`
	Thumbnail          string   `json:"thumbnail"`
	Duration           *int     `json:"duration,omitempty"` // in minutes
	DurationText       string   `json:"durationText,omitempty"`
	IsFrequent         bool     `json:"isFrequent"`
	Frequency          string   `json:"frequency,omitempty"`
	BasePrice          *float64 `json:"basePrice,omitempty"`
	FormattedPrice     string   `json:"formattedPrice,omitempty"`
}

type ServiceListResponse struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Slug           string   `json:"slug"`
	CategorySlug   string   `json:"categorySlug"`
	Description    string   `json:"description"`
	Thumbnail      string   `json:"thumbnail"`
	Duration       *int     `json:"duration,omitempty"`
	DurationText   string   `json:"durationText,omitempty"`
	IsFrequent     bool     `json:"isFrequent"`
	BasePrice      *float64 `json:"basePrice,omitempty"`
	FormattedPrice string   `json:"formattedPrice,omitempty"`
}

type ServiceDetailResponse struct {
	Service ServiceResponse `json:"service"`
	Addons  []AddonResponse `json:"addons"`
}

// ==================== Addon Responses ====================

type AddonResponse struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Slug               string   `json:"slug"`
	CategorySlug       string   `json:"categorySlug"`
	Description        string   `json:"description"`
	WhatsIncluded      []string `json:"whatsIncluded"`
	Notes              []string `json:"notes,omitempty"`
	Image              string   `json:"image"`
	Price              float64  `json:"price"`
	FormattedPrice     string   `json:"formattedPrice"`
	StrikethroughPrice *float64 `json:"strikethroughPrice,omitempty"`
	DiscountPercentage float64  `json:"discountPercentage,omitempty"`
	HasDiscount        bool     `json:"hasDiscount"`
}

type AddonListResponse struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Slug               string   `json:"slug"`
	CategorySlug       string   `json:"categorySlug"`
	Image              string   `json:"image"`
	Price              float64  `json:"price"`
	FormattedPrice     string   `json:"formattedPrice"`
	StrikethroughPrice *float64 `json:"strikethroughPrice,omitempty"`
	DiscountPercentage float64  `json:"discountPercentage,omitempty"`
	HasDiscount        bool     `json:"hasDiscount"`
}

// ==================== Search Response ====================

type SearchResultItem struct {
	Type           string   `json:"type"` // "service" or "addon"
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Slug           string   `json:"slug"`
	CategorySlug   string   `json:"categorySlug"`
	Description    string   `json:"description"`
	Image          string   `json:"image"`
	Price          *float64 `json:"price,omitempty"`
	FormattedPrice string   `json:"formattedPrice,omitempty"`
}

type SearchResponse struct {
	Query   string             `json:"query"`
	Results []SearchResultItem `json:"results"`
	Total   int                `json:"total"`
}

// ==================== Conversion Functions ====================

func FormatDuration(minutes *int) string {
	if minutes == nil {
		return ""
	}
	m := *minutes
	if m < 60 {
		return fmt.Sprintf("%d min", m)
	}
	hours := m / 60
	mins := m % 60
	if mins == 0 {
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	if hours == 1 {
		return fmt.Sprintf("1 hour %d min", mins)
	}
	return fmt.Sprintf("%d hours %d min", hours, mins)
}

func FormatPrice(price *float64) string {
	if price == nil {
		return ""
	}
	return fmt.Sprintf("$%.2f", *price)
}

func FormatPriceValue(price float64) string {
	return fmt.Sprintf("$%.2f", price)
}

func ToServiceResponse(service *models.ServiceNew) ServiceResponse {

	whatsIncluded := make([]string, len(service.WhatsIncluded))
	copy(whatsIncluded, service.WhatsIncluded)

	termsAndConditions := make([]string, len(service.TermsAndConditions))
	copy(termsAndConditions, service.TermsAndConditions)

	return ServiceResponse{
		ID:                 service.ID,
		Title:              service.Title,
		LongTitle:          service.LongTitle,
		Slug:               service.ServiceSlug,
		CategorySlug:       service.CategorySlug,
		Description:        service.Description,
		LongDescription:    service.LongDescription,
		Highlights:         service.Highlights,
		WhatsIncluded:      whatsIncluded,
		TermsAndConditions: termsAndConditions,
		BannerImage:        service.BannerImage,
		Thumbnail:          service.Thumbnail,
		Duration:           service.Duration,
		DurationText:       FormatDuration(service.Duration),
		IsFrequent:         service.IsFrequent,
		Frequency:          service.Frequency,
		BasePrice:          service.BasePrice,
		FormattedPrice:     FormatPrice(service.BasePrice),
	}
}

func ToServiceListResponse(service *models.ServiceNew) ServiceListResponse {
	return ServiceListResponse{
		ID:             service.ID,
		Title:          service.Title,
		Slug:           service.ServiceSlug,
		CategorySlug:   service.CategorySlug,
		Description:    service.Description,
		Thumbnail:      service.Thumbnail,
		Duration:       service.Duration,
		DurationText:   FormatDuration(service.Duration),
		IsFrequent:     service.IsFrequent,
		BasePrice:      service.BasePrice,
		FormattedPrice: FormatPrice(service.BasePrice),
	}
}

func ToServiceListResponses(services []*models.ServiceNew) []ServiceListResponse {
	if services == nil {
		return []ServiceListResponse{}
	}
	responses := make([]ServiceListResponse, len(services))
	for i, service := range services {
		responses[i] = ToServiceListResponse(service)
	}
	return responses
}

func ToServiceResponses(services []*models.ServiceNew) []ServiceResponse {
	if services == nil {
		return []ServiceResponse{}
	}
	responses := make([]ServiceResponse, len(services))
	for i, service := range services {
		responses[i] = ToServiceResponse(service)
	}
	return responses
}

func ToAddonResponse(addon *models.Addon) AddonResponse {
	whatsIncluded := make([]string, len(addon.WhatsIncluded))
	copy(whatsIncluded, addon.WhatsIncluded)

	notes := make([]string, len(addon.Notes))
	copy(notes, addon.Notes)

	return AddonResponse{
		ID:                 addon.ID,
		Title:              addon.Title,
		Slug:               addon.AddonSlug,
		CategorySlug:       addon.CategorySlug,
		Description:        addon.Description,
		WhatsIncluded:      whatsIncluded,
		Notes:              notes,
		Image:              addon.Image,
		Price:              addon.Price,
		FormattedPrice:     FormatPriceValue(addon.Price),
		StrikethroughPrice: addon.StrikethroughPrice,
		DiscountPercentage: addon.DiscountPercentage(),
		HasDiscount:        addon.HasDiscount(),
	}
}

func ToAddonListResponse(addon *models.Addon) AddonListResponse {
	return AddonListResponse{
		ID:                 addon.ID,
		Title:              addon.Title,
		Slug:               addon.AddonSlug,
		CategorySlug:       addon.CategorySlug,
		Image:              addon.Image,
		Price:              addon.Price,
		FormattedPrice:     FormatPriceValue(addon.Price),
		StrikethroughPrice: addon.StrikethroughPrice,
		DiscountPercentage: addon.DiscountPercentage(),
		HasDiscount:        addon.HasDiscount(),
	}
}

func ToAddonListResponses(addons []*models.Addon) []AddonListResponse {
	if addons == nil {
		return []AddonListResponse{}
	}
	responses := make([]AddonListResponse, len(addons))
	for i, addon := range addons {
		responses[i] = ToAddonListResponse(addon)
	}
	return responses
}

func ToAddonResponses(addons []*models.Addon) []AddonResponse {
	if addons == nil {
		return []AddonResponse{}
	}
	responses := make([]AddonResponse, len(addons))
	for i, addon := range addons {
		responses[i] = ToAddonResponse(addon)
	}
	return responses
}

func ToSearchResultFromService(service *models.ServiceNew) SearchResultItem {
	return SearchResultItem{
		Type:           "service",
		ID:             service.ID,
		Title:          service.Title,
		Slug:           service.ServiceSlug,
		CategorySlug:   service.CategorySlug,
		Description:    service.Description,
		Image:          service.Thumbnail,
		Price:          service.BasePrice,
		FormattedPrice: FormatPrice(service.BasePrice),
	}
}

func ToSearchResultFromAddon(addon *models.Addon) SearchResultItem {
	price := addon.Price
	return SearchResultItem{
		Type:           "addon",
		ID:             addon.ID,
		Title:          addon.Title,
		Slug:           addon.AddonSlug,
		CategorySlug:   addon.CategorySlug,
		Description:    addon.Description,
		Image:          addon.Image,
		Price:          &price,
		FormattedPrice: FormatPriceValue(addon.Price),
	}
}

// ==================== Order Responses ====================

type OrderServiceItem struct {
	ServiceSlug string  `json:"serviceSlug"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

type OrderAddonItem struct {
	AddonSlug string  `json:"addonSlug"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
}

type OrderCustomerInfo struct {
	Name    string  `json:"name"`
	Phone   string  `json:"phone"`
	Email   string  `json:"email,omitempty"`
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

type OrderBookingInfo struct {
	Day            string `json:"day"`
	Date           string `json:"date"`
	Time           string `json:"time"`
	PreferredTime  string `json:"preferredTime,omitempty"`
	FormattedDate  string `json:"formattedDate"` // e.g., "Monday, January 15, 2024"
	FormattedTime  string `json:"formattedTime"` // e.g., "2:30 PM"
	QuantityOfPros int    `json:"quantityOfPros"`
}

type OrderPricing struct {
	ServicesTotal      float64 `json:"servicesTotal"`
	AddonsTotal        float64 `json:"addonsTotal"`
	Subtotal           float64 `json:"subtotal"`
	PlatformCommission float64 `json:"platformCommission"`
	TotalPrice         float64 `json:"totalPrice"`
	FormattedTotal     string  `json:"formattedTotal"`
}

type OrderPaymentInfo struct {
	Method        string  `json:"method"`
	Status        string  `json:"status"`
	AmountPaid    float64 `json:"amountPaid"`
	TransactionID string  `json:"transactionId,omitempty"`
}

type OrderProviderInfo struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Phone        string  `json:"phone,omitempty"` // Only shown after acceptance
	Rating       float64 `json:"rating"`
	TotalReviews int     `json:"totalReviews"`
	Photo        string  `json:"photo,omitempty"`
}

type OrderCancellationInfo struct {
	CancelledBy     string    `json:"cancelledBy"`
	CancelledAt     time.Time `json:"cancelledAt"`
	Reason          string    `json:"reason"`
	CancellationFee float64   `json:"cancellationFee"`
	RefundAmount    float64   `json:"refundAmount"`
}

type OrderRatingInfo struct {
	CustomerRating  *int       `json:"customerRating,omitempty"`
	CustomerReview  string     `json:"customerReview,omitempty"`
	CustomerRatedAt *time.Time `json:"customerRatedAt,omitempty"`
	ProviderRating  *int       `json:"providerRating,omitempty"`
	ProviderReview  string     `json:"providerReview,omitempty"`
	ProviderRatedAt *time.Time `json:"providerRatedAt,omitempty"`
	CanRate         bool       `json:"canRate"`
}

type OrderStatusInfo struct {
	Current            string     `json:"current"`
	DisplayStatus      string     `json:"displayStatus"`
	ProviderAcceptedAt *time.Time `json:"providerAcceptedAt,omitempty"`
	ProviderStartedAt  *time.Time `json:"providerStartedAt,omitempty"`
	CompletedAt        *time.Time `json:"completedAt,omitempty"`
	CanCancel          bool       `json:"canCancel"`
}

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

func FormatOrderDate(dateStr string) string {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return date.Format("Monday, January 2, 2006")
}

func FormatOrderTime(timeStr string) string {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("3:04 PM")
}

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

func ToOrderBookingInfo(info models.BookingInfo) OrderBookingInfo {
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
		response.Provider = &OrderProviderInfo{
			ID: *order.AssignedProviderID,
		}
	}

	return response
}

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
