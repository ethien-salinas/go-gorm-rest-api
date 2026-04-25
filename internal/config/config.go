// Package config reads application configuration from environment variables.
package config

import "os"

// Config holds the application configuration loaded from environment variables.
type Config struct {
	DBHost      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBPort      string
	DBSSLMode   string
	Port        string
	AutoMigrate bool
}

// Load returns a [Config] populated from environment variables.
// The PORT variable defaults to "3000" if not set.
func Load() Config {
	return Config{
		DBHost:      os.Getenv("DB_HOST"),
		DBUser:      os.Getenv("DB_USER"),
		DBPassword:  os.Getenv("DB_PASSWORD"),
		DBName:      os.Getenv("DB_NAME"),
		DBPort:      os.Getenv("DB_PORT"),
		DBSSLMode:   os.Getenv("DB_SSLMODE"),
		Port:        envOrDefault("PORT", "3000"),
		AutoMigrate: os.Getenv("AUTO_MIGRATE") == "true",
	}
}

func envOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
