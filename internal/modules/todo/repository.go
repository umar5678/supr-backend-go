package todos

import (
	"github.com/umar5678/go-backend/internal/models"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(todo *models.Todo) error {
	return r.db.Create(todo).Error
}

func (r *Repository) FindByID(id, userID string) (*models.Todo, error) {
	var todo models.Todo
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&todo).Error
	return &todo, err
}

func (r *Repository) FindAll(userID string) ([]models.Todo, error) {
	var todos []models.Todo
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&todos).Error
	return todos, err
}

func (r *Repository) Update(todo *models.Todo) error {
	return r.db.Save(todo).Error
}

func (r *Repository) Delete(id, userID string) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Todo{}).Error
}
