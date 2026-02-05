package dto

import (
	"errors"
	"time"
)

type CreateRideRequest struct {
	PickupLat       float64 `json:"pickupLat" binding:"required,min=-90,max=90"`
	PickupLon       float64 `json:"pickupLon" binding:"required,min=-180,max=180"`
	PickupAddress   string  `json:"pickupAddress" binding:"required,max=500"`
	DropoffLat      float64 `json:"dropoffLat" binding:"required,min=-90,max=90"`
	DropoffLon      float64 `json:"dropoffLon" binding:"required,min=-180,max=180"`
	DropoffAddress  string  `json:"dropoffAddress" binding:"required,max=500"`
	SavedLocationID *string `json:"savedLocationId" binding:"omitempty,uuid"`
	UseSavedAs      string  `json:"useSavedAs" binding:"omitempty,oneof=pickup dropoff"`
	VehicleTypeID   string  `json:"vehicleTypeId" binding:"required,uuid"`
	RiderNotes      string  `json:"riderNotes" binding:"omitempty,max=500"`
	PromoCode       string  `json:"promoCode" binding:"omitempty,min=3,max=50"`
	IsScheduled     bool    `json:"isScheduled" binding:"omitempty"`
	ScheduledAt string `json:"scheduledAt" binding:"omitempty"`
}

func (r *CreateRideRequest) Validate() error {
	if r.PickupLat == r.DropoffLat && r.PickupLon == r.DropoffLon {
		return errors.New("pickup and dropoff locations must be different")
	}
	if r.PickupAddress == "" {
		return errors.New("pickup address is required")
	}
	if r.DropoffAddress == "" {
		return errors.New("dropoff address is required")
	}

	if r.ScheduledAt != "" {
		t, err := time.Parse(time.RFC3339, r.ScheduledAt)
		if err != nil {
			return errors.New("scheduledAt must be a valid RFC3339 timestamp")
		}
		if !t.After(time.Now()) {
			return errors.New("scheduledAt must be a future time")
		}
	}
	return nil
}

type AcceptRideRequest struct {
	RideID string `json:"rideId" binding:"required,uuid"`
}

type RejectRideRequest struct {
	RideID string `json:"rideId" binding:"required,uuid"`
	Reason string `json:"reason" binding:"omitempty,max=500"`
}

type CancelRideRequest struct {
	Reason string `json:"reason" binding:"omitempty,max=500"`
}

type StartRideRequest struct {
	RiderPIN string `json:"riderPin" binding:"required,len=4"`
}

type CompleteRideRequest struct {
	ActualDistance float64 `json:"actualDistance" binding:"required,min=0"`
	ActualDuration int     `json:"actualDuration" binding:"required,min=0"`
	DriverLat      float64 `json:"driverLat" binding:"required"`
	DriverLon      float64 `json:"driverLon" binding:"required"`
}

type ListRidesRequest struct {
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Status string `form:"status" binding:"omitempty,oneof=searching accepted arrived started completed cancelled"`
}

func (r *ListRidesRequest) SetDefaults() {
	if r.Page == 0 {
		r.Page = 1
	}
	if r.Limit == 0 {
		r.Limit = 20
	}
}

type AvailableCarRequest struct {
	Latitude  float64 `json:"latitude" binding:"required,latitude"`
	Longitude float64 `json:"longitude" binding:"required,longitude"`
	RadiusKm  float64 `json:"radiusKm" binding:"required,min=0.1,max=50"`
}

type VehicleDetailsRequest struct {
	PickupLat      float64 `json:"pickupLat" binding:"required,latitude"`
	PickupLon      float64 `json:"pickupLon" binding:"required,longitude"`
	PickupAddress  string  `json:"pickupAddress" binding:"required,max=500"`
	DropoffLat     float64 `json:"dropoffLat" binding:"required,latitude"`
	DropoffLon     float64 `json:"dropoffLon" binding:"required,longitude"`
	DropoffAddress string  `json:"dropoffAddress" binding:"required,max=500"`
	RadiusKm       float64 `json:"radiusKm" binding:"omitempty,min=0.1,max=50"`
}
