package request

type ProductRequest struct {
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	ImageUrl    string  `json:"image_url"`
}

type UpdateProductRequest struct {
	Name        string
	Price       float64
	Description string
}
