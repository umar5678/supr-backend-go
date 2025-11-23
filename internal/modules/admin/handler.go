package admin

// now

import (
	"github.com/gin-gonic/gin"
	dto "github.com/umar5678/go-backend/internal/modules/admin/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// ListUsers godoc
// @Summary List all users (Admin)
// @Description Get a paginated list of users with optional filtering by role and status
// @Tags Admin routes
// @Accept json
// @Produce json
// @Param role query string false "Filter by user role (e.g., admin, service_provider, customer)"
// @Param status query string false "Filter by user status (e.g., active, pending, suspended)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Response{data=dto.ListUsersResponse} "Users retrieved successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/users [get]
// @Security BearerAuth
func (h *Handler) ListUsers(c *gin.Context) {
	role := c.Query("role")
	status := c.Query("status")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")

	result, err := h.service.ListUsers(c.Request.Context(), role, status, page, limit)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, result, "Users retrieved")
}

// ApproveServiceProvider godoc
// @Summary Approve a service provider (Admin)
// @Description Approve a pending service provider registration
// @Tags Admin routes
// @Accept json
// @Produce json
// @Param id path string true "Service Provider ID"
// @Success 200 {object} response.Response "Service provider approved successfully"
// @Failure 400 {object} response.Response "Bad request"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "Service provider not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/service-providers/{id}/approve [post]
// @Security BearerAuth
func (h *Handler) ApproveServiceProvider(c *gin.Context) {
	providerID := c.Param("id")

	if err := h.service.ApproveServiceProvider(c.Request.Context(), providerID); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Service provider approved")
}

// SuspendUser godoc
// @Summary Suspend a user (Admin)
// @Description Suspend a user account with a reason
// @Tags Admin routes
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body dto.SuspendUserRequest true "Suspension reason"
// @Success 200 {object} response.Response "User suspended successfully"
// @Failure 400 {object} response.Response "Bad request - Invalid input"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/users/{id}/suspend [post]
// @Security BearerAuth
func (h *Handler) SuspendUser(c *gin.Context) {
	userID := c.Param("id")

	var req dto.SuspendUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request"))
		return
	}

	if err := h.service.SuspendUser(c.Request.Context(), userID, req.Reason); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "User suspended")
}

// UpdateUserStatus godoc
// @Summary Update user status (Admin)
// @Description Change the status of a user account
// @Tags Admin routes
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body dto.UpdateUserStatusRequest true "New user status"
// @Success 200 {object} response.Response "User status updated successfully"
// @Failure 400 {object} response.Response "Bad request - Invalid status"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 404 {object} response.Response "User not found"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/users/{id}/status [put]
// @Security BearerAuth
func (h *Handler) UpdateUserStatus(c *gin.Context) {
	userID := c.Param("id")

	var req dto.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request"))
		return
	}

	if err := h.service.UpdateUserStatus(c.Request.Context(), userID, req.Status); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "User status updated")
}

// GetDashboardStats godoc
// @Summary Get admin dashboard statistics
// @Description Retrieve statistical data for the admin dashboard
// @Tags Admin routes
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=dto.DashboardStatsResponse} "Dashboard statistics retrieved successfully"
// @Failure 401 {object} response.Response "Unauthorized"
// @Failure 500 {object} response.Response "Internal server error"
// @Router /admin/dashboard/stats [get]
// @Security BearerAuth
func (h *Handler) GetDashboardStats(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, stats, "Dashboard stats retrieved")
}
