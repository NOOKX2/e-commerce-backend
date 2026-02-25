package response

type SellerCustomerResponse struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	TotalOrders   int     `json:"totalOrders"`
	TotalSpent    float64 `json:"totalSpent"`
	LastOrderDate string  `json:"lastOrderDate"`
	Location      string  `json:"location"`
	Status        string  `json:"status"`
}