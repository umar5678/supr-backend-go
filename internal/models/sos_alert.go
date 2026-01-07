package models

import (
    "time"
    "gorm.io/gorm"
)

type SOSAlert struct {
    ID                      string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
    UserID                  string         `gorm:"type:uuid;not null;index" json:"userId"`
    User                    User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
    RideID                  *string        `gorm:"type:uuid;index" json:"rideId,omitempty"`
    Ride                    *Ride          `gorm:"foreignKey:RideID" json:"ride,omitempty"`
    AlertType               string         `gorm:"type:varchar(50);not null" json:"alertType"` // 'manual', 'auto_triggered'
    Latitude                float64        `gorm:"type:decimal(10,8);not null" json:"latitude"`
    Longitude               float64        `gorm:"type:decimal(11,8);not null" json:"longitude"`
    Status                  string         `gorm:"type:varchar(50);default:'active'" json:"status"` // 'active', 'resolved', 'cancelled'
    EmergencyContactsNotified bool         `gorm:"default:false" json:"emergencyContactsNotified"`
    Severity                string         `gorm:"type:varchar(50);default:'low'" json:"severity"` // 'low', 'medium', 'high'
    TriggeredAt             time.Time      `gorm:"not null" json:"triggeredAt"`
    SafetyTeamNotifiedAt    *time.Time     `json:"safetyTeamNotifiedAt,omitempty"`
    ResolvedAt              *time.Time     `json:"resolvedAt,omitempty"`
    ResolvedBy              *string        `gorm:"type:uuid" json:"resolvedBy,omitempty"`
    Notes                   string         `gorm:"type:text" json:"notes,omitempty"`
    CreatedAt               time.Time      `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt               time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
    DeletedAt               gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SOSAlert) TableName() string {
    return "sos_alerts"
}
