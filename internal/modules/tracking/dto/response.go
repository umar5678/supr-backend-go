package dto

import (
	"time"

	driverdto "github.com/umar5678/go-backend/internal/modules/drivers/dto"
)

type LocationResponse struct {
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Heading   int       `json:"heading"`
	Speed     float64   `json:"speed"`
	Accuracy  float64   `json:"accuracy"`
	Timestamp time.Time `json:"timestamp"`
}

type DriverLocationResponse struct {
	DriverID string                           `json:"driverId"`
	Driver   *driverdto.DriverProfileResponse `json:"driver,omitempty"`
	Location LocationResponse                 `json:"location"`
	Distance float64                          `json:"distance"` // km from search point
	ETA      int                              `json:"eta"`      // seconds
}

type NearbyDriversResponse struct {
	SearchLocation LocationResponse         `json:"searchLocation"`
	RadiusKm       float64                  `json:"radiusKm"`
	Drivers        []DriverLocationResponse `json:"drivers"`
	Count          int                      `json:"count"`
}
