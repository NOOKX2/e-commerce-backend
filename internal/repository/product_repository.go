package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"gorm.io/gorm"
)

type ProductRepositoryInterface interface {
	Create(ctx context.Context, product *models.Product) error
	GetAllProduct(category string, price string, sort string, limit uint, offset uint) ([]models.Product, int64, error)
	GetProductByID(ctx context.Context, id uint) (*models.Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*models.Product, error)
	UpdateProduct(product *models.Product) error
	DeleteProduct(id uint) error
	GetProductBySKU(ctx context.Context, sku string) (*models.Product, error)
	AddToStock(ctx context.Context, id uint, amount uint) error
	RemoveFromStock(ctx context.Context, id uint, amount uint) error
	GetProductsBySellerID(ctx context.Context, sellerID uint, page, limit int, search string) ([]models.Product, int64, error)
	GetOrCreateCategory(ctx context.Context, name string, slug string) (uint, error)
	GetAllCategories(ctx context.Context) ([]models.Category, error)
	GetProductByNormalizedName(ctx context.Context, sellerID uint, normalizedName string) (*models.Product, error)
	UpdateProductBySKU(ctx context.Context, sku string, sellerID uint, product *models.Product) (*models.Product, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error)
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

func (r *productRepository) GetAllProduct(category string, price string, sort string, limit uint, offset uint) ([]models.Product, int64, error) {
	var total int64
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

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
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

	if err := query.Where("status = ?", "active").Find(&products).Error; err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *productRepository) GetProductByID(ctx context.Context, id uint) (*models.Product, error) {
	var product models.Product
	result := r.db.Where("id = ?", id).First(&product)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, result.Error
	}
	return &product, nil
}

func (r *productRepository) GetProductBySlug(ctx context.Context, slug string) (*models.Product, error) {
	var product models.Product
	err := r.db.Preload("Category").WithContext(ctx).Where("slug = ?", slug).First(&product).Error
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

func (r *productRepository) GetProductBySKU(ctx context.Context, sku string) (*models.Product, error) {
	var product models.Product

	err := r.db.Preload("Category").WithContext(ctx).Where("sku = ?", sku).First(&product).Error

	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) AddToStock(ctx context.Context, id uint, amount uint) error {
	result := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("id = ?", id).
		Update("quantity", gorm.Expr("quantity + ?", amount))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("Product not found")
	}
	return nil
}

func (r *productRepository) RemoveFromStock(ctx context.Context, id uint, amount uint) error {
	var product models.Product
	if err := r.db.WithContext(ctx).Select("id, quantity").First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("Product not found")
		}
		return err
	}

	if uint(product.Quantity) < amount {
		return fmt.Errorf("Insufficient Stock: available %d, requested %d", product.Quantity, amount)
	}

	result := r.db.WithContext(ctx).Model(&models.Product{}).
		Where("id = ? AND quantity >= ?", id, amount).
		Update("quantity", gorm.Expr("quantity - ?", amount))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("Product not found")
	}
	return nil
}

func (r *productRepository) GetProductsBySellerID(ctx context.Context, sellerID uint, page, limit int, search string) ([]models.Product, int64, error) {
	var products []models.Product
	var total int64
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&models.Product{}).Where("seller_id = ?", sellerID)

	if search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(search)+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	result := query.Preload("Category").Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return products, total, nil
}

func (r *productRepository) GetOrCreateCategory(ctx context.Context, name string, slug string) (uint, error) {
	var category models.Category

	if err := r.db.WithContext(ctx).Where(models.Category{Slug: slug}).
		Attrs(models.Category{Name: name}).
		FirstOrCreate(&category).Error; err != nil {
		return 0, err
	}

	return category.ID, nil
}

func (r *productRepository) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	if err := r.db.WithContext(ctx).Order("name asc").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *productRepository) GetProductByNormalizedName(ctx context.Context, sellerID uint, normalizedName string) (*models.Product, error) {
	var product models.Product

	if err := r.db.Preload("Category").WithContext(ctx).
		Where("seller_id = ? AND normalized_name = ?", sellerID, normalizedName).
		First(&product).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) UpdateProductBySKU(ctx context.Context, sku string, sellerID uint, product *models.Product) (*models.Product, error) {
	var updatedProduct models.Product
	fmt.Println("update product")
	fmt.Printf("product category %v\n", product.Category)

	result := r.db.Debug().WithContext(ctx).
		Model(&models.Product{}).
		Where("sku = ? AND seller_id = ?", sku, sellerID).
		Updates(map[string]interface{}{
			"name":        product.Name,
			"description": product.Description,
			"price":       product.Price,
			"sale_price":  product.SalePrice,
			"cost_price":  product.CostPrice,
			"quantity":    product.Quantity,
			"status":      product.Status,
			"category_id": product.CategoryID,
			"image_url": product.ImageURL,
			"image_hash": product.ImageHash,
		})

	fmt.Printf("result %v\n", result)
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("product with this SKU not found")
	}

	if err := r.db.Debug().WithContext(ctx).
		Preload("Category").
		Where("sku = ? AND seller_id = ?", sku, sellerID).
		First(&updatedProduct).Error; err != nil {
		return nil, err
	}

	return &updatedProduct, nil
}

func (r *productRepository) GetCategoryBySlug(ctx context.Context, slug string) (*models.Category, error) {
	var category models.Category

	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}
