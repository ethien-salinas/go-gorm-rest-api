// Package database provides the PostgreSQL connection via GORM.
package database

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ethien-salinas/go-gorm-rest-api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect opens a GORM connection to the PostgreSQL database described by cfg.
// The connection is terminated if it cannot be established.
func Connect(cfg config.Config, logger *slog.Logger) *gorm.DB {
	logger.Info("connecting to database", "host", cfg.DBHost, "dbname", cfg.DBName)

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	logger.Info("database connection established")
	return db
}
