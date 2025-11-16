package pricing

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/modules/pricing/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetFareEstimate godoc
// @Summary Get fare estimate for a trip
// @Tags pricing
// @Accept json
// @Produce json
// @Param request body dto.FareEstimateRequest true "Fare estimate request"
// @Success 200 {object} response.Response{data=dto.FareEstimateResponse}
// @Router /pricing/estimate [post]
func (h *Handler) GetFareEstimate(c *gin.Context) {
	var req dto.FareEstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	estimate, err := h.service.GetFareEstimate(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, estimate, "Fare estimate calculated successfully")
}

// GetSurgeMultiplier godoc
// @Summary Get surge multiplier for a location
// @Tags pricing
// @Produce json
// @Param latitude query number true "Latitude"
// @Param longitude query number true "Longitude"
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Router /pricing/surge [get]
func (h *Handler) GetSurgeMultiplier(c *gin.Context) {
	latStr := c.Query("latitude")
	lonStr := c.Query("longitude")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.Error(response.BadRequest("Invalid latitude"))
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.Error(response.BadRequest("Invalid longitude"))
		return
	}

	multiplier, err := h.service.GetSurgeMultiplier(c.Request.Context(), lat, lon)
	if err != nil {
		c.Error(err)
		return
	}

	data := map[string]interface{}{
		"latitude":        lat,
		"longitude":       lon,
		"surgeMultiplier": multiplier,
		"message":         h.getSurgeMessage(multiplier),
	}

	response.Success(c, data, "Surge multiplier retrieved successfully")
}

// GetActiveSurgeZones godoc
// @Summary Get all active surge zones
// @Tags pricing
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.SurgeZoneResponse}
// @Router /pricing/surge/zones [get]
func (h *Handler) GetActiveSurgeZones(c *gin.Context) {
	zones, err := h.service.GetActiveSurgeZones(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, zones, "Active surge zones retrieved successfully")
}

func (h *Handler) getSurgeMessage(multiplier float64) string {
	switch {
	case multiplier == 1.0:
		return "Normal pricing"
	case multiplier <= 1.5:
		return "Slightly increased demand"
	case multiplier <= 2.0:
		return "High demand in this area"
	default:
		return "Very high demand in this area"
	}
}
