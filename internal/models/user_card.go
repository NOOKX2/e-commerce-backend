package models

import "time"

type UserCard struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	UserID                uint      `gorm:"not null;index" json:"user_id"`
	CardUniqueKey         string    `json:"card_unique_key" gorm:"index:idx_user_card_key"`
	StripePaymentMethodID string    `gorm:"unique;not null" json:"payment_method_id"`
	CardBrand             string    `json:"brand"`
	LastFour              string    `json:"last_four"`
	ExpiryMonth           int       `json:"expiry_month"`
	ExpiryYear            int       `json:"expiry_year"`
	IsDefault             bool      `gorm:"default:false" json:"is_default"`
	CreatedAt             time.Time `json:"created_at"`
}
