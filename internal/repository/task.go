package repository

import (
	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) FindAll() ([]models.Task, error) {
	var tasks []models.Task
	return tasks, r.db.Preload("User").Find(&tasks).Error
}

func (r *TaskRepository) FindByID(id string) (models.Task, error) {
	var task models.Task
	return task, r.db.Preload("User").First(&task, id).Error
}

func (r *TaskRepository) Create(task *models.Task) error {
	return r.db.Create(task).Error
}

func (r *TaskRepository) Update(task *models.Task) error {
	return r.db.Save(task).Error
}

func (r *TaskRepository) Delete(task *models.Task) error {
	return r.db.Delete(task).Error
}
