package sos

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/modules/sos/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// TriggerSOS godoc
// @Summary Trigger SOS emergency alert
// @Tags sos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.TriggerSOSRequest true "SOS data"
// @Success 201 {object} response.Response{data=dto.SOSAlertResponse}
// @Router /sos/trigger [post]
func (h *Handler) TriggerSOS(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.TriggerSOSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	alert, err := h.service.TriggerSOS(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, alert, "SOS alert triggered successfully")
}

// GetSOS godoc
// @Summary Get SOS alert details
// @Tags sos
// @Security BearerAuth
// @Produce json
// @Param id path string true "Alert ID"
// @Success 200 {object} response.Response{data=dto.SOSAlertResponse}
// @Router /sos/{id} [get]
func (h *Handler) GetSOS(c *gin.Context) {
	userID, _ := c.Get("userID")
	alertID := c.Param("id")

	alert, err := h.service.GetSOS(c.Request.Context(), userID.(string), alertID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, alert, "SOS alert retrieved successfully")
}

// GetActiveSOS godoc
// @Summary Get active SOS alert
// @Tags sos
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=dto.SOSAlertResponse}
// @Router /sos/active [get]
func (h *Handler) GetActiveSOS(c *gin.Context) {
	userID, _ := c.Get("userID")

	alert, err := h.service.GetActiveSOS(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, alert, "Active SOS alert retrieved successfully")
}

// ListSOS godoc
// @Summary List SOS alerts
// @Tags sos
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response{data=[]dto.SOSAlertListResponse}
// @Router /sos [get]
func (h *Handler) ListSOS(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.ListSOSRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}

	alerts, total, err := h.service.ListSOS(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	pagination := response.NewPaginationMeta(total, req.Page, req.Limit)
	response.Paginated(c, alerts, pagination, "SOS alerts retrieved successfully")
}

// ResolveSOS godoc
// @Summary Resolve SOS alert
// @Tags sos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Alert ID"
// @Param request body dto.ResolveSOSRequest true "Resolution notes"
// @Success 200 {object} response.Response
// @Router /sos/{id}/resolve [post]
func (h *Handler) ResolveSOS(c *gin.Context) {
	userID, _ := c.Get("userID")
	alertID := c.Param("id")

	var req dto.ResolveSOSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = dto.ResolveSOSRequest{}
	}

	if err := h.service.ResolveSOS(c.Request.Context(), userID.(string), alertID, req); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "SOS alert resolved successfully")
}

// CancelSOS godoc
// @Summary Cancel SOS alert
// @Tags sos
// @Security BearerAuth
// @Param id path string true "Alert ID"
// @Success 200 {object} response.Response
// @Router /sos/{id}/cancel [post]
func (h *Handler) CancelSOS(c *gin.Context) {
	userID, _ := c.Get("userID")
	alertID := c.Param("id")

	if err := h.service.CancelSOS(c.Request.Context(), userID.(string), alertID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "SOS alert cancelled successfully")
}
