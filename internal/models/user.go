package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email            string    `json:"email" gorm:"unique;not null;size:255"`
	PasswordHash     string    `json:"password" gorm:"not null"`
	Name             string    `json:"name" gorm:"not null;size:100"`
	Role             UserRole  `json:"role" gorm:"not null;size:50;default:'buyer'"`
	CreatedAt        time.Time `json:"createdAt"`
	StripeCustomerID string    `gorm:"column:stripe_customer_id"`
}
