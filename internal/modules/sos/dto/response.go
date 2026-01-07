package dto

import (
    "time"
    "github.com/umar5678/go-backend/internal/models"
    authdto "github.com/umar5678/go-backend/internal/modules/auth/dto"
)

type SOSAlertResponse struct {
    ID                        string                `json:"id"`
    UserID                    string                `json:"userId"`
    User                      *authdto.UserResponse `json:"user,omitempty"`
    RideID                    *string               `json:"rideId,omitempty"`
    AlertType                 string                `json:"alertType"`
    Latitude                  float64               `json:"latitude"`
    Longitude                 float64               `json:"longitude"`
    Status                    string                `json:"status"`
    EmergencyContactsNotified bool                  `json:"emergencyContactsNotified"`
    SafetyTeamNotifiedAt      *time.Time            `json:"safetyTeamNotifiedAt,omitempty"`
    ResolvedAt                *time.Time            `json:"resolvedAt,omitempty"`
    ResolvedBy                *string               `json:"resolvedBy,omitempty"`
    Notes                     string                `json:"notes,omitempty"`
    CreatedAt                 time.Time             `json:"createdAt"`
}

type SOSAlertListResponse struct {
    ID        string    `json:"id"`
    UserID    string    `json:"userId"`
    RideID    *string   `json:"rideId,omitempty"`
    AlertType string    `json:"alertType"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"createdAt"`
}

func ToSOSAlertResponse(alert *models.SOSAlert) *SOSAlertResponse {
    resp := &SOSAlertResponse{
        ID:                        alert.ID,
        UserID:                    alert.UserID,
        RideID:                    alert.RideID,
        AlertType:                 alert.AlertType,
        Latitude:                  alert.Latitude,
        Longitude:                 alert.Longitude,
        Status:                    alert.Status,
        EmergencyContactsNotified: alert.EmergencyContactsNotified,
        SafetyTeamNotifiedAt:      alert.SafetyTeamNotifiedAt,
        ResolvedAt:                alert.ResolvedAt,
        ResolvedBy:                alert.ResolvedBy,
        Notes:                     alert.Notes,
        CreatedAt:                 alert.CreatedAt,
    }

    if alert.User.ID != "" {
        resp.User = authdto.ToUserResponse(&alert.User)
    }

    return resp
}

func ToSOSAlertListResponse(alert *models.SOSAlert) *SOSAlertListResponse {
    return &SOSAlertListResponse{
        ID:        alert.ID,
        UserID:    alert.UserID,
        RideID:    alert.RideID,
        AlertType: alert.AlertType,
        Status:    alert.Status,
        CreatedAt: alert.CreatedAt,
    }
}
