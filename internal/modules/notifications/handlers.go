package notifications

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/umar5678/go-backend/internal/models"
	notificationservice "github.com/umar5678/go-backend/internal/modules/notifications/service"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// RideEventHandler handles ride-related events
type RideEventHandler struct {
	pushService notificationservice.PushService
}

func NewRideEventHandler(pushService notificationservice.PushService) *RideEventHandler {
	return &RideEventHandler{
		pushService: pushService,
	}
}

func (h *RideEventHandler) EventType() EventType {
	return EventRideRequested // Base event type for ride events consumer
}

func (h *RideEventHandler) CanHandle(eventType EventType) bool {
	// Handle all ride-related events
	switch eventType {
	case EventRideRequestSent,
		EventRideAccepted,
		EventRideAssigned,
		EventRideStarted,
		EventRideCompleted,
		EventRideCancelled,
		EventRideRequested,
		EventRideRequestCancelledBySystem,
		EventRideRequestAlreadyAccepted,
		EventRideRequestAccepted,
		EventRideRequestRejected,
		EventRideRequestExpired,
		EventHighRiskRider,
		EventDriverArrived,
		EventInvalidRidePINAttempt,
		EventRideDestinationChanged,
		EventRideRouteUpdated,
		EventRideUpdated:
		return true
	default:
		return false
	}
}

func (h *RideEventHandler) Handle(ctx context.Context, event *ConsumedEvent) error {
	eventTypeStr := event.Headers["event_type"]

	// Parse the payload to get ride details
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		logger.Error("failed to unmarshal ride event payload", "error", err)
		return fmt.Errorf("failed to unmarshal ride event payload: %w", err)
	}

	// Check for rider_id or riderId (both formats supported)
	riderID, ok := payload["rider_id"].(string)
	if !ok {
		riderID, ok = payload["riderId"].(string)
	}
	if !ok {
		logger.Warn("missing rider_id/riderId in ride event", "event_type", eventTypeStr)
		return fmt.Errorf("missing rider_id in ride event")
	}

	riderUUID, err := uuid.Parse(riderID)
	if err != nil {
		logger.Error("invalid rider_id format", "error", err, "rider_id", riderID)
		return fmt.Errorf("invalid rider_id format: %w", err)
	}

	driverID, _ := payload["driver_id"].(string)
	if driverID == "" {
		driverID, _ = payload["driverId"].(string)
	}

	// Create notification message based on event type
	var notificationMsg string
	var metadataMap map[string]interface{}

	switch EventType(eventTypeStr) {
	case EventRideRequestSent:
		notificationMsg = "Your ride request has been sent"
		metadataMap = map[string]interface{}{
			"event_type": EventRideRequestSent,
			"ride_id":    payload["ride_id"],
			"fare":       payload["fare"],
		}

	case EventRideAccepted:
		notificationMsg = "Your ride has been accepted"
		metadataMap = map[string]interface{}{
			"event_type": EventRideAccepted,
			"ride_id":    payload["ride_id"],
			"driver_id":  driverID,
			"driver":     payload["driver"],
		}

	case EventRideAssigned:
		notificationMsg = "Your driver is arriving"
		metadataMap = map[string]interface{}{
			"event_type": EventRideAssigned,
			"ride_id":    payload["ride_id"],
			"driver_id":  driverID,
			"eta":        payload["eta"],
		}

	case EventRideStarted:
		notificationMsg = "Your ride has started"
		metadataMap = map[string]interface{}{
			"event_type": EventRideStarted,
			"ride_id":    payload["ride_id"],
		}

	case EventRideCompleted:
		notificationMsg = "Your ride is complete"
		metadataMap = map[string]interface{}{
			"event_type": EventRideCompleted,
			"ride_id":    payload["ride_id"],
			"fare":       payload["fare"],
		}

	case EventRideCancelled:
		notificationMsg = "Your ride has been cancelled"
		metadataMap = map[string]interface{}{
			"event_type": EventRideCancelled,
			"ride_id":    payload["ride_id"],
		}

	default:
		logger.Debug("unknown ride event type, skipping", "event_type", eventTypeStr)
		return nil
	}

	// Marshal metadata to JSON
	metadataJSON, err := json.Marshal(metadataMap)
	if err != nil {
		logger.Error("failed to marshal notification metadata", "error", err)
		return fmt.Errorf("failed to marshal notification metadata: %w", err)
	}

	// Send via push service
	notification := &models.Notification{
		UserID:   riderUUID,
		Title:    "Ride Update",
		Message:  notificationMsg,
		Channel:  models.ChannelInApp,
		Status:   models.NotificationStatusPending,
		Metadata: metadataJSON,
	}

	if err := h.pushService.SendPush(ctx, riderUUID, notification.Title, notification.Message, metadataMap); err != nil {
		logger.Error("failed to send push notification to rider", "error", err, "rider_id", riderID)
		// Don't return error - log and continue
	}

	// If driver exists, also notify them for certain events
	if driverID != "" {
		driverUUID, err := uuid.Parse(driverID)
		if err == nil {
			switch EventType(eventTypeStr) {
			case EventRideAccepted, EventRideStarted, EventRideCompleted:
				if err := h.pushService.SendPush(ctx, driverUUID, "Ride Update", notificationMsg, metadataMap); err != nil {
					logger.Error("failed to send push notification to driver", "error", err, "driver_id", driverID)
				}
			}
		}
	}

	return nil
}

// PaymentEventHandler handles payment-related events
type PaymentEventHandler struct {
	pushService notificationservice.PushService
}

func NewPaymentEventHandler(pushService notificationservice.PushService) *PaymentEventHandler {
	return &PaymentEventHandler{
		pushService: pushService,
	}
}

func (h *PaymentEventHandler) EventType() EventType {
	return EventPaymentProcessed // Base event type for payment events consumer
}

func (h *PaymentEventHandler) CanHandle(eventType EventType) bool {
	// Handle all payment-related events
	switch eventType {
	case EventPaymentProcessed,
		EventPaymentFailed,
		EventRefundIssued:
		return true
	default:
		return false
	}
}

func (h *PaymentEventHandler) Handle(ctx context.Context, event *ConsumedEvent) error {
	eventTypeStr := event.Headers["event_type"]

	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		logger.Error("failed to unmarshal payment event payload", "error", err)
		return fmt.Errorf("failed to unmarshal payment event payload: %w", err)
	}

	userID, ok := payload["user_id"].(string)
	if !ok {
		userID, ok = payload["userId"].(string)
	}
	if !ok {
		logger.Warn("missing user_id/userId in payment event", "event_type", eventTypeStr)
		return fmt.Errorf("missing user_id in payment event")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		logger.Error("invalid user_id format", "error", err, "user_id", userID)
		return fmt.Errorf("invalid user_id format: %w", err)
	}

	var notificationMsg string
	var metadataMap map[string]interface{}

	switch EventType(eventTypeStr) {
	case EventPaymentProcessed:
		notificationMsg = "Payment processed successfully"
		metadataMap = map[string]interface{}{
			"event_type": EventPaymentProcessed,
			"amount":     payload["amount"],
			"ride_id":    payload["ride_id"],
		}

	case EventPaymentFailed:
		notificationMsg = "Payment failed"
		metadataMap = map[string]interface{}{
			"event_type": EventPaymentFailed,
			"amount":     payload["amount"],
			"reason":     payload["reason"],
		}

	case EventRefundIssued:
		notificationMsg = "Refund issued"
		metadataMap = map[string]interface{}{
			"event_type": EventRefundIssued,
			"amount":     payload["amount"],
		}

	default:
		logger.Debug("unknown payment event type, skipping", "event_type", eventTypeStr)
		return nil
	}

	if err := h.pushService.SendPush(ctx, userUUID, "Payment Update", notificationMsg, metadataMap); err != nil {
		logger.Error("failed to send payment notification", "error", err, "user_id", userID)
	}

	return nil
}

// SOSEventHandler handles SOS-related events
type SOSEventHandler struct {
	pushService notificationservice.PushService
}

func NewSOSEventHandler(pushService notificationservice.PushService) *SOSEventHandler {
	return &SOSEventHandler{
		pushService: pushService,
	}
}

func (h *SOSEventHandler) EventType() EventType {
	return EventSOSAlert // Base event type for SOS events consumer
}

func (h *SOSEventHandler) CanHandle(eventType EventType) bool {
	// Handle all SOS-related events
	switch eventType {
	case EventSOSAlert,
		EventSOSTriggered,
		EventSOSResolved:
		return true
	default:
		return false
	}
}

func (h *SOSEventHandler) Handle(ctx context.Context, event *ConsumedEvent) error {
	eventTypeStr := event.Headers["event_type"]

	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		logger.Error("failed to unmarshal SOS event payload", "error", err)
		return fmt.Errorf("failed to unmarshal SOS event payload: %w", err)
	}

	riderID, ok := payload["rider_id"].(string)
	if !ok {
		riderID, ok = payload["riderId"].(string)
	}
	if !ok {
		logger.Warn("missing rider_id/riderId in SOS event", "event_type", eventTypeStr)
		return fmt.Errorf("missing rider_id in SOS event")
	}

	riderUUID, err := uuid.Parse(riderID)
	if err != nil {
		logger.Error("invalid rider_id format", "error", err, "rider_id", riderID)
		return fmt.Errorf("invalid rider_id format: %w", err)
	}

	var notificationMsg string
	var metadataMap map[string]interface{}

	switch EventType(eventTypeStr) {
	case EventSOSTriggered:
		notificationMsg = "SOS alert activated"
		metadataMap = map[string]interface{}{
			"event_type": EventSOSTriggered,
			"sos_id":     payload["sos_id"],
		}

	case EventSOSResolved:
		notificationMsg = "SOS alert resolved"
		metadataMap = map[string]interface{}{
			"event_type": EventSOSResolved,
			"sos_id":     payload["sos_id"],
		}

	default:
		logger.Debug("unknown SOS event type, skipping", "event_type", eventTypeStr)
		return nil
	}

	if err := h.pushService.SendPush(ctx, riderUUID, "Security Alert", notificationMsg, metadataMap); err != nil {
		logger.Error("failed to send SOS notification", "error", err, "rider_id", riderID)
	}

	return nil
}

// FraudEventHandler handles fraud detection events
type FraudEventHandler struct {
	pushService notificationservice.PushService
}

func NewFraudEventHandler(pushService notificationservice.PushService) *FraudEventHandler {
	return &FraudEventHandler{
		pushService: pushService,
	}
}

func (h *FraudEventHandler) EventType() EventType {
	return EventFraudPatternDetected // Base event type for fraud events consumer
}

func (h *FraudEventHandler) CanHandle(eventType EventType) bool {
	// Handle all fraud-related events
	switch eventType {
	case EventFraudPatternDetected,
		EventFraudAlertCreated,
		EventHighRiskRider:
		return true
	default:
		return false
	}
}

func (h *FraudEventHandler) Handle(ctx context.Context, event *ConsumedEvent) error {
	eventTypeStr := event.Headers["event_type"]

	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		logger.Error("failed to unmarshal fraud event payload", "error", err)
		return fmt.Errorf("failed to unmarshal fraud event payload: %w", err)
	}

	userID, ok := payload["user_id"].(string)
	if !ok {
		userID, ok = payload["userId"].(string)
	}
	if !ok {
		logger.Warn("missing user_id/userId in fraud event", "event_type", eventTypeStr)
		return fmt.Errorf("missing user_id in fraud event")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		logger.Error("invalid user_id format", "error", err, "user_id", userID)
		return fmt.Errorf("invalid user_id format: %w", err)
	}

	var notificationMsg string
	var metadataMap map[string]interface{}

	switch EventType(eventTypeStr) {
	case EventFraudPatternDetected:
		notificationMsg = "Suspicious activity detected on your account"
		metadataMap = map[string]interface{}{
			"event_type": EventFraudPatternDetected,
		}

	case EventHighRiskRider:
		notificationMsg = "Account flagged for verification"
		metadataMap = map[string]interface{}{
			"event_type": EventHighRiskRider,
		}

	default:
		logger.Debug("unknown fraud event type, skipping", "event_type", eventTypeStr)
		return nil
	}

	if err := h.pushService.SendPush(ctx, userUUID, "Account Security", notificationMsg, metadataMap); err != nil {
		logger.Error("failed to send fraud notification", "error", err, "user_id", userID)
	}

	return nil
}

// UserEventHandler handles user-related events
type UserEventHandler struct {
	pushService notificationservice.PushService
}

func NewUserEventHandler(pushService notificationservice.PushService) *UserEventHandler {
	return &UserEventHandler{
		pushService: pushService,
	}
}

func (h *UserEventHandler) EventType() EventType {
	return EventUserRegistered // Base event type for user events consumer
}

func (h *UserEventHandler) CanHandle(eventType EventType) bool {
	// Handle all user-related events
	switch eventType {
	case EventUserRegistered,
		EventUserVerified,
		EventUserSuspended:
		return true
	default:
		return false
	}
}

func (h *UserEventHandler) Handle(ctx context.Context, event *ConsumedEvent) error {
	eventTypeStr := event.Headers["event_type"]

	var payload map[string]interface{}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		logger.Error("failed to unmarshal user event payload", "error", err)
		return fmt.Errorf("failed to unmarshal user event payload: %w", err)
	}

	userID, ok := payload["user_id"].(string)
	if !ok {
		userID, ok = payload["userId"].(string)
	}
	if !ok {
		logger.Warn("missing user_id/userId in user event", "event_type", eventTypeStr)
		return fmt.Errorf("missing user_id in user event")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		logger.Error("invalid user_id format", "error", err, "user_id", userID)
		return fmt.Errorf("invalid user_id format: %w", err)
	}

	var notificationMsg string
	var metadataMap map[string]interface{}

	switch EventType(eventTypeStr) {
	case EventUserVerified:
		notificationMsg = "Your account has been verified"
		metadataMap = map[string]interface{}{
			"event_type": EventUserVerified,
		}

	case EventUserSuspended:
		notificationMsg = "Your account has been suspended"
		metadataMap = map[string]interface{}{
			"event_type": EventUserSuspended,
		}

	default:
		logger.Debug("unknown user event type, skipping", "event_type", eventTypeStr)
		return nil
	}

	if err := h.pushService.SendPush(ctx, userUUID, "Account Update", notificationMsg, metadataMap); err != nil {
		logger.Error("failed to send user notification", "error", err, "user_id", userID)
	}

	return nil
}
