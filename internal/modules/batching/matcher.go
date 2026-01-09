package batching

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/umar5678/go-backend/internal/modules/batching/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

// Matcher implements intelligent driver-request pairing logic
type Matcher struct {
	ranker *DriverRanker
}

// NewMatcher creates a new matching engine
func NewMatcher(ranker *DriverRanker) *Matcher {
	return &Matcher{
		ranker: ranker,
	}
}

// MatchRequestsToDrivers performs optimal driver-ride assignment for a batch
// Returns assignments prioritized by score, with unmatched request IDs
func (m *Matcher) MatchRequestsToDrivers(
	ctx context.Context,
	requests []dto.RideRequestInfo,
	rankedDrivers []dto.DriverRankingScore,
	acceptanceThreshold float64, // Minimum confidence score to offer (0.0-1.0)
) *dto.BatchMatchingResult {
	result := &dto.BatchMatchingResult{
		BatchID:      uuid.New().String(),
		Assignments:  []dto.OptimalAssignment{},
		UnmatchedIDs: []string{},
	}

	if len(requests) == 0 || len(rankedDrivers) == 0 {
		return result
	}

	// Build a map of available drivers (not yet assigned)
	availableDrivers := make(map[string]dto.DriverRankingScore)
	for _, driver := range rankedDrivers {
		availableDrivers[driver.DriverID] = driver
	}

	// Sort requests by estimated complexity (longer dropoff distances = harder to match)
	sortedRequests := sortRequestsByComplexity(requests)

	// Match each request to the best available driver
	for _, request := range sortedRequests {
		// Calculate distance and confidence for each available driver
		driverScores := make([]driverMatchScore, 0)

		for _, driver := range availableDrivers {
			// Calculate distance from driver to pickup
			distance := calculateHaversineDistance(
				driver.Distance, driver.Distance, // Placeholder - should get actual driver position
				request.PickupLat, request.PickupLon,
			)

			// Calculate confidence score
			// Combines driver ranking score with proximity
			// Distance bonus: closer drivers get higher confidence (max 10%)
			// Farther drivers (beyond 5km) get penalty
			var distanceBonus float64
			if distance < 0.5 {
				distanceBonus = 10.0 // Extra 10% for drivers <0.5km away
			} else if distance < 2.0 {
				distanceBonus = 5.0 // Extra 5% for drivers <2km
			} else if distance < 5.0 {
				distanceBonus = 0 // No bonus
			} else {
				distanceBonus = -5.0 // Penalty for drivers >5km
			}

			// Base confidence from driver score (0-100 scale to 0-1)
			baseConfidence := driver.TotalScore / 100.0
			adjustedConfidence := baseConfidence + (distanceBonus / 100.0)

			// Cap at 1.0
			if adjustedConfidence > 1.0 {
				adjustedConfidence = 1.0
			}
			if adjustedConfidence < 0 {
				adjustedConfidence = 0
			}

			driverScores = append(driverScores, driverMatchScore{
				Driver:     driver,
				Distance:   distance,
				Confidence: adjustedConfidence,
			})
		}

		// Sort by confidence descending
		sort.Slice(driverScores, func(i, j int) bool {
			return driverScores[i].Confidence > driverScores[j].Confidence
		})

		// Find best match that passes threshold
		matched := false
		for _, match := range driverScores {
			if match.Confidence >= acceptanceThreshold {
				// Create assignment
				assignment := dto.OptimalAssignment{
					RequestID:  request.RideID,
					RideID:     request.RideID,
					DriverID:   match.Driver.DriverID,
					UserID:     match.Driver.UserID,
					Score:      match.Driver.TotalScore,
					Distance:   match.Distance,
					ETA:        match.Driver.ResponseTime,
					Confidence: match.Confidence,
				}

				result.Assignments = append(result.Assignments, assignment)

				// Remove driver from available pool
				delete(availableDrivers, match.Driver.DriverID)
				matched = true

				logger.Debug("Request matched to driver",
					"requestID", request.RideID,
					"driverID", match.Driver.DriverID,
					"confidence", match.Confidence,
					"distance", match.Distance,
				)
				break
			}
		}

		if !matched {
			result.UnmatchedIDs = append(result.UnmatchedIDs, request.RideID)
			logger.Debug("Request not matched",
				"requestID", request.RideID,
				"reason", "no_driver_above_threshold",
			)
		}
	}

	result.MatchedCount = len(result.Assignments)
	return result
}

// MatchSingleRequestToDriver finds the best driver for a single request
func (m *Matcher) MatchSingleRequestToDriver(
	ctx context.Context,
	request dto.RideRequestInfo,
	availableDriverIDs []string,
	acceptanceThreshold float64,
) (*dto.OptimalAssignment, error) {
	// Rank the available drivers
	rankedDrivers, err := m.ranker.RankDrivers(ctx, availableDriverIDs, request.PickupLat, request.PickupLon)
	if err != nil {
		return nil, err
	}

	if len(rankedDrivers) == 0 {
		return nil, nil
	}

	// Find best match
	for _, driver := range rankedDrivers {
		distance := calculateHaversineDistance(
			driver.Distance, driver.Distance,
			request.PickupLat, request.PickupLon,
		)

		// Calculate confidence
		baseConfidence := driver.TotalScore / 100.0
		var distanceBonus float64
		if distance < 0.5 {
			distanceBonus = 10.0
		} else if distance < 2.0 {
			distanceBonus = 5.0
		} else if distance < 5.0 {
			distanceBonus = 0
		} else {
			distanceBonus = -5.0
		}

		adjustedConfidence := baseConfidence + (distanceBonus / 100.0)
		if adjustedConfidence > 1.0 {
			adjustedConfidence = 1.0
		}
		if adjustedConfidence < 0 {
			adjustedConfidence = 0
		}

		if adjustedConfidence >= acceptanceThreshold {
			return &dto.OptimalAssignment{
				RequestID:  request.RideID,
				RideID:     request.RideID,
				DriverID:   driver.DriverID,
				UserID:     driver.UserID,
				Score:      driver.TotalScore,
				Distance:   distance,
				ETA:        driver.ResponseTime,
				Confidence: adjustedConfidence,
			}, nil
		}
	}

	return nil, nil
}

// sortRequestsByComplexity sorts requests by difficulty to match
// Requests with longer distances are harder to match and should be prioritized
func sortRequestsByComplexity(requests []dto.RideRequestInfo) []dto.RideRequestInfo {
	// Sort by dropoff distance descending (harder to match first)
	sorted := make([]dto.RideRequestInfo, len(requests))
	copy(sorted, requests)

	sort.Slice(sorted, func(i, j int) bool {
		distI := calculateHaversineDistance(
			sorted[i].PickupLat, sorted[i].PickupLon,
			sorted[i].DropoffLat, sorted[i].DropoffLon,
		)
		distJ := calculateHaversineDistance(
			sorted[j].PickupLat, sorted[j].PickupLon,
			sorted[j].DropoffLat, sorted[j].DropoffLon,
		)
		return distI > distJ // Longer distances first
	})

	return sorted
}

// driverMatchScore represents a driver's score for a specific request match
type driverMatchScore struct {
	Driver     dto.DriverRankingScore
	Distance   float64 // Distance from driver to pickup location
	Confidence float64 // 0-1 likelihood of acceptance
}
