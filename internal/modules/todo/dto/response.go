package dto

import (
	"time"

	"github.com/umar5678/go-backend/internal/models"
)

type TodoResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToTodoResponse(todo *models.Todo) *TodoResponse {
	return &TodoResponse{
		ID:          todo.ID,
		Title:       todo.Title,
		Description: todo.Description,
		Completed:   todo.Completed,
		UserID:      todo.UserID,
		CreatedAt:   todo.CreatedAt,
		UpdatedAt:   todo.UpdatedAt,
	}
}

func ToTodoResponses(todos []models.Todo) []*TodoResponse {
	responses := make([]*TodoResponse, len(todos))
	for i, todo := range todos {
		responses[i] = ToTodoResponse(&todo)
	}
	return responses
}
