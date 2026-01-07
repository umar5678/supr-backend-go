package dto

import (
    "time"
    "github.com/umar5678/go-backend/internal/models"
)

type KYCResponse struct {
    ID              string     `json:"id"`
    UserID          string     `json:"userId"`
    IDType          string     `json:"idType"`
    IDNumber        string     `json:"idNumber"`
    IDDocumentURL   string     `json:"idDocumentUrl"`
    SelfieURL       string     `json:"selfieUrl"`
    Status          string     `json:"status"`
    RejectionReason string     `json:"rejectionReason,omitempty"`
    VerifiedAt      *time.Time `json:"verifiedAt,omitempty"`
    CreatedAt       time.Time  `json:"createdAt"`
}

type SavedLocationResponse struct {
    ID         string    `json:"id"`
    UserID     string    `json:"userId"`
    Label      string    `json:"label"`
    CustomName string    `json:"customName,omitempty"`
    Address    string    `json:"address"`
    Latitude   float64   `json:"latitude"`
    Longitude  float64   `json:"longitude"`
    IsDefault  bool      `json:"isDefault"`
    CreatedAt  time.Time `json:"createdAt"`
}

type RecentLocationResponse struct {
    Address   string    `json:"address"`
    Latitude  float64   `json:"latitude"`
    Longitude float64   `json:"longitude"`
    LastUsed  time.Time `json:"lastUsed"`
}

type ReferralInfoResponse struct {
    ReferralCode  string `json:"referralCode"`
    ReferralCount int64  `json:"referralCount"`
    ReferralBonus float64 `json:"referralBonus"`
}

func ToKYCResponse(kyc *models.UserKYC) *KYCResponse {
    return &KYCResponse{
        ID:              kyc.ID,
        UserID:          kyc.UserID,
        IDType:          kyc.IDType,
        IDNumber:        kyc.IDNumber,
        IDDocumentURL:   kyc.IDDocumentURL,
        SelfieURL:       kyc.SelfieURL,
        Status:          kyc.Status,
        RejectionReason: kyc.RejectionReason,
        VerifiedAt:      kyc.VerifiedAt,
        CreatedAt:       kyc.CreatedAt,
    }
}

func ToSavedLocationResponse(loc *models.SavedLocation) *SavedLocationResponse {
    return &SavedLocationResponse{
        ID:         loc.ID,
        UserID:     loc.UserID,
        Label:      loc.Label,
        CustomName: loc.CustomName,
        Address:    loc.Address,
        Latitude:   loc.Latitude,
        Longitude:  loc.Longitude,
        IsDefault:  loc.IsDefault,
        CreatedAt:  loc.CreatedAt,
    }
}
