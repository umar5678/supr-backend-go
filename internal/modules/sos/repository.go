package sos

import (
	"context"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, alert *models.SOSAlert) error
	FindByID(ctx context.Context, id string) (*models.SOSAlert, error)
	FindActiveByUserID(ctx context.Context, userID string) (*models.SOSAlert, error)
	List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.SOSAlert, int64, error)
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
	var alert models.SOSAlert
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Ride").
		Where("id = ?", id).
		First(&alert).Error
	return &alert, err
}

func (r *repository) FindActiveByUserID(ctx context.Context, userID string) (*models.SOSAlert, error) {
	var alert models.SOSAlert
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID, "active").
		Order("created_at DESC").
		First(&alert).Error
	return &alert, err
}

func (r *repository) List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.SOSAlert, int64, error) {
	// Build base query with filters
	query := r.db.WithContext(ctx)

	// Apply userId filter (required)
	if userID, ok := filters["userId"].(string); ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Apply status filter (optional)
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count
	var total int64
	if err := query.Model(&models.SOSAlert{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch alerts with pagination
	var alerts []*models.SOSAlert
	offset := (page - 1) * limit

	// Build fresh query for fetch (to avoid state corruption)
	fetchQuery := r.db.WithContext(ctx)

	if userID, ok := filters["userId"].(string); ok && userID != "" {
		fetchQuery = fetchQuery.Where("user_id = ?", userID)
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		fetchQuery = fetchQuery.Where("status = ?", status)
	}

	err := fetchQuery.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Preload("User").
		Preload("Ride").
		Find(&alerts).Error

	if err != nil {
		return nil, 0, err
	}

	return alerts, total, nil
}

func (r *repository) Resolve(ctx context.Context, alertID, resolvedBy, notes string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.SOSAlert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"status":      "resolved",
			"resolved_at": now,
			"resolved_by": resolvedBy,
			"notes":       notes,
		}).Error
}

func (r *repository) Cancel(ctx context.Context, alertID string) error {
	return r.db.WithContext(ctx).
		Model(&models.SOSAlert{}).
		Where("id = ?", alertID).
		Update("status", "cancelled").Error
}

func (r *repository) UpdateLocation(ctx context.Context, alertID string, latitude, longitude float64) error {
	return r.db.WithContext(ctx).
		Model(&models.SOSAlert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"latitude":   latitude,
			"longitude":  longitude,
			"updated_at": time.Now(),
		}).Error
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
