package ratings

import (
	"github.com/gin-gonic/gin"

	"github.com/umar5678/go-backend/internal/modules/ratings/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreateRating godoc
// @Summary Rate a service
// @Description Rate a completed service order
// @Tags ratings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateRatingRequest true "Rating details"
// @Success 201 {object} response.Response{data=dto.RatingResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /ratings [post]
func (h *Handler) CreateRating(c *gin.Context) {
	var req dto.CreateRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	userID, _ := c.Get("userID")

	rating, err := h.service.CreateRating(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, rating, "Rating submitted successfully")
}

// RateDriver godoc
// @Summary Rate driver (Rider)
// @Tags ratings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.RateDriverRequest true "Rating data"
// @Success 200 {object} response.Response
// @Router /ratings/driver [post]
func (h *Handler) RateDriver(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.RateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.RateDriver(c.Request.Context(), userID.(string), req); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Driver rated successfully")
}

// RateRider godoc
// @Summary Rate rider (Driver)
// @Tags ratings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.RateRiderRequest true "Rating data"
// @Success 200 {object} response.Response
// @Router /ratings/rider [post]
func (h *Handler) RateRider(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.RateRiderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.RateRider(c.Request.Context(), userID.(string), req); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Rider rated successfully")
}

// GetDriverRatingStats godoc
// @Summary Get driver rating statistics
// @Tags ratings
// @Produce json
// @Param driverId path string true "Driver ID"
// @Success 200 {object} response.Response{data=dto.RatingStatsResponse}
// @Router /ratings/driver/{driverId}/stats [get]
func (h *Handler) GetDriverRatingStats(c *gin.Context) {
	driverID := c.Param("driverId")

	stats, err := h.service.GetDriverRatingStats(c.Request.Context(), driverID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, stats, "Driver rating stats retrieved successfully")
}

// GetRiderRatingStats godoc
// @Summary Get rider rating statistics
// @Tags ratings
// @Security BearerAuth
// @Produce json
// @Param riderId path string true "Rider ID"
// @Success 200 {object} response.Response{data=dto.RatingStatsResponse}
// @Router /ratings/rider/{riderId}/stats [get]
func (h *Handler) GetRiderRatingStats(c *gin.Context) {
	riderID := c.Param("riderId")

	stats, err := h.service.GetRiderRatingStats(c.Request.Context(), riderID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, stats, "Rider rating stats retrieved successfully")
}

// GetDriverRatingBreakdown godoc
// @Summary Get driver rating breakdown
// @Tags ratings
// @Produce json
// @Param driverId path string true "Driver ID"
// @Success 200 {object} response.Response{data=dto.RatingBreakdownResponse}
// @Router /ratings/driver/{driverId}/breakdown [get]
func (h *Handler) GetDriverRatingBreakdown(c *gin.Context) {
	driverID := c.Param("driverId")

	breakdown, err := h.service.GetDriverRatingBreakdown(c.Request.Context(), driverID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, breakdown, "Rating breakdown retrieved successfully")
}

// GetRiderRatingBreakdown godoc
// @Summary Get rider rating breakdown
// @Tags ratings
// @Security BearerAuth
// @Produce json
// @Param riderId path string true "Rider ID"
// @Success 200 {object} response.Response{data=dto.RatingBreakdownResponse}
// @Router /ratings/rider/{riderId}/breakdown [get]
func (h *Handler) GetRiderRatingBreakdown(c *gin.Context) {
	riderID := c.Param("riderId")

	breakdown, err := h.service.GetRiderRatingBreakdown(c.Request.Context(), riderID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, breakdown, "Rating breakdown retrieved successfully")
}
