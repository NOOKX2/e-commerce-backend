package request

type ProductRequest struct {
	Name        string  `json:"name"`
	SKU         string  `json:"sku" gorm:"uniqueIndex;not null"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	ImageUrl    string  `json:"image_url"`
	Category    string  `json:"category"`
	Quantity    uint    `json:"quantity"`
}

type UpdateProductRequest struct {
	Name        string
	Price       float64
	Description string
}
