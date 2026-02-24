package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	UserID                uint            `json:"userId"`
	Status                string          `json:"status"`
	TotalAmount           float64         `json:"totalAmount" gorm:"type:numeric(10,2)"`
	ShippingAddress       ShippingAddress `json:"shippingAddress" gorm:"embedded;embeddedPrefix:shipping_"`
	StripePaymentIntentID *string         `json:"stripePaymentIntentId" gorm:"unique"`
	CreatedAt             time.Time       `json:"createdAt"`
	UpdatedAt             time.Time       `json:"updatedAt"`
	Items                 []OrderItem     `gorm:"foreignKey:OrderID" json:"items"`
	UserCardID            *uint           `json:"userCardId"`
}

type OrderItem struct {
	gorm.Model
	OrderID         uint    `json:"orderId"`
	ProductID       uint    `json:"productId"`
	Product         Product `gorm:"foreignKey:ProductID" json:"product"`
	Quantity        uint    `json:"quantity"`
	PriceAtPurchase float64 `json:"priceAtPurchase" gorm:"type:numeric(10,2)"`
}

type ShippingAddress struct {
	Email         string `json:"email" gorm:"not null;default:''" validate:"required,email"`
	ReceiverName  string `json:"receiverName" gorm:"not null;default:''" validate:"required"`
	PhoneNumber   string `json:"phoneNumber" gorm:"not null;default:''" validate:"required"`
	StreetAddress string `json:"streetAddress" gorm:"not null;default:''" validate:"required"`
	SubDistrict   string `json:"subDistrict" gorm:"not null;default:''" validate:"required"`
	District      string `json:"district" gorm:"not null;default:''" validate:"required"`
	Province      string `json:"province" gorm:"not null;default:''" validate:"required"`
	PostalCode    string `json:"postalCode" gorm:"not null;default:''" validate:"required,numeric"`
}
