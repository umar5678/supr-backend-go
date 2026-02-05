package pricing

import (
	"math"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/utils/location"
)

type FareCalculator struct{}

func NewFareCalculator() *FareCalculator {
	return &FareCalculator{}
}

func (c *FareCalculator) CalculateEstimate(
	pickupLat, pickupLon, dropoffLat, dropoffLon float64,
	vehicleType *models.VehicleType,
	surgeMultiplier float64,
) *models.FareEstimate {

	distanceKm := location.HaversineDistance(pickupLat, pickupLon, dropoffLat, dropoffLon)

	estimatedDistance := distanceKm * 1.2

	averageSpeedKmh := 40.0
	estimatedDuration := location.CalculateETA(estimatedDistance, averageSpeedKmh)

	distanceFare := estimatedDistance * vehicleType.PerKmRate
	durationMinutes := float64(estimatedDuration) / 60.0
	durationFare := durationMinutes * vehicleType.PerMinuteRate

	subTotal := vehicleType.BaseFare + distanceFare + durationFare
	surgeAmount := subTotal * (surgeMultiplier - 1.0)
	totalFare := subTotal + surgeAmount + vehicleType.BookingFee

	totalFare = math.Round(totalFare*100) / 100

	return &models.FareEstimate{
		BaseFare:          vehicleType.BaseFare,
		DistanceFare:      math.Round(distanceFare*100) / 100,
		DurationFare:      math.Round(durationFare*100) / 100,
		BookingFee:        vehicleType.BookingFee,
		SurgeMultiplier:   surgeMultiplier,
		SubTotal:          math.Round(subTotal*100) / 100,
		SurgeAmount:       math.Round(surgeAmount*100) / 100,
		TotalFare:         totalFare,
		EstimatedDistance: math.Round(estimatedDistance*100) / 100,
		EstimatedDuration: estimatedDuration,
		VehicleTypeName:   vehicleType.DisplayName,
	}
}

func (c *FareCalculator) CalculateActualFare(
	actualDistanceKm float64,
	actualDurationSec int,
	vehicleType *models.VehicleType,
	surgeMultiplier float64,
) *models.FareEstimate {

	distanceFare := actualDistanceKm * vehicleType.PerKmRate
	durationMinutes := float64(actualDurationSec) / 60.0
	durationFare := durationMinutes * vehicleType.PerMinuteRate

	subTotal := vehicleType.BaseFare + distanceFare + durationFare
	surgeAmount := subTotal * (surgeMultiplier - 1.0)
	totalFare := subTotal + surgeAmount + vehicleType.BookingFee

	totalFare = math.Round(totalFare*100) / 100

	return &models.FareEstimate{
		BaseFare:          vehicleType.BaseFare,
		DistanceFare:      math.Round(distanceFare*100) / 100,
		DurationFare:      math.Round(durationFare*100) / 100,
		BookingFee:        vehicleType.BookingFee,
		SurgeMultiplier:   surgeMultiplier,
		SubTotal:          math.Round(subTotal*100) / 100,
		SurgeAmount:       math.Round(surgeAmount*100) / 100,
		TotalFare:         totalFare,
		EstimatedDistance: actualDistanceKm,
		EstimatedDuration: actualDurationSec,
		VehicleTypeName:   vehicleType.DisplayName,
	}
}

func (c *FareCalculator) CalculateMinimumFare(vehicleType *models.VehicleType) float64 {
	return vehicleType.BaseFare + vehicleType.BookingFee
}
