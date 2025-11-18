package models

import (
	"time"
)

type DriverLocation struct {
	ID        string    `gorm:"primaryKey;autoIncrement" json:"id"`
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

func (DriverLocation) TableName() string {
	return "driver_locations"
}

// Real-time location broadcast message
type LocationBroadcast struct {
	DriverID  string    `json:"driverId"`
	RideID    string    `json:"rideId"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Heading   int       `json:"heading"`
	Speed     float64   `json:"speed"`
	Timestamp time.Time `json:"timestamp"`
	ETA       int       `json:"eta"` // Estimated seconds to destination
}

// Route information for polyline
type Route struct {
	Polyline    string    `json:"polyline"` // Encoded polyline points
	Distance    float64   `json:"distance"` // meters
	Duration    int       `json:"duration"` // seconds
	Summary     string    `json:"summary"`  // Route summary
	BoundingBox []float64 `json:"bbox"`     // [minLon, minLat, maxLon, maxLat]
}
