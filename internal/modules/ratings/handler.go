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
