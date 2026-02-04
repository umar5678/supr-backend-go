package pricing

import (
	"strconv"
	"time"

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

// CreateSurgeZone godoc
// @Summary Create a new surge pricing zone
// @Tags pricing - admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateSurgeZoneRequest true "Surge zone details"
// @Success 201 {object} response.Response{data=dto.CreateSurgeZoneResponse}
// @Router /pricing/surge/zones [post]
func (h *Handler) CreateSurgeZone(c *gin.Context) {
	var req dto.CreateSurgeZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	if err := req.Validate(); err != nil {
		c.Error(response.BadRequest(err.Error()))
		return
	}

	result, err := h.service.CreateSurgeZone(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, result, "Surge zone created successfully")
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

// GetFareBreakdown godoc
// @Summary Get fare breakdown and estimate
// @Tags pricing
// @Produce json
// @Param pickupLat query float64 true "Pickup latitude"
// @Param pickupLon query float64 true "Pickup longitude"
// @Param dropoffLat query float64 true "Dropoff latitude"
// @Param dropoffLon query float64 true "Dropoff longitude"
// @Param vehicleTypeId query string true "Vehicle type ID"
// @Success 200 {object} response.Response{data=dto.FareBreakdownResponse}
// @Router /pricing/fare-breakdown [get]
func (h *Handler) GetFareBreakdown(c *gin.Context) {
	var req dto.GetFareBreakdownRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}

	breakdown, err := h.service.GetFareBreakdown(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, breakdown, "Fare breakdown calculated successfully")
}

// CalculateWaitTimeCharge godoc
// @Summary Calculate wait time charges
// @Tags pricing
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CalculateWaitTimeRequest true "Wait time data"
// @Success 200 {object} response.Response{data=dto.WaitTimeChargeResponse}
// @Router /pricing/wait-time [post]
func (h *Handler) CalculateWaitTimeCharge(c *gin.Context) {
	var req dto.CalculateWaitTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	arrivedAt, err := time.Parse(time.RFC3339, req.ArrivedAt)
	if err != nil {
		c.Error(response.BadRequest("Invalid arrivedAt timestamp"))
		return
	}

	charge, err := h.service.CalculateWaitTimeCharge(c.Request.Context(), req.RideID, arrivedAt)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, charge, "Wait time charge calculated successfully")
}

// ChangeDestination godoc
// @Summary Change ride destination (Driver)
// @Tags pricing
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ChangeDestinationRequest true "New destination data"
// @Success 200 {object} response.Response{data=dto.DestinationChangeResponse}
func (h *Handler) ChangeDestination(c *gin.Context) {
	userIDVal, ok := c.Get("userID")
	if !ok {
		c.Error(response.BadRequest("Missing or invalid user authentication"))
		return
	}
	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		c.Error(response.BadRequest("Missing or invalid user authentication"))
		return
	}

	var req dto.ChangeDestinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	result, err := h.service.ChangeDestination(c.Request.Context(), userID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, result, "Destination changed successfully")
}

// GetSurgePricingRules godoc
// @Summary Get all active surge pricing rules
// @Description Returns all active surge pricing rules for time-based and demand-based surge
// @Tags pricing
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.SurgePricingRuleResponse}
// @Router /pricing/surge-rules [get]
func (h *Handler) GetSurgePricingRules(c *gin.Context) {
	rules, err := h.service.GetActiveSurgePricingRules(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, rules, "Surge pricing rules retrieved successfully")
}

// CreateSurgePricingRule godoc
// @Summary Create a new surge pricing rule
// @Description Creates a new surge pricing rule for time-based or demand-based surging
// @Tags pricing
// @Accept json
// @Produce json
// @Param request body dto.CreateSurgePricingRuleRequest true "Surge pricing rule"
// @Success 201 {object} response.Response{data=dto.SurgePricingRuleResponse}
// @Router /pricing/surge-rules [post]
func (h *Handler) CreateSurgePricingRule(c *gin.Context) {
	var req dto.CreateSurgePricingRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	rule, err := h.service.CreateSurgePricingRule(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data":    rule,
		"message": "Surge pricing rule created successfully",
	})
}

// CalculateSurge godoc
// @Summary Calculate surge pricing for a location
// @Description Calculates combined time-based and demand-based surge multiplier
// @Tags pricing
// @Accept json
// @Produce json
// @Param request body dto.SurgeCalculationRequest true "Surge calculation request"
// @Success 200 {object} response.Response{data=dto.SurgeCalculationResponse}
// @Router /pricing/calculate-surge [post]
func (h *Handler) CalculateSurge(c *gin.Context) {
	var req dto.SurgeCalculationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	surge, err := h.service.CalculateCombinedSurge(c.Request.Context(), req.VehicleTypeID, req.Geohash, req.PickupLat, req.PickupLon)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, surge, "Surge calculated successfully")
}

// GetCurrentDemand godoc
// @Summary Get current demand for a zone
// @Description Returns current demand metrics (pending requests, available drivers) for a zone
// @Tags pricing
// @Produce json
// @Param geohash query string true "Geohash of the zone"
// @Success 200 {object} response.Response{data=dto.DemandTrackingResponse}
// @Router /pricing/demand [get]
func (h *Handler) GetCurrentDemand(c *gin.Context) {
	geohash := c.Query("geohash")
	if geohash == "" {
		c.Error(response.BadRequest("geohash parameter is required"))
		return
	}

	demand, err := h.service.GetCurrentDemand(c.Request.Context(), geohash)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, demand, "Current demand retrieved successfully")
}

// CalculateETA godoc
// @Summary Calculate ETA for a route
// @Description Calculates estimated time of arrival for pickup and dropoff
// @Tags pricing
// @Accept json
// @Produce json
// @Param request body dto.ETAEstimateRequest true "ETA calculation request"
// @Success 200 {object} response.Response{data=dto.ETAEstimateResponse}
// @Router /pricing/calculate-eta [post]
func (h *Handler) CalculateETA(c *gin.Context) {
	var req dto.ETAEstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	eta, err := h.service.CalculateETAEstimate(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, eta, "ETA calculated successfully")
}
