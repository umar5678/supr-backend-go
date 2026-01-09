// internal/modules/pricing/dto/response.go
package dto

type FareEstimateResponse struct {
	BaseFare           float64               `json:"baseFare"`
	DistanceFare       float64               `json:"distanceFare"`
	DurationFare       float64               `json:"durationFare"`
	BookingFee         float64               `json:"bookingFee"`
	SurgeMultiplier    float64               `json:"surgeMultiplier"`
	SubTotal           float64               `json:"subTotal"`
	SurgeAmount        float64               `json:"surgeAmount"`
	TotalFare          float64               `json:"totalFare"`
	DriverPayout       float64               `json:"driverPayout"`       // What driver receives (after commission)
	PlatformCommission float64               `json:"platformCommission"` // Platform fee
	CommissionRate     float64               `json:"commissionRate"`     // Percentage (e.g., 5.0 for 5%)
	EstimatedDistance  float64               `json:"estimatedDistance"`  // km
	EstimatedDuration  int                   `json:"estimatedDuration"`  // seconds
	VehicleTypeName    string                `json:"vehicleTypeName"`
	Currency           string                `json:"currency"`
	SurgeDetails       *SurgeDetailsResponse `json:"surgeDetails,omitempty"`
}

type SurgeZoneResponse struct {
	ID         string  `json:"id"`
	AreaName   string  `json:"areaName"`
	Multiplier float64 `json:"multiplier"`
	RadiusKm   float64 `json:"radiusKm"`
	IsActive   bool    `json:"isActive"`
}

type FareBreakdownResponse struct {
	Components        []FareComponent `json:"components"`
	BaseFare          float64         `json:"baseFare"`
	DistanceCharge    float64         `json:"distanceCharge"`
	TimeCharge        float64         `json:"timeCharge"`
	BookingFee        float64         `json:"bookingFee"`
	SurgeCharge       float64         `json:"surgeCharge"`
	SurgeMultiplier   float64         `json:"surgeMultiplier"`
	SubTotal          float64         `json:"subTotal"`
	TotalFare         float64         `json:"totalFare"`
	CustomerPrice     float64         `json:"customerPrice"`
	DriverEarning     float64         `json:"driverEarning"`
	PlatformFee       float64         `json:"platformFee"`
	EstimatedDistance float64         `json:"estimatedDistance"`
	EstimatedDuration int             `json:"estimatedDuration"`
	PriceCapped       bool            `json:"priceCapped"`
	PlatformAbsorbed  float64         `json:"platformAbsorbed"`
}

type WaitTimeChargeResponse struct {
	RideID           string  `json:"rideId"`
	TotalWaitMinutes int     `json:"totalWaitMinutes"`
	ChargeAmount     float64 `json:"chargeAmount"`
	FreeWaitMinutes  int     `json:"freeWaitMinutes"`
}

type DestinationChangeResponse struct {
	RideID             string  `json:"rideId"`
	AdditionalDistance float64 `json:"additionalDistance"`
	AdditionalCharge   float64 `json:"additionalCharge"`
	NewTotalFare       float64 `json:"newTotalFare"`
}

type FareComponent struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Type   string  `json:"type"` // base, distance, duration, surge, booking_fee
}

// SurgeResponse contains basic surge information
type SurgeResponse struct {
	Multiplier float64 `json:"multiplier"`
	ZoneID     string  `json:"zoneId,omitempty"`
	ZoneName   string  `json:"zoneName,omitempty"`
	Reason     string  `json:"reason"`
}

// CombinedSurgeResponse contains detailed surge breakdown
type CombinedSurgeResponse struct {
	AppliedMultiplier     float64 `json:"appliedMultiplier"`
	ZoneBasedMultiplier   float64 `json:"zoneBasedMultiplier"`
	TimeBasedMultiplier   float64 `json:"timeBasedMultiplier"`
	DemandBasedMultiplier float64 `json:"demandBasedMultiplier"`
	Reason                string  `json:"reason"`
	ZoneID                string  `json:"zoneId,omitempty"`
	ZoneName              string  `json:"zoneName,omitempty"`
}

// SurgeDetailsResponse is included in fare estimate
type SurgeDetailsResponse struct {
	IsActive              bool    `json:"isActive"`
	AppliedMultiplier     float64 `json:"appliedMultiplier"`
	ZoneBasedMultiplier   float64 `json:"zoneBasedMultiplier"`
	TimeBasedMultiplier   float64 `json:"timeBasedMultiplier"`
	DemandBasedMultiplier float64 `json:"demandBasedMultiplier"`
	Reason                string  `json:"reason"`
	ZoneID                string  `json:"zoneId,omitempty"`
	ZoneName              string  `json:"zoneName,omitempty"`
}

// // Update FareEstimateResponse to include surge details
// type FareEstimateResponse struct {
//     BaseFare          float64              `json:"baseFare"`
//     DistanceFare      float64              `json:"distanceFare"`
//     DurationFare      float64              `json:"durationFare"`
//     BookingFee        float64              `json:"bookingFee"`
//     SurgeMultiplier   float64              `json:"surgeMultiplier"`
//     SubTotal          float64              `json:"subTotal"`
//     SurgeAmount       float64              `json:"surgeAmount"`
//     TotalFare         float64              `json:"totalFare"`
//     EstimatedDistance float64              `json:"estimatedDistance"`
//     EstimatedDuration int                  `json:"estimatedDuration"`
//     VehicleTypeName   string               `json:"vehicleTypeName"`
//     Currency          string               `json:"currency"`
//     SurgeDetails      *SurgeDetailsResponse `json:"surgeDetails,omitempty"`
// }
