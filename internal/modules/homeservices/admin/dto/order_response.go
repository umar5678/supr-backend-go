package dto

import (
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
)

// ==================== Order Responses ====================

// AdminOrderListResponse represents an order in admin list view
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

// AdminOrderDetailResponse represents full order details for admin
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

// OrderPartyInfo represents customer or provider info
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

// AdminBookingInfo represents booking info for admin
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

// AdminOrderServiceItem represents a service item
type AdminOrderServiceItem struct {
	ServiceSlug string  `json:"serviceSlug"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Subtotal    float64 `json:"subtotal"`
}

// AdminOrderAddonItem represents an addon item
type AdminOrderAddonItem struct {
	AddonSlug string  `json:"addonSlug"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
}

// AdminOrderPricing represents pricing breakdown
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

// AdminPaymentInfo represents payment info
type AdminPaymentInfo struct {
	Method        string  `json:"method"`
	Status        string  `json:"status"`
	Total         float64 `json:"total"`
	AmountPaid    float64 `json:"amountPaid"`
	Voucher       string  `json:"voucher,omitempty"`
	TransactionID string  `json:"transactionId,omitempty"`
	WalletHoldID  string  `json:"walletHoldId,omitempty"`
}

// AdminOrderStatus represents order status info
type AdminOrderStatus struct {
	Current            string     `json:"current"`
	DisplayStatus      string     `json:"displayStatus"`
	ProviderAcceptedAt *time.Time `json:"providerAcceptedAt,omitempty"`
	ProviderStartedAt  *time.Time `json:"providerStartedAt,omitempty"`
	CompletedAt        *time.Time `json:"completedAt,omitempty"`
	ExpiresAt          *time.Time `json:"expiresAt,omitempty"`
	IsExpired          bool       `json:"isExpired"`
}

// AdminCancellationInfo represents cancellation info
type AdminCancellationInfo struct {
	CancelledBy     string    `json:"cancelledBy"`
	CancelledAt     time.Time `json:"cancelledAt"`
	Reason          string    `json:"reason"`
	CancellationFee float64   `json:"cancellationFee"`
	RefundAmount    float64   `json:"refundAmount"`
}

// AdminRatingsInfo represents ratings info
type AdminRatingsInfo struct {
	CustomerRating  *int       `json:"customerRating,omitempty"`
	CustomerReview  string     `json:"customerReview,omitempty"`
	CustomerRatedAt *time.Time `json:"customerRatedAt,omitempty"`
	ProviderRating  *int       `json:"providerRating,omitempty"`
	ProviderReview  string     `json:"providerReview,omitempty"`
	ProviderRatedAt *time.Time `json:"providerRatedAt,omitempty"`
}

// StatusHistoryItem represents a status change in history
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

// OverviewAnalyticsResponse represents overview analytics
type OverviewAnalyticsResponse struct {
	Period           AnalyticsPeriod    `json:"period"`
	Summary          AnalyticsSummary   `json:"summary"`
	OrdersByStatus   []StatusCount      `json:"ordersByStatus"`
	OrdersByCategory []CategoryCount    `json:"ordersByCategory"`
	RevenueBreakdown []RevenueBreakdown `json:"revenueBreakdown"`
	Trends           AnalyticsTrends    `json:"trends"`
}

// AnalyticsPeriod represents the analytics period
type AnalyticsPeriod struct {
	FromDate string `json:"fromDate"`
	ToDate   string `json:"toDate"`
	GroupBy  string `json:"groupBy"`
}

// AnalyticsSummary represents summary statistics
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

// StatusCount represents count by status
type StatusCount struct {
	Status      string  `json:"status"`
	DisplayName string  `json:"displayName"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
}

// CategoryCount represents count by category
type CategoryCount struct {
	CategorySlug  string  `json:"categorySlug"`
	CategoryTitle string  `json:"categoryTitle"`
	OrderCount    int     `json:"orderCount"`
	Revenue       float64 `json:"revenue"`
	Percentage    float64 `json:"percentage"`
}

// RevenueBreakdown represents revenue over time
type RevenueBreakdown struct {
	Period            string  `json:"period"` // Date, week, or month
	OrderCount        int     `json:"orderCount"`
	Revenue           float64 `json:"revenue"`
	Commission        float64 `json:"commission"`
	ProviderPayouts   float64 `json:"providerPayouts"`
	AverageOrderValue float64 `json:"averageOrderValue"`
}

// AnalyticsTrends represents trend comparisons
type AnalyticsTrends struct {
	OrdersChange     TrendChange `json:"ordersChange"`
	RevenueChange    TrendChange `json:"revenueChange"`
	CompletionChange TrendChange `json:"completionChange"`
}

// TrendChange represents a trend comparison
type TrendChange struct {
	CurrentValue  float64 `json:"currentValue"`
	PreviousValue float64 `json:"previousValue"`
	Change        float64 `json:"change"`        // Absolute change
	ChangePercent float64 `json:"changePercent"` // Percentage change
	Trend         string  `json:"trend"`         // "up", "down", "stable"
}

// ProviderAnalyticsResponse represents provider analytics
type ProviderAnalyticsResponse struct {
	Period    AnalyticsPeriod         `json:"period"`
	Providers []ProviderAnalyticsItem `json:"providers"`
	Total     int                     `json:"total"`
}

// ProviderAnalyticsItem represents analytics for a single provider
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

// RevenueReportResponse represents revenue report
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

// CategoryRevenue represents revenue by category
type CategoryRevenue struct {
	CategorySlug  string  `json:"categorySlug"`
	CategoryTitle string  `json:"categoryTitle"`
	Revenue       float64 `json:"revenue"`
	Commission    float64 `json:"commission"`
	OrderCount    int     `json:"orderCount"`
	Percentage    float64 `json:"percentage"`
}

// PaymentMethodStats represents stats by payment method
type PaymentMethodStats struct {
	Method      string  `json:"method"`
	OrderCount  int     `json:"orderCount"`
	TotalAmount float64 `json:"totalAmount"`
	Percentage  float64 `json:"percentage"`
}

// ==================== Dashboard Responses ====================

// DashboardResponse represents admin dashboard data
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

// TodayStats represents today's statistics
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

// PendingActions represents items needing attention
type PendingActions struct {
	OrdersNeedingProvider int `json:"ordersNeedingProvider"`
	ExpiredOrders         int `json:"expiredOrders"`
	DisputedOrders        int `json:"disputedOrders"`
	UnratedOrders         int `json:"unratedOrders"`
}

// WeeklyStats represents last 7 days stats
type WeeklyStats struct {
	TotalOrders     int          `json:"totalOrders"`
	CompletedOrders int          `json:"completedOrders"`
	TotalRevenue    float64      `json:"totalRevenue"`
	AverageRating   float64      `json:"averageRating"`
	DailyBreakdown  []DailyStats `json:"dailyBreakdown"`
}

// DailyStats represents stats for a single day
type DailyStats struct {
	Date       string  `json:"date"`
	DayName    string  `json:"dayName"`
	OrderCount int     `json:"orderCount"`
	Revenue    float64 `json:"revenue"`
}

// ==================== Conversion Functions ====================

// GetCategoryTitle returns category title from slug
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

// GetDisplayStatus returns human-readable status
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

// FormatPrice formats price for display
func FormatPrice(price float64) string {
	return fmt.Sprintf("$%.2f", price)
}

// FormatDate formats date for display
func FormatDate(dateStr string) string {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return date.Format("January 2, 2006")
}

// FormatTime formats time for display
func FormatTime(timeStr string) string {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("3:04 PM")
}

// CalculateProviderPayout calculates provider payout (90%)
func CalculateProviderPayout(totalPrice float64) float64 {
	return shared.RoundToTwoDecimals(totalPrice * (1 - shared.PlatformCommissionRate))
}

// GetAvailableActions returns available admin actions for an order
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

// ToAdminOrderListResponse converts order to list response
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

	// TODO: Load provider name from user table
	if order.AssignedProviderID != nil {
		response.ProviderName = "Provider" // Placeholder
	}

	return response
}

// ToAdminOrderListResponses converts multiple orders
func ToAdminOrderListResponses(orders []*models.ServiceOrderNew) []AdminOrderListResponse {
	responses := make([]AdminOrderListResponse, len(orders))
	for i, order := range orders {
		responses[i] = ToAdminOrderListResponse(order)
	}
	return responses
}

// ToAdminOrderDetailResponse converts order to detail response
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
		Day:           order.BookingInfo.Day,
		Date:          order.BookingInfo.Date,
		Time:          order.BookingInfo.Time,
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
