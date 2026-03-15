package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// PushService interface for push notifications (device agnostic)
type PushService interface {
	SendPush(ctx context.Context, userID uuid.UUID, title, body string, data map[string]interface{}) error
	RegisterToken(ctx context.Context, userID uuid.UUID, token, deviceID, deviceOS string) error
	UnregisterToken(ctx context.Context, token string) error
	GetPendingMessages(ctx context.Context, userID uuid.UUID) ([]PushMessage, error)
	GetUserTokens(ctx context.Context, userID uuid.UUID) ([]PushToken, error)
	SubscribeToUser(userID uuid.UUID, subscriberID string, msgChan chan PushMessage) error
	UnsubscribeFromUser(userID uuid.UUID, subscriberID string) error
}

// SendSecurityAlert sends a security alert push notification
func SendSecurityAlert(ctx context.Context, svc PushService, userID uuid.UUID, patternID string, riskScore float64) error {
	data := map[string]interface{}{
		"type":       "security_alert",
		"pattern_id": patternID,
		"risk_score": fmt.Sprintf("%.0f", riskScore),
	}

	body := "Unusual activity detected on your account"
	if riskScore > 80 {
		body = "⚠️ High-risk suspicious activity detected on your account"
	}

	if err := svc.SendPush(ctx, userID, "Security Alert", body, data); err != nil {
		logger.Error("failed to send security alert", "error", err, "userID", userID)
		return err
	}

	return nil
}

// SendRideCompleteNotification sends a ride completion push notification
func SendRideCompleteNotification(ctx context.Context, svc PushService, userID uuid.UUID, rideID string, finalFare float64) error {
	data := map[string]interface{}{
		"type":    "ride_complete",
		"ride_id": rideID,
		"fare":    fmt.Sprintf("%.2f", finalFare),
	}

	body := fmt.Sprintf("Your ride is complete. Total: ₹%.2f", finalFare)

	if err := svc.SendPush(ctx, userID, "Ride Completed", body, data); err != nil {
		logger.Error("failed to send ride complete notification", "error", err, "userID", userID)
		return err
	}

	return nil
}

// SendRideAcceptedNotification sends a ride accepted push notification
func SendRideAcceptedNotification(ctx context.Context, svc PushService, userID uuid.UUID, rideID, driverName string, eta int) error {
	data := map[string]interface{}{
		"type":        "ride_accepted",
		"ride_id":     rideID,
		"driver_name": driverName,
		"eta":         eta,
	}

	body := fmt.Sprintf("%s accepted your ride. ETA: %d min", driverName, eta)

	if err := svc.SendPush(ctx, userID, "Driver Assigned", body, data); err != nil {
		logger.Error("failed to send ride accepted notification", "error", err, "userID", userID)
		return err
	}

	return nil
}

// SendPaymentNotification sends a payment notification
func SendPaymentNotification(ctx context.Context, svc PushService, userID uuid.UUID, amount float64, status string) error {
	title := "Payment Processed"
	if status == "failed" {
		title = "Payment Failed"
	}

	data := map[string]interface{}{
		"type":   "payment",
		"amount": fmt.Sprintf("%.2f", amount),
		"status": status,
	}

	body := fmt.Sprintf("Payment of ₹%.2f %s", amount, status)

	if err := svc.SendPush(ctx, userID, title, body, data); err != nil {
		logger.Error("failed to send payment notification", "error", err, "userID", userID)
		return err
	}

	return nil
}
