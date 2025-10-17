package domain

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email        string    `json:"email" gorm:"unique;not null"`
	PasswordHash string    `gorm:"not null"`
	DisplayName  string    `json:"display_name" gorm:"not null"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}
