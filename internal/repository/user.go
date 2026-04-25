package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewUserRepository(db *gorm.DB, logger *slog.Logger) *UserRepository {
	return &UserRepository{db: db, logger: logger}
}

func (r *UserRepository) FindAll(ctx context.Context) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Preload("Tasks").Find(&users).Error; err != nil {
		r.logger.Error("repository: failed to find all users", "error", err)
		return nil, fmt.Errorf("userRepository.FindAll: %w", err)
	}
	return users, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Tasks").First(&user, id).Error; err != nil {
		r.logger.Error("repository: failed to find user", "id", id, "error", err)
		return models.User{}, fmt.Errorf("userRepository.FindByID: %w", err)
	}
	return user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		r.logger.Error("repository: failed to create user", "error", err)
		return fmt.Errorf("userRepository.Create: %w", err)
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		r.logger.Error("repository: failed to update user", "id", user.ID, "error", err)
		return fmt.Errorf("userRepository.Update: %w", err)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Delete(user).Error; err != nil {
		r.logger.Error("repository: failed to delete user", "id", user.ID, "error", err)
		return fmt.Errorf("userRepository.Delete: %w", err)
	}
	return nil
}
