// Package config reads application configuration from environment variables.
package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	DBHost         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBPort         string
	DBSSLMode      string
	Port           string
	AutoMigrate    bool
	LogToFile      bool   // LOG_TO_FILE=true habilita escritura a archivos rotativos diarios
	LogDir         string // LOG_DIR directorio para archivos de log (default: "logs")
	JWTSecret      string // JWT_SECRET clave de firma para los tokens
	JWTExpiryHours int    // JWT_EXPIRY_HOURS duración del token en horas (default: 24)
}

// Load returns a [Config] populated from environment variables.
// PORT defaults to "3000", LOG_DIR to "logs" and JWT_EXPIRY_HOURS to 24 if not set.
func Load() Config {
	return Config{
		DBHost:         os.Getenv("DB_HOST"),
		DBUser:         os.Getenv("DB_USER"),
		DBPassword:     os.Getenv("DB_PASSWORD"),
		DBName:         os.Getenv("DB_NAME"),
		DBPort:         os.Getenv("DB_PORT"),
		DBSSLMode:      os.Getenv("DB_SSLMODE"),
		Port:           envOrDefault("PORT", "3000"),
		AutoMigrate:    os.Getenv("AUTO_MIGRATE") == "true",
		LogToFile:      os.Getenv("LOG_TO_FILE") == "true",
		LogDir:         envOrDefault("LOG_DIR", "logs"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		JWTExpiryHours: envIntOrDefault("JWT_EXPIRY_HOURS", 24),
	}
}

func envOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func envIntOrDefault(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultValue
}
