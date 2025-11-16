package rides

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/umar5678/go-backend/internal/models"
	driversrepo "github.com/umar5678/go-backend/internal/modules/drivers"
	pricingservice "github.com/umar5678/go-backend/internal/modules/pricing"
	pricingdto "github.com/umar5678/go-backend/internal/modules/pricing/dto"
	ridersrepo "github.com/umar5678/go-backend/internal/modules/riders"
	"github.com/umar5678/go-backend/internal/modules/rides/dto"
	trackingservice "github.com/umar5678/go-backend/internal/modules/tracking"
	trackingdto "github.com/umar5678/go-backend/internal/modules/tracking/dto"
	walletservice "github.com/umar5678/go-backend/internal/modules/wallet"
	walletdto "github.com/umar5678/go-backend/internal/modules/wallet/dto"
	"github.com/umar5678/go-backend/internal/services/cache"
	"github.com/umar5678/go-backend/internal/utils/helpers"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
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
	StartRide(ctx context.Context, driverID, rideID string) (*dto.RideResponse, error)
	CompleteRide(ctx context.Context, driverID, rideID string, req dto.CompleteRideRequest) (*dto.RideResponse, error)

	// Internal
	FindDriverForRide(ctx context.Context, rideID string) error
}

type service struct {
	repo            Repository
	driversRepo     driversrepo.Repository
	ridersRepo      ridersrepo.Repository
	pricingService  pricingservice.Service
	trackingService trackingservice.Service
	walletService   walletservice.Service
}

func NewService(
	repo Repository,
	driversRepo driversrepo.Repository,
	ridersRepo ridersrepo.Repository,
	pricingService pricingservice.Service,
	trackingService trackingservice.Service,
	walletService walletservice.Service,
) Service {
	return &service{
		repo:            repo,
		driversRepo:     driversRepo,
		ridersRepo:      ridersRepo,
		pricingService:  pricingService,
		trackingService: trackingService,
		walletService:   walletService,
	}
}

func (s *service) CreateRide(ctx context.Context, riderID string, req dto.CreateRideRequest) (*dto.RideResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
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

	// 2. Hold funds from rider's wallet
	holdReq := walletdto.HoldFundsRequest{
		Amount:        fareEstimate.TotalFare,
		ReferenceType: "ride",
		HoldDuration:  30, // 30 minutes
	}

	holdResp, err := s.walletService.HoldFunds(ctx, riderID, holdReq)
	if err != nil {
		return nil, response.BadRequest("Insufficient wallet balance. Please add funds.")
	}

	// 3. Create ride
	rideID := uuid.New().String()
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
		EstimatedFare:     fareEstimate.TotalFare,
		SurgeMultiplier:   fareEstimate.SurgeMultiplier,
		WalletHoldID:      &holdResp.ID,
		RiderNotes:        req.RiderNotes,
		RequestedAt:       time.Now(),
	}

	// Update hold reference
	holdResp.ReferenceID = rideID

	if err := s.repo.CreateRide(ctx, ride); err != nil {
		// Release hold if ride creation fails
		s.walletService.ReleaseHold(ctx, riderID, walletdto.ReleaseHoldRequest{HoldID: holdResp.ID})
		logger.Error("failed to create ride", "error", err, "riderID", riderID)
		return nil, response.InternalServerError("Failed to create ride", err)
	}

	// 4. Store in Redis for quick access
	cacheKey := fmt.Sprintf("ride:active:%s", rideID)
	cache.SetJSON(ctx, cacheKey, ride, 30*time.Minute)

	// 5. Start finding driver asynchronously
	go func() {
		bgCtx := context.Background()
		if err := s.FindDriverForRide(bgCtx, rideID); err != nil {
			logger.Error("failed to find driver", "error", err, "rideID", rideID)
			// Handle no drivers found - cancel ride and refund
		}
	}()

	// 6. Notify rider
	helpers.SendNotification(riderID, map[string]interface{}{
		"type":    "ride_searching",
		"rideId":  rideID,
		"message": "Searching for nearby drivers...",
	})

	logger.Info("ride created",
		"rideID", rideID,
		"riderID", riderID,
		"estimatedFare", fareEstimate.TotalFare,
		"surge", fareEstimate.SurgeMultiplier,
	)

	// Fetch with relations
	ride, _ = s.repo.FindRideByID(ctx, rideID)

	return dto.ToRideResponse(ride), nil
}

func (s *service) FindDriverForRide(ctx context.Context, rideID string) error {
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return err
	}

	if ride.Status != "searching" {
		return nil // Already assigned
	}

	// Find nearby drivers
	nearbyReq := trackingdto.FindNearbyDriversRequest{
		Latitude:      ride.PickupLat,
		Longitude:     ride.PickupLon,
		RadiusKm:      5.0, // Start with 5km
		VehicleTypeID: ride.VehicleTypeID,
		Limit:         10,
	}

	nearbyDrivers, err := s.trackingService.FindNearbyDrivers(ctx, nearbyReq)
	if err != nil || nearbyDrivers.Count == 0 {
		// Try expanding radius
		nearbyReq.RadiusKm = 10.0
		nearbyDrivers, err = s.trackingService.FindNearbyDrivers(ctx, nearbyReq)

		if err != nil || nearbyDrivers.Count == 0 {
			// No drivers found
			helpers.SendNotification(ride.RiderID, map[string]interface{}{
				"type":    "no_drivers_available",
				"rideId":  rideID,
				"message": "No drivers available in your area. Please try again later.",
			})

			// Auto-cancel and refund
			go func() {
				time.Sleep(2 * time.Minute)
				s.CancelRide(context.Background(), ride.RiderID, rideID, dto.CancelRideRequest{
					Reason: "No drivers available",
				})
			}()

			return errors.New("no drivers available")
		}
	}

	// Send requests to drivers (one at a time, max 3 drivers)
	maxAttempts := 3
	if nearbyDrivers.Count < maxAttempts {
		maxAttempts = nearbyDrivers.Count
	}

	for i := 0; i < maxAttempts; i++ {
		driver := nearbyDrivers.Drivers[i]

		// Create ride request
		requestID := uuid.New().String()
		rideRequest := &models.RideRequest{
			ID:        requestID,
			RideID:    rideID,
			DriverID:  driver.DriverID,
			Status:    "pending",
			SentAt:    time.Now(),
			ExpiresAt: time.Now().Add(30 * time.Second),
		}

		if err := s.repo.CreateRideRequest(ctx, rideRequest); err != nil {
			logger.Error("failed to create ride request", "error", err)
			continue
		}

		// Send notification to driver
		helpers.SendNotification(driver.DriverID, map[string]interface{}{
			"type":      "ride_request",
			"rideId":    rideID,
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
			"expiresIn":     30,
		})

		logger.Info("ride request sent to driver",
			"rideID", rideID,
			"driverID", driver.DriverID,
			"distance", driver.Distance,
		)

		// Wait for response (30 seconds)
		time.Sleep(30 * time.Second)

		// Check if ride was accepted
		ride, _ = s.repo.FindRideByID(ctx, rideID)
		if ride.Status == "accepted" {
			return nil // Driver accepted
		}

		// Mark request as expired
		s.repo.UpdateRideRequestStatus(ctx, requestID, "expired")
	}

	// All drivers rejected/timeout
	helpers.SendNotification(ride.RiderID, map[string]interface{}{
		"type":    "drivers_not_available",
		"rideId":  rideID,
		"message": "Drivers are busy right now. Please try again.",
	})

	return errors.New("no driver accepted")
}

func (s *service) AcceptRide(ctx context.Context, driverID, rideID string) (*dto.RideResponse, error) {
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return nil, response.NotFoundError("Ride")
	}

	if ride.Status != "searching" {
		return nil, response.BadRequest("Ride is no longer available")
	}

	// Check if driver is online
	driver, err := s.driversRepo.FindDriverByUserID(ctx, driverID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	if driver.Status != "online" {
		return nil, response.BadRequest("Driver must be online to accept rides")
	}

	// Update ride
	ride.DriverID = &driver.ID
	ride.Status = "accepted"
	ride.AcceptedAt = &time.Time{}
	*ride.AcceptedAt = time.Now()

	if err := s.repo.UpdateRide(ctx, ride); err != nil {
		return nil, response.InternalServerError("Failed to accept ride", err)
	}

	// Update driver status
	s.driversRepo.UpdateDriverStatus(ctx, driver.ID, "busy")

	// Update cache
	cache.Delete(ctx, fmt.Sprintf("ride:active:%s", rideID))
	cache.SetJSON(ctx, fmt.Sprintf("ride:active:%s", rideID), ride, 30*time.Minute)

	// Notify rider
	helpers.SendNotification(ride.RiderID, map[string]interface{}{
		"type":     "ride_accepted",
		"rideId":   rideID,
		"driverId": driver.ID,
		"driver": map[string]interface{}{
			"name":   driver.User.Name,
			"phone":  driver.User.Phone,
			"rating": driver.Rating,
		},
		"vehicle": map[string]interface{}{
			// Add vehicle details
		},
		"message": "Driver is on the way!",
	})

	logger.Info("ride accepted",
		"rideID", rideID,
		"driverID", driver.ID,
	)

	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
}

func (s *service) RejectRide(ctx context.Context, driverID, rideID string, req dto.RejectRideRequest) error {
	// Find pending request for this driver
	requests, err := s.repo.FindPendingRequestsForDriver(ctx, driverID)
	if err != nil {
		return err
	}

	var requestID string
	for _, r := range requests {
		if r.RideID == rideID {
			requestID = r.ID
			break
		}
	}

	if requestID == "" {
		return response.NotFoundError("Ride request")
	}

	// Update request status
	if err := s.repo.UpdateRideRequestStatus(ctx, requestID, "rejected"); err != nil {
		return response.InternalServerError("Failed to reject ride", err)
	}

	logger.Info("ride rejected by driver",
		"rideID", rideID,
		"driverID", driverID,
		"reason", req.Reason,
	)

	return nil
}

func (s *service) MarkArrived(ctx context.Context, driverID, rideID string) (*dto.RideResponse, error) {
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return nil, response.NotFoundError("Ride")
	}

	if ride.DriverID == nil || *ride.DriverID != driverID {
		return nil, response.ForbiddenError("Not authorized")
	}

	if ride.Status != "accepted" {
		return nil, response.BadRequest("Invalid ride status")
	}

	// Update status
	if err := s.repo.UpdateRideStatus(ctx, rideID, "arrived"); err != nil {
		return nil, response.InternalServerError("Failed to update status", err)
	}

	// Notify rider
	helpers.SendNotification(ride.RiderID, map[string]interface{}{
		"type":    "driver_arrived",
		"rideId":  rideID,
		"message": "Your driver has arrived",
	})

	logger.Info("driver arrived at pickup", "rideID", rideID, "driverID", driverID)

	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
}

func (s *service) StartRide(ctx context.Context, driverID, rideID string) (*dto.RideResponse, error) {
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return nil, response.NotFoundError("Ride")
	}

	if ride.DriverID == nil || *ride.DriverID != driverID {
		return nil, response.ForbiddenError("Not authorized")
	}

	if ride.Status != "arrived" {
		return nil, response.BadRequest("Driver must arrive at pickup first")
	}

	// Update status
	if err := s.repo.UpdateRideStatus(ctx, rideID, "started"); err != nil {
		return nil, response.InternalServerError("Failed to start ride", err)
	}

	// Update driver status
	s.driversRepo.UpdateDriverStatus(ctx, *ride.DriverID, "on_trip")

	// Notify both parties
	helpers.SendNotification(ride.RiderID, map[string]interface{}{
		"type":    "ride_started",
		"rideId":  rideID,
		"message": "Your ride has started",
	})

	logger.Info("ride started", "rideID", rideID, "driverID", driverID)

	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
}

func (s *service) CompleteRide(ctx context.Context, driverID, rideID string, req dto.CompleteRideRequest) (*dto.RideResponse, error) {
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return nil, response.NotFoundError("Ride")
	}

	if ride.DriverID == nil || *ride.DriverID != driverID {
		return nil, response.ForbiddenError("Not authorized")
	}

	if ride.Status != "started" {
		return nil, response.BadRequest("Ride must be started first")
	}

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

	// Update ride
	ride.ActualDistance = &req.ActualDistance
	ride.ActualDuration = &req.ActualDuration
	ride.ActualFare = &actualFareResp.TotalFare
	ride.Status = "completed"
	ride.CompletedAt = &time.Time{}
	*ride.CompletedAt = time.Now()

	if err := s.repo.UpdateRide(ctx, ride); err != nil {
		return nil, response.InternalServerError("Failed to complete ride", err)
	}

	// Capture hold
	if ride.WalletHoldID != nil {
		captureReq := walletdto.CaptureHoldRequest{
			HoldID:      *ride.WalletHoldID,
			Amount:      &actualFareResp.TotalFare,
			Description: fmt.Sprintf("Payment for ride %s", rideID),
		}

		if _, err := s.walletService.CaptureHold(ctx, ride.RiderID, captureReq); err != nil {
			logger.Error("failed to capture hold", "error", err, "rideID", rideID)
		}
	}

	// Credit driver (80% commission)
	driverEarnings := actualFareResp.TotalFare * 0.80
	s.walletService.CreditWallet(
		ctx,
		driverID,
		driverEarnings,
		"ride",
		rideID,
		fmt.Sprintf("Earnings from ride %s", rideID),
		map[string]interface{}{
			"totalFare":  actualFareResp.TotalFare,
			"commission": 0.20,
		},
	)

	// Update stats
	s.driversRepo.IncrementTrips(ctx, *ride.DriverID)
	s.driversRepo.UpdateEarnings(ctx, *ride.DriverID, driverEarnings)
	s.ridersRepo.IncrementTotalRides(ctx, ride.RiderID)

	// Update driver status back to online
	s.driversRepo.UpdateDriverStatus(ctx, *ride.DriverID, "online")

	// Clear cache
	cache.Delete(ctx, fmt.Sprintf("ride:active:%s", rideID))

	// Notify both parties
	helpers.SendNotification(ride.RiderID, map[string]interface{}{
		"type":       "ride_completed",
		"rideId":     rideID,
		"actualFare": actualFareResp.TotalFare,
		"message":    "Your ride is complete",
	})

	helpers.SendNotification(driverID, map[string]interface{}{
		"type":     "ride_completed",
		"rideId":   rideID,
		"earnings": driverEarnings,
		"message":  "Ride completed successfully",
	})

	logger.Info("ride completed",
		"rideID", rideID,
		"actualFare", actualFareResp.TotalFare,
		"driverEarnings", driverEarnings,
	)

	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
}

func (s *service) GetRide(ctx context.Context, userID, rideID string) (*dto.RideResponse, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("ride:active:%s", rideID)
	var cached models.Ride

	err := cache.GetJSON(ctx, cacheKey, &cached)
	if err == nil {
		// Verify user has access
		if cached.RiderID == userID || (cached.DriverID != nil && *cached.DriverID == userID) {
			return dto.ToRideResponse(&cached), nil
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

	return dto.ToRideResponse(ride), nil
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
		return response.NotFoundError("Ride")
	}

	// Check authorization
	isRider := ride.RiderID == userID
	isDriver := ride.DriverID != nil && *ride.DriverID == userID

	if !isRider && !isDriver {
		return response.ForbiddenError("Not authorized to cancel this ride")
	}

	// Check if ride can be cancelled
	if ride.Status == "completed" || ride.Status == "cancelled" {
		return response.BadRequest("Cannot cancel a completed or already cancelled ride")
	}

	if ride.Status == "started" {
		return response.BadRequest("Cannot cancel an ongoing ride")
	}

	// Determine cancellation fee
	var cancellationFee float64
	cancelledBy := "rider"
	if isDriver {
		cancelledBy = "driver"
	}

	// Apply cancellation fee rules
	if ride.Status == "accepted" || ride.Status == "arrived" {
		if isRider {
			cancellationFee = 2.0 // $2 cancellation fee for rider
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

	// Process wallet transactions
	if ride.WalletHoldID != nil {
		if cancellationFee > 0 {
			// Capture cancellation fee, release rest
			if _, err := s.walletService.CaptureHold(ctx, ride.RiderID, walletdto.CaptureHoldRequest{
				HoldID:      *ride.WalletHoldID,
				Amount:      &cancellationFee,
				Description: "Cancellation fee",
			}); err != nil {
				logger.Error("failed to capture cancellation fee", "error", err)
			}

			// Credit driver with cancellation fee if assigned
			if ride.DriverID != nil {
				s.walletService.CreditWallet(
					ctx,
					*ride.DriverID,
					cancellationFee,
					"cancellation_fee",
					rideID,
					"Cancellation fee compensation",
					nil,
				)
			}
		} else {
			// Release hold completely
			if err := s.walletService.ReleaseHold(ctx, ride.RiderID, walletdto.ReleaseHoldRequest{
				HoldID: *ride.WalletHoldID,
			}); err != nil {
				logger.Error("failed to release hold", "error", err)
			}
		}
	}

	// Update driver status if assigned
	if ride.DriverID != nil {
		s.driversRepo.UpdateDriverStatus(ctx, *ride.DriverID, "online")

		// Notify driver
		helpers.SendNotification(*ride.DriverID, map[string]interface{}{
			"type":    "ride_cancelled",
			"rideId":  rideID,
			"message": "Ride was cancelled by rider",
		})
	}

	// Clear cache
	cache.Delete(ctx, fmt.Sprintf("ride:active:%s", rideID))

	// Notify rider
	if isDriver {
		helpers.SendNotification(ride.RiderID, map[string]interface{}{
			"type":    "ride_cancelled",
			"rideId":  rideID,
			"message": "Ride was cancelled by driver",
		})
	}

	logger.Info("ride cancelled",
		"rideID", rideID,
		"cancelledBy", cancelledBy,
		"cancellationFee", cancellationFee,
	)

	return nil
}
