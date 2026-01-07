package models

import (
    "time"
    "gorm.io/gorm"
)

type PriceCappingRule struct {
    ID                  string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    VehicleTypeID       string         `gorm:"type:uuid;not null;index" json:"vehicleTypeId"`
    VehicleType         VehicleType    `gorm:"foreignKey:VehicleTypeID" json:"vehicleType,omitempty"`
    MaxCustomerPrice    float64        `gorm:"type:decimal(10,2);not null" json:"maxCustomerPrice"`
    MaxDriverEarning    float64        `gorm:"type:decimal(10,2);not null" json:"maxDriverEarning"`
    PlatformAbsorbsDiff bool           `gorm:"default:true" json:"platformAbsorbsDiff"`
    IsActive            bool           `gorm:"default:true" json:"isActive"`
    CreatedAt           time.Time      `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt           time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
    DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PriceCappingRule) TableName() string {
    return "price_capping_rules"
}
