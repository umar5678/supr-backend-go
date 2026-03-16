package sos

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, alert *models.SOSAlert) error
	FindByID(ctx context.Context, id string) (*models.SOSAlert, error)
	FindActiveByUserID(ctx context.Context, userID string) (*models.SOSAlert, error)
	List(ctx context.Context, userID string, status string, page, limit int) ([]*models.SOSAlert, int64, error)
	Resolve(ctx context.Context, alertID, resolvedBy, notes string) error
	Cancel(ctx context.Context, alertID string) error
	UpdateLocation(ctx context.Context, alertID string, latitude, longitude float64) error
	MarkEmergencyContactsNotified(ctx context.Context, alertID string) error
	MarkSafetyTeamNotified(ctx context.Context, alertID string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, alert *models.SOSAlert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*models.SOSAlert, error) {
	if id == "" {
		return nil, errors.New("alert ID is required")
	}

	var alert models.SOSAlert
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Ride").
		Where("id = ?", id).
		First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *repository) FindActiveByUserID(ctx context.Context, userID string) (*models.SOSAlert, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	var alert models.SOSAlert
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Ride").
		Where("user_id = ? AND status = ?", userID, "active").
		Order("created_at DESC").
		First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *repository) List(ctx context.Context, userID string, status string, page, limit int) ([]*models.SOSAlert, int64, error) {
	var alerts []*models.SOSAlert
	var total int64

	base := r.db.WithContext(ctx).Model(&models.SOSAlert{})
	if userID != "" {
		base = base.Where("user_id = ?", userID)
	}
	if status != "" {
		base = base.Where("status = ?", status)
	}

	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count SOS alerts: %w", err)
	}

	if total == 0 {
		return []*models.SOSAlert{}, 0, nil
	}

	offset := (page - 1) * limit

	if err := base.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Preload("User").
		Preload("Ride").
		Find(&alerts).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch SOS alerts: %w", err)
	}

	return alerts, total, nil
}

func (r *repository) Resolve(ctx context.Context, alertID, resolvedBy, notes string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&models.SOSAlert{}).
		Where("id = ? AND status = ?", alertID, "active").
		Updates(map[string]interface{}{
			"status":      "resolved",
			"resolved_at": now,
			"resolved_by": resolvedBy,
			"notes":       notes,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("alert not found or already resolved")
	}
	return nil
}

func (r *repository) Cancel(ctx context.Context, alertID string) error {
	result := r.db.WithContext(ctx).
		Model(&models.SOSAlert{}).
		Where("id = ? AND status = ?", alertID, "active").
		Update("status", "cancelled")

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("alert not found or already cancelled")
	}
	return nil
}

func (r *repository) UpdateLocation(ctx context.Context, alertID string, latitude, longitude float64) error {
	result := r.db.WithContext(ctx).
		Model(&models.SOSAlert{}).
		Where("id = ? AND status = ?", alertID, "active").
		Updates(map[string]interface{}{
			"latitude":  latitude,
			"longitude": longitude,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("alert not found or not active")
	}
	return nil
}

func (r *repository) MarkEmergencyContactsNotified(ctx context.Context, alertID string) error {
	return r.db.WithContext(ctx).
		Model(&models.SOSAlert{}).
		Where("id = ?", alertID).
		Update("emergency_contacts_notified", true).Error
}

func (r *repository) MarkSafetyTeamNotified(ctx context.Context, alertID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.SOSAlert{}).
		Where("id = ?", alertID).
		Update("safety_team_notified_at", now).Error
}
