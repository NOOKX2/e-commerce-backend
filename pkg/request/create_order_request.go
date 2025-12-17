package request

import (
	"github.com/NOOKX2/e-commerce-backend/internal/models"
)

type CreateOrderRequest struct {
	ShippingAddress *models.ShippingAddress `json:"shippingAddress" validate:"required"`
	Items           []CartItemRequest         `json:"items" validate:"required,min=1,dive"`

	PaymentIntentID string `json:"paymentIntentId" validate:"required"`
}

type CreatePaymentIntentRequest struct {
	Items []CartItemRequest `json:"items" validate:"required,min=1"`
}

type CartItemRequest struct {
	ProductID uint `json:"productId" validate:"required,gt=0"`
	Quantity  uint `json:"quantity" validate:"required,gt=0"`
}
