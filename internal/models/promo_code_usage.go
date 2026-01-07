// internal/models/promo_code_usage.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type PromoCodeUsage struct {
	ID             string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	PromoCodeID    string         `gorm:"type:uuid;not null;index" json:"promoCodeId"`
	PromoCode      PromoCode      `gorm:"foreignKey:PromoCodeID" json:"promoCode,omitempty"`
	UserID         string         `gorm:"type:uuid;not null;index" json:"userId"`
	User           User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	RideID         string         `gorm:"type:uuid;not null" json:"rideId"`
	Ride           Ride           `gorm:"foreignKey:RideID" json:"ride,omitempty"`
	DiscountAmount float64        `gorm:"type:decimal(10,2);not null" json:"discountAmount"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PromoCodeUsage) TableName() string {
	return "promo_code_usage"
}
