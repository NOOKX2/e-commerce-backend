package models

import "time"

type Media struct {
	ID        uint   `gorm:"primaryKey"`
	Hash      string `gorm:"uniqueIndex"`
	PublicURL string `gorm:"not null"`
	CreatedAt time.Time
}
