package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"gorm.io/gorm"
)

// TaskRepository provides database operations for [models.Task] records.
type TaskRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewTaskRepository returns a [TaskRepository] backed by the given database connection.
func NewTaskRepository(db *gorm.DB, logger *slog.Logger) *TaskRepository {
	return &TaskRepository{db: db, logger: logger}
}

// FindAll returns all tasks from the database with their owning user preloaded.
func (r *TaskRepository) FindAll(ctx context.Context) ([]models.Task, error) {
	var tasks []models.Task
	if err := r.db.WithContext(ctx).Preload("User").Find(&tasks).Error; err != nil {
		r.logger.Error("repository: failed to find all tasks", "error", err)
		return nil, fmt.Errorf("taskRepository.FindAll: %w", err)
	}
	return tasks, nil
}

// FindByID returns the task identified by id, with the owning user preloaded.
func (r *TaskRepository) FindByID(ctx context.Context, id string) (models.Task, error) {
	var task models.Task
	if err := r.db.WithContext(ctx).Preload("User").First(&task, id).Error; err != nil {
		r.logger.Error("repository: failed to find task", "id", id, "error", err)
		return models.Task{}, fmt.Errorf("taskRepository.FindByID: %w", err)
	}
	return task, nil
}

// Create inserts a new task record into the database.
func (r *TaskRepository) Create(ctx context.Context, task *models.Task) error {
	if err := r.db.WithContext(ctx).Create(task).Error; err != nil {
		r.logger.Error("repository: failed to create task", "error", err)
		return fmt.Errorf("taskRepository.Create: %w", err)
	}
	return nil
}

// Update persists changes to an existing task record.
func (r *TaskRepository) Update(ctx context.Context, task *models.Task) error {
	if err := r.db.WithContext(ctx).Save(task).Error; err != nil {
		r.logger.Error("repository: failed to update task", "id", task.ID, "error", err)
		return fmt.Errorf("taskRepository.Update: %w", err)
	}
	return nil
}

// Delete performs a soft-delete on the given task record.
func (r *TaskRepository) Delete(ctx context.Context, task *models.Task) error {
	if err := r.db.WithContext(ctx).Delete(task).Error; err != nil {
		r.logger.Error("repository: failed to delete task", "id", task.ID, "error", err)
		return fmt.Errorf("taskRepository.Delete: %w", err)
	}
	return nil
}
