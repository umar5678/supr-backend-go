package models

import (
	"time"

	"gorm.io/gorm"
)

type VehicleType struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name          string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"` // economy, comfort, premium, xl, bike
	DisplayName   string         `gorm:"type:varchar(100);not null" json:"displayName"`
	BaseFare      float64        `gorm:"type:decimal(10,2);not null" json:"baseFare"`
	PerKmRate     float64        `gorm:"type:decimal(10,2);not null" json:"perKmRate"`
	PerMinuteRate float64        `gorm:"type:decimal(10,2);not null" json:"perMinuteRate"`
	BookingFee    float64        `gorm:"type:decimal(10,2);not null;default:0.50" json:"bookingFee"`
	Capacity      int            `gorm:"not null" json:"capacity"`
	Description   string         `gorm:"type:text" json:"description"`
	IsActive      bool           `gorm:"default:true" json:"isActive"`
	IconURL       string         `gorm:"type:varchar(255)" json:"iconUrl"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (VehicleType) TableName() string {
	return "vehicle_types"
}
