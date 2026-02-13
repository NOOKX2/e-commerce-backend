package request

type ProductRequest struct {
	Name        string  `json:"name"`
	SKU         string  `json:"sku" gorm:"uniqueIndex;not null"`
	Price       float64 `json:"price"`
	CostPrice   float64 `json:"costPrice"`
	Description string  `json:"description"`
	ImageUrl    string  `json:"image_url"`
	Category    string  `json:"category"`
	Quantity    uint    `json:"quantity"`
	ImageHash   string  `json:"image_hash"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	SalePrice   float64 `json:"sale_price"`
	Quantity    uint    `json:"quantity"`
	Status      string  `json:"status"`
	CostPrice   float64 `json:"cost_price"`
}
