package dto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
)

// ServiceCategoryResponse for category listings
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

// ServiceResponse for detailed service info
type ServiceResponse struct {
	ID                  uint                     `json:"id"`
	CategoryID          uint                     `json:"categoryId"`
	Name                string                   `json:"name"`
	Description         string                   `json:"description"`
	ImageURL            string                   `json:"imageUrl"`
	BasePrice           float64                  `json:"basePrice"`
	PricingModel        string                   `json:"pricingModel"`
	BaseDurationMinutes int                      `json:"baseDurationMinutes"`
	IsActive            bool                     `json:"isActive"`
	CreatedAt           time.Time                `json:"createdAt"`
	Category            *ServiceCategoryResponse `json:"category,omitempty"`
	Options             []ServiceOptionResponse  `json:"options,omitempty"`
}

type ServiceListResponse struct {
	ID         uint      `json:"id"`
	CategoryID uint      `json:"categoryId"`
	Name       string    `json:"name"`
	ImageURL   string    `json:"imageUrl"`
	BasePrice  float64   `json:"basePrice"`
	IsActive   bool      `json:"isActive"`
	CreatedAt  time.Time `json:"createdAt"`
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

func ToServiceResponse(svc *models.Service) *ServiceResponse {
	resp := &ServiceResponse{
		ID:                  svc.ID,
		CategoryID:          svc.CategoryID,
		Name:                svc.Name,
		Description:         svc.Description,
		ImageURL:            svc.ImageURL,
		BasePrice:           svc.BasePrice,
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
		ID:         svc.ID,
		CategoryID: svc.CategoryID,
		Name:       svc.Name,
		ImageURL:   svc.ImageURL,
		BasePrice:  svc.BasePrice,
		IsActive:   svc.IsActive,
		CreatedAt:  svc.CreatedAt,
	}
}

func ToServiceListResponses(svcs []*models.Service) []*ServiceListResponse {
	result := make([]*ServiceListResponse, len(svcs))
	for i, svc := range svcs {
		result[i] = ToServiceListResponse(svc)
	}
	return result
}

// OrderResponse for order details
type OrderResponse struct {
	ID          string              `json:"id"`
	Code        string              `json:"code"`
	UserID      string              `json:"userId"`
	ProviderID  *string             `json:"providerId,omitempty"`
	Status      string              `json:"status"`
	Address     string              `json:"address"`
	ServiceDate time.Time           `json:"serviceDate"`
	Frequency   string              `json:"frequency"`
	Subtotal    float64             `json:"subtotal"`
	Discount    float64             `json:"discount"`
	SurgeFee    float64             `json:"surgeFee"`
	PlatformFee float64             `json:"platformFee"`
	Total       float64             `json:"total"`
	CouponCode  *string             `json:"couponCode,omitempty"`
	CreatedAt   time.Time           `json:"createdAt"`
	AcceptedAt  *time.Time          `json:"acceptedAt,omitempty"`
	CompletedAt *time.Time          `json:"completedAt,omitempty"`
	Items       []OrderItemResponse `json:"items,omitempty"`
	Provider    *ProviderResponse   `json:"provider,omitempty"`
}

type OrderListResponse struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Status      string    `json:"status"`
	Address     string    `json:"address"`
	ServiceDate time.Time `json:"serviceDate"`
	Total       float64   `json:"total"`
	CreatedAt   time.Time `json:"createdAt"`
}

type OrderItemResponse struct {
	ID              uint                   `json:"id"`
	ServiceID       uint                   `json:"serviceId"`
	ServiceName     string                 `json:"serviceName"`
	BasePrice       float64                `json:"basePrice"`
	CalculatedPrice float64                `json:"calculatedPrice"`
	DurationMinutes int                    `json:"durationMinutes"`
	SelectedOptions map[string]interface{} `json:"selectedOptions"`
}

type ProviderResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Photo         *string `json:"photo,omitempty"`
	Rating        float64 `json:"rating"`
	CompletedJobs int     `json:"completedJobs"`
}

func ToOrderResponse(order *models.ServiceOrder) *OrderResponse {
	resp := &OrderResponse{
		ID:          order.ID,
		Code:        order.Code,
		UserID:      order.UserID,
		ProviderID:  order.ProviderID,
		Status:      order.Status,
		Address:     order.Address,
		ServiceDate: order.ServiceDate,
		Frequency:   order.Frequency,
		Subtotal:    order.Subtotal,
		Discount:    order.Discount,
		SurgeFee:    order.SurgeFee,
		PlatformFee: order.PlatformFee,
		Total:       order.Total,
		CouponCode:  order.CouponCode,
		CreatedAt:   order.CreatedAt,
		AcceptedAt:  order.AcceptedAt,
		CompletedAt: order.CompletedAt,
	}

	if len(order.Items) > 0 {
		resp.Items = make([]OrderItemResponse, len(order.Items))
		for i, item := range order.Items {
			resp.Items[i] = OrderItemResponse{
				ID:              item.ID,
				ServiceID:       item.ServiceID,
				ServiceName:     item.ServiceName,
				BasePrice:       item.BasePrice,
				CalculatedPrice: item.CalculatedPrice,
				DurationMinutes: item.DurationMinutes,
				SelectedOptions: item.SelectedOptions,
			}
		}
	}

	if order.Provider != nil && order.Provider.User != nil {
		resp.Provider = &ProviderResponse{
			ID:            order.Provider.ID,
			Name:          order.Provider.User.Name,
			Photo:         order.Provider.Photo,
			Rating:        order.Provider.Rating,
			CompletedJobs: order.Provider.CompletedJobs,
		}
	}

	return resp
}

func ToOrderListResponse(order *models.ServiceOrder) *OrderListResponse {
	return &OrderListResponse{
		ID:          order.ID,
		Code:        order.Code,
		Status:      order.Status,
		Address:     order.Address,
		ServiceDate: order.ServiceDate,
		Total:       order.Total,
		CreatedAt:   order.CreatedAt,
	}
}

func ToOrderListResponses(orders []*models.ServiceOrder) []*OrderListResponse {
	result := make([]*OrderListResponse, len(orders))
	for i, order := range orders {
		result[i] = ToOrderListResponse(order)
	}
	return result
}
