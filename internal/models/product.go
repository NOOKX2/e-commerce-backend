package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	SKU         string  `json:"sku" gorm:"uniqueIndex;not null"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	ImageURL    string  `json:"imageUrl"`
	SellerID    uint    `json:"sellerId"`
	Slug        string  `json:"slug"`
	Quantity    uint    `json:"quantity" gorm:"default:0"`
}
