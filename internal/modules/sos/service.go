package sos

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	notificationsmodule "github.com/umar5678/go-backend/internal/modules/notifications"
	"github.com/umar5678/go-backend/internal/modules/sos/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
	"github.com/umar5678/go-backend/internal/websocket"
	websocketutil "github.com/umar5678/go-backend/internal/websocket/websocketutils"
	"gorm.io/gorm"
)

type Service interface {
	TriggerSOS(ctx context.Context, userID string, req dto.TriggerSOSRequest) (*dto.SOSAlertResponse, error)
	GetSOS(ctx context.Context, userID, alertID string, isAdmin bool) (*dto.SOSAlertResponse, error)
	GetActiveSOS(ctx context.Context, userID string) (*dto.SOSAlertResponse, error)
	ListSOS(ctx context.Context, userID string, req dto.ListSOSRequest) ([]*dto.SOSAlertListResponse, int64, error)
	ResolveSOS(ctx context.Context, userID, alertID string, req dto.ResolveSOSRequest) (*dto.SOSAlertResponse, error)
	CancelSOS(ctx context.Context, userID, alertID string) (*dto.SOSAlertResponse, error)
	UpdateSOSLocation(ctx context.Context, userID, alertID string, latitude, longitude float64) (*dto.SOSAlertResponse, error)
}

type service struct {
	repo          Repository
	userDB        *gorm.DB
	eventProducer notificationsmodule.EventProducer
}

func NewService(repo Repository, db *gorm.DB) Service {
	return NewServiceWithNotifications(repo, db, nil)
}

func NewServiceWithNotifications(repo Repository, db *gorm.DB, eventProducer notificationsmodule.EventProducer) Service {
	return &service{
		repo:          repo,
		userDB:        db,
		eventProducer: eventProducer,
	}
}

func (s *service) TriggerSOS(ctx context.Context, userID string, req dto.TriggerSOSRequest) (*dto.SOSAlertResponse, error) {
	if userID == "" {
		return nil, response.BadRequest("User ID is required")
	}

	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	existingAlert, err := s.repo.FindActiveByUserID(ctx, userID)
	if err == nil && existingAlert != nil && existingAlert.Status == "active" {
		return nil, response.BadRequest("You already have an active SOS alert")
	}

	now := time.Now()
	alert := &models.SOSAlert{
		UserID:      userID,
		RideID:      req.RideID,
		AlertType:   "manual",
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Status:      "active",
		Severity:    "critical",
		TriggeredAt: now,
	}

	if err := s.repo.Create(ctx, alert); err != nil {
		logger.Error("failed to create SOS alert", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to trigger SOS", err)
	}

	alert, err = s.repo.FindByID(ctx, alert.ID)
	if err != nil {
		logger.Error("failed to fetch created SOS alert", "error", err, "alertID", alert.ID)
		return nil, response.InternalServerError("Failed to fetch created alert", err)
	}

	go s.notifyEmergencyContacts(context.Background(), alert)
	go s.notifySafetyTeam(context.Background(), alert)

	logger.Warn("SOS ALERT TRIGGERED",
		"alertID", alert.ID,
		"userID", userID,
		"rideID", req.RideID,
		"location", fmt.Sprintf("%f,%f", req.Latitude, req.Longitude),
	)

	// Publish SOS alert event - use background context to ensure notification reaches admins
	s.publishSOSEvent(context.Background(), notificationsmodule.EventSOSAlert, alert.ID, userID, map[string]interface{}{
		"rideID":    req.RideID,
		"latitude":  req.Latitude,
		"longitude": req.Longitude,
		"severity":  "critical",
	})

	return dto.ToSOSAlertResponse(alert), nil
}

func (s *service) GetSOS(ctx context.Context, userID, alertID string, isAdmin bool) (*dto.SOSAlertResponse, error) {
	if alertID == "" {
		return nil, response.BadRequest("Alert ID is required")
	}

	alert, err := s.repo.FindByID(ctx, alertID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NotFoundError("SOS alert")
		}
		logger.Error("failed to fetch SOS alert", "error", err, "alertID", alertID)
		return nil, response.InternalServerError("Failed to fetch SOS alert", err)
	}

	// Allow access if user owns the alert or is an admin
	if alert.UserID != userID && !isAdmin {
		return nil, response.ForbiddenError("Unauthorized")
	}

	return dto.ToSOSAlertResponse(alert), nil
}

func (s *service) GetActiveSOS(ctx context.Context, userID string) (*dto.SOSAlertResponse, error) {
	if userID == "" {
		return nil, response.BadRequest("User ID is required")
	}

	alert, err := s.repo.FindActiveByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NotFoundError("No active SOS alert")
		}
		logger.Error("failed to fetch active SOS alert", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to fetch active SOS alert", err)
	}

	return dto.ToSOSAlertResponse(alert), nil
}

func (s *service) ListSOS(ctx context.Context, userID string, req dto.ListSOSRequest) ([]*dto.SOSAlertListResponse, int64, error) {
	alerts, total, err := s.repo.List(ctx, userID, req.Status, req.Page, req.Limit)
	if err != nil {
		logger.Error("failed to list SOS alerts", "error", err, "userID", userID)
		return nil, 0, response.InternalServerError("Failed to fetch SOS alerts", err)
	}

	result := make([]*dto.SOSAlertListResponse, 0, len(alerts))
	for _, alert := range alerts {
		result = append(result, dto.ToSOSAlertListResponse(alert))
	}

	return result, total, nil
}

func (s *service) ResolveSOS(ctx context.Context, userID, alertID string, req dto.ResolveSOSRequest) (*dto.SOSAlertResponse, error) {
	if userID == "" || alertID == "" {
		return nil, response.BadRequest("User ID and Alert ID are required")
	}

	alert, err := s.repo.FindByID(ctx, alertID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NotFoundError("SOS alert")
		}
		return nil, response.InternalServerError("Failed to fetch SOS alert", err)
	}

	if alert.UserID != userID {
		return nil, response.ForbiddenError("You can only resolve your own alerts")
	}

	if alert.Status != "active" {
		return nil, response.BadRequest(fmt.Sprintf("Alert cannot be resolved, current status: %s", alert.Status))
	}

	if err := s.repo.Resolve(ctx, alertID, userID, req.Notes); err != nil {
		logger.Error("failed to resolve SOS alert", "error", err, "alertID", alertID)
		return nil, response.InternalServerError("Failed to resolve SOS alert", err)
	}

	updatedAlert, err := s.repo.FindByID(ctx, alertID)
	if err != nil {
		logger.Error("failed to fetch updated alert after resolve", "error", err, "alertID", alertID)
		return nil, response.InternalServerError("Failed to fetch updated alert", err)
	}

	go NotifySOSResolved(context.Background(), updatedAlert, userID)

	logger.Info("SOS alert resolved", "alertID", alertID, "userID", userID)
	return dto.ToSOSAlertResponse(updatedAlert), nil
}

func (s *service) CancelSOS(ctx context.Context, userID, alertID string) (*dto.SOSAlertResponse, error) {
	if userID == "" || alertID == "" {
		return nil, response.BadRequest("User ID and Alert ID are required")
	}

	alert, err := s.repo.FindByID(ctx, alertID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NotFoundError("SOS alert")
		}
		return nil, response.InternalServerError("Failed to fetch SOS alert", err)
	}

	if alert.UserID != userID {
		return nil, response.ForbiddenError("You can only cancel your own alerts")
	}

	if alert.Status != "active" {
		return nil, response.BadRequest(fmt.Sprintf("Alert cannot be cancelled, current status: %s", alert.Status))
	}

	if err := s.repo.Cancel(ctx, alertID); err != nil {
		logger.Error("failed to cancel SOS alert", "error", err, "alertID", alertID)
		return nil, response.InternalServerError("Failed to cancel SOS alert", err)
	}

	updatedAlert, err := s.repo.FindByID(ctx, alertID)
	if err != nil {
		logger.Error("failed to fetch updated alert after cancel", "error", err, "alertID", alertID)
		return nil, response.InternalServerError("Failed to fetch updated alert", err)
	}

	logger.Info("SOS alert cancelled", "alertID", alertID, "userID", userID)
	return dto.ToSOSAlertResponse(updatedAlert), nil
}

func (s *service) UpdateSOSLocation(ctx context.Context, userID, alertID string, latitude, longitude float64) (*dto.SOSAlertResponse, error) {
	if userID == "" || alertID == "" {
		return nil, response.BadRequest("User ID and Alert ID are required")
	}

	alert, err := s.repo.FindByID(ctx, alertID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NotFoundError("SOS alert")
		}
		return nil, response.InternalServerError("Failed to fetch SOS alert", err)
	}

	if alert.UserID != userID {
		return nil, response.ForbiddenError("You can only update your own alerts")
	}

	if alert.Status != "active" {
		return nil, response.BadRequest("Alert is not active")
	}

	if err := s.repo.UpdateLocation(ctx, alertID, latitude, longitude); err != nil {
		logger.Error("failed to update SOS location", "error", err, "alertID", alertID)
		return nil, response.InternalServerError("Failed to update SOS location", err)
	}

	go websocketutil.SendSOSLocationUpdate(userID, map[string]interface{}{
		"alertId":   alertID,
		"rideId":    alert.RideID,
		"latitude":  latitude,
		"longitude": longitude,
		"timestamp": time.Now().UTC(),
	}, true)

	updatedAlert, err := s.repo.FindByID(ctx, alertID)
	if err != nil {
		logger.Error("failed to fetch updated alert after location update", "error", err, "alertID", alertID)
		return nil, response.InternalServerError("Failed to fetch updated alert", err)
	}

	logger.Info("SOS location updated",
		"alertID", alertID,
		"userID", userID,
		"latitude", latitude,
		"longitude", longitude,
	)
	return dto.ToSOSAlertResponse(updatedAlert), nil
}

func (s *service) notifyEmergencyContacts(ctx context.Context, alert *models.SOSAlert) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic in notifyEmergencyContacts", "recover", r, "alertID", alert.ID)
		}
	}()

	var user models.User
	if err := s.userDB.WithContext(ctx).Where("id = ?", alert.UserID).First(&user).Error; err != nil {
		logger.Error("failed to fetch user for emergency contact", "error", err, "userID", alert.UserID)
		return
	}

	if user.EmergencyContactPhone == "" {
		logger.Warn("user has no emergency contact", "userID", alert.UserID)
		return
	}

	message := fmt.Sprintf(
		"EMERGENCY ALERT: %s has triggered an SOS. Location: https://maps.google.com/?q=%f,%f",
		user.Name, alert.Latitude, alert.Longitude,
	)

	logger.Info("emergency contact notification",
		"phone", user.EmergencyContactPhone,
		"message", message,
	)

	websocketutil.SendToUser(user.ID, websocket.TypeSOSAlert, map[string]interface{}{
		"alertId":  alert.ID,
		"userName": user.Name,
		"location": map[string]float64{
			"latitude":  alert.Latitude,
			"longitude": alert.Longitude,
		},
		"message": "SOS Alert from your emergency contact",
	})

	s.publishSOSEvent(ctx, notificationsmodule.EventSOSAlert, alert.ID, alert.UserID, map[string]interface{}{
		"location": map[string]float64{
			"latitude":  alert.Latitude,
			"longitude": alert.Longitude,
		},
		"message": "SOS Alert from your emergency contact",
	})

	if err := s.repo.MarkEmergencyContactsNotified(ctx, alert.ID); err != nil {
		logger.Error("failed to mark emergency contacts notified", "error", err, "alertID", alert.ID)
	}
}

func (s *service) notifySafetyTeam(ctx context.Context, alert *models.SOSAlert) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("panic in notifySafetyTeam", "recover", r, "alertID", alert.ID)
		}
	}()

	s.publishSOSEvent(context.Background(), notificationsmodule.EventSOSAlert, alert.ID, alert.UserID, map[string]interface{}{
		"rideId":    alert.RideID,
		"latitude":  alert.Latitude,
		"longitude": alert.Longitude,
		"timestamp": alert.CreatedAt,
	})

	websocketutil.BroadcastToRole("admin", websocket.TypeSOSAlert, map[string]interface{}{
		"type":      "sos_alert",
		"alertId":   alert.ID,
		"userId":    alert.UserID,
		"rideId":    alert.RideID,
		"latitude":  alert.Latitude,
		"longitude": alert.Longitude,
		"timestamp": alert.CreatedAt,
	})

	logger.Warn("SAFETY TEAM NOTIFIED - SOS ALERT BROADCAST",
		"alertID", alert.ID,
		"userID", alert.UserID,
		"rideID", alert.RideID,
	)

	if err := s.repo.MarkSafetyTeamNotified(ctx, alert.ID); err != nil {
		logger.Error("failed to mark safety team notified", "error", err, "alertID", alert.ID)
	}
}

func (s *service) publishSOSEvent(ctx context.Context, eventType notificationsmodule.EventType, alertID, userID string, data map[string]interface{}) {
	if s.eventProducer == nil {
		logger.Debug("event producer not available, skipping SOS event publication", "eventType", eventType, "alertID", alertID)
		return
	}

	payload := map[string]interface{}{
		"alert_id":  alertID,
		"user_id":   userID,
		"timestamp": time.Now().UTC(),
	}

	for k, v := range data {
		payload[k] = v
	}

	go func() {
		// Use background context to prevent cancellation when HTTP request completes
		bgCtx := context.Background()
		if err := s.eventProducer.PublishEventWithKey(bgCtx, eventType, alertID, payload); err != nil {
			logger.Error("failed to publish SOS event",
				"error", err,
				"eventType", eventType,
				"alertID", alertID,
			)
		}
	}()
}
