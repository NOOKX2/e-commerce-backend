package response

import (
	"time"

	"github.com/NOOKX2/e-commerce-backend/internal/domain"
)

type ProductResponse struct {
	ID          uint      `json:"id"` 
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	ImageURL    string    `json:"imageUrl"`  
	SellerID    uint      `json:"sellerId"`  
	CreatedAt   time.Time `json:"createdAt"` 
	UpdatedAt   time.Time `json:"updatedAt"` 
	Slug        string    `json:"slug"`      
}

func ToProductResponse(product domain.Product) ProductResponse {
	return ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Category:    product.Category,
		ImageURL:    product.ImageURL,
		SellerID:    product.SellerID,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

func ToProductResponses(products []domain.Product) []ProductResponse {
	responses := make([]ProductResponse, 0, len(products))
	for _, product := range products {
		responses = append(responses, ToProductResponse(product))
	}

	return responses
}