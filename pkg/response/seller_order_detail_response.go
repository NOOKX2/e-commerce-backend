package response

type SellerOrderDetailResponse struct {
	OrderID         uint                 `json:"orderId"`
	Status          string               `json:"status"`
	PlacedAt        string               `json:"placedAt"`
	CustomerInfo    CustomerInfoDTO      `json:"customerInfo"`
	ShippingAddress ShippingAddressDTO   `json:"shippingAddress"`
	Items           []SellerOrderItemDTO `json:"items"`

	SellerSubtotal float64 `json:"sellerSubtotal"`
}

type CustomerInfoDTO struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}

type ShippingAddressDTO struct {
	AddressLine string `json:"addressLine"`
}

type SellerOrderItemDTO struct {
	ProductID uint    `json:"productId"`
	Name      string  `json:"name"`
	SKU       string  `json:"sku"`
	ImageURL  string  `json:"imageUrl"`
	Price     float64 `json:"price"`
	Quantity  uint    `json:"quantity"`
	Total     float64 `json:"total"`
}
