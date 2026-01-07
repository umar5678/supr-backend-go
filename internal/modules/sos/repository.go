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
    var alerts []*models.SOSAlert
    var total int64

    query := r.db.WithContext(ctx).Model(&models.SOSAlert{})

    if status, ok := filters["status"].(string); ok && status != "" {
        query = query.Where("status = ?", status)
    }
    if userID, ok := filters["userId"].(string); ok && userID != "" {
        query = query.Where("user_id = ?", userID)
    }

    query.Count(&total)

    offset := (page - 1) * limit
    err := query.
        Preload("User").
        Preload("Ride").
        Order("created_at DESC").
        Offset(offset).
        Limit(limit).
        Find(&alerts).Error

    return alerts, total, err
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
