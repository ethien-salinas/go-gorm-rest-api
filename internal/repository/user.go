package repository

import (
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

func (r *UserRepository) FindAll() ([]models.User, error) {
	var users []models.User
	err := r.db.Preload("Tasks").Find(&users).Error
	if err != nil {
		r.logger.Error("repository: failed to find all users", "error", err)
	}
	return users, err
}

func (r *UserRepository) FindByID(id string) (models.User, error) {
	var user models.User
	err := r.db.Preload("Tasks").First(&user, id).Error
	if err != nil {
		r.logger.Error("repository: failed to find user", "id", id, "error", err)
	}
	return user, err
}

func (r *UserRepository) Create(user *models.User) error {
	err := r.db.Create(user).Error
	if err != nil {
		r.logger.Error("repository: failed to create user", "error", err)
	}
	return err
}

func (r *UserRepository) Update(user *models.User) error {
	err := r.db.Save(user).Error
	if err != nil {
		r.logger.Error("repository: failed to update user", "id", user.ID, "error", err)
	}
	return err
}

func (r *UserRepository) Delete(user *models.User) error {
	err := r.db.Delete(user).Error
	if err != nil {
		r.logger.Error("repository: failed to delete user", "id", user.ID, "error", err)
	}
	return err
}
