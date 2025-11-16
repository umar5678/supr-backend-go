package dto

import "errors"

type CreateTodoRequest struct {
	Title       string `json:"title" binding:"required,min=3"`
	Description string `json:"description" binding:"omitempty"`
}

func (r CreateTodoRequest) Validate() error {
	if len(r.Title) < 3 {
		return errors.New("title must be at least 3 characters")
	}
	return nil
}

type UpdateTodoRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=3"`
	Description *string `json:"description" binding:"omitempty"`
	Completed   *bool   `json:"completed" binding:"omitempty"`
}

func (r UpdateTodoRequest) Validate() error {
	if r.Title != nil && len(*r.Title) < 3 {
		return errors.New("title must be at least 3 characters if provided")
	}
	return nil
}
