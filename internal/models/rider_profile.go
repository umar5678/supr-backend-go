package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Address represents a location with coordinates
type Address struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address"`
}

// Value implements the driver.Valuer interface for JSON encoding
func (a Address) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface for JSON decoding
func (a *Address) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Address")
	}

	return json.Unmarshal(bytes, a)
}

// RiderProfile model
type RiderProfile struct {
	ID                   string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID               string    `gorm:"type:uuid;not null;uniqueIndex" json:"userId"`
	HomeAddress          *Address  `gorm:"type:jsonb" json:"homeAddress,omitempty"`
	WorkAddress          *Address  `gorm:"type:jsonb" json:"workAddress,omitempty"`
	PreferredVehicleType *string   `gorm:"type:varchar(50)" json:"preferredVehicleType,omitempty"`
	Rating               float64   `gorm:"type:decimal(3,2);not null;default:5.0" json:"rating"`
	TotalRides           int       `gorm:"not null;default:0" json:"totalRides"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relations
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Wallet Wallet `gorm:"foreignKey:UserID;references:UserID" json:"wallet,omitempty"`
}

func (RiderProfile) TableName() string {
	return "rider_profiles"
}
