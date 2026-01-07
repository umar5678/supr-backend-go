package dto

import "errors"

type CreateRatingRequest struct {
	OrderID string  `json:"orderId" binding:"required"`
	Score   int     `json:"score" binding:"required,min=1,max=5"`
	Comment *string `json:"comment" binding:"omitempty,max=500"`
}

type RateDriverRequest struct {
	RideID  string `json:"rideId" binding:"required,uuid"`
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment" binding:"omitempty,max=500"`
}

func (r *RateDriverRequest) Validate() error {
	if r.Rating < 1 || r.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}
	return nil
}

type RateRiderRequest struct {
	RideID  string `json:"rideId" binding:"required,uuid"`
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment" binding:"omitempty,max=500"`
}

func (r *RateRiderRequest) Validate() error {
	if r.Rating < 1 || r.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}
	return nil
}
