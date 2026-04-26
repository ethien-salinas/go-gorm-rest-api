// Package repository encapsulates GORM queries for each entity type.
package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"gorm.io/gorm"
)

// UserRepository provides database operations for [models.User] records.
type UserRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewUserRepository returns a [UserRepository] backed by the given database connection.
func NewUserRepository(db *gorm.DB, logger *slog.Logger) *UserRepository {
	return &UserRepository{db: db, logger: logger}
}

// FindAll returns all users from the database with their associated tasks preloaded.
func (r *UserRepository) FindAll(ctx context.Context) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Preload("Tasks").Find(&users).Error; err != nil {
		r.logger.Error("repository: failed to find all users", "error", err)
		return nil, fmt.Errorf("userRepository.FindAll: %w", err)
	}
	return users, nil
}

// FindByID returns the user identified by id, with tasks preloaded.
func (r *UserRepository) FindByID(ctx context.Context, id string) (models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Preload("Tasks").First(&user, id).Error; err != nil {
		r.logger.Error("repository: failed to find user", "id", id, "error", err)
		return models.User{}, fmt.Errorf("userRepository.FindByID: %w", err)
	}
	return user, nil
}

// Create inserts a new user record into the database.
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		r.logger.Error("repository: failed to create user", "error", err)
		return fmt.Errorf("userRepository.Create: %w", err)
	}
	return nil
}

// Update persists changes to an existing user record.
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		r.logger.Error("repository: failed to update user", "id", user.ID, "error", err)
		return fmt.Errorf("userRepository.Update: %w", err)
	}
	return nil
}

// Delete performs a soft-delete on the given user record.
func (r *UserRepository) Delete(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Delete(user).Error; err != nil {
		r.logger.Error("repository: failed to delete user", "id", user.ID, "error", err)
		return fmt.Errorf("userRepository.Delete: %w", err)
	}
	return nil
}

// Count returns the total number of non-deleted user records.
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error; err != nil {
		r.logger.Error("repository: failed to count users", "error", err)
		return 0, fmt.Errorf("userRepository.Count: %w", err)
	}
	return count, nil
}
