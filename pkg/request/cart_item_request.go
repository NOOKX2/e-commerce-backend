package request

type CartItemRequest struct {
	ProductID uint `json:"productId" validate:"required,gt=0"`
	Quantity  uint `json:"quantity" validate:"required,gt=0"`
}