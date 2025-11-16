package todos

import (
	"context"
	"errors"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/todo/dto"
	"github.com/umar5678/go-backend/internal/utils/response"

	"gorm.io/gorm"
)

type Service interface {
	Create(ctx context.Context, userID string, req dto.CreateTodoRequest) (*dto.TodoResponse, error)
	GetAll(ctx context.Context, userID string) ([]*dto.TodoResponse, error)
	GetByID(ctx context.Context, id, userID string) (*dto.TodoResponse, error)
	Update(ctx context.Context, id, userID string, req dto.UpdateTodoRequest) (*dto.TodoResponse, error)
	Delete(ctx context.Context, id, userID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, userID string, req dto.CreateTodoRequest) (*dto.TodoResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	todo := &models.Todo{
		Title:       req.Title,
		Description: req.Description,
		UserID:      userID,
		Completed:   false,
	}

	if err := s.repo.Create(todo); err != nil {
		return nil, response.InternalServerError("Failed to create todo", err)
	}

	return dto.ToTodoResponse(todo), nil
}

func (s *service) GetAll(ctx context.Context, userID string) ([]*dto.TodoResponse, error) {
	todos, err := s.repo.FindAll(userID)
	if err != nil {
		return nil, response.InternalServerError("Failed to retrieve todos", err)
	}

	return dto.ToTodoResponses(todos), nil
}

func (s *service) GetByID(ctx context.Context, id, userID string) (*dto.TodoResponse, error) {
	todo, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NotFoundError("Todo")
		}
		return nil, response.InternalServerError("Failed to retrieve todo", err)
	}

	return dto.ToTodoResponse(todo), nil
}

func (s *service) Update(ctx context.Context, id, userID string, req dto.UpdateTodoRequest) (*dto.TodoResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, response.BadRequest(err.Error())
	}

	todo, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NotFoundError("Todo")
		}
		return nil, response.InternalServerError("Failed to retrieve todo", err)
	}

	if req.Title != nil {
		todo.Title = *req.Title
	}
	if req.Description != nil {
		todo.Description = *req.Description
	}
	if req.Completed != nil {
		todo.Completed = *req.Completed
	}

	if err := s.repo.Update(todo); err != nil {
		return nil, response.InternalServerError("Failed to update todo", err)
	}

	return dto.ToTodoResponse(todo), nil
}

func (s *service) Delete(ctx context.Context, id, userID string) error {
	_, err := s.repo.FindByID(id, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NotFoundError("Todo")
		}
		return response.InternalServerError("Failed to retrieve todo", err)
	}

	if err := s.repo.Delete(id, userID); err != nil {
		return response.InternalServerError("Failed to delete todo", err)
	}

	return nil
}
