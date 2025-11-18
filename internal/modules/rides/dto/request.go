package dto

import "errors"

type CreateRideRequest struct {
	PickupLat      float64 `json:"pickupLat" binding:"required,min=-90,max=90"`
	PickupLon      float64 `json:"pickupLon" binding:"required,min=-180,max=180"`
	PickupAddress  string  `json:"pickupAddress" binding:"required,max=500"`
	DropoffLat     float64 `json:"dropoffLat" binding:"required,min=-90,max=90"`
	DropoffLon     float64 `json:"dropoffLon" binding:"required,min=-180,max=180"`
	DropoffAddress string  `json:"dropoffAddress" binding:"required,max=500"`
	VehicleTypeID  string  `json:"vehicleTypeId" binding:"required,uuid"`
	RiderNotes     string  `json:"riderNotes" binding:"omitempty,max=500"`
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

type CompleteRideRequest struct {
	ActualDistance float64 `json:"actualDistance" binding:"required,min=0"`
	ActualDuration int     `json:"actualDuration" binding:"required,min=0"`
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

// // internal/modules/rides/dto/request.go
// package dto

// import "errors"

// type CreateRideRequest struct {
// 	PickupLat      float64 `json:"pickupLat" binding:"required,min=-90,max=90"`
// 	PickupLon      float64 `json:"pickupLon" binding:"required,min=-180,max=180"`
// 	PickupAddress  string  `json:"pickupAddress" binding:"omitempty,max=500"`
// 	DropoffLat     float64 `json:"dropoffLat" binding:"required,min=-90,max=90"`
// 	DropoffLon     float64 `json:"dropoffLon" binding:"required,min=-180,max=180"`
// 	DropoffAddress string  `json:"dropoffAddress" binding:"omitempty,max=500"`
// 	VehicleTypeID  string  `json:"vehicleTypeId" binding:"required,uuid"`
// 	RiderNotes     string  `json:"riderNotes" binding:"omitempty,max=500"`
// }

// func (r *CreateRideRequest) Validate() error {
// 	if r.PickupLat == r.DropoffLat && r.PickupLon == r.DropoffLon {
// 		return errors.New("pickup and dropoff locations must be different")
// 	}
// 	return nil
// }

// type AcceptRideRequest struct {
// 	RideID string `json:"rideId" binding:"required,uuid"`
// }

// type RejectRideRequest struct {
// 	RideID string `json:"rideId" binding:"required,uuid"`
// 	Reason string `json:"reason" binding:"omitempty,max=500"`
// }

// type CancelRideRequest struct {
// 	Reason string `json:"reason" binding:"omitempty,max=500"`
// }

// type CompleteRideRequest struct {
// 	ActualDistance float64 `json:"actualDistance" binding:"required,min=0"`
// 	ActualDuration int     `json:"actualDuration" binding:"required,min=0"`
// }

// type ListRidesRequest struct {
// 	Page   int    `form:"page" binding:"omitempty,min=1"`
// 	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
// 	Status string `form:"status" binding:"omitempty,oneof=searching accepted arrived started completed cancelled"`
// }

// func (r *ListRidesRequest) SetDefaults() {
// 	if r.Page == 0 {
// 		r.Page = 1
// 	}
// 	if r.Limit == 0 {
// 		r.Limit = 20
// 	}
// }
