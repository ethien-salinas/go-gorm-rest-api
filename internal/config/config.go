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
	RateLimitRPS   int    // RATE_LIMIT_RPS peticiones por segundo por IP (default: 10)
	RateLimitBurst int    // RATE_LIMIT_BURST capacidad máxima del bucket (default: 20)

	// Política de contraseñas (PCI DSS v4.0 + NIST SP 800-63B)
	PasswordHistoryCount          int // PASSWORD_HISTORY_COUNT          contraseñas anteriores a recordar (default: 12)
	PasswordMaxAgeDays            int // PASSWORD_MAX_AGE_DAYS            días antes de forzar cambio (default: 90, reservado)
	AccountLockoutThreshold       int // ACCOUNT_LOCKOUT_THRESHOLD        intentos fallidos antes de bloquear (default: 5)
	AccountLockoutDurationMinutes int // ACCOUNT_LOCKOUT_DURATION_MINUTES minutos de bloqueo (default: 30)
}

// Load returns a [Config] populated from environment variables.
// Defaults: PORT=3000, LOG_DIR=logs, JWT_EXPIRY_HOURS=24, RATE_LIMIT_RPS=10, RATE_LIMIT_BURST=20.
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
		RateLimitRPS:   envIntOrDefault("RATE_LIMIT_RPS", 10),
		RateLimitBurst: envIntOrDefault("RATE_LIMIT_BURST", 20),

		PasswordHistoryCount:          envIntOrDefault("PASSWORD_HISTORY_COUNT", 12),
		PasswordMaxAgeDays:            envIntOrDefault("PASSWORD_MAX_AGE_DAYS", 90),
		AccountLockoutThreshold:       envIntOrDefault("ACCOUNT_LOCKOUT_THRESHOLD", 5),
		AccountLockoutDurationMinutes: envIntOrDefault("ACCOUNT_LOCKOUT_DURATION_MINUTES", 30),
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
