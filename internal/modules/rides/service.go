package rides

import (
	"context"
	"errors"
	"fmt"
	"math"
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

	rideID := uuid.New().String()

	// 2. Hold funds with ReferenceID
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

			// ✅ CRITICAL FIX: Check ride status before canceling
			currentRide, statusErr := s.repo.FindRideByID(bgCtx, rideID)
			if statusErr != nil {
				logger.Error("failed to fetch ride status", "error", statusErr, "rideID", rideID)
				return
			}

			// ✅ If ride is already accepted/arrived/started, don't cancel
			if currentRide.Status != "searching" {
				logger.Info("ride already accepted by driver, not canceling",
					"rideID", rideID,
					"currentStatus", currentRide.Status,
					"driverID", currentRide.DriverID,
				)
				return
			}

			// Only cancel if still searching
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
			OnlyAvailable: true,
		}

		nearbyDrivers, err = s.trackingService.FindNearbyDrivers(ctx, nearbyReq)
		if err == nil && nearbyDrivers.Count > 0 {
			logger.Info("found nearby drivers at radius",
				"radius", radius,
				"count", nearbyDrivers.Count,
				"rideID", rideID)
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

// ✅ FIXED: Accept userID and fetch driver profile first
func (s *service) AcceptRide(ctx context.Context, userID, rideID string) (*dto.RideResponse, error) {
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

	logger.Info("driver attempting to accept ride",
		"userID", userID,
		"driverID", driverID,
		"driverName", driver.User.Name,
		"rideID", rideID,
	)

	// ✅ STEP 2: Find ride request using driver profile ID
	rideRequest, err := s.repo.FindRideRequestByRideAndDriver(ctx, rideID, driverID)
	if err != nil {
		logger.Error("ride request not found",
			"error", err,
			"rideID", rideID,
			"driverID", driverID,
			"userID", userID,
		)
		return nil, response.NotFoundError("Ride request not found")
	}

	if rideRequest.Status != "pending" {
		return nil, response.BadRequest("Ride request is no longer available")
	}

	if time.Now().After(rideRequest.ExpiresAt) {
		return nil, response.BadRequest("Ride request has expired")
	}

	// ✅ STEP 3: Validate ride status
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return nil, response.NotFoundError("Ride")
	}

	if ride.Status != "searching" {
		return nil, response.BadRequest("Ride is no longer available")
	}

	// ✅ STEP 4: Validate driver status
	if driver.Status != "online" {
		return nil, response.BadRequest("Driver must be online to accept rides")
	}

	logger.Info("driver attempting to accept ride request",
		"rideID", rideID,
		"driverID", driverID,
		"userID", userID,
		"driverName", driver.User.Name,
		"requestID", rideRequest.ID,
		"riderID", ride.RiderID,
	)

	// ✅ STEP 5: Update ride request status
	if err := s.repo.UpdateRideRequestStatus(ctx, rideRequest.ID, "accepted", nil); err != nil {
		logger.Error("failed to update ride request status",
			"error", err,
			"requestID", rideRequest.ID,
			"driverID", driverID,
			"userID", userID,
			"driverName", driver.User.Name,
		)
		return nil, response.InternalServerError("Failed to accept ride", err)
	}

	// ✅ STEP 6: Assign driver to ride
	if err := s.assignDriverToRide(ctx, rideID, driverID); err != nil {
		s.repo.UpdateRideRequestStatus(ctx, rideRequest.ID, "pending", nil)
		logger.Error("failed to assign driver to ride",
			"error", err,
			"rideID", rideID,
			"driverID", driverID,
			"userID", userID,
			"driverName", driver.User.Name,
		)
		return nil, err
	}

	logger.Info("ride successfully accepted by driver",
		"rideID", rideID,
		"driverID", driverID,
		"userID", userID,
		"driverName", driver.User.Name,
		"driverPhone", driver.User.Phone,
		"vehicleType", driver.Vehicle.VehicleType.Name,
		"vehiclePlate", driver.Vehicle.LicensePlate,
		"riderID", ride.RiderID,
	)

	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
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

	if ride.DriverID == nil || *ride.DriverID != driverID {
		logger.Warn("unauthorized attempt to mark arrived",
			"userID", userID,
			"driverID", driverID,
			"rideDriverID", ride.DriverID,
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
	websocketutil.SendToUser(ride.RiderID, websocket.TypeRideStatusUpdate, map[string]interface{}{
		"rideId":    rideID,
		"status":    "arrived",
		"message":   "Your driver has arrived at the pickup location",
		"timestamp": time.Now().UTC(),
	})

	logger.Info("driver arrived at pickup location",
		"rideID", rideID,
		"driverID", driverID,
		"userID", userID,
		"driverName", driver.User.Name,
		"riderID", ride.RiderID,
	)

	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
}

// ✅ FIXED: Accept userID and fetch driver profile first
func (s *service) StartRide(ctx context.Context, userID, rideID string) (*dto.RideResponse, error) {
	// ✅ STEP 1: Fetch driver profile using user ID
	driver, err := s.driversRepo.FindDriverByUserID(ctx, userID)
	if err != nil {
		logger.Error("driver profile not found for user",
			"error", err,
			"userID", userID,
		)
		return nil, response.NotFoundError("Driver profile not found")
	}

	driverID := driver.ID         // ✅ Driver profile ID
	driverUserID := driver.UserID // ✅ Driver user ID for WebSocket

	// ✅ STEP 2: Fetch ride and validate authorization
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return nil, response.NotFoundError("Ride")
	}

	if ride.DriverID == nil || *ride.DriverID != driverID {
		logger.Warn("unauthorized attempt to start ride",
			"userID", userID,
			"driverID", driverID,
			"rideDriverID", ride.DriverID,
			"rideID", rideID,
		)
		return nil, response.ForbiddenError("Not authorized")
	}

	// ✅ FIXED: Allow both "accepted" and "arrived" status
	if ride.Status != "accepted" && ride.Status != "arrived" {
		return nil, response.BadRequest("Ride must be accepted or arrived to start")
	}

	// ✅ FIXED: Update ride status AND set StartedAt timestamp
	ride.Status = "started"
	startedAt := time.Now()
	ride.StartedAt = &startedAt

	if err := s.repo.UpdateRide(ctx, ride); err != nil {
		return nil, response.InternalServerError("Failed to start ride", err)
	}

	// ✅ Update driver status to on_trip
	s.driversRepo.UpdateDriverStatus(ctx, driverID, "on_trip")

	// ✅ NEW: Update ride cache
	cacheKey := fmt.Sprintf("ride:active:%s", rideID)
	cache.SetJSON(ctx, cacheKey, ride, 30*time.Minute)

	// ✅ Notify both parties via WebSocket using user IDs
	websocketutil.SendRideStatusUpdate(ride.RiderID, driverUserID, map[string]interface{}{
		"rideId":    rideID,
		"status":    "started",
		"message":   "Your ride has started",
		"timestamp": time.Now().UTC(),
	})

	logger.Info("ride started",
		"rideID", rideID,
		"driverID", driverID,
		"userID", userID,
		"driverUserID", driverUserID,
		"driverName", driver.User.Name,
		"riderID", ride.RiderID,
		"pickup", ride.PickupAddress,
		"dropoff", ride.DropoffAddress,
		"startedAt", startedAt,
	)

	// ✅ Refresh ride from DB to get all updated fields
	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
}

// ✅ FIXED: Accept userID and fetch driver profile first
// ✅ UPDATED CompleteRide with tracking cache cleanup
func (s *service) CompleteRide(ctx context.Context, userID, rideID string, req dto.CompleteRideRequest) (*dto.RideResponse, error) {
	// ✅ STEP 1: Fetch driver profile using user ID
	driver, err := s.driversRepo.FindDriverByUserID(ctx, userID)
	if err != nil {
		logger.Error("driver profile not found for user",
			"error", err,
			"userID", userID,
		)
		return nil, response.NotFoundError("Driver profile not found")
	}

	driverID := driver.ID         // ✅ Driver profile ID
	driverUserID := driver.UserID // ✅ Driver user ID for WebSocket and wallet

	// ✅ STEP 2: Fetch ride and validate authorization
	ride, err := s.repo.FindRideByID(ctx, rideID)
	if err != nil {
		return nil, response.NotFoundError("Ride")
	}

	if ride.DriverID == nil || *ride.DriverID != driverID {
		logger.Warn("unauthorized attempt to complete ride",
			"userID", userID,
			"driverID", driverID,
			"rideDriverID", ride.DriverID,
			"rideID", rideID,
		)
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
			logger.Error("failed to capture hold",
				"error", err,
				"rideID", rideID,
				"driverID", driverID,
				"userID", userID,
				"driverUserID", driverUserID,
				"driverName", driver.User.Name,
			)
		}
	}

	// Credit driver (80% commission)
	driverEarnings := actualFareResp.TotalFare * 0.80

	// ✅ Use driver user ID for wallet credit
	s.walletService.CreditWallet(
		ctx,
		driverUserID, // ✅ Use driver's user ID, not driver profile ID
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
	s.driversRepo.IncrementTrips(ctx, driverID)
	s.driversRepo.UpdateEarnings(ctx, driverID, driverEarnings)
	s.ridersRepo.IncrementTotalRides(ctx, ride.RiderID)

	// Update driver status back to online
	s.driversRepo.UpdateDriverStatus(ctx, driverID, "online")

	// ✅ CLEANUP: Clear all driver/ride caches
	// Clear driver busy flag
	busyKey := fmt.Sprintf("driver:busy:%s", driverID)
	cache.Delete(ctx, busyKey)

	// ✅ NEW: Clear active ride tracking cache
	activeRideCacheKey := fmt.Sprintf("driver:active:ride:%s", driverID)
	cache.Delete(ctx, activeRideCacheKey)

	// Clear ride cache
	rideCacheKey := fmt.Sprintf("ride:active:%s", rideID)
	cache.Delete(ctx, rideCacheKey)

	// ✅ Notify both parties via WebSocket using user IDs
	websocketutil.SendToUser(ride.RiderID, websocket.TypeRideCompleted, map[string]interface{}{
		"rideId":     rideID,
		"actualFare": actualFareResp.TotalFare,
		"message":    "Your ride is complete. Thank you for riding with us!",
		"timestamp":  time.Now().UTC(),
	})

	websocketutil.SendToUser(driverUserID, websocket.TypeRideCompleted, map[string]interface{}{
		"rideId":    rideID,
		"earnings":  driverEarnings,
		"message":   "Ride completed successfully",
		"timestamp": time.Now().UTC(),
	})

	logger.Info("ride completed successfully",
		"rideID", rideID,
		"driverID", driverID,
		"userID", userID,
		"driverUserID", driverUserID,
		"driverName", driver.User.Name,
		"riderID", ride.RiderID,
		"actualFare", actualFareResp.TotalFare,
		"driverEarnings", driverEarnings,
		"actualDistance", req.ActualDistance,
		"actualDuration", req.ActualDuration,
	)

	// Refresh ride data from DB
	ride, _ = s.repo.FindRideByID(ctx, rideID)

	return dto.ToRideResponse(ride), nil
}

// ✅ UPDATED assignDriverToRide with all status updates + improvements
func (s *service) assignDriverToRide(ctx context.Context, rideID, driverID string) error {
	fmt.Println("Run func: assignDriverToRide")

	// Atomic update: only one driver can change status from "searching" to "accepted"
	err := s.repo.UpdateRideStatusAndDriver(ctx, rideID, "accepted", "searching", driverID)
	if err != nil {
		logger.Warn("failed to assign driver - ride may be already accepted",
			"error", err,
			"rideID", rideID,
			"driverID", driverID,
		)
		return response.BadRequest("Ride already accepted by another driver")
	}

	// Cancel all other pending requests for this ride
	go s.repo.CancelPendingRequestsExcept(context.Background(), rideID, driverID)

	// Update driver status
	s.driversRepo.UpdateDriverStatus(ctx, driverID, "busy")

	// Mark driver as busy in cache
	busyKey := fmt.Sprintf("driver:busy:%s", driverID)
	cache.Set(ctx, busyKey, "true", 30*time.Minute)

	// Get ride for cache update and location calculation
	ride, _ := s.repo.FindRideByID(ctx, rideID)

	// Update cache
	cacheKey := fmt.Sprintf("ride:active:%s", rideID)
	cache.SetJSON(ctx, cacheKey, ride, 30*time.Minute)

	// Get driver details for notification
	driver, err := s.driversRepo.FindDriverByID(ctx, driverID)
	if err != nil {
		logger.Error("failed to fetch driver details for notification",
			"error", err,
			"driverID", driverID,
		)
		return nil
	}

	// Safely check if vehicle exists before accessing
	if driver.Vehicle == nil {
		logger.Warn("driver has no vehicle assigned",
			"driverID", driverID,
			"driverName", driver.User.Name,
		)
		return nil
	}

	//  Cache active ride for tracking module
	activeRideCacheKey := fmt.Sprintf("driver:active:ride:%s", driverID)
	cache.SetJSON(ctx, activeRideCacheKey, map[string]string{
		"rideID":  rideID,
		"riderID": ride.RiderID,
	}, 30*time.Minute)

	// ✅ Get driver's current location
	var driverLat, driverLon float64
	var calculatedETA int

	driverLocation, err := s.trackingService.GetDriverLocation(ctx, driverID)
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
			"driverID", driverID,
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
			"driverID", driverID,
			"driverName", driver.User.Name,
		)
	}

	logger.Info("ride assigned to driver successfully",
		"rideID", rideID,
		"driverID", driverID,
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
		return response.NotFoundError("Ride")
	}

	// ✅ FIX: Check authorization properly
	var isRider bool
	var isDriver bool
	var driverProfile *models.DriverProfile
	var driverProfileID string

	// Check if user is the rider (riderID is user ID)
	isRider = ride.RiderID == userID

	// ✅ Check if user is the driver (need to fetch driver profile first)
	if ride.DriverID != nil {
		// Try to fetch driver profile by user ID
		driverProfile, err = s.driversRepo.FindDriverByUserID(ctx, userID)
		if err == nil && driverProfile != nil {
			// User has a driver profile - check if it matches the ride's driver
			isDriver = driverProfile.ID == *ride.DriverID
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

	logger.Info("ride cancellation initiated",
		"rideID", rideID,
		"cancelledBy", cancelledBy,
		"userID", userID,
		"isRider", isRider,
		"isDriver", isDriver,
		"driverProfileID", driverProfileID,
		"cancellationFee", cancellationFee,
	)

	// Process wallet transactions
	if ride.WalletHoldID != nil {
		if cancellationFee > 0 {
			// Capture cancellation fee, release rest
			if _, err := s.walletService.CaptureHold(ctx, ride.RiderID, walletdto.CaptureHoldRequest{
				HoldID:      *ride.WalletHoldID,
				Amount:      &cancellationFee,
				Description: "Cancellation fee",
			}); err != nil {
				logger.Error("failed to capture cancellation fee", "error", err, "rideID", rideID)
			} else {
				logger.Info("cancellation fee captured",
					"rideID", rideID,
					"amount", cancellationFee,
					"riderID", ride.RiderID,
				)
			}

			// ✅ Credit driver with cancellation fee if assigned
			if ride.DriverID != nil {
				// If we already have driver profile from authorization, use it
				// Otherwise, fetch it
				var driver *models.DriverProfile
				if driverProfile != nil && driverProfile.ID == *ride.DriverID {
					driver = driverProfile
				} else {
					driver, err = s.driversRepo.FindDriverByID(ctx, *ride.DriverID)
					if err != nil {
						logger.Error("failed to fetch driver for cancellation fee credit",
							"error", err,
							"driverID", *ride.DriverID,
							"rideID", rideID,
						)
					}
				}

				if driver != nil {
					// ✅ Use driver's user ID for wallet credit
					s.walletService.CreditWallet(
						ctx,
						driver.UserID, // ✅ Use user ID, not driver profile ID
						cancellationFee,
						"cancellation_fee",
						rideID,
						"Cancellation fee compensation",
						nil,
					)

					logger.Info("cancellation fee credited to driver",
						"rideID", rideID,
						"driverID", driver.ID,
						"driverUserID", driver.UserID,
						"driverName", driver.User.Name,
						"amount", cancellationFee,
					)
				}
			}
		} else {
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

	// ✅ CLEANUP: Update driver status and clear all caches if driver was assigned
	if ride.DriverID != nil {
		driverID := *ride.DriverID

		// Update driver status to online
		s.driversRepo.UpdateDriverStatus(ctx, driverID, "online")

		// ✅ Clear driver busy flag from cache
		busyKey := fmt.Sprintf("driver:busy:%s", driverID)
		cache.Delete(ctx, busyKey)

		// ✅ NEW: Clear active ride tracking cache
		activeRideCacheKey := fmt.Sprintf("driver:active:ride:%s", driverID)
		cache.Delete(ctx, activeRideCacheKey)

		logger.Debug("driver caches cleared",
			"driverID", driverID,
			"rideID", rideID,
			"clearedKeys", []string{busyKey, activeRideCacheKey},
		)

		// ✅ Get driver details for WebSocket notification
		// If we already have driver profile from authorization, use it
		// Otherwise, fetch it
		var driver *models.DriverProfile
		if driverProfile != nil && driverProfile.ID == driverID {
			driver = driverProfile
		} else {
			driver, err = s.driversRepo.FindDriverByID(ctx, driverID)
			if err != nil {
				logger.Error("failed to fetch driver for cancellation notification",
					"error", err,
					"driverID", driverID,
					"rideID", rideID,
				)
			}
		}

		if driver != nil {
			// ✅ Notify driver via WebSocket using user ID (only if driver didn't cancel)
			if !isDriver {
				websocketutil.SendToUser(driver.UserID, websocket.TypeRideCancelled, map[string]interface{}{
					"rideId":    rideID,
					"message":   "Ride was cancelled by rider",
					"reason":    req.Reason,
					"timestamp": time.Now().UTC(),
				})

				logger.Info("driver notified of cancellation",
					"rideID", rideID,
					"driverID", driver.ID,
					"driverUserID", driver.UserID,
					"driverName", driver.User.Name,
					"cancelledBy", "rider",
				)
			}
		}
	}

	// ✅ Clear ride cache
	rideCacheKey := fmt.Sprintf("ride:active:%s", rideID)
	cache.Delete(ctx, rideCacheKey)

	// ✅ Notify rider via WebSocket (ride.RiderID is user ID, so OK)
	if isDriver {
		websocketutil.SendToUser(ride.RiderID, websocket.TypeRideCancelled, map[string]interface{}{
			"rideId":    rideID,
			"message":   "Ride was cancelled by driver",
			"reason":    req.Reason,
			"timestamp": time.Now().UTC(),
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
		"cancellationFee", cancellationFee,
		"driverID", ride.DriverID,
		"status", "success",
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
