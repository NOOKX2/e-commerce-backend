package response

type UserResponse struct {
	ID     uint   `json:"ID"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Status string `json:"status"`
}
