package promotions

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/modules/promotions/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreatePromoCode godoc
// @Summary Create promo code (Admin)
// @Tags Admin routes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreatePromoCodeRequest true "Promo code data"
// @Success 201 {object} response.Response{data=dto.PromoCodeResponse}
// @Router /promotions/promo-codes [post]
func (h *Handler) CreatePromoCode(c *gin.Context) {
	var req dto.CreatePromoCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	promo, err := h.service.CreatePromoCode(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, promo, "Promo code created successfully")
}

// GetPromoCode godoc
// @Summary Get promo code details
// @Tags promotions
// @Security BearerAuth
// @Produce json
// @Param code path string true "Promo code"
// @Success 200 {object} response.Response{data=dto.PromoCodeResponse}
// @Router /promotions/promo-codes/{code} [get]
func (h *Handler) GetPromoCode(c *gin.Context) {
	code := c.Param("code")

	promo, err := h.service.GetPromoCode(c.Request.Context(), code)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, promo, "Promo code retrieved successfully")
}

// ListPromoCodes godoc
// @Summary List promo codes (Admin)
// @Tags Admin routes
// @Security BearerAuth
// @Produce json
// @Param isActive query bool false "Filter by active status"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response{data=[]dto.PromoCodeResponse}
// @Router /promotions/promo-codes [get]
func (h *Handler) ListPromoCodes(c *gin.Context) {
	isActive := c.Query("isActive") == "true"
	page := 1
	limit := 20

	if _, ok := c.GetQuery("page"); ok {
		c.ShouldBindQuery(&struct {
			Page int `form:"page"`
		}{})
	}

	promos, total, err := h.service.ListPromoCodes(c.Request.Context(), isActive, page, limit)
	if err != nil {
		c.Error(err)
		return
	}

	pagination := response.NewPaginationMeta(total, page, limit)
	response.Paginated(c, promos, pagination, "Promo codes retrieved successfully")
}

// ValidatePromoCode godoc
// @Summary Validate promo code
// @Tags promotions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ValidatePromoCodeRequest true "Validation data"
// @Success 200 {object} response.Response{data=dto.ValidatePromoCodeResponse}
// @Router /promotions/validate [post]
func (h *Handler) ValidatePromoCode(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.ValidatePromoCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	validation, err := h.service.ValidatePromoCode(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, validation, "Promo code validated")
}

// ApplyPromoCode godoc
// @Summary Apply promo code to ride
// @Tags promotions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ApplyPromoCodeRequest true "Apply data"
// @Success 200 {object} response.Response{data=dto.ApplyPromoCodeResponse}
// @Router /promotions/apply [post]
func (h *Handler) ApplyPromoCode(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.ApplyPromoCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	result, err := h.service.ApplyPromoCode(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, result, "Promo code applied successfully")
}

// DeactivatePromoCode godoc
// @Summary Deactivate promo code (Admin)
// @Tags Admin routes
// @Security BearerAuth
// @Param id path string true "Promo code ID"
// @Success 200 {object} response.Response
// @Router /promotions/promo-codes/{id}/deactivate [post]
func (h *Handler) DeactivatePromoCode(c *gin.Context) {
	promoID := c.Param("id")

	if err := h.service.DeactivatePromoCode(c.Request.Context(), promoID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Promo code deactivated successfully")
}

// GetFreeRideCredits godoc
// @Summary Get user's free ride credits
// @Tags promotions
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=float64}
// @Router /promotions/free-ride-credits [get]
func (h *Handler) GetFreeRideCredits(c *gin.Context) {
	userID, _ := c.Get("userID")

	credits, err := h.service.GetFreeRideCredits(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, map[string]float64{"credits": credits}, "Free ride credits retrieved successfully")
}

// AddFreeRideCredit godoc
// @Summary Add free ride credits (Admin)
// @Tags Admin routes
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.AddFreeRideCreditRequest true "Credit data"
// @Success 200 {object} response.Response
// @Router /promotions/free-ride-credits [post]
func (h *Handler) AddFreeRideCredit(c *gin.Context) {
	var req dto.AddFreeRideCreditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.AddFreeRideCredit(c.Request.Context(), req); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Free ride credits added successfully")
}
