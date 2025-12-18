package auth

import (
	"fmt"

	"github.com/gin-gonic/gin"
	authdto "github.com/umar5678/go-backend/internal/modules/auth/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// PhoneSignup godoc
// @Summary Signup with phone (riders / drivers / service providers)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authdto.PhoneSignupRequest true "Signup data"
// @Success 201 {object} response.Response{data=authdto.AuthResponse}
// @Router /auth/phone/signup [post]
func (h *Handler) PhoneSignup(c *gin.Context) {
	var req authdto.PhoneSignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	authResp, err := h.service.PhoneSignup(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, authResp, "Account created successfully")
}

// PhoneLogin godoc
// @Summary Login with phone (riders / drivers / service providers)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authdto.PhoneLoginRequest true "Login data"
// @Success 200 {object} response.Response{data=authdto.AuthResponse}
// @Router /auth/phone/login [post]
func (h *Handler) PhoneLogin(c *gin.Context) {
	var req authdto.PhoneLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	authResp, err := h.service.PhoneLogin(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, authResp, "Login successful")
}

// EmailSignup godoc
// @Summary Signup with email (other roles)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authdto.EmailSignupRequest true "Signup data"
// @Success 201 {object} response.Response{data=authdto.AuthResponse}
// @Router /auth/email/signup [post]
func (h *Handler) EmailSignup(c *gin.Context) {
	var req authdto.EmailSignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	authResp, err := h.service.EmailSignup(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, authResp, "Account created successfully")
}

// EmailLogin godoc
// @Summary Login with email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authdto.EmailLoginRequest true "Login data"
// @Success 200 {object} response.Response{data=authdto.AuthResponse}
// @Router /auth/email/login [post]
func (h *Handler) EmailLogin(c *gin.Context) {

	fmt.Println("=========================================================called")
	var req authdto.EmailLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	authResp, err := h.service.EmailLogin(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, authResp, "Login successful")
}

// RefreshToken godoc
// @Summary Refresh access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authdto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} response.Response{data=authdto.AuthResponse}
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req authdto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	authResp, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, authResp, "Token refreshed successfully")
}

// Logout godoc
// @Summary Logout
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body authdto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} response.Response
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req authdto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	if err := h.service.Logout(c.Request.Context(), userID.(string), req.RefreshToken); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Logged out successfully")
}

// GetProfile godoc
// @Summary Get user profile
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=authdto.UserResponse}
// @Router /auth/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	profile, err := h.service.GetProfile(c.Request.Context(), userID.(string))
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, profile, "Profile retrieved successfully")
}

// // // UpdateProfile godoc
// // // @Summary Update user profile
// // // @Tags auth
// // // @Security BearerAuth
// // // @Accept json
// // // @Produce json
// // // @Param request body authdto.UpdateProfileRequest true "Profile data"
// // // @Success 200 {object} response.Response{data=authdto.ToUserResponse}
// // // @Router /auth/profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req authdto.UpdateProfileRequest
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
