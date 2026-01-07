// internal/models/promo_code.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type PromoCode struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Code          string         `gorm:"type:varchar(50);not null;uniqueIndex" json:"code"`
	DiscountType  string         `gorm:"type:varchar(20);not null" json:"discountType"` // 'percentage', 'flat', 'free_ride'
	DiscountValue float64        `gorm:"type:decimal(10,2);not null" json:"discountValue"`
	MaxDiscount   float64        `gorm:"type:decimal(10,2)" json:"maxDiscount,omitempty"`
	MinRideAmount float64        `gorm:"type:decimal(10,2);default:0" json:"minRideAmount"`
	UsageLimit    int            `gorm:"default:0" json:"usageLimit"` // 0 = unlimited
	UsageCount    int            `gorm:"default:0" json:"usageCount"`
	PerUserLimit  int            `gorm:"default:1" json:"perUserLimit"`
	ValidFrom     time.Time      `gorm:"not null" json:"validFrom"`
	ValidUntil    time.Time      `gorm:"not null" json:"validUntil"`
	IsActive      bool           `gorm:"default:true" json:"isActive"`
	Description   string         `gorm:"type:text" json:"description,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PromoCode) TableName() string {
	return "promo_codes"
}
