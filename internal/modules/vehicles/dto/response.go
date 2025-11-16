package dto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
)

type VehicleTypeResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	DisplayName   string    `json:"displayName"`
	BaseFare      float64   `json:"baseFare"`
	PerKmRate     float64   `json:"perKmRate"`
	PerMinuteRate float64   `json:"perMinuteRate"`
	BookingFee    float64   `json:"bookingFee"`
	Capacity      int       `json:"capacity"`
	Description   string    `json:"description"`
	IsActive      bool      `json:"isActive"`
	IconURL       string    `json:"iconUrl"`
	CreatedAt     time.Time `json:"createdAt"`
}

func ToVehicleTypeResponse(vt *models.VehicleType) *VehicleTypeResponse {
	return &VehicleTypeResponse{
		ID:            vt.ID,
		Name:          vt.Name,
		DisplayName:   vt.DisplayName,
		BaseFare:      vt.BaseFare,
		PerKmRate:     vt.PerKmRate,
		PerMinuteRate: vt.PerMinuteRate,
		BookingFee:    vt.BookingFee,
		Capacity:      vt.Capacity,
		Description:   vt.Description,
		IsActive:      vt.IsActive,
		IconURL:       vt.IconURL,
		CreatedAt:     vt.CreatedAt,
	}
}
