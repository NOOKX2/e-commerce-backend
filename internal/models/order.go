package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	UserID      uint    `json:"userId"`
	Status      string  `json:"status"`
	TotalAmount float64 `json:"totalAmount" gorm:"type:numeric(10,2)"`

	// ข้อมูลที่อยู่จัดส่ง (Flattened fields) เพื่อเป็น Snapshot ของออเดอร์นั้นๆ
	ShippingEmail         string `json:"shippingEmail" gorm:"column:shipping_email"`
	ShippingReceiverName  string `json:"shippingReceiverName" gorm:"column:shipping_receiver_name"`
	ShippingPhoneNumber   string `json:"shippingPhoneNumber" gorm:"column:shipping_phone_number"`
	ShippingStreetAddress string `json:"shippingStreetAddress" gorm:"column:shipping_street_address"`
	ShippingSubDistrict   string `json:"shippingSubDistrict" gorm:"column:shipping_sub_district"`
	ShippingDistrict      string `json:"shippingDistrict" gorm:"column:shipping_district"`
	ShippingProvince      string `json:"shippingProvince" gorm:"column:shipping_province"`
	ShippingPostalCode    string `json:"shippingPostalCode" gorm:"column:shipping_postal_code"`

	StripePaymentIntentID *string     `json:"stripePaymentIntentId" gorm:"unique"`
	CreatedAt             time.Time   `json:"createdAt"`
	UpdatedAt             time.Time   `json:"updatedAt"`
	Items                 []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
	UserCardID            *uint       `json:"userCardId"`
}

type OrderItem struct {
	gorm.Model
	OrderID         uint    `json:"orderId"`
	ProductID       uint    `json:"productId"`
	Product         Product `gorm:"foreignKey:ProductID" json:"product"`
	Quantity        uint    `json:"quantity"`
	PriceAtPurchase float64 `json:"priceAtPurchase" gorm:"type:numeric(10,2)"`
}

// ShippingAddress ใช้เป็น DTO (Data Transfer Object) สำหรับรับค่าจาก Request
type ShippingAddress struct {
	Email         string `json:"email" validate:"required,email"`
	ReceiverName  string `json:"receiverName" validate:"required"`
	PhoneNumber   string `json:"phoneNumber" validate:"required"`
	StreetAddress string `json:"streetAddress" validate:"required"`
	SubDistrict   string `json:"subDistrict" validate:"required"`
	District      string `json:"district" validate:"required"`
	Province      string `json:"province" validate:"required"`
	PostalCode    string `json:"postalCode" validate:"required,numeric"`
}