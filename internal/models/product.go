package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	SKU            string   `json:"sku" gorm:"uniqueIndex;not null"`
	Name           string   `json:"name"`
	NormalizedName string   `gorm:"type:varchar(255);index;not null"`
	Description    string   `json:"description"`
	Price          float64  `json:"price"`
	CostPrice      float64  `json:"costPrice" gorm:"default:0"`
	CategoryID     uint     `json:"categoryID"`
	Category       Category `json:"category"`
	ImageURL       string   `json:"imageUrl"`
	SellerID       uint     `json:"sellerId"`
	Seller         User     `json:"-" gorm:"foreignKey:SellerID"`
	Slug           string   `json:"slug"`
	Quantity       uint     `json:"quantity" gorm:"default:0"`
	Status         string   `json:"status" gorm:"default:'active';size:20"`
	TotalSales     int      `json:"totalSales" gorm:"default:0"`
	Rating         float64  `json:"rating" gorm:"default:0"`
	ImageHash      string   `json:"imageHash"`
}
