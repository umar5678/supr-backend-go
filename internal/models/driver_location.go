package models

import (
	"time"
)

type DriverLocationHistory struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	DriverID  string    `gorm:"type:uuid;not null;index" json:"driverId"`
	Location  string    `gorm:"type:geometry(Point,4326);not null" json:"location"`
	Latitude  float64   `gorm:"type:decimal(10,8);not null" json:"latitude"`
	Longitude float64   `gorm:"type:decimal(11,8);not null" json:"longitude"`
	Heading   int       `gorm:"default:0" json:"heading"`
	Speed     float64   `gorm:"type:decimal(6,2);default:0" json:"speed"`    // km/h
	Accuracy  float64   `gorm:"type:decimal(6,2);default:0" json:"accuracy"` // meters
	Timestamp time.Time `gorm:"not null;index" json:"timestamp"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`

	// Relations
	Driver DriverProfile `gorm:"foreignKey:DriverID;references:ID" json:"driver,omitempty"`
}

func (DriverLocationHistory) TableName() string {
	return "driver_locations_history"
}
