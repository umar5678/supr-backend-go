package batching

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/umar5678/go-backend/internal/modules/batching/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
)

type Matcher struct {
	ranker *DriverRanker
}

func NewMatcher(ranker *DriverRanker) *Matcher {
	return &Matcher{
		ranker: ranker,
	}
}

func (m *Matcher) MatchRequestsToDrivers(
	ctx context.Context,
	requests []dto.RideRequestInfo,
	rankedDrivers []dto.DriverRankingScore,
	acceptanceThreshold float64,
) *dto.BatchMatchingResult {
	result := &dto.BatchMatchingResult{
		BatchID:      uuid.New().String(),
		Assignments:  []dto.OptimalAssignment{},
		UnmatchedIDs: []string{},
	}

	if len(requests) == 0 || len(rankedDrivers) == 0 {
		return result
	}

	availableDrivers := make(map[string]dto.DriverRankingScore)
	for _, driver := range rankedDrivers {
		availableDrivers[driver.DriverID] = driver
	}

	sortedRequests := sortRequestsByComplexity(requests)

	for _, request := range sortedRequests {
		driverScores := make([]driverMatchScore, 0)

		for _, driver := range availableDrivers {
			distance := calculateHaversineDistance(
				driver.Distance, driver.Distance,
				request.PickupLat, request.PickupLon,
			)

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

			baseConfidence := driver.TotalScore / 100.0
			adjustedConfidence := baseConfidence + (distanceBonus / 100.0)

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

		sort.Slice(driverScores, func(i, j int) bool {
			return driverScores[i].Confidence > driverScores[j].Confidence
		})

		matched := false
		for _, match := range driverScores {
			if match.Confidence >= acceptanceThreshold {
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

func (m *Matcher) MatchSingleRequestToDriver(
	ctx context.Context,
	request dto.RideRequestInfo,
	availableDriverIDs []string,
	acceptanceThreshold float64,
) (*dto.OptimalAssignment, error) {

	rankedDrivers, err := m.ranker.RankDrivers(ctx, availableDriverIDs, request.PickupLat, request.PickupLon)
	if err != nil {
		return nil, err
	}

	if len(rankedDrivers) == 0 {
		return nil, nil
	}

	for _, driver := range rankedDrivers {
		distance := calculateHaversineDistance(
			driver.Distance, driver.Distance,
			request.PickupLat, request.PickupLon,
		)

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

func sortRequestsByComplexity(requests []dto.RideRequestInfo) []dto.RideRequestInfo {
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
		return distI > distJ
	})

	return sorted
}

type driverMatchScore struct {
	Driver     dto.DriverRankingScore
	Distance   float64 
	Confidence float64 
}
