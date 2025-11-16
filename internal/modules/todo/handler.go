package todos

import (
	"github.com/gin-gonic/gin"

	"github.com/umar5678/go-backend/internal/modules/todo/dto"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Create godoc
// @Summary Create new todo
// @Tags todos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateTodoRequest true "Create Todo"
// @Success 201 {object} response.Response{data=dto.TodoResponse}
// @Router /todos [post]
func (h *Handler) Create(c *gin.Context) {
	userID := c.GetString("userID")
	var req dto.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request"))
		return
	}
	result, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, result, "Todo created successfully")
}

// GetAll godoc
// @Summary Get all todos
// @Tags todos
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{data=[]dto.TodoResponse}
// @Router /todos [get]
func (h *Handler) GetAll(c *gin.Context) {
	userID := c.GetString("userID")
	result, err := h.service.GetAll(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, result, "Todos retrieved successfully")
}

// GetByID godoc
// @Summary Get todo by ID
// @Tags todos
// @Security BearerAuth
// @Produce json
// @Param id path string true "Todo ID"
// @Success 200 {object} response.Response{data=dto.TodoResponse}
// @Router /todos/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	userID := c.GetString("userID")
	todoID := c.Param("id")
	result, err := h.service.GetByID(c.Request.Context(), todoID, userID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, result, "Todo retrieved successfully")
}

// Update godoc
// @Summary Update todo
// @Tags todos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Todo ID"
// @Param request body dto.UpdateTodoRequest true "Update Todo"
// @Success 200 {object} response.Response{data=dto.TodoResponse}
// @Router /todos/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	userID := c.GetString("userID")
	todoID := c.Param("id")
	var req dto.UpdateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request"))
		return
	}
	result, err := h.service.Update(c.Request.Context(), todoID, userID, req)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, result, "Todo updated successfully")
}

// Delete godoc
// @Summary Delete todo
// @Tags todos
// @Security BearerAuth
// @Produce json
// @Param id path string true "Todo ID"
// @Success 200 {object} response.Response
// @Router /todos/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	userID := c.GetString("userID")
	todoID := c.Param("id")
	err := h.service.Delete(c.Request.Context(), todoID, userID)
	if err != nil {
		c.Error(err)
		return
	}
	response.Success(c, nil, "Todo deleted successfully")
}
