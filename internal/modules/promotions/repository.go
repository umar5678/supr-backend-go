package promotions

import (
	"context"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	CreatePromoCode(ctx context.Context, promo *models.PromoCode) error
	FindPromoCodeByCode(ctx context.Context, code string) (*models.PromoCode, error)
	FindPromoCodeByID(ctx context.Context, id string) (*models.PromoCode, error)
	ListPromoCodes(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.PromoCode, int64, error)
	UpdatePromoCode(ctx context.Context, promo *models.PromoCode) error
	IncrementUsageCount(ctx context.Context, promoID string) error
	DeactivatePromoCode(ctx context.Context, promoID string) error

	CreatePromoUsage(ctx context.Context, usage *models.PromoCodeUsage) error
	CountUserUsage(ctx context.Context, promoID, userID string) (int64, error)

	GetFreeRideCredits(ctx context.Context, userID string) (float64, error)
	DeductFreeRideCredits(ctx context.Context, userID string, amount float64) error
	AddFreeRideCredits(ctx context.Context, userID string, amount float64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreatePromoCode(ctx context.Context, promo *models.PromoCode) error {
	return r.db.WithContext(ctx).Create(promo).Error
}

func (r *repository) FindPromoCodeByCode(ctx context.Context, code string) (*models.PromoCode, error) {
	var promo models.PromoCode
	err := r.db.WithContext(ctx).
		Where("code = ? AND is_active = true AND valid_from <= ? AND valid_until >= ?",
			code, time.Now(), time.Now()).
		First(&promo).Error
	return &promo, err
}

func (r *repository) FindPromoCodeByID(ctx context.Context, id string) (*models.PromoCode, error) {
	var promo models.PromoCode
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&promo).Error
	return &promo, err
}

func (r *repository) ListPromoCodes(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.PromoCode, int64, error) {
	var promos []*models.PromoCode
	var total int64

	query := r.db.WithContext(ctx).Model(&models.PromoCode{})

	if isActive, ok := filters["isActive"].(bool); ok {
		query = query.Where("is_active = ?", isActive)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&promos).Error

	return promos, total, err
}

func (r *repository) UpdatePromoCode(ctx context.Context, promo *models.PromoCode) error {
	return r.db.WithContext(ctx).Save(promo).Error
}

func (r *repository) IncrementUsageCount(ctx context.Context, promoID string) error {
	return r.db.WithContext(ctx).
		Model(&models.PromoCode{}).
		Where("id = ?", promoID).
		Update("usage_count", gorm.Expr("usage_count + 1")).Error
}

func (r *repository) DeactivatePromoCode(ctx context.Context, promoID string) error {
	return r.db.WithContext(ctx).
		Model(&models.PromoCode{}).
		Where("id = ?", promoID).
		Update("is_active", false).Error
}

func (r *repository) CreatePromoUsage(ctx context.Context, usage *models.PromoCodeUsage) error {
	return r.db.WithContext(ctx).Create(usage).Error
}

func (r *repository) CountUserUsage(ctx context.Context, promoID, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.PromoCodeUsage{}).
		Where("promo_code_id = ? AND user_id = ?", promoID, userID).
		Count(&count).Error
	return count, err
}

func (r *repository) GetFreeRideCredits(ctx context.Context, userID string) (float64, error) {
	var wallet models.Wallet
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&wallet).Error
	if err != nil {
		return 0, err
	}
	return wallet.FreeRideCredits, nil
}

func (r *repository) DeductFreeRideCredits(ctx context.Context, userID string, amount float64) error {
	return r.db.WithContext(ctx).
		Model(&models.Wallet{}).
		Where("user_id = ? AND free_ride_credits >= ?", userID, amount).
		Update("free_ride_credits", gorm.Expr("free_ride_credits - ?", amount)).Error
}

func (r *repository) AddFreeRideCredits(ctx context.Context, userID string, amount float64) error {
	return r.db.WithContext(ctx).
		Model(&models.Wallet{}).
		Where("user_id = ?", userID).
		Update("free_ride_credits", gorm.Expr("free_ride_credits + ?", amount)).Error
}
