package riders

import (
	"github.com/gin-gonic/gin"
	riderdto "github.com/umar5678/go-backend/internal/modules/riders/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetProfile godoc
// @Summary Get rider profile
// @Tags riders
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=riderdto.RiderProfileResponse}
// @Router /riders/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	profile, err := h.service.GetProfile(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, profile, "Profile retrieved successfully")
}

// UpdateProfile godoc
// @Summary Update rider profile
// @Tags riders
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body riderdto.UpdateProfileRequest true "Profile data"
// @Success 200 {object} response.Response{data=riderdto.RiderProfileResponse}
// @Router /riders/profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req riderdto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	profile, err := h.service.UpdateProfile(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, profile, "Profile updated successfully")
}

// GetStats godoc
// @Summary Get rider statistics
// @Tags riders
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=riderdto.RiderStatsResponse}
// @Router /riders/stats [get]
func (h *Handler) GetStats(c *gin.Context) {
	userID, _ := c.Get("userID")

	stats, err := h.service.GetStats(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, stats, "Statistics retrieved successfully")
}
