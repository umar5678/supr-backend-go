// internal/modules/promotions/dto/request.go
package dto

import "errors"

type CreatePromoCodeRequest struct {
    Code          string  `json:"code" binding:"required,min=3,max=50"`
    DiscountType  string  `json:"discountType" binding:"required,oneof=percentage flat free_ride"`
    DiscountValue float64 `json:"discountValue" binding:"required,min=0"`
    MaxDiscount   float64 `json:"maxDiscount" binding:"omitempty,min=0"`
    MinRideAmount float64 `json:"minRideAmount" binding:"omitempty,min=0"`
    UsageLimit    int     `json:"usageLimit" binding:"omitempty,min=0"`
    PerUserLimit  int     `json:"perUserLimit" binding:"omitempty,min=1"`
    ValidFrom     string  `json:"validFrom" binding:"required"`
    ValidUntil    string  `json:"validUntil" binding:"required"`
    Description   string  `json:"description" binding:"omitempty,max=500"`
}

func (r *CreatePromoCodeRequest) Validate() error {
    if r.Code == "" {
        return errors.New("code is required")
    }
    if r.DiscountType == "percentage" && r.DiscountValue > 100 {
        return errors.New("percentage discount cannot exceed 100")
    }
    return nil
}

type ValidatePromoCodeRequest struct {
    Code       string  `json:"code" binding:"required"`
    RideAmount float64 `json:"rideAmount" binding:"required,min=0"`
}

type ApplyPromoCodeRequest struct {
    Code       string  `json:"code" binding:"required"`
    RideID     string  `json:"rideId" binding:"required,uuid"`
    RideAmount float64 `json:"rideAmount" binding:"required,min=0"`
}

type AddFreeRideCreditRequest struct {
    UserID string  `json:"userId" binding:"required,uuid"`
    Amount float64 `json:"amount" binding:"required,min=0"`
    Reason string  `json:"reason" binding:"required,max=200"`
}
