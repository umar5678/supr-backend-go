package models

import (
	"time"

	"gorm.io/gorm"
)

type SurgePricingZone struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	AreaName    string         `gorm:"type:varchar(255);not null" json:"areaName"`
	AreaGeohash string         `gorm:"type:varchar(12);index;not null" json:"areaGeohash"`
	CenterLat   float64        `gorm:"type:decimal(10,8);not null" json:"centerLat"`
	CenterLon   float64        `gorm:"type:decimal(11,8);not null" json:"centerLon"`
	RadiusKm    float64        `gorm:"type:decimal(6,2);not null" json:"radiusKm"`
	Multiplier  float64        `gorm:"type:decimal(3,2);not null;default:1.0" json:"multiplier"`
	ActiveFrom  time.Time      `gorm:"not null" json:"activeFrom"`
	ActiveUntil time.Time      `gorm:"not null" json:"activeUntil"`
	IsActive    bool           `gorm:"default:true" json:"isActive"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SurgePricingZone) TableName() string {
	return "surge_pricing_zones"
}

type FareEstimate struct {
	BaseFare          float64 `json:"baseFare"`
	DistanceFare      float64 `json:"distanceFare"`
	DurationFare      float64 `json:"durationFare"`
	BookingFee        float64 `json:"bookingFee"`
	SurgeMultiplier   float64 `json:"surgeMultiplier"`
	SubTotal          float64 `json:"subTotal"`
	SurgeAmount       float64 `json:"surgeAmount"`
	TotalFare         float64 `json:"totalFare"`
	EstimatedDistance float64 `json:"estimatedDistance"` // km
	EstimatedDuration int     `json:"estimatedDuration"` // seconds
	VehicleTypeName   string  `json:"vehicleTypeName"`
}
