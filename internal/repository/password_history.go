package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/models"
	"gorm.io/gorm"
)

// PasswordHistoryRepository provides database operations for [models.PasswordHistory] records.
type PasswordHistoryRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewPasswordHistoryRepository returns a [PasswordHistoryRepository] backed by the given database connection.
func NewPasswordHistoryRepository(db *gorm.DB, logger *slog.Logger) *PasswordHistoryRepository {
	return &PasswordHistoryRepository{db: db, logger: logger}
}

// Create inserts a new password history record.
func (r *PasswordHistoryRepository) Create(ctx context.Context, h *models.PasswordHistory) error {
	if err := r.db.WithContext(ctx).Create(h).Error; err != nil {
		r.logger.Error("repository: failed to create password history", "userID", h.UserID, "error", err)
		return fmt.Errorf("passwordHistoryRepository.Create: %w", err)
	}
	return nil
}

// FindRecentByUserID returns the n most recent password hashes for the given user,
// ordered newest first.
func (r *PasswordHistoryRepository) FindRecentByUserID(ctx context.Context, userID uint, limit int) ([]models.PasswordHistory, error) {
	var history []models.PasswordHistory
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error; err != nil {
		r.logger.Error("repository: failed to find password history", "userID", userID, "error", err)
		return nil, fmt.Errorf("passwordHistoryRepository.FindRecentByUserID: %w", err)
	}
	return history, nil
}
