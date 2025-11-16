package vehicles

import (
	"github.com/gin-gonic/gin"
	_ "github.com/umar5678/go-backend/internal/modules/vehicles/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetAllVehicleTypes godoc
// @Summary Get all vehicle types
// @Tags vehicles
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.VehicleTypeResponse}
// @Router /vehicles/types [get]
func (h *Handler) GetAllVehicleTypes(c *gin.Context) {
	vehicleTypes, err := h.service.GetAllVehicleTypes(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, vehicleTypes, "Vehicle types retrieved successfully")
}

// GetActiveVehicleTypes godoc
// @Summary Get active vehicle types
// @Tags vehicles
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.VehicleTypeResponse}
// @Router /vehicles/types/active [get]
func (h *Handler) GetActiveVehicleTypes(c *gin.Context) {
	vehicleTypes, err := h.service.GetActiveVehicleTypes(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, vehicleTypes, "Active vehicle types retrieved successfully")
}

// GetVehicleTypeByID godoc
// @Summary Get vehicle type by ID
// @Tags vehicles
// @Produce json
// @Param id path string true "Vehicle Type ID"
// @Success 200 {object} response.Response{data=dto.VehicleTypeResponse}
// @Router /vehicles/types/{id} [get]
func (h *Handler) GetVehicleTypeByID(c *gin.Context) {
	id := c.Param("id")

	vehicleType, err := h.service.GetVehicleTypeByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, vehicleType, "Vehicle type retrieved successfully")
}
