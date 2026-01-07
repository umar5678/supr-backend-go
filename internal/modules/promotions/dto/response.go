// internal/modules/promotions/dto/response.go
package dto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
)

type PromoCodeResponse struct {
	ID            string    `json:"id"`
	Code          string    `json:"code"`
	DiscountType  string    `json:"discountType"`
	DiscountValue float64   `json:"discountValue"`
	MaxDiscount   float64   `json:"maxDiscount,omitempty"`
	MinRideAmount float64   `json:"minRideAmount"`
	UsageLimit    int       `json:"usageLimit"`
	UsageCount    int       `json:"usageCount"`
	PerUserLimit  int       `json:"perUserLimit"`
	ValidFrom     time.Time `json:"validFrom"`
	ValidUntil    time.Time `json:"validUntil"`
	IsActive      bool      `json:"isActive"`
	Description   string    `json:"description,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

type ValidatePromoCodeResponse struct {
	Valid          bool    `json:"valid"`
	DiscountAmount float64 `json:"discountAmount"`
	FinalAmount    float64 `json:"finalAmount"`
	Message        string  `json:"message,omitempty"`
}

type ApplyPromoCodeResponse struct {
	RideID         string  `json:"rideId"`
	DiscountAmount float64 `json:"discountAmount"`
	FinalAmount    float64 `json:"finalAmount"`
}

func ToPromoCodeResponse(promo *models.PromoCode) *PromoCodeResponse {
	return &PromoCodeResponse{
		ID:            promo.ID,
		Code:          promo.Code,
		DiscountType:  promo.DiscountType,
		DiscountValue: promo.DiscountValue,
		MaxDiscount:   promo.MaxDiscount,
		MinRideAmount: promo.MinRideAmount,
		UsageLimit:    promo.UsageLimit,
		UsageCount:    promo.UsageCount,
		PerUserLimit:  promo.PerUserLimit,
		ValidFrom:     promo.ValidFrom,
		ValidUntil:    promo.ValidUntil,
		IsActive:      promo.IsActive,
		Description:   promo.Description,
		CreatedAt:     promo.CreatedAt,
	}
}
