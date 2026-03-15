package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/umar5678/go-backend/internal/modules/notifications"
	"github.com/umar5678/go-backend/internal/modules/notifications/service"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// EventHandlerFactory creates and routes event handlers
type EventHandlerFactory struct {
	pushService service.PushService
}

func NewEventHandlerFactory(pushService service.PushService) *EventHandlerFactory {
	return &EventHandlerFactory{
		pushService: pushService,
	}
}

// HandleEvent routes events to appropriate handlers
func (f *EventHandlerFactory) HandleEvent(ctx context.Context, eventType notifications.EventType, payload []byte) error {
	switch eventType {
	// Fraud events
	case notifications.EventFraudPatternDetected:
		return f.handleFraudPatternDetected(ctx, payload)

	// Ride events
	case notifications.EventRideRequested:
		return f.handleRideRequested(ctx, payload)
	case notifications.EventRideAccepted:
		return f.handleRideAccepted(ctx, payload)
	case notifications.EventRideStarted:
		return f.handleRideStarted(ctx, payload)
	case notifications.EventRideCompleted:
		return f.handleRideCompleted(ctx, payload)
	case notifications.EventRideCancelled:
		return f.handleRideCancelled(ctx, payload)

	// Payment events
	case notifications.EventPaymentProcessed:
		return f.handlePaymentProcessed(ctx, payload)
	case notifications.EventPaymentFailed:
		return f.handlePaymentFailed(ctx, payload)

	// Vehicle events
	case notifications.EventVehicleRegistered:
		return f.handleVehicleRegistered(ctx, payload)

	// User events
	case notifications.EventUserRegistered:
		return f.handleUserRegistered(ctx, payload)
	case notifications.EventUserVerified:
		return f.handleUserVerified(ctx, payload)

	default:
		logger.Warn("unknown event type", "type", eventType)
		return fmt.Errorf("unknown event type: %s", eventType)
	}
}

// ============ FRAUD HANDLERS ============

type FraudPatternDetectedPayload struct {
	PatternID   uuid.UUID `json:"pattern_id"`
	UserID      uuid.UUID `json:"user_id"`
	RiskScore   float64   `json:"risk_score"`
	PatternType string    `json:"pattern_type"`
}

func (f *EventHandlerFactory) handleFraudPatternDetected(ctx context.Context, payload []byte) error {
	var event FraudPatternDetectedPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal fraud event: %w", err)
	}

	// Send push notification if risk score is high
	if event.RiskScore > 50 {
		if err := service.SendSecurityAlert(ctx, f.pushService, event.UserID, event.PatternID.String(), event.RiskScore); err != nil {
			logger.Error("failed to send security alert", "error", err)
			// Don't fail the entire event if notification fails
		}
	}

	logger.Info("fraud pattern notification sent", "userID", event.UserID, "riskScore", event.RiskScore)
	return nil
}

// ============ RIDE HANDLERS ============

type RideRequestedPayload struct {
	RideID        uuid.UUID `json:"ride_id"`
	RiderID       uuid.UUID `json:"rider_id"`
	PickupLat     float64   `json:"pickup_lat"`
	PickupLon     float64   `json:"pickup_lon"`
	DropoffLat    float64   `json:"dropoff_lat"`
	DropoffLon    float64   `json:"dropoff_lon"`
	EstimatedFare float64   `json:"estimated_fare"`
}

func (f *EventHandlerFactory) handleRideRequested(ctx context.Context, payload []byte) error {
	var event RideRequestedPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal ride requested event: %w", err)
	}

	data := map[string]interface{}{
		"type":    "ride_requested",
		"ride_id": event.RideID.String(),
		"fare":    fmt.Sprintf("%.2f", event.EstimatedFare),
	}

	if err := f.pushService.SendPush(ctx, event.RiderID, "Ride Requested", "Your ride request has been sent to nearby drivers", data); err != nil {
		logger.Error("failed to send ride requested notification", "error", err)
	}

	return nil
}

type RideAcceptedPayload struct {
	RideID     uuid.UUID `json:"ride_id"`
	RiderID    uuid.UUID `json:"rider_id"`
	DriverID   uuid.UUID `json:"driver_id"`
	DriverName string    `json:"driver_name"`
	ETA        int       `json:"eta"` // in seconds
}

func (f *EventHandlerFactory) handleRideAccepted(ctx context.Context, payload []byte) error {
	var event RideAcceptedPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal ride accepted event: %w", err)
	}

	etaMin := event.ETA / 60
	if err := service.SendRideAcceptedNotification(ctx, f.pushService, event.RiderID, event.RideID.String(), event.DriverName, etaMin); err != nil {
		logger.Error("failed to send ride accepted notification", "error", err)
	}

	return nil
}

type RideStartedPayload struct {
	RideID   uuid.UUID `json:"ride_id"`
	RiderID  uuid.UUID `json:"rider_id"`
	DriverID uuid.UUID `json:"driver_id"`
}

func (f *EventHandlerFactory) handleRideStarted(ctx context.Context, payload []byte) error {
	var event RideStartedPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal ride started event: %w", err)
	}

	data := map[string]interface{}{
		"type":    "ride_started",
		"ride_id": event.RideID.String(),
	}

	if err := f.pushService.SendPush(ctx, event.RiderID, "Ride Started", "Your ride has started", data); err != nil {
		logger.Error("failed to send ride started notification", "error", err)
	}

	return nil
}

type RideCompletedPayload struct {
	RideID         uuid.UUID `json:"ride_id"`
	RiderID        uuid.UUID `json:"rider_id"`
	DriverID       uuid.UUID `json:"driver_id"`
	FinalFare      float64   `json:"final_fare"`
	ActualDistance float64   `json:"actual_distance"`
}

func (f *EventHandlerFactory) handleRideCompleted(ctx context.Context, payload []byte) error {
	var event RideCompletedPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal ride completed event: %w", err)
	}

	// Send to rider
	if err := service.SendRideCompleteNotification(ctx, f.pushService, event.RiderID, event.RideID.String(), event.FinalFare); err != nil {
		logger.Error("failed to send ride completed notification to rider", "error", err)
	}

	// Send to driver
	data := map[string]interface{}{
		"type":    "ride_completed",
		"ride_id": event.RideID.String(),
		"earning": fmt.Sprintf("%.2f", event.FinalFare),
	}

	if err := f.pushService.SendPush(ctx, event.DriverID, "Ride Completed", fmt.Sprintf("Ride completed. Your earning: ₹%.2f", event.FinalFare), data); err != nil {
		logger.Error("failed to send ride completed notification to driver", "error", err)
	}

	return nil
}

type RideCancelledPayload struct {
	RideID   uuid.UUID  `json:"ride_id"`
	RiderID  uuid.UUID  `json:"rider_id"`
	DriverID *uuid.UUID `json:"driver_id,omitempty"`
	Reason   string     `json:"reason"`
}

func (f *EventHandlerFactory) handleRideCancelled(ctx context.Context, payload []byte) error {
	var event RideCancelledPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal ride cancelled event: %w", err)
	}

	data := map[string]interface{}{
		"type":    "ride_cancelled",
		"ride_id": event.RideID.String(),
		"reason":  event.Reason,
	}

	// Notify rider
	if err := f.pushService.SendPush(ctx, event.RiderID, "Ride Cancelled", fmt.Sprintf("Your ride has been cancelled. Reason: %s", event.Reason), data); err != nil {
		logger.Error("failed to send ride cancelled notification to rider", "error", err)
	}

	// Notify driver if assigned
	if event.DriverID != nil {
		if err := f.pushService.SendPush(ctx, *event.DriverID, "Ride Cancelled", fmt.Sprintf("A ride has been cancelled. Reason: %s", event.Reason), data); err != nil {
			logger.Error("failed to send ride cancelled notification to driver", "error", err)
		}
	}

	return nil
}

// ============ PAYMENT HANDLERS ============

type PaymentProcessedPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Amount float64   `json:"amount"`
	RideID uuid.UUID `json:"ride_id"`
}

func (f *EventHandlerFactory) handlePaymentProcessed(ctx context.Context, payload []byte) error {
	var event PaymentProcessedPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal payment processed event: %w", err)
	}

	if err := service.SendPaymentNotification(ctx, f.pushService, event.UserID, event.Amount, "success"); err != nil {
		logger.Error("failed to send payment notification", "error", err)
	}

	return nil
}

type PaymentFailedPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Amount float64   `json:"amount"`
	Reason string    `json:"reason"`
}

func (f *EventHandlerFactory) handlePaymentFailed(ctx context.Context, payload []byte) error {
	var event PaymentFailedPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal payment failed event: %w", err)
	}

	if err := service.SendPaymentNotification(ctx, f.pushService, event.UserID, event.Amount, "failed"); err != nil {
		logger.Error("failed to send payment failed notification", "error", err)
	}

	return nil
}

// ============ VEHICLE HANDLERS ============

type VehicleRegisteredPayload struct {
	VehicleID   uuid.UUID `json:"vehicle_id"`
	UserID      uuid.UUID `json:"user_id"`
	VehicleName string    `json:"vehicle_name"`
}

func (f *EventHandlerFactory) handleVehicleRegistered(ctx context.Context, payload []byte) error {
	var event VehicleRegisteredPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal vehicle registered event: %w", err)
	}

	data := map[string]interface{}{
		"type":       "vehicle_registered",
		"vehicle_id": event.VehicleID.String(),
	}

	if err := f.pushService.SendPush(ctx, event.UserID, "Vehicle Registered", fmt.Sprintf("Your %s has been registered successfully", event.VehicleName), data); err != nil {
		logger.Error("failed to send vehicle registered notification", "error", err)
	}

	return nil
}

// ============ USER HANDLERS ============

type UserRegisteredPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
	Role   string    `json:"role"`
}

func (f *EventHandlerFactory) handleUserRegistered(ctx context.Context, payload []byte) error {
	var event UserRegisteredPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal user registered event: %w", err)
	}

	data := map[string]interface{}{
		"type": "user_registered",
		"role": event.Role,
	}

	greeting := fmt.Sprintf("Welcome to Ghartak, %s!", event.Name)
	if err := f.pushService.SendPush(ctx, event.UserID, "Welcome!", greeting, data); err != nil {
		logger.Error("failed to send welcome notification", "error", err)
	}

	return nil
}

type UserVerifiedPayload struct {
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
}

func (f *EventHandlerFactory) handleUserVerified(ctx context.Context, payload []byte) error {
	var event UserVerifiedPayload
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal user verified event: %w", err)
	}

	data := map[string]interface{}{
		"type": "user_verified",
	}

	if err := f.pushService.SendPush(ctx, event.UserID, "Account Verified", "Your account has been verified successfully!", data); err != nil {
		logger.Error("failed to send account verified notification", "error", err)
	}

	return nil
}

// EventHandlerAdapter wraps EventHandlerFactory to implement the consumer's EventHandler interface
// This allows the factory to be registered as a handler with the Kafka consumer
type EventHandlerAdapter struct {
	factory   *EventHandlerFactory
	eventType notifications.EventType
}

// NewEventHandlerAdapter creates a new adapter for a specific event type
func NewEventHandlerAdapter(factory *EventHandlerFactory, eventType notifications.EventType) *EventHandlerAdapter {
	return &EventHandlerAdapter{
		factory:   factory,
		eventType: eventType,
	}
}

// Handle implements the EventHandler interface
func (a *EventHandlerAdapter) Handle(ctx context.Context, event *notifications.ConsumedEvent) error {
	// Route to the factory's HandleEvent method
	return a.factory.HandleEvent(ctx, event.EventType, event.Payload)
}

// EventType implements the EventHandler interface
func (a *EventHandlerAdapter) EventType() notifications.EventType {
	return a.eventType
}
