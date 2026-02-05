// internal/services/websocket/service.go
package websocketservice

import (
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/websocket"
)

type Service struct {
	manager *websocket.Manager
}

func NewService(manager *websocket.Manager) *Service {
	return &Service{
		manager: manager,
	}
}

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

func (s *Service) SendNotification(userID string, notification interface{}) error {
	return s.manager.SendNotification(userID, notification)
}

func (s *Service) IsUserOnline(userID string) bool {
	return s.manager.Hub().IsUserConnected(userID)
}

func (s *Service) GetOnlineUsers() (int, int) {
	stats := s.manager.GetStats()
	return stats.ConnectedUsers, stats.TotalConnections
}

func (s *Service) SendRideRequest(driverID string, rideDetails map[string]interface{}) error {
	return s.SendToUser(driverID, websocket.TypeRideRequest, rideDetails)
}

func (s *Service) SendRideAccepted(riderID string, rideDetails map[string]interface{}) error {
	return s.SendToUser(riderID, websocket.TypeRideAccepted, rideDetails)
}

func (s *Service) SendRideLocationUpdate(riderID string, locationData map[string]interface{}) error {
	return s.SendToUser(riderID, websocket.TypeRideLocation, locationData)
}

func (s *Service) SendRideStatusUpdate(riderID, driverID string, statusData map[string]interface{}) error {
	if riderID != "" {
		s.SendToUser(riderID, websocket.TypeRideStatusUpdate, statusData)
	}

	if driverID != "" {
		s.SendToUser(driverID, websocket.TypeRideStatusUpdate, statusData)
	}

	return nil
}

func (s *Service) SendPaymentUpdate(userID string, paymentData map[string]interface{}) error {
	return s.SendToUser(userID, websocket.TypePaymentCompleted, paymentData)
}
