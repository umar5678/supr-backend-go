package rides

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/modules/rides/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreateRide godoc
// @Summary Create a new ride request
// @Tags rides
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateRideRequest true "Ride request data"
// @Success 201 {object} response.Response{data=dto.RideResponse}
// @Router /rides [post]
func (h *Handler) CreateRide(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.CreateRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	ride, err := h.service.CreateRide(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, ride, "Ride requested successfully")
}

// GetRide godoc
// @Summary Get ride details
// @Tags rides
// @Security BearerAuth
// @Produce json
// @Param id path string true "Ride ID"
// @Success 200 {object} response.Response{data=dto.RideResponse}
// @Router /rides/{id} [get]
func (h *Handler) GetRide(c *gin.Context) {
	userID, _ := c.Get("userID")
	rideID := c.Param("id")

	ride, err := h.service.GetRide(c.Request.Context(), userID.(string), rideID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, ride, "Ride retrieved successfully")
}

// ListRides godoc
// @Summary List user's rides
// @Tags rides
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param status query string false "Filter by status"
// @Param role query string true "User role (rider or driver)"
// @Success 200 {object} response.Response{data=[]dto.RideListResponse}
// @Router /rides [get]
func (h *Handler) ListRides(c *gin.Context) {
	userID, _ := c.Get("userID")
	role := c.Query("role") // "rider" or "driver"

	if role != "rider" && role != "driver" {
		c.Error(response.BadRequest("Role must be 'rider' or 'driver'"))
		return
	}

	var req dto.ListRidesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}

	rides, total, err := h.service.ListRides(c.Request.Context(), userID.(string), role, req)
	if err != nil {
		c.Error(err)
		return
	}

	pagination := response.NewPaginationMeta(total, req.Page, req.Limit)
	response.Paginated(c, rides, pagination, "Rides retrieved successfully")
}

// AcceptRide godoc
// @Summary Accept a ride request (Driver)
// @Tags rides
// @Security BearerAuth
// @Param id path string true "Ride ID"
// @Success 200 {object} response.Response{data=dto.RideResponse}
// @Router /rides/{id}/accept [post]
func (h *Handler) AcceptRide(c *gin.Context) {
	userID, _ := c.Get("userID") // ✅ This is user ID from JWT token, NOT driver profile ID
	rideID := c.Param("id")

	// ✅ Service will fetch driver profile using this userID
	ride, err := h.service.AcceptRide(c.Request.Context(), userID.(string), rideID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, ride, "Ride accepted successfully")
}

// RejectRide godoc
// @Summary Reject a ride request (Driver)
// @Tags rides
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Ride ID"
// @Param request body dto.RejectRideRequest true "Rejection data"
// @Success 200 {object} response.Response
// @Router /rides/{id}/reject [post]
func (h *Handler) RejectRide(c *gin.Context) {
	userID, _ := c.Get("userID") // ✅ User ID from JWT
	rideID := c.Param("id")

	var req dto.RejectRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest(err.Error()))
		return
	}

	err := h.service.RejectRide(c.Request.Context(), userID.(string), rideID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Ride rejected successfully")
}

// MarkArrived godoc
// @Summary Mark driver as arrived at pickup (Driver)
// @Tags rides
// @Security BearerAuth
// @Param id path string true "Ride ID"
// @Success 200 {object} response.Response{data=dto.RideResponse}
// @Router /rides/{id}/arrived [post]
func (h *Handler) MarkArrived(c *gin.Context) {
	userID, _ := c.Get("userID") // ✅ User ID from JWT
	rideID := c.Param("id")

	ride, err := h.service.MarkArrived(c.Request.Context(), userID.(string), rideID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, ride, "Marked as arrived")
}

// StartRide godoc
// @Summary Start the ride (Driver)
// @Tags rides
// @Security BearerAuth
// @Param id path string true "Ride ID"
// @Param request body dto.StartRideRequest true "Rider PIN"
// @Success 200 {object} response.Response{data=dto.RideResponse}
// @Router /rides/{id}/start [post]
func (h *Handler) StartRide(c *gin.Context) {
	userID, _ := c.Get("userID") // ✅ User ID from JWT
	rideID := c.Param("id")
	logger.Info("StartRide handler called", "userID", userID, "rideID", rideID)

	var req dto.StartRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("failed to bind request body", "error", err)
		c.Error(response.BadRequest("Invalid request body"))
		return
	}
	logger.Info("request body parsed", "rideID", rideID)

	logger.Info("calling service.StartRide", "rideID", rideID)
	ride, err := h.service.StartRide(c.Request.Context(), userID.(string), rideID, req)
	if err != nil {
		logger.Error("service.StartRide returned error", "error", err, "rideID", rideID)
		c.Error(err)
		return
	}
	logger.Info("service.StartRide returned successfully", "rideID", rideID)

	logger.Info("calling response.Success", "rideID", rideID)
	response.Success(c, ride, "Ride started successfully")
	logger.Info("response.Success completed", "rideID", rideID)
}

// CompleteRide godoc
// @Summary Complete the ride (Driver)
// @Tags rides
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Ride ID"
// @Param request body dto.CompleteRideRequest true "Completion data"
// @Success 200 {object} response.Response{data=dto.RideResponse}
// @Router /rides/{id}/complete [post]
func (h *Handler) CompleteRide(c *gin.Context) {
	userID, _ := c.Get("userID") // ✅ User ID from JWT
	rideID := c.Param("id")

	var req dto.CompleteRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest(err.Error()))
		return
	}

	ride, err := h.service.CompleteRide(c.Request.Context(), userID.(string), rideID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, ride, "Ride completed successfully")
}

// CancelRide godoc
// @Summary Cancel a ride
// @Tags rides
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Ride ID"
// @Param request body dto.CancelRideRequest true "Cancellation data"
// @Success 200 {object} response.Response
// @Router /rides/{id}/cancel [post]
func (h *Handler) CancelRide(c *gin.Context) {
	userID, _ := c.Get("userID")
	rideID := c.Param("id")

	var req dto.CancelRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = dto.CancelRideRequest{}
	}

	if err := h.service.CancelRide(c.Request.Context(), userID.(string), rideID, req); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Ride cancelled successfully")
}

// TriggerSOS godoc
// @Summary Trigger SOS alert during active ride (Rider)
// @Tags rides
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Ride ID"
// @Param request body map[string]interface{} true "Location data with latitude and longitude"
// @Success 200 {object} response.Response
// @Router /rides/{id}/emergency [post]
func (h *Handler) TriggerSOS(c *gin.Context) {
	userID, _ := c.Get("userID")
	rideID := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	// Extract location from request
	var latitude, longitude float64
	if lat, ok := req["latitude"].(float64); ok {
		latitude = lat
	} else {
		c.Error(response.BadRequest("Latitude is required"))
		return
	}

	if lon, ok := req["longitude"].(float64); ok {
		longitude = lon
	} else {
		c.Error(response.BadRequest("Longitude is required"))
		return
	}

	if err := h.service.TriggerSOS(c.Request.Context(), userID.(string), rideID, latitude, longitude); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "SOS alert triggered - Help is on the way")
}

// GetAvailableCars godoc
// @Summary Get available cars near the rider
// @Tags rides
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.AvailableCarRequest true "Location and radius"
// @Success 200 {object} response.Response{data=dto.AvailableCarsListResponse}
// @Router /rides/available-cars [post]
func (h *Handler) GetAvailableCars(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.AvailableCarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	// Set default radius if not provided
	if req.RadiusKm == 0 {
		req.RadiusKm = 5.0
	}

	cars, err := h.service.GetAvailableCars(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, cars, "Available cars fetched successfully")
}
