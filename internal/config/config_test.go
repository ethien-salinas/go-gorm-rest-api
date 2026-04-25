package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want Config
	}{
		{
			name: "all vars set",
			env: map[string]string{
				"DB_HOST":      "localhost",
				"DB_USER":      "postgres",
				"DB_PASSWORD":  "secret",
				"DB_NAME":      "mydb",
				"DB_PORT":      "5432",
				"DB_SSLMODE":   "disable",
				"PORT":         "8080",
				"AUTO_MIGRATE": "false",
			},
			want: Config{
				DBHost:      "localhost",
				DBUser:      "postgres",
				DBPassword:  "secret",
				DBName:      "mydb",
				DBPort:      "5432",
				DBSSLMode:   "disable",
				Port:        "8080",
				AutoMigrate: false,
			},
		},
		{
			name: "PORT not set defaults to 3000",
			env: map[string]string{
				"PORT": "",
			},
			want: Config{
				Port: "3000",
			},
		},
		{
			name: "AUTO_MIGRATE true",
			env: map[string]string{
				"AUTO_MIGRATE": "true",
				"PORT":         "",
			},
			want: Config{
				AutoMigrate: true,
				Port:        "3000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			got := Load()
			assert.Equal(t, tt.want, got)
		})
	}
}
