package models

import "gorm.io/gorm"

type User struct {
	gorm.Model

	FirstName string `gorm:"size:100;not null" json:"first_name"`
	LastName  string `gorm:"size:100;not null" json:"last_name"`
	Email     string `gorm:"size:100;not null;uniqueIndex" json:"email"`
	Tasks     []Task `gorm:"foreignKey:UserID" json:"tasks"`
}
