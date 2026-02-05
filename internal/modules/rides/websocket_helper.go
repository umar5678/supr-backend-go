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

type RideWebSocketHelper struct {
	service Service
}

func NewRideWebSocketHelper() *RideWebSocketHelper {
	return &RideWebSocketHelper{}
}

func (h *RideWebSocketHelper) SetService(service Service) {
	h.service = service
}

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

func (h *RideWebSocketHelper) CheckUserOnline(userID string) bool {
	return websocketutil.IsUserOnline(userID)
}

func (h *RideWebSocketHelper) SendRideRequest(driverID string, rideDetails map[string]interface{}) error {
	return websocketutil.SendRideRequest(driverID, rideDetails)
}

func (h *RideWebSocketHelper) SendRideAccepted(riderID string, rideDetails map[string]interface{}) error {
	return websocketutil.SendRideAccepted(riderID, rideDetails)
}

func (h *RideWebSocketHelper) HandleAvailableCarsStream(conn *websocket.Conn, riderID string) error {
	if h.service == nil {
		logger.Error("service not initialized for available cars handler")
		return websocket.ErrCloseSent
	}

	defer conn.Close()

	ctx := context.Background()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var currentRequest *dto.AvailableCarRequest
	var updateInterval time.Duration = 5 * time.Second
	var lastUpdateTime time.Time

	for {
		select {
		case <-ticker.C:
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
					req.RadiusKm = 5.0
				}

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
				msg := dto.WebSocketAvailableCarsMessage{
					Type:      "subscribed",
					Timestamp: time.Now(),
				}
				if err := conn.WriteJSON(msg); err != nil {
					logger.Warn("failed to send subscription confirmation", "error", err)
					return err
				}

			case "update_location":
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
