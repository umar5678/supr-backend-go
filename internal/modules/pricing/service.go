// internal/modules/pricing/service.go
package pricing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/pricing/dto"
	vehiclesrepo "github.com/umar5678/go-backend/internal/modules/vehicles"
	"github.com/umar5678/go-backend/internal/services/cache"
	"gorm.io/gorm"

	"github.com/umar5678/go-backend/internal/utils/location"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
	GetFareEstimate(ctx context.Context, req dto.FareEstimateRequest) (*dto.FareEstimateResponse, error)
	CalculateActualFare(ctx context.Context, req dto.CalculateActualFareRequest) (*dto.FareEstimateResponse, error)
	GetSurgeMultiplier(ctx context.Context, lat, lon float64) (float64, error)
	GetActiveSurgeZones(ctx context.Context) ([]*dto.SurgeZoneResponse, error)
	CreateSurgeZone(ctx context.Context, req dto.CreateSurgeZoneRequest) (*dto.CreateSurgeZoneResponse, error)
	GetFareBreakdown(ctx context.Context, req dto.GetFareBreakdownRequest) (*dto.FareBreakdownResponse, error)
	CalculateWaitTimeCharge(ctx context.Context, rideID string, arrivedAt time.Time) (*dto.WaitTimeChargeResponse, error)
	ChangeDestination(ctx context.Context, driverID string, req dto.ChangeDestinationRequest) (*dto.DestinationChangeResponse, error)
	ApplyPriceCapping(ctx context.Context, vehicleTypeID string, calculatedFare float64) (*dto.FareBreakdownResponse, error)

	CalculateCombinedSurge(ctx context.Context, vehicleTypeID, geohash string, lat, lon float64) (*dto.SurgeCalculationResponse, error)
	CreateSurgePricingRule(ctx context.Context, req dto.CreateSurgePricingRuleRequest) (*dto.SurgePricingRuleResponse, error)
	GetActiveSurgePricingRules(ctx context.Context) ([]*dto.SurgePricingRuleResponse, error)
	GetCurrentDemand(ctx context.Context, geohash string) (*dto.DemandTrackingResponse, error)
	CalculateETAEstimate(ctx context.Context, req dto.ETAEstimateRequest) (*dto.ETAEstimateResponse, error)
}

type service struct {
	repo         Repository
	db           *gorm.DB
	vehiclesRepo vehiclesrepo.Repository
	calculator   *FareCalculator
	surgeManager *SurgeManager
}

func NewService(repo Repository, db *gorm.DB, vehiclesRepo vehiclesrepo.Repository) Service {
	return &service{
		repo:         repo,
		vehiclesRepo: vehiclesRepo,
		db:           db,
		calculator:   NewFareCalculator(),
		surgeManager: NewSurgeManager(repo),
	}
}

func (s *service) GetFareEstimate(ctx context.Context, req dto.FareEstimateRequest) (*dto.FareEstimateResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	distance := location.HaversineDistance(req.PickupLat, req.PickupLon, req.DropoffLat, req.DropoffLon)
	if distance < 0.5 {
		return nil, response.BadRequest("Minimum trip distance is 0.5 km")
	}

	if distance > 100 {
		return nil, response.BadRequest("Maximum trip distance is 100 km")
	}

	vehicleType, err := s.vehiclesRepo.FindByID(ctx, req.VehicleTypeID)
	if err != nil {
		return nil, response.NotFoundError("Vehicle type")
	}

	if !vehicleType.IsActive {
		return nil, response.BadRequest("Vehicle type is not available")
	}

	geohash := fmt.Sprintf("%.1f_%.1f", req.PickupLat, req.PickupLon)
	combinedMultiplier, timeMultiplier, demandMultiplier, reason, err := s.surgeManager.CalculateCombinedSurge(ctx, req.VehicleTypeID, geohash, req.PickupLat, req.PickupLon)
	if err != nil {
		logger.Warn("surge calculation failed", "error", err)
		combinedMultiplier = 1.0
		timeMultiplier = 1.0
		demandMultiplier = 1.0
		reason = "normal"
	}

	surgeMultiplier := combinedMultiplier

	estimate := s.calculator.CalculateEstimate(
		req.PickupLat, req.PickupLon,
		req.DropoffLat, req.DropoffLon,
		vehicleType,
		surgeMultiplier,
	)

	fareResponse := &dto.FareEstimateResponse{
		BaseFare:          estimate.BaseFare,
		DistanceFare:      estimate.DistanceFare,
		DurationFare:      estimate.DurationFare,
		BookingFee:        estimate.BookingFee,
		SurgeMultiplier:   surgeMultiplier,
		SubTotal:          estimate.SubTotal,
		SurgeAmount:       estimate.SurgeAmount,
		TotalFare:         estimate.TotalFare,
		EstimatedDistance: estimate.EstimatedDistance,
		EstimatedDuration: estimate.EstimatedDuration,
		VehicleTypeName:   estimate.VehicleTypeName,
		Currency:          "INR",

		DriverPayout:       estimate.TotalFare,
		PlatformCommission: 0,
		CommissionRate:     0,

		SurgeDetails: &dto.SurgeDetailsResponse{
			IsActive:              surgeMultiplier > 1.0,
			AppliedMultiplier:     combinedMultiplier,
			TimeBasedMultiplier:   timeMultiplier,
			DemandBasedMultiplier: demandMultiplier,
			Reason:                reason,
		},
	}

	cacheKey := fmt.Sprintf("fare:estimate:%s:%f:%f:%f:%f",
		req.VehicleTypeID, req.PickupLat, req.PickupLon, req.DropoffLat, req.DropoffLon)
	cache.SetJSON(ctx, cacheKey, fareResponse, 1*time.Minute)

	logger.Info("fare estimate calculated",
		"vehicleType", vehicleType.Name,
		"distance", estimate.EstimatedDistance,
		"duration", estimate.EstimatedDuration,
		"surge", surgeMultiplier,
		"surgeReason", reason,
		"totalFare", estimate.TotalFare,
		"driverPayout", estimate.TotalFare,
		"platformCommission", 0,
	)

	return fareResponse, nil
}

func (s *service) CalculateActualFare(ctx context.Context, req dto.CalculateActualFareRequest) (*dto.FareEstimateResponse, error) {
	if req.ActualDistanceKm < 0 {
		return nil, response.BadRequest("Invalid distance")
	}
	if req.ActualDurationSec < 0 {
		return nil, response.BadRequest("Invalid duration")
	}

	const maxRideDurationSeconds = 12 * 60 * 60
	if req.ActualDurationSec > maxRideDurationSeconds {
		logger.Error("invalid ride duration - exceeds 12 hours",
			"durationSeconds", req.ActualDurationSec,
			"durationMinutes", float64(req.ActualDurationSec)/60.0,
			"durationHours", float64(req.ActualDurationSec)/3600.0,
		)
		return nil, response.BadRequest("Ride duration cannot exceed 12 hours")
	}

	vehicleType, err := s.vehiclesRepo.FindByID(ctx, req.VehicleTypeID)
	if err != nil {
		return nil, response.NotFoundError("Vehicle type")
	}

	surgeMultiplier := req.SurgeMultiplier
	if surgeMultiplier == 0 {
		surgeMultiplier = 1.0
	}

	estimate := s.calculator.CalculateActualFare(
		req.ActualDistanceKm,
		req.ActualDurationSec,
		vehicleType,
		surgeMultiplier,
	)

	fareResponse := &dto.FareEstimateResponse{
		BaseFare:           estimate.BaseFare,
		DistanceFare:       estimate.DistanceFare,
		DurationFare:       estimate.DurationFare,
		BookingFee:         estimate.BookingFee,
		SurgeMultiplier:    estimate.SurgeMultiplier,
		SubTotal:           estimate.SubTotal,
		SurgeAmount:        estimate.SurgeAmount,
		TotalFare:          estimate.TotalFare,
		DriverPayout:       estimate.TotalFare,
		PlatformCommission: 0,
		CommissionRate:     0,
		EstimatedDistance:  estimate.EstimatedDistance,
		EstimatedDuration:  estimate.EstimatedDuration,
		VehicleTypeName:    estimate.VehicleTypeName,
		Currency:           "INR",
	}

	logger.Info("actual fare calculated",
		"vehicleType", vehicleType.Name,
		"distance", req.ActualDistanceKm,
		"durationSeconds", req.ActualDurationSec,
		"durationMinutes", float64(req.ActualDurationSec)/60.0,
		"distanceFare", estimate.DistanceFare,
		"durationFare", estimate.DurationFare,
		"surge", surgeMultiplier,
		"totalFare", estimate.TotalFare,
	)

	return fareResponse, nil
}

func (s *service) GetSurgeMultiplier(ctx context.Context, lat, lon float64) (float64, error) {
	if err := location.ValidateCoordinates(lat, lon); err != nil {
		return 1.0, response.BadRequest(err.Error())
	}

	geohash := location.Encode(lat, lon, 7)
	cacheKey := fmt.Sprintf("surge:zone:%s", geohash)

	var cached float64
	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil {
		return cached, nil
	}

	multiplier, err := s.surgeManager.GetSurgeMultiplier(ctx, lat, lon)
	if err != nil {
		logger.Error("failed to get surge multiplier", "error", err)
		return 1.0, nil
	}

	cache.Set(ctx, cacheKey, fmt.Sprintf("%.2f", multiplier), 5*time.Minute)

	return multiplier, nil
}

func (s *service) GetActiveSurgeZones(ctx context.Context) ([]*dto.SurgeZoneResponse, error) {
	cacheKey := "surge:zones:active"
	var cached []*dto.SurgeZoneResponse

	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil {
		return cached, nil
	}

	zones, err := s.repo.GetAllActiveSurgeZones(ctx)
	if err != nil {
		logger.Error("failed to get active surge zones", "error", err)
		return nil, response.InternalServerError("Failed to fetch surge zones", err)
	}

	result := make([]*dto.SurgeZoneResponse, len(zones))
	for i, zone := range zones {
		result[i] = &dto.SurgeZoneResponse{
			ID:         zone.ID,
			AreaName:   zone.AreaName,
			Geohash:    zone.AreaGeohash,
			CenterLat:  zone.CenterLat,
			CenterLon:  zone.CenterLon,
			Multiplier: zone.Multiplier,
			RadiusKm:   zone.RadiusKm,
			IsActive:   zone.IsActive,
		}
	}

	cache.SetJSON(ctx, cacheKey, result, 5*time.Minute)

	return result, nil
}

func (s *service) CreateSurgeZone(ctx context.Context, req dto.CreateSurgeZoneRequest) (*dto.CreateSurgeZoneResponse, error) {
	activeFrom, err := time.Parse(time.RFC3339, req.ActiveFrom)
	if err != nil {
		return nil, response.BadRequest("Invalid activeFrom timestamp format (expected RFC3339)")
	}

	activeUntil, err := time.Parse(time.RFC3339, req.ActiveUntil)
	if err != nil {
		return nil, response.BadRequest("Invalid activeUntil timestamp format (expected RFC3339)")
	}

	if activeUntil.Before(activeFrom) {
		return nil, response.BadRequest("activeUntil must be after activeFrom")
	}

	zone := &models.SurgePricingZone{
		AreaName:    req.AreaName,
		AreaGeohash: req.AreaGeohash,
		CenterLat:   req.CenterLat,
		CenterLon:   req.CenterLon,
		RadiusKm:    req.RadiusKm,
		Multiplier:  req.Multiplier,
		ActiveFrom:  activeFrom,
		ActiveUntil: activeUntil,
		IsActive:    req.IsActive,
	}

	if err := s.repo.CreateSurgeZone(ctx, zone); err != nil {
		logger.Error("failed to create surge zone", "error", err)
		return nil, response.InternalServerError("Failed to create surge zone", err)
	}

	cacheKey := "surge:zones:active"
	cache.Delete(ctx, cacheKey)

	result := &dto.CreateSurgeZoneResponse{
		ID:          zone.ID,
		AreaName:    zone.AreaName,
		AreaGeohash: zone.AreaGeohash,
		CenterLat:   zone.CenterLat,
		CenterLon:   zone.CenterLon,
		RadiusKm:    zone.RadiusKm,
		Multiplier:  zone.Multiplier,
		IsActive:    zone.IsActive,
		ActiveFrom:  zone.ActiveFrom,
		ActiveUntil: zone.ActiveUntil,
		CreatedAt:   zone.CreatedAt,
	}

	return result, nil
}

func (s *service) GetFareBreakdown(ctx context.Context, req dto.GetFareBreakdownRequest) (*dto.FareBreakdownResponse, error) {
	vehicleType, err := s.vehiclesRepo.FindByID(ctx, req.VehicleTypeID)
	if err != nil {
		return nil, response.NotFoundError("Vehicle type")
	}

	distance := location.HaversineDistance(req.PickupLat, req.PickupLon, req.DropoffLat, req.DropoffLon)

	duration := int((distance / 30.0) * 60)

	// Get combined surge with detailed breakdown
	geohash := fmt.Sprintf("%.1f_%.1f", req.PickupLat, req.PickupLon)
	combinedMultiplier, timeMultiplier, demandMultiplier, surgeReason, err := s.surgeManager.CalculateCombinedSurge(ctx, req.VehicleTypeID, geohash, req.PickupLat, req.PickupLon)
	if err != nil {
		logger.Warn("failed to calculate combined surge, using basic surge", "error", err)
		combinedMultiplier = 1.0
		timeMultiplier = 1.0
		demandMultiplier = 1.0
		surgeReason = "normal"
	}

	// Also get zone-based surge if available
	zoneMultiplier, _ := s.surgeManager.CalculateZoneBasedSurge(ctx, req.PickupLat, req.PickupLon)
	if zoneMultiplier == 0 {
		zoneMultiplier = 1.0
	}

	surgeMultiplier := combinedMultiplier

	logger.Info("surge details retrieved for fare breakdown",
		"lat", req.PickupLat,
		"lon", req.PickupLon,
		"combinedMultiplier", combinedMultiplier,
		"timeMultiplier", timeMultiplier,
		"demandMultiplier", demandMultiplier,
		"zoneMultiplier", zoneMultiplier,
		"surgeReason", surgeReason,
	)

	baseFare := vehicleType.BaseFare
	distanceCharge := distance * vehicleType.PerKmRate
	timeCharge := float64(duration) * vehicleType.PerMinuteRate
	bookingFee := vehicleType.BookingFee

	subTotal := baseFare + distanceCharge + timeCharge + bookingFee
	surgeCharge := subTotal * (surgeMultiplier - 1.0)
	totalFare := subTotal + surgeCharge

	logger.Info("fare breakdown calculation",
		"baseFare", baseFare,
		"distanceCharge", distanceCharge,
		"timeCharge", timeCharge,
		"bookingFee", bookingFee,
		"subTotal", subTotal,
		"surgeMultiplier", surgeMultiplier,
		"surgeCharge", surgeCharge,
		"totalFare", totalFare,
	)

	components := []dto.FareComponent{
		{
			Name:   "Base Fare",
			Amount: baseFare,
			Type:   "base",
		},
		{
			Name:   fmt.Sprintf("Distance (%.2f km)", distance),
			Amount: distanceCharge,
			Type:   "distance",
		},
		{
			Name:   fmt.Sprintf("Duration (%d min)", duration),
			Amount: timeCharge,
			Type:   "duration",
		},
		{
			Name:   "Booking Fee",
			Amount: bookingFee,
			Type:   "booking_fee",
		},
	}

	if surgeMultiplier > 1.0 {
		components = append(components, dto.FareComponent{
			Name:   fmt.Sprintf("Surge (%.1fx)", surgeMultiplier),
			Amount: surgeCharge,
			Type:   "surge",
		})
	}

	breakdown := &dto.FareBreakdownResponse{
		Components:      components,
		BaseFare:        baseFare,
		DistanceCharge:  distanceCharge,
		TimeCharge:      timeCharge,
		BookingFee:      bookingFee,
		SurgeCharge:     surgeCharge,
		SurgeMultiplier: surgeMultiplier,
		SurgeDetails: &dto.SurgeDetailsResponse{
			IsActive:              surgeMultiplier > 1.0,
			AppliedMultiplier:     combinedMultiplier,
			TimeBasedMultiplier:   timeMultiplier,
			DemandBasedMultiplier: demandMultiplier,
			ZoneBasedMultiplier:   zoneMultiplier,
			Reason:                surgeReason,
		},
		SubTotal:          subTotal,
		TotalFare:         totalFare,
		EstimatedDistance: distance,
		EstimatedDuration: duration,
	}

	rule, err := s.repo.FindPriceCappingRule(ctx, req.VehicleTypeID)
	if err == nil && rule != nil {
		if totalFare > rule.MaxCustomerPrice {
			breakdown.CustomerPrice = rule.MaxCustomerPrice
			breakdown.DriverEarning = rule.MaxDriverEarning
			breakdown.PlatformFee = breakdown.CustomerPrice - breakdown.DriverEarning
			breakdown.PlatformAbsorbed = totalFare - breakdown.CustomerPrice
			breakdown.PriceCapped = true
		} else {
			breakdown.CustomerPrice = totalFare
			breakdown.DriverEarning = totalFare
			breakdown.PlatformFee = 0.0
			breakdown.PriceCapped = false
		}
	} else {
		breakdown.CustomerPrice = totalFare
		breakdown.DriverEarning = totalFare
		breakdown.PlatformFee = 0.0
		breakdown.PriceCapped = false
	}

	return breakdown, nil
}

func (s *service) ApplyPriceCapping(ctx context.Context, vehicleTypeID string, calculatedFare float64) (*dto.FareBreakdownResponse, error) {
	rule, err := s.repo.FindPriceCappingRule(ctx, vehicleTypeID)
	if err != nil {
		return &dto.FareBreakdownResponse{
			TotalFare:     calculatedFare,
			CustomerPrice: calculatedFare,
			DriverEarning: calculatedFare,
			PlatformFee:   0.0,
			PriceCapped:   false,
		}, nil
	}

	breakdown := &dto.FareBreakdownResponse{
		TotalFare: calculatedFare,
	}

	if calculatedFare > rule.MaxCustomerPrice {
		breakdown.CustomerPrice = rule.MaxCustomerPrice
		breakdown.DriverEarning = rule.MaxDriverEarning
		breakdown.PlatformFee = breakdown.CustomerPrice - breakdown.DriverEarning
		breakdown.PlatformAbsorbed = calculatedFare - breakdown.CustomerPrice
		breakdown.PriceCapped = true

		logger.Info("price capped",
			"originalFare", calculatedFare,
			"cappedFare", breakdown.CustomerPrice,
			"driverEarning", breakdown.DriverEarning,
			"platformAbsorbed", breakdown.PlatformAbsorbed,
		)
	} else {
		breakdown.CustomerPrice = calculatedFare
		breakdown.DriverEarning = calculatedFare
		breakdown.PlatformFee = 0.0
		breakdown.PriceCapped = false
	}

	return breakdown, nil
}

func (s *service) CalculateDriverPayout(totalFare float64, commissionRate float64) (driverAmount, platformCommission float64) {
	platformCommission = totalFare * (commissionRate / 100.0)
	driverAmount = totalFare - platformCommission
	return driverAmount, platformCommission
}

func (s *service) CalculateWaitTimeCharge(ctx context.Context, rideID string, arrivedAt time.Time) (*dto.WaitTimeChargeResponse, error) {
	var ride models.Ride
	if err := s.db.WithContext(ctx).Where("id = ?", rideID).First(&ride).Error; err != nil {
		return nil, response.NotFoundError("Ride")
	}

	waitMinutes := int(time.Since(arrivedAt).Minutes())
	freeWaitMinutes := 3
	chargeAmount := 0.0

	if waitMinutes > freeWaitMinutes {
		chargeableMinutes := waitMinutes - freeWaitMinutes
		chargeAmount = float64(chargeableMinutes) * 1.00
	}

	waitCharge, err := s.repo.FindWaitTimeChargeByRideID(ctx, rideID)
	if err != nil {
		waitCharge = &models.WaitTimeCharge{
			RideID:           rideID,
			WaitStartedAt:    arrivedAt,
			TotalWaitMinutes: waitMinutes,
			ChargeAmount:     chargeAmount,
		}
		if err := s.repo.CreateWaitTimeCharge(ctx, waitCharge); err != nil {
			return nil, response.InternalServerError("Failed to create wait time charge", err)
		}
	} else {
		now := time.Now()
		waitCharge.WaitEndedAt = &now
		waitCharge.TotalWaitMinutes = waitMinutes
		waitCharge.ChargeAmount = chargeAmount
		s.repo.UpdateWaitTimeCharge(ctx, waitCharge)
	}

	s.repo.UpdateRideWaitTimeCharge(ctx, rideID, chargeAmount)

	logger.Info("wait time charge calculated",
		"rideID", rideID,
		"waitMinutes", waitMinutes,
		"chargeAmount", chargeAmount,
	)

	return &dto.WaitTimeChargeResponse{
		RideID:           rideID,
		TotalWaitMinutes: waitMinutes,
		ChargeAmount:     chargeAmount,
		FreeWaitMinutes:  freeWaitMinutes,
	}, nil
}

func (s *service) ChangeDestination(ctx context.Context, driverID string, req dto.ChangeDestinationRequest) (*dto.DestinationChangeResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	var ride models.Ride
	if err := s.db.WithContext(ctx).
		Preload("VehicleType").
		Where("id = ?", req.RideID).
		First(&ride).Error; err != nil {
		return nil, response.NotFoundError("Ride")
	}

	if ride.DriverID == nil || *ride.DriverID != driverID {
		return nil, response.ForbiddenError("Unauthorized")
	}

	if ride.Status != "started" {
		return nil, response.BadRequest("Can only change destination during active ride")
	}

	additionalDistance := location.HaversineDistance(
		ride.DropoffLat, ride.DropoffLon,
		req.NewLatitude, req.NewLongitude,
	)

	additionalCharge := additionalDistance * ride.VehicleType.PerKmRate

	if err := s.repo.UpdateRideDestination(ctx, req.RideID,
		req.NewLatitude, req.NewLongitude, req.NewAddress, additionalCharge); err != nil {
		return nil, response.InternalServerError("Failed to update destination", err)
	}

	newTotalFare := ride.EstimatedFare + additionalCharge

	logger.Info("destination changed",
		"rideID", req.RideID,
		"additionalDistance", additionalDistance,
		"additionalCharge", additionalCharge,
	)

	return &dto.DestinationChangeResponse{
		RideID:             req.RideID,
		AdditionalDistance: additionalDistance,
		AdditionalCharge:   additionalCharge,
		NewTotalFare:       newTotalFare,
	}, nil
}

func (s *service) CalculateCombinedSurge(ctx context.Context, vehicleTypeID, geohash string, lat, lon float64) (*dto.SurgeCalculationResponse, error) {
	combined, timeSurge, demandSurge, reason, err := s.surgeManager.CalculateCombinedSurge(ctx, vehicleTypeID, geohash, lat, lon)
	if err != nil {
		logger.Warn("surge manager calculation failed", "error", err)
		return &dto.SurgeCalculationResponse{
			AppliedMultiplier:     1.0,
			TimeBasedMultiplier:   1.0,
			DemandBasedMultiplier: 1.0,
			Reason:                "normal",
			BaseFare:              0,
			SurgeAmount:           0,
			TotalFare:             0,
			Details: dto.SurgeDetails{
				TimeOfDay: "normal",
				DayType:   "weekday",
			},
		}, nil
	}

	return &dto.SurgeCalculationResponse{
		AppliedMultiplier:     combined,
		TimeBasedMultiplier:   timeSurge,
		DemandBasedMultiplier: demandSurge,
		Reason:                reason,
		BaseFare:              0,
		SurgeAmount:           0,
		TotalFare:             0,
		Details: dto.SurgeDetails{
			TimeOfDay: "peak",
			DayType:   "weekday",
		},
	}, nil
}

func (s *service) CreateSurgePricingRule(ctx context.Context, req dto.CreateSurgePricingRuleRequest) (*dto.SurgePricingRuleResponse, error) {
	rule := &models.SurgePricingRule{
		ID:                         uuid.New().String(),
		Name:                       req.Name,
		Description:                req.Description,
		VehicleTypeID:              req.VehicleTypeID,
		DayOfWeek:                  req.DayOfWeek,
		StartTime:                  req.StartTime,
		EndTime:                    req.EndTime,
		BaseMultiplier:             req.BaseMultiplier,
		MinMultiplier:              req.MinMultiplier,
		MaxMultiplier:              req.MaxMultiplier,
		EnableDemandBasedSurge:     req.EnableDemandBasedSurge,
		DemandThreshold:            req.DemandThreshold,
		DemandMultiplierPerRequest: req.DemandMultiplierPerRequest,
		IsActive:                   true,
	}

	if err := s.repo.CreateSurgePricingRule(ctx, rule); err != nil {
		logger.Error("failed to create surge pricing rule", "error", err)
		return nil, response.InternalServerError("Failed to create rule", err)
	}

	return &dto.SurgePricingRuleResponse{
		ID:                         rule.ID,
		Name:                       rule.Name,
		Description:                rule.Description,
		VehicleTypeID:              rule.VehicleTypeID,
		DayOfWeek:                  rule.DayOfWeek,
		StartTime:                  rule.StartTime,
		EndTime:                    rule.EndTime,
		BaseMultiplier:             rule.BaseMultiplier,
		MinMultiplier:              rule.MinMultiplier,
		MaxMultiplier:              rule.MaxMultiplier,
		EnableDemandBasedSurge:     rule.EnableDemandBasedSurge,
		DemandThreshold:            rule.DemandThreshold,
		DemandMultiplierPerRequest: rule.DemandMultiplierPerRequest,
		IsActive:                   rule.IsActive,
		CreatedAt:                  rule.CreatedAt,
		UpdatedAt:                  rule.UpdatedAt,
	}, nil
}

func (s *service) GetActiveSurgePricingRules(ctx context.Context) ([]*dto.SurgePricingRuleResponse, error) {
	rules, err := s.repo.GetActiveSurgePricingRules(ctx)
	if err != nil {
		logger.Error("failed to get surge pricing rules", "error", err)
		return nil, response.InternalServerError("Failed to get rules", err)
	}

	var responses []*dto.SurgePricingRuleResponse
	for _, rule := range rules {
		responses = append(responses, &dto.SurgePricingRuleResponse{
			ID:                         rule.ID,
			Name:                       rule.Name,
			Description:                rule.Description,
			VehicleTypeID:              rule.VehicleTypeID,
			DayOfWeek:                  rule.DayOfWeek,
			StartTime:                  rule.StartTime,
			EndTime:                    rule.EndTime,
			BaseMultiplier:             rule.BaseMultiplier,
			MinMultiplier:              rule.MinMultiplier,
			MaxMultiplier:              rule.MaxMultiplier,
			EnableDemandBasedSurge:     rule.EnableDemandBasedSurge,
			DemandThreshold:            rule.DemandThreshold,
			DemandMultiplierPerRequest: rule.DemandMultiplierPerRequest,
			IsActive:                   rule.IsActive,
			CreatedAt:                  rule.CreatedAt,
			UpdatedAt:                  rule.UpdatedAt,
		})
	}

	return responses, nil
}

func (s *service) GetCurrentDemand(ctx context.Context, geohash string) (*dto.DemandTrackingResponse, error) {
	demand, err := s.repo.GetLatestDemandByGeohash(ctx, geohash)
	if err != nil {
		logger.Error("failed to get demand tracking", "error", err)
		return nil, response.NotFoundError("Demand data")
	}

	return &dto.DemandTrackingResponse{
		ID:                string(demand.ID),
		ZoneID:            demand.ZoneID,
		ZoneGeohash:       demand.ZoneGeohash,
		PendingRequests:   demand.PendingRequests,
		AvailableDrivers:  demand.AvailableDrivers,
		CompletedRides:    demand.CompletedRides,
		AverageWaitTime:   demand.AverageWaitTime,
		DemandSupplyRatio: demand.DemandSupplyRatio,
		SurgeMultiplier:   demand.SurgeMultiplier,
		RecordedAt:        demand.RecordedAt,
	}, nil
}

func (s *service) CalculateETAEstimate(ctx context.Context, req dto.ETAEstimateRequest) (*dto.ETAEstimateResponse, error) {
	distance := location.HaversineDistance(req.PickupLat, req.PickupLon, req.DropoffLat, req.DropoffLon)

	if distance < 0.5 {
		return nil, response.BadRequest("Minimum trip distance is 0.5 km")
	}
	estimatedSpeedKmh := 40.0
	estimatedDurationHours := distance / estimatedSpeedKmh
	estimatedDurationSeconds := int(estimatedDurationHours * 3600)

	estimatedPickupETA := 5 * 60

	estimatedDropoffETA := estimatedPickupETA + estimatedDurationSeconds

	eta := &models.ETAEstimate{
		ID:                  uuid.New().String(),
		RideID:              nil,
		PickupLat:           req.PickupLat,
		PickupLon:           req.PickupLon,
		DropoffLat:          req.DropoffLat,
		DropoffLon:          req.DropoffLon,
		DistanceKm:          distance,
		DurationSeconds:     estimatedDurationSeconds,
		EstimatedPickupETA:  estimatedPickupETA,
		EstimatedDropoffETA: estimatedDropoffETA,
		TrafficCondition:    "normal",
		TrafficMultiplier:   1.0,
		Source:              "calculated",
	}

	if err := s.repo.CreateETAEstimate(ctx, eta); err != nil {
		logger.Warn("failed to create eta estimate (non-fatal)", "error", err)
	}

	return &dto.ETAEstimateResponse{
		ID:                  eta.ID,
		DistanceKm:          eta.DistanceKm,
		DurationSeconds:     eta.DurationSeconds,
		EstimatedPickupETA:  eta.EstimatedPickupETA,
		EstimatedDropoffETA: eta.EstimatedDropoffETA,
		TrafficCondition:    eta.TrafficCondition,
		TrafficMultiplier:   eta.TrafficMultiplier,
		Source:              eta.Source,
		CreatedAt:           eta.CreatedAt,
	}, nil
}

func (s *service) isRuleActiveNow(rule *models.SurgePricingRule) bool {
	return true
}

func (s *service) calculateDemandMultiplier(demand *models.DemandTracking, rule *models.SurgePricingRule) float64 {
	if demand == nil {
		return rule.BaseMultiplier
	}
	return demand.SurgeMultiplier
}

func (s *service) getSurgeReason(multiplier float64) string {
	if multiplier > 2.0 {
		return "high_demand"
	} else if multiplier > 1.5 {
		return "medium_demand"
	} else if multiplier > 1.0 {
		return "low_demand"
	}
	return "normal"
}
