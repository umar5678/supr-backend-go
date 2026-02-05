// internal/modules/pricing/surge.go
package pricing

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/umar5678/go-backend/internal/models"
)

type SurgeManager struct {
	repo Repository
}

func NewSurgeManager(repo Repository) *SurgeManager {
	return &SurgeManager{repo: repo}
}

func (m *SurgeManager) GetSurgeMultiplier(ctx context.Context, lat, lon float64) (float64, error) {
	zones, err := m.repo.GetActiveSurgeZones(ctx, lat, lon)
	if err != nil {
		return 1.0, err
	}

	if len(zones) == 0 {
		return 1.0, nil
	}

	maxMultiplier := 1.0
	for _, zone := range zones {
		if zone.Multiplier > maxMultiplier {
			maxMultiplier = zone.Multiplier
		}
	}

	return maxMultiplier, nil
}

func (m *SurgeManager) getFakeSurgeMultiplier() float64 {
	rand.Seed(time.Now().UnixNano())

	random := rand.Float64()

	switch {
	case random < 0.70:
		return 1.0
	case random < 0.90:
		return 1.5
	default:
		return 2.0
	}
}

func (m *SurgeManager) GetSurgeZoneByGeohash(ctx context.Context, geohash string) (*models.SurgePricingZone, error) {
	return m.repo.GetSurgeZoneByGeohash(ctx, geohash)
}

func (m *SurgeManager) IsLocationInSurgeZone(ctx context.Context, lat, lon float64) (bool, float64, error) {
	zones, err := m.repo.GetActiveSurgeZones(ctx, lat, lon)
	if err != nil {
		return false, 1.0, err
	}

	if len(zones) == 0 {
		return false, 1.0, nil
	}

	maxMultiplier := 1.0
	for _, zone := range zones {
		if zone.Multiplier > maxMultiplier {
			maxMultiplier = zone.Multiplier
		}
	}

	return true, maxMultiplier, nil
}

func (m *SurgeManager) CalculateTimeBasedSurge(ctx context.Context, vehicleTypeID string) (float64, error) {
	rules, err := m.repo.GetActiveSurgePricingRules(ctx)
	if err != nil {
		return 1.0, err
	}

	now := time.Now()
	currentTime := now.Format("15:04")
	currentDay := int(now.Weekday())

	maxMultiplier := 1.0

	for _, rule := range rules {
		if rule.VehicleTypeID != "" && rule.VehicleTypeID != vehicleTypeID {
			continue
		}

		if rule.DayOfWeek != -1 && rule.DayOfWeek != currentDay {
			continue
		}

		if currentTime >= rule.StartTime && currentTime <= rule.EndTime {
			if rule.BaseMultiplier > maxMultiplier {
				maxMultiplier = rule.BaseMultiplier
			}
		}
	}

	return maxMultiplier, nil
}

func (m *SurgeManager) CalculateDemandBasedSurge(ctx context.Context, geohash string, lat, lon float64) (float64, error) {
	demand, err := m.repo.GetLatestDemandByGeohash(ctx, geohash)
	if err != nil || demand == nil {
		return 1.0, nil
	}

	if demand.AvailableDrivers == 0 {
		return 2.0, nil
	}

	ratio := float64(demand.PendingRequests) / float64(demand.AvailableDrivers)

	multiplier := 1.0 + (ratio * 0.25)
	if multiplier > 2.0 {
		multiplier = 2.0
	}

	return multiplier, nil
}

func (m *SurgeManager) CalculateCombinedSurge(ctx context.Context, vehicleTypeID, geohash string, lat, lon float64) (float64, float64, float64, string, error) {
	timeSurge, err := m.CalculateTimeBasedSurge(ctx, vehicleTypeID)
	if err != nil {
		timeSurge = 1.0
	}

	demandSurge, err := m.CalculateDemandBasedSurge(ctx, geohash, lat, lon)
	if err != nil {
		demandSurge = 1.0
	}

	combined := math.Max(timeSurge, demandSurge)

	reason := "normal"
	if timeSurge > 1.0 && demandSurge > 1.0 {
		reason = "combined"
	} else if timeSurge > 1.0 {
		reason = "time_based"
	} else if demandSurge > 1.0 {
		reason = "demand_based"
	}

	return combined, timeSurge, demandSurge, reason, nil
}

func (m *SurgeManager) ApplyMaxMultiplierCap(multiplier float64, vehicleTypeID string, ctx context.Context) float64 {
	rule, err := m.repo.GetSurgePricingRuleByVehicleType(ctx, vehicleTypeID)
	if err != nil || rule == nil {
		return multiplier
	}

	if multiplier > rule.MaxMultiplier {
		return rule.MaxMultiplier
	}

	if multiplier < rule.MinMultiplier {
		return rule.MinMultiplier
	}

	return multiplier
}

func (m *SurgeManager) CalculateZoneBasedSurge(ctx context.Context, lat, lon float64) (float64, error) {
	zones, err := m.repo.GetActiveSurgeZones(ctx, lat, lon)
	if err != nil {
		return 1.0, err
	}

	if len(zones) == 0 {
		return 1.0, nil
	}

	maxMultiplier := 1.0
	for _, zone := range zones {
		if zone.Multiplier > maxMultiplier {
			maxMultiplier = zone.Multiplier
		}
	}

	return maxMultiplier, nil
}

func (m *SurgeManager) GetZoneInfo(ctx context.Context, lat, lon float64) (*models.SurgePricingZone, error) {
	zones, err := m.repo.GetActiveSurgeZones(ctx, lat, lon)
	if err != nil {
		return nil, err
	}

	if len(zones) == 0 {
		return nil, nil
	}

	return zones[0], nil
}
