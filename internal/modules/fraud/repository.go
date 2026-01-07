// internal/modules/fraud/repository.go
package fraud

import (
	"context"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, pattern *models.FraudPattern) error
	FindByID(ctx context.Context, id string) (*models.FraudPattern, error)
	List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.FraudPattern, int64, error)
	Update(ctx context.Context, pattern *models.FraudPattern) error
	Review(ctx context.Context, patternID, reviewerID, status, notes string) error

	// Pattern detection queries
	CheckFrequentCancellations(ctx context.Context, userID string, days int) (int, error)
	CheckSameRiderDriverPair(ctx context.Context, riderID, driverID string, days int) (int, error)
	CheckShortDistanceHighFare(ctx context.Context, rideID string) (bool, error)
	CheckLocationGaming(ctx context.Context, userID string, days int) (bool, error)

	// Statistics
	GetFraudStats(ctx context.Context) (map[string]interface{}, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, pattern *models.FraudPattern) error {
	return r.db.WithContext(ctx).Create(pattern).Error
}

func (r *repository) FindByID(ctx context.Context, id string) (*models.FraudPattern, error) {
	var pattern models.FraudPattern
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Driver").
		Where("id = ?", id).
		First(&pattern).Error
	return &pattern, err
}

func (r *repository) List(ctx context.Context, filters map[string]interface{}, page, limit int) ([]*models.FraudPattern, int64, error) {
	var patterns []*models.FraudPattern
	var total int64

	query := r.db.WithContext(ctx).Model(&models.FraudPattern{})

	if patternType, ok := filters["patternType"].(string); ok && patternType != "" {
		query = query.Where("pattern_type = ?", patternType)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if minRiskScore, ok := filters["minRiskScore"].(int); ok && minRiskScore > 0 {
		query = query.Where("risk_score >= ?", minRiskScore)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.
		Preload("User").
		Preload("Driver").
		Order("risk_score DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&patterns).Error

	return patterns, total, err
}

func (r *repository) Update(ctx context.Context, pattern *models.FraudPattern) error {
	return r.db.WithContext(ctx).Save(pattern).Error
}

func (r *repository) Review(ctx context.Context, patternID, reviewerID, status, notes string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.FraudPattern{}).
		Where("id = ?", patternID).
		Updates(map[string]interface{}{
			"status":       status,
			"reviewed_by":  reviewerID,
			"review_notes": notes,
			"reviewed_at":  now,
		}).Error
}

func (r *repository) CheckFrequentCancellations(ctx context.Context, userID string, days int) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Ride{}).
		Where("(rider_id = ? OR driver_id = ?) AND status = ? AND created_at > ?",
			userID, userID, "cancelled", time.Now().AddDate(0, 0, -days)).
		Count(&count).Error
	return int(count), err
}

func (r *repository) CheckSameRiderDriverPair(ctx context.Context, riderID, driverID string, days int) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Ride{}).
		Where("rider_id = ? AND driver_id = ? AND created_at > ?",
			riderID, driverID, time.Now().AddDate(0, 0, -days)).
		Count(&count).Error
	return int(count), err
}

func (r *repository) CheckShortDistanceHighFare(ctx context.Context, rideID string) (bool, error) {
	var ride models.Ride
	err := r.db.WithContext(ctx).Where("id = ?", rideID).First(&ride).Error
	if err != nil {
		return false, err
	}

	// Flag if actual distance < 1km but fare > $10
	if ride.ActualDistance != nil && *ride.ActualDistance < 1.0 &&
		ride.ActualFare != nil && *ride.ActualFare > 10.0 {
		return true, nil
	}

	return false, nil
}

func (r *repository) CheckLocationGaming(ctx context.Context, userID string, days int) (bool, error) {
	// Check for multiple rides with same exact pickup/dropoff within short time
	var count int64
	err := r.db.WithContext(ctx).Raw(`
        SELECT COUNT(DISTINCT DATE_TRUNC('hour', requested_at))
        FROM rides
        WHERE rider_id = ?
        AND created_at > ?
        GROUP BY pickup_lat, pickup_lon, dropoff_lat, dropoff_lon
        HAVING COUNT(*) > 5
    `, userID, time.Now().AddDate(0, 0, -days)).Count(&count).Error

	return count > 0, err
}

func (r *repository) GetFraudStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total patterns
	var total int64
	r.db.WithContext(ctx).Model(&models.FraudPattern{}).Count(&total)
	stats["total"] = total

	// By status
	var statusCounts []struct {
		Status string
		Count  int64
	}
	r.db.WithContext(ctx).
		Model(&models.FraudPattern{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts)

	for _, sc := range statusCounts {
		stats[sc.Status] = sc.Count
	}

	// By type
	var typeCounts []struct {
		Type  string
		Count int64
	}
	r.db.WithContext(ctx).
		Model(&models.FraudPattern{}).
		Select("pattern_type as type, COUNT(*) as count").
		Group("pattern_type").
		Scan(&typeCounts)

	byType := make(map[string]int64)
	for _, tc := range typeCounts {
		byType[tc.Type] = tc.Count
	}
	stats["byType"] = byType

	// High risk count (>= 80)
	var highRisk int64
	r.db.WithContext(ctx).
		Model(&models.FraudPattern{}).
		Where("risk_score >= ?", 80).
		Count(&highRisk)
	stats["highRisk"] = highRisk

	return stats, nil
}
