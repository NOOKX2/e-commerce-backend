package domain

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email        string    `json:"email" gorm:"unique;not null;size:255"`
	PasswordHash string    `json:"password" gorm:"not null"`
	DisplayName  string    `json:"display_name" gorm:"not null;size:100"`
	Role         UserRole    `json:"role" gorm:"not null;size:50;default:'buyer'"`
	CreatedAt    time.Time `json:"created_at"`
}
