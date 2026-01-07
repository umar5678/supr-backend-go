package models

import (
	"time"

	"gorm.io/gorm"
)

type Ride struct {
	ID            string  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	RiderID       string  `gorm:"type:uuid;not null;index" json:"riderId"`
	DriverID      *string `gorm:"type:uuid;index" json:"driverId"`
	VehicleTypeID string  `gorm:"type:uuid;not null" json:"vehicleTypeId"`
	Status        string  `gorm:"type:varchar(50);not null;index" json:"status"` // searching, accepted, arrived, started, completed, cancelled

	// Locations
	PickupLocation string  `gorm:"type:geometry(Point,4326);not null" json:"pickupLocation"`
	PickupLat      float64 `gorm:"type:decimal(10,8);not null" json:"pickupLat"`
	PickupLon      float64 `gorm:"type:decimal(11,8);not null" json:"pickupLon"`
	PickupAddress  string  `gorm:"type:text" json:"pickupAddress"`

	DropoffLocation string  `gorm:"type:geometry(Point,4326);not null" json:"dropoffLocation"`
	DropoffLat      float64 `gorm:"type:decimal(10,8);not null" json:"dropoffLat"`
	DropoffLon      float64 `gorm:"type:decimal(11,8);not null" json:"dropoffLon"`
	DropoffAddress  string  `gorm:"type:text" json:"dropoffAddress"`

	// Estimates
	EstimatedDistance float64 `gorm:"type:decimal(10,2)" json:"estimatedDistance"` // km
	EstimatedDuration int     `json:"estimatedDuration"`                           // seconds
	EstimatedFare     float64 `gorm:"type:decimal(10,2)" json:"estimatedFare"`

	// Actuals
	ActualDistance *float64 `gorm:"type:decimal(10,2)" json:"actualDistance"` // km
	ActualDuration *int     `json:"actualDuration"`                           // seconds
	ActualFare     *float64 `gorm:"type:decimal(10,2)" json:"actualFare"`

	// Pricing
	SurgeMultiplier         float64  `gorm:"type:decimal(3,2);default:1.0" json:"surgeMultiplier"`
	WaitTimeCharge          *float64 `gorm:"type:decimal(10,2)" json:"waitTimeCharge"`
	PromoDiscount           *float64 `gorm:"type:decimal(10,2)" json:"promoDiscount"`
	PromoCodeID             *string  `gorm:"type:uuid" json:"promoCodeId"`
	PromoCode               *string  `gorm:"type:varchar(50)" json:"promoCode"`
	DestinationChangeCharge *float64 `gorm:"type:decimal(10,2)" json:"destinationChangeCharge"`

	// Rating
	DriverRating *float64 `gorm:"type:decimal(2,1)" json:"driverRating"`
	RiderRating  *float64 `gorm:"type:decimal(2,1)" json:"riderRating"`

	// Wallet
	WalletHoldID *string `gorm:"type:uuid" json:"walletHoldId"`

	// Notes
	RiderNotes         string  `gorm:"type:text" json:"riderNotes"`
	CancellationReason string  `gorm:"type:text" json:"cancellationReason"`
	CancelledBy        *string `gorm:"type:varchar(50)" json:"cancelledBy"` // rider, driver, system

	// Timestamps
	RequestedAt time.Time  `gorm:"not null" json:"requestedAt"`
	AcceptedAt  *time.Time `json:"acceptedAt"`
	ArrivedAt   *time.Time `json:"arrivedAt"`
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
	CancelledAt *time.Time `json:"cancelledAt"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Rider       User        `gorm:"foreignKey:RiderID" json:"rider,omitempty"`
	Driver      *User       `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	VehicleType VehicleType `gorm:"foreignKey:VehicleTypeID" json:"vehicleType,omitempty"`
}

func (Ride) TableName() string {
	return "rides"
}

// Add to ride.go models
type RideRequest struct {
	ID              string     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	RideID          string     `gorm:"type:uuid;not null;index" json:"rideId"`
	DriverID        string     `gorm:"type:uuid;not null;index" json:"driverId"`
	Status          string     `gorm:"type:varchar(50);not null" json:"status"` // pending, accepted, rejected, expired, cancelled
	SentAt          time.Time  `gorm:"not null" json:"sentAt"`
	RespondedAt     *time.Time `json:"respondedAt,omitempty"`
	ExpiresAt       time.Time  `gorm:"not null;index" json:"expiresAt"`
	RejectionReason string     `gorm:"type:text" json:"rejectionReason,omitempty"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`

	// Relations
	Ride   Ride          `gorm:"foreignKey:RideID" json:"ride,omitempty"`
	Driver DriverProfile `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
}

func (RideRequest) TableName() string {
	return "ride_requests"
}
