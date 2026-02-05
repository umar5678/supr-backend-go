// internal/modules/promotions/service.go
package promotions

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/umar5678/go-backend/internal/models"
    "github.com/umar5678/go-backend/internal/modules/promotions/dto"
    "github.com/umar5678/go-backend/internal/utils/logger"
    "github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
    CreatePromoCode(ctx context.Context, req dto.CreatePromoCodeRequest) (*dto.PromoCodeResponse, error)
    GetPromoCode(ctx context.Context, code string) (*dto.PromoCodeResponse, error)
    ListPromoCodes(ctx context.Context, isActive bool, page, limit int) ([]*dto.PromoCodeResponse, int64, error)
    ValidatePromoCode(ctx context.Context, userID string, req dto.ValidatePromoCodeRequest) (*dto.ValidatePromoCodeResponse, error)
    ApplyPromoCode(ctx context.Context, userID string, req dto.ApplyPromoCodeRequest) (*dto.ApplyPromoCodeResponse, error)
    DeactivatePromoCode(ctx context.Context, promoID string) error
    AddFreeRideCredit(ctx context.Context, req dto.AddFreeRideCreditRequest) error
    GetFreeRideCredits(ctx context.Context, userID string) (float64, error)
}

type service struct {
    repo Repository
}

func NewService(repo Repository) Service {
    return &service{repo: repo}
}

func (s *service) CreatePromoCode(ctx context.Context, req dto.CreatePromoCodeRequest) (*dto.PromoCodeResponse, error) {
    if err := req.Validate(); err != nil {
        return nil, response.BadRequest(err.Error())
    }

    // Parse dates
    validFrom, err := time.Parse(time.RFC3339, req.ValidFrom)
    if err != nil {
        return nil, response.BadRequest("Invalid validFrom date format")
    }

    validUntil, err := time.Parse(time.RFC3339, req.ValidUntil)
    if err != nil {
        return nil, response.BadRequest("Invalid validUntil date format")
    }

    if validUntil.Before(validFrom) {
        return nil, response.BadRequest("validUntil must be after validFrom")
    }

    // Uppercase code
    code := strings.ToUpper(req.Code)

    promo := &models.PromoCode{
        Code:          code,
        DiscountType:  req.DiscountType,
        DiscountValue: req.DiscountValue,
        MaxDiscount:   req.MaxDiscount,
        MinRideAmount: req.MinRideAmount,
        UsageLimit:    req.UsageLimit,
        PerUserLimit:  req.PerUserLimit,
        ValidFrom:     validFrom,
        ValidUntil:    validUntil,
        IsActive:      true,
        Description:   req.Description,
    }

    if err := s.repo.CreatePromoCode(ctx, promo); err != nil {
        logger.Error("failed to create promo code", "error", err, "code", code)
        return nil, response.InternalServerError("Failed to create promo code", err)
    }

    logger.Info("promo code created", "promoID", promo.ID, "code", code)
    return dto.ToPromoCodeResponse(promo), nil
}

func (s *service) GetPromoCode(ctx context.Context, code string) (*dto.PromoCodeResponse, error) {
    code = strings.ToUpper(code)
    promo, err := s.repo.FindPromoCodeByCode(ctx, code)
    if err != nil {
        return nil, response.NotFoundError("Promo code")
    }

    return dto.ToPromoCodeResponse(promo), nil
}

func (s *service) ListPromoCodes(ctx context.Context, isActive bool, page, limit int) ([]*dto.PromoCodeResponse, int64, error) {
    filters := map[string]interface{}{
        "isActive": isActive,
    }

    promos, total, err := s.repo.ListPromoCodes(ctx, filters, page, limit)
    if err != nil {
        return nil, 0, response.InternalServerError("Failed to list promo codes", err)
    }

    result := make([]*dto.PromoCodeResponse, len(promos))
    for i, promo := range promos {
        result[i] = dto.ToPromoCodeResponse(promo)
    }

    return result, total, nil
}

func (s *service) ValidatePromoCode(ctx context.Context, userID string, req dto.ValidatePromoCodeRequest) (*dto.ValidatePromoCodeResponse, error) {
    code := strings.ToUpper(req.Code)
    promo, err := s.repo.FindPromoCodeByCode(ctx, code)
    if err != nil {
        return &dto.ValidatePromoCodeResponse{
            Valid:   false,
            Message: "Invalid or expired promo code",
        }, nil
    }

    if promo.UsageLimit > 0 && promo.UsageCount >= promo.UsageLimit {
        return &dto.ValidatePromoCodeResponse{
            Valid:   false,
            Message: "Promo code usage limit reached",
        }, nil
    }

    userUsageCount, _ := s.repo.CountUserUsage(ctx, promo.ID, userID)
    if userUsageCount >= int64(promo.PerUserLimit) {
        return &dto.ValidatePromoCodeResponse{
            Valid:   false,
            Message: "You have already used this promo code",
        }, nil
    }

    if req.RideAmount < promo.MinRideAmount {
        return &dto.ValidatePromoCodeResponse{
            Valid:   false,
            Message: fmt.Sprintf("Minimum ride amount of $%.2f required", promo.MinRideAmount),
        }, nil
    }

    discount := s.calculateDiscount(promo, req.RideAmount)
    finalAmount := req.RideAmount - discount

    return &dto.ValidatePromoCodeResponse{
        Valid:          true,
        DiscountAmount: discount,
        FinalAmount:    finalAmount,
        Message:        "Promo code is valid",
    }, nil
}

func (s *service) ApplyPromoCode(ctx context.Context, userID string, req dto.ApplyPromoCodeRequest) (*dto.ApplyPromoCodeResponse, error) {

    validateReq := dto.ValidatePromoCodeRequest{
        Code:       req.Code,
        RideAmount: req.RideAmount,
    }
    validation, err := s.ValidatePromoCode(ctx, userID, validateReq)
    if err != nil {
        return nil, err
    }

    if !validation.Valid {
        return nil, response.BadRequest(validation.Message)
    }

    code := strings.ToUpper(req.Code)
    promo, _ := s.repo.FindPromoCodeByCode(ctx, code)

    usage := &models.PromoCodeUsage{
        PromoCodeID:    promo.ID,
        UserID:         userID,
        RideID:         req.RideID,
        DiscountAmount: validation.DiscountAmount,
    }

    if err := s.repo.CreatePromoUsage(ctx, usage); err != nil {
        return nil, response.InternalServerError("Failed to apply promo code", err)
    }

    s.repo.IncrementUsageCount(ctx, promo.ID)

    logger.Info("promo code applied",
        "promoID", promo.ID,
        "userID", userID,
        "rideID", req.RideID,
        "discount", validation.DiscountAmount,
    )

    return &dto.ApplyPromoCodeResponse{
        RideID:         req.RideID,
        DiscountAmount: validation.DiscountAmount,
        FinalAmount:    validation.FinalAmount,
    }, nil
}

func (s *service) DeactivatePromoCode(ctx context.Context, promoID string) error {
    if err := s.repo.DeactivatePromoCode(ctx, promoID); err != nil {
        return response.InternalServerError("Failed to deactivate promo code", err)
    }

    logger.Info("promo code deactivated", "promoID", promoID)
    return nil
}

func (s *service) AddFreeRideCredit(ctx context.Context, req dto.AddFreeRideCreditRequest) error {
    if err := s.repo.AddFreeRideCredits(ctx, req.UserID, req.Amount); err != nil {
        return response.InternalServerError("Failed to add free ride credits", err)
    }

    logger.Info("free ride credits added",
        "userID", req.UserID,
        "amount", req.Amount,
        "reason", req.Reason,
    )

    return nil
}

func (s *service) GetFreeRideCredits(ctx context.Context, userID string) (float64, error) {
    credits, err := s.repo.GetFreeRideCredits(ctx, userID)
    if err != nil {
        return 0, response.InternalServerError("Failed to get free ride credits", err)
    }

    return credits, nil
}

func (s *service) calculateDiscount(promo *models.PromoCode, rideAmount float64) float64 {
    var discount float64

    switch promo.DiscountType {
    case "percentage":
        discount = rideAmount * (promo.DiscountValue / 100)
        if promo.MaxDiscount > 0 && discount > promo.MaxDiscount {
            discount = promo.MaxDiscount
        }
    case "flat":
        discount = promo.DiscountValue
        if discount > rideAmount {
            discount = rideAmount
        }
    case "free_ride":
        discount = rideAmount
        if promo.MaxDiscount > 0 && discount > promo.MaxDiscount {
            discount = promo.MaxDiscount
        }
    }

    return discount
}
