package repository

import (
	"context"
	"strconv"
	"strings"

	"github.com/NOOKX2/e-commerce-backend/internal/domain"
	"gorm.io/gorm"
)

type ProductRepositoryInterface interface {
	Create(ctx context.Context, product *domain.Product) error
	GetAllProduct(category string, price string, sort string, limit uint, offset uint) ([]domain.Product, error)
	GetProductByID(id uint) (*domain.Product, error)
	UpdateProduct(product *domain.Product) error
	DeleteProduct(id uint) error
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

func (r *productRepository) GetAllProduct(category string, price string, sort string, limit uint, offset uint) ([]domain.Product, error) {
	products := make([]domain.Product, 0)
	query := r.db.Model(&domain.Product{})

	if category != "" {
		categories := strings.Split(category, ",")
		query = query.Where("category IN ?", categories)
	}

	if price != "" {
		prices := strings.Split(price, ",")
		if len(prices) == 2 {
			minPrice, errMin := strconv.Atoi(prices[0])
			maxPrice, errMax := strconv.Atoi(prices[1])
			if errMin == nil && errMax == nil {
				query = query.Where("price BETWEEN ? AND ?", minPrice, maxPrice)
			}
		}
	}

	if sort != "" {
		sortParts := strings.Split(sort, "_")
		if len(sortParts) == 2 {
			column := sortParts[0]
			direction := sortParts[1]

			if (column == "price" || column == "created_at") && (direction == "asc" || direction == "desc") {
				query = query.Order(column + " " + direction)
			}
		}
	} else {
		query = query.Order("created_at desc")
	}

	query = query.Limit(int(limit)).Offset(int(offset))

	if err := query.Find(&products).Error; err != nil {
		return nil, err
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

func (r *productRepository) UpdateProduct(product *domain.Product) error {
	result := r.db.Save(product)
	return result.Error
}

func (r productRepository) DeleteProduct(id uint) error {
	result := r.db.Delete(&domain.Product{}, id)
	return result.Error
}
