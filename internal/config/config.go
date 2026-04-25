package config

import "os"

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
