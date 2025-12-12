package response

type UserResponse struct {
	ID    uint   
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}
