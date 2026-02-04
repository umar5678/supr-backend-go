package dto

import "errors"

type FareEstimateRequest struct {
	PickupLat     float64 `json:"pickupLat" binding:"required,min=-90,max=90"`
	PickupLon     float64 `json:"pickupLon" binding:"required,min=-180,max=180"`
	DropoffLat    float64 `json:"dropoffLat" binding:"required,min=-90,max=90"`
	DropoffLon    float64 `json:"dropoffLon" binding:"required,min=-180,max=180"`
	VehicleTypeID string  `json:"vehicleTypeId" binding:"required,uuid"`
}

func (r *FareEstimateRequest) Validate() error {
	if r.PickupLat < -90 || r.PickupLat > 90 {
		return errors.New("invalid pickup latitude")
	}
	if r.PickupLon < -180 || r.PickupLon > 180 {
		return errors.New("invalid pickup longitude")
	}
	if r.DropoffLat < -90 || r.DropoffLat > 90 {
		return errors.New("invalid dropoff latitude")
	}
	if r.DropoffLon < -180 || r.DropoffLon > 180 {
		return errors.New("invalid dropoff longitude")
	}
	if r.VehicleTypeID == "" {
		return errors.New("vehicle type is required")
	}

	// Check if pickup and dropoff are different
	if r.PickupLat == r.DropoffLat && r.PickupLon == r.DropoffLon {
		return errors.New("pickup and dropoff locations must be different")
	}

	return nil
}

type CalculateActualFareRequest struct {
	ActualDistanceKm  float64 `json:"actualDistanceKm" binding:"required,min=0"`
	ActualDurationSec int     `json:"actualDurationSec" binding:"required,min=0"`
	VehicleTypeID     string  `json:"vehicleTypeId" binding:"required,uuid"`
	SurgeMultiplier   float64 `json:"surgeMultiplier" binding:"omitempty,min=1,max=5"`
}

type CalculateWaitTimeRequest struct {
	RideID    string `json:"rideId" binding:"required,uuid"`
	ArrivedAt string `json:"arrivedAt" binding:"required"` // ISO 8601 timestamp
}

type ChangeDestinationRequest struct {
	RideID       string  `json:"rideId" binding:"required,uuid"`
	NewLatitude  float64 `json:"newLatitude" binding:"required,min=-90,max=90"`
	NewLongitude float64 `json:"newLongitude" binding:"required,min=-180,max=180"`
	NewAddress   string  `json:"newAddress" binding:"required,max=500"`
}

func (r *ChangeDestinationRequest) Validate() error {
	if r.NewLatitude == 0 && r.NewLongitude == 0 {
		return errors.New("new destination location is required")
	}
	if r.NewAddress == "" {
		return errors.New("new address is required")
	}
	return nil
}

type GetFareBreakdownRequest struct {
	PickupLat     float64 `form:"pickupLat" binding:"required,min=-90,max=90"`
	PickupLon     float64 `form:"pickupLon" binding:"required,min=-180,max=180"`
	DropoffLat    float64 `form:"dropoffLat" binding:"required,min=-90,max=90"`
	DropoffLon    float64 `form:"dropoffLon" binding:"required,min=-180,max=180"`
	VehicleTypeID string  `form:"vehicleTypeId" binding:"required,uuid"`
}

// CreateSurgeZoneRequest creates a new surge pricing zone
type CreateSurgeZoneRequest struct {
	AreaName    string  `json:"areaName" binding:"required,max=255"`
	AreaGeohash string  `json:"areaGeohash" binding:"required,max=12"`
	CenterLat   float64 `json:"centerLat" binding:"required,min=-90,max=90"`
	CenterLon   float64 `json:"centerLon" binding:"required,min=-180,max=180"`
	RadiusKm    float64 `json:"radiusKm" binding:"required,min=0.1,max=100"`
	Multiplier  float64 `json:"multiplier" binding:"required,min=1.0,max=5.0"`
	ActiveFrom  string  `json:"activeFrom" binding:"required"`  // ISO 8601 timestamp
	ActiveUntil string  `json:"activeUntil" binding:"required"` // ISO 8601 timestamp
	IsActive    bool    `json:"isActive" binding:"omitempty"`
}

func (r *CreateSurgeZoneRequest) Validate() error {
	if r.AreaName == "" {
		return errors.New("area name is required")
	}
	if r.AreaGeohash == "" {
		return errors.New("area geohash is required")
	}
	if r.CenterLat == 0 && r.CenterLon == 0 {
		return errors.New("center coordinates are required")
	}
	if r.RadiusKm <= 0 {
		return errors.New("radius must be greater than 0")
	}
	if r.Multiplier < 1.0 {
		return errors.New("multiplier must be at least 1.0")
	}
	if r.ActiveFrom == "" || r.ActiveUntil == "" {
		return errors.New("active from and until times are required")
	}
	return nil
}
