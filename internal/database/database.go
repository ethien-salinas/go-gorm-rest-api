// Package database provides the PostgreSQL connection via GORM.
package database

import (
	"fmt"
	"log/slog"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect opens a GORM connection to the PostgreSQL database described by cfg.
// Returns an error if the connection cannot be established.
func Connect(cfg config.Config, logger *slog.Logger) (*gorm.DB, error) {
	logger.Info("connecting to database", "host", cfg.DBHost, "dbname", cfg.DBName)

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
		cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("database.Connect: %w", err)
	}

	logger.Info("database connection established")
	return db, nil
}
