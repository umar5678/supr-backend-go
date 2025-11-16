// internal/modules/pricing/dto/request.go
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
