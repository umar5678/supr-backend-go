package dto

import (
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
)

type AddonResponse struct {
	ID                 string    `json:"id"`
	Title              string    `json:"title"`
	AddonSlug          string    `json:"addonSlug"`
	CategorySlug       string    `json:"categorySlug"`
	Description        string    `json:"description"`
	WhatsIncluded      []string  `json:"whatsIncluded"`
	Notes              []string  `json:"notes"`
	Image              string    `json:"image"`
	Price              float64   `json:"price"`
	StrikethroughPrice *float64  `json:"strikethroughPrice"`
	DiscountPercentage float64   `json:"discountPercentage"`
	IsActive           bool      `json:"isActive"`
	IsAvailable        bool      `json:"isAvailable"`
	SortOrder          int       `json:"sortOrder"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type AddonListResponse struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	AddonSlug          string   `json:"addonSlug"`
	CategorySlug       string   `json:"categorySlug"`
	Image              string   `json:"image"`
	Price              float64  `json:"price"`
	StrikethroughPrice *float64 `json:"strikethroughPrice"`
	DiscountPercentage float64  `json:"discountPercentage"`
	IsActive           bool     `json:"isActive"`
	IsAvailable        bool     `json:"isAvailable"`
	SortOrder          int      `json:"sortOrder"`
}

func ToAddonResponse(addon *models.Addon) *AddonResponse {
	if addon == nil {
		return nil
	}

	// Convert pq.StringArray to []string
	whatsIncluded := make([]string, len(addon.WhatsIncluded))
	copy(whatsIncluded, addon.WhatsIncluded)

	notes := make([]string, len(addon.Notes))
	copy(notes, addon.Notes)

	return &AddonResponse{
		ID:                 addon.ID,
		Title:              addon.Title,
		AddonSlug:          addon.AddonSlug,
		CategorySlug:       addon.CategorySlug,
		Description:        addon.Description,
		WhatsIncluded:      whatsIncluded,
		Notes:              notes,
		Image:              addon.Image,
		Price:              addon.Price,
		StrikethroughPrice: addon.StrikethroughPrice,
		DiscountPercentage: addon.DiscountPercentage(),
		IsActive:           addon.IsActive,
		IsAvailable:        addon.IsAvailable,
		SortOrder:          addon.SortOrder,
		CreatedAt:          addon.CreatedAt,
		UpdatedAt:          addon.UpdatedAt,
	}
}

func ToAddonListResponse(addon *models.Addon) *AddonListResponse {
	if addon == nil {
		return nil
	}

	return &AddonListResponse{
		ID:                 addon.ID,
		Title:              addon.Title,
		AddonSlug:          addon.AddonSlug,
		CategorySlug:       addon.CategorySlug,
		Image:              addon.Image,
		Price:              addon.Price,
		StrikethroughPrice: addon.StrikethroughPrice,
		DiscountPercentage: addon.DiscountPercentage(),
		IsActive:           addon.IsActive,
		IsAvailable:        addon.IsAvailable,
		SortOrder:          addon.SortOrder,
	}
}

func ToAddonListResponses(addons []*models.Addon) []*AddonListResponse {
	if addons == nil {
		return []*AddonListResponse{}
	}

	responses := make([]*AddonListResponse, len(addons))
	for i, addon := range addons {
		responses[i] = ToAddonListResponse(addon)
	}
	return responses
}

// ==================== Order Responses ====================

type AdminOrderListResponse struct {
	ID             string     `json:"id"`
	OrderNumber    string     `json:"orderNumber"`
	CategorySlug   string     `json:"categorySlug"`
	CategoryTitle  string     `json:"categoryTitle"`
	CustomerName   string     `json:"customerName"`
	CustomerPhone  string     `json:"customerPhone"`
	ProviderName   string     `json:"providerName,omitempty"`
	BookingDate    string     `json:"bookingDate"`
	BookingTime    string     `json:"bookingTime"`
	TotalPrice     float64    `json:"totalPrice"`
	ProviderPayout float64    `json:"providerPayout"`
	Commission     float64    `json:"commission"`
	Status         string     `json:"status"`
	DisplayStatus  string     `json:"displayStatus"`
	PaymentMethod  string     `json:"paymentMethod"`
	PaymentStatus  string     `json:"paymentStatus"`
	CreatedAt      time.Time  `json:"createdAt"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
}

type AdminOrderDetailResponse struct {
	ID          string `json:"id"`
	OrderNumber string `json:"orderNumber"`

	// Category
	CategorySlug  string `json:"categorySlug"`
	CategoryTitle string `json:"categoryTitle"`

	// Customer
	Customer OrderPartyInfo `json:"customer"`

	// Provider
	Provider *OrderPartyInfo `json:"provider,omitempty"`

	// Booking
	BookingInfo AdminBookingInfo `json:"bookingInfo"`

	// Services
	Services     []AdminOrderServiceItem `json:"services"`
	Addons       []AdminOrderAddonItem   `json:"addons,omitempty"`
	SpecialNotes string                  `json:"specialNotes,omitempty"`

	// Pricing
	Pricing AdminOrderPricing `json:"pricing"`

	// Payment
	Payment AdminPaymentInfo `json:"payment"`

	// Status
	Status AdminOrderStatus `json:"status"`

	// Cancellation (if cancelled)
	Cancellation *AdminCancellationInfo `json:"cancellation,omitempty"`

	// Ratings
	Ratings *AdminRatingsInfo `json:"ratings,omitempty"`

	// History
	StatusHistory []StatusHistoryItem `json:"statusHistory"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Admin actions
	AvailableActions []string `json:"availableActions"`
}

type OrderPartyInfo struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Phone   string  `json:"phone"`
	Address string  `json:"address,omitempty"`
	Lat     float64 `json:"lat,omitempty"`
	Lng     float64 `json:"lng,omitempty"`
	Rating  float64 `json:"rating,omitempty"`
	Photo   string  `json:"photo,omitempty"`
}

type AdminBookingInfo struct {
	Day           string `json:"day"`
	Date          string `json:"date"`
	Time          string `json:"time"`
	PreferredTime string `json:"preferredTime,omitempty"`
	FormattedDate string `json:"formattedDate"`
	FormattedTime string `json:"formattedTime"`
	IsPast        bool   `json:"isPast"`
	IsToday       bool   `json:"isToday"`
}

type AdminOrderServiceItem struct {
	ServiceSlug string  `json:"serviceSlug"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

type AdminOrderAddonItem struct {
	AddonSlug string  `json:"addonSlug"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
}

type AdminOrderPricing struct {
	ServicesTotal      float64 `json:"servicesTotal"`
	AddonsTotal        float64 `json:"addonsTotal"`
	Subtotal           float64 `json:"subtotal"`
	PlatformCommission float64 `json:"platformCommission"`
	CommissionRate     float64 `json:"commissionRate"` // e.g., 0.10 for 10%
	TotalPrice         float64 `json:"totalPrice"`
	ProviderPayout     float64 `json:"providerPayout"`
	FormattedTotal     string  `json:"formattedTotal"`
}

type AdminPaymentInfo struct {
	Method        string  `json:"method"`
	Status        string  `json:"status"`
	Total         float64 `json:"total"`
	AmountPaid    float64 `json:"amountPaid"`
	Voucher       string  `json:"voucher,omitempty"`
	TransactionID string  `json:"transactionId,omitempty"`
	WalletHoldID  string  `json:"walletHoldId,omitempty"`
}

type AdminOrderStatus struct {
	Current            string     `json:"current"`
	DisplayStatus      string     `json:"displayStatus"`
	ProviderAcceptedAt *time.Time `json:"providerAcceptedAt,omitempty"`
	ProviderStartedAt  *time.Time `json:"providerStartedAt,omitempty"`
	CompletedAt        *time.Time `json:"completedAt,omitempty"`
	ExpiresAt          *time.Time `json:"expiresAt,omitempty"`
	IsExpired          bool       `json:"isExpired"`
}

type AdminCancellationInfo struct {
	CancelledBy     string    `json:"cancelledBy"`
	CancelledAt     time.Time `json:"cancelledAt"`
	Reason          string    `json:"reason"`
	CancellationFee float64   `json:"cancellationFee"`
	RefundAmount    float64   `json:"refundAmount"`
}

type AdminRatingsInfo struct {
	CustomerRating  *int       `json:"customerRating,omitempty"`
	CustomerReview  string     `json:"customerReview,omitempty"`
	CustomerRatedAt *time.Time `json:"customerRatedAt,omitempty"`
	ProviderRating  *int       `json:"providerRating,omitempty"`
	ProviderReview  string     `json:"providerReview,omitempty"`
	ProviderRatedAt *time.Time `json:"providerRatedAt,omitempty"`
}

type StatusHistoryItem struct {
	ID            string                 `json:"id"`
	FromStatus    string                 `json:"fromStatus"`
	ToStatus      string                 `json:"toStatus"`
	ChangedBy     string                 `json:"changedBy,omitempty"`
	ChangedByRole string                 `json:"changedByRole"`
	Notes         string                 `json:"notes,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
}

// ==================== Analytics Responses ====================

type OverviewAnalyticsResponse struct {
	Period           AnalyticsPeriod    `json:"period"`
	Summary          AnalyticsSummary   `json:"summary"`
	OrdersByStatus   []StatusCount      `json:"ordersByStatus"`
	OrdersByCategory []CategoryCount    `json:"ordersByCategory"`
	RevenueBreakdown []RevenueBreakdown `json:"revenueBreakdown"`
	Trends           AnalyticsTrends    `json:"trends"`
}

type AnalyticsPeriod struct {
	FromDate string `json:"fromDate"`
	ToDate   string `json:"toDate"`
	GroupBy  string `json:"groupBy"`
}

type AnalyticsSummary struct {
	TotalOrders          int     `json:"totalOrders"`
	CompletedOrders      int     `json:"completedOrders"`
	CancelledOrders      int     `json:"cancelledOrders"`
	PendingOrders        int     `json:"pendingOrders"`
	TotalRevenue         float64 `json:"totalRevenue"`
	TotalCommission      float64 `json:"totalCommission"`
	TotalProviderPayouts float64 `json:"totalProviderPayouts"`
	AverageOrderValue    float64 `json:"averageOrderValue"`
	CompletionRate       float64 `json:"completionRate"`
	CancellationRate     float64 `json:"cancellationRate"`
	AverageRating        float64 `json:"averageRating"`
}

type StatusCount struct {
	Status      string  `json:"status"`
	DisplayName string  `json:"displayName"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
}

type CategoryCount struct {
	CategorySlug  string  `json:"categorySlug"`
	CategoryTitle string  `json:"categoryTitle"`
	OrderCount    int     `json:"orderCount"`
	Revenue       float64 `json:"revenue"`
	Percentage    float64 `json:"percentage"`
}

type RevenueBreakdown struct {
	Period            string  `json:"period"` // Date, week, or month
	OrderCount        int     `json:"orderCount"`
	Revenue           float64 `json:"revenue"`
	Commission        float64 `json:"commission"`
	ProviderPayouts   float64 `json:"providerPayouts"`
	AverageOrderValue float64 `json:"averageOrderValue"`
}

type AnalyticsTrends struct {
	OrdersChange     TrendChange `json:"ordersChange"`
	RevenueChange    TrendChange `json:"revenueChange"`
	CompletionChange TrendChange `json:"completionChange"`
}

type TrendChange struct {
	CurrentValue  float64 `json:"currentValue"`
	PreviousValue float64 `json:"previousValue"`
	Change        float64 `json:"change"`        // Absolute change
	ChangePercent float64 `json:"changePercent"` // Percentage change
	Trend         string  `json:"trend"`         // "up", "down", "stable"
}

type ProviderAnalyticsResponse struct {
	Period    AnalyticsPeriod         `json:"period"`
	Providers []ProviderAnalyticsItem `json:"providers"`
	Total     int                     `json:"total"`
}

type ProviderAnalyticsItem struct {
	ProviderID      string   `json:"providerId"`
	ProviderName    string   `json:"providerName"`
	Photo           string   `json:"photo,omitempty"`
	CompletedOrders int      `json:"completedOrders"`
	CancelledOrders int      `json:"cancelledOrders"`
	TotalEarnings   float64  `json:"totalEarnings"`
	AverageRating   float64  `json:"averageRating"`
	TotalRatings    int      `json:"totalRatings"`
	CompletionRate  float64  `json:"completionRate"`
	Categories      []string `json:"categories"`
}

type RevenueReportResponse struct {
	Period           AnalyticsPeriod      `json:"period"`
	TotalRevenue     float64              `json:"totalRevenue"`
	TotalCommission  float64              `json:"totalCommission"`
	TotalPayouts     float64              `json:"totalPayouts"`
	TotalRefunds     float64              `json:"totalRefunds"`
	NetRevenue       float64              `json:"netRevenue"`
	FormattedRevenue string               `json:"formattedRevenue"`
	Breakdown        []RevenueBreakdown   `json:"breakdown"`
	ByCategory       []CategoryRevenue    `json:"byCategory"`
	ByPaymentMethod  []PaymentMethodStats `json:"byPaymentMethod"`
}

type CategoryRevenue struct {
	CategorySlug  string  `json:"categorySlug"`
	CategoryTitle string  `json:"categoryTitle"`
	Revenue       float64 `json:"revenue"`
	Commission    float64 `json:"commission"`
	OrderCount    int     `json:"orderCount"`
	Percentage    float64 `json:"percentage"`
}

type PaymentMethodStats struct {
	Method      string  `json:"method"`
	OrderCount  int     `json:"orderCount"`
	TotalAmount float64 `json:"totalAmount"`
	Percentage  float64 `json:"percentage"`
}

// ==================== Dashboard Responses ====================

type DashboardResponse struct {
	// Today's stats
	Today TodayStats `json:"today"`

	// Recent orders
	RecentOrders []AdminOrderListResponse `json:"recentOrders"`

	// Pending actions
	PendingActions PendingActions `json:"pendingActions"`

	// Quick stats (last 7 days)
	WeeklyStats WeeklyStats `json:"weeklyStats"`
}

type TodayStats struct {
	TotalOrders      int     `json:"totalOrders"`
	CompletedOrders  int     `json:"completedOrders"`
	PendingOrders    int     `json:"pendingOrders"`
	InProgressOrders int     `json:"inProgressOrders"`
	Revenue          float64 `json:"revenue"`
	Commission       float64 `json:"commission"`
	NewCustomers     int     `json:"newCustomers"`
	ActiveProviders  int     `json:"activeProviders"`
}

type PendingActions struct {
	OrdersNeedingProvider int `json:"ordersNeedingProvider"`
	ExpiredOrders         int `json:"expiredOrders"`
	DisputedOrders        int `json:"disputedOrders"`
	UnratedOrders         int `json:"unratedOrders"`
}

type WeeklyStats struct {
	TotalOrders     int          `json:"totalOrders"`
	CompletedOrders int          `json:"completedOrders"`
	TotalRevenue    float64      `json:"totalRevenue"`
	AverageRating   float64      `json:"averageRating"`
	DailyBreakdown  []DailyStats `json:"dailyBreakdown"`
}

type DailyStats struct {
	Date       string  `json:"date"`
	DayName    string  `json:"dayName"`
	OrderCount int     `json:"orderCount"`
	Revenue    float64 `json:"revenue"`
}

// ==================== Conversion Functions ====================

func GetCategoryTitle(slug string) string {
	titles := map[string]string{
		"pest-control": "Pest Control",
		"cleaning":     "Cleaning Services",
		"iv-therapy":   "IV Therapy",
		"massage":      "Massage Therapy",
		"handyman":     "Handyman Services",
	}
	if title, ok := titles[slug]; ok {
		return title
	}
	return slug
}

func GetDisplayStatus(status string) string {
	statuses := map[string]string{
		"pending":            "Pending",
		"searching_provider": "Searching Provider",
		"assigned":           "Assigned",
		"accepted":           "Accepted",
		"in_progress":        "In Progress",
		"completed":          "Completed",
		"cancelled":          "Cancelled",
		"expired":            "Expired",
	}
	if display, ok := statuses[status]; ok {
		return display
	}
	return status
}

func FormatPrice(price float64) string {
	return fmt.Sprintf("$%.2f", price)
}

func FormatDate(dateStr string) string {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return date.Format("January 2, 2006")
}

func FormatTime(timeStr string) string {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("3:04 PM")
}

func CalculateProviderPayout(totalPrice float64) float64 {
	return shared.RoundToTwoDecimals(totalPrice * (1 - shared.PlatformCommissionRate))
}

func GetAvailableActions(status string) []string {
	actions := []string{"view", "view_history"}

	switch status {
	case shared.OrderStatusPending, shared.OrderStatusSearchingProvider:
		actions = append(actions, "cancel", "assign_provider", "update_status")
	case shared.OrderStatusAssigned:
		actions = append(actions, "cancel", "reassign", "update_status")
	case shared.OrderStatusAccepted:
		actions = append(actions, "cancel", "reassign", "force_start", "update_status")
	case shared.OrderStatusInProgress:
		actions = append(actions, "force_complete", "cancel", "update_status")
	case shared.OrderStatusCompleted:
		actions = append(actions, "refund")
	case shared.OrderStatusCancelled:
		actions = append(actions, "reopen")
	}

	return actions
}

func ToAdminOrderListResponse(order *models.ServiceOrderNew) AdminOrderListResponse {
	providerPayout := CalculateProviderPayout(order.TotalPrice)
	commission := order.TotalPrice - providerPayout

	response := AdminOrderListResponse{
		ID:             order.ID,
		OrderNumber:    order.OrderNumber,
		CategorySlug:   order.CategorySlug,
		CategoryTitle:  GetCategoryTitle(order.CategorySlug),
		CustomerName:   order.CustomerInfo.Name,
		CustomerPhone:  order.CustomerInfo.Phone,
		BookingDate:    order.BookingInfo.Date,
		BookingTime:    order.BookingInfo.Time,
		TotalPrice:     order.TotalPrice,
		ProviderPayout: providerPayout,
		Commission:     commission,
		Status:         order.Status,
		DisplayStatus:  GetDisplayStatus(order.Status),
		CreatedAt:      order.CreatedAt,
		CompletedAt:    order.CompletedAt,
	}

	if order.PaymentInfo != nil {
		response.PaymentMethod = order.PaymentInfo.Method
		response.PaymentStatus = order.PaymentInfo.Status
	}

	if order.AssignedProviderID != nil {
		response.ProviderName = "Provider" // Placeholder
	}

	return response
}

func ToAdminOrderListResponses(orders []*models.ServiceOrderNew) []AdminOrderListResponse {
	responses := make([]AdminOrderListResponse, len(orders))
	for i, order := range orders {
		responses[i] = ToAdminOrderListResponse(order)
	}
	return responses
}

func ToAdminOrderDetailResponse(order *models.ServiceOrderNew, history []models.OrderStatusHistory) *AdminOrderDetailResponse {
	providerPayout := CalculateProviderPayout(order.TotalPrice)

	// Build services
	services := make([]AdminOrderServiceItem, len(order.SelectedServices))
	for i, s := range order.SelectedServices {
		services[i] = AdminOrderServiceItem{
			ServiceSlug: s.ServiceSlug,
			Title:       s.Title,
			Price:       s.Price,
			Quantity:    s.Quantity,
			Subtotal:    s.Price * float64(s.Quantity),
		}
	}

	// Build addons
	var addons []AdminOrderAddonItem
	if order.SelectedAddons != nil {
		addons = make([]AdminOrderAddonItem, len(order.SelectedAddons))
		for i, a := range order.SelectedAddons {
			addons[i] = AdminOrderAddonItem{
				AddonSlug: a.AddonSlug,
				Title:     a.Title,
				Price:     a.Price,
				Quantity:  a.Quantity,
				Subtotal:  a.Price * float64(a.Quantity),
			}
		}
	}

	// Build booking info
	bookingDate, _ := time.Parse("2006-01-02", order.BookingInfo.Date)
	today := time.Now().Truncate(24 * time.Hour)
	bookingInfo := AdminBookingInfo{
		Day:  order.BookingInfo.Day,
		Date: order.BookingInfo.Date,
		Time: order.BookingInfo.Time,
		PreferredTime: func() string {
			if order.BookingInfo.PreferredTime.IsZero() {
				return ""
			}
			// format PreferredTime to a human-readable string (e.g., "3:04 PM")
			return FormatTime(order.BookingInfo.PreferredTime.Format("15:04"))
		}(),
		FormattedDate: FormatDate(order.BookingInfo.Date),
		FormattedTime: FormatTime(order.BookingInfo.Time),
		IsPast:        bookingDate.Before(today),
		IsToday:       bookingDate.Equal(today),
	}

	// Build status history
	statusHistory := make([]StatusHistoryItem, len(history))
	for i, h := range history {
		statusHistory[i] = StatusHistoryItem{
			ID:            h.ID,
			FromStatus:    h.FromStatus,
			ToStatus:      h.ToStatus,
			ChangedByRole: h.ChangedByRole,
			Notes:         h.Notes,
			Metadata:      h.Metadata,
			CreatedAt:     h.CreatedAt,
		}
		if h.ChangedBy != nil {
			statusHistory[i].ChangedBy = *h.ChangedBy
		}
	}

	response := &AdminOrderDetailResponse{
		ID:            order.ID,
		OrderNumber:   order.OrderNumber,
		CategorySlug:  order.CategorySlug,
		CategoryTitle: GetCategoryTitle(order.CategorySlug),
		Customer: OrderPartyInfo{
			ID:      order.CustomerID,
			Name:    order.CustomerInfo.Name,
			Email:   order.CustomerInfo.Email,
			Phone:   order.CustomerInfo.Phone,
			Address: order.CustomerInfo.Address,
			Lat:     order.CustomerInfo.Lat,
			Lng:     order.CustomerInfo.Lng,
		},
		BookingInfo:  bookingInfo,
		Services:     services,
		Addons:       addons,
		SpecialNotes: order.SpecialNotes,
		Pricing: AdminOrderPricing{
			ServicesTotal:      order.ServicesTotal,
			AddonsTotal:        order.AddonsTotal,
			Subtotal:           order.Subtotal,
			PlatformCommission: order.PlatformCommission,
			CommissionRate:     shared.PlatformCommissionRate,
			TotalPrice:         order.TotalPrice,
			ProviderPayout:     providerPayout,
			FormattedTotal:     FormatPrice(order.TotalPrice),
		},
		Status: AdminOrderStatus{
			Current:            order.Status,
			DisplayStatus:      GetDisplayStatus(order.Status),
			ProviderAcceptedAt: order.ProviderAcceptedAt,
			ProviderStartedAt:  order.ProviderStartedAt,
			CompletedAt:        order.CompletedAt,
			ExpiresAt:          order.ExpiresAt,
			IsExpired:          order.ExpiresAt != nil && time.Now().After(*order.ExpiresAt),
		},
		StatusHistory:    statusHistory,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
		AvailableActions: GetAvailableActions(order.Status),
	}

	// Add payment info
	if order.PaymentInfo != nil {
		response.Payment = AdminPaymentInfo{
			Method:        order.PaymentInfo.Method,
			Status:        order.PaymentInfo.Status,
			Total:         order.PaymentInfo.Total,
			AmountPaid:    order.PaymentInfo.AmountPaid,
			Voucher:       order.PaymentInfo.Voucher,
			TransactionID: order.PaymentInfo.TransactionID,
		}
		if order.WalletHoldID != nil {
			response.Payment.WalletHoldID = *order.WalletHoldID
		}
	}

	// Add provider if assigned
	if order.AssignedProviderID != nil {
		response.Provider = &OrderPartyInfo{
			ID: *order.AssignedProviderID,
			// TODO: Load from user table
			Name: "Provider Name",
		}
	}

	// Add cancellation info
	if order.CancellationInfo != nil {
		response.Cancellation = &AdminCancellationInfo{
			CancelledBy:     order.CancellationInfo.CancelledBy,
			CancelledAt:     order.CancellationInfo.CancelledAt,
			Reason:          order.CancellationInfo.Reason,
			CancellationFee: order.CancellationInfo.CancellationFee,
			RefundAmount:    order.CancellationInfo.RefundAmount,
		}
	}

	// Add ratings
	if order.CustomerRating != nil || order.ProviderRating != nil {
		response.Ratings = &AdminRatingsInfo{
			CustomerRating:  order.CustomerRating,
			CustomerReview:  order.CustomerReview,
			CustomerRatedAt: order.CustomerRatedAt,
			ProviderRating:  order.ProviderRating,
			ProviderReview:  order.ProviderReview,
			ProviderRatedAt: order.ProviderRatedAt,
		}
	}

	return response
}

// ServiceResponse represents a full service response
type ServiceResponse struct {
	ID                 string    `json:"id"`
	Title              string    `json:"title"`
	LongTitle          string    `json:"longTitle"`
	ServiceSlug        string    `json:"serviceSlug"`
	CategorySlug       string    `json:"categorySlug"`
	Description        string    `json:"description"`
	LongDescription    string    `json:"longDescription"`
	Highlights         string    `json:"highlights"`
	WhatsIncluded      []string  `json:"whatsIncluded"`
	TermsAndConditions []string  `json:"termsAndConditions"`
	BannerImage        string    `json:"bannerImage"`
	Thumbnail          string    `json:"thumbnail"`
	Duration           *int      `json:"duration"`
	IsFrequent         bool      `json:"isFrequent"`
	Frequency          string    `json:"frequency"`
	SortOrder          int       `json:"sortOrder"`
	IsActive           bool      `json:"isActive"`
	IsAvailable        bool      `json:"isAvailable"`
	BasePrice          *float64  `json:"basePrice"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// // HomeCleaningServiceResponse represents a home cleaning service response
// type HomeCleaningServiceResponse struct {
// 	ID          string    `json:"id"`
// 	Title       string    `json:"title"`
// 	ServiceSlug string    `json:"serviceSlug"`
// 	BasePrice   float64   `json:"basePrice"`
// 	CreatedAt   time.Time `json:"createdAt"`
// }

// // ToHomeCleaningServiceResponse converts model to response
// func ToHomeCleaningServiceResponse(service *models.HomeCleaningService) *HomeCleaningServiceResponse {
// 	if service == nil {
// 		return nil
// 	}
// 	return &HomeCleaningServiceResponse{
// 		ID:          service.ID,
// 		Title:       service.Title,
// 		ServiceSlug: service.ServiceSlug,
// 		BasePrice:   service.BasePrice,
// 		CreatedAt:   service.CreatedAt,
// 	}
// }

// // HomeCleaningServiceAddonResponse represents an addon in service response
// type HomeCleaningServiceAddonResponse struct {
// 	CategorySlug string               `json:"categorySlug"`
// 	Addons       []*AddonListResponse `json:"addons"`
// 	TotalCount   int                  `json:"totalCount"`
// }

// ServiceListResponse represents a service in list view (lighter)
type ServiceListResponse struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	ServiceSlug  string    `json:"serviceSlug"`
	CategorySlug string    `json:"categorySlug"`
	Thumbnail    string    `json:"thumbnail"`
	Duration     *int      `json:"duration"`
	IsActive     bool      `json:"isActive"`
	IsAvailable  bool      `json:"isAvailable"`
	BasePrice    *float64  `json:"basePrice"`
	SortOrder    int       `json:"sortOrder"`
	CreatedAt    time.Time `json:"createdAt"`
}

// ToServiceResponse converts model to full response
func ToServiceResponse(service *models.ServiceNew) *ServiceResponse {
	if service == nil {
		return nil
	}

	// // Convert pq.StringArray to []string
	// highlights := make([]string, len(service.Highlights))
	// copy(highlights, service.Highlights)

	whatsIncluded := make([]string, len(service.WhatsIncluded))
	copy(whatsIncluded, service.WhatsIncluded)

	termsAndConditions := make([]string, len(service.TermsAndConditions))
	copy(termsAndConditions, service.TermsAndConditions)

	return &ServiceResponse{
		ID:                 service.ID,
		Title:              service.Title,
		LongTitle:          service.LongTitle,
		ServiceSlug:        service.ServiceSlug,
		CategorySlug:       service.CategorySlug,
		Description:        service.Description,
		LongDescription:    service.LongDescription,
		Highlights:         service.Highlights,
		WhatsIncluded:      whatsIncluded,
		TermsAndConditions: termsAndConditions,
		BannerImage:        service.BannerImage,
		Thumbnail:          service.Thumbnail,
		Duration:           service.Duration,
		IsFrequent:         service.IsFrequent,
		Frequency:          service.Frequency,
		SortOrder:          service.SortOrder,
		IsActive:           service.IsActive,
		IsAvailable:        service.IsAvailable,
		BasePrice:          service.BasePrice,
		CreatedAt:          service.CreatedAt,
		UpdatedAt:          service.UpdatedAt,
	}
}

// ToServiceListResponse converts model to list response
func ToServiceListResponse(service *models.ServiceNew) *ServiceListResponse {
	if service == nil {
		return nil
	}

	return &ServiceListResponse{
		ID:           service.ID,
		Title:        service.Title,
		ServiceSlug:  service.ServiceSlug,
		CategorySlug: service.CategorySlug,
		Thumbnail:    service.Thumbnail,
		Duration:     service.Duration,
		IsActive:     service.IsActive,
		IsAvailable:  service.IsAvailable,
		BasePrice:    service.BasePrice,
		SortOrder:    service.SortOrder,
		CreatedAt:    service.CreatedAt,
	}
}

// ToServiceListResponses converts multiple models to list responses
func ToServiceListResponses(services []*models.ServiceNew) []*ServiceListResponse {
	if services == nil {
		return []*ServiceListResponse{}
	}

	responses := make([]*ServiceListResponse, len(services))
	for i, service := range services {
		responses[i] = ToServiceListResponse(service)
	}
	return responses
}

// CategoryServicesResponse represents services grouped by category
type CategoryServicesResponse struct {
	CategorySlug string                 `json:"categorySlug"`
	Services     []*ServiceListResponse `json:"services"`
	Addons       []*AddonListResponse   `json:"addons"`
	TotalCount   int                    `json:"totalCount"`
}
