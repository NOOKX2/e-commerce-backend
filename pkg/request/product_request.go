package request

type ProductRequest struct {
	Name        string
	Price       float64
	Description string
	SellerID    uint
}

type UpdateProductRequest struct {
	Name        string
	Price       float64
	Description string
}
