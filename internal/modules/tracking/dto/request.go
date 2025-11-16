package dto

import "errors"

type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
	Heading   int     `json:"heading" binding:"omitempty,min=0,max=360"`
	Speed     float64 `json:"speed" binding:"omitempty,min=0,max=300"`
	Accuracy  float64 `json:"accuracy" binding:"omitempty,min=0"`
}

func (r *UpdateLocationRequest) Validate() error {
	if r.Latitude < -90 || r.Latitude > 90 {
		return errors.New("invalid latitude")
	}
	if r.Longitude < -180 || r.Longitude > 180 {
		return errors.New("invalid longitude")
	}
	if r.Heading < 0 || r.Heading > 360 {
		r.Heading = 0
	}
	if r.Speed < 0 {
		r.Speed = 0
	}
	return nil
}

type FindNearbyDriversRequest struct {
	Latitude      float64 `form:"latitude" binding:"required,min=-90,max=90"`
	Longitude     float64 `form:"longitude" binding:"required,min=-180,max=180"`
	RadiusKm      float64 `form:"radiusKm" binding:"omitempty,min=0.1,max=50"`
	VehicleTypeID string  `form:"vehicleTypeId" binding:"omitempty,uuid"`
	Limit         int     `form:"limit" binding:"omitempty,min=1,max=50"`
}

func (r *FindNearbyDriversRequest) SetDefaults() {
	if r.RadiusKm == 0 {
		r.RadiusKm = 5.0 // Default 5km radius
	}
	if r.Limit == 0 {
		r.Limit = 20
	}
}
