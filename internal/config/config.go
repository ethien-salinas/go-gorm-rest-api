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
	LogToFile   bool   // LOG_TO_FILE=true habilita escritura a archivos rotativos diarios
	LogDir      string // LOG_DIR directorio para archivos de log (default: "logs")
}

// Load returns a [Config] populated from environment variables.
// PORT defaults to "3000" and LOG_DIR defaults to "logs" if not set.
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
		LogToFile:   os.Getenv("LOG_TO_FILE") == "true",
		LogDir:      envOrDefault("LOG_DIR", "logs"),
	}
}

func envOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
