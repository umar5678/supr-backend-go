package profile

import (
	"github.com/gin-gonic/gin"
	"github.com/umar5678/go-backend/internal/modules/profile/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// UpdateEmergencyContact godoc
// @Summary Update emergency contact
// @Tags profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.UpdateEmergencyContactRequest true "Emergency contact data"
// @Success 200 {object} response.Response
// @Router /profile/emergency-contact [put]
func (h *Handler) UpdateEmergencyContact(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.UpdateEmergencyContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.UpdateEmergencyContact(c.Request.Context(), userID.(string), req); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Emergency contact updated successfully")
}

// GenerateReferralCode godoc
// @Summary Generate referral code
// @Tags profile
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=dto.ReferralInfoResponse}
// @Router /profile/referral/generate [post]
func (h *Handler) GenerateReferralCode(c *gin.Context) {
	userID, _ := c.Get("userID")

	info, err := h.service.GenerateReferralCode(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, info, "Referral code generated successfully")
}

// ApplyReferralCode godoc
// @Summary Apply referral code
// @Tags profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ApplyReferralRequest true "Referral code"
// @Success 200 {object} response.Response
// @Router /profile/referral/apply [post]
func (h *Handler) ApplyReferralCode(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.ApplyReferralRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.ApplyReferralCode(c.Request.Context(), userID.(string), req); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Referral code applied successfully")
}

// GetReferralInfo godoc
// @Summary Get referral information
// @Tags profile
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=dto.ReferralInfoResponse}
// @Router /profile/referral [get]
func (h *Handler) GetReferralInfo(c *gin.Context) {
	userID, _ := c.Get("userID")

	info, err := h.service.GetReferralInfo(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, info, "Referral info retrieved successfully")
}

// SubmitKYC godoc
// @Summary Submit KYC documents
// @Tags profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.SubmitKYCRequest true "KYC data"
// @Success 201 {object} response.Response{data=dto.KYCResponse}
// @Router /profile/kyc [post]
func (h *Handler) SubmitKYC(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.SubmitKYCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	kyc, err := h.service.SubmitKYC(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, kyc, "KYC submitted successfully")
}

// GetKYC godoc
// @Summary Get KYC status
// @Tags profile
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=dto.KYCResponse}
// @Router /profile/kyc [get]
func (h *Handler) GetKYC(c *gin.Context) {
	userID, _ := c.Get("userID")

	kyc, err := h.service.GetKYC(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, kyc, "KYC retrieved successfully")
}

// SaveLocation godoc
// @Summary Save a location
// @Tags profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.SaveLocationRequest true "Location data"
// @Success 201 {object} response.Response{data=dto.SavedLocationResponse}
// @Router /profile/locations [post]
func (h *Handler) SaveLocation(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.SaveLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	location, err := h.service.SaveLocation(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, location, "Location saved successfully")
}

// GetSavedLocations godoc
// @Summary Get saved locations
// @Tags profile
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.SavedLocationResponse}
// @Router /profile/locations [get]
func (h *Handler) GetSavedLocations(c *gin.Context) {
	userID, _ := c.Get("userID")

	locations, err := h.service.GetSavedLocations(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, locations, "Locations retrieved successfully")
}

// DeleteLocation godoc
// @Summary Delete saved location
// @Tags profile
// @Security BearerAuth
// @Param id path string true "Location ID"
// @Success 200 {object} response.Response
// @Router /profile/locations/{id} [delete]
func (h *Handler) DeleteLocation(c *gin.Context) {
	userID, _ := c.Get("userID")
	locationID := c.Param("id")

	if err := h.service.DeleteLocation(c.Request.Context(), userID.(string), locationID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Location deleted successfully")
}

// SetDefaultLocation godoc
// @Summary Set default location
// @Tags profile
// @Security BearerAuth
// @Param id path string true "Location ID"
// @Success 200 {object} response.Response
// @Router /profile/locations/{id}/default [post]
func (h *Handler) SetDefaultLocation(c *gin.Context) {
	userID, _ := c.Get("userID")
	locationID := c.Param("id")

	if err := h.service.SetDefaultLocation(c.Request.Context(), userID.(string), locationID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Default location set successfully")
}

// GetRecentLocations godoc
// @Summary Get recent locations
// @Tags profile
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.RecentLocationResponse}
// @Router /profile/locations/recent [get]
func (h *Handler) GetRecentLocations(c *gin.Context) {
	userID, _ := c.Get("userID")

	locations, err := h.service.GetRecentLocations(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, locations, "Recent locations retrieved successfully")
}
