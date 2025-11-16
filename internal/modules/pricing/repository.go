package pricing

import (
	"context"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/utils/location"

	"gorm.io/gorm"
)

type Repository interface {
	// Surge zones
	GetActiveSurgeZones(ctx context.Context, lat, lon float64) ([]*models.SurgePricingZone, error)
	GetSurgeZoneByGeohash(ctx context.Context, geohash string) (*models.SurgePricingZone, error)
	GetAllActiveSurgeZones(ctx context.Context) ([]*models.SurgePricingZone, error)
	CreateSurgeZone(ctx context.Context, zone *models.SurgePricingZone) error
	UpdateSurgeZone(ctx context.Context, zone *models.SurgePricingZone) error
	DeactivateSurgeZone(ctx context.Context, id string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetActiveSurgeZones(ctx context.Context, lat, lon float64) ([]*models.SurgePricingZone, error) {
	var zones []*models.SurgePricingZone

	now := time.Now()

	// Find zones where point is within radius
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Where("active_from <= ?", now).
		Where("active_until >= ?", now).
		Find(&zones).Error

	if err != nil {
		return nil, err
	}

	// Filter by distance (check if point is within zone radius)
	var activeZones []*models.SurgePricingZone
	for _, zone := range zones {
		distance := location.HaversineDistance(lat, lon, zone.CenterLat, zone.CenterLon)
		if distance <= zone.RadiusKm {
			activeZones = append(activeZones, zone)
		}
	}

	return activeZones, nil
}

func (r *repository) GetSurgeZoneByGeohash(ctx context.Context, geohash string) (*models.SurgePricingZone, error) {
	var zone models.SurgePricingZone

	err := r.db.WithContext(ctx).
		Where("area_geohash = ?", geohash).
		Where("is_active = ?", true).
		First(&zone).Error

	return &zone, err
}

func (r *repository) GetAllActiveSurgeZones(ctx context.Context) ([]*models.SurgePricingZone, error) {
	var zones []*models.SurgePricingZone

	now := time.Now()

	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Where("active_from <= ?", now).
		Where("active_until >= ?", now).
		Find(&zones).Error

	return zones, err
}

func (r *repository) CreateSurgeZone(ctx context.Context, zone *models.SurgePricingZone) error {
	return r.db.WithContext(ctx).Create(zone).Error
}

func (r *repository) UpdateSurgeZone(ctx context.Context, zone *models.SurgePricingZone) error {
	return r.db.WithContext(ctx).Save(zone).Error
}

func (r *repository) DeactivateSurgeZone(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&models.SurgePricingZone{}).
		Where("id = ?", id).
		Update("is_active", false).Error
}
