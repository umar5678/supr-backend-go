// internal/modules/rides/websocket_helper.go
package rides

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"github.com/umar5678/go-backend/internal/modules/rides/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
	websocketutil "github.com/umar5678/go-backend/internal/websocket/websocketutils"
)

// RideWebSocketHelper provides ride-specific WebSocket functionality
type RideWebSocketHelper struct {
	service Service
}

func NewRideWebSocketHelper() *RideWebSocketHelper {
	return &RideWebSocketHelper{}
}

// SetService sets the rides service for WebSocket operations
func (h *RideWebSocketHelper) SetService(service Service) {
	h.service = service
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

// ============================================================================
// Available Cars WebSocket Handler
// ============================================================================

// HandleAvailableCarsStream handles WebSocket connections for streaming available cars
// This allows riders to see real-time updates of available cars near them
func (h *RideWebSocketHelper) HandleAvailableCarsStream(conn *websocket.Conn, riderID string) error {
	if h.service == nil {
		logger.Error("service not initialized for available cars handler")
		return websocket.ErrCloseSent
	}

	defer conn.Close()

	ctx := context.Background()
	ticker := time.NewTicker(5 * time.Second) // Default update interval
	defer ticker.Stop()

	var currentRequest *dto.AvailableCarRequest
	var updateInterval time.Duration = 5 * time.Second
	var lastUpdateTime time.Time

	for {
		select {
		case <-ticker.C:
			// Auto-refresh available cars at specified interval
			if currentRequest != nil && time.Since(lastUpdateTime) >= updateInterval {
				cars, err := h.service.GetAvailableCars(ctx, riderID, *currentRequest)
				if err != nil {
					logger.Error("failed to fetch available cars", "error", err, "riderID", riderID)
					msg := dto.WebSocketAvailableCarsMessage{
						Type:      "error",
						Error:     "Failed to fetch cars: " + err.Error(),
						Timestamp: time.Now(),
					}
					if err := conn.WriteJSON(msg); err != nil {
						logger.Warn("failed to write error to websocket", "error", err)
						return err
					}
					continue
				}

				// Send update to client
				msg := dto.WebSocketAvailableCarsMessage{
					Type:      "cars_update",
					Data:      cars,
					Timestamp: time.Now(),
				}

				if err := conn.WriteJSON(msg); err != nil {
					logger.Warn("failed to write cars update to websocket", "error", err)
					return err
				}

				lastUpdateTime = time.Now()
				logger.Info("sent available cars update", "riderID", riderID, "carsCount", cars.CarsCount)
			}

		default:
			// Read incoming messages from client
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))

			var incomingMsg map[string]interface{}
			err := conn.ReadJSON(&incomingMsg)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Error("WebSocket error", "error", err)
				}
				return err
			}

			msgType, ok := incomingMsg["type"].(string)
			if !ok {
				continue
			}

			switch msgType {
			case "subscribe_available_cars":
				// Parse request parameters
				req := &dto.AvailableCarRequest{}

				if lat, ok := incomingMsg["latitude"].(float64); ok {
					req.Latitude = lat
				} else {
					msg := dto.WebSocketAvailableCarsMessage{
						Type:      "error",
						Error:     "Missing or invalid latitude",
						Timestamp: time.Now(),
					}
					conn.WriteJSON(msg)
					continue
				}

				if lon, ok := incomingMsg["longitude"].(float64); ok {
					req.Longitude = lon
				} else {
					msg := dto.WebSocketAvailableCarsMessage{
						Type:      "error",
						Error:     "Missing or invalid longitude",
						Timestamp: time.Now(),
					}
					conn.WriteJSON(msg)
					continue
				}

				if radius, ok := incomingMsg["radiusKm"].(float64); ok && radius > 0 {
					req.RadiusKm = radius
				} else {
					req.RadiusKm = 5.0 // Default
				}

				// Set update interval if provided
				if interval, ok := incomingMsg["updateIntervalSeconds"].(float64); ok && interval > 0 {
					updateInterval = time.Duration(int(interval)) * time.Second
					ticker.Stop()
					ticker = time.NewTicker(updateInterval)
				}

				currentRequest = req
				lastUpdateTime = time.Now().Add(-updateInterval) // Force immediate first update

				logger.Info("subscribed to available cars",
					"riderID", riderID,
					"latitude", req.Latitude,
					"longitude", req.Longitude,
					"radiusKm", req.RadiusKm,
					"updateInterval", updateInterval.Seconds())

				// Send subscription confirmation
				msg := dto.WebSocketAvailableCarsMessage{
					Type:      "subscribed",
					Timestamp: time.Now(),
				}
				if err := conn.WriteJSON(msg); err != nil {
					logger.Warn("failed to send subscription confirmation", "error", err)
					return err
				}

			case "update_location":
				// Update rider location for available cars search
				if lat, ok := incomingMsg["latitude"].(float64); ok {
					if currentRequest != nil {
						currentRequest.Latitude = lat
					}
				}

				if lon, ok := incomingMsg["longitude"].(float64); ok {
					if currentRequest != nil {
						currentRequest.Longitude = lon
					}
				}

				// Force immediate update
				if currentRequest != nil {
					lastUpdateTime = time.Now().Add(-updateInterval)
				}

				logger.Info("location updated", "riderID", riderID)

			case "unsubscribe":
				logger.Info("unsubscribed from available cars", "riderID", riderID)
				msg := dto.WebSocketAvailableCarsMessage{
					Type:      "unsubscribed",
					Timestamp: time.Now(),
				}
				conn.WriteJSON(msg)
				return nil

			case "ping":
				// Respond to ping with pong
				msg := dto.WebSocketAvailableCarsMessage{
					Type:      "pong",
					Timestamp: time.Now(),
				}
				conn.WriteJSON(msg)

			default:
				logger.Warn("unknown message type", "type", msgType)
			}
		}
	}
}
