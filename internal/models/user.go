// Package models defines the GORM entity types for the application.
package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a registered user with an associated list of tasks.
// PasswordHash is excluded from JSON responses via json:"-".
type User struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	FirstName    string `gorm:"size:100;not null"              json:"first_name"`
	LastName     string `gorm:"size:100;not null"              json:"last_name"`
	Email        string `gorm:"size:100;not null;uniqueIndex"  json:"email"`
	PasswordHash string `gorm:"size:255;not null;default:''"   json:"-"` // bcrypt hash; never serialized
	Tasks        []Task `gorm:"foreignKey:UserID"              json:"tasks"`
}
