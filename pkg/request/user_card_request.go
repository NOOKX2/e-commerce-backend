package request

type SaveCardRequest struct {
    PaymentMethodID string `json:"payment_method_id" validate:"required"`
    CardBrand       string `json:"brand"`
    LastFour        string `json:"last_four"`
    ExpiryMonth     int    `json:"expiry_month"`
    ExpiryYear      int    `json:"expiry_year"`
    SetDefault      bool   `json:"set_default"`
}