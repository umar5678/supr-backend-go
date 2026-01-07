package dto

import (
	"time"
)

// ===== SURGE PRICING RULES =====

// CreateSurgePricingRuleRequest creates a new surge pricing rule
type CreateSurgePricingRuleRequest struct {
	Name                       string  `json:"name" binding:"required"`
	Description                string  `json:"description"`
	VehicleTypeID              string  `json:"vehicleTypeId"`
	DayOfWeek                  int     `json:"dayOfWeek" binding:"min:-1,max:6"` // -1 = all days
	StartTime                  string  `json:"startTime" binding:"required"`     // HH:MM
	EndTime                    string  `json:"endTime" binding:"required"`       // HH:MM
	BaseMultiplier             float64 `json:"baseMultiplier" binding:"required,min:0.5,max:5.0"`
	MinMultiplier              float64 `json:"minMultiplier" binding:"min:0.5,max:5.0"`
	MaxMultiplier              float64 `json:"maxMultiplier" binding:"min:0.5,max:5.0"`
	EnableDemandBasedSurge     bool    `json:"enableDemandBasedSurge"`
	DemandThreshold            int     `json:"demandThreshold" binding:"min:1"`
	DemandMultiplierPerRequest float64 `json:"demandMultiplierPerRequest"`
}

// SurgePricingRuleResponse returns surge pricing rule info
type SurgePricingRuleResponse struct {
	ID                         string    `json:"id"`
	Name                       string    `json:"name"`
	Description                string    `json:"description"`
	VehicleTypeID              string    `json:"vehicleTypeId,omitempty"`
	DayOfWeek                  int       `json:"dayOfWeek"`
	StartTime                  string    `json:"startTime"`
	EndTime                    string    `json:"endTime"`
	BaseMultiplier             float64   `json:"baseMultiplier"`
	MinMultiplier              float64   `json:"minMultiplier"`
	MaxMultiplier              float64   `json:"maxMultiplier"`
	EnableDemandBasedSurge     bool      `json:"enableDemandBasedSurge"`
	DemandThreshold            int       `json:"demandThreshold"`
	DemandMultiplierPerRequest float64   `json:"demandMultiplierPerRequest"`
	IsActive                   bool      `json:"isActive"`
	CreatedAt                  time.Time `json:"createdAt"`
	UpdatedAt                  time.Time `json:"updatedAt"`
}

// ===== DEMAND TRACKING =====

// DemandTrackingResponse returns current demand info
type DemandTrackingResponse struct {
	ID                string    `json:"id"`
	ZoneID            string    `json:"zoneId"`
	ZoneGeohash       string    `json:"zoneGeohash"`
	PendingRequests   int       `json:"pendingRequests"`
	AvailableDrivers  int       `json:"availableDrivers"`
	CompletedRides    int       `json:"completedRides"`
	AverageWaitTime   int       `json:"averageWaitTime"`
	DemandSupplyRatio float64   `json:"demandSupplyRatio"`
	SurgeMultiplier   float64   `json:"surgeMultiplier"`
	RecordedAt        time.Time `json:"recordedAt"`
}

// ===== ETA ESTIMATES =====

// ETAEstimateRequest requests ETA calculation
type ETAEstimateRequest struct {
	PickupLat  float64 `json:"pickupLat" binding:"required"`
	PickupLon  float64 `json:"pickupLon" binding:"required"`
	DropoffLat float64 `json:"dropoffLat" binding:"required"`
	DropoffLon float64 `json:"dropoffLon" binding:"required"`
}

// ETAEstimateResponse returns ETA details
type ETAEstimateResponse struct {
	ID                  string    `json:"id"`
	RideID              string    `json:"rideId,omitempty"`
	DistanceKm          float64   `json:"distanceKm"`
	DurationSeconds     int       `json:"durationSeconds"`
	EstimatedPickupETA  int       `json:"estimatedPickupETA"`
	EstimatedDropoffETA int       `json:"estimatedDropoffETA"`
	TrafficCondition    string    `json:"trafficCondition"`
	TrafficMultiplier   float64   `json:"trafficMultiplier"`
	Source              string    `json:"source"`
	CreatedAt           time.Time `json:"createdAt"`
}

// ===== SURGE CALCULATION =====

// SurgeCalculationRequest requests surge calculation for a location
type SurgeCalculationRequest struct {
	PickupLat     float64 `json:"pickupLat" binding:"required"`
	PickupLon     float64 `json:"pickupLon" binding:"required"`
	VehicleTypeID string  `json:"vehicleTypeId" binding:"required"`
	Geohash       string  `json:"geohash"`
}

// SurgeCalculationResponse returns surge details
type SurgeCalculationResponse struct {
	AppliedMultiplier     float64      `json:"appliedMultiplier"`
	TimeBasedMultiplier   float64      `json:"timeBasedMultiplier"`
	DemandBasedMultiplier float64      `json:"demandBasedMultiplier"`
	Reason                string       `json:"reason"` // 'normal', 'time_based', 'demand_based', 'combined'
	BaseFare              float64      `json:"baseFare"`
	SurgeAmount           float64      `json:"surgeAmount"`
	TotalFare             float64      `json:"totalFare"`
	Details               SurgeDetails `json:"details"`
}

// SurgeDetails provides breakdown of surge factors
type SurgeDetails struct {
	TimeOfDay         string  `json:"timeOfDay"` // "peak" or "off-peak"
	DayType           string  `json:"dayType"`   // "weekday" or "weekend"
	PendingRequests   int     `json:"pendingRequests"`
	AvailableDrivers  int     `json:"availableDrivers"`
	DemandSupplyRatio float64 `json:"demandSupplyRatio"`
	TrafficCondition  string  `json:"trafficCondition"`
	ActivePricingRule string  `json:"activePricingRule,omitempty"`
}

// ===== SURGE HISTORY =====

// SurgeHistoryResponse returns surge pricing history for audit
type SurgeHistoryResponse struct {
	ID                    string    `json:"id"`
	RideID                string    `json:"rideId"`
	AppliedMultiplier     float64   `json:"appliedMultiplier"`
	BaseAmount            float64   `json:"baseAmount"`
	SurgeAmount           float64   `json:"surgeAmount"`
	Reason                string    `json:"reason"`
	TimeBasedMultiplier   float64   `json:"timeBasedMultiplier"`
	DemandBasedMultiplier float64   `json:"demandBasedMultiplier"`
	PendingRequests       int       `json:"pendingRequests"`
	AvailableDrivers      int       `json:"availableDrivers"`
	CreatedAt             time.Time `json:"createdAt"`
}

// ===== FARE ESTIMATE WITH SURGE =====

// EnhancedFareEstimateResponse returns detailed fare with all surges
type EnhancedFareEstimateResponse struct {
	*FareEstimateResponse
	// Additional surge details
	TimeBasedSurge   float64                  `json:"timeBasedSurge"`
	DemandBasedSurge float64                  `json:"demandBasedSurge"`
	SurgeReason      string                   `json:"surgeReason"`
	ETA              *ETAEstimateResponse     `json:"eta,omitempty"`
	SurgeCalculation SurgeCalculationResponse `json:"surgeCalculation,omitempty"`
}
