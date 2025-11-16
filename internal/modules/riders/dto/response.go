package riderdto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
	authdto "github.com/umar5678/go-backend/internal/modules/auth/dto"
	walletdto "github.com/umar5678/go-backend/internal/modules/wallet/dto"
)

type AddressResponse struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address"`
}

type RiderProfileResponse struct {
	ID                   string                    `json:"id"`
	UserID               string                    `json:"userId"`
	HomeAddress          *AddressResponse          `json:"homeAddress,omitempty"`
	WorkAddress          *AddressResponse          `json:"workAddress,omitempty"`
	PreferredVehicleType *string                   `json:"preferredVehicleType,omitempty"`
	Rating               float64                   `json:"rating"`
	TotalRides           int                       `json:"totalRides"`
	CreatedAt            time.Time                 `json:"createdAt"`
	UpdatedAt            time.Time                 `json:"updatedAt"`
	User                 *authdto.UserResponse     `json:"user,omitempty"`
	Wallet               *walletdto.WalletResponse `json:"wallet,omitempty"`
}

type RiderStatsResponse struct {
	TotalRides    int     `json:"totalRides"`
	Rating        float64 `json:"rating"`
	WalletBalance float64 `json:"walletBalance"`
	MemberSince   string  `json:"memberSince"`
}

func ToAddressResponse(addr *models.Address) *AddressResponse {
	if addr == nil {
		return nil
	}
	return &AddressResponse{
		Lat:     addr.Lat,
		Lng:     addr.Lng,
		Address: addr.Address,
	}
}

func ToRiderProfileResponse(profile *models.RiderProfile) *RiderProfileResponse {
	resp := &RiderProfileResponse{
		ID:                   profile.ID,
		UserID:               profile.UserID,
		HomeAddress:          ToAddressResponse(profile.HomeAddress),
		WorkAddress:          ToAddressResponse(profile.WorkAddress),
		PreferredVehicleType: profile.PreferredVehicleType,
		Rating:               profile.Rating,
		TotalRides:           profile.TotalRides,
		CreatedAt:            profile.CreatedAt,
		UpdatedAt:            profile.UpdatedAt,
	}

	if profile.User.ID != "" {
		resp.User = authdto.ToUserResponse(&profile.User)
	}

	if profile.Wallet.ID != "" {
		resp.Wallet = walletdto.ToWalletResponse(&profile.Wallet)
	}

	return resp
}
