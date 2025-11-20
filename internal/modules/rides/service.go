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

// ✅ FIX: Get rider user ID for WebSocket
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

			// ✅ FIX: riderID is already user ID, so this is OK
			s.wsHelper.SendRideStatusToBoth(bgCtx, riderID, "", rideID, "cancelled", "No drivers available. Your payment has been refunded.")
		}
	}()

	// 6. Notify rider via WebSocket (riderID is user ID, so OK)
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

// ✅ FIXED: Get driver user ID for WebSocket
func (s *service) sendRideRequestToDriver(
	ctx context.Context,
	ride *models.Ride,
	driver trackingdto.DriverLocationResponse,
	resultChan chan<- string,
	errorChan chan<- error,
) {
	requestID := uuid.New().String()
	expiresAt := time.Now().Add(10 * time.Second)

	// ✅ FIX: Fetch driver details to get USER ID
	driverDetails, err := s.driversRepo.FindDriverByID(ctx, driver.DriverID)
	if err != nil {
		logger.Error("failed to fetch driver details",
			"error", err,
			"driverID", driver.DriverID,
		)
		errorChan <- err
		return
	}

	// ✅ CRITICAL: Use user ID for WebSocket, not driver ID
	userIDForWebSocket := driverDetails.UserID

	rideRequest := &models.RideRequest{
		ID:        requestID,
		RideID:    ride.ID,
		DriverID:  driver.DriverID, // Still use driver ID for database
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

	// ✅ FIX: Send to USER ID, not driver ID
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

func (s *service) AcceptRide(ctx context.Context, driverID, rideID string) (*dto.RideResponse, error) {
	rideRequest, err := s.repo.FindRideRequestByRideAndDriver(ctx, rideID, driverID)
	if err != nil {
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

	driver, err := s.driversRepo.FindDriverByID(ctx, driverID)
	if err != nil {
		return nil, response.NotFoundError("Driver profile")
	}

	if driver.Status != "online" {
		return nil, response.BadRequest("Driver must be online to accept rides")
	}

	logger.Info("driver attempting to accept ride request",
		"rideID", rideID,
		"driverID", driverID,
		"driverName", driver.User.Name,
		"requestID", rideRequest.ID,
		"riderID", ride.RiderID,
	)

	if err := s.repo.UpdateRideRequestStatus(ctx, rideRequest.ID, "accepted", nil); err != nil {
		logger.Error("failed to update ride request status",
			"error", err,
			"requestID", rideRequest.ID,
			"driverID", driverID,
			"driverName", driver.User.Name,
		)
		return nil, response.InternalServerError("Failed to accept ride", err)
	}

	if err := s.assignDriverToRide(ctx, rideID, driverID); err != nil {
		s.repo.UpdateRideRequestStatus(ctx, rideRequest.ID, "pending", nil)
		logger.Error("failed to assign driver to ride",
			"error", err,
			"rideID", rideID,
			"driverID", driverID,
			"driverName", driver.User.Name,
		)
		return nil, err
	}

	logger.Info("ride successfully accepted by driver",
		"rideID", rideID,
		"driverID", driverID,
		"driverName", driver.User.Name,
		"driverPhone", driver.User.Phone,
		"vehicleType", driver.Vehicle.VehicleType.Name,
		"vehiclePlate", driver.Vehicle.LicensePlate,
		"riderID", ride.RiderID,
	)

	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
}

// ✅ FIX: Get rider user ID for WebSocket (riderID is already user ID, so OK)
func (s *service) assignDriverToRide(ctx context.Context, rideID, driverID string) error {
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

	// Update cache
	ride, _ := s.repo.FindRideByID(ctx, rideID)
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

	// ✅ ride.RiderID is user ID, so this is OK
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
	)

	return nil
}

func (s *service) RejectRide(ctx context.Context, driverID, rideID string, req dto.RejectRideRequest) error {
	rideRequest, err := s.repo.FindRideRequestByRideAndDriver(ctx, rideID, driverID)
	if err != nil {
		return response.NotFoundError("Ride request not found")
	}

	if rideRequest.Status != "pending" {
		return response.BadRequest("Ride request is no longer available")
	}

	// Fetch driver details for logging
	driver, err := s.driversRepo.FindDriverByID(ctx, driverID)
	if err != nil {
		logger.Error("failed to fetch driver details",
			"error", err,
			"driverID", driverID,
		)
		// Continue with rejection even if we can't get driver details
	}

	if err := s.repo.UpdateRideRequestStatus(ctx, rideRequest.ID, "rejected", &req.Reason); err != nil {
		return response.InternalServerError("Failed to reject ride", err)
	}

	if driver != nil {
		logger.Info("ride rejected by driver",
			"rideID", rideID,
			"driverID", driverID,
			"driverName", driver.User.Name,
			"requestID", rideRequest.ID,
			"reason", req.Reason,
		)
	} else {
		logger.Info("ride rejected by driver",
			"rideID", rideID,
			"driverID", driverID,
			"requestID", rideRequest.ID,
			"reason", req.Reason,
		)
	}

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

	// Get driver details for logging
	driver, err := s.driversRepo.FindDriverByID(ctx, driverID)
	var driverName string
	if err == nil && driver != nil {
		driverName = driver.User.Name
	}

	if err := s.repo.UpdateRideStatus(ctx, rideID, "arrived"); err != nil {
		return nil, response.InternalServerError("Failed to update status", err)
	}

	// ✅ Notify rider via WebSocket (ride.RiderID is user ID, so OK)
	websocketutil.SendToUser(ride.RiderID, websocket.TypeRideStatusUpdate, map[string]interface{}{
		"rideId":    rideID,
		"status":    "arrived",
		"message":   "Your driver has arrived at the pickup location",
		"timestamp": time.Now().UTC(),
	})

	logger.Info("driver arrived at pickup location",
		"rideID", rideID,
		"driverID", driverID,
		"driverName", driverName,
		"riderID", ride.RiderID,
	)

	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
}

// ✅ FIX: Get driver user ID for WebSocket
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

	// ✅ FIX: Get driver details to fetch user ID for WebSocket
	driver, err := s.driversRepo.FindDriverByID(ctx, driverID)
	if err != nil {
		logger.Error("failed to fetch driver details",
			"error", err,
			"driverID", driverID,
		)
		return nil, response.InternalServerError("Failed to fetch driver details", err)
	}

	driverUserID := driver.UserID // ✅ This is the user ID for WebSocket

	if err := s.repo.UpdateRideStatus(ctx, rideID, "started"); err != nil {
		return nil, response.InternalServerError("Failed to start ride", err)
	}

	// Update driver status
	s.driversRepo.UpdateDriverStatus(ctx, *ride.DriverID, "on_trip")

	// ✅ FIX: Use user IDs for both rider and driver
	// ride.RiderID is already user ID, driverUserID is the driver's user ID
	websocketutil.SendRideStatusUpdate(ride.RiderID, driverUserID, map[string]interface{}{
		"rideId":    rideID,
		"status":    "started",
		"message":   "Your ride has started",
		"timestamp": time.Now().UTC(),
	})

	logger.Info("ride started",
		"rideID", rideID,
		"driverID", driverID,
		"driverUserID", driverUserID,
		"driverName", driver.User.Name,
		"riderID", ride.RiderID,
		"pickup", ride.PickupAddress,
		"dropoff", ride.DropoffAddress,
	)

	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
}

// ✅ FIX: Get driver user ID for WebSocket
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

	// ✅ FIX: Get driver details to fetch user ID for WebSocket
	driver, err := s.driversRepo.FindDriverByID(ctx, driverID)
	if err != nil {
		logger.Error("failed to fetch driver details",
			"error", err,
			"driverID", driverID,
		)
		return nil, response.InternalServerError("Failed to fetch driver details", err)
	}

	driverUserID := driver.UserID // ✅ This is the user ID for WebSocket

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
				"driverUserID", driverUserID,
				"driverName", driver.User.Name,
			)
		}
	}

	// Credit driver (80% commission)
	driverEarnings := actualFareResp.TotalFare * 0.80

	// ✅ FIX: Use driver user ID for wallet credit
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
	s.driversRepo.IncrementTrips(ctx, *ride.DriverID)
	s.driversRepo.UpdateEarnings(ctx, *ride.DriverID, driverEarnings)
	s.ridersRepo.IncrementTotalRides(ctx, ride.RiderID)

	// Update driver status back to online
	s.driversRepo.UpdateDriverStatus(ctx, *ride.DriverID, "online")

	// Clear driver busy flag from cache
	busyKey := fmt.Sprintf("driver:busy:%s", *ride.DriverID)
	cache.Delete(ctx, busyKey)

	// Clear ride cache
	cache.Delete(ctx, fmt.Sprintf("ride:active:%s", rideID))

	// ✅ FIX: Notify both parties via WebSocket using user IDs
	// Notify rider (ride.RiderID is user ID, so OK)
	websocketutil.SendToUser(ride.RiderID, websocket.TypeRideCompleted, map[string]interface{}{
		"rideId":     rideID,
		"actualFare": actualFareResp.TotalFare,
		"message":    "Your ride is complete. Thank you for riding with us!",
		"timestamp":  time.Now().UTC(),
	})

	// ✅ FIX: Notify driver using user ID, not driver profile ID
	websocketutil.SendToUser(driverUserID, websocket.TypeRideCompleted, map[string]interface{}{
		"rideId":    rideID,
		"earnings":  driverEarnings,
		"message":   "Ride completed successfully",
		"timestamp": time.Now().UTC(),
	})

	logger.Info("ride completed successfully",
		"rideID", rideID,
		"driverID", driverID,
		"driverUserID", driverUserID,
		"driverName", driver.User.Name,
		"riderID", ride.RiderID,
		"actualFare", actualFareResp.TotalFare,
		"driverEarnings", driverEarnings,
		"actualDistance", req.ActualDistance,
		"actualDuration", req.ActualDuration,
	)

	ride, _ = s.repo.FindRideByID(ctx, rideID)
	return dto.ToRideResponse(ride), nil
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

// ✅ FIX: Get driver user ID for WebSocket
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

			// ✅ FIX: Get driver user ID for wallet credit
			if ride.DriverID != nil {
				driver, err := s.driversRepo.FindDriverByID(ctx, *ride.DriverID)
				if err != nil {
					logger.Error("failed to fetch driver for cancellation fee credit",
						"error", err,
						"driverID", *ride.DriverID,
					)
				} else {
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
				}
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

	// ✅ FIX: Update driver status and notify via WebSocket using user ID
	if ride.DriverID != nil {
		s.driversRepo.UpdateDriverStatus(ctx, *ride.DriverID, "online")

		// Clear driver busy flag from cache
		busyKey := fmt.Sprintf("driver:busy:%s", *ride.DriverID)
		cache.Delete(ctx, busyKey)

		// ✅ FIX: Get driver user ID for WebSocket notification
		driver, err := s.driversRepo.FindDriverByID(ctx, *ride.DriverID)
		if err != nil {
			logger.Error("failed to fetch driver for cancellation notification",
				"error", err,
				"driverID", *ride.DriverID,
			)
		} else {
			// ✅ Notify driver via WebSocket using user ID
			websocketutil.SendToUser(driver.UserID, websocket.TypeRideCancelled, map[string]interface{}{
				"rideId":    rideID,
				"message":   "Ride was cancelled by rider",
				"timestamp": time.Now().UTC(),
			})

			logger.Info("notified driver of cancellation",
				"rideID", rideID,
				"driverID", *ride.DriverID,
				"driverUserID", driver.UserID,
				"driverName", driver.User.Name,
			)
		}
	}

	// Clear cache
	cache.Delete(ctx, fmt.Sprintf("ride:active:%s", rideID))

	// ✅ Notify rider via WebSocket (ride.RiderID is user ID, so OK)
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
