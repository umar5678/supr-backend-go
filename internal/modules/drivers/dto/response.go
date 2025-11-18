package driverdto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
	authdto "github.com/umar5678/go-backend/internal/modules/auth/dto"
	vehicledto "github.com/umar5678/go-backend/internal/modules/vehicles/dto"
)

type DriverProfileResponse struct {
	ID               string                `json:"id"`
	UserID           string                `json:"userId"`
	User             *authdto.UserResponse `json:"user,omitempty"`
	LicenseNumber    string                `json:"licenseNumber"`
	Status           string                `json:"status"`
	CurrentLocation  *LocationResponse     `json:"currentLocation,omitempty"`
	Heading          int                   `json:"heading"`
	Rating           float64               `json:"rating"`
	TotalTrips       int                   `json:"totalTrips"`
	TotalEarnings    float64               `json:"totalEarnings"`
	AcceptanceRate   float64               `json:"acceptanceRate"`
	CancellationRate float64               `json:"cancellationRate"`
	IsVerified       bool                  `json:"isVerified"`
	WalletBalance    float64               `json:"walletBalance"`
	Vehicle          *VehicleResponse      `json:"vehicle,omitempty"`
	CreatedAt        time.Time             `json:"createdAt"`
	UpdatedAt        time.Time             `json:"updatedAt"`
}

type LocationResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type VehicleResponse struct {
	ID            string                          `json:"id"`
	DriverID      string                          `json:"driverId"`
	VehicleTypeID string                          `json:"vehicleTypeId"`
	VehicleType   *vehicledto.VehicleTypeResponse `json:"vehicleType,omitempty"`
	Make          string                          `json:"make"`
	Model         string                          `json:"model"`
	Year          int                             `json:"year"`
	Color         string                          `json:"color"`
	LicensePlate  string                          `json:"licensePlate"`
	Capacity      int                             `json:"capacity"`
	IsActive      bool                            `json:"isActive"`
	CreatedAt     time.Time                       `json:"createdAt"`
	UpdatedAt     time.Time                       `json:"updatedAt"`
}

type DriverDashboardResponse struct {
	TodayTrips     int                    `json:"todayTrips"`
	TodayEarnings  float64                `json:"todayEarnings"`
	WeekEarnings   float64                `json:"weekEarnings"`
	Rating         float64                `json:"rating"`
	AcceptanceRate float64                `json:"acceptanceRate"`
	CompletionRate float64                `json:"completionRate"`
	TotalTrips     int                    `json:"totalTrips"`
	WalletBalance  float64                `json:"walletBalance"`
	Status         string                 `json:"status"`
	Profile        *DriverProfileResponse `json:"profile,omitempty"`
}

type WalletResponse struct {
	Balance         float64   `json:"balance"`
	TotalEarnings   float64   `json:"totalEarnings"`
	PendingAmount   float64   `json:"pendingAmount"`
	LastTransaction time.Time `json:"lastTransaction,omitempty"`
}

func ToDriverProfileResponse(driver *models.DriverProfile) *DriverProfileResponse {
	resp := &DriverProfileResponse{
		ID:               driver.ID,
		UserID:           driver.UserID,
		LicenseNumber:    driver.LicenseNumber,
		Status:           driver.Status,
		Heading:          driver.Heading,
		Rating:           driver.Rating,
		TotalTrips:       driver.TotalTrips,
		TotalEarnings:    driver.TotalEarnings,
		AcceptanceRate:   driver.AcceptanceRate,
		CancellationRate: driver.CancellationRate,
		IsVerified:       driver.IsVerified,
		WalletBalance:    driver.WalletBalance,
		CreatedAt:        driver.CreatedAt,
		UpdatedAt:        driver.UpdatedAt,
	}

	// Parse location if exists
	if driver.CurrentLocation != nil && *driver.CurrentLocation != "" {
		// Location is stored as "POINT(longitude latitude)"
		// You'll need to parse this - for now using placeholder
		resp.CurrentLocation = &LocationResponse{
			Latitude:  0, // Parse from driver.CurrentLocation
			Longitude: 0,
		}
	}

	if driver.User.ID != "" {
		resp.User = authdto.ToUserResponse(&driver.User)
	}

	if driver.Vehicle != nil {
		resp.Vehicle = ToVehicleResponse(driver.Vehicle)
	}

	return resp
}

func ToVehicleResponse(vehicle *models.Vehicle) *VehicleResponse {
	resp := &VehicleResponse{
		ID:            vehicle.ID,
		DriverID:      vehicle.DriverID,
		VehicleTypeID: vehicle.VehicleTypeID,
		Make:          vehicle.Make,
		Model:         vehicle.Model,
		Year:          vehicle.Year,
		Color:         vehicle.Color,
		LicensePlate:  vehicle.LicensePlate,
		Capacity:      vehicle.Capacity,
		IsActive:      vehicle.IsActive,
		CreatedAt:     vehicle.CreatedAt,
		UpdatedAt:     vehicle.UpdatedAt,
	}

	if vehicle.VehicleType.ID != "" {
		resp.VehicleType = vehicledto.ToVehicleTypeResponse(&vehicle.VehicleType)
	}

	return resp
}
