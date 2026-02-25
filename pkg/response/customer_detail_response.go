package response

type CustomerDetailResponse struct {
	ID           uint            `json:"id"`
	Name         string          `json:"name"`
	Email        string          `json:"email"`
	PhoneNumber  string          `json:"phoneNumber"`
	Location     string          `json:"location"`
	JoinedDate   string          `json:"joinedDate"`
	TotalSpent   float64         `json:"totalSpent"`
	TotalOrders  int             `json:"totalOrders"`
	Status       string          `json:"status"`
	OrderHistory []CustomerOrder `json:"orderHistory"`
}

type CustomerOrder struct {
	OrderID    uint  `json:"orderId"`
	Date       string  `json:"date"`
	Total      float64 `json:"total"`
	Status     string  `json:"status"`
	ItemsCount int     `json:"itemsCount"`
}