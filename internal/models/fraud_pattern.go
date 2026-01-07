// internal/models/fraud_pattern.go
package models

import (
	"time"

	"gorm.io/gorm"
)

type FraudPattern struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	PatternType   string         `gorm:"type:varchar(50);not null;index" json:"patternType"` // 'frequent_cancellation', 'same_rider_driver', 'location_gaming', 'fake_trips'
	UserID        *string        `gorm:"type:uuid;index" json:"userId,omitempty"`
	User          *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	DriverID      *string        `gorm:"type:uuid;index" json:"driverId,omitempty"`
	Driver        *User          `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	RelatedUserID *string        `gorm:"type:uuid" json:"relatedUserId,omitempty"`
	RideID        *string        `gorm:"type:uuid" json:"rideId,omitempty"`
	Details       string         `gorm:"type:jsonb" json:"details"`
	RiskScore     int            `gorm:"not null" json:"riskScore"`                        // 0-100
	Status        string         `gorm:"type:varchar(50);default:'flagged'" json:"status"` // 'flagged', 'investigating', 'confirmed', 'dismissed'
	ReviewedBy    *string        `gorm:"type:uuid" json:"reviewedBy,omitempty"`
	ReviewNotes   string         `gorm:"type:text" json:"reviewNotes,omitempty"`
	ReviewedAt    *time.Time     `json:"reviewedAt,omitempty"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (FraudPattern) TableName() string {
	return "fraud_patterns"
}
