package websocket

import "time"

const (

	TypeRideAccepted MessageType = "ride_accepted"
	TypeRideRejected MessageType = "ride_rejected"
	TypeRideLocation MessageType = "ride_location"

	TypeDriverAvailable   MessageType = "driver_available"
	TypeDriverUnavailable MessageType = "driver_unavailable"
	TypeDriverLocation    MessageType = "driver_location"

	TypePaymentInitiated MessageType = "payment_initiated"
	TypePaymentCompleted MessageType = "payment_completed"
	TypePaymentFailed    MessageType = "payment_failed"
)

func NewRideMessage(msgType MessageType, rideID string, data map[string]interface{}) *Message {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["rideId"] = rideID
	return NewMessage(msgType, data)
}

func NewRideLocationMessage(rideID string, location map[string]interface{}) *Message {
	return NewRideMessage(TypeRideLocation, rideID, map[string]interface{}{
		"location":  location,
		"timestamp": time.Now().UTC(),
	})
}

func NewRideStatusMessage(rideID, status string, additionalData map[string]interface{}) *Message {
	data := map[string]interface{}{
		"status": status,
	}
	for k, v := range additionalData {
		data[k] = v
	}
	return NewRideMessage(TypeRideStatusUpdate, rideID, data)
}
