package sos

import (
	"context"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/sos/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
	"github.com/umar5678/go-backend/internal/websocket"
	websocketutil "github.com/umar5678/go-backend/internal/websocket/websocketutils"
	"gorm.io/gorm"
)

type Service interface {
	TriggerSOS(ctx context.Context, userID string, req dto.TriggerSOSRequest) (*dto.SOSAlertResponse, error)
	GetSOS(ctx context.Context, userID, alertID string) (*dto.SOSAlertResponse, error)
	GetActiveSOS(ctx context.Context, userID string) (*dto.SOSAlertResponse, error)
	ListSOS(ctx context.Context, userID string, req dto.ListSOSRequest) ([]*dto.SOSAlertListResponse, int64, error)
	ResolveSOS(ctx context.Context, userID, alertID string, req dto.ResolveSOSRequest) error
	CancelSOS(ctx context.Context, userID, alertID string) error
}

type service struct {
	repo   Repository
	userDB *gorm.DB // To fetch user emergency contacts
}

func NewService(repo Repository, db *gorm.DB) Service {
	return &service{
		repo:   repo,
		userDB: db,
	}
}

func (s *service) TriggerSOS(ctx context.Context, userID string, req dto.TriggerSOSRequest) (*dto.SOSAlertResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Check if user already has an active SOS
	existingAlert, err := s.repo.FindActiveByUserID(ctx, userID)
	if err == nil && existingAlert.Status == "active" {
		return nil, response.BadRequest("You already have an active SOS alert")
	}

	alert := &models.SOSAlert{
		UserID:      userID,
		RideID:      req.RideID,
		AlertType:   "manual",
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Status:      "active",
		Severity:    "critical", // SOS alerts are always critical
		TriggeredAt: time.Now(),
	}

	if err := s.repo.Create(ctx, alert); err != nil {
		logger.Error("failed to create SOS alert", "error", err, "userID", userID)
		return nil, response.InternalServerError("Failed to trigger SOS", err)
	}

	// Fetch alert with user info
	alert, _ = s.repo.FindByID(ctx, alert.ID)

	// Send notifications asynchronously
	go s.notifyEmergencyContacts(context.Background(), alert)
	go s.notifySafetyTeam(context.Background(), alert)

	logger.Warn("SOS ALERT TRIGGERED",
		"alertID", alert.ID,
		"userID", userID,
		"rideID", req.RideID,
		"location", fmt.Sprintf("%f,%f", req.Latitude, req.Longitude),
	)

	return dto.ToSOSAlertResponse(alert), nil
}

func (s *service) GetSOS(ctx context.Context, userID, alertID string) (*dto.SOSAlertResponse, error) {
	alert, err := s.repo.FindByID(ctx, alertID)
	if err != nil {
		return nil, response.NotFoundError("SOS alert")
	}

	// Users can only see their own alerts (admins see all)
	if alert.UserID != userID {
		return nil, response.ForbiddenError("Unauthorized")
	}

	return dto.ToSOSAlertResponse(alert), nil
}

func (s *service) GetActiveSOS(ctx context.Context, userID string) (*dto.SOSAlertResponse, error) {
	alert, err := s.repo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return nil, response.NotFoundError("No active SOS alert")
	}

	return dto.ToSOSAlertResponse(alert), nil
}

func (s *service) ListSOS(ctx context.Context, userID string, req dto.ListSOSRequest) ([]*dto.SOSAlertListResponse, int64, error) {
	req.SetDefaults()

	filters := map[string]interface{}{
		"userId": userID,
	}
	if req.Status != "" {
		filters["status"] = req.Status
	}

	alerts, total, err := s.repo.List(ctx, filters, req.Page, req.Limit)
	if err != nil {
		return nil, 0, response.InternalServerError("Failed to fetch SOS alerts", err)
	}

	result := make([]*dto.SOSAlertListResponse, len(alerts))
	for i, alert := range alerts {
		result[i] = dto.ToSOSAlertListResponse(alert)
	}

	return result, total, nil
}

func (s *service) ResolveSOS(ctx context.Context, userID, alertID string, req dto.ResolveSOSRequest) error {
	alert, err := s.repo.FindByID(ctx, alertID)
	if err != nil {
		return response.NotFoundError("SOS alert")
	}

	if alert.UserID != userID {
		return response.ForbiddenError("You can only resolve your own alerts")
	}

	if alert.Status != "active" {
		return response.BadRequest("Alert is not active")
	}

	if err := s.repo.Resolve(ctx, alertID, userID, req.Notes); err != nil {
		logger.Error("failed to resolve SOS alert", "error", err, "alertID", alertID)
		return response.InternalServerError("Failed to resolve SOS alert", err)
	}

	logger.Info("SOS alert resolved", "alertID", alertID, "userID", userID)
	return nil
}

func (s *service) CancelSOS(ctx context.Context, userID, alertID string) error {
	alert, err := s.repo.FindByID(ctx, alertID)
	if err != nil {
		return response.NotFoundError("SOS alert")
	}

	if alert.UserID != userID {
		return response.ForbiddenError("You can only cancel your own alerts")
	}

	if alert.Status != "active" {
		return response.BadRequest("Alert is not active")
	}

	if err := s.repo.Cancel(ctx, alertID); err != nil {
		logger.Error("failed to cancel SOS alert", "error", err, "alertID", alertID)
		return response.InternalServerError("Failed to cancel SOS alert", err)
	}

	logger.Info("SOS alert cancelled", "alertID", alertID, "userID", userID)
	return nil
}

func (s *service) notifyEmergencyContacts(ctx context.Context, alert *models.SOSAlert) {
	// Fetch user emergency contact
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
		"🚨 EMERGENCY ALERT: %s has triggered an SOS. Location: https://maps.google.com/?q=%f,%f",
		user.Name, alert.Latitude, alert.Longitude,
	)

	// TODO: Send SMS via Twilio/SNS
	logger.Info("emergency contact notification",
		"phone", user.EmergencyContactPhone,
		"message", message,
	)

	// Send via WebSocket to emergency contact if online
	websocketutil.SendToUser(user.ID, websocket.TypeSOSAlert, map[string]interface{}{
		"alertId":  alert.ID,
		"userName": user.Name,
		"location": map[string]float64{
			"latitude":  alert.Latitude,
			"longitude": alert.Longitude,
		},
		"message": "SOS Alert from your emergency contact",
	})

	s.repo.MarkEmergencyContactsNotified(ctx, alert.ID)
}

func (s *service) notifySafetyTeam(ctx context.Context, alert *models.SOSAlert) {
	// Send to safety monitoring dashboard via WebSocket to all connected users
	websocketutil.BroadcastToRole("admin", websocket.TypeSOSAlert, map[string]interface{}{
		"type":      "sos_alert",
		"alertId":   alert.ID,
		"userId":    alert.UserID,
		"rideId":    alert.RideID,
		"latitude":  alert.Latitude,
		"longitude": alert.Longitude,
		"timestamp": alert.CreatedAt,
	})

	// TODO: Trigger call to safety line
	logger.Warn("SAFETY TEAM NOTIFIED - SOS ALERT BROADCAST",
		"alertID", alert.ID,
		"userID", alert.UserID,
		"rideID", alert.RideID,
	)

	s.repo.MarkSafetyTeamNotified(ctx, alert.ID)
}
