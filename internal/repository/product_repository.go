package repository

import (
	"context"

	"github.com/NOOKX2/e-commerce-backend/internal/domain"
	"gorm.io/gorm"
)

type ProductRepositoryInterface interface {
	Create(ctx context.Context, product *domain.Product) error
	GetAllProduct() ([]domain.Product, error)
	GetProductByID(id uint) (*domain.Product, error)
	UpdateProduct(product *domain.Product) (error)
	DeleteProduct(id uint) (error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepositoryInterface {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	result := r.db.WithContext(ctx).Create(product)
	return result.Error
}

func (r *productRepository) GetAllProduct() ([]domain.Product, error) {
	var products []domain.Product
	result := r.db.Find(&products)

	if result.Error != nil {
		return nil, result.Error
	}

	return products, nil
}

func (r *productRepository) GetProductByID(id uint) (*domain.Product, error) {
	var product domain.Product
	result := r.db.Where("id = ?", id).First(&product)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound { // user not found
			return nil, nil
		}

		return nil, result.Error
	}
	return &product, nil
}


func (r *productRepository) UpdateProduct(product *domain.Product) (error) {
	result := r.db.Save(product)
	return result.Error
}


func (r productRepository) DeleteProduct(id uint) (error) {
	result := r.db.Delete(&domain.Product{}, id)
	return result.Error
}