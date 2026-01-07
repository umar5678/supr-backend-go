package models

import (
	"time"

	"gorm.io/gorm"
)

// SurgePricingRule defines time-based surge pricing rules
type SurgePricingRule struct {
	ID            string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name          string `gorm:"type:varchar(255);not null" json:"name"`
	Description   string `gorm:"type:text" json:"description"`
	VehicleTypeID string `gorm:"type:uuid;index" json:"vehicleTypeId"`

	// Time-based surge
	DayOfWeek int    `gorm:"type:integer" json:"dayOfWeek"` // 0 = Sunday, 1 = Monday, etc. -1 = all days
	StartTime string `gorm:"type:time" json:"startTime"`    // HH:MM format
	EndTime   string `gorm:"type:time" json:"endTime"`      // HH:MM format

	// Surge multiplier
	BaseMultiplier float64 `gorm:"type:decimal(3,2);not null;default:1.0" json:"baseMultiplier"`
	MinMultiplier  float64 `gorm:"type:decimal(3,2);not null;default:1.0" json:"minMultiplier"`
	MaxMultiplier  float64 `gorm:"type:decimal(3,2);not null;default:2.0" json:"maxMultiplier"`

	// Demand-based surge
	EnableDemandBasedSurge     bool    `gorm:"default:false" json:"enableDemandBasedSurge"`
	DemandThreshold            int     `gorm:"type:integer;default:10" json:"demandThreshold"`                   // Pending requests threshold
	DemandMultiplierPerRequest float64 `gorm:"type:decimal(3,2);default:0.05" json:"demandMultiplierPerRequest"` // +5% per pending request

	// Status
	IsActive  bool           `gorm:"default:true" json:"isActive"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SurgePricingRule) TableName() string {
	return "surge_pricing_rules"
}

// DemandTracking tracks real-time demand for surge calculation
type DemandTracking struct {
	ID          string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ZoneID      string `gorm:"type:uuid;index;not null" json:"zoneId"`
	ZoneGeohash string `gorm:"type:varchar(12);index" json:"zoneGeohash"`

	// Demand metrics
	PendingRequests  int `gorm:"type:integer;default:0" json:"pendingRequests"`  // Requests waiting for drivers
	AvailableDrivers int `gorm:"type:integer;default:0" json:"availableDrivers"` // Drivers available in zone
	CompletedRides   int `gorm:"type:integer;default:0" json:"completedRides"`   // Rides completed in last hour
	AverageWaitTime  int `gorm:"type:integer;default:0" json:"averageWaitTime"`  // Avg wait in seconds

	// Calculated metrics
	DemandSupplyRatio float64 `gorm:"type:decimal(5,2)" json:"demandSupplyRatio"`           // pending_requests / available_drivers
	SurgeMultiplier   float64 `gorm:"type:decimal(3,2);default:1.0" json:"surgeMultiplier"` // Current surge based on demand

	// Timestamps
	RecordedAt time.Time      `gorm:"index;not null" json:"recordedAt"`
	ExpiresAt  time.Time      `gorm:"index" json:"expiresAt"` // Data validity period
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DemandTracking) TableName() string {
	return "demand_tracking"
}

// ETAEstimate stores ETA calculations for route optimization
type ETAEstimate struct {
	ID     string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	RideID string `gorm:"type:uuid;index" json:"rideId"`

	// Route information
	PickupLat  float64 `gorm:"type:decimal(10,8)" json:"pickupLat"`
	PickupLon  float64 `gorm:"type:decimal(11,8)" json:"pickupLon"`
	DropoffLat float64 `gorm:"type:decimal(10,8)" json:"dropoffLat"`
	DropoffLon float64 `gorm:"type:decimal(11,8)" json:"dropoffLon"`

	// Distance and duration
	DistanceKm      float64 `gorm:"type:decimal(10,2)" json:"distanceKm"`
	DurationSeconds int     `gorm:"type:integer" json:"durationSeconds"`

	// ETA (in seconds from now)
	EstimatedPickupETA  int `gorm:"type:integer" json:"estimatedPickupETA"`  // How long till driver reaches pickup
	EstimatedDropoffETA int `gorm:"type:integer" json:"estimatedDropoffETA"` // How long till reaching dropoff

	// Traffic conditions
	TrafficCondition  string  `gorm:"type:varchar(50);default:'normal'" json:"trafficCondition"` // normal, slow, heavy
	TrafficMultiplier float64 `gorm:"type:decimal(3,2);default:1.0" json:"trafficMultiplier"`    // Time multiplier due to traffic

	// Source
	Source string `gorm:"type:varchar(50)" json:"source"` // 'google_maps', 'osrm', 'calculated'

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ETAEstimate) TableName() string {
	return "eta_estimates"
}

// SurgeHistory tracks surge pricing history for analytics
type SurgeHistory struct {
	ID     string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	RideID string `gorm:"type:uuid;index" json:"rideId"`
	ZoneID string `gorm:"type:uuid;index" json:"zoneId"`

	// Surge details
	AppliedMultiplier float64 `gorm:"type:decimal(3,2)" json:"appliedMultiplier"`
	BaseAmount        float64 `gorm:"type:decimal(10,2)" json:"baseAmount"`
	SurgeAmount       float64 `gorm:"type:decimal(10,2)" json:"surgeAmount"`

	// Reason for surge
	Reason string `gorm:"type:varchar(255)" json:"reason"` // 'time_based', 'demand_based', 'combined'

	// Contributing factors
	TimeBasedMultiplier   float64 `gorm:"type:decimal(3,2)" json:"timeBasedMultiplier"`
	DemandBasedMultiplier float64 `gorm:"type:decimal(3,2)" json:"demandBasedMultiplier"`
	PendingRequests       int     `gorm:"type:integer" json:"pendingRequests"`
	AvailableDrivers      int     `gorm:"type:integer" json:"availableDrivers"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SurgeHistory) TableName() string {
	return "surge_history"
}
