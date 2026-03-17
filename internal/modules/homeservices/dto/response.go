package homeServiceDto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
	authdto "github.com/umar5678/go-backend/internal/modules/auth/dto"
)

type ProviderProfileResponse struct {
	ID     string                `json:"id"`
	UserID string                `json:"userId"`
	User   *authdto.UserResponse `json:"user,omitempty"`

	BusinessName    *string `json:"businessName,omitempty"`
	Description     *string `json:"description,omitempty"`
	ServiceCategory string  `json:"serviceCategory"`

	Status           string   `json:"status"`
	IsVerified       bool     `json:"isVerified"`
	VerificationDocs []string `json:"verificationDocs,omitempty"`

	Rating        float64 `json:"rating"`
	TotalReviews  int     `json:"totalReviews"`
	CompletedJobs int     `json:"completedJobs"`

	IsAvailable  bool     `json:"isAvailable"`
	WorkingHours *string  `json:"workingHours,omitempty"`
	ServiceAreas []string `json:"serviceAreas,omitempty"`

	HourlyRate *float64 `json:"hourlyRate,omitempty"`
	Currency   string   `json:"currency"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type CategoryWithTabsResponse struct {
	ID          uint                 `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	IconURL     string               `json:"iconUrl"`
	BannerImage string               `json:"bannerImage"`
	Highlights  []string             `json:"highlights"`
	IsActive    bool                 `json:"isActive"`
	SortOrder   int                  `json:"sortOrder"`
	CreatedAt   time.Time            `json:"createdAt"`
	Tabs        []ServiceTabResponse `json:"tabs"`
}

type ServiceTabResponse struct {
	ID            uint      `json:"id"`
	CategoryID    uint      `json:"categoryId"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	IconURL       string    `json:"iconUrl"`
	BannerTitle   string    `json:"bannerTitle,omitempty"`
	BannerDesc    string    `json:"bannerDescription,omitempty"`
	BannerImage   string    `json:"bannerImage,omitempty"`
	IsActive      bool      `json:"isActive"`
	SortOrder     int       `json:"sortOrder"`
	ServicesCount int       `json:"servicesCount"`
	CreatedAt     time.Time `json:"createdAt"`
}

type ServiceCategoryResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IconURL     string    `json:"iconUrl"`
	IsActive    bool      `json:"isActive"`
	SortOrder   int       `json:"sortOrder"`
	CreatedAt   time.Time `json:"createdAt"`
}

func ToServiceCategoryResponse(cat *models.ServiceCategory) *ServiceCategoryResponse {
	return &ServiceCategoryResponse{
		ID:          cat.ID,
		Name:        cat.Name,
		Description: cat.Description,
		IconURL:     cat.IconURL,
		IsActive:    cat.IsActive,
		SortOrder:   cat.SortOrder,
		CreatedAt:   cat.CreatedAt,
	}
}

func ToServiceCategoryList(cats []models.ServiceCategory) []*ServiceCategoryResponse {
	result := make([]*ServiceCategoryResponse, len(cats))
	for i, cat := range cats {
		result[i] = ToServiceCategoryResponse(&cat)
	}
	return result
}

type ServiceResponse struct {
	ID                  uint                     `json:"id"`
	CategoryID          uint                     `json:"categoryId"`
	Name                string                   `json:"name"`
	Description         string                   `json:"description"`
	ImageURL            string                   `json:"imageUrl"`
	BasePrice           float64                  `json:"basePrice"`
	OriginalPrice       float64                  `json:"originalPrice,omitempty"`
	DiscountPercentage  int                      `json:"discountPercentage"`
	PricingModel        string                   `json:"pricingModel"`
	BaseDurationMinutes int                      `json:"baseDurationMinutes"`
	IsActive            bool                     `json:"isActive"`
	CreatedAt           time.Time                `json:"createdAt"`
	Category            *ServiceCategoryResponse `json:"category,omitempty"`
	Options             []ServiceOptionResponse  `json:"options,omitempty"`
}

type ServiceListResponse struct {
	ID                 uint      `json:"id"`
	CategoryID         uint      `json:"categoryId"`
	TabID              uint      `json:"tabId"`
	Name               string    `json:"name"`
	ImageURL           string    `json:"imageUrl"`
	BasePrice          float64   `json:"basePrice"`
	OriginalPrice      float64   `json:"originalPrice,omitempty"`
	DiscountPercentage int       `json:"discountPercentage"`
	DurationMinutes    int       `json:"durationMinutes"`
	IsActive           bool      `json:"isActive"`
	IsFeatured         bool      `json:"isFeatured"`
	CreatedAt          time.Time `json:"createdAt"`
}

type ServiceDetailResponse struct {
	ID                  uint                    `json:"id"`
	CategoryID          uint                    `json:"categoryId"`
	TabID               uint                    `json:"tabId"`
	Name                string                  `json:"name"`
	Description         string                  `json:"description"`
	ImageURL            string                  `json:"imageUrl"`
	BasePrice           float64                 `json:"basePrice"`
	OriginalPrice       float64                 `json:"originalPrice,omitempty"`
	DiscountPercentage  int                     `json:"discountPercentage"`
	PricingModel        string                  `json:"pricingModel"`
	BaseDurationMinutes int                     `json:"baseDurationMinutes"`
	MaxQuantity         int                     `json:"maxQuantity"`
	IsActive            bool                    `json:"isActive"`
	IsFeatured          bool                    `json:"isFeatured"`
	CreatedAt           time.Time               `json:"createdAt"`
	Category            *CategoryBasicResponse  `json:"category,omitempty"`
	Tab                 *TabBasicResponse       `json:"tab,omitempty"`
	Options             []ServiceOptionResponse `json:"options,omitempty"`
}

type CategoryBasicResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type TabBasicResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ServiceOptionResponse struct {
	ID         uint                          `json:"id"`
	ServiceID  uint                          `json:"serviceId"`
	Name       string                        `json:"name"`
	Type       string                        `json:"type"`
	IsRequired bool                          `json:"isRequired"`
	Choices    []ServiceOptionChoiceResponse `json:"choices,omitempty"`
}

type ServiceOptionChoiceResponse struct {
	ID                      uint    `json:"id"`
	OptionID                uint    `json:"optionId"`
	Label                   string  `json:"label"`
	PriceModifier           float64 `json:"priceModifier"`
	DurationModifierMinutes int     `json:"durationModifierMinutes"`
}

func calculateDiscountPercentage(originalPrice, basePrice float64) int {
	if originalPrice > 0 && basePrice < originalPrice {
		return int(((originalPrice - basePrice) / originalPrice) * 100)
	}
	return 0
}

func ToServiceResponse(svc *models.Service) *ServiceResponse {
	resp := &ServiceResponse{
		ID:                  svc.ID,
		CategoryID:          svc.CategoryID,
		Name:                svc.Name,
		Description:         svc.Description,
		ImageURL:            svc.ImageURL,
		BasePrice:           svc.BasePrice,
		OriginalPrice:       svc.OriginalPrice,
		DiscountPercentage:  calculateDiscountPercentage(svc.OriginalPrice, svc.BasePrice),
		PricingModel:        svc.PricingModel,
		BaseDurationMinutes: svc.BaseDurationMinutes,
		IsActive:            svc.IsActive,
		CreatedAt:           svc.CreatedAt,
	}

	if svc.Category != nil {
		resp.Category = ToServiceCategoryResponse(svc.Category)
	}

	if len(svc.Options) > 0 {
		resp.Options = make([]ServiceOptionResponse, len(svc.Options))
		for i, opt := range svc.Options {
			resp.Options[i] = ServiceOptionResponse{
				ID:         opt.ID,
				ServiceID:  opt.ServiceID,
				Name:       opt.Name,
				Type:       opt.Type,
				IsRequired: opt.IsRequired,
			}
			if len(opt.Choices) > 0 {
				resp.Options[i].Choices = make([]ServiceOptionChoiceResponse, len(opt.Choices))
				for j, choice := range opt.Choices {
					resp.Options[i].Choices[j] = ServiceOptionChoiceResponse{
						ID:                      choice.ID,
						OptionID:                choice.OptionID,
						Label:                   choice.Label,
						PriceModifier:           choice.PriceModifier,
						DurationModifierMinutes: choice.DurationModifierMinutes,
					}
				}
			}
		}
	}

	return resp
}

func ToServiceListResponse(svc *models.Service) *ServiceListResponse {
	return &ServiceListResponse{
		ID:                 svc.ID,
		CategoryID:         svc.CategoryID,
		TabID:              svc.TabID,
		Name:               svc.Name,
		ImageURL:           svc.ImageURL,
		BasePrice:          svc.BasePrice,
		OriginalPrice:      svc.OriginalPrice,
		DiscountPercentage: calculateDiscountPercentage(svc.OriginalPrice, svc.BasePrice),
		DurationMinutes:    svc.BaseDurationMinutes,
		IsActive:           svc.IsActive,
		IsFeatured:         svc.IsFeatured,
		CreatedAt:          svc.CreatedAt,
	}
}

func ToServiceListResponses(svcs []*models.Service) []*ServiceListResponse {
	result := make([]*ServiceListResponse, len(svcs))
	for i, svc := range svcs {
		result[i] = ToServiceListResponse(svc)
	}
	return result
}

type AddOnResponse struct {
	ID              uint    `json:"id"`
	CategoryID      uint    `json:"categoryId"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	ImageURL        string  `json:"imageUrl"`
	Price           float64 `json:"price"`
	OriginalPrice   float64 `json:"originalPrice,omitempty"`
	DurationMinutes int     `json:"durationMinutes"`
	IsActive        bool    `json:"isActive"`
	SortOrder       int     `json:"sortOrder"`
}

type OrderResponse struct {
	ID             string               `json:"id"`
	Code           string               `json:"code"`
	UserID         string               `json:"userId"`
	ProviderID     *string              `json:"providerId,omitempty"`
	Status         string               `json:"status"`
	Address        string               `json:"address"`
	ServiceDate    time.Time            `json:"serviceDate"`
	Frequency      string               `json:"frequency"`
	QuantityOfPros int                  `json:"quantityOfPros"`
	HoursOfService float64              `json:"hoursOfService"`
	Subtotal       float64              `json:"subtotal"`
	Discount       float64              `json:"discount"`
	SurgeFee       float64              `json:"surgeFee"`
	PlatformFee    float64              `json:"platformFee"`
	Total          float64              `json:"total"`
	CouponCode     *string              `json:"couponCode,omitempty"`
	CreatedAt      time.Time            `json:"createdAt"`
	AcceptedAt     *time.Time           `json:"acceptedAt,omitempty"`
	CompletedAt    *time.Time           `json:"completedAt,omitempty"`
	Items          []OrderItemResponse  `json:"items,omitempty"`
	AddOns         []OrderAddOnResponse `json:"addOns,omitempty"`
	Provider       *ProviderResponse    `json:"provider,omitempty"`
}

type OrderListResponse struct {
	ID             string    `json:"id"`
	Code           string    `json:"code"`
	Status         string    `json:"status"`
	Address        string    `json:"address"`
	ServiceDate    time.Time `json:"serviceDate"`
	QuantityOfPros int       `json:"quantityOfPros"` 
	HoursOfService float64   `json:"hoursOfService"` 
	Total          float64   `json:"total"`
	CreatedAt      time.Time `json:"createdAt"`
}

type OrderItemResponse struct {
	ID              uint                   `json:"id"`
	ServiceID       string                 `json:"serviceId"`
	ServiceName     string                 `json:"serviceName"`
	BasePrice       float64                `json:"basePrice"`
	CalculatedPrice float64                `json:"calculatedPrice"`
	DurationMinutes int                    `json:"durationMinutes"`
	SelectedOptions map[string]interface{} `json:"selectedOptions"`
}

type OrderAddOnResponse struct {
	ID      uint    `json:"id"`
	AddOnID uint    `json:"addOnId"`
	Title   string  `json:"title"`
	Price   float64 `json:"price"`
}

type ProviderResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Phone         *string `json:"phone,omitempty"`
	Photo         *string `json:"photo,omitempty"`
	Rating        float64 `json:"rating"`
	CompletedJobs int     `json:"completedJobs"`
}

func ToCategoryWithTabsResponse(cat *models.ServiceCategory) *CategoryWithTabsResponse {
	resp := &CategoryWithTabsResponse{
		ID:          cat.ID,
		Name:        cat.Name,
		Description: cat.Description,
		IconURL:     cat.IconURL,
		BannerImage: cat.BannerImage,
		Highlights:  []string(cat.Highlights),
		IsActive:    cat.IsActive,
		SortOrder:   cat.SortOrder,
		CreatedAt:   cat.CreatedAt,
		Tabs:        make([]ServiceTabResponse, 0),
	}

	for _, tab := range cat.Tabs {
		resp.Tabs = append(resp.Tabs, ServiceTabResponse{
			ID:          tab.ID,
			CategoryID:  tab.CategoryID,
			Name:        tab.Name,
			Description: tab.Description,
			IconURL:     tab.IconURL,
			BannerTitle: tab.BannerTitle,
			BannerDesc:  tab.BannerDesc,
			BannerImage: tab.BannerImage,
			IsActive:    tab.IsActive,
			SortOrder:   tab.SortOrder,
			CreatedAt:   tab.CreatedAt,
		})
	}

	return resp
}

func ToOrderResponseFromNew(orderNew *models.ServiceOrderNew) *OrderResponse {
	if orderNew == nil {
		return nil
	}

	resp := &OrderResponse{
		ID:             orderNew.ID,
		Code:           orderNew.OrderNumber,
		UserID:         orderNew.CustomerID,
		ProviderID:     orderNew.AssignedProviderID,
		Status:         orderNew.Status,
		Address:        orderNew.CustomerInfo.Address,
		ServiceDate:    orderNew.CreatedAt,
		Frequency:      "once", 
		QuantityOfPros: orderNew.BookingInfo.QuantityOfPros,
		Subtotal:       orderNew.Subtotal,
		PlatformFee:    orderNew.PlatformCommission,
		Total:          orderNew.TotalPrice,
		CreatedAt:      orderNew.CreatedAt,
		AcceptedAt:     orderNew.ProviderAcceptedAt,
		CompletedAt:    orderNew.ProviderCompletedAt,
	}

	if len(orderNew.SelectedServices) > 0 {
		resp.Items = make([]OrderItemResponse, len(orderNew.SelectedServices))
		for i, svc := range orderNew.SelectedServices {
			resp.Items[i] = OrderItemResponse{
				ServiceID:       svc.ServiceSlug,
				ServiceName:     svc.Title,
				BasePrice:       svc.Price,
				CalculatedPrice: svc.Price * float64(svc.Quantity),
				SelectedOptions: make(map[string]interface{}),
			}
		}
	}

	if len(orderNew.SelectedAddons) > 0 {
		resp.AddOns = make([]OrderAddOnResponse, len(orderNew.SelectedAddons))
		for i, addon := range orderNew.SelectedAddons {
			resp.AddOns[i] = OrderAddOnResponse{
				Title: addon.Title,
				Price: addon.Price * float64(addon.Quantity),
			}
		}
	}

	if orderNew.AssignedProvider != nil {
		providerName := "Service Provider"
		var providerPhone *string
		var providerPhoto *string

		if orderNew.AssignedProvider.User != nil {
			providerName = orderNew.AssignedProvider.User.Name
			providerPhone = orderNew.AssignedProvider.User.Phone
			providerPhoto = orderNew.AssignedProvider.User.ProfilePhotoURL
		}

		resp.Provider = &ProviderResponse{
			ID:            orderNew.AssignedProvider.ID,
			Name:          providerName,
			Phone:         providerPhone,
			Photo:         providerPhoto,
			Rating:        orderNew.AssignedProvider.Rating,
			CompletedJobs: orderNew.AssignedProvider.CompletedJobs,
		}
	}

	return resp
}

func ToOrderListResponse(order *models.ServiceOrderNew) *OrderListResponse {
	return &OrderListResponse{
		ID:             order.ID,
		Code:           order.OrderNumber,
		Status:         order.Status,
		Address:        order.CustomerInfo.Address,
		ServiceDate:    order.CreatedAt,
		QuantityOfPros: order.BookingInfo.QuantityOfPros,
		Total:          order.TotalPrice,
		CreatedAt:      order.CreatedAt,
	}
}

func ToOrderListResponses(orders []*models.ServiceOrderNew) []*OrderListResponse {
	result := make([]*OrderListResponse, len(orders))
	for i, order := range orders {
		result[i] = ToOrderListResponse(order)
	}
	return result
}

func ToServiceDetailResponse(svc *models.Service) *ServiceDetailResponse {
	resp := &ServiceDetailResponse{
		ID:                  svc.ID,
		CategoryID:          svc.CategoryID,
		TabID:               svc.TabID,
		Name:                svc.Name,
		Description:         svc.Description,
		ImageURL:            svc.ImageURL,
		BasePrice:           svc.BasePrice,
		OriginalPrice:       svc.OriginalPrice,
		DiscountPercentage:  calculateDiscountPercentage(svc.OriginalPrice, svc.BasePrice),
		PricingModel:        svc.PricingModel,
		BaseDurationMinutes: svc.BaseDurationMinutes,
		MaxQuantity:         svc.MaxQuantity,
		IsActive:            svc.IsActive,
		IsFeatured:          svc.IsFeatured,
		CreatedAt:           svc.CreatedAt,
	}

	if svc.Category != nil {
		resp.Category = &CategoryBasicResponse{
			ID:   svc.Category.ID,
			Name: svc.Category.Name,
		}
	}

	if svc.Tab != nil {
		resp.Tab = &TabBasicResponse{
			ID:   svc.Tab.ID,
			Name: svc.Tab.Name,
		}
	}

	if len(svc.Options) > 0 {
		resp.Options = make([]ServiceOptionResponse, len(svc.Options))
		for i, opt := range svc.Options {
			resp.Options[i] = ServiceOptionResponse{
				ID:         opt.ID,
				ServiceID:  opt.ServiceID,
				Name:       opt.Name,
				Type:       opt.Type,
				IsRequired: opt.IsRequired,
			}
			if len(opt.Choices) > 0 {
				resp.Options[i].Choices = make([]ServiceOptionChoiceResponse, len(opt.Choices))
				for j, choice := range opt.Choices {
					resp.Options[i].Choices[j] = ServiceOptionChoiceResponse{
						ID:                      choice.ID,
						OptionID:                choice.OptionID,
						Label:                   choice.Label,
						PriceModifier:           choice.PriceModifier,
						DurationModifierMinutes: choice.DurationModifierMinutes,
					}
				}
			}
		}
	}

	return resp
}

func ToAddOnResponse(addon *models.AddOnService) *AddOnResponse {
	return &AddOnResponse{
		ID:              addon.ID,
		CategoryID:      addon.CategoryID,
		Title:           addon.Title,
		Description:     addon.Description,
		ImageURL:        addon.ImageURL,
		Price:           addon.Price,
		OriginalPrice:   addon.OriginalPrice,
		DurationMinutes: addon.DurationMinutes,
		IsActive:        addon.IsActive,
		SortOrder:       addon.SortOrder,
	}
}
