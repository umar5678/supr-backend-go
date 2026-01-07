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

	// Surge pricing rules (time-based and demand-based)
	GetActiveSurgePricingRules(ctx context.Context) ([]*models.SurgePricingRule, error)
	GetSurgePricingRuleByVehicleType(ctx context.Context, vehicleTypeID string) (*models.SurgePricingRule, error)
	CreateSurgePricingRule(ctx context.Context, rule *models.SurgePricingRule) error
	UpdateSurgePricingRule(ctx context.Context, rule *models.SurgePricingRule) error

	// Demand tracking
	GetDemandByZone(ctx context.Context, zoneID string) (*models.DemandTracking, error)
	CreateDemandTracking(ctx context.Context, demand *models.DemandTracking) error
	UpdateDemandTracking(ctx context.Context, demand *models.DemandTracking) error
	GetLatestDemandByGeohash(ctx context.Context, geohash string) (*models.DemandTracking, error)

	// ETA estimates
	CreateETAEstimate(ctx context.Context, eta *models.ETAEstimate) error
	GetETAEstimate(ctx context.Context, rideID string) (*models.ETAEstimate, error)
	UpdateETAEstimate(ctx context.Context, eta *models.ETAEstimate) error

	// Surge history
	CreateSurgeHistory(ctx context.Context, history *models.SurgeHistory) error
	GetSurgeHistory(ctx context.Context, rideID string) (*models.SurgeHistory, error)

	// Price capping
	FindPriceCappingRule(ctx context.Context, vehicleTypeID string) (*models.PriceCappingRule, error)
	CreatePriceCappingRule(ctx context.Context, rule *models.PriceCappingRule) error

	// Wait time
	CreateWaitTimeCharge(ctx context.Context, charge *models.WaitTimeCharge) error
	FindWaitTimeChargeByRideID(ctx context.Context, rideID string) (*models.WaitTimeCharge, error)
	UpdateWaitTimeCharge(ctx context.Context, charge *models.WaitTimeCharge) error

	// Surge zones
	GetSurgeMultiplier(ctx context.Context, lat, lon float64, vehicleTypeID string) (float64, error)

	// Ride updates
	UpdateRideDestination(ctx context.Context, rideID string, lat, lon float64, address string, additionalCharge float64) error
	UpdateRideWaitTimeCharge(ctx context.Context, rideID string, charge float64) error
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

func (r *repository) FindPriceCappingRule(ctx context.Context, vehicleTypeID string) (*models.PriceCappingRule, error) {
	var rule models.PriceCappingRule
	err := r.db.WithContext(ctx).
		Where("vehicle_type_id = ? AND is_active = true", vehicleTypeID).
		First(&rule).Error
	return &rule, err
}

func (r *repository) CreatePriceCappingRule(ctx context.Context, rule *models.PriceCappingRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *repository) CreateWaitTimeCharge(ctx context.Context, charge *models.WaitTimeCharge) error {
	return r.db.WithContext(ctx).Create(charge).Error
}

func (r *repository) FindWaitTimeChargeByRideID(ctx context.Context, rideID string) (*models.WaitTimeCharge, error) {
	var charge models.WaitTimeCharge
	err := r.db.WithContext(ctx).
		Where("ride_id = ?", rideID).
		First(&charge).Error
	return &charge, err
}

func (r *repository) UpdateWaitTimeCharge(ctx context.Context, charge *models.WaitTimeCharge) error {
	return r.db.WithContext(ctx).Save(charge).Error
}

func (r *repository) GetSurgeMultiplier(ctx context.Context, lat, lon float64, vehicleTypeID string) (float64, error) {
	var multiplier float64

	err := r.db.WithContext(ctx).Raw(`
        SELECT COALESCE(MAX(multiplier), 1.0) as multiplier
        FROM surge_zones
        WHERE vehicle_type_id = ?
        AND ST_Contains(zone_boundary, ST_SetSRID(ST_MakePoint(?, ?), 4326))
        AND is_active = true
        AND (valid_until IS NULL OR valid_until > NOW())
    `, vehicleTypeID, lon, lat).Scan(&multiplier).Error

	if err != nil || multiplier == 0 {
		return 1.0, nil
	}

	return multiplier, nil
}

func (r *repository) UpdateRideDestination(ctx context.Context, rideID string, lat, lon float64, address string, additionalCharge float64) error {
	return r.db.WithContext(ctx).
		Model(&models.Ride{}).
		Where("id = ?", rideID).
		Updates(map[string]interface{}{
			"dropoff_lat":               lat,
			"dropoff_lon":               lon,
			"dropoff_address":           address,
			"destination_changed":       true,
			"destination_change_charge": additionalCharge,
		}).Error
}

func (r *repository) UpdateRideWaitTimeCharge(ctx context.Context, rideID string, charge float64) error {
	return r.db.WithContext(ctx).
		Model(&models.Ride{}).
		Where("id = ?", rideID).
		Update("wait_time_charge", charge).Error
}

// GetActiveSurgePricingRules returns all active surge pricing rules
func (r *repository) GetActiveSurgePricingRules(ctx context.Context) ([]*models.SurgePricingRule, error) {
	var rules []*models.SurgePricingRule
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Find(&rules).Error
	return rules, err
}

// GetSurgePricingRuleByVehicleType returns surge rule for a vehicle type
func (r *repository) GetSurgePricingRuleByVehicleType(ctx context.Context, vehicleTypeID string) (*models.SurgePricingRule, error) {
	var rule models.SurgePricingRule
	err := r.db.WithContext(ctx).
		Where("vehicle_type_id = ? AND is_active = ?", vehicleTypeID, true).
		First(&rule).Error
	return &rule, err
}

// CreateSurgePricingRule creates a new surge pricing rule
func (r *repository) CreateSurgePricingRule(ctx context.Context, rule *models.SurgePricingRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// UpdateSurgePricingRule updates an existing surge pricing rule
func (r *repository) UpdateSurgePricingRule(ctx context.Context, rule *models.SurgePricingRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

// GetDemandByZone returns demand tracking for a zone
func (r *repository) GetDemandByZone(ctx context.Context, zoneID string) (*models.DemandTracking, error) {
	var demand models.DemandTracking
	err := r.db.WithContext(ctx).
		Where("zone_id = ?", zoneID).
		Order("recorded_at DESC").
		First(&demand).Error
	return &demand, err
}

// CreateDemandTracking creates a new demand tracking record
func (r *repository) CreateDemandTracking(ctx context.Context, demand *models.DemandTracking) error {
	return r.db.WithContext(ctx).Create(demand).Error
}

// UpdateDemandTracking updates demand tracking record
func (r *repository) UpdateDemandTracking(ctx context.Context, demand *models.DemandTracking) error {
	return r.db.WithContext(ctx).Save(demand).Error
}

// GetLatestDemandByGeohash returns latest demand for a geohash
func (r *repository) GetLatestDemandByGeohash(ctx context.Context, geohash string) (*models.DemandTracking, error) {
	var demand models.DemandTracking
	err := r.db.WithContext(ctx).
		Where("zone_geohash = ?", geohash).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Order("recorded_at DESC").
		First(&demand).Error
	return &demand, err
}

// CreateETAEstimate creates a new ETA estimate
func (r *repository) CreateETAEstimate(ctx context.Context, eta *models.ETAEstimate) error {
	return r.db.WithContext(ctx).Create(eta).Error
}

// GetETAEstimate retrieves ETA estimate for a ride
func (r *repository) GetETAEstimate(ctx context.Context, rideID string) (*models.ETAEstimate, error) {
	var eta models.ETAEstimate
	err := r.db.WithContext(ctx).
		Where("ride_id = ?", rideID).
		First(&eta).Error
	return &eta, err
}

// UpdateETAEstimate updates an ETA estimate
func (r *repository) UpdateETAEstimate(ctx context.Context, eta *models.ETAEstimate) error {
	return r.db.WithContext(ctx).Save(eta).Error
}

// CreateSurgeHistory creates a surge history record
func (r *repository) CreateSurgeHistory(ctx context.Context, history *models.SurgeHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// GetSurgeHistory retrieves surge history for a ride
func (r *repository) GetSurgeHistory(ctx context.Context, rideID string) (*models.SurgeHistory, error) {
	var history models.SurgeHistory
	err := r.db.WithContext(ctx).
		Where("ride_id = ?", rideID).
		First(&history).Error
	return &history, err
}
