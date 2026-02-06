package models

import "time"

type Category struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Slug      string    `gorm:"uniqueIndex;not null" json:"slug"` // กันซ้ำที่ระดับ DB
	CreatedAt time.Time `json:"createdAt"`
}
