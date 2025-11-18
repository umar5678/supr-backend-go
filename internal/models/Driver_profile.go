package models

import (
	"time"

	"gorm.io/gorm"
)

type DriverProfile struct {
	ID               string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID           string         `gorm:"type:uuid;uniqueIndex;not null" json:"userId"`
	LicenseNumber    string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"licenseNumber"`
	Status           string         `gorm:"type:varchar(50);default:'offline'" json:"status"` // offline, online, busy, on_trip
	CurrentLocation  *string        `gorm:"type:geometry(Point,4326)" json:"currentLocation"`
	Heading          int            `gorm:"default:0" json:"heading"` // 0-360 degrees
	Rating           float64        `gorm:"type:decimal(3,2);default:5.0" json:"rating"`
	TotalTrips       int            `gorm:"default:0" json:"totalTrips"`
	TotalEarnings    float64        `gorm:"type:decimal(10,2);default:0" json:"totalEarnings"`
	AcceptanceRate   float64        `gorm:"type:decimal(5,2);default:100.0" json:"acceptanceRate"`
	CancellationRate float64        `gorm:"type:decimal(5,2);default:0.0" json:"cancellationRate"`
	IsVerified       bool           `gorm:"default:true" json:"isVerified"`
	WalletBalance    float64        `gorm:"type:decimal(10,2);default:0" json:"walletBalance"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User    User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Vehicle *Vehicle `gorm:"foreignKey:DriverID" json:"vehicle,omitempty"`
}

func (DriverProfile) TableName() string {
	return "driver_profiles"
}

type Vehicle struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	DriverID      string         `gorm:"type:uuid;uniqueIndex;not null" json:"driverId"`
	VehicleTypeID string         `gorm:"type:uuid;not null" json:"vehicleTypeId"`
	Make          string         `gorm:"type:varchar(100);not null" json:"make"`
	Model         string         `gorm:"type:varchar(100);not null" json:"model"`
	Year          int            `gorm:"not null" json:"year"`
	Color         string         `gorm:"type:varchar(50);not null" json:"color"`
	LicensePlate  string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"licensePlate"`
	Capacity      int            `gorm:"not null" json:"capacity"`
	IsActive      bool           `gorm:"default:true" json:"isActive"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	VehicleType VehicleType `gorm:"foreignKey:VehicleTypeID" json:"vehicleType,omitempty"`
}

func (Vehicle) TableName() string {
	return "vehicles"
}
