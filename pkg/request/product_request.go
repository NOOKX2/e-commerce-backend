package request

type ProductRequest struct {
    Name        string  `json:"name"`
    SKU         string  `json:"sku"`
    Price       float64 `json:"price"`
    CostPrice   float64 `json:"costPrice"`
    Description string  `json:"description"`
    ImageUrl    string  `json:"imageUrl"`    
    Category    string  `json:"category"`
    Quantity    uint    `json:"quantity"`
    ImageHash   string  `json:"imageHash"`   
}

type UpdateProductRequest struct {
    Name        string  `json:"name"`
    Price       float64 `json:"price"`
    Description string  `json:"description"`
    Category    string  `json:"category"`
    SalePrice   float64 `json:"salePrice"`   
    Quantity    uint    `json:"quantity"`
    Status      string  `json:"status"`
    CostPrice   float64 `json:"costPrice"`   
    ImageUrl    string  `json:"imageUrl"`    
    ImageHash   string  `json:"imageHash"`   
}