package dto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
	authdto "github.com/umar5678/go-backend/internal/modules/auth/dto"
	vehicledto "github.com/umar5678/go-backend/internal/modules/vehicles/dto"
)

type RideResponse struct {
	ID            string                          `json:"id"`
	RiderID       string                          `json:"riderId"`
	Rider         *authdto.UserResponse           `json:"rider,omitempty"`
	DriverID      *string                         `json:"driverId"`
	Driver        *authdto.UserResponse           `json:"driver,omitempty"`
	VehicleTypeID string                          `json:"vehicleTypeId"`
	VehicleType   *vehicledto.VehicleTypeResponse `json:"vehicleType,omitempty"`
	Status        string                          `json:"status"`

	PickupLat      float64 `json:"pickupLat"`
	PickupLon      float64 `json:"pickupLon"`
	PickupAddress  string  `json:"pickupAddress"`
	DropoffLat     float64 `json:"dropoffLat"`
	DropoffLon     float64 `json:"dropoffLon"`
	DropoffAddress string  `json:"dropoffAddress"`

	EstimatedDistance float64 `json:"estimatedDistance"`
	EstimatedDuration int     `json:"estimatedDuration"`
	EstimatedFare     float64 `json:"estimatedFare"`

	ActualDistance *float64 `json:"actualDistance,omitempty"`
	ActualDuration *int     `json:"actualDuration,omitempty"`
	ActualFare     *float64 `json:"actualFare,omitempty"`
	PromoDiscount  *float64 `json:"promoDiscount,omitempty"`
	WaitTimeCharge *float64 `json:"waitTimeCharge,omitempty"`

	SurgeMultiplier    float64 `json:"surgeMultiplier"`
	RiderNotes         string  `json:"riderNotes,omitempty"`
	CancellationReason string  `json:"cancellationReason,omitempty"`
	CancelledBy        *string `json:"cancelledBy,omitempty"`

	IsScheduled bool       `json:"isScheduled"`
	ScheduledAt *time.Time `json:"scheduledAt,omitempty"`

	RequestedAt time.Time  `json:"requestedAt"`
	AcceptedAt  *time.Time `json:"acceptedAt,omitempty"`
	ArrivedAt   *time.Time `json:"arrivedAt,omitempty"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CancelledAt *time.Time `json:"cancelledAt,omitempty"`

	HasActiveSOS bool    `json:"hasActiveSos"`
	SOSAlertID   *string `json:"sosAlertId,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	DriverLocation *LocationDTO `json:"driverLocation,omitempty"`
}

type RideListResponse struct {
	ID             string    `json:"id"`
	Status         string    `json:"status"`
	PickupAddress  string    `json:"pickupAddress"`
	DropoffAddress string    `json:"dropoffAddress"`
	EstimatedFare  float64   `json:"estimatedFare"`
	RequestedAt    time.Time `json:"requestedAt"`
}

func ToRideResponse(ride *models.Ride) *RideResponse {
	// ✅ CRITICAL: Handle nil ride to prevent panic
	if ride == nil {
		return nil
	}

	resp := &RideResponse{
		ID:                 ride.ID,
		RiderID:            ride.RiderID,
		DriverID:           ride.DriverID,
		VehicleTypeID:      ride.VehicleTypeID,
		Status:             ride.Status,
		PickupLat:          ride.PickupLat,
		PickupLon:          ride.PickupLon,
		PickupAddress:      ride.PickupAddress,
		DropoffLat:         ride.DropoffLat,
		DropoffLon:         ride.DropoffLon,
		DropoffAddress:     ride.DropoffAddress,
		EstimatedDistance:  ride.EstimatedDistance,
		EstimatedDuration:  ride.EstimatedDuration,
		EstimatedFare:      ride.EstimatedFare,
		ActualDistance:     ride.ActualDistance,
		ActualDuration:     ride.ActualDuration,
		ActualFare:         ride.ActualFare,
		SurgeMultiplier:    ride.SurgeMultiplier,
		RiderNotes:         ride.RiderNotes,
		CancellationReason: ride.CancellationReason,
		CancelledBy:        ride.CancelledBy,
		IsScheduled:        ride.IsScheduled,
		ScheduledAt:        ride.ScheduledAt,
		RequestedAt:        ride.RequestedAt,
		AcceptedAt:         ride.AcceptedAt,
		ArrivedAt:          ride.ArrivedAt,
		StartedAt:          ride.StartedAt,
		CompletedAt:        ride.CompletedAt,
		CancelledAt:        ride.CancelledAt,
		CreatedAt:          ride.CreatedAt,
		UpdatedAt:          ride.UpdatedAt,
	}

	if ride.Rider.ID != "" {
		resp.Rider = authdto.ToUserResponse(&ride.Rider)
	}
	if ride.Driver != nil && ride.Driver.ID != "" {
		resp.Driver = authdto.ToUserResponse(ride.Driver)
	}
	if ride.VehicleType.ID != "" {
		resp.VehicleType = vehicledto.ToVehicleTypeResponse(&ride.VehicleType)
	}

	return resp
}

func ToRideListResponse(ride *models.Ride) *RideListResponse {
	return &RideListResponse{
		ID:             ride.ID,
		Status:         ride.Status,
		PickupAddress:  ride.PickupAddress,
		DropoffAddress: ride.DropoffAddress,
		EstimatedFare:  ride.EstimatedFare,
		RequestedAt:    ride.RequestedAt,
	}
}

type LocationDTO struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// type RideResponse struct {
//     // ... existing fields ...
//     DriverLocation *LocationDTO `json:"driverLocation,omitempty"`
// }
