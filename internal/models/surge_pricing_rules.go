package models

import (
	"time"

	"gorm.io/gorm"
)

type SurgePricingRule struct {
	ID            string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Name          string `gorm:"type:varchar(255);not null" json:"name"`
	Description   string `gorm:"type:text" json:"description"`
	VehicleTypeID string `gorm:"type:uuid;index" json:"vehicleTypeId"`

	DayOfWeek int    `gorm:"type:integer" json:"dayOfWeek"`
	StartTime string `gorm:"type:time" json:"startTime"`   
	EndTime   string `gorm:"type:time" json:"endTime"`     

	BaseMultiplier float64 `gorm:"type:decimal(3,2);not null;default:1.0" json:"baseMultiplier"`
	MinMultiplier  float64 `gorm:"type:decimal(3,2);not null;default:1.0" json:"minMultiplier"`
	MaxMultiplier  float64 `gorm:"type:decimal(3,2);not null;default:2.0" json:"maxMultiplier"`

	EnableDemandBasedSurge     bool    `gorm:"default:false" json:"enableDemandBasedSurge"`
	DemandThreshold            int     `gorm:"type:integer;default:10" json:"demandThreshold"`                   
	DemandMultiplierPerRequest float64 `gorm:"type:decimal(3,2);default:0.05" json:"demandMultiplierPerRequest"` 

	IsActive  bool           `gorm:"default:true" json:"isActive"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SurgePricingRule) TableName() string {
	return "surge_pricing_rules"
}

type DemandTracking struct {
	ID          string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	ZoneID      string `gorm:"type:uuid;index;not null" json:"zoneId"`
	ZoneGeohash string `gorm:"type:varchar(12);index" json:"zoneGeohash"`

	PendingRequests  int `gorm:"type:integer;default:0" json:"pendingRequests"` 
	AvailableDrivers int `gorm:"type:integer;default:0" json:"availableDrivers"`
	CompletedRides   int `gorm:"type:integer;default:0" json:"completedRides"`  
	AverageWaitTime  int `gorm:"type:integer;default:0" json:"averageWaitTime"` 

	DemandSupplyRatio float64 `gorm:"type:decimal(5,2)" json:"demandSupplyRatio"`          
	SurgeMultiplier   float64 `gorm:"type:decimal(3,2);default:1.0" json:"surgeMultiplier"`

	RecordedAt time.Time      `gorm:"index;not null" json:"recordedAt"`
	ExpiresAt  time.Time      `gorm:"index" json:"expiresAt"` 
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DemandTracking) TableName() string {
	return "demand_tracking"
}

type ETAEstimate struct {
	ID     string  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	RideID *string `gorm:"type:uuid;index" json:"rideId,omitempty"`

	PickupLat  float64 `gorm:"type:decimal(10,8)" json:"pickupLat"`
	PickupLon  float64 `gorm:"type:decimal(11,8)" json:"pickupLon"`
	DropoffLat float64 `gorm:"type:decimal(10,8)" json:"dropoffLat"`
	DropoffLon float64 `gorm:"type:decimal(11,8)" json:"dropoffLon"`

	DistanceKm      float64 `gorm:"type:decimal(10,2)" json:"distanceKm"`
	DurationSeconds int     `gorm:"type:integer" json:"durationSeconds"`

	EstimatedPickupETA  int `gorm:"type:integer" json:"estimatedPickupETA"`  
	EstimatedDropoffETA int `gorm:"type:integer" json:"estimatedDropoffETA"` 

	TrafficCondition  string  `gorm:"type:varchar(50);default:'normal'" json:"trafficCondition"` 
	TrafficMultiplier float64 `gorm:"type:decimal(3,2);default:1.0" json:"trafficMultiplier"`    

	Source string `gorm:"type:varchar(50)" json:"source"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ETAEstimate) TableName() string {
	return "eta_estimates"
}

type SurgeHistory struct {
	ID     string `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	RideID string `gorm:"type:uuid;index" json:"rideId"`
	ZoneID string `gorm:"type:uuid;index" json:"zoneId"`

	AppliedMultiplier float64 `gorm:"type:decimal(3,2)" json:"appliedMultiplier"`
	BaseAmount        float64 `gorm:"type:decimal(10,2)" json:"baseAmount"`
	SurgeAmount       float64 `gorm:"type:decimal(10,2)" json:"surgeAmount"`

	Reason string `gorm:"type:varchar(255)" json:"reason"`

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
