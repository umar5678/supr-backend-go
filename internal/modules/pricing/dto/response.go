// internal/modules/pricing/dto/response.go
package dto

type FareEstimateResponse struct {
	BaseFare          float64 `json:"baseFare"`
	DistanceFare      float64 `json:"distanceFare"`
	DurationFare      float64 `json:"durationFare"`
	BookingFee        float64 `json:"bookingFee"`
	SurgeMultiplier   float64 `json:"surgeMultiplier"`
	SubTotal          float64 `json:"subTotal"`
	SurgeAmount       float64 `json:"surgeAmount"`
	TotalFare         float64 `json:"totalFare"`
	EstimatedDistance float64 `json:"estimatedDistance"` // km
	EstimatedDuration int     `json:"estimatedDuration"` // seconds
	VehicleTypeName   string  `json:"vehicleTypeName"`
	Currency          string  `json:"currency"`
}

type SurgeZoneResponse struct {
	ID         string  `json:"id"`
	AreaName   string  `json:"areaName"`
	Multiplier float64 `json:"multiplier"`
	RadiusKm   float64 `json:"radiusKm"`
	IsActive   bool    `json:"isActive"`
}

type FareBreakdownResponse struct {
	Components []FareComponent `json:"components"`
	Total      float64         `json:"total"`
}

type FareComponent struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Type   string  `json:"type"` // base, distance, duration, surge, booking_fee
}
