package tracking

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/modules/tracking/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// UpdateLocation godoc
// @Summary Update driver location (called by driver app)
// @Tags tracking
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.UpdateLocationRequest true "Location data"
// @Success 200 {object} response.Response
// @Router /tracking/location [post]
func (h *Handler) UpdateLocation(c *gin.Context) {
	logger.Info("============================================================================")
	logger.Info("TRACKING HANDLER: UpdateLocation CALLED")

	userID, exists := c.Get("userID")
	if !exists {
		logger.Error("No userID in context")
		c.Error(response.BadRequest("User not authenticated"))
		return
	}

	driverUserID := userID.(string)
	logger.Info("Driver User ID from auth", "userID", driverUserID)

	var req dto.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request body", "error", err)
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	logger.Info("Location data received",
		"lat", req.Latitude,
		"lng", req.Longitude,
		"heading", req.Heading,
		"speed", req.Speed,
	)

	logger.Info(" Converting user ID to profile ID...")
	driverProfileID, err := h.service.GetDriverProfileID(c.Request.Context(), driverUserID)
	if err != nil {
		logger.Error("Driver profile NOT found",
			"userID", driverUserID,
			"error", err,
		)

		if err := h.service.UpdateDriverLocation(c.Request.Context(), driverUserID, req); err != nil {
			c.Error(err)
			return
		}
		response.Success(c, nil, "Location updated successfully")
		return
	}

	logger.Info("Driver profile found",
		"userID", driverUserID,
		"profileID", driverProfileID,
	)

	logger.Info("Checking for active ride...")
	activeRide, riderID, err := h.service.GetDriverActiveRide(c.Request.Context(), driverProfileID)

	if err != nil {
		logger.Warn("No active ride found",
			"driverProfileID", driverProfileID,
			"error", err,
		)

		if err := h.service.UpdateDriverLocation(c.Request.Context(), driverProfileID, req); err != nil {
			c.Error(err)
			return
		}

		logger.Info("Location updated (no active ride)")
		response.Success(c, nil, "Location updated successfully")
		return
	}

	if activeRide == "" || riderID == "" {
		logger.Warn("Active ride data incomplete",
			"rideID", activeRide,
			"riderID", riderID,
		)

		if err := h.service.UpdateDriverLocation(c.Request.Context(), driverProfileID, req); err != nil {
			c.Error(err)
			return
		}

		logger.Info("Location updated (incomplete ride data)")
		response.Success(c, nil, "Location updated successfully")
		return
	}

	logger.Info("Active ride FOUND",
		"driverProfileID", driverProfileID,
		"rideID", activeRide,
		"riderUserID", riderID,
	)

	logger.Info("Calling UpdateDriverLocationWithStreaming...")
	if err := h.service.UpdateDriverLocationWithStreaming(
		c.Request.Context(),
		driverProfileID,
		req,
		activeRide,
		riderID,
	); err != nil {
		logger.Error("UpdateDriverLocationWithStreaming FAILED", "error", err)
		c.Error(err)
		return
	}

	logger.Info("Location updated with streaming SUCCESS")
	logger.Info("========================================================================================")
	response.Success(c, nil, "Location updated successfully")
}

// GetDriverLocation godoc
// @Summary Get driver's current location
// @Tags tracking
// @Produce json
// @Param driverId path string true "Driver ID"
// @Success 200 {object} response.Response{data=dto.LocationResponse}
// @Router /tracking/driver/{driverId} [get]
func (h *Handler) GetDriverLocation(c *gin.Context) {
	driverID := c.Param("driverId")

	location, err := h.service.GetDriverLocation(c.Request.Context(), driverID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, location, "Location retrieved successfully")
}

// FindNearbyDrivers godoc
// @Summary Find drivers near a location
// @Tags tracking
// @Produce json
// @Param latitude query number true "Latitude"
// @Param longitude query number true "Longitude"
// @Param radiusKm query number false "Search radius in km"
// @Param vehicleTypeId query string false "Filter by vehicle type"
// @Param limit query int false "Maximum results"
// @Param onlyAvailable query bool false "Filter only available drivers (default: true)"
// @Success 200 {object} response.Response{data=dto.NearbyDriversResponse}
// @Router /tracking/nearby [get]
func (h *Handler) FindNearbyDrivers(c *gin.Context) {
	var req dto.FindNearbyDriversRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}

	result, err := h.service.FindNearbyDrivers(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, result, "Nearby drivers found")
}
