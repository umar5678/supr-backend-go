package dto

import (
	"errors"
	"fmt"
)

type TriggerSOSRequest struct {
	RideID    *string `json:"rideId" binding:"omitempty,uuid"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

func (r *TriggerSOSRequest) Validate() error {
	if r.Latitude == 0 && r.Longitude == 0 {
		return errors.New("valid location is required")
	}
	if r.Latitude < -90 || r.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90, got %f", r.Latitude)
	}
	if r.Longitude < -180 || r.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180, got %f", r.Longitude)
	}
	return nil
}

type ResolveSOSRequest struct {
	Notes string `json:"notes" binding:"omitempty,max=1000"`
}

type ListSOSRequest struct {
	Status string `form:"status" binding:"omitempty,oneof=active resolved cancelled"`
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

func (r *ListSOSRequest) SetDefaults() {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Limit <= 0 {
		r.Limit = 20
	}
	if r.Limit > 100 {
		r.Limit = 100
	}
}

type UpdateSOSLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}