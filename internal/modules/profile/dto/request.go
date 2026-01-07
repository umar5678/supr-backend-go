package dto

import "errors"

type UpdateEmergencyContactRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=255"`
	Phone string `json:"phone" binding:"required,min=10,max=20"`
}

func (r *UpdateEmergencyContactRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.Phone == "" {
		return errors.New("phone is required")
	}
	return nil
}

type ApplyReferralRequest struct {
	ReferralCode string `json:"referralCode" binding:"required,min=6,max=20"`
}

type SubmitKYCRequest struct {
	IDType        string `json:"idType" binding:"required,oneof=national_id passport driver_license"`
	IDNumber      string `json:"idNumber" binding:"required,min=5,max=100"`
	IDDocumentURL string `json:"idDocumentUrl" binding:"required,url"`
	SelfieURL     string `json:"selfieUrl" binding:"required,url"`
}

func (r *SubmitKYCRequest) Validate() error {
	if r.IDType == "" {
		return errors.New("ID type is required")
	}
	if r.IDNumber == "" {
		return errors.New("ID number is required")
	}
	return nil
}

type SaveLocationRequest struct {
	Label      string  `json:"label" binding:"required,oneof=home work custom"`
	CustomName string  `json:"customName" binding:"omitempty,max=100"`
	Address    string  `json:"address" binding:"required,max=500"`
	Latitude   float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude  float64 `json:"longitude" binding:"required,min=-180,max=180"`
	IsDefault  bool    `json:"isDefault"`
}

func (r *SaveLocationRequest) Validate() error {
	if r.Label == "" {
		return errors.New("label is required")
	}
	if r.Label == "custom" && r.CustomName == "" {
		return errors.New("custom name is required for custom locations")
	}
	if r.Address == "" {
		return errors.New("address is required")
	}
	return nil
}
