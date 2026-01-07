// internal/modules/pricing/surge.go
package pricing

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/umar5678/go-backend/internal/models"
)

// SurgeManager handles surge pricing logic
type SurgeManager struct {
	repo Repository
}

func NewSurgeManager(repo Repository) *SurgeManager {
	return &SurgeManager{repo: repo}
}

// GetSurgeMultiplier returns surge multiplier for a location
func (m *SurgeManager) GetSurgeMultiplier(ctx context.Context, lat, lon float64) (float64, error) {
	// Check if location is in any active surge zone
	zones, err := m.repo.GetActiveSurgeZones(ctx, lat, lon)
	if err != nil {
		return 1.0, err // Default to no surge
	}

	// If no surge zones found, return default or random fake surge
	if len(zones) == 0 {
		return m.getFakeSurgeMultiplier(), nil
	}

	// Return highest surge multiplier if in multiple zones
	maxMultiplier := 1.0
	for _, zone := range zones {
		if zone.Multiplier > maxMultiplier {
			maxMultiplier = zone.Multiplier
		}
	}

	return maxMultiplier, nil
}

// getFakeSurgeMultiplier generates random surge for demo
func (m *SurgeManager) getFakeSurgeMultiplier() float64 {
	rand.Seed(time.Now().UnixNano())

	// 70% chance: no surge (1.0x)
	// 20% chance: 1.5x surge
	// 10% chance: 2.0x surge
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

// GetSurgeZoneByGeohash returns surge zone by geohash
func (m *SurgeManager) GetSurgeZoneByGeohash(ctx context.Context, geohash string) (*models.SurgePricingZone, error) {
	return m.repo.GetSurgeZoneByGeohash(ctx, geohash)
}

// IsLocationInSurgeZone checks if location is in any surge zone
func (m *SurgeManager) IsLocationInSurgeZone(ctx context.Context, lat, lon float64) (bool, float64, error) {
	zones, err := m.repo.GetActiveSurgeZones(ctx, lat, lon)
	if err != nil {
		return false, 1.0, err
	}

	if len(zones) == 0 {
		return false, 1.0, nil
	}

	// Return highest multiplier
	maxMultiplier := 1.0
	for _, zone := range zones {
		if zone.Multiplier > maxMultiplier {
			maxMultiplier = zone.Multiplier
		}
	}

	return true, maxMultiplier, nil
}

// CalculateTimeBasedSurge calculates surge multiplier based on time of day
func (m *SurgeManager) CalculateTimeBasedSurge(ctx context.Context, vehicleTypeID string) (float64, error) {
	// Get all active rules
	rules, err := m.repo.GetActiveSurgePricingRules(ctx)
	if err != nil {
		return 1.0, err
	}

	now := time.Now()
	currentTime := now.Format("15:04")
	currentDay := int(now.Weekday())

	maxMultiplier := 1.0

	// Check each rule
	for _, rule := range rules {
		// Skip if vehicle type doesn't match
		if rule.VehicleTypeID != "" && rule.VehicleTypeID != vehicleTypeID {
			continue
		}

		// Check if rule applies to current day
		if rule.DayOfWeek != -1 && rule.DayOfWeek != currentDay {
			continue
		}

		// Check if current time is within rule's time window
		if currentTime >= rule.StartTime && currentTime <= rule.EndTime {
			if rule.BaseMultiplier > maxMultiplier {
				maxMultiplier = rule.BaseMultiplier
			}
		}
	}

	return maxMultiplier, nil
}

// CalculateDemandBasedSurge calculates surge based on pending requests vs available drivers
func (m *SurgeManager) CalculateDemandBasedSurge(ctx context.Context, geohash string, lat, lon float64) (float64, error) {
	// Get latest demand data
	demand, err := m.repo.GetLatestDemandByGeohash(ctx, geohash)
	if err != nil || demand == nil {
		return 1.0, nil // No demand data available
	}

	// Calculate demand-supply ratio
	if demand.AvailableDrivers == 0 {
		// No drivers available - max surge
		return 2.0, nil
	}

	ratio := float64(demand.PendingRequests) / float64(demand.AvailableDrivers)

	// Convert ratio to multiplier
	// 0-1: 1.0x (normal)
	// 1-2: 1.25x
	// 2-3: 1.5x
	// 3+: 2.0x (cap at 2.0)
	multiplier := 1.0 + (ratio * 0.25)
	if multiplier > 2.0 {
		multiplier = 2.0
	}

	return multiplier, nil
}

// CalculateCombinedSurge combines time-based and demand-based surge
// Returns: combined multiplier, time-based, demand-based, and reason string
func (m *SurgeManager) CalculateCombinedSurge(ctx context.Context, vehicleTypeID, geohash string, lat, lon float64) (float64, float64, float64, string, error) {
	// Get time-based surge
	timeSurge, err := m.CalculateTimeBasedSurge(ctx, vehicleTypeID)
	if err != nil {
		timeSurge = 1.0
	}

	// Get demand-based surge
	demandSurge, err := m.CalculateDemandBasedSurge(ctx, geohash, lat, lon)
	if err != nil {
		demandSurge = 1.0
	}

	// Combine surges (apply both but cap at max)
	combined := math.Max(timeSurge, demandSurge)

	// Determine reason
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

// ApplyMaxMultiplierCap applies the max multiplier cap from surge rule
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
