package fraud

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/modules/fraud/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GetFraudPattern godoc
// @Summary Get fraud pattern details (Admin)
// @Tags fraud
// @Security BearerAuth
// @Produce json
// @Param id path string true "Pattern ID"
// @Success 200 {object} response.Response{data=dto.FraudPatternResponse}
// @Router /fraud/patterns/{id} [get]
func (h *Handler) GetFraudPattern(c *gin.Context) {
	patternID := c.Param("id")

	pattern, err := h.service.GetFraudPattern(c.Request.Context(), patternID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, pattern, "Fraud pattern retrieved successfully")
}

// ListFraudPatterns godoc
// @Summary List fraud patterns (Admin)
// @Tags fraud
// @Security BearerAuth
// @Produce json
// @Param patternType query string false "Filter by pattern type"
// @Param status query string false "Filter by status"
// @Param minRiskScore query int false "Minimum risk score"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response{data=[]dto.FraudPatternListResponse}
// @Router /fraud/patterns [get]
func (h *Handler) ListFraudPatterns(c *gin.Context) {
	var req dto.ListFraudPatternsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}

	patterns, total, err := h.service.ListFraudPatterns(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	pagination := response.NewPaginationMeta(total, req.Page, req.Limit)
	response.Paginated(c, patterns, pagination, "Fraud patterns retrieved successfully")
}

// ReviewFraudPattern godoc
// @Summary Review fraud pattern (Admin)
// @Tags fraud
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Pattern ID"
// @Param request body dto.ReviewFraudPatternRequest true "Review data"
// @Success 200 {object} response.Response
// @Router /fraud/patterns/{id}/review [post]
func (h *Handler) ReviewFraudPattern(c *gin.Context) {
	userID, _ := c.Get("userID")
	patternID := c.Param("id")

	var req dto.ReviewFraudPatternRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.ReviewFraudPattern(c.Request.Context(), patternID, userID.(string), req); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Fraud pattern reviewed successfully")
}

// GetFraudStats godoc
// @Summary Get fraud detection statistics (Admin)
// @Tags fraud
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=dto.FraudStatsResponse}
// @Router /fraud/stats [get]
func (h *Handler) GetFraudStats(c *gin.Context) {
	stats, err := h.service.GetFraudStats(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, stats, "Fraud stats retrieved successfully")
}

// CheckUserRiskScore godoc
// @Summary Check user risk score (Admin)
// @Tags fraud
// @Security BearerAuth
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} response.Response{data=int}
// @Router /fraud/users/{userId}/risk-score [get]
func (h *Handler) CheckUserRiskScore(c *gin.Context) {
	userID := c.Param("userId")

	riskScore, err := h.service.CheckUserRiskScore(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, map[string]int{"riskScore": riskScore}, "Risk score calculated successfully")
}
