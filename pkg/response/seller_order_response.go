package response

type SellerOrderResponse struct {
    ID       uint `json:"id"`
    Product  string `json:"product"`
    Customer string `json:"customer"`
    Date     string `json:"date"`
    Amount   float64 `json:"amount"` 
    Status   string `json:"status"`
}