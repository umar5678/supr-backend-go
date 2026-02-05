package dto

import (
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/homeservices/shared"
)

type ProviderProfileResponse struct {
	ID                string                    `json:"id"`
	UserID            string                    `json:"userId"`
	Name              string                    `json:"name"`
	Email             string                    `json:"email"`
	Phone             string                    `json:"phone"`
	Photo             string                    `json:"photo,omitempty"`
	Bio               string                    `json:"bio,omitempty"`
	IsVerified        bool                      `json:"isVerified"`
	IsAvailable       bool                      `json:"isAvailable"`
	YearsOfExperience int                       `json:"yearsOfExperience"`
	ServiceCategories []ServiceCategoryResponse `json:"serviceCategories"`
	Statistics        ProviderStatistics        `json:"statistics"`
	CreatedAt         time.Time                 `json:"createdAt"`
}

type ServiceCategoryResponse struct {
	ID                string    `json:"id"`
	CategorySlug      string    `json:"categorySlug"`
	CategoryTitle     string    `json:"categoryTitle"`
	ExpertiseLevel    string    `json:"expertiseLevel"`
	YearsOfExperience int       `json:"yearsOfExperience"`
	IsActive          bool      `json:"isActive"`
	CompletedJobs     int       `json:"completedJobs"`
	TotalEarnings     float64   `json:"totalEarnings"`
	AverageRating     float64   `json:"averageRating"`
	TotalRatings      int       `json:"totalRatings"`
	CreatedAt         time.Time `json:"createdAt"`
}

type ProviderStatistics struct {
	TotalCompletedJobs   int     `json:"totalCompletedJobs"`
	TotalEarnings        float64 `json:"totalEarnings"`
	OverallRating        float64 `json:"overallRating"`
	TotalRatings         int     `json:"totalRatings"`
	AcceptanceRate       float64 `json:"acceptanceRate"`
	CompletionRate       float64 `json:"completionRate"`
	AverageResponseTime  int     `json:"averageResponseTime"` // in minutes
	TotalActiveOrders    int     `json:"totalActiveOrders"`
	TodayCompletedOrders int     `json:"todayCompletedOrders"`
	TodayEarnings        float64 `json:"todayEarnings"`
}

type AvailableOrderResponse struct {
	ID              string             `json:"id"`
	OrderNumber     string             `json:"orderNumber"`
	CategorySlug    string             `json:"categorySlug"`
	CategoryTitle   string             `json:"categoryTitle"`
	CustomerInfo    OrderCustomerInfo  `json:"customerInfo"`
	BookingInfo     OrderBookingInfo   `json:"bookingInfo"`
	Services        []OrderServiceItem `json:"services"`
	Addons          []OrderAddonItem   `json:"addons,omitempty"`
	SpecialNotes    string             `json:"specialNotes,omitempty"`
	TotalPrice      float64            `json:"totalPrice"`
	ProviderPayout  float64            `json:"providerPayout"` // 90% of total
	FormattedPayout string             `json:"formattedPayout"`
	Distance        *float64           `json:"distance,omitempty"` // km from provider
	CreatedAt       time.Time          `json:"createdAt"`
	ExpiresAt       *time.Time         `json:"expiresAt,omitempty"`
}

type OrderCustomerInfo struct {
	Name    string  `json:"name"`
	Phone   string  `json:"phone,omitempty"` // Only shown after acceptance
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

type OrderBookingInfo struct {
	Day           string `json:"day"`
	Date          string `json:"date"`
	Time          string `json:"time"`
	PreferredTime string `json:"preferredTime,omitempty"`
	FormattedDate string `json:"formattedDate"`
	FormattedTime string `json:"formattedTime"`
}

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

type ProviderOrderResponse struct {
	ID              string             `json:"id"`
	OrderNumber     string             `json:"orderNumber"`
	CategorySlug    string             `json:"categorySlug"`
	CategoryTitle   string             `json:"categoryTitle"`
	CustomerInfo    OrderCustomerInfo  `json:"customerInfo"`
	BookingInfo     OrderBookingInfo   `json:"bookingInfo"`
	Services        []OrderServiceItem `json:"services"`
	Addons          []OrderAddonItem   `json:"addons,omitempty"`
	SpecialNotes    string             `json:"specialNotes,omitempty"`
	TotalPrice      float64            `json:"totalPrice"`
	ProviderPayout  float64            `json:"providerPayout"`
	FormattedPayout string             `json:"formattedPayout"`
	Status          OrderStatusInfo    `json:"status"`
	Rating          *OrderRatingInfo   `json:"rating,omitempty"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`
}

type OrderStatusInfo struct {
	Current       string     `json:"current"`
	DisplayStatus string     `json:"displayStatus"`
	AcceptedAt    *time.Time `json:"acceptedAt,omitempty"`
	StartedAt     *time.Time `json:"startedAt,omitempty"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
	CanStart      bool       `json:"canStart"`
	CanComplete   bool       `json:"canComplete"`
	CanRate       bool       `json:"canRate"`
}

type OrderRatingInfo struct {
	CustomerRating  *int       `json:"customerRating,omitempty"`
	CustomerReview  string     `json:"customerReview,omitempty"`
	CustomerRatedAt *time.Time `json:"customerRatedAt,omitempty"`
	ProviderRating  *int       `json:"providerRating,omitempty"`
	ProviderReview  string     `json:"providerReview,omitempty"`
	ProviderRatedAt *time.Time `json:"providerRatedAt,omitempty"`
}

type ProviderOrderListResponse struct {
	ID              string           `json:"id"`
	OrderNumber     string           `json:"orderNumber"`
	CategorySlug    string           `json:"categorySlug"`
	CategoryTitle   string           `json:"categoryTitle"`
	CustomerName    string           `json:"customerName"`
	BookingInfo     OrderBookingInfo `json:"bookingInfo"`
	ProviderPayout  float64          `json:"providerPayout"`
	FormattedPayout string           `json:"formattedPayout"`
	Status          string           `json:"status"`
	DisplayStatus   string           `json:"displayStatus"`
	CreatedAt       time.Time        `json:"createdAt"`
}

type EarningsSummaryResponse struct {
	TotalEarnings   float64             `json:"totalEarnings"`
	TotalOrders     int                 `json:"totalOrders"`
	AveragePerOrder float64             `json:"averagePerOrder"`
	FormattedTotal  string              `json:"formattedTotal"`
	Period          EarningsPeriod      `json:"period"`
	Breakdown       []EarningsBreakdown `json:"breakdown"`
	ByCategory      []CategoryEarnings  `json:"byCategory"`
}

type EarningsPeriod struct {
	FromDate string `json:"fromDate"`
	ToDate   string `json:"toDate"`
}

type EarningsBreakdown struct {
	Period            string  `json:"period"` // date, week number, or month
	Earnings          float64 `json:"earnings"`
	OrderCount        int     `json:"orderCount"`
	FormattedEarnings string  `json:"formattedEarnings"`
}

type CategoryEarnings struct {
	CategorySlug  string  `json:"categorySlug"`
	CategoryTitle string  `json:"categoryTitle"`
	Earnings      float64 `json:"earnings"`
	OrderCount    int     `json:"orderCount"`
	Percentage    float64 `json:"percentage"`
}

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
		"searching_provider": "Searching",
		"assigned":           "Assigned to You",
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

func FormatDate(dateStr string) string {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return date.Format("Monday, January 2, 2006")
}

func FormatTime(timeStr string) string {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("3:04 PM")
}

func FormatPrice(price float64) string {
	return fmt.Sprintf("$%.2f", price)
}

func CalculateProviderPayout(totalPrice float64) float64 {
	return shared.RoundToTwoDecimals(totalPrice * (1 - shared.PlatformCommissionRate))
}

func ToOrderBookingInfo(info models.BookingInfo) OrderBookingInfo {
	var preferred string
	if !info.PreferredTime.IsZero() {
		preferred = FormatTime(info.PreferredTime.Format("15:04"))
	}
	return OrderBookingInfo{
		Day:           info.Day,
		Date:          info.Date,
		Time:          info.Time,
		PreferredTime: preferred,
		FormattedDate: FormatDate(info.Date),
		FormattedTime: FormatTime(info.Time),
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

func ToServiceCategoryResponse(category *models.ProviderServiceCategory) ServiceCategoryResponse {
	return ServiceCategoryResponse{
		ID:                category.ID,
		CategorySlug:      category.CategorySlug,
		CategoryTitle:     GetCategoryTitle(category.CategorySlug),
		ExpertiseLevel:    category.ExpertiseLevel,
		YearsOfExperience: category.YearsOfExperience,
		IsActive:          category.IsActive,
		CompletedJobs:     category.CompletedJobs,
		TotalEarnings:     category.TotalEarnings,
		AverageRating:     category.AverageRating,
		TotalRatings:      category.TotalRatings,
		CreatedAt:         category.CreatedAt,
	}
}

func ToServiceCategoryResponses(categories []*models.ProviderServiceCategory) []ServiceCategoryResponse {
	responses := make([]ServiceCategoryResponse, len(categories))
	for i, cat := range categories {
		responses[i] = ToServiceCategoryResponse(cat)
	}
	return responses
}

func ToAvailableOrderResponse(order *models.ServiceOrderNew, distance *float64) AvailableOrderResponse {
	providerPayout := CalculateProviderPayout(order.TotalPrice)

	return AvailableOrderResponse{
		ID:            order.ID,
		OrderNumber:   order.OrderNumber,
		CategorySlug:  order.CategorySlug,
		CategoryTitle: GetCategoryTitle(order.CategorySlug),
		CustomerInfo: OrderCustomerInfo{
			Name:    order.CustomerInfo.Name,
			Address: order.CustomerInfo.Address,
			Lat:     order.CustomerInfo.Lat,
			Lng:     order.CustomerInfo.Lng,
		},
		BookingInfo:     ToOrderBookingInfo(order.BookingInfo),
		Services:        ToOrderServiceItems(order.SelectedServices),
		Addons:          ToOrderAddonItems(order.SelectedAddons),
		SpecialNotes:    order.SpecialNotes,
		TotalPrice:      order.TotalPrice,
		ProviderPayout:  providerPayout,
		FormattedPayout: FormatPrice(providerPayout),
		Distance:        distance,
		CreatedAt:       order.CreatedAt,
		ExpiresAt:       order.ExpiresAt,
	}
}

func ToProviderOrderResponse(order *models.ServiceOrderNew) *ProviderOrderResponse {
	providerPayout := CalculateProviderPayout(order.TotalPrice)

	response := &ProviderOrderResponse{
		ID:            order.ID,
		OrderNumber:   order.OrderNumber,
		CategorySlug:  order.CategorySlug,
		CategoryTitle: GetCategoryTitle(order.CategorySlug),
		CustomerInfo: OrderCustomerInfo{
			Name:    order.CustomerInfo.Name,
			Phone:   order.CustomerInfo.Phone, // Shown for accepted orders
			Address: order.CustomerInfo.Address,
			Lat:     order.CustomerInfo.Lat,
			Lng:     order.CustomerInfo.Lng,
		},
		BookingInfo:     ToOrderBookingInfo(order.BookingInfo),
		Services:        ToOrderServiceItems(order.SelectedServices),
		Addons:          ToOrderAddonItems(order.SelectedAddons),
		SpecialNotes:    order.SpecialNotes,
		TotalPrice:      order.TotalPrice,
		ProviderPayout:  providerPayout,
		FormattedPayout: FormatPrice(providerPayout),
		Status: OrderStatusInfo{
			Current:       order.Status,
			DisplayStatus: GetDisplayStatus(order.Status),
			AcceptedAt:    order.ProviderAcceptedAt,
			StartedAt:     order.ProviderStartedAt,
			CompletedAt:   order.CompletedAt,
			CanStart:      order.Status == shared.OrderStatusAccepted,
			CanComplete:   order.Status == shared.OrderStatusInProgress,
			CanRate:       order.Status == shared.OrderStatusCompleted && order.ProviderRating == nil,
		},
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}

	if order.Status == shared.OrderStatusCompleted {
		response.Rating = &OrderRatingInfo{
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

func ToProviderOrderListResponse(order *models.ServiceOrderNew) ProviderOrderListResponse {
	providerPayout := CalculateProviderPayout(order.TotalPrice)

	return ProviderOrderListResponse{
		ID:              order.ID,
		OrderNumber:     order.OrderNumber,
		CategorySlug:    order.CategorySlug,
		CategoryTitle:   GetCategoryTitle(order.CategorySlug),
		CustomerName:    order.CustomerInfo.Name,
		BookingInfo:     ToOrderBookingInfo(order.BookingInfo),
		ProviderPayout:  providerPayout,
		FormattedPayout: FormatPrice(providerPayout),
		Status:          order.Status,
		DisplayStatus:   GetDisplayStatus(order.Status),
		CreatedAt:       order.CreatedAt,
	}
}

func ToProviderOrderListResponses(orders []*models.ServiceOrderNew) []ProviderOrderListResponse {
	responses := make([]ProviderOrderListResponse, len(orders))
	for i, order := range orders {
		responses[i] = ToProviderOrderListResponse(order)
	}
	return responses
}
