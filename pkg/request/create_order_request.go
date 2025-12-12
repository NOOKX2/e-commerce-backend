package request

type CreateOrderRequest struct {
	ShippingAddress string            `json:"shippingAddress" validate:"required"`
	Items           []CartItemRequest `json:"items" validate:"required,min=1"`
	PaymentIntentID string            `json:"paymentIntentId" validate:"required"`
}