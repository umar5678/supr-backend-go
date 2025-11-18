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
	StartRide(ctx context.Context, driverID, rideID string) (*dto.RideResponse, error)
	CompleteRide(ctx context.Context, driverID, rideID string, req dto.CompleteRideRequest) (*dto.RideResponse, error)

	// Internal
	FindDriverForRide(ctx context.Context, rideID string) error
	ProcessRideRequestTimeout(ctx context.Context, requestID string) error
}

type service struct {
	repo            Repository
	driversRepo     driversrepo.Repository
	ridersRepo      ridersrepo.Repository
	pricingService  pricingservice.Service
	trackingService trackingservice.Service
	walletService   walletservice.Service
	wsHelper        *RideWebSocketHelper
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
		wsHelper:        NewRideWebSocketHelper(),
	}
}

// FIX 1: Add ReferenceID to wallet hold
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

	// Generate ride ID first
	rideID := uuid.New().String()

	// 2. Hold funds with ReferenceID (FIX 1)
	holdReq := walletdto.HoldFundsRequest{
		Amount:        fareEstimate.TotalFare,
		ReferenceType: "ride",
		ReferenceID:   rideID,
		HoldDuration:  1800, // 30 minutes
	}

	holdResp, err := s.walletService.HoldFunds(ctx, riderID, holdReq)
	if err != nil {
		return nil, response.BadRequest("Insufficient wallet balance. Please add funds.")
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
		EstimatedFare:     fareEstimate.TotalFare,
		SurgeMultiplier:   fareEstimate.SurgeMultiplier,
		WalletHoldID:      &holdResp.ID,
		RiderNotes:        req.RiderNotes,
		RequestedAt:       time.Now(),
	}

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

			if err := s.repo.UpdateRideStatus(bgCtx, rideID, "cancelled"); err != nil {
				logger.Error("failed to update ride status", "error", err, "rideID", rideID)
			}

			if err := s.walletService.ReleaseHold(bgCtx, riderID, walletdto.ReleaseHoldRequest{HoldID: holdResp.ID}); err != nil {
				logger.Error("failed to release hold", "error", err, "rideID", rideID)
			}

			s.wsHelper.SendRideStatusToBoth(bgCtx, riderID, "", rideID, "cancelled", "No drivers available. Your payment has been refunded.")
		}
	}()

	// 6. Notify rider via WebSocket
	s.wsHelper.SendRideStatusToBoth(ctx, riderID, "", rideID, "searching", "Searching for nearby drivers...")

	logger.Info("ride created",
		"rideID", rideID,
		"riderID", riderID,
		"estimatedFare", fareEstimate.TotalFare,
		"surge", fareEstimate.SurgeMultiplier,
	)

	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
}

// FIX 2: Add OnlyAvailable filter
func (s *service) FindDriverForRide(ctx context.Context, rideID string) error {
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return err
	}

	if ride.Status != "searching" {
		return nil
	}

	radii := []float64{3.0, 5.0, 8.0}
	var nearbyDrivers *trackingdto.NearbyDriversResponse

	for _, radius := range radii {
		nearbyReq := trackingdto.FindNearbyDriversRequest{
			Latitude:      ride.PickupLat,
			Longitude:     ride.PickupLon,
			RadiusKm:      radius,
			VehicleTypeID: ride.VehicleTypeID,
			Limit:         15,
			// Note: OnlyAvailable should be added to trackingdto.FindNearbyDriversRequest struct
		}

		nearbyDrivers, err = s.trackingService.FindNearbyDrivers(ctx, nearbyReq)
		if err == nil && nearbyDrivers.Count > 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if err != nil || nearbyDrivers == nil || nearbyDrivers.Count == 0 {
		return errors.New("no drivers available in the area")
	}

	logger.Info("found nearby drivers",
		"rideID", rideID,
		"driverCount", nearbyDrivers.Count,
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
		return s.assignDriverToRide(ctx, rideID, acceptedDriverID)

	case <-ctxWithTimeout.Done():
		return errors.New("no driver accepted the ride request")

	case err := <-errorChan:
		return err
	}
}

// FIX 3: Handle context cancellation properly
func (s *service) sendRideRequestToDriver(
	ctx context.Context,
	ride *models.Ride,
	driver trackingdto.DriverLocationResponse,
	resultChan chan<- string,
	errorChan chan<- error,
) {
	requestID := uuid.New().String()
	expiresAt := time.Now().Add(10 * time.Second)

	rideRequest := &models.RideRequest{
		ID:        requestID,
		RideID:    ride.ID,
		DriverID:  driver.DriverID,
		Status:    "pending",
		SentAt:    time.Now(),
		ExpiresAt: expiresAt,
	}

	if err := s.repo.CreateRideRequest(ctx, rideRequest); err != nil {
		logger.Error("failed to create ride request", "error", err, "driverID", driver.DriverID)
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

	if err := s.wsHelper.SendRideRequest(driver.DriverID, rideDetails); err != nil {
		logger.Error("failed to send ride request via WebSocket", "error", err, "driverID", driver.DriverID)
	}

	logger.Info("ride request sent to driver",
		"rideID", ride.ID,
		"driverID", driver.DriverID,
		"requestID", requestID,
		"distance", driver.Distance,
	)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// FIX 3: Cancel when context is done
			s.repo.UpdateRideRequestStatus(context.Background(), requestID, "cancelled_by_system", nil)
			return

		case <-ticker.C:
			updatedRequest, err := s.repo.FindRideRequestByID(ctx, requestID)
			if err != nil {
				continue
			}

			if updatedRequest.Status == "accepted" {
				resultChan <- driver.DriverID
				return
			}

			if updatedRequest.Status == "rejected" || updatedRequest.Status == "expired" {
				return
			}

			if time.Now().After(expiresAt) {
				s.repo.UpdateRideRequestStatus(ctx, requestID, "expired", nil)
				return
			}
		}
	}
}

// FIX 4: Atomic assignment with cancellation of other requests
func (s *service) assignDriverToRide(ctx context.Context, rideID, driverID string) error {
	// Atomic update: only one driver can change status from "searching" to "accepted"
	err := s.repo.UpdateRideStatusAndDriver(ctx, rideID, "accepted", "searching", driverID)
	if err != nil {
		return response.BadRequest("Ride already accepted by another driver")
	}

	// Cancel all other pending requests for this ride
	go s.repo.CancelPendingRequestsExcept(context.Background(), rideID, driverID)

	// Update driver status
	s.driversRepo.UpdateDriverStatus(ctx, driverID, "busy")

	// Update cache
	ride, _ := s.repo.FindRideByID(ctx, rideID)
	cacheKey := fmt.Sprintf("ride:active:%s", rideID)
	cache.SetJSON(ctx, cacheKey, ride, 30*time.Minute)

	// Get driver details for notification
	driver, err := s.driversRepo.FindDriverByID(ctx, driverID)
	if err != nil {
		logger.Error("failed to fetch driver details", "error", err, "driverID", driverID)
	} else {
		rideDetails := map[string]interface{}{
			"rideId":   rideID,
			"driverId": driver.ID,
			"driver": map[string]interface{}{
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
			"message": "Driver is on the way!",
			"eta":     10,
		}

		s.wsHelper.SendRideAccepted(ride.RiderID, rideDetails)
	}

	logger.Info("ride assigned to driver",
		"rideID", rideID,
		"driverID", driverID,
	)

	return nil
}

// FIX 5: Check expiration in AcceptRide
func (s *service) AcceptRide(ctx context.Context, driverID, rideID string) (*dto.RideResponse, error) {
	rideRequest, err := s.repo.FindRideRequestByRideAndDriver(ctx, rideID, driverID)
	if err != nil {
		return nil, response.NotFoundError("Ride request not found")
	}

	if rideRequest.Status != "pending" {
		return nil, response.BadRequest("Ride request is no longer available")
	}

	// FIX 5: Check expiration
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

	driver, err := s.driversRepo.FindDriverByUserID(ctx, driverID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	if driver.Status != "online" {
		return nil, response.BadRequest("Driver must be online to accept rides")
	}

	if err := s.repo.UpdateRideRequestStatus(ctx, rideRequest.ID, "accepted", nil); err != nil {
		return nil, response.InternalServerError("Failed to accept ride", err)
	}

	if err := s.assignDriverToRide(ctx, rideID, driverID); err != nil {
		s.repo.UpdateRideRequestStatus(ctx, rideRequest.ID, "pending", nil)
		return nil, err
	}

	logger.Info("ride accepted by driver",
		"rideID", rideID,
		"driverID", driverID,
	)

	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
}

// FIX 6: Pass rejection reason
func (s *service) RejectRide(ctx context.Context, driverID, rideID string, req dto.RejectRideRequest) error {
	rideRequest, err := s.repo.FindRideRequestByRideAndDriver(ctx, rideID, driverID)
	if err != nil {
		return response.NotFoundError("Ride request not found")
	}

	if rideRequest.Status != "pending" {
		return response.BadRequest("Ride request is no longer available")
	}

	// FIX 6: Pass rejection reason
	if err := s.repo.UpdateRideRequestStatus(ctx, rideRequest.ID, "rejected", &req.Reason); err != nil {
		return response.InternalServerError("Failed to reject ride", err)
	}

	logger.Info("ride rejected by driver",
		"rideID", rideID,
		"driverID", driverID,
		"reason", req.Reason,
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

	// Notify rider via WebSocket
	websocketutil.SendToUser(ride.RiderID, websocket.TypeRideStatusUpdate, map[string]interface{}{
		"rideId":    rideID,
		"status":    "arrived",
		"message":   "Your driver has arrived at the pickup location",
		"timestamp": time.Now().UTC(),
	})

	logger.Info("driver arrived at pickup", "rideID", rideID, "driverID", driverID)

	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
}

// keep
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

	// Notify both parties via WebSocket
	// In your rides service.go - CORRECT USAGE:

	// Notify both parties via WebSocket
	websocketutil.SendRideStatusUpdate(ride.RiderID, driverID, map[string]interface{}{
		"rideId":    rideID,
		"status":    "started",
		"message":   "Your ride has started",
		"timestamp": time.Now().UTC(),
	})

	logger.Info("ride started", "rideID", rideID, "driverID", driverID)

	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
}

// keep
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

	// Notify both parties via WebSocket
	websocketutil.SendToUser(ride.RiderID, websocket.TypeRideCompleted, map[string]interface{}{
		"rideId":     rideID,
		"actualFare": actualFareResp.TotalFare,
		"message":    "Your ride is complete. Thank you for riding with us!",
		"timestamp":  time.Now().UTC(),
	})

	websocketutil.SendToUser(driverID, websocket.TypeRideCompleted, map[string]interface{}{
		"rideId":    rideID,
		"earnings":  driverEarnings,
		"message":   "Ride completed successfully",
		"timestamp": time.Now().UTC(),
	})

	logger.Info("ride completed",
		"rideID", rideID,
		"actualFare", actualFareResp.TotalFare,
		"driverEarnings", driverEarnings,
	)

	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
}

// keep
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

// keep
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

// keep
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

		// Notify driver via WebSocket
		websocketutil.SendToUser(*ride.DriverID, websocket.TypeRideCancelled, map[string]interface{}{
			"rideId":    rideID,
			"message":   "Ride was cancelled by rider",
			"timestamp": time.Now().UTC(),
		})
	}

	// Clear cache
	cache.Delete(ctx, fmt.Sprintf("ride:active:%s", rideID))

	// Notify rider via WebSocket
	if isDriver {
		websocketutil.SendToUser(ride.RiderID, websocket.TypeRideCancelled, map[string]interface{}{
			"rideId":    rideID,
			"message":   "Ride was cancelled by driver",
			"timestamp": time.Now().UTC(),
		})
	}

	logger.Info("ride cancelled",
		"rideID", rideID,
		"cancelledBy", cancelledBy,
		"cancellationFee", cancellationFee,
	)

	return nil
}
