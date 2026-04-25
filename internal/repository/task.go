package repository

import (
	"log/slog"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewTaskRepository(db *gorm.DB, logger *slog.Logger) *TaskRepository {
	return &TaskRepository{db: db, logger: logger}
}

func (r *TaskRepository) FindAll() ([]models.Task, error) {
	var tasks []models.Task
	err := r.db.Preload("User").Find(&tasks).Error
	if err != nil {
		r.logger.Error("repository: failed to find all tasks", "error", err)
	}
	return tasks, err
}

func (r *TaskRepository) FindByID(id string) (models.Task, error) {
	var task models.Task
	err := r.db.Preload("User").First(&task, id).Error
	if err != nil {
		r.logger.Error("repository: failed to find task", "id", id, "error", err)
	}
	return task, err
}

func (r *TaskRepository) Create(task *models.Task) error {
	err := r.db.Create(task).Error
	if err != nil {
		r.logger.Error("repository: failed to create task", "error", err)
	}
	return err
}

func (r *TaskRepository) Update(task *models.Task) error {
	err := r.db.Save(task).Error
	if err != nil {
		r.logger.Error("repository: failed to update task", "id", task.ID, "error", err)
	}
	return err
}

func (r *TaskRepository) Delete(task *models.Task) error {
	err := r.db.Delete(task).Error
	if err != nil {
		r.logger.Error("repository: failed to delete task", "id", task.ID, "error", err)
	}
	return err
}
