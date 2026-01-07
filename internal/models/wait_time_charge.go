package models

import (
    "time"
    "gorm.io/gorm"
)

type WaitTimeCharge struct {
    ID               string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    RideID           string         `gorm:"type:uuid;not null;uniqueIndex" json:"rideId"`
    Ride             Ride           `gorm:"foreignKey:RideID" json:"ride,omitempty"`
    WaitStartedAt    time.Time      `gorm:"not null" json:"waitStartedAt"`
    WaitEndedAt      *time.Time     `json:"waitEndedAt,omitempty"`
    TotalWaitMinutes int            `gorm:"default:0" json:"totalWaitMinutes"`
    ChargeAmount     float64        `gorm:"type:decimal(10,2);default:0" json:"chargeAmount"`
    CreatedAt        time.Time      `gorm:"autoCreateTime" json:"createdAt"`
    DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WaitTimeCharge) TableName() string {
    return "wait_time_charges"
}
