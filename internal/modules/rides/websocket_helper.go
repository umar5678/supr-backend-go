// internal/modules/rides/websocket_helper.go
package rides

import (
	"context"
	"time"

	"github.com/umar5678/go-backend/internal/utils/logger"
	websocketutil "github.com/umar5678/go-backend/internal/websocket/websocketutils"
)

// RideWebSocketHelper provides ride-specific WebSocket functionality
type RideWebSocketHelper struct{}

func NewRideWebSocketHelper() *RideWebSocketHelper {
	return &RideWebSocketHelper{}
}

// SendDriverLocationUpdate sends real-time driver location to rider
func (h *RideWebSocketHelper) SendDriverLocationUpdate(ctx context.Context, riderID, rideID string, location map[string]interface{}) {
	locationData := map[string]interface{}{
		"rideId":    rideID,
		"location":  location,
		"timestamp": time.Now().UTC(), // Always use current timestamp
	}

	if err := websocketutil.SendRideLocationUpdate(riderID, locationData); err != nil {
		logger.Error("failed to send driver location update",
			"error", err,
			"riderID", riderID,
			"rideID", rideID,
		)
	}
}

// SendRideStatusToBoth sends status update to both rider and driver
func (h *RideWebSocketHelper) SendRideStatusToBoth(ctx context.Context, riderID, driverID, rideID, status, message string) {
	statusData := map[string]interface{}{
		"rideId":    rideID,
		"status":    status,
		"message":   message,
		"timestamp": time.Now().UTC(),
	}

	if err := websocketutil.SendRideStatusUpdate(riderID, driverID, statusData); err != nil {
		logger.Error("failed to send ride status update",
			"error", err,
			"rideID", rideID,
			"status", status,
		)
	}
}

// CheckUserOnline checks if a user is currently connected via WebSocket
func (h *RideWebSocketHelper) CheckUserOnline(userID string) bool {
	return websocketutil.IsUserOnline(userID)
}

// SendRideRequest sends ride request to driver
func (h *RideWebSocketHelper) SendRideRequest(driverID string, rideDetails map[string]interface{}) error {
	return websocketutil.SendRideRequest(driverID, rideDetails)
}

// SendRideAccepted sends ride acceptance to rider
func (h *RideWebSocketHelper) SendRideAccepted(riderID string, rideDetails map[string]interface{}) error {
	return websocketutil.SendRideAccepted(riderID, rideDetails)
}
