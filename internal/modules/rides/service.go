package rides

import (
	"context"
	"errors"
	"fmt"
	"math"
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
	// Rider actions
	CreateRide(ctx context.Context, riderID string, req dto.CreateRideRequest) (*dto.RideResponse, error)
	GetRide(ctx context.Context, userID, rideID string) (*dto.RideResponse, error)
	ListRides(ctx context.Context, userID string, role string, req dto.ListRidesRequest) ([]*dto.RideListResponse, int64, error)
	CancelRide(ctx context.Context, userID, rideID string, req dto.CancelRideRequest) error

	// Driver actions
	AcceptRide(ctx context.Context, driverID, rideID string) (*dto.RideResponse, error)
	RejectRide(ctx context.Context, driverID, rideID string, req dto.RejectRideRequest) error
	MarkArrived(ctx context.Context, driverID, rideID string) (*dto.RideResponse, error)
	StartRide(ctx context.Context, driverID, rideID string, req dto.StartRideRequest) (*dto.RideResponse, error)
	CompleteRide(ctx context.Context, driverID, rideID string, req dto.CompleteRideRequest) (*dto.RideResponse, error)

	// Safety features
	TriggerSOS(ctx context.Context, riderID, rideID string, latitude, longitude float64) error

	// Internal
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
	ridePINService    ridepinservice.Service    // NEW
	profileService    profile.Service           // NEW
	sosService        sos.Service               // NEW
	promotionsService promotionsservice.Service // NEW
	ratingsService    ratingsservice.Service    // NEW
	fraudService      fraudservice.Service      // NEW
	batchingService   batchingservice.Service   // NEW: For request batching
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

	// Set up batch expiration callback to process batches when they expire
	if batchingService != nil {
		batchingService.SetBatchExpireCallback(func(batchID string) {
			ctx := context.Background()
			logger.Info("🔄 Processing expired batch",
				"batchID", batchID,
			)
			if result, err := batchingService.ProcessBatch(ctx, batchID); err != nil {
				logger.Error("Failed to process batch",
					"batchID", batchID,
					"error", err,
				)
			} else {
				logger.Info("✅ Batch processing complete",
					"batchID", batchID,
					"assignments", len(result.Assignments),
					"unmatched", len(result.UnmatchedIDs),
				)
				// Process the matching result
				svc.processMatchingResult(ctx, result)
			}
		})
	}

	return svc
}

// ========================================
// CreateRide with Profile & Promotions
// ========================================
func (s *service) CreateRide(ctx context.Context, riderID string, req dto.CreateRideRequest) (*dto.RideResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	// Check if user has emergency contact set (optional warning)
	user, err := s.adminRepo.FindUserByID(ctx, riderID)
	if err == nil && user.EmergencyContactPhone == "" {
		logger.Warn("Ride created without emergency contact", "userID", riderID)
	}

	// Handle saved locations if provided
	// Integrate with profile service when methods are available
	if req.SavedLocationID != nil && s.profileService != nil {
		savedLocs, err := s.profileService.GetSavedLocations(ctx, riderID)
		if err != nil {
			return nil, response.BadRequest("Failed to fetch saved locations")
		}

		// Find the matching saved location
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

	// 1. Get fare estimate
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

	// 1a. Calculate enhanced surge with time-based and demand-based logic
	// Generate simple geohash from pickup location for demand tracking (5-char geohash)
	geohash := fmt.Sprintf("%.1f_%.1f", req.PickupLat, req.PickupLon)
	surgeCalc, err := s.pricingService.CalculateCombinedSurge(ctx, req.VehicleTypeID, geohash, req.PickupLat, req.PickupLon)
	if err != nil {
		logger.Warn("failed to calculate combined surge, using basic surge", "error", err)
	} else if surgeCalc != nil && surgeCalc.AppliedMultiplier > fareEstimate.SurgeMultiplier {
		// Use the higher surge multiplier
		fareEstimate.SurgeMultiplier = surgeCalc.AppliedMultiplier
		fareEstimate.SurgeAmount = (fareEstimate.SubTotal) * (surgeCalc.AppliedMultiplier - 1.0)
		fareEstimate.TotalFare = fareEstimate.SubTotal + fareEstimate.SurgeAmount
		logger.Info("Enhanced surge applied",
			"timeBasedSurge", surgeCalc.TimeBasedMultiplier,
			"demandBasedSurge", surgeCalc.DemandBasedMultiplier,
			"combinedSurge", surgeCalc.AppliedMultiplier,
			"reason", surgeCalc.Reason)
	}

	// 1b. Calculate and store ETA estimate
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

	// Apply promo code if provided
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

	// ✅ Calculate amount to hold (after applying free credits)
	walletInfo, err := s.walletService.GetWallet(ctx, riderID)
	if err != nil {
		logger.Warn("failed to get wallet info", "error", err, "riderID", riderID)
	}

	amountToHold := finalAmount
	if walletInfo != nil && walletInfo.FreeRideCredits > 0 {
		// Free ride credits can cover part or all of the fare
		if walletInfo.FreeRideCredits >= finalAmount {
			amountToHold = 0 // Free credits cover everything
			logger.Info("ride fully covered by free credits", "riderID", riderID, "fareAmount", finalAmount, "freeCredits", walletInfo.FreeRideCredits)
		} else {
			// Free credits cover partial amount, hold the remaining
			amountToHold = finalAmount - walletInfo.FreeRideCredits
			logger.Info("ride partially covered by free credits", "riderID", riderID, "fareAmount", finalAmount, "freeCredits", walletInfo.FreeRideCredits, "amountToHold", amountToHold)
		}
	}

	rideID := uuid.New().String()

	// 2. Hold funds with ReferenceID (only if amount > 0)
	var holdID *string
	if amountToHold > 0 {
		holdReq := walletdto.HoldFundsRequest{
			Amount:        amountToHold,
			ReferenceType: "ride",
			ReferenceID:   rideID,
			HoldDuration:  1800, // 30 minutes
		}

		holdResp, err := s.walletService.HoldFunds(ctx, riderID, holdReq)
		if err != nil {
			return nil, response.BadRequest("Insufficient wallet balance. Please add funds.")
		}
		holdID = &holdResp.ID
		logger.Info("funds held for ride", "rideID", rideID, "holdID", holdResp.ID, "amount", amountToHold)
	} else {
		logger.Info("no funds to hold - free credits cover entire fare", "rideID", rideID)
	}

	// 3. Create ride
	ride := &models.Ride{
		ID:                rideID,
		RiderID:           riderID,
		VehicleTypeID:     req.VehicleTypeID,
		Status:            "searching",
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

	// 4. Store in Redis for quick access
	cacheKey := fmt.Sprintf("ride:active:%s", rideID)
	cache.SetJSON(ctx, cacheKey, ride, 30*time.Minute)

	// 5. Start finding driver asynchronously using batching service
	go func() {
		bgCtx := context.Background()

		// ✅ OPTION A: Use intelligent request batching with 4-factor driver ranking
		if s.batchingService != nil {
			// Add request to batch for intelligent matching
			// Requests in the same vehicle type get batched for 10 seconds
			// and matched using 4-factor driver ranking algorithm
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
				// Batch will be processed when:
				// - 10-second window expires, or
				// - Maximum batch size (20 requests) is reached
				// Matches are then delivered to drivers via processMatchingResult()
				return
			}

			logger.Warn("Failed to add request to batch, falling back to sequential matching",
				"error", err,
				"rideID", ride.ID,
			)
		}

		// ✅ FALLBACK: Sequential driver matching (if batching unavailable)
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

			s.wsHelper.SendRideStatusToBoth(bgCtx, riderID, "", rideID, "cancelled", "No drivers available. Your payment has been refunded.")
		}
	}()

	// 6. Notify rider via WebSocket
	s.wsHelper.SendRideStatusToBoth(ctx, riderID, "", rideID, "searching", "Searching for nearby drivers...")

	logger.Info("ride created",
		"rideID", rideID,
		"riderID", riderID,
		"estimatedFare", finalAmount,
		"surge", fareEstimate.SurgeMultiplier,
		"promoDiscount", promoDiscount,
	)

	// Fetch fresh ride data for response
	freshRide, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		logger.Error("failed to fetch fresh ride data for response", "error", err, "rideID", rideID)
		// Return the ride we have instead of nil
		return dto.ToRideResponse(ride), nil
	}
	return dto.ToRideResponse(freshRide), nil
}

// processMatchingResult processes the results from batch matching
// and processes assignments for each matched ride-driver pair
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

	// Process successful assignments
	for _, assignment := range result.Assignments {
		logger.Info("✅ Batch assignment match found",
			"rideID", assignment.RideID,
			"driverID", assignment.DriverID,
			"score", assignment.Score,
			"distance", assignment.Distance,
			"eta", assignment.ETA,
		)

		// Use the existing driver assignment logic
		// which handles all the database updates, notifications, etc.
		go func(assign batchingdto.OptimalAssignment) {
			bgCtx := context.Background()

			// Fetch the ride
			ride, err := s.repo.FindRideByID(bgCtx, assign.RideID)
			if err != nil {
				logger.Error("failed to fetch ride for batch assignment",
					"rideID", assign.RideID,
					"error", err,
				)
				return
			}

			// Update ride with assigned driver
			driverIDPtr := assign.DriverID // Convert to pointer
			ride.DriverID = &driverIDPtr
			ride.Status = "accepted"
			ride.AcceptedAt = ptr(time.Now())

			// Persist to database
			if err := s.repo.UpdateRide(bgCtx, ride); err != nil {
				logger.Error("failed to update ride with driver assignment",
					"rideID", assign.RideID,
					"driverID", assign.DriverID,
					"error", err,
				)
				return
			}

			logger.Info("📤 Driver assigned to ride from batch",
				"rideID", assign.RideID,
				"driverID", assign.DriverID,
				"riderID", ride.RiderID,
			)

			// Send rider notification using existing helper
			s.wsHelper.SendRideAccepted(ride.RiderID, map[string]interface{}{
				"rideId":    assign.RideID,
				"driverId":  assign.DriverID,
				"eta":       assign.ETA,
				"distance":  assign.Distance,
				"timestamp": time.Now().Format(time.RFC3339),
			})

			// Send driver notification using existing helper
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

	// Log unmatched requests for fallback matching
	if len(result.UnmatchedIDs) > 0 {
		logger.Info("Processing unmatched rides from batch",
			"batchID", result.BatchID,
			"unmatchedCount", len(result.UnmatchedIDs),
		)

		// Fallback to sequential matching for unmatched rides
		for _, rideID := range result.UnmatchedIDs {
			logger.Warn("🔄 Unmatched ride, attempting sequential matching",
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

// Helper function to get pointer to time.Time
func ptr(t time.Time) *time.Time {
	return &t
}

// start from here
func (s *service) FindDriverForRide(ctx context.Context, rideID string) error {
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return err
	}

	if ride.Status != "searching" {
		return nil
	}

	// ========================================
	// RIDER RATING FILTER
	// ========================================
	// Get rider profile to check rating
	// Lower rated riders (< 4.0) get longer wait times by narrowing search radius
	riderProfile, err := s.ridersRepo.FindByUserID(ctx, ride.RiderID)
	var riderRating float64 = 5.0 // Default to high rating
	if err == nil && riderProfile != nil {
		riderRating = riderProfile.Rating
		logger.Info("rider rating check",
			"rideID", rideID,
			"riderID", ride.RiderID,
			"riderRating", riderRating,
		)
	}

	// Adjust radius search based on rider rating
	// High rated riders (4.5+): Start from 3km
	// Medium rated (3.5-4.5): Start from 5km
	// Low rated (< 3.5): Start from 8km only
	radii := []float64{3.0, 5.0, 8.0} // Default for good riders
	if riderRating < 3.5 {
		radii = []float64{8.0} // Low-rated riders only search at 8km
		logger.Warn("low-rated rider search reduced",
			"rideID", rideID,
			"riderRating", riderRating,
			"searchRadius", "8km only",
		)
	} else if riderRating < 4.0 {
		radii = []float64{5.0, 8.0} // Medium-rated riders skip 3km
		logger.Info("medium-rated rider search radius adjusted",
			"rideID", rideID,
			"riderRating", riderRating,
			"searchRadius", "5-8km",
		)
	}

	// ========================================
	// BATCHING INTEGRATION - OPTION A (Recommended)
	// ========================================
	// Uncomment to enable request batching with intelligent driver matching
	// This collects the request in a 10-second batch window and performs
	// optimal matching across all requests in the batch using the 4-factor
	// driver ranking algorithm (Rating 40%, Acceptance 30%, Cancellation 20%, Completion 10%)
	//
	// if s.batchingService != nil {
	//     // Add request to batch for this vehicle type
	//     batchID, err := s.batchingService.AddRequestToBatch(ctx, RideRequestInfo{
	//         RideID:        ride.ID,
	//         RiderID:       ride.RiderID,
	//         PickupLat:     ride.PickupLat,
	//         PickupLon:     ride.PickupLon,
	//         DropoffLat:    ride.DropoffLat,
	//         DropoffLon:    ride.DropoffLon,
	//         VehicleTypeID: ride.VehicleTypeID,
	//         RequestedAt:   time.Now(),
	//     })
	//     if err == nil {
	//         logger.Info("Request added to batch", "batchID", batchID, "rideID", ride.ID)
	//         // Batch will be processed when window expires or max size reached
	//         // Assignments will be delivered to drivers via processMatchingResult()
	//     }
	//     return nil
	// }
	//
	// ========================================
	// Current flow: Sequential driver assignment (OPTION B - Fallback)
	// ========================================

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
		// acceptedDriverID is the driver profile ID
		// We need to fetch the user ID to pass to assignDriverToRide
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
			// ✅ CRITICAL FIX: Check ride status first
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

// ========================================
// AcceptRide with Fraud Detection
// ========================================
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

	// Check rider's fraud risk score
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

	// ✅ FIXED: Pass both userID and driverID to assignDriverToRide
	// userID is needed for the rides table foreign key
	// driverID (driver profile ID) is needed for driver operations
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

	// Fetch fresh ride data for response
	freshRide, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		logger.Error("failed to fetch fresh ride data for response", "error", err, "rideID", rideID)
		// Return the ride we have instead of nil
		if ride != nil {
			return dto.ToRideResponse(ride), nil
		}
		return nil, response.InternalServerError("Failed to fetch ride data", err)
	}

	if freshRide == nil {
		logger.Error("fetched ride is nil", "rideID", rideID)
		// Return the ride we have
		return dto.ToRideResponse(ride), nil
	}

	return dto.ToRideResponse(freshRide), nil
}

// ✅ FIXED: Accept userID and fetch driver profile first
func (s *service) RejectRide(ctx context.Context, userID, rideID string, req dto.RejectRideRequest) error {
	// ✅ STEP 1: Fetch driver profile using user ID
	driver, err := s.driversRepo.FindDriverByUserID(ctx, userID)
	if err != nil {
		logger.Error("driver profile not found for user",
			"error", err,
			"userID", userID,
		)
		return response.NotFoundError("Driver profile not found")
	}

	driverID := driver.ID // ✅ Now we have the actual driver profile ID

	// ✅ STEP 2: Find ride request using driver profile ID
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

// ✅ FIXED: Accept userID and fetch driver profile first
func (s *service) MarkArrived(ctx context.Context, userID, rideID string) (*dto.RideResponse, error) {
	// ✅ STEP 1: Fetch driver profile using user ID
	driver, err := s.driversRepo.FindDriverByUserID(ctx, userID)
	if err != nil {
		logger.Error("driver profile not found for user",
			"error", err,
			"userID", userID,
		)
		return nil, response.NotFoundError("Driver profile not found")
	}

	driverID := driver.ID // ✅ Now we have the actual driver profile ID

	// ✅ STEP 2: Fetch ride and validate authorization
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return nil, response.NotFoundError("Ride")
	}

	// ✅ FIXED: Compare with userID, not driverID
	// ride.DriverID now contains the USER ID, not the driver profile ID
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

	// ✅ Notify rider via WebSocket (ride.RiderID is user ID)
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

	// Fetch fresh ride data for response
	freshRide, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		logger.Error("failed to fetch fresh ride data for response", "error", err, "rideID", rideID)
		// Return the ride we have instead of nil
		if ride != nil {
			return dto.ToRideResponse(ride), nil
		}
		return nil, response.InternalServerError("Failed to fetch ride data", err)
	}

	if freshRide == nil {
		logger.Error("fetched ride is nil", "rideID", rideID)
		// Return the ride we have
		return dto.ToRideResponse(ride), nil
	}

	return dto.ToRideResponse(freshRide), nil
}

// ========================================
// StartRide with RidePin Verification
// ========================================
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

	// ✅ FIXED: Compare with userID, not driverID
	// ride.DriverID now contains the USER ID, not the driver profile ID
	if ride.DriverID == nil || *ride.DriverID != userID {
		logger.Warn("unauthorized attempt to start ride", "userID", userID, "driverProfileID", driverID, "rideID", rideID)
		return nil, response.ForbiddenError("Not authorized")
	}

	if ride.Status != "accepted" && ride.Status != "arrived" {
		logger.Warn("invalid ride status for start", "rideID", rideID, "status", ride.Status)
		return nil, response.BadRequest("Ride must be accepted or arrived to start")
	}

	// CRITICAL: Verify rider's PIN
	logger.Info("verifying ride PIN", "rideID", rideID, "riderID", ride.RiderID)
	if err := s.ridePINService.VerifyRidePIN(ctx, ride.RiderID, req.RiderPIN); err != nil {
		logger.Warn("invalid ride PIN attempt at start",
			"rideID", rideID,
			"driverID", driverID,
			"riderID", ride.RiderID)
		return nil, response.BadRequest("Invalid Rider PIN. Please ask the rider for their 4-digit Ride PIN.")
	}

	logger.Info("ride PIN verified at start", "rideID", rideID)

	// Calculate wait time charge if applicable
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

	// Update driver status and cache
	logger.Info("updating driver status", "driverID", driverID)
	if err := s.driversRepo.UpdateDriverStatus(ctx, driverID, "on_trip"); err != nil {
		logger.Warn("failed to update driver status", "error", err, "driverID", driverID)
	}

	cacheKey := fmt.Sprintf("ride:active:%s", rideID)
	logger.Info("setting ride cache", "cacheKey", cacheKey)
	if err := cache.SetJSON(ctx, cacheKey, ride, 30*time.Minute); err != nil {
		logger.Warn("failed to cache ride", "error", err, "rideID", rideID)
	}

	// Notify via WebSocket (safely ignore errors)
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

	// Fetch fresh ride data for response
	logger.Info("fetching fresh ride data", "rideID", rideID)
	freshRide, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		logger.Error("failed to fetch fresh ride data for response", "error", err, "rideID", rideID)
		// Return the updated ride we have instead of nil
		if ride != nil {
			logger.Info("returning cached ride response", "rideID", rideID)
			return dto.ToRideResponse(ride), nil
		}
		return nil, response.InternalServerError("Failed to fetch ride data", err)
	}

	if freshRide == nil {
		logger.Error("fetched ride is nil", "rideID", rideID)
		// Return the ride we have
		return dto.ToRideResponse(ride), nil
	}

	logger.Info("converting freshRide to response", "rideID", rideID)
	resp := dto.ToRideResponse(freshRide)
	logger.Info("response conversion complete", "rideID", rideID)
	return resp, nil
}

// ========================================
// CompleteRide with all integrations
// ========================================
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

	// ✅ FIXED: Compare with userID, not driverID
	// ride.DriverID now contains the USER ID, not the driver profile ID
	if ride.DriverID == nil || *ride.DriverID != userID {
		logger.Warn("unauthorized attempt to complete ride", "userID", userID, "driverProfileID", driverID, "rideID", rideID)
		return nil, response.ForbiddenError("Not authorized")
	}

	if ride.Status != "started" {
		return nil, response.BadRequest("Ride must be started first")
	}

	// CRITICAL: Verify rider's PIN
	if err := s.ridePINService.VerifyRidePIN(ctx, ride.RiderID, req.RiderPIN); err != nil {
		logger.Warn("invalid ride PIN attempt at completion", "rideID", rideID, "driverID", driverID, "riderID", ride.RiderID)
		return nil, response.BadRequest("Invalid Rider PIN. Please ask the rider for their 4-digit Ride PIN.")
	}

	logger.Info("ride PIN verified at completion", "rideID", rideID)

	// ✅ VERIFY: Driver is within 150m of dropoff location
	distanceToDropoff := location.HaversineDistance(req.DriverLat, req.DriverLon, ride.DropoffLat, ride.DropoffLon)
	distanceKm := distanceToDropoff
	const maxCompletionRadiusKm = 0.15 // 150 meters

	if distanceKm > maxCompletionRadiusKm {
		logger.Warn("driver outside completion radius", "rideID", rideID, "distanceKm", distanceKm, "maxRadiusKm", maxCompletionRadiusKm)
		return nil, response.BadRequest(fmt.Sprintf("You must be within 150 meters of the destination. Current distance: %.0f meters", distanceKm*1000))
	}

	logger.Info("driver verified within 150 meter radius", "rideID", rideID, "distanceKm", distanceKm)

	// Calculate actual fare
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

	actualFare := actualFareResp.TotalFare

	// Add wait time charge if any
	if ride.WaitTimeCharge != nil {
		actualFare += *ride.WaitTimeCharge
	}

	// Add destination change charge if any
	if ride.DestinationChangeCharge != nil {
		actualFare += *ride.DestinationChangeCharge
	}

	// Apply promo discount
	if ride.PromoDiscount != nil {
		actualFare -= *ride.PromoDiscount
	}

	// Check for active SOS and resolve when sosService is available
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

	// Update ride
	ride.ActualDistance = &req.ActualDistance
	ride.ActualDuration = &req.ActualDuration
	ride.ActualFare = &actualFare
	ride.Status = "completed"
	completedAt := time.Now()
	ride.CompletedAt = &completedAt

	if err := s.repo.UpdateRide(ctx, ride); err != nil {
		return nil, response.InternalServerError("Failed to complete ride", err)
	}

	// ✅ Capture hold (marks cash payment as received)
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

	// Apply promo code usage
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

	// ✅ Credit driver earnings with commission-based split
	// Using actual fare response which already has driver payout calculated
	driverEarnings := actualFareResp.DriverPayout           // Driver gets reduced amount after commission
	platformCommission := actualFareResp.PlatformCommission // Platform keeps commission

	logger.Info("ride payout breakdown",
		"rideID", rideID,
		"totalFare", actualFare,
		"driverEarnings", driverEarnings,
		"platformCommission", platformCommission,
		"commissionRate", actualFareResp.CommissionRate,
	)

	_, err = s.walletService.CreditDriverWallet(
		ctx,
		driver.UserID,
		driverEarnings,
		"ride_earnings",
		rideID,
		fmt.Sprintf("Earnings from ride %s (after %.1f%% commission)", rideID, actualFareResp.CommissionRate),
		nil,
	)
	if err != nil {
		logger.Error("failed to credit driver wallet", "error", err, "rideID", rideID)
	}

	// Update stats
	s.driversRepo.IncrementTrips(ctx, driverID)
	s.driversRepo.UpdateEarnings(ctx, driverID, driverEarnings)
	s.ridersRepo.IncrementTotalRides(ctx, ride.RiderID)

	// Update driver status back to online
	if err := s.driversRepo.UpdateDriverStatus(ctx, driverID, "online"); err != nil {
		logger.Warn("failed to update driver status", "error", err, "driverID", driverID)
	}

	// CLEANUP: Clear all caches (safely ignore errors)
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

	// Notify both parties (safely ignore errors)
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

	// Prompt for ratings asynchronously
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

	// Trigger fraud detection
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

	// Fetch fresh ride data for response
	freshRide, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		logger.Error("failed to fetch fresh ride data for response", "error", err, "rideID", rideID)
		// Return the ride we have instead of nil
		if ride != nil {
			return dto.ToRideResponse(ride), nil
		}
		return nil, response.InternalServerError("Failed to fetch ride data", err)
	}

	if freshRide == nil {
		logger.Error("fetched ride is nil", "rideID", rideID)
		// Return the ride we have
		return dto.ToRideResponse(ride), nil
	}

	return dto.ToRideResponse(freshRide), nil
}

// ========================================
// Helper: Prompt Ratings
// ========================================
func (s *service) promptRatings(ctx context.Context, riderID, driverUserID, rideID string) {
	time.Sleep(5 * time.Second)

	// Prompt rider to rate driver (safely ignore errors)
	if err := websocketutil.SendToUser(riderID, websocket.TypeRatingPrompt, map[string]interface{}{
		"rideId":  rideID,
		"message": "How was your ride? Please rate your driver",
	}); err != nil {
		logger.Warn("failed to send rating prompt to rider", "error", err, "rideID", rideID, "riderID", riderID)
	}

	// Prompt driver to rate rider (safely ignore errors)
	if err := websocketutil.SendToUser(driverUserID, websocket.TypeRatingPrompt, map[string]interface{}{
		"rideId":  rideID,
		"message": "Please rate your rider",
	}); err != nil {
		logger.Warn("failed to send rating prompt to driver", "error", err, "rideID", rideID, "driverUserID", driverUserID)
	}
}

// ✅ UPDATED assignDriverToRide with both userID and driverProfileID
// userID: the ID of the user (for rides table foreign key)
// driverProfileID: the ID of the driver profile (for driver operations)
func (s *service) assignDriverToRide(ctx context.Context, rideID, userID, driverProfileID string) error {
	fmt.Println("Run func: assignDriverToRide")

	// Atomic update: only one driver can change status from "searching" to "accepted"
	// ✅ Use userID for the rides table foreign key
	err := s.repo.UpdateRideStatusAndDriver(ctx, rideID, "accepted", "searching", userID)
	if err != nil {
		logger.Warn("failed to assign driver - ride may be already accepted",
			"error", err,
			"rideID", rideID,
			"userID", userID,
		)
		return response.BadRequest("Ride already accepted by another driver")
	}

	// Cancel all other pending requests for this ride
	// ✅ Use driverProfileID for ride requests
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

	// Update driver status
	// ✅ Use driverProfileID for driver operations
	s.driversRepo.UpdateDriverStatus(ctx, driverProfileID, "busy")

	// Mark driver as busy in cache
	busyKey := fmt.Sprintf("driver:busy:%s", driverProfileID)
	cache.Set(ctx, busyKey, "true", 30*time.Minute)

	// Get ride for cache update and location calculation
	ride, _ := s.repo.FindRideByID(ctx, rideID)

	// Update cache
	cacheKey := fmt.Sprintf("ride:active:%s", rideID)
	cache.SetJSON(ctx, cacheKey, ride, 30*time.Minute)

	// Get driver details for notification
	// ✅ Use driverProfileID to fetch driver details
	driver, err := s.driversRepo.FindDriverByID(ctx, driverProfileID)
	if err != nil {
		logger.Error("failed to fetch driver details for notification",
			"error", err,
			"driverProfileID", driverProfileID,
		)
		return nil
	}

	// Safely check if vehicle exists before accessing
	if driver.Vehicle == nil {
		logger.Warn("driver has no vehicle assigned",
			"driverProfileID", driverProfileID,
			"driverName", driver.User.Name,
		)
		return nil
	}

	//  Cache active ride for tracking module
	activeRideCacheKey := fmt.Sprintf("driver:active:ride:%s", driverProfileID)
	cache.SetJSON(ctx, activeRideCacheKey, map[string]string{
		"rideID":  rideID,
		"riderID": ride.RiderID,
	}, 30*time.Minute)

	// ✅ Get driver's current location
	var driverLat, driverLon float64
	var calculatedETA int

	driverLocation, err := s.trackingService.GetDriverLocation(ctx, driverProfileID)
	if err == nil && driverLocation != nil {
		driverLat = driverLocation.Latitude
		driverLon = driverLocation.Longitude

		// Calculate real ETA based on distance
		distance := calculateDistance(driverLat, driverLon, ride.PickupLat, ride.PickupLon)
		calculatedETA = int((distance / 30.0) * 60) // 30 km/h avg city speed
		if calculatedETA < 1 {
			calculatedETA = 1
		}
	} else {
		// Fallback to pickup location and default ETA
		driverLat = ride.PickupLat
		driverLon = ride.PickupLon
		calculatedETA = 10
		logger.Warn("could not get driver location, using defaults",
			"error", err,
			"driverProfileID", driverProfileID,
		)
	}

	// ✅ IMPROVED PAYLOAD with complete driver info
	rideDetails := map[string]interface{}{
		"rideId":   rideID,
		"driverId": driver.ID,
		"userId":   driver.UserID, // ✅ ADDED
		"driver": map[string]interface{}{
			"id":     driver.ID,
			"userId": driver.UserID, // ✅ ADDED
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
		"location": map[string]interface{}{ // ✅ ADDED
			"latitude":  driverLat,
			"longitude": driverLon,
		},
		"message": "Driver is on the way!",
		"eta":     calculatedETA, // ✅ CALCULATED
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

// ✅ UPDATED GetRide with driver location enrichment
func (s *service) GetRide(ctx context.Context, userID, rideID string) (*dto.RideResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("ride:active:%s", rideID)
	var cached models.Ride

	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil {
		// Verify user has access
		if cached.RiderID == userID || (cached.DriverID != nil && *cached.DriverID == userID) {
			response := dto.ToRideResponse(&cached)

			// ✅ Enrich with live driver location if ride is accepted
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

	// Get from database
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NotFoundError("Ride")
		}
		return nil, response.InternalServerError("Failed to fetch ride", err)
	}

	// Verify access
	if ride.RiderID != userID && (ride.DriverID == nil || *ride.DriverID != userID) {
		return nil, response.ForbiddenError("Not authorized to view this ride")
	}

	response := dto.ToRideResponse(ride)

	// ✅ Enrich with live driver location if ride is accepted
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

// ✅ FIXED: Handle both rider and driver authorization correctly
func (s *service) CancelRide(ctx context.Context, userID, rideID string, req dto.CancelRideRequest) error {
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return response.NotFoundError("Ride (to cancel)")
	}

	// ✅ Validate ride data integrity
	if ride.RiderID == "" {
		logger.Error("invalid ride: missing riderID", "rideID", rideID)
		return response.InternalServerError("Invalid ride data", nil)
	}

	// ✅ Check authorization
	var isRider bool
	var isDriver bool
	var driverProfile *models.DriverProfile
	var driverProfileID string

	// Check if user is the rider
	isRider = ride.RiderID == userID

	// Check if user is the driver
	// ✅ FIXED: ride.DriverID contains USER ID, not driver profile ID
	isDriver = ride.DriverID != nil && *ride.DriverID == userID

	// If user is the driver, get their driver profile ID
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

	// Check if ride can be cancelled
	// ✅ Prevent double cancellation and completed rides
	if ride.Status == "completed" {
		return response.BadRequest("Cannot cancel a completed ride")
	}
	if ride.Status == "cancelled" {
		return response.BadRequest("Ride was already cancelled")
	}

	// ✅ REMOVED: Now allowing cancellation of "started" rides

	// ✅ UPDATED: Dynamic cancellation fee based on ride status
	var riderCancellationFee float64
	var driverPenalty float64
	cancelledBy := "rider"
	originalStatus := ride.Status // Save original status for later use

	if isDriver {
		cancelledBy = "driver"
	}

	// ✅ Fee Structure based on ride status
	switch ride.Status {
	case "searching":
		// No fees during search
		riderCancellationFee = 0.0
		driverPenalty = 0.0

	case "accepted", "arrived":
		// Standard cancellation fees
		if isRider {
			riderCancellationFee = 2.0 // $2 for rider
		} else {
			driverPenalty = 3.0 // $3 penalty for driver cancelling after accepting
		}

	case "started":
		// ✅ NEW: Higher fees for ongoing rides
		if isRider {
			// Rider cancels ongoing ride - higher fee
			riderCancellationFee = 5.0 // $5 cancellation fee
			driverPenalty = 0.0        // Driver gets compensated from rider's fee
		} else {
			// Driver cancels ongoing ride - penalty + rider compensation
			driverPenalty = 10.0       // $10 penalty from driver
			riderCancellationFee = 0.0 // Rider gets refunded + compensation
		}
	}

	// Update ride
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

	// ✅ Get driver details for wallet operations
	var driver *models.DriverProfile
	if ride.DriverID != nil {
		// ✅ FIXED: ride.DriverID contains USER ID, not driver profile ID
		// Check if the driverProfile we already fetched is for this ride's driver
		if driverProfile != nil && driverProfile.UserID == *ride.DriverID {
			driver = driverProfile
		} else {
			// Need to fetch the driver by user ID
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

	// ✅ UPDATED: Wallet transaction processing based on scenario
	if ride.WalletHoldID != nil {
		switch {
		// Scenario 1: Rider cancels ongoing ride
		case ride.Status == "started" && isRider:
			// Capture full cancellation fee from rider's hold
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

				// Credit driver with cancellation fee
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

		// Scenario 2: Driver cancels ongoing ride
		case ride.Status == "started" && isDriver:
			// Release rider's hold completely
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

			// Debit driver's wallet as penalty
			if driver != nil {
				if _, err := s.walletService.DebitWallet(
					ctx,
					driver.UserID,
					driverPenalty,
					"cancellation_penalty",
					rideID,
					"Penalty for cancelling ongoing ride",
					nil,
				); err != nil {
					logger.Error("failed to debit driver penalty", "error", err, "driverID", driver.ID)
				} else {
					logger.Info("driver penalty applied",
						"rideID", rideID,
						"driverID", driver.ID,
						"penalty", driverPenalty,
					)

					// Credit rider with compensation from driver's penalty
					compensationAmount := driverPenalty * 0.5 // Give rider 50% of penalty
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

		// Scenario 3: Standard cancellation (accepted/arrived)
		case riderCancellationFee > 0:
			// Capture standard cancellation fee
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

				// Credit driver with cancellation fee
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

		// Scenario 4: Driver penalty (driver cancels accepted/arrived ride)
		case driverPenalty > 0 && driver != nil:
			// Release rider's hold
			if err := s.walletService.ReleaseHold(ctx, ride.RiderID, walletdto.ReleaseHoldRequest{
				HoldID: *ride.WalletHoldID,
			}); err != nil {
				logger.Error("failed to release hold", "error", err, "rideID", rideID)
			}

			// Debit driver's wallet as penalty
			if _, err := s.walletService.DebitWallet(
				ctx,
				driver.UserID,
				driverPenalty,
				"cancellation_penalty",
				rideID,
				"Penalty for cancelling accepted ride",
				nil,
			); err != nil {
				logger.Error("failed to debit driver penalty", "error", err, "driverID", driver.ID)
			} else {
				logger.Info("driver cancellation penalty applied",
					"rideID", rideID,
					"driverID", driver.ID,
					"penalty", driverPenalty,
				)
			}

		// Scenario 5: No fees (searching status)
		default:
			// Release hold completely
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

	// ✅ CLEANUP: Update driver status and clear caches
	if ride.DriverID != nil {
		driverUserID := *ride.DriverID

		// Get driver profile ID if not already fetched
		if driverProfile == nil {
			driverProfile, _ = s.driversRepo.FindDriverByUserID(ctx, driverUserID)
		}

		if driverProfile != nil && driverProfile.ID != "" {
			driverProfileID = driverProfile.ID // Use assignment, not declaration to avoid shadowing

			// Update driver status to online (safely ignore errors)
			if err := s.driversRepo.UpdateDriverStatus(ctx, driverProfileID, "online"); err != nil {
				logger.Warn("failed to update driver status", "error", err, "driverID", driverProfileID)
			}

			// Clear driver busy flag
			busyKey := fmt.Sprintf("driver:busy:%s", driverProfileID)
			if err := cache.Delete(ctx, busyKey); err != nil {
				logger.Warn("failed to clear driver busy cache", "error", err, "busyKey", busyKey)
			}

			// Clear active ride tracking cache
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

	// Notify driver (if driver didn't cancel)
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

	// Clear ride cache
	rideCacheKey := fmt.Sprintf("ride:active:%s", rideID)
	if err := cache.Delete(ctx, rideCacheKey); err != nil {
		logger.Warn("failed to clear ride cache", "error", err, "rideCacheKey", rideCacheKey)
	}

	// Notify rider (if rider didn't cancel)
	if isDriver {
		var message string
		var compensation float64

		// Use original status since ride.Status is now "cancelled"
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

// ========================================
// TriggerSOS - Allow rider to trigger SOS during emergency
// ========================================
func (s *service) TriggerSOS(ctx context.Context, riderID, rideID string, latitude, longitude float64) error {
	// Verify the ride exists and user is the rider
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return response.NotFoundError("Ride")
	}

	if ride.RiderID != riderID {
		return response.ForbiddenError("Not authorized to trigger SOS for this ride")
	}

	// Only allow SOS during active rides
	if ride.Status != "started" && ride.Status != "accepted" && ride.Status != "arrived" {
		return response.BadRequest("SOS can only be triggered during an active ride")
	}

	// Check if sosService is available
	if s.sosService == nil {
		logger.Warn("SOS service not initialized, logging emergency only",
			"riderID", riderID,
			"rideID", rideID,
			"location", map[string]float64{"lat": latitude, "lon": longitude},
		)
		// Still notify via WebSocket even if SOS service unavailable
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

	// Trigger SOS via the SOS service
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

	// Broadcast SOS alert to all connected users (safety team can monitor)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ✅ HELPER FUNCTION: Calculate distance between two coordinates
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Haversine formula
	const earthRadius = 6371.0 // kilometers

	dLat := (lat2 - lat1) * (3.14159265359 / 180.0)
	dLon := (lon2 - lon1) * (3.14159265359 / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*(3.14159265359/180.0))*math.Cos(lat2*(3.14159265359/180.0))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// func (s *service) CancelRide(ctx context.Context, userID, rideID string, req dto.CancelRideRequest) error {
// 	ride, err := s.repo.FindRideByID(ctx, rideID)
// 	if err != nil {
// 		return response.NotFoundError("Ride")
// 	}

// 	// ✅ FIX: Check authorization properly
// 	var isRider bool
// 	var isDriver bool
// 	var driverProfile *models.DriverProfile
// 	var driverProfileID string

// 	// Check if user is the rider (riderID is user ID)
// 	isRider = ride.RiderID == userID

// 	// ✅ Check if user is the driver (need to fetch driver profile first)
// 	if ride.DriverID != nil {
// 		// Try to fetch driver profile by user ID
// 		driverProfile, err = s.driversRepo.FindDriverByUserID(ctx, userID)
// 		if err == nil && driverProfile != nil {
// 			// User has a driver profile - check if it matches the ride's driver
// 			isDriver = driverProfile.ID == *ride.DriverID
// 			driverProfileID = driverProfile.ID
// 		}
// 	}

// 	if !isRider && !isDriver {
// 		logger.Warn("unauthorized cancellation attempt",
// 			"userID", userID,
// 			"rideID", rideID,
// 			"riderID", ride.RiderID,
// 			"rideDriverID", ride.DriverID,
// 		)
// 		return response.ForbiddenError("Not authorized to cancel this ride")
// 	}

// 	// Check if ride can be cancelled
// 	if ride.Status == "completed" || ride.Status == "cancelled" {
// 		return response.BadRequest("Cannot cancel a completed or already cancelled ride")
// 	}

// 	if ride.Status == "started" {
// 		return response.BadRequest("Cannot cancel an ongoing ride")
// 	}

// 	// Determine cancellation fee
// 	var cancellationFee float64
// 	cancelledBy := "rider"
// 	if isDriver {
// 		cancelledBy = "driver"
// 	}

// 	// Apply cancellation fee rules
// 	if ride.Status == "accepted" || ride.Status == "arrived" {
// 		if isRider {
// 			cancellationFee = 2.0 // $2 cancellation fee for rider
// 		}
// 	}

// 	// Update ride
// 	ride.Status = "cancelled"
// 	ride.CancellationReason = req.Reason
// 	ride.CancelledBy = &cancelledBy
// 	ride.CancelledAt = &time.Time{}
// 	*ride.CancelledAt = time.Now()

// 	if err := s.repo.UpdateRide(ctx, ride); err != nil {
// 		return response.InternalServerError("Failed to cancel ride", err)
// 	}

// 	logger.Info("ride cancellation initiated",
// 		"rideID", rideID,
// 		"cancelledBy", cancelledBy,
// 		"userID", userID,
// 		"isRider", isRider,
// 		"isDriver", isDriver,
// 		"driverProfileID", driverProfileID,
// 		"cancellationFee", cancellationFee,
// 	)

// 	// Process wallet transactions
// 	if ride.WalletHoldID != nil {
// 		if cancellationFee > 0 {
// 			// Capture cancellation fee, release rest
// 			if _, err := s.walletService.CaptureHold(ctx, ride.RiderID, walletdto.CaptureHoldRequest{
// 				HoldID:      *ride.WalletHoldID,
// 				Amount:      &cancellationFee,
// 				Description: "Cancellation fee",
// 			}); err != nil {
// 				logger.Error("failed to capture cancellation fee", "error", err, "rideID", rideID)
// 			} else {
// 				logger.Info("cancellation fee captured",
// 					"rideID", rideID,
// 					"amount", cancellationFee,
// 					"riderID", ride.RiderID,
// 				)
// 			}

// 			// ✅ Credit driver with cancellation fee if assigned
// 			if ride.DriverID != nil {
// 				// If we already have driver profile from authorization, use it
// 				// Otherwise, fetch it
// 				var driver *models.DriverProfile
// 				if driverProfile != nil && driverProfile.ID == *ride.DriverID {
// 					driver = driverProfile
// 				} else {
// 					driver, err = s.driversRepo.FindDriverByID(ctx, *ride.DriverID)
// 					if err != nil {
// 						logger.Error("failed to fetch driver for cancellation fee credit",
// 							"error", err,
// 							"driverID", *ride.DriverID,
// 							"rideID", rideID,
// 						)
// 					}
// 				}

// 				if driver != nil {
// 					// ✅ Use driver's user ID for wallet credit
// 					s.walletService.CreditWallet(
// 						ctx,
// 						driver.UserID, // ✅ Use user ID, not driver profile ID
// 						cancellationFee,
// 						"cancellation_fee",
// 						rideID,
// 						"Cancellation fee compensation",
// 						nil,
// 					)

// 					logger.Info("cancellation fee credited to driver",
// 						"rideID", rideID,
// 						"driverID", driver.ID,
// 						"driverUserID", driver.UserID,
// 						"driverName", driver.User.Name,
// 						"amount", cancellationFee,
// 					)
// 				}
// 			}
// 		} else {
// 			// Release hold completely
// 			if err := s.walletService.ReleaseHold(ctx, ride.RiderID, walletdto.ReleaseHoldRequest{
// 				HoldID: *ride.WalletHoldID,
// 			}); err != nil {
// 				logger.Error("failed to release hold", "error", err, "rideID", rideID)
// 			} else {
// 				logger.Info("wallet hold released",
// 					"rideID", rideID,
// 					"riderID", ride.RiderID,
// 				)
// 			}
// 		}
// 	}

// 	// ✅ CLEANUP: Update driver status and clear all caches if driver was assigned
// 	if ride.DriverID != nil {
// 		driverID := *ride.DriverID

// 		// Update driver status to online
// 		s.driversRepo.UpdateDriverStatus(ctx, driverID, "online")

// 		// ✅ Clear driver busy flag from cache
// 		busyKey := fmt.Sprintf("driver:busy:%s", driverID)
// 		cache.Delete(ctx, busyKey)

// 		// ✅ NEW: Clear active ride tracking cache
// 		activeRideCacheKey := fmt.Sprintf("driver:active:ride:%s", driverID)
// 		cache.Delete(ctx, activeRideCacheKey)

// 		logger.Debug("driver caches cleared",
// 			"driverID", driverID,
// 			"rideID", rideID,
// 			"clearedKeys", []string{busyKey, activeRideCacheKey},
// 		)

// 		// ✅ Get driver details for WebSocket notification
// 		// If we already have driver profile from authorization, use it
// 		// Otherwise, fetch it
// 		var driver *models.DriverProfile
// 		if driverProfile != nil && driverProfile.ID == driverID {
// 			driver = driverProfile
// 		} else {
// 			driver, err = s.driversRepo.FindDriverByID(ctx, driverID)
// 			if err != nil {
// 				logger.Error("failed to fetch driver for cancellation notification",
// 					"error", err,
// 					"driverID", driverID,
// 					"rideID", rideID,
// 				)
// 			}
// 		}

// 		if driver != nil {
// 			// ✅ Notify driver via WebSocket using user ID (only if driver didn't cancel)
// 			if !isDriver {
// 				websocketutil.SendToUser(driver.UserID, websocket.TypeRideCancelled, map[string]interface{}{
// 					"rideId":    rideID,
// 					"message":   "Ride was cancelled by rider",
// 					"reason":    req.Reason,
// 					"timestamp": time.Now().UTC(),
// 				})

// 				logger.Info("driver notified of cancellation",
// 					"rideID", rideID,
// 					"driverID", driver.ID,
// 					"driverUserID", driver.UserID,
// 					"driverName", driver.User.Name,
// 					"cancelledBy", "rider",
// 				)
// 			}
// 		}
// 	}

// 	// ✅ Clear ride cache
// 	rideCacheKey := fmt.Sprintf("ride:active:%s", rideID)
// 	cache.Delete(ctx, rideCacheKey)

// 	// ✅ Notify rider via WebSocket (ride.RiderID is user ID, so OK)
// 	if isDriver {
// 		websocketutil.SendToUser(ride.RiderID, websocket.TypeRideCancelled, map[string]interface{}{
// 			"rideId":    rideID,
// 			"message":   "Ride was cancelled by driver",
// 			"reason":    req.Reason,
// 			"timestamp": time.Now().UTC(),
// 		})

// 		logger.Info("rider notified of cancellation",
// 			"rideID", rideID,
// 			"riderID", ride.RiderID,
// 			"cancelledBy", "driver",
// 		)
// 	}

// 	logger.Info("ride cancelled successfully",
// 		"rideID", rideID,
// 		"cancelledBy", cancelledBy,
// 		"userID", userID,
// 		"cancellationFee", cancellationFee,
// 		"driverID", ride.DriverID,
// 		"status", "success",
// 	)

// 	return nil
// }
