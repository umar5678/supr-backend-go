package tracking

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/modules/tracking/dto"
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
	userID, _ := c.Get("userID")
	driverID := userID.(string) // Assuming userID is driverID for drivers

	var req dto.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.UpdateDriverLocation(c.Request.Context(), driverID, req); err != nil {
		c.Error(err)
		return
	}

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
