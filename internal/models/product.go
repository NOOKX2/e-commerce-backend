package models

import "gorm.io/gorm"

type Product struct {
    gorm.Model
    SKU            string        `json:"sku" gorm:"uniqueIndex;not null"`
    Name           string        `json:"name" gorm:"index"` 
    NormalizedName string        `json:"-" gorm:"type:varchar(255);index;not null"` 
    Description    string        `json:"description" gorm:"type:text"` 
    Price          float64       `json:"price" gorm:"not null"`
    CostPrice      float64       `json:"costPrice" gorm:"default:0"`
    SalePrice      float64       `json:"salePrice" gorm:"default:0"`
    
    // Relationship
    CategoryID     uint          `json:"categoryID" gorm:"index"`
    Category       Category      `json:"category" gorm:"foreignKey:CategoryID"`
    SellerID       uint          `json:"sellerId" gorm:"index"` 
    Seller         User          `json:"-" gorm:"foreignKey:SellerID"`
    
    // SEO & Inventory
    Slug           string        `json:"slug" gorm:"uniqueIndex;not null"`
    Quantity       uint          `json:"quantity" gorm:"default:0"`
    Status         ProductStatus `json:"status" gorm:"type:varchar(20);default:'active';index"`
    
    // Metrics
    TotalSales     int           `json:"totalSales" gorm:"default:0"`
    Rating         float64       `json:"rating" gorm:"default:0"`
    
    // Media
    ImageURL       string        `json:"imageUrl"`
    ImageHash      string        `json:"imageHash"`
}

type ProductStatus string

const (
	StatusActive   ProductStatus = "active"
	StatusInactive ProductStatus = "inactive"
	StatusDraft    ProductStatus = "draft"
	StatusArchived ProductStatus = "archived"
)
