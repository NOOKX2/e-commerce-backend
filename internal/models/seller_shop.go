package models

import "gorm.io/gorm"

// SellerShop holds seller-facing shop profile and payout info (one per seller user).
type SellerShop struct {
	gorm.Model
	SellerID        uint   `json:"sellerId" gorm:"uniqueIndex;not null"`
	ShopName        string `json:"shopName" gorm:"size:200"`
	Description     string `json:"description" gorm:"type:text"`
	LogoURL         string `json:"logoUrl" gorm:"size:1024"`
	PickupAddress   string `json:"pickupAddress" gorm:"type:text"`
	BankName        string `json:"bankName" gorm:"size:120"`
	AccountNumber   string `json:"accountNumber" gorm:"size:64"`
	AccountHolder   string `json:"accountHolder" gorm:"size:200"`
}

func (SellerShop) TableName() string {
	return "seller_shops"
}
