package dto

import "errors"

type TriggerSOSRequest struct {
    RideID    *string `json:"rideId" binding:"omitempty,uuid"`
    Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
    Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
}

func (r *TriggerSOSRequest) Validate() error {
    if r.Latitude == 0 && r.Longitude == 0 {
        return errors.New("location is required")
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
    if r.Page == 0 {
        r.Page = 1
    }
    if r.Limit == 0 {
        r.Limit = 20
    }
}
