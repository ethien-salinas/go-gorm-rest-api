package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	FirstName string `gorm:"size:100;not null"             json:"first_name"`
	LastName  string `gorm:"size:100;not null"             json:"last_name"`
	Email     string `gorm:"size:100;not null;uniqueIndex"  json:"email"`
	Tasks     []Task `gorm:"foreignKey:UserID"             json:"tasks"`
}
