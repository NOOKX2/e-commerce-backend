package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	UserID                uint        `json:"userId"`
	Status                string      `json:"status"`
	TotalAmount           float64     `json:"totalAmount" gorm:"type:numeric(10,2)"`
	ShippingAddress       string      `json:"shippingAddress"`
	StripePaymentIntentID *string     `json:"stripePaymentIntentId"`
	CreatedAt             time.Time   `json:"createdAt"`
	UpdatedAt             time.Time   `json:"updatedAt"`
	Items                 []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

type OrderItem struct {
	gorm.Model
	OrderID         uint    `json:"orderId"`
	ProductID       uint    `json:"productId"`
	Product         Product `gorm:"foreignKey:ProductID" json:"product"`
	Quantity        uint    `json:"quantity"`
	PriceAtPurchase float64 `json:"priceAtPurchase" gorm:"type:numeric(10,2)"`
}




