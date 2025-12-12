package repository

import (
	"context"
	"strconv"
	"strings"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"gorm.io/gorm"
)

type ProductRepositoryInterface interface {
	Create(ctx context.Context, product *models.Product) error
	GetAllProduct(category string, price string, sort string, limit uint, offset uint) ([]models.Product, error)
	GetProductByID(id uint) (*models.Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*models.Product, error)
	UpdateProduct(product *models.Product) error
	DeleteProduct(id uint) error
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepositoryInterface {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *models.Product) error {
	result := r.db.WithContext(ctx).Create(product)
	return result.Error
}

func (r *productRepository) GetAllProduct(category string, price string, sort string, limit uint, offset uint) ([]models.Product, error) {
	products := make([]models.Product, 0)
	query := r.db.Model(&models.Product{})

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

func (r *productRepository) GetProductByID(id uint) (*models.Product, error) {
	var product models.Product
	result := r.db.Where("id = ?", id).First(&product)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound { // user not found
			return nil, nil
		}

		return nil, result.Error
	}
	return &product, nil
}

func (r *productRepository) GetProductBySlug(ctx context.Context, slug string) (*models.Product, error) {
	var product models.Product
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&product).Error
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) UpdateProduct(product *models.Product) error {
	result := r.db.Save(product)
	return result.Error
}

func (r productRepository) DeleteProduct(id uint) error {
	result := r.db.Delete(&models.Product{}, id)
	return result.Error
}
