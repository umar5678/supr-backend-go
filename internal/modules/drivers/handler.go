// internal/modules/drivers/handler.go
package drivers

import (
	"github.com/gin-gonic/gin"
	driverdto "github.com/umar5678/go-backend/internal/modules/drivers/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterDriver godoc
// @Summary Register driver profile
// @Tags drivers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body driverdto.RegisterDriverRequest true "Driver registration data"
// @Success 200 {object} response.Response{data=driverdto.DriverProfileResponse}
// @Router /drivers/register [post]
func (h *Handler) RegisterDriver(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req driverdto.RegisterDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	profile, err := h.service.RegisterDriver(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, profile, "Driver registered successfully")
}

// GetProfile godoc
// @Summary Get driver profile
// @Tags drivers
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=driverdto.DriverProfileResponse}
// @Router /drivers/profile [get]
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
// @Summary Update driver profile
// @Tags drivers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body driverdto.UpdateDriverProfileRequest true "Update data"
// @Success 200 {object} response.Response{data=driverdto.DriverProfileResponse}
// @Router /drivers/profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req driverdto.UpdateDriverProfileRequest
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

// UpdateVehicle godoc
// @Summary Update vehicle information
// @Tags drivers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body driverdto.UpdateVehicleRequest true "Vehicle update data"
// @Success 200 {object} response.Response{data=driverdto.VehicleResponse}
// @Router /drivers/vehicle [put]
func (h *Handler) UpdateVehicle(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req driverdto.UpdateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	vehicle, err := h.service.UpdateVehicle(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, vehicle, "Vehicle updated successfully")
}

// UpdateStatus godoc
// @Summary Update driver status (online/offline)
// @Tags drivers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body driverdto.UpdateStatusRequest true "Status update"
// @Success 200 {object} response.Response{data=driverdto.DriverProfileResponse}
// @Router /drivers/status [post]
func (h *Handler) UpdateStatus(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req driverdto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	profile, err := h.service.UpdateStatus(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, profile, "Status updated successfully")
}

// UpdateLocation godoc
// @Summary Update driver location
// @Tags drivers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body driverdto.UpdateLocationRequest true "Location data"
// @Success 200 {object} response.Response
// @Router /drivers/location [post]
func (h *Handler) UpdateLocation(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req driverdto.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.UpdateLocation(c.Request.Context(), userID.(string), req); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Location updated successfully")
}

// GetWallet godoc
// @Summary Get driver wallet balance
// @Tags drivers
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=driverdto.WalletResponse}
// @Router /drivers/wallet [get]
func (h *Handler) GetWallet(c *gin.Context) {
	userID, _ := c.Get("userID")

	wallet, err := h.service.GetWallet(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, wallet, "Wallet retrieved successfully")
}

// GetDashboard godoc
// @Summary Get driver dashboard stats
// @Tags drivers
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=driverdto.DriverDashboardResponse}
// @Router /drivers/dashboard [get]
func (h *Handler) GetDashboard(c *gin.Context) {
	userID, _ := c.Get("userID")

	dashboard, err := h.service.GetDashboard(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, dashboard, "Dashboard retrieved successfully")
}
