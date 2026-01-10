package driverdto

import "errors"

type RegisterDriverRequest struct {
	LicenseNumber string       `json:"licenseNumber" binding:"required,min=5,max=100"`
	Vehicle       VehicleInput `json:"vehicle" binding:"required"`
}

type VehicleInput struct {
	VehicleTypeID string `json:"vehicleTypeId" binding:"required,uuid"`
	Make          string `json:"make" binding:"required,min=2,max=100"`
	Model         string `json:"model" binding:"required,min=2,max=100"`
	Year          int    `json:"year" binding:"required,min=1900"`
	Color         string `json:"color" binding:"required,min=2,max=50"`
	LicensePlate  string `json:"licensePlate" binding:"required,min=2,max=50"`
	Capacity      int    `json:"capacity" binding:"required,min=1,max=20"`
}

func (r *RegisterDriverRequest) Validate() error {
	if r.LicenseNumber == "" {
		return errors.New("license number is required")
	}
	if r.Vehicle.VehicleTypeID == "" {
		return errors.New("vehicle type is required")
	}
	if r.Vehicle.Make == "" {
		return errors.New("vehicle make is required")
	}
	if r.Vehicle.Model == "" {
		return errors.New("vehicle model is required")
	}
	if r.Vehicle.Year < 1900 {
		return errors.New("invalid vehicle year")
	}
	if r.Vehicle.LicensePlate == "" {
		return errors.New("license plate is required")
	}
	return nil
}

type UpdateDriverProfileRequest struct {
	LicenseNumber *string `json:"licenseNumber" binding:"omitempty,min=5,max=100"`
}

type UpdateVehicleRequest struct {
	VehicleTypeID *string `json:"vehicleTypeId" binding:"omitempty,uuid"`
	Make          *string `json:"make" binding:"omitempty,min=2,max=100"`
	Model         *string `json:"model" binding:"omitempty,min=2,max=100"`
	Year          *int    `json:"year" binding:"omitempty,min=1900"`
	Color         *string `json:"color" binding:"omitempty,min=2,max=50"`
	LicensePlate  *string `json:"licensePlate" binding:"omitempty,min=2,max=50"`
	Capacity      *int    `json:"capacity" binding:"omitempty,min=1,max=20"`
	IsActive      *bool   `json:"isActive"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=offline online"`
}

func (r *UpdateStatusRequest) Validate() error {
	if r.Status != "offline" && r.Status != "online" {
		return errors.New("status must be either 'offline' or 'online'")
	}
	return nil
}

type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
	Heading   int     `json:"heading" binding:"omitempty,min=0,max=360"`
}

func (r *UpdateLocationRequest) Validate() error {
	if r.Latitude < -90 || r.Latitude > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if r.Longitude < -180 || r.Longitude > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	if r.Heading < 0 || r.Heading > 360 {
		r.Heading = 0
	}
	return nil
}

// ✅ WalletTopUpRequest - Request to add funds to wallet
type WalletTopUpRequest struct {
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PaymentMethod string  `json:"paymentMethod" binding:"required,oneof=card upi netbanking wallet"` // card, upi, netbanking, wallet
	Reference     *string `json:"reference" binding:"omitempty,max=100"`                             // Transaction reference/order ID
}

func (r *WalletTopUpRequest) Validate() error {
	if r.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if r.Amount > 100000 { // Max top-up limit
		return errors.New("amount cannot exceed 100,000")
	}
	if r.PaymentMethod == "" {
		return errors.New("payment method is required")
	}
	validMethods := map[string]bool{"card": true, "upi": true, "netbanking": true, "wallet": true}
	if !validMethods[r.PaymentMethod] {
		return errors.New("invalid payment method")
	}
	return nil
}

