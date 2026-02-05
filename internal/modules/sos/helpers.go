// internal/modules/sos/helpers.go
package sos

import (
	"context"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/websocket"
	websocketutil "github.com/umar5678/go-backend/internal/websocket/websocketutils"
)


func sendEmergencyContactNotification(ctx context.Context, user *models.User, sosAlert *models.SOSAlert, ride *models.Ride) {
	if user.EmergencyContactPhone == "" {
		logger.Warn("no emergency contact phone available", "userID", user.ID)
		return
	}

	message := fmt.Sprintf(
		"EMERGENCY: %s has triggered an SOS alert. Location: https://maps.google.com/?q=%f,%f",
		user.Name,
		sosAlert.Latitude,
		sosAlert.Longitude,
	)

	if ride != nil {
		message += fmt.Sprintf("\nRide ID: %s\nPickup: %s\nDropoff: %s",
			ride.ID,
			ride.PickupAddress,
			ride.DropoffAddress,
		)
	}

	logger.Info("sending emergency contact notification",
		"userID", user.ID,
		"emergencyContactName", user.EmergencyContactName,
		"emergencyContactPhone", user.EmergencyContactPhone,
	)

	logger.Info("emergency contact SMS would be sent",
		"to", user.EmergencyContactPhone,
		"message", message,
	)
}

func sendSafetyTeamExternalAlerts(ctx context.Context, sosAlert *models.SOSAlert, payload map[string]interface{}) {
	logger.Info("sending external alerts to safety team", "sosAlertID", sosAlert.ID)
	logger.Info("external safety team alerts would be sent",
		"sosAlertID", sosAlert.ID,
		"severity", sosAlert.Severity,
	)
}

func NotifySOSResolved(ctx context.Context, sosAlert *models.SOSAlert, resolvedBy string) {
	if sosAlert == nil {
		return
	}

	resolvedPayload := map[string]interface{}{
		"sosAlertId": sosAlert.ID,
		"status":     "resolved",
		"resolvedBy": resolvedBy,
		"resolvedAt": time.Now().UTC(),
		"message":    "SOS alert has been resolved",
	}

	websocketutil.SendToUser(sosAlert.UserID, websocket.TypeSOSResolved, resolvedPayload)

	websocketutil.BroadcastToRole("safety_team", websocket.TypeSOSResolved, resolvedPayload)
	websocketutil.BroadcastToRole("admin", websocket.TypeSOSResolved, resolvedPayload)

	logger.Info("SOS resolved notifications sent",
		"sosAlertID", sosAlert.ID,
		"resolvedBy", resolvedBy,
	)
}