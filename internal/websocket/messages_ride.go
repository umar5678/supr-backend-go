// internal/websocket/messages_ride.go
package websocket

import "time"

// Ride-specific message types
const (
	// Ride Management
	TypeRideRequested    MessageType = "ride_requested"
	TypeRideAccepted     MessageType = "ride_accepted"
	TypeRideRejected     MessageType = "ride_rejected"
	TypeRideStarted      MessageType = "ride_started"
	TypeRideCompleted    MessageType = "ride_completed"
	TypeRideCancelled    MessageType = "ride_cancelled"
	TypeRideLocation     MessageType = "ride_location"
	TypeRideStatusUpdate MessageType = "ride_status_update"

	// Driver Management
	TypeDriverAvailable   MessageType = "driver_available"
	TypeDriverUnavailable MessageType = "driver_unavailable"
	TypeDriverLocation    MessageType = "driver_location"

	// Payment
	TypePaymentInitiated MessageType = "payment_initiated"
	TypePaymentCompleted MessageType = "payment_completed"
	TypePaymentFailed    MessageType = "payment_failed"
)

// NewRideMessage creates a ride-specific message
func NewRideMessage(msgType MessageType, rideID string, data map[string]interface{}) *Message {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["rideId"] = rideID
	return NewMessage(msgType, data)
}

// NewRideLocationMessage creates a ride location update message
func NewRideLocationMessage(rideID string, location map[string]interface{}) *Message {
	return NewRideMessage(TypeRideLocation, rideID, map[string]interface{}{
		"location":  location,
		"timestamp": time.Now().UTC(),
	})
}

// NewRideStatusMessage creates a ride status update message
func NewRideStatusMessage(rideID, status string, additionalData map[string]interface{}) *Message {
	data := map[string]interface{}{
		"status": status,
	}
	for k, v := range additionalData {
		data[k] = v
	}
	return NewRideMessage(TypeRideStatusUpdate, rideID, data)
}
