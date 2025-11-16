package riderdto

import "errors"

type UpdateProfileRequest struct {
	HomeAddress          *AddressInput `json:"homeAddress,omitempty"`
	WorkAddress          *AddressInput `json:"workAddress,omitempty"`
	PreferredVehicleType *string       `json:"preferredVehicleType,omitempty"`
}

type AddressInput struct {
	Lat     float64 `json:"lat" binding:"required"`
	Lng     float64 `json:"lng" binding:"required"`
	Address string  `json:"address" binding:"required,min=5,max=500"`
}

func (a *AddressInput) Validate() error {
	if a.Lat < -90 || a.Lat > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if a.Lng < -180 || a.Lng > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	if a.Address == "" {
		return errors.New("address is required")
	}
	return nil
}

func (r *UpdateProfileRequest) Validate() error {
	if r.HomeAddress != nil {
		if err := r.HomeAddress.Validate(); err != nil {
			return err
		}
	}
	if r.WorkAddress != nil {
		if err := r.WorkAddress.Validate(); err != nil {
			return err
		}
	}
	if r.PreferredVehicleType != nil && *r.PreferredVehicleType == "" {
		return errors.New("preferredVehicleType cannot be empty")
	}
	return nil
}
