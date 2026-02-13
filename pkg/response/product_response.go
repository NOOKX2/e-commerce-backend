package response

import (
	"time"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
)

type ProductResponse struct {
	ID          uint
	Name        string          `json:"name"`
	SKU         string          `json:"sku" gorm:"uniqueIndex;not null"`
	Description string          `json:"description"`
	Price       float64         `json:"price"`
	Category    models.Category `json:"category"`
	ImageURL    string          `json:"imageUrl"`
	SellerID    uint            `json:"sellerId"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	Slug        string          `json:"slug"`
	Quantity    uint            `json:"quantity"`
	Status      string          `json:"status"`
	CostPrice   float64         `json:"costPrice"`
	SalePrice   float64         `json:"salePrice"`
}

func ToProductResponse(product models.Product) ProductResponse {
	return ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		SKU:         product.SKU,
		Description: product.Description,
		Price:       product.Price,
		Category:    product.Category,
		ImageURL:    product.ImageURL,
		SellerID:    product.SellerID,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
		Slug:        product.Slug,
		Quantity:    product.Quantity,
		Status:      product.Status,
		CostPrice:   product.CostPrice,
		SalePrice:   product.SalePrice,
	}
}

func ToProductResponses(products []models.Product) []ProductResponse {
	responses := make([]ProductResponse, 0, len(products))
	for _, product := range products {
		responses = append(responses, ToProductResponse(product))
	}

	return responses
}
