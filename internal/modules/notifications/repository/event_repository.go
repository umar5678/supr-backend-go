package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/models"
)

type EventRepository interface {
	Create(ctx context.Context, event *models.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Event, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.EventStatus) error
	GetFailedEvents(ctx context.Context, limit int) ([]*models.Event, error)
	GetEventsByType(ctx context.Context, eventType string, since time.Time) ([]*models.Event, error)
}

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(ctx context.Context, event *models.Event) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *eventRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	var event models.Event
	err := r.db.WithContext(ctx).First(&event, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.EventStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.Event{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *eventRepository) GetFailedEvents(ctx context.Context, limit int) ([]*models.Event, error) {
	var events []*models.Event
	err := r.db.WithContext(ctx).
		Where("status = ?", models.EventStatusFailed).
		Order("created_at DESC").
		Limit(limit).
		Find(&events).Error

	return events, err
}

func (r *eventRepository) GetEventsByType(ctx context.Context, eventType string, since time.Time) ([]*models.Event, error) {
	var events []*models.Event
	err := r.db.WithContext(ctx).
		Where("event_type = ? AND created_at >= ?", eventType, since).
		Order("created_at DESC").
		Find(&events).Error

	return events, err
}
