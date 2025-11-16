// internal/modules/pricing/surge.go
package pricing

import (
	"context"
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
