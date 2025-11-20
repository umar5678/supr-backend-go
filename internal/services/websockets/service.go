// internal/services/websocket/service.go
package websocketservice

import (
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/websocket"
)

// Service provides a simple interface for WebSocket operations
type Service struct {
	manager *websocket.Manager
}

// NewService creates a new WebSocket service
func NewService(manager *websocket.Manager) *Service {
	return &Service{
		manager: manager,
	}
}

// SendToUser sends a message to a specific user
func (s *Service) SendToUser(userID string, messageType websocket.MessageType, data map[string]interface{}) error {
	if data == nil {
		data = make(map[string]interface{})
	}

	msg := websocket.NewTargetedMessage(messageType, userID, data)
	s.manager.Hub().SendToUser(userID, msg)

	logger.Debug("websocket message sent to user",
		"userID", userID,
		"type", messageType,
	)
	return nil
}

// BroadcastToAll sends a message to all connected users
func (s *Service) BroadcastToAll(messageType websocket.MessageType, data map[string]interface{}) error {
	if data == nil {
		data = make(map[string]interface{})
	}

	msg := websocket.NewMessage(messageType, data)
	s.manager.Hub().BroadcastToAll(msg)

	logger.Debug("websocket message broadcasted",
		"type", messageType,
	)
	return nil
}

// SendNotification sends a notification to a user
func (s *Service) SendNotification(userID string, notification interface{}) error {
	return s.manager.SendNotification(userID, notification)
}

// IsUserOnline checks if a user is currently connected
func (s *Service) IsUserOnline(userID string) bool {
	return s.manager.Hub().IsUserConnected(userID)
}

// GetOnlineUsers returns statistics about connected users
func (s *Service) GetOnlineUsers() (int, int) {
	stats := s.manager.GetStats()
	return stats.ConnectedUsers, stats.TotalConnections
}

// Ride-specific methods

// SendRideRequest sends ride request to driver
func (s *Service) SendRideRequest(driverID string, rideDetails map[string]interface{}) error {
	return s.SendToUser(driverID, websocket.TypeRideRequest, rideDetails)
}

// SendRideAccepted notifies rider that driver accepted
func (s *Service) SendRideAccepted(riderID string, rideDetails map[string]interface{}) error {
	return s.SendToUser(riderID, websocket.TypeRideAccepted, rideDetails)
}

// SendRideLocationUpdate sends location update to rider
func (s *Service) SendRideLocationUpdate(riderID string, locationData map[string]interface{}) error {
	return s.SendToUser(riderID, websocket.TypeRideLocation, locationData)
}

// SendRideStatusUpdate sends status update to both rider and driver
func (s *Service) SendRideStatusUpdate(riderID, driverID string, statusData map[string]interface{}) error {
	// Send to rider
	if riderID != "" {
		s.SendToUser(riderID, websocket.TypeRideStatusUpdate, statusData)
	}

	// Send to driver
	if driverID != "" {
		s.SendToUser(driverID, websocket.TypeRideStatusUpdate, statusData)
	}

	return nil
}

// SendPaymentUpdate sends payment status update
func (s *Service) SendPaymentUpdate(userID string, paymentData map[string]interface{}) error {
	return s.SendToUser(userID, websocket.TypePaymentCompleted, paymentData)
}
