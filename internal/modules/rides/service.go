package rides

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/umar5678/go-backend/internal/models"
	adminrepo "github.com/umar5678/go-backend/internal/modules/admin"
	batchingservice "github.com/umar5678/go-backend/internal/modules/batching"
	batchingdto "github.com/umar5678/go-backend/internal/modules/batching/dto"
	driversrepo "github.com/umar5678/go-backend/internal/modules/drivers"
	fraudservice "github.com/umar5678/go-backend/internal/modules/fraud"
	pricingservice "github.com/umar5678/go-backend/internal/modules/pricing"
	pricingdto "github.com/umar5678/go-backend/internal/modules/pricing/dto"
	"github.com/umar5678/go-backend/internal/modules/profile"
	profiledto "github.com/umar5678/go-backend/internal/modules/profile/dto"
	promotionsservice "github.com/umar5678/go-backend/internal/modules/promotions"
	promotionsdto "github.com/umar5678/go-backend/internal/modules/promotions/dto"
	ratingsservice "github.com/umar5678/go-backend/internal/modules/ratings"
	ridepinservice "github.com/umar5678/go-backend/internal/modules/ridepin"
	ridersrepo "github.com/umar5678/go-backend/internal/modules/riders"
	"github.com/umar5678/go-backend/internal/modules/rides/dto"
	"github.com/umar5678/go-backend/internal/modules/sos"
	sosdto "github.com/umar5678/go-backend/internal/modules/sos/dto"
	trackingservice "github.com/umar5678/go-backend/internal/modules/tracking"
	trackingdto "github.com/umar5678/go-backend/internal/modules/tracking/dto"
	walletservice "github.com/umar5678/go-backend/internal/modules/wallet"
	walletdto "github.com/umar5678/go-backend/internal/modules/wallet/dto"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/location"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
	"github.com/umar5678/go-backend/internal/websocket"
	websocketutil "github.com/umar5678/go-backend/internal/websocket/websocketutils"
	"gorm.io/gorm"
)

type Service interface {
	CreateRide(ctx context.Context, riderID string, req dto.CreateRideRequest) (*dto.RideResponse, error)
	GetRide(ctx context.Context, userID, rideID string) (*dto.RideResponse, error)
	ListRides(ctx context.Context, userID string, role string, req dto.ListRidesRequest) ([]*dto.RideListResponse, int64, error)
	CancelRide(ctx context.Context, userID, rideID string, req dto.CancelRideRequest) error

	GetAvailableCars(ctx context.Context, riderID string, req dto.AvailableCarRequest) (*dto.AvailableCarsListResponse, error)
	GetVehiclesWithDetails(ctx context.Context, riderID string, req dto.VehicleDetailsRequest) (*dto.VehiclesWithDetailsListResponse, error)

	AcceptRide(ctx context.Context, driverID, rideID string) (*dto.RideResponse, error)
	RejectRide(ctx context.Context, driverID, rideID string, req dto.RejectRideRequest) error
	MarkArrived(ctx context.Context, driverID, rideID string) (*dto.RideResponse, error)
	StartRide(ctx context.Context, driverID, rideID string, req dto.StartRideRequest) (*dto.RideResponse, error)
	CompleteRide(ctx context.Context, driverID, rideID string, req dto.CompleteRideRequest) (*dto.RideResponse, error)

	TriggerSOS(ctx context.Context, riderID, rideID string, latitude, longitude float64) error

	FindDriverForRide(ctx context.Context, rideID string) error
	ProcessRideRequestTimeout(ctx context.Context, requestID string) error
}

type service struct {
	repo              Repository
	adminRepo         adminrepo.Repository
	driversRepo       driversrepo.Repository
	ridersRepo        ridersrepo.Repository
	pricingService    pricingservice.Service
	trackingService   trackingservice.Service
	walletService     walletservice.Service
	ridePINService    ridepinservice.Service    
	profileService    profile.Service           
	sosService        sos.Service               
	promotionsService promotionsservice.Service 
	ratingsService    ratingsservice.Service    
	fraudService      fraudservice.Service      
	batchingService   batchingservice.Service   
	wsHelper          *RideWebSocketHelper
}

func NewService(
	repo Repository,
	driversRepo driversrepo.Repository,
	ridersRepo ridersrepo.Repository,
	pricingService pricingservice.Service,
	trackingService trackingservice.Service,
	walletService walletservice.Service,
	ridePINService ridepinservice.Service,
	profileService profile.Service,
	sosService sos.Service,
	promotionsService promotionsservice.Service,
	ratingsService ratingsservice.Service,
	fraudService fraudservice.Service,
	batchingService batchingservice.Service,
	adminRepo adminrepo.Repository,
) Service {
	svc := &service{
		repo:              repo,
		adminRepo:         adminRepo,
		driversRepo:       driversRepo,
		ridersRepo:        ridersRepo,
		pricingService:    pricingService,
		trackingService:   trackingService,
		walletService:     walletService,
		ridePINService:    ridePINService,
		profileService:    profileService,
		sosService:        sosService,
		promotionsService: promotionsService,
		ratingsService:    ratingsService,
		fraudService:      fraudService,
		batchingService:   batchingService,
		wsHelper:          NewRideWebSocketHelper(),
	}

	svc.wsHelper.SetService(svc)

	if batchingService != nil {
		batchingService.SetBatchExpireCallback(func(batchID string) {
			ctx := context.Background()
			logger.Info("Processing expired batch",
				"batchID", batchID,
			)
			if result, err := batchingService.ProcessBatch(ctx, batchID); err != nil {
				logger.Error("Failed to process batch",
					"batchID", batchID,
					"error", err,
				)
			} else {
				logger.Info("Batch processing complete",
					"batchID", batchID,
					"assignments", len(result.Assignments),
					"unmatched", len(result.UnmatchedIDs),
				)
				svc.processMatchingResult(ctx, result)
			}
		})
	}

	return svc
}

func (s *service) CreateRide(ctx context.Context, riderID string, req dto.CreateRideRequest) (*dto.RideResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	user, err := s.adminRepo.FindUserByID(ctx, riderID)
	if err == nil && user.EmergencyContactPhone == "" {
		logger.Warn("Ride created without emergency contact", "userID", riderID)
	}

	var scheduledAtPtr *time.Time
	isScheduled := false
	if req.ScheduledAt != "" {
		t, err := time.Parse(time.RFC3339, req.ScheduledAt)
		if err != nil {
			return nil, response.BadRequest("scheduledAt must be a valid RFC3339 timestamp")
		}
		if !t.After(time.Now()) {
			return nil, response.BadRequest("scheduledAt must be a future time")
		}
		scheduledAtPtr = &t
		isScheduled = true
		logger.Info("ride scheduled in advance", "riderID", riderID, "scheduledAt", t)
	}

	if req.SavedLocationID != nil && s.profileService != nil {
		savedLocs, err := s.profileService.GetSavedLocations(ctx, riderID)
		if err != nil {
			return nil, response.BadRequest("Failed to fetch saved locations")
		}

		var savedLoc *profiledto.SavedLocationResponse
		for _, loc := range savedLocs {
			if loc.ID == *req.SavedLocationID {
				savedLoc = loc
				break
			}
		}

		if savedLoc == nil {
			return nil, response.BadRequest("Invalid saved location")
		}

		if req.UseSavedAs == "pickup" {
			req.PickupLat = savedLoc.Latitude
			req.PickupLon = savedLoc.Longitude
			req.PickupAddress = savedLoc.Address
		} else if req.UseSavedAs == "dropoff" {
			req.DropoffLat = savedLoc.Latitude
			req.DropoffLon = savedLoc.Longitude
			req.DropoffAddress = savedLoc.Address
		}
	}

	fareReq := pricingdto.FareEstimateRequest{
		PickupLat:     req.PickupLat,
		PickupLon:     req.PickupLon,
		DropoffLat:    req.DropoffLat,
		DropoffLon:    req.DropoffLon,
		VehicleTypeID: req.VehicleTypeID,
	}

	fareEstimate, err := s.pricingService.GetFareEstimate(ctx, fareReq)
	if err != nil {
		return nil, err
	}

	geohash := fmt.Sprintf("%.1f_%.1f", req.PickupLat, req.PickupLon)
	surgeCalc, err := s.pricingService.CalculateCombinedSurge(ctx, req.VehicleTypeID, geohash, req.PickupLat, req.PickupLon)
	if err != nil {
		logger.Warn("failed to calculate combined surge, using basic surge", "error", err)
	} else if surgeCalc != nil && surgeCalc.AppliedMultiplier > fareEstimate.SurgeMultiplier {
		fareEstimate.SurgeMultiplier = surgeCalc.AppliedMultiplier
		fareEstimate.SurgeAmount = (fareEstimate.SubTotal) * (surgeCalc.AppliedMultiplier - 1.0)
		fareEstimate.TotalFare = fareEstimate.SubTotal + fareEstimate.SurgeAmount
		logger.Info("Enhanced surge applied",
			"timeBasedSurge", surgeCalc.TimeBasedMultiplier,
			"demandBasedSurge", surgeCalc.DemandBasedMultiplier,
			"combinedSurge", surgeCalc.AppliedMultiplier,
			"reason", surgeCalc.Reason)
	}

	etaReq := pricingdto.ETAEstimateRequest{
		PickupLat:  req.PickupLat,
		PickupLon:  req.PickupLon,
		DropoffLat: req.DropoffLat,
		DropoffLon: req.DropoffLon,
	}
	if etaEstimate, err := s.pricingService.CalculateETAEstimate(ctx, etaReq); err == nil && etaEstimate != nil {
		logger.Info("ETA calculated",
			"pickupETA", etaEstimate.EstimatedPickupETA,
			"dropoffETA", etaEstimate.EstimatedDropoffETA,
			"distance", etaEstimate.DistanceKm)
	} else if err != nil {
		logger.Warn("failed to calculate ETA", "error", err)
	}

	finalAmount := fareEstimate.TotalFare
	var promoDiscount float64
	var promoCodeID *string

	if req.PromoCode != "" {
		validateReq := promotionsdto.ValidatePromoCodeRequest{
			Code:       req.PromoCode,
			RideAmount: fareEstimate.TotalFare,
		}

		validation, err := s.promotionsService.ValidatePromoCode(ctx, riderID, validateReq)
		if err != nil {
			logger.Warn("Invalid promo code", "code", req.PromoCode, "error", err)
		} else if validation.Valid {
			finalAmount = validation.FinalAmount
			promoDiscount = validation.DiscountAmount

			promoCode, _ := s.promotionsService.GetPromoCode(ctx, req.PromoCode)
			if promoCode != nil {
				promoCodeID = &promoCode.ID
			}

			logger.Info("Promo code validated",
				"code", req.PromoCode,
				"discount", promoDiscount,
				"finalAmount", finalAmount)
		}
	}

	walletInfo, err := s.walletService.GetWallet(ctx, riderID)
	if err != nil {
		logger.Warn("could not get free credits info", "error", err, "riderID", riderID)
	} else if walletInfo != nil && walletInfo.FreeRideCredits > 0 {
		if walletInfo.FreeRideCredits >= finalAmount {
			logger.Info("ride fully covered by free credits", "riderID", riderID, "fareAmount", finalAmount, "freeCredits", walletInfo.FreeRideCredits)
		} else {
			logger.Info("ride partially covered by free credits", "riderID", riderID, "fareAmount", finalAmount, "freeCredits", walletInfo.FreeRideCredits)
		}
	}

	rideID := uuid.New().String()

	var holdID *string
	if !isScheduled && finalAmount > 0 {
		holdReq := walletdto.HoldFundsRequest{
			Amount:        finalAmount,
			ReferenceType: "ride",
			ReferenceID:   rideID,
			HoldDuration:  1800, 
		}

		holdResp, err := s.walletService.HoldFunds(ctx, riderID, holdReq)
		if err != nil {
			logger.Warn("failed to create cash payment tracking hold", "error", err, "rideID", rideID)

		} else {
			holdID = &holdResp.ID
			logger.Info("cash payment hold created for tracking", "rideID", rideID, "holdID", holdResp.ID, "amount", finalAmount)
		}
	} else if isScheduled {
		logger.Info("scheduled ride - cash payment tracking deferred until activation", "rideID", rideID)
	}

	status := "searching"
	if isScheduled {
		status = "scheduled"
	}

	ride := &models.Ride{
		ID:                rideID,
		RiderID:           riderID,
		VehicleTypeID:     req.VehicleTypeID,
		Status:            status,
		PickupLat:         req.PickupLat,
		PickupLon:         req.PickupLon,
		PickupAddress:     req.PickupAddress,
		DropoffLat:        req.DropoffLat,
		DropoffLon:        req.DropoffLon,
		DropoffAddress:    req.DropoffAddress,
		EstimatedDistance: fareEstimate.EstimatedDistance,
		EstimatedDuration: fareEstimate.EstimatedDuration,
		EstimatedFare:     finalAmount,
		SurgeMultiplier:   fareEstimate.SurgeMultiplier,
		WalletHoldID:      holdID,
		ScheduledAt:       scheduledAtPtr,
		IsScheduled:       isScheduled,
		RiderNotes:        req.RiderNotes,
		PromoCodeID:       promoCodeID,
		PromoDiscount:     &promoDiscount,
		RequestedAt:       time.Now(),
	}

	if err := s.repo.CreateRide(ctx, ride); err != nil {
		if holdID != nil {
			s.walletService.ReleaseHold(ctx, riderID, walletdto.ReleaseHoldRequest{HoldID: *holdID})
		}
		logger.Error("failed to create ride", "error", err, "riderID", riderID)
		return nil, response.InternalServerError("Failed to create ride", err)
	}

	cacheKey := fmt.Sprintf("ride:active:%s", rideID)
	if isScheduled {
		cacheKey = fmt.Sprintf("ride:scheduled:%s", rideID)
		cache.SetJSON(ctx, cacheKey, ride, 24*time.Hour)
		logger.Info("scheduled ride stored in cache", "rideID", rideID, "scheduledAt", scheduledAtPtr)
	} else {
		cache.SetJSON(ctx, cacheKey, ride, 30*time.Minute)
	}

	if isScheduled && scheduledAtPtr != nil {
		go func(rideID string, scheduledTime time.Time) {
			bgCtx := context.Background()
			wait := time.Until(scheduledTime)
			if wait > 0 {
				logger.Info("ride scheduled for future matching", "rideID", rideID, "waitSeconds", int(wait.Seconds()))
				time.Sleep(wait)
			}

			if err := s.repo.UpdateRideStatus(bgCtx, rideID, "searching"); err != nil {
				logger.Error("failed to update scheduled ride status to searching", "rideID", rideID, "error", err)
				return
			}
			logger.Info("scheduled ride activated, starting driver matching", "rideID", rideID)

			if s.batchingService != nil {
				ride, err := s.repo.FindRideByID(bgCtx, rideID)
				if err != nil {
					logger.Error("failed to fetch ride for scheduled matching", "rideID", rideID, "error", err)
					return
				}

				batchID, err := s.batchingService.AddRequestToBatch(bgCtx, struct {
					RideID        string
					RiderID       string
					PickupLat     float64
					PickupLon     float64
					DropoffLat    float64
					DropoffLon    float64
					PickupGeohash string
					VehicleTypeID string
					RequestedAt   time.Time
				}{
					RideID:        ride.ID,
					RiderID:       ride.RiderID,
					PickupLat:     ride.PickupLat,
					PickupLon:     ride.PickupLon,
					DropoffLat:    ride.DropoffLat,
					DropoffLon:    ride.DropoffLon,
					VehicleTypeID: ride.VehicleTypeID,
					RequestedAt:   time.Now(),
				})

				if err == nil {
					logger.Info("scheduled ride added to batch for matching",
						"batchID", batchID,
						"rideID", rideID,
					)
					if result, err := s.batchingService.ProcessBatch(bgCtx, batchID); err != nil {
						logger.Error("failed to process scheduled batch", "batchID", batchID, "error", err)
					} else {
						s.processMatchingResult(bgCtx, result)
					}
				}
			} else {
				if err := s.FindDriverForRide(bgCtx, rideID); err != nil {
					logger.Error("failed to find driver for scheduled ride", "rideID", rideID, "error", err)
				}
			}
		}(rideID, *scheduledAtPtr)
	} else {
		go func() {
			bgCtx := context.Background()

			if s.batchingService != nil {
				batchID, err := s.batchingService.AddRequestToBatch(bgCtx, struct {
					RideID        string
					RiderID       string
					PickupLat     float64
					PickupLon     float64
					DropoffLat    float64
					DropoffLon    float64
					PickupGeohash string
					VehicleTypeID string
					RequestedAt   time.Time
				}{
					RideID:        ride.ID,
					RiderID:       ride.RiderID,
					PickupLat:     ride.PickupLat,
					PickupLon:     ride.PickupLon,
					DropoffLat:    ride.DropoffLat,
					DropoffLon:    ride.DropoffLon,
					VehicleTypeID: ride.VehicleTypeID,
					RequestedAt:   time.Now(),
				})

				if err == nil {
					logger.Info("Request added to batch for intelligent matching",
						"batchID", batchID,
						"rideID", ride.ID,
						"riderID", ride.RiderID,
						"vehicleType", ride.VehicleTypeID,
					)

					if result, err := s.batchingService.ProcessBatch(bgCtx, batchID); err != nil {
						logger.Error("Failed to process batch",
							"batchID", batchID,
							"error", err,
						)
					} else {
						s.processMatchingResult(bgCtx, result)
					}
				}

				if err != nil {
					logger.Warn("Failed to add request to batch, falling back to sequential matching",
						"error", err,
						"rideID", ride.ID,
					)
				}
			}

			if err := s.FindDriverForRide(bgCtx, rideID); err != nil {
				logger.Error("failed to find driver", "error", err, "rideID", rideID)

				currentRide, statusErr := s.repo.FindRideByID(bgCtx, rideID)
				if statusErr != nil {
					logger.Error("failed to fetch ride status", "error", statusErr, "rideID", rideID)
					return
				}

				if currentRide.Status != "searching" {
					logger.Info("ride already accepted by driver, not canceling",
						"rideID", rideID,
						"currentStatus", currentRide.Status,
						"driverID", currentRide.DriverID,
					)
					return
				}

				if err := s.repo.UpdateRideStatus(bgCtx, rideID, "cancelled"); err != nil {
					logger.Error("failed to update ride status", "error", err, "rideID", rideID)
				}

				if holdID != nil {
					if err := s.walletService.ReleaseHold(bgCtx, riderID, walletdto.ReleaseHoldRequest{HoldID: *holdID}); err != nil {
						logger.Error("failed to release hold", "error", err, "rideID", rideID)
					}
				}

				s.wsHelper.SendRideStatusToBoth(bgCtx, riderID, "", rideID, "cancelled", "No drivers are currently active in you area.")
			}
		}()
	}

	if !isScheduled {
		s.wsHelper.SendRideStatusToBoth(ctx, riderID, "", rideID, "searching", "Searching for nearby drivers...")
	} else {
		s.wsHelper.SendRideStatusToBoth(ctx, riderID, "", rideID, "scheduled", "Your ride has been scheduled and will start matching at the scheduled time")
	}

	logger.Info("ride created",
		"rideID", rideID,
		"riderID", riderID,
		"estimatedFare", finalAmount,
		"surge", fareEstimate.SurgeMultiplier,
		"promoDiscount", promoDiscount,
		"isScheduled", isScheduled,
		"status", status,
	)

	freshRide, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		logger.Error("failed to fetch fresh ride data for response", "error", err, "rideID", rideID)
		return dto.ToRideResponse(ride), nil
	}
	return dto.ToRideResponse(freshRide), nil
}

func (s *service) processMatchingResult(ctx context.Context, result *batchingdto.BatchMatchingResult) {
	if result == nil {
		logger.Warn("Null batch matching result received")
		return
	}

	logger.Info("Processing batch matching results",
		"batchID", result.BatchID,
		"assignments", len(result.Assignments),
		"unmatched", len(result.UnmatchedIDs),
	)

	for _, assignment := range result.Assignments {
		logger.Info("Batch assignment match found",
			"rideID", assignment.RideID,
			"driverID", assignment.DriverID,
			"score", assignment.Score,
			"distance", assignment.Distance,
			"eta", assignment.ETA,
		)

		go func(assign batchingdto.OptimalAssignment) {
			bgCtx := context.Background()

			ride, err := s.repo.FindRideByID(bgCtx, assign.RideID)
			if err != nil {
				logger.Error("failed to fetch ride for batch assignment",
					"rideID", assign.RideID,
					"error", err,
				)
				return
			}

			driverIDPtr := assign.DriverID 
			ride.DriverID = &driverIDPtr
			ride.Status = "accepted"
			ride.AcceptedAt = ptr(time.Now())

			if err := s.repo.UpdateRide(bgCtx, ride); err != nil {
				logger.Error("failed to update ride with driver assignment",
					"rideID", assign.RideID,
					"driverID", assign.DriverID,
					"error", err,
				)
				return
			}

			logger.Info("Driver assigned to ride from batch",
				"rideID", assign.RideID,
				"driverID", assign.DriverID,
				"riderID", ride.RiderID,
			)

			s.wsHelper.SendRideAccepted(ride.RiderID, map[string]interface{}{
				"rideId":    assign.RideID,
				"driverId":  assign.DriverID,
				"eta":       assign.ETA,
				"distance":  assign.Distance,
				"timestamp": time.Now().Format(time.RFC3339),
			})

			s.wsHelper.SendRideRequest(assign.DriverID, map[string]interface{}{
				"rideId":     assign.RideID,
				"riderId":    ride.RiderID,
				"pickupLat":  ride.PickupLat,
				"pickupLon":  ride.PickupLon,
				"pickupAddr": ride.PickupAddress,
				"distance":   assign.Distance,
				"timestamp":  time.Now().Format(time.RFC3339),
			})

		}(assignment)
	}

	if len(result.UnmatchedIDs) > 0 {
		logger.Info("Processing unmatched rides from batch",
			"batchID", result.BatchID,
			"unmatchedCount", len(result.UnmatchedIDs),
		)

		for _, rideID := range result.UnmatchedIDs {
			logger.Warn("Unmatched ride, attempting sequential matching",
				"rideID", rideID,
			)

			go func(rid string) {
				bgCtx := context.Background()
				if err := s.FindDriverForRide(bgCtx, rid); err != nil {
					logger.Error("Sequential matching also failed",
						"rideID", rid,
						"error", err,
					)
				}
			}(rideID)
		}
	}
}

func ptr(t time.Time) *time.Time {
	return &t
}

func (s *service) FindDriverForRide(ctx context.Context, rideID string) error {
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return err
	}

	if ride.Status != "searching" {
		return nil
	}

	riderProfile, err := s.ridersRepo.FindByUserID(ctx, ride.RiderID)
	var riderRating float64 = 5.0
	if err == nil && riderProfile != nil {
		riderRating = riderProfile.Rating
		logger.Info("rider rating check",
			"rideID", rideID,
			"riderID", ride.RiderID,
			"riderRating", riderRating,
		)
	}

	radii := []float64{3.0, 5.0, 8.0} 
	if riderRating < 3.5 {
		radii = []float64{8.0}
		logger.Warn("low-rated rider search reduced",
			"rideID", rideID,
			"riderRating", riderRating,
			"searchRadius", "8km only",
		)
	} else if riderRating < 4.0 {
		radii = []float64{5.0, 8.0} 
		logger.Info("medium-rated rider search radius adjusted",
			"rideID", rideID,
			"riderRating", riderRating,
			"searchRadius", "5-8km",
		)
	}

	var nearbyDrivers *trackingdto.NearbyDriversResponse

	for _, radius := range radii {
		nearbyReq := trackingdto.FindNearbyDriversRequest{
			Latitude:      ride.PickupLat,
			Longitude:     ride.PickupLon,
			RadiusKm:      radius,
			VehicleTypeID: ride.VehicleTypeID,
			Limit:         15,
			OnlyAvailable: true,
		}

		nearbyDrivers, err = s.trackingService.FindNearbyDrivers(ctx, nearbyReq)
		if err != nil {
			logger.Warn("driver search failed at radius",
				"radius", radius,
				"error", err,
				"rideID", rideID,
				"lat", ride.PickupLat,
				"lon", ride.PickupLon,
			)
			continue
		}

		if nearbyDrivers != nil && nearbyDrivers.Count > 0 {
			logger.Info("found nearby drivers at radius",
				"radius", radius,
				"count", nearbyDrivers.Count,
				"rideID", rideID,
				"riderRating", riderRating,
			)
			break
		} else {
			logger.Info("no drivers found at radius",
				"radius", radius,
				"rideID", rideID,
				"vehicleTypeID", ride.VehicleTypeID,
				"pickupLat", ride.PickupLat,
				"pickupLon", ride.PickupLon,
			)
		}
		time.Sleep(1 * time.Second)
	}

	if err != nil || nearbyDrivers == nil || nearbyDrivers.Count == 0 {
		logger.Error("no drivers available after checking all radii",
			"rideID", rideID,
			"radii", radii,
			"riderRating", riderRating,
			"vehicleTypeID", ride.VehicleTypeID,
		)
		return errors.New("no drivers available in the area")
	}

	logger.Info("found nearby drivers",
		"rideID", rideID,
		"driverCount", nearbyDrivers.Count,
		"riderRating", riderRating,
	)

	maxConcurrentRequests := 3
	timeout := 30 * time.Second

	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	driversToContact := min(maxConcurrentRequests, nearbyDrivers.Count)
	for i := 0; i < driversToContact; i++ {
		driver := nearbyDrivers.Drivers[i]
		go s.sendRideRequestToDriver(ctxWithTimeout, ride, driver, resultChan, errorChan)
	}

	select {
	case acceptedDriverID := <-resultChan:
		cancel()
		driver, err := s.driversRepo.FindDriverByID(ctx, acceptedDriverID)
		if err != nil {
			logger.Error("failed to fetch driver details for ride assignment",
				"error", err,
				"driverProfileID", acceptedDriverID,
			)
			return err
		}
		logger.Info("driver accepted ride",
			"rideID", rideID,
			"driverID", acceptedDriverID,
			"driverName", driver.User.Name,
		)
		return s.assignDriverToRide(ctx, rideID, driver.UserID, acceptedDriverID)

	case <-ctxWithTimeout.Done():
		logger.Error("no driver accepted ride request (timeout)",
			"rideID", rideID,
			"timeoutSeconds", timeout.Seconds(),
			"driversContacted", driversToContact,
		)
		return errors.New("no driver accepted the ride request")

	case err := <-errorChan:
		logger.Error("error during driver search",
			"rideID", rideID,
			"error", err,
		)
		return err
	}
}

func (s *service) sendRideRequestToDriver(
	ctx context.Context,
	ride *models.Ride,
	driver trackingdto.DriverLocationResponse,
	resultChan chan<- string,
	errorChan chan<- error,
) {
	requestID := uuid.New().String()
	expiresAt := time.Now().Add(10 * time.Second)

	driverDetails, err := s.driversRepo.FindDriverByID(ctx, driver.DriverID)
	if err != nil {
		logger.Error("failed to fetch driver details",
			"error", err,
			"driverID", driver.DriverID,
		)
		errorChan <- err
		return
	}

	userIDForWebSocket := driverDetails.UserID

	rideRequest := &models.RideRequest{
		ID:        requestID,
		RideID:    ride.ID,
		DriverID:  driver.DriverID,
		Status:    "pending",
		SentAt:    time.Now(),
		ExpiresAt: expiresAt,
	}

	if err := s.repo.CreateRideRequest(ctx, rideRequest); err != nil {
		logger.Error("failed to create ride request",
			"error", err,
			"driverID", driver.DriverID,
			"driverName", driverDetails.User.Name,
		)
		errorChan <- err
		return
	}

	rideDetails := map[string]interface{}{
		"rideId":    ride.ID,
		"requestId": requestID,
		"pickup": map[string]interface{}{
			"lat":     ride.PickupLat,
			"lon":     ride.PickupLon,
			"address": ride.PickupAddress,
		},
		"dropoff": map[string]interface{}{
			"lat":     ride.DropoffLat,
			"lon":     ride.DropoffLon,
			"address": ride.DropoffAddress,
		},
		"estimatedFare": ride.EstimatedFare,
		"distance":      driver.Distance,
		"eta":           driver.ETA,
		"expiresIn":     10,
		"riderNotes":    ride.RiderNotes,
	}

	if err := s.wsHelper.SendRideRequest(userIDForWebSocket, rideDetails); err != nil {
		logger.Error("failed to send ride request via WebSocket",
			"error", err,
			"userID", userIDForWebSocket,
			"driverID", driver.DriverID,
			"driverName", driverDetails.User.Name,
		)
	}

	logger.Info("ride request sent to driver",
		"rideID", ride.ID,
		"userID", userIDForWebSocket,
		"driverID", driver.DriverID,
		"driverName", driverDetails.User.Name,
		"requestID", requestID,
		"distance", driver.Distance,
		"eta", driver.ETA,
	)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.repo.UpdateRideRequestStatus(context.Background(), requestID, "cancelled_by_system", nil)
			logger.Info("ride request cancelled due to context done",
				"requestID", requestID,
				"driverID", driver.DriverID,
				"driverName", driverDetails.User.Name,
			)
			return

		case <-ticker.C:
			currentRide, err := s.repo.FindRideByID(ctx, ride.ID)
			if err == nil && currentRide.Status != "searching" {
				logger.Info("ride already accepted by another driver, stopping polling",
					"requestID", requestID,
					"driverID", driver.DriverID,
					"driverName", driverDetails.User.Name,
					"rideID", ride.ID,
					"rideStatus", currentRide.Status,
				)
				return
			}

			updatedRequest, err := s.repo.FindRideRequestByID(ctx, requestID)
			if err != nil {
				continue
			}

			if updatedRequest.Status == "accepted" {
				logger.Info("ride request accepted by driver",
					"requestID", requestID,
					"driverID", driver.DriverID,
					"driverName", driverDetails.User.Name,
					"rideID", ride.ID,
				)
				resultChan <- driver.DriverID
				return
			}

			if updatedRequest.Status == "rejected" {
				logger.Info("ride request rejected by driver",
					"requestID", requestID,
					"driverID", driver.DriverID,
					"driverName", driverDetails.User.Name,
					"rideID", ride.ID,
				)
				return
			}

			if updatedRequest.Status == "expired" {
				logger.Info("ride request expired",
					"requestID", requestID,
					"driverID", driver.DriverID,
					"driverName", driverDetails.User.Name,
					"rideID", ride.ID,
				)
				return
			}

			if time.Now().After(expiresAt) {
				s.repo.UpdateRideRequestStatus(ctx, requestID, "expired", nil)
				logger.Info("ride request expired due to timeout",
					"requestID", requestID,
					"driverID", driver.DriverID,
					"driverName", driverDetails.User.Name,
					"rideID", ride.ID,
				)
				return
			}
		}
	}
}

func (s *service) AcceptRide(ctx context.Context, userID, rideID string) (*dto.RideResponse, error) {
	driver, err := s.driversRepo.FindDriverByUserID(ctx, userID)
	if err != nil {
		logger.Error("driver profile not found for user", "error", err, "userID", userID)
		return nil, response.NotFoundError("Driver profile not found")
	}

	driverID := driver.ID

	rideRequest, err := s.repo.FindRideRequestByRideAndDriver(ctx, rideID, driverID)
	if err != nil {
		logger.Error("ride request not found", "error", err, "rideID", rideID, "driverID", driverID)
		return nil, response.NotFoundError("Ride request not found")
	}

	if rideRequest.Status != "pending" {
		return nil, response.BadRequest("Ride request is no longer available")
	}

	if time.Now().After(rideRequest.ExpiresAt) {
		return nil, response.BadRequest("Ride request has expired")
	}

	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return nil, response.NotFoundError("Ride")
	}

	if ride.Status != "searching" {
		return nil, response.BadRequest("Ride is no longer available")
	}

	if driver.Status != "online" {
		return nil, response.BadRequest("Driver must be online to accept rides")
	}

	riderRiskScore, _ := s.fraudService.CheckUserRiskScore(ctx, ride.RiderID)
	if riderRiskScore > 80 {
		logger.Warn("High-risk rider",
			"riderID", ride.RiderID,
			"riskScore", riderRiskScore,
			"driverID", driverID)
	}

	if err := s.repo.UpdateRideRequestStatus(ctx, rideRequest.ID, "accepted", nil); err != nil {
		logger.Error("failed to update ride request status", "error", err, "requestID", rideRequest.ID)
		return nil, response.InternalServerError("Failed to accept ride", err)
	}

	if err := s.assignDriverToRide(ctx, rideID, userID, driverID); err != nil {
		s.repo.UpdateRideRequestStatus(ctx, rideRequest.ID, "pending", nil)
		logger.Error("failed to assign driver to ride", "error", err, "rideID", rideID, "userID", userID, "driverProfileID", driverID)
		return nil, err
	}

	logger.Info("ride successfully accepted by driver",
		"rideID", rideID,
		"driverID", driverID,
		"userID", userID,
		"driverName", driver.User.Name,
		"riderID", ride.RiderID,
	)

	freshRide, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		logger.Error("failed to fetch fresh ride data for response", "error", err, "rideID", rideID)
		if ride != nil {
			return dto.ToRideResponse(ride), nil
		}
		return nil, response.InternalServerError("Failed to fetch ride data", err)
	}

	if freshRide == nil {
		logger.Error("fetched ride is nil", "rideID", rideID)
		return dto.ToRideResponse(ride), nil
	}

	return dto.ToRideResponse(freshRide), nil
}

func (s *service) RejectRide(ctx context.Context, userID, rideID string, req dto.RejectRideRequest) error {
	driver, err := s.driversRepo.FindDriverByUserID(ctx, userID)
	if err != nil {
		logger.Error("driver profile not found for user",
			"error", err,
			"userID", userID,
		)
		return response.NotFoundError("Driver profile not found")
	}

	driverID := driver.ID

	rideRequest, err := s.repo.FindRideRequestByRideAndDriver(ctx, rideID, driverID)
	if err != nil {
		logger.Error("ride request not found",
			"error", err,
			"rideID", rideID,
			"driverID", driverID,
			"userID", userID,
		)
		return response.NotFoundError("Ride request not found")
	}

	if rideRequest.Status != "pending" {
		return response.BadRequest("Ride request is no longer available")
	}

	if err := s.repo.UpdateRideRequestStatus(ctx, rideRequest.ID, "rejected", &req.Reason); err != nil {
		return response.InternalServerError("Failed to reject ride", err)
	}

	logger.Info("ride rejected by driver",
		"rideID", rideID,
		"driverID", driverID,
		"userID", userID,
		"driverName", driver.User.Name,
		"requestID", rideRequest.ID,
		"reason", req.Reason,
	)

	return nil
}

func (s *service) MarkArrived(ctx context.Context, userID, rideID string) (*dto.RideResponse, error) {
	driver, err := s.driversRepo.FindDriverByUserID(ctx, userID)
	if err != nil {
		logger.Error("driver profile not found for user",
			"error", err,
			"userID", userID,
		)
		return nil, response.NotFoundError("Driver profile not found")
	}

	driverID := driver.ID

	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return nil, response.NotFoundError("Ride")
	}

	if ride.DriverID == nil || *ride.DriverID != userID {
		logger.Warn("unauthorized attempt to mark arrived",
			"userID", userID,
			"driverProfileID", driverID,
			"rideDriverUserID", ride.DriverID,
			"rideID", rideID,
		)
		return nil, response.ForbiddenError("Not authorized")
	}

	if ride.Status != "accepted" {
		return nil, response.BadRequest("Invalid ride status")
	}

	if err := s.repo.UpdateRideStatus(ctx, rideID, "arrived"); err != nil {
		return nil, response.InternalServerError("Failed to update status", err)
	}

	if err := websocketutil.SendToUser(ride.RiderID, websocket.TypeRideStatusUpdate, map[string]interface{}{
		"rideId":    rideID,
		"status":    "arrived",
		"message":   "Your driver has arrived at the pickup location",
		"timestamp": time.Now().UTC(),
	}); err != nil {
		logger.Warn("failed to notify rider of arrival", "error", err, "rideID", rideID)
	}

	logger.Info("driver arrived at pickup location",
		"rideID", rideID,
		"driverID", driverID,
		"userID", userID,
		"driverName", driver.User.Name,
		"riderID", ride.RiderID,
	)

	time.AfterFunc(3*time.Minute, func() {
		if ride.WaitTimeCharge == nil {
			ride.WaitTimeCharge = new(float64)
		}
		*ride.WaitTimeCharge += 1.0
	})

	freshRide, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		logger.Error("failed to fetch fresh ride data for response", "error", err, "rideID", rideID)
		if ride != nil {
			return dto.ToRideResponse(ride), nil
		}
		return nil, response.InternalServerError("Failed to fetch ride data", err)
	}

	if freshRide == nil {
		logger.Error("fetched ride is nil", "rideID", rideID)
		return dto.ToRideResponse(ride), nil
	}

	return dto.ToRideResponse(freshRide), nil
}

func (s *service) StartRide(ctx context.Context, userID, rideID string, req dto.StartRideRequest) (*dto.RideResponse, error) {
	logger.Info("StartRide called", "userID", userID, "rideID", rideID)

	driver, err := s.driversRepo.FindDriverByUserID(ctx, userID)
	if err != nil {
		logger.Error("driver profile not found for user", "error", err, "userID", userID)
		return nil, response.NotFoundError("Driver profile not found")
	}

	driverID := driver.ID
	driverUserID := driver.UserID
	logger.Info("driver fetched", "driverID", driverID, "driverUserID", driverUserID)

	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		logger.Error("ride not found", "error", err, "rideID", rideID)
		return nil, response.NotFoundError("Ride")
	}
	logger.Info("ride fetched", "rideID", rideID, "rideStatus", ride.Status)

	if ride.DriverID == nil || *ride.DriverID != userID {
		logger.Warn("unauthorized attempt to start ride", "userID", userID, "driverProfileID", driverID, "rideID", rideID)
		return nil, response.ForbiddenError("Not authorized")
	}

	if ride.Status != "accepted" && ride.Status != "arrived" {
		logger.Warn("invalid ride status for start", "rideID", rideID, "status", ride.Status)
		return nil, response.BadRequest("Ride must be accepted or arrived to start")
	}

	logger.Info("verifying ride PIN", "rideID", rideID, "riderID", ride.RiderID)
	if err := s.ridePINService.VerifyRidePIN(ctx, ride.RiderID, req.RiderPIN); err != nil {
		logger.Warn("invalid ride PIN attempt at start",
			"rideID", rideID,
			"driverID", driverID,
			"riderID", ride.RiderID)
		return nil, response.BadRequest("Invalid Rider PIN. Please ask the rider for their 4-digit Ride PIN.")
	}

	logger.Info("ride PIN verified at start", "rideID", rideID)

	now := time.Now()
	if ride.ArrivedAt != nil {
		waitMinutes := int(now.Sub(*ride.ArrivedAt).Minutes())
		if waitMinutes > 3 {
			waitCharge := float64(waitMinutes-3) * 0.50
			ride.WaitTimeCharge = &waitCharge
			logger.Info("wait time charge applied", "rideID", rideID, "waitMinutes", waitMinutes, "charge", waitCharge)
		}
	}

	ride.Status = "started"
	ride.StartedAt = &now
	logger.Info("updating ride status to started", "rideID", rideID)

	if err := s.repo.UpdateRide(ctx, ride); err != nil {
		logger.Error("failed to update ride", "error", err, "rideID", rideID)
		return nil, response.InternalServerError("Failed to start ride", err)
	}
	logger.Info("ride status updated", "rideID", rideID)

	logger.Info("updating driver status", "driverID", driverID)
	if err := s.driversRepo.UpdateDriverStatus(ctx, driverID, "on_trip"); err != nil {
		logger.Warn("failed to update driver status", "error", err, "driverID", driverID)
	}

	cacheKey := fmt.Sprintf("ride:active:%s", rideID)
	logger.Info("setting ride cache", "cacheKey", cacheKey)
	if err := cache.SetJSON(ctx, cacheKey, ride, 30*time.Minute); err != nil {
		logger.Warn("failed to cache ride", "error", err, "rideID", rideID)
	}

	logger.Info("sending websocket notification", "rideID", rideID)
	if err := websocketutil.SendRideStatusUpdate(ride.RiderID, driverUserID, map[string]interface{}{
		"rideId":    rideID,
		"status":    "started",
		"message":   "Your ride has started",
		"timestamp": time.Now().UTC(),
	}); err != nil {
		logger.Warn("failed to send websocket notification", "error", err, "rideID", rideID)
	}

	logger.Info("ride started successfully", "rideID", rideID, "driverID", driverID, "riderID", ride.RiderID)

	logger.Info("fetching fresh ride data", "rideID", rideID)
	freshRide, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		logger.Error("failed to fetch fresh ride data for response", "error", err, "rideID", rideID)

		if ride != nil {
			logger.Info("returning cached ride response", "rideID", rideID)
			return dto.ToRideResponse(ride), nil
		}
		return nil, response.InternalServerError("Failed to fetch ride data", err)
	}

	if freshRide == nil {
		logger.Error("fetched ride is nil", "rideID", rideID)
		return dto.ToRideResponse(ride), nil
	}

	logger.Info("converting freshRide to response", "rideID", rideID)
	resp := dto.ToRideResponse(freshRide)
	logger.Info("response conversion complete", "rideID", rideID)
	return resp, nil
}

func (s *service) CompleteRide(ctx context.Context, userID, rideID string, req dto.CompleteRideRequest) (*dto.RideResponse, error) {
	driver, err := s.driversRepo.FindDriverByUserID(ctx, userID)
	if err != nil {
		logger.Error("driver profile not found for user", "error", err, "userID", userID)
		return nil, response.NotFoundError("Driver profile not found")
	}

	driverID := driver.ID
	driverUserID := driver.UserID

	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return nil, response.NotFoundError("Ride")
	}

	if ride.DriverID == nil || *ride.DriverID != userID {
		logger.Warn("unauthorized attempt to complete ride", "userID", userID, "driverProfileID", driverID, "rideID", rideID)
		return nil, response.ForbiddenError("Not authorized")
	}

	if ride.Status != "started" {
		return nil, response.BadRequest("Ride must be started first")
	}

	distanceToDropoff := location.HaversineDistance(req.DriverLat, req.DriverLon, ride.DropoffLat, ride.DropoffLon)
	distanceKm := distanceToDropoff
	const maxCompletionRadiusKm = 0.100 
	const epsilon = 0.001          

	if distanceKm > maxCompletionRadiusKm+epsilon {
		logger.Warn("driver outside completion radius", "rideID", rideID, "distanceKm", distanceKm, "maxRadiusKm", maxCompletionRadiusKm)
		return nil, response.BadRequest(fmt.Sprintf("You must be within 100 meters of the destination. Current distance: %.0f meters", distanceKm*1000))
	}

	logger.Info("driver verified within 100 meter radius", "rideID", rideID, "distanceKm", distanceKm)

	actualFareReq := pricingdto.CalculateActualFareRequest{
		ActualDistanceKm:  req.ActualDistance,
		ActualDurationSec: req.ActualDuration,
		VehicleTypeID:     ride.VehicleTypeID,
		SurgeMultiplier:   ride.SurgeMultiplier,
	}

	actualFareResp, err := s.pricingService.CalculateActualFare(ctx, actualFareReq)
	if err != nil {
		return nil, err
	}

	driverFare := actualFareResp.TotalFare
	riderFare := actualFareResp.TotalFare

	if ride.WaitTimeCharge != nil {
		driverFare += *ride.WaitTimeCharge
		riderFare += *ride.WaitTimeCharge
	}

	if ride.DestinationChangeCharge != nil {
		driverFare += *ride.DestinationChangeCharge
		riderFare += *ride.DestinationChangeCharge
	}

	if ride.PromoDiscount != nil && *ride.PromoDiscount > 0 {
		driverFare -= *ride.PromoDiscount
		riderFare -= *ride.PromoDiscount
		logger.Info("promo code applied - both driver and rider get discount",
			"rideID", rideID,
			"originalFare", actualFareResp.TotalFare,
			"driverFare", driverFare,
			"riderFare", riderFare,
			"promoDiscount", *ride.PromoDiscount,
		)
	}

	actualFare := driverFare

	if s.sosService != nil {
		activeSOS, _ := s.sosService.GetActiveSOS(ctx, ride.RiderID)
		if activeSOS != nil {
			resolveReq := sosdto.ResolveSOSRequest{
				Notes: "Ride completed successfully",
			}
			s.sosService.ResolveSOS(ctx, ride.RiderID, activeSOS.ID, resolveReq)
			logger.Info("Auto-resolved SOS alert on ride completion", "sosID", activeSOS.ID)
		}
	}

	ride.ActualDistance = &req.ActualDistance
	ride.ActualDuration = &req.ActualDuration
	ride.ActualFare = &actualFare
	ride.DriverFare = &driverFare
	ride.RiderFare = &riderFare  
	ride.Status = "completed"
	completedAt := time.Now()
	ride.CompletedAt = &completedAt

	if err := s.repo.UpdateRide(ctx, ride); err != nil {
		return nil, response.InternalServerError("Failed to complete ride", err)
	}

	if ride.WalletHoldID != nil {
		captureReq := walletdto.CaptureHoldRequest{
			HoldID:      *ride.WalletHoldID,
			Amount:      ride.ActualFare,
			Description: fmt.Sprintf("Cash payment for ride %s", rideID),
		}

		_, err := s.walletService.CaptureHold(ctx, ride.RiderID, captureReq)
		if err != nil {
			logger.Error("failed to capture hold", "error", err, "rideID", rideID)
			return nil, response.InternalServerError("Failed to process payment", err)
		}
	}

	if ride.PromoCodeID != nil && ride.PromoCode != nil {
		applyReq := promotionsdto.ApplyPromoCodeRequest{
			Code:       *ride.PromoCode,
			RideID:     rideID,
			RideAmount: actualFare,
		}

		_, err := s.promotionsService.ApplyPromoCode(ctx, ride.RiderID, applyReq)
		if err != nil {
			logger.Error("Failed to apply promo code", "error", err, "rideID", rideID)
		}
	}

	driverEarnings := actualFareResp.DriverPayout

	logger.Info("ride payout breakdown",
		"rideID", rideID,
		"totalFare", actualFare,
		"driverEarnings", driverEarnings,
		"platformCommission", 0,
		"commissionRate", 0,
	)

	_, err = s.walletService.CreditDriverWallet(
		ctx,
		driver.UserID,
		actualFare,
		"ride_earnings",
		rideID,
		fmt.Sprintf("Cash earned from ride %s", rideID),
		map[string]interface{}{"total_fare": actualFare},
	)
	if err != nil {
		logger.Error("failed to credit driver wallet", "error", err, "rideID", rideID)
	}

	s.driversRepo.IncrementTrips(ctx, driverID)
	s.driversRepo.UpdateEarnings(ctx, driverID, driverEarnings)
	s.ridersRepo.IncrementTotalRides(ctx, ride.RiderID)

	if err := s.driversRepo.UpdateDriverStatus(ctx, driverID, "online"); err != nil {
		logger.Warn("failed to update driver status", "error", err, "driverID", driverID)
	}

	busyKey := fmt.Sprintf("driver:busy:%s", driverID)
	if err := cache.Delete(ctx, busyKey); err != nil {
		logger.Warn("failed to clear driver busy cache", "error", err, "busyKey", busyKey)
	}

	activeRideCacheKey := fmt.Sprintf("driver:active:ride:%s", driverID)
	if err := cache.Delete(ctx, activeRideCacheKey); err != nil {
		logger.Warn("failed to clear active ride cache", "error", err, "activeRideCacheKey", activeRideCacheKey)
	}

	rideCacheKey := fmt.Sprintf("ride:active:%s", rideID)
	if err := cache.Delete(ctx, rideCacheKey); err != nil {
		logger.Warn("failed to clear ride cache", "error", err, "rideCacheKey", rideCacheKey)
	}

	if err := websocketutil.SendToUser(ride.RiderID, websocket.TypeRideCompleted, map[string]interface{}{
		"rideId":     rideID,
		"actualFare": actualFare,
		"message":    "Your ride is complete. Thank you for riding with us!",
		"timestamp":  time.Now().UTC(),
	}); err != nil {
		logger.Warn("failed to notify rider of completion", "error", err, "rideID", rideID)
	}

	if err := websocketutil.SendToUser(driverUserID, websocket.TypeRideCompleted, map[string]interface{}{
		"rideId":    rideID,
		"earnings":  driverEarnings,
		"message":   "Ride completed successfully",
		"timestamp": time.Now().UTC(),
	}); err != nil {
		logger.Warn("failed to notify driver of completion", "error", err, "rideID", rideID)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic in promptRatings goroutine",
					"error", r,
					"rideID", rideID,
					"riderID", ride.RiderID,
				)
			}
		}()
		s.promptRatings(context.Background(), ride.RiderID, driverUserID, rideID)
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic in fraud detection goroutine",
					"error", r,
					"rideID", rideID,
				)
			}
		}()
		s.fraudService.DetectFraudPatterns(context.Background(), rideID)
	}()

	logger.Info("ride completed successfully",
		"rideID", rideID,
		"driverID", driverID,
		"actualFare", actualFare,
		"driverEarnings", driverEarnings,
	)

	freshRide, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		logger.Error("failed to fetch fresh ride data for response", "error", err, "rideID", rideID)
		if ride != nil {
			return dto.ToRideResponse(ride), nil
		}
		return nil, response.InternalServerError("Failed to fetch ride data", err)
	}

	if freshRide == nil {
		logger.Error("fetched ride is nil", "rideID", rideID)
		return dto.ToRideResponse(ride), nil
	}

	return dto.ToRideResponse(freshRide), nil
}

func (s *service) promptRatings(ctx context.Context, riderID, driverUserID, rideID string) {
	time.Sleep(5 * time.Second)

	if err := websocketutil.SendToUser(riderID, websocket.TypeRatingPrompt, map[string]interface{}{
		"rideId":  rideID,
		"message": "How was your ride? Please rate your driver",
	}); err != nil {
		logger.Warn("failed to send rating prompt to rider", "error", err, "rideID", rideID, "riderID", riderID)
	}

	if err := websocketutil.SendToUser(driverUserID, websocket.TypeRatingPrompt, map[string]interface{}{
		"rideId":  rideID,
		"message": "Please rate your rider",
	}); err != nil {
		logger.Warn("failed to send rating prompt to driver", "error", err, "rideID", rideID, "driverUserID", driverUserID)
	}
}

func (s *service) assignDriverToRide(ctx context.Context, rideID, userID, driverProfileID string) error {
	fmt.Println("Run func: assignDriverToRide")

	err := s.repo.UpdateRideStatusAndDriver(ctx, rideID, "accepted", "searching", userID)
	if err != nil {
		logger.Warn("failed to assign driver - ride may be already accepted",
			"error", err,
			"rideID", rideID,
			"userID", userID,
		)
		return response.BadRequest("Ride already accepted by another driver")
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic in cancel pending requests goroutine",
					"error", r,
					"rideID", rideID,
					"driverProfileID", driverProfileID,
				)
			}
		}()
		s.repo.CancelPendingRequestsExcept(context.Background(), rideID, driverProfileID)
	}()

	s.driversRepo.UpdateDriverStatus(ctx, driverProfileID, "busy")

	busyKey := fmt.Sprintf("driver:busy:%s", driverProfileID)
	cache.Set(ctx, busyKey, "true", 30*time.Minute)

	ride, _ := s.repo.FindRideByID(ctx, rideID)

	cacheKey := fmt.Sprintf("ride:active:%s", rideID)
	cache.SetJSON(ctx, cacheKey, ride, 30*time.Minute)

	driver, err := s.driversRepo.FindDriverByID(ctx, driverProfileID)
	if err != nil {
		logger.Error("failed to fetch driver details for notification",
			"error", err,
			"driverProfileID", driverProfileID,
		)
		return nil
	}

	if driver.Vehicle == nil {
		logger.Warn("driver has no vehicle assigned",
			"driverProfileID", driverProfileID,
			"driverName", driver.User.Name,
		)
		return nil
	}

	activeRideCacheKey := fmt.Sprintf("driver:active:ride:%s", driverProfileID)
	cache.SetJSON(ctx, activeRideCacheKey, map[string]string{
		"rideID":  rideID,
		"riderID": ride.RiderID,
	}, 30*time.Minute)

	var driverLat, driverLon float64
	var calculatedETA int

	driverLocation, err := s.trackingService.GetDriverLocation(ctx, driverProfileID)
	if err == nil && driverLocation != nil {
		driverLat = driverLocation.Latitude
		driverLon = driverLocation.Longitude

		distance := calculateDistance(driverLat, driverLon, ride.PickupLat, ride.PickupLon)
		calculatedETA = int((distance / 30.0) * 60)
		if calculatedETA < 1 {
			calculatedETA = 1
		}
	} else {
		driverLat = ride.PickupLat
		driverLon = ride.PickupLon
		calculatedETA = 10
		logger.Warn("could not get driver location, using defaults",
			"error", err,
			"driverProfileID", driverProfileID,
		)
	}

	rideDetails := map[string]interface{}{
		"rideId":   rideID,
		"driverId": driver.ID,
		"userId":   driver.UserID, 
		"driver": map[string]interface{}{
			"id":     driver.ID,
			"userId": driver.UserID, 
			"name":   driver.User.Name,
			"phone":  driver.User.Phone,
			"rating": driver.Rating,
		},
		"vehicle": map[string]interface{}{
			"type":  driver.Vehicle.VehicleType.Name,
			"model": driver.Vehicle.Model,
			"color": driver.Vehicle.Color,
			"plate": driver.Vehicle.LicensePlate,
		},
		"location": map[string]interface{}{ 
			"latitude":  driverLat,
			"longitude": driverLon,
		},
		"message": "Driver is on the way!",
		"eta":     calculatedETA,
	}

	// Send to rider via WebSocket
	if err := s.wsHelper.SendRideAccepted(ride.RiderID, rideDetails); err != nil {
		logger.Error("failed to send ride acceptance notification",
			"error", err,
			"riderID", ride.RiderID,
			"driverProfileID", driverProfileID,
			"driverName", driver.User.Name,
		)
	}

	logger.Info("ride assigned to driver successfully",
		"rideID", rideID,
		"driverProfileID", driverProfileID,
		"driverName", driver.User.Name,
		"driverPhone", driver.User.Phone,
		"vehicleType", driver.Vehicle.VehicleType.Name,
		"vehiclePlate", driver.Vehicle.LicensePlate,
		"riderID", ride.RiderID,
		"calculatedETA", calculatedETA,
	)

	return nil
}

func (s *service) ProcessRideRequestTimeout(ctx context.Context, requestID string) error {
	request, err := s.repo.FindRideRequestByID(ctx, requestID)
	if err != nil {
		return err
	}

	if request.Status == "pending" && time.Now().After(request.ExpiresAt) {
		return s.repo.UpdateRideRequestStatus(ctx, requestID, "expired", nil)
	}

	return nil
}

func (s *service) GetAvailableCars(ctx context.Context, riderID string, req dto.AvailableCarRequest) (*dto.AvailableCarsListResponse, error) {
	drivers, err := s.driversRepo.FindNearbyDrivers(ctx, req.Latitude, req.Longitude, req.RadiusKm, "")
	if err != nil {
		logger.Error("failed to find nearby drivers", "error", err, "riderID", riderID)
		return nil, response.InternalServerError("Failed to fetch available cars", err)
	}

	if len(drivers) == 0 {
		return &dto.AvailableCarsListResponse{
			TotalCount: 0,
			CarsCount:  0,
			RiderLat:   req.Latitude,
			RiderLon:   req.Longitude,
			RadiusKm:   req.RadiusKm,
			Cars:       []*dto.AvailableCarResponse{},
			Timestamp:  time.Now(),
		}, nil
	}

	riderProf, err := s.ridersRepo.FindByUserID(ctx, riderID)
	if err != nil {
		logger.Warn("failed to fetch rider profile", "error", err, "riderID", riderID)
	}

	var availableCars []*dto.AvailableCarResponse
	var filteredStatus, filteredUnverified, filteredRating, filteredNoVehicle, filteredLocationParse, filteredNoUser int

	for _, driver := range drivers {
		if driver.Status != "online" {
			filteredStatus++
			logger.Warn("driver status not online (unexpected - db should have filtered)", "driverID", driver.ID, "status", driver.Status)
			continue
		}
		if !driver.IsVerified {
			filteredUnverified++
			logger.Warn("driver not verified (unexpected - db should have filtered)", "driverID", driver.ID)
			continue
		}

		if driver.User.ID == "" {
			filteredNoUser++
			logger.Warn("driver user data not loaded", "driverID", driver.ID)
			continue
		}

		if riderProf != nil && riderProf.Rating >= 4.5 && driver.Rating < 2.5 {
			filteredRating++
			logger.Info("driver filtered by rating", "driverID", driver.ID, "driverRating", driver.Rating, "riderRating", riderProf.Rating)
			continue
		}

		vehicle := driver.Vehicle
		if vehicle == nil {
			filteredNoVehicle++
			logger.Warn("driver has no vehicle", "driverID", driver.ID)
			continue
		}

		if vehicle.VehicleType.ID == "" {
			filteredNoVehicle++
			logger.Warn("vehicle type not loaded for driver", "driverID", driver.ID, "vehicleID", vehicle.ID)
			continue
		}

		if driver.CurrentLocation == nil {
			filteredLocationParse++
			logger.Warn("driver current location is nil", "driverID", driver.ID)
			continue
		}

		driverLat, driverLon, err := parseDriverLocation(driver.CurrentLocation)
		if err != nil {
			filteredLocationParse++
			logger.Warn("failed to parse driver location", "error", err, "driverID", driver.ID, "location", driver.CurrentLocation)
			continue
		}

		distanceKm := location.HaversineDistance(req.Latitude, req.Longitude, driverLat, driverLon)

		const avgSpeedKmh = 40.0
		etaSeconds := location.CalculateETA(distanceKm, avgSpeedKmh)
		etaMinutes := (etaSeconds + 30) / 60

		estimatedFare := 0.0
		if vehicle.VehicleType.BaseFare > 0 {
			estimatedFare = vehicle.VehicleType.BaseFare + (5.0 * vehicle.VehicleType.PerKmRate)
		}

		driverName := driver.User.Name
		driverImage := driver.User.ProfilePhotoURL

		carResp := &dto.AvailableCarResponse{
			ID:                 vehicle.ID,
			DriverID:           driver.ID,
			DriverName:         driverName,
			DriverRating:       driver.Rating,
			DriverImage:        driverImage,
			VehicleTypeID:      vehicle.VehicleTypeID,
			VehicleType:        vehicle.VehicleType.Name,
			VehicleDisplayName: vehicle.VehicleType.DisplayName,
			Make:               vehicle.Make,
			Model:              vehicle.Model,
			Color:              vehicle.Color,
			LicensePlate:       vehicle.LicensePlate,
			Capacity:           vehicle.Capacity,
			CurrentLatitude:    driverLat,
			CurrentLongitude:   driverLon,
			Heading:            driver.Heading,
			DistanceKm:         math.Round(distanceKm*10) / 10, 
			ETASeconds:         etaSeconds,
			ETAMinutes:         etaMinutes,
			EstimatedFare:      math.Round(estimatedFare*100) / 100, 
			SurgeMultiplier:    1.0,                                 
			AcceptanceRate:     driver.AcceptanceRate,
			CancellationRate:   driver.CancellationRate,
			TotalTrips:         driver.TotalTrips,
			Status:             driver.Status,
			IsVerified:         driver.IsVerified,
			UpdatedAt:          driver.UpdatedAt,
		}

		availableCars = append(availableCars, carResp)
	}

	sort.Slice(availableCars, func(i, j int) bool {
		return availableCars[i].DistanceKm < availableCars[j].DistanceKm
	})

	if len(availableCars) > 20 {
		availableCars = availableCars[:20]
	}

	if len(availableCars) == 0 {
		logger.Info("no available cars after filtering",
			"totalNearbyDrivers", len(drivers),
			"filteredStatus", filteredStatus,
			"filteredUnverified", filteredUnverified,
			"filteredNoUser", filteredNoUser,
			"filteredRating", filteredRating,
			"filteredNoVehicle", filteredNoVehicle,
			"filteredLocationParse", filteredLocationParse,
		)

		availableCars = make([]*dto.AvailableCarResponse, 0)
	}

	return &dto.AvailableCarsListResponse{
		TotalCount: len(drivers),
		CarsCount:  len(availableCars),
		RiderLat:   req.Latitude,
		RiderLon:   req.Longitude,
		RadiusKm:   req.RadiusKm,
		Cars:       availableCars,
		Timestamp:  time.Now(),
	}, nil
}

func (s *service) GetVehiclesWithDetails(ctx context.Context, riderID string, req dto.VehicleDetailsRequest) (*dto.VehiclesWithDetailsListResponse, error) {

	tripDistance := location.HaversineDistance(req.PickupLat, req.PickupLon, req.DropoffLat, req.DropoffLon)
	if tripDistance < 0.5 {
		return nil, response.BadRequest("Minimum trip distance is 0.5 km")
	}
	if tripDistance > 100 {
		return nil, response.BadRequest("Maximum trip distance is 100 km")
	}

	const avgSpeedKmh = 40.0
	tripDurationSeconds := location.CalculateETA(tripDistance, avgSpeedKmh)

	drivers, err := s.driversRepo.FindNearbyDrivers(ctx, req.PickupLat, req.PickupLon, req.RadiusKm, "")
	if err != nil {
		logger.Error("failed to find nearby drivers", "error", err, "riderID", riderID)
		return nil, response.InternalServerError("Failed to fetch vehicles", err)
	}

	if len(drivers) == 0 {
		return &dto.VehiclesWithDetailsListResponse{
			PickupLat:        req.PickupLat,
			PickupLon:        req.PickupLon,
			PickupAddress:    req.PickupAddress,
			DropoffLat:       req.DropoffLat,
			DropoffLon:       req.DropoffLon,
			DropoffAddress:   req.DropoffAddress,
			RadiusKm:         req.RadiusKm,
			TripDistance:     math.Round(tripDistance*10) / 10,
			TripDuration:     tripDurationSeconds,
			TripDurationMins: (tripDurationSeconds + 30) / 60,
			TotalCount:       0,
			CarsCount:        0,
			Vehicles:         make([]*dto.VehicleWithDetailsResponse, 0),
			Timestamp:        time.Now(),
		}, nil
	}

	riderProf, err := s.ridersRepo.FindByUserID(ctx, riderID)
	if err != nil {
		logger.Warn("failed to fetch rider profile", "error", err, "riderID", riderID)
	}

	demandGeohash := fmt.Sprintf("%.1f_%.1f", req.PickupLat, req.PickupLon)
	demandData, err := s.pricingService.GetCurrentDemand(ctx, demandGeohash)
	if err != nil {
		logger.Warn("failed to get demand data", "error", err, "geohash", demandGeohash)
		demandData = &pricingdto.DemandTrackingResponse{
			PendingRequests:  0,
			AvailableDrivers: len(drivers),
		}
	}

	var vehicles []*dto.VehicleWithDetailsResponse
	var filteredStatus, filteredUnverified, filteredNoUser, filteredRating, filteredNoVehicle int

	for _, driver := range drivers {
		if driver.Status != "online" {
			filteredStatus++
			logger.Warn("driver status not online (unexpected - db should have filtered)", "driverID", driver.ID, "status", driver.Status)
			continue
		}
		if !driver.IsVerified {
			filteredUnverified++
			logger.Warn("driver not verified (unexpected - db should have filtered)", "driverID", driver.ID)
			continue
		}

		if driver.User.ID == "" {
			filteredNoUser++
			logger.Warn("driver user data not loaded", "driverID", driver.ID)
			continue
		}

		if riderProf != nil && riderProf.Rating >= 4.5 && driver.Rating < 2.5 {
			filteredRating++
			logger.Info("driver filtered by rating", "driverID", driver.ID, "driverRating", driver.Rating, "riderRating", riderProf.Rating)
			continue
		}

		vehicle := driver.Vehicle
		if vehicle == nil {
			filteredNoVehicle++
			logger.Warn("driver has no vehicle", "driverID", driver.ID)
			continue
		}

		if vehicle.VehicleType.ID == "" {
			filteredNoVehicle++
			logger.Warn("vehicle type not loaded for driver", "driverID", driver.ID, "vehicleID", vehicle.ID)
			continue
		}

		if driver.CurrentLocation == nil {
			logger.Warn("driver current location is nil", "driverID", driver.ID)
			continue
		}

		driverLat, driverLon, err := parseDriverLocation(driver.CurrentLocation)
		if err != nil {
			logger.Warn("failed to parse driver location", "error", err, "driverID", driver.ID)
			continue
		}

		distanceToPickup := location.HaversineDistance(req.PickupLat, req.PickupLon, driverLat, driverLon)
		etaSecondsToPickup := location.CalculateETA(distanceToPickup, avgSpeedKmh)
		etaMinutesToPickup := (etaSecondsToPickup + 30) / 60

		geohash := fmt.Sprintf("%.1f_%.1f", req.PickupLat, req.PickupLon)
		surgeResp, err := s.pricingService.CalculateCombinedSurge(ctx, vehicle.VehicleTypeID, geohash, req.PickupLat, req.PickupLon)
		if err != nil {
			logger.Warn("failed to get surge info", "error", err, "vehicleTypeID", vehicle.VehicleTypeID)
			surgeResp = &pricingdto.SurgeCalculationResponse{
				AppliedMultiplier: 1.0,
				Reason:            "normal",
			}
		}

		fareReq := pricingdto.FareEstimateRequest{
			PickupLat:     req.PickupLat,
			PickupLon:     req.PickupLon,
			DropoffLat:    req.DropoffLat,
			DropoffLon:    req.DropoffLon,
			VehicleTypeID: vehicle.VehicleTypeID,
		}
		fareResp, err := s.pricingService.GetFareEstimate(ctx, fareReq)
		if err != nil {
			logger.Warn("failed to get fare estimate", "error", err, "vehicleTypeID", vehicle.VehicleTypeID)
			estimatedFare := vehicle.VehicleType.BaseFare + (tripDistance*vehicle.VehicleType.PerKmRate)*surgeResp.AppliedMultiplier
			fareResp = &pricingdto.FareEstimateResponse{
				BaseFare:        vehicle.VehicleType.BaseFare,
				DistanceFare:    tripDistance * vehicle.VehicleType.PerKmRate,
				DurationFare:    0,
				SurgeMultiplier: surgeResp.AppliedMultiplier,
				TotalFare:       estimatedFare,
			}
		}

		demandLevel := "normal"
		if surgeResp.AppliedMultiplier > 2.0 {
			demandLevel = "extreme"
		} else if surgeResp.AppliedMultiplier > 1.5 {
			demandLevel = "high"
		} else if surgeResp.AppliedMultiplier > 1.0 {
			demandLevel = "high"
		} else if surgeResp.AppliedMultiplier < 1.0 {
			demandLevel = "low"
		}

		etaFormatted := fmt.Sprintf("%d min", etaMinutesToPickup)

		vehicleResp := &dto.VehicleWithDetailsResponse{
			ID:               vehicle.ID,
			DriverID:         driver.ID,
			DriverName:       driver.User.Name,
			DriverRating:     driver.Rating,
			DriverImage:      driver.User.ProfilePhotoURL,
			AcceptanceRate:   driver.AcceptanceRate,
			CancellationRate: driver.CancellationRate,
			TotalTrips:       driver.TotalTrips,
			IsVerified:       driver.IsVerified,
			Status:           driver.Status,

			VehicleTypeID:      vehicle.VehicleTypeID,
			VehicleType:        vehicle.VehicleType.Name,
			VehicleDisplayName: vehicle.VehicleType.DisplayName,
			Make:               vehicle.Make,
			Model:              vehicle.Model,
			Year:               vehicle.Year,
			Color:              vehicle.Color,
			LicensePlate:       vehicle.LicensePlate,
			Capacity:           vehicle.Capacity,

			CurrentLatitude:  driverLat,
			CurrentLongitude: driverLon,
			Heading:          driver.Heading,
			DistanceKm:       math.Round(distanceToPickup*10) / 10,

			ETASeconds:   etaSecondsToPickup,
			ETAMinutes:   etaMinutesToPickup,
			ETAFormatted: etaFormatted,

			BaseFare:              vehicle.VehicleType.BaseFare,
			PerKmRate:             vehicle.VehicleType.PerKmRate,
			PerMinRate:            vehicle.VehicleType.PerMinuteRate,
			EstimatedFare:         math.Round(fareResp.TotalFare*100) / 100,
			EstimatedDistance:     math.Round(tripDistance*10) / 10,
			EstimatedDuration:     tripDurationSeconds,
			EstimatedDurationMins: (tripDurationSeconds + 30) / 60,

			SurgeMultiplier: surgeResp.AppliedMultiplier,
			SurgeReason:     surgeResp.Reason,

			PendingRequests:  demandData.PendingRequests,
			AvailableDrivers: demandData.AvailableDrivers,
			Demand:           demandLevel,

			UpdatedAt: driver.UpdatedAt,
			Timestamp: time.Now(),
		}

		vehicles = append(vehicles, vehicleResp)
		logger.Debug("vehicle added to list", "driverID", driver.ID, "vehicleID", vehicle.ID, "distance", vehicleResp.DistanceKm, "fare", vehicleResp.EstimatedFare)
	}

	sort.Slice(vehicles, func(i, j int) bool {
		return vehicles[i].DistanceKm < vehicles[j].DistanceKm
	})

	if len(vehicles) > 20 {
		vehicles = vehicles[:20]
	}

	if len(vehicles) == 0 {
		logger.Info("no available vehicles after filtering",
			"totalNearbyDrivers", len(drivers),
			"filteredStatus", filteredStatus,
			"filteredUnverified", filteredUnverified,
			"filteredNoUser", filteredNoUser,
			"filteredRating", filteredRating,
			"filteredNoVehicle", filteredNoVehicle,
		)
		vehicles = make([]*dto.VehicleWithDetailsResponse, 0)
	}

	return &dto.VehiclesWithDetailsListResponse{
		PickupLat:        req.PickupLat,
		PickupLon:        req.PickupLon,
		PickupAddress:    req.PickupAddress,
		DropoffLat:       req.DropoffLat,
		DropoffLon:       req.DropoffLon,
		DropoffAddress:   req.DropoffAddress,
		RadiusKm:         req.RadiusKm,
		TripDistance:     math.Round(tripDistance*10) / 10,
		TripDuration:     tripDurationSeconds,
		TripDurationMins: (tripDurationSeconds + 30) / 60,
		TotalCount:       len(drivers),
		CarsCount:        len(vehicles),
		Vehicles:         vehicles,
		Timestamp:        time.Now(),
	}, nil
}

func parseDriverLocation(geom *string) (float64, float64, error) {
	if geom == nil {
		return 0, 0, fmt.Errorf("driver location is nil")
	}

	var lat, lon float64
	_, err := fmt.Sscanf(*geom, "POINT(%f %f)", &lon, &lat)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse location: %w", err)
	}
	return lat, lon, nil
}

func (s *service) GetRide(ctx context.Context, userID, rideID string) (*dto.RideResponse, error) {
	cacheKey := fmt.Sprintf("ride:active:%s", rideID)
	var cached models.Ride

	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil {
		if cached.RiderID == userID || (cached.DriverID != nil && *cached.DriverID == userID) {
			response := dto.ToRideResponse(&cached)

			if cached.Status == "accepted" && cached.DriverID != nil {
				driverLocation, locErr := s.trackingService.GetDriverLocation(ctx, *cached.DriverID)
				if locErr == nil && driverLocation != nil {
					response.DriverLocation = &dto.LocationDTO{
						Latitude:  driverLocation.Latitude,
						Longitude: driverLocation.Longitude,
					}
				}
			}

			return response, nil
		}
	}

	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NotFoundError("Ride")
		}
		return nil, response.InternalServerError("Failed to fetch ride", err)
	}

	if ride.RiderID != userID && (ride.DriverID == nil || *ride.DriverID != userID) {
		return nil, response.ForbiddenError("Not authorized to view this ride")
	}

	response := dto.ToRideResponse(ride)

	if ride.Status == "accepted" && ride.DriverID != nil {
		driverLocation, locErr := s.trackingService.GetDriverLocation(ctx, *ride.DriverID)
		if locErr == nil && driverLocation != nil {
			response.DriverLocation = &dto.LocationDTO{
				Latitude:  driverLocation.Latitude,
				Longitude: driverLocation.Longitude,
			}
		}
	}

	return response, nil
}

func (s *service) ListRides(ctx context.Context, userID string, role string, req dto.ListRidesRequest) ([]*dto.RideListResponse, int64, error) {
	req.SetDefaults()

	filters := map[string]interface{}{
		"role": role,
	}

	if req.Status != "" {
		filters["status"] = req.Status
	}

	rides, total, err := s.repo.ListRides(ctx, userID, filters, req.Page, req.Limit)
	if err != nil {
		return nil, 0, response.InternalServerError("Failed to fetch rides", err)
	}

	result := make([]*dto.RideListResponse, len(rides))
	for i, ride := range rides {
		result[i] = dto.ToRideListResponse(ride)
	}

	return result, total, nil
}

func (s *service) CancelRide(ctx context.Context, userID, rideID string, req dto.CancelRideRequest) error {
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return response.NotFoundError("Ride (to cancel)")
	}

	if ride.RiderID == "" {
		logger.Error("invalid ride: missing riderID", "rideID", rideID)
		return response.InternalServerError("Invalid ride data", nil)
	}

	var isRider bool
	var isDriver bool
	var driverProfile *models.DriverProfile
	var driverProfileID string

	isRider = ride.RiderID == userID

	isDriver = ride.DriverID != nil && *ride.DriverID == userID

	if isDriver {
		driverProfile, err = s.driversRepo.FindDriverByUserID(ctx, userID)
		if err == nil && driverProfile != nil {
			driverProfileID = driverProfile.ID
		}
	}

	if !isRider && !isDriver {
		logger.Warn("unauthorized cancellation attempt",
			"userID", userID,
			"rideID", rideID,
			"riderID", ride.RiderID,
			"rideDriverID", ride.DriverID,
		)
		return response.ForbiddenError("Not authorized to cancel this ride")
	}

	if ride.Status == "completed" {
		return response.BadRequest("Cannot cancel a completed ride")
	}
	if ride.Status == "cancelled" {
		return response.BadRequest("Ride was already cancelled")
	}

	var riderCancellationFee float64
	var driverPenalty float64
	cancelledBy := "rider"
	originalStatus := ride.Status

	if isDriver {
		cancelledBy = "driver"
	}

	switch ride.Status {
	case "searching":
		riderCancellationFee = 0.0
		driverPenalty = 0.0
	case "accepted", "arrived":
		if isRider {
			riderCancellationFee = 2.0 
		} else {
			driverPenalty = 3.0 
		}
	case "started":
		if isRider {
			riderCancellationFee = 5.0 
			driverPenalty = 0.0        
		} else {
			driverPenalty = 10.0       
			riderCancellationFee = 0.0 
		}
	}

	ride.Status = "cancelled"
	ride.CancellationReason = req.Reason
	ride.CancelledBy = &cancelledBy
	ride.CancelledAt = &time.Time{}
	*ride.CancelledAt = time.Now()

	if err := s.repo.UpdateRide(ctx, ride); err != nil {
		return response.InternalServerError("Failed to cancel ride", err)
	}

	logger.Info("ride cancellation initiated",
		"rideID", rideID,
		"cancelledBy", cancelledBy,
		"userID", userID,
		"isRider", isRider,
		"isDriver", isDriver,
		"driverProfileID", driverProfileID,
		"rideStatus", ride.Status,
		"riderCancellationFee", riderCancellationFee,
		"driverPenalty", driverPenalty,
	)

	var driver *models.DriverProfile
	if ride.DriverID != nil {
		if driverProfile != nil && driverProfile.UserID == *ride.DriverID {
			driver = driverProfile
		} else {
			driver, err = s.driversRepo.FindDriverByUserID(ctx, *ride.DriverID)
			if err != nil {
				logger.Error("failed to fetch driver for cancellation",
					"error", err,
					"driverUserID", *ride.DriverID,
					"rideID", rideID,
				)
			}
		}
	}
	if ride.WalletHoldID != nil {
		switch {
		case ride.Status == "started" && isRider:
			if _, err := s.walletService.CaptureHold(ctx, ride.RiderID, walletdto.CaptureHoldRequest{
				HoldID:      *ride.WalletHoldID,
				Amount:      &riderCancellationFee,
				Description: "Cancellation fee for ongoing ride",
			}); err != nil {
				logger.Error("failed to capture ongoing ride cancellation fee", "error", err, "rideID", rideID)
			} else {
				logger.Info("ongoing ride cancellation fee captured",
					"rideID", rideID,
					"amount", riderCancellationFee,
					"riderID", ride.RiderID,
				)

				if driver != nil {
					s.walletService.CreditWallet(
						ctx,
						driver.UserID,
						riderCancellationFee,
						"cancellation_fee",
						rideID,
						"Compensation for cancelled ongoing ride",
						nil,
					)

					logger.Info("driver compensated for ongoing ride cancellation",
						"rideID", rideID,
						"driverID", driver.ID,
						"driverUserID", driver.UserID,
						"amount", riderCancellationFee,
					)
				}
			}

		case ride.Status == "started" && isDriver:
			if err := s.walletService.ReleaseHold(ctx, ride.RiderID, walletdto.ReleaseHoldRequest{
				HoldID: *ride.WalletHoldID,
			}); err != nil {
				logger.Error("failed to release rider hold", "error", err, "rideID", rideID)
			} else {
				logger.Info("rider hold released due to driver cancellation",
					"rideID", rideID,
					"riderID", ride.RiderID,
				)
			}

			if driver != nil && driverPenalty > 0 {
				_, err := s.walletService.DeductPenalty(
					ctx,
					driver.UserID,
					driverPenalty,
					"cancelled_ongoing_ride",
					rideID,
				)
				if err != nil {
					logger.Error("failed to deduct driver penalty", "error", err, "driverID", driver.ID, "rideID", rideID)
				} else {
					logger.Info("driver cancellation penalty deducted",
						"rideID", rideID,
						"driverID", driver.ID,
						"driverUserID", driver.UserID,
						"penalty", driverPenalty,
					)

					compensationAmount := driverPenalty * 0.5
					s.walletService.CreditWallet(
						ctx,
						ride.RiderID,
						compensationAmount,
						"cancellation_compensation",
						rideID,
						"Compensation for driver cancelling ongoing ride",
						nil,
					)

					logger.Info("rider compensated for driver cancellation",
						"rideID", rideID,
						"riderID", ride.RiderID,
						"amount", compensationAmount,
					)
				}
			}

		case riderCancellationFee > 0:
			if _, err := s.walletService.CaptureHold(ctx, ride.RiderID, walletdto.CaptureHoldRequest{
				HoldID:      *ride.WalletHoldID,
				Amount:      &riderCancellationFee,
				Description: "Cancellation fee",
			}); err != nil {
				logger.Error("failed to capture cancellation fee", "error", err, "rideID", rideID)
			} else {
				logger.Info("cancellation fee captured",
					"rideID", rideID,
					"amount", riderCancellationFee,
					"riderID", ride.RiderID,
				)
				if driver != nil {
					s.walletService.CreditWallet(
						ctx,
						driver.UserID,
						riderCancellationFee,
						"cancellation_fee",
						rideID,
						"Cancellation fee compensation",
						nil,
					)

					logger.Info("cancellation fee credited to driver",
						"rideID", rideID,
						"driverID", driver.ID,
						"amount", riderCancellationFee,
					)
				}
			}
		case driverPenalty > 0 && driver != nil:
			if err := s.walletService.ReleaseHold(ctx, ride.RiderID, walletdto.ReleaseHoldRequest{
				HoldID: *ride.WalletHoldID,
			}); err != nil {
				logger.Error("failed to release hold", "error", err, "rideID", rideID)
			}
			_, err := s.walletService.DeductPenalty(
				ctx,
				driver.UserID,
				driverPenalty,
				"cancelled_accepted_ride",
				rideID,
			)
			if err != nil {
				logger.Error("failed to deduct driver penalty", "error", err, "driverID", driver.ID, "rideID", rideID)
			} else {
				logger.Info("driver cancellation penalty deducted",
					"rideID", rideID,
					"driverID", driver.ID,
					"driverUserID", driver.UserID,
					"penalty", driverPenalty,
				)
			}
		default:
			if err := s.walletService.ReleaseHold(ctx, ride.RiderID, walletdto.ReleaseHoldRequest{
				HoldID: *ride.WalletHoldID,
			}); err != nil {
				logger.Error("failed to release hold", "error", err, "rideID", rideID)
			} else {
				logger.Info("wallet hold released",
					"rideID", rideID,
					"riderID", ride.RiderID,
				)
			}
		}
	}

	if ride.DriverID != nil {
		driverUserID := *ride.DriverID

		if driverProfile == nil {
			driverProfile, _ = s.driversRepo.FindDriverByUserID(ctx, driverUserID)
		}

		if driverProfile != nil && driverProfile.ID != "" {
			driverProfileID = driverProfile.ID 

			if err := s.driversRepo.UpdateDriverStatus(ctx, driverProfileID, "online"); err != nil {
				logger.Warn("failed to update driver status", "error", err, "driverID", driverProfileID)
			}

			busyKey := fmt.Sprintf("driver:busy:%s", driverProfileID)
			if err := cache.Delete(ctx, busyKey); err != nil {
				logger.Warn("failed to clear driver busy cache", "error", err, "busyKey", busyKey)
			}

			activeRideCacheKey := fmt.Sprintf("driver:active:ride:%s", driverProfileID)
			if err := cache.Delete(ctx, activeRideCacheKey); err != nil {
				logger.Warn("failed to clear active ride cache", "error", err, "activeRideCacheKey", activeRideCacheKey)
			}

			logger.Debug("driver caches cleared",
				"driverProfileID", driverProfileID,
				"rideID", rideID,
				"clearedKeys", []string{busyKey, activeRideCacheKey},
			)
		}
	}

	if driver != nil && !isDriver {
		websocketutil.SendToUser(driver.UserID, websocket.TypeRideCancelled, map[string]interface{}{
			"rideId":       rideID,
			"message":      "Ride was cancelled by rider",
			"reason":       req.Reason,
			"compensation": riderCancellationFee,
			"timestamp":    time.Now().UTC(),
		})

		logger.Info("driver notified of cancellation",
			"rideID", rideID,
			"driverID", driver.ID,
			"cancelledBy", "rider",
		)
	}

	rideCacheKey := fmt.Sprintf("ride:active:%s", rideID)
	if err := cache.Delete(ctx, rideCacheKey); err != nil {
		logger.Warn("failed to clear ride cache", "error", err, "rideCacheKey", rideCacheKey)
	}

	if isDriver {
		var message string
		var compensation float64

		if originalStatus == "started" {
			message = "Ongoing ride was cancelled by driver. You will receive compensation."
			compensation = driverPenalty * 0.5
		} else {
			message = "Ride was cancelled by driver"
			compensation = 0
		}

		websocketutil.SendToUser(ride.RiderID, websocket.TypeRideCancelled, map[string]interface{}{
			"rideId":       rideID,
			"message":      message,
			"reason":       req.Reason,
			"compensation": compensation,
			"timestamp":    time.Now().UTC(),
		})

		logger.Info("rider notified of cancellation",
			"rideID", rideID,
			"riderID", ride.RiderID,
			"cancelledBy", "driver",
		)
	}

	logger.Info("ride cancelled successfully",
		"rideID", rideID,
		"cancelledBy", cancelledBy,
		"userID", userID,
		"rideStatus", ride.Status,
		"riderFee", riderCancellationFee,
		"driverPenalty", driverPenalty,
		"status", "success",
	)

	return nil
}

func (s *service) TriggerSOS(ctx context.Context, riderID, rideID string, latitude, longitude float64) error {
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return response.NotFoundError("Ride")
	}

	if ride.RiderID != riderID {
		return response.ForbiddenError("Not authorized to trigger SOS for this ride")
	}

	if ride.Status != "started" && ride.Status != "accepted" && ride.Status != "arrived" {
		return response.BadRequest("SOS can only be triggered during an active ride")
	}

	if s.sosService == nil {
		logger.Warn("SOS service not initialized, logging emergency only",
			"riderID", riderID,
			"rideID", rideID,
			"location", map[string]float64{"lat": latitude, "lon": longitude},
		)
		websocketutil.BroadcastToAll(websocket.TypeSOSAlert, map[string]interface{}{
			"type":      "sos_emergency",
			"rideId":    rideID,
			"riderId":   riderID,
			"latitude":  latitude,
			"longitude": longitude,
			"timestamp": time.Now().UTC(),
		})
		return nil
	}

	sosReq := sosdto.TriggerSOSRequest{
		RideID:    &rideID,
		Latitude:  latitude,
		Longitude: longitude,
	}

	alert, err := s.sosService.TriggerSOS(ctx, riderID, sosReq)
	if err != nil {
		logger.Error("failed to trigger SOS",
			"error", err,
			"riderID", riderID,
			"rideID", rideID,
		)
		return response.InternalServerError("Failed to trigger SOS", err)
	}

	websocketutil.BroadcastToAll(websocket.TypeSOSAlert, map[string]interface{}{
		"alertId":   alert.ID,
		"rideId":    rideID,
		"riderId":   riderID,
		"latitude":  latitude,
		"longitude": longitude,
		"severity":  "critical",
		"timestamp": time.Now().UTC(),
	})

	logger.Warn("SOS ALERT TRIGGERED",
		"alertId", alert.ID,
		"riderID", riderID,
		"rideID", rideID,
		"latitude", latitude,
		"longitude", longitude,
	)

	return nil
}

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0

	dLat := (lat2 - lat1) * (3.14159265359 / 180.0)
	dLon := (lon2 - lon1) * (3.14159265359 / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(3.14159265359/180.0))*math.Cos(lat2*(3.14159265359/180.0))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}