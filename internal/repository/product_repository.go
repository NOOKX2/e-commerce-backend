package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/pkg/cursor"
	"gorm.io/gorm"
)

type ProductRepositoryInterface interface {
	Create(ctx context.Context, product *models.Product) error
	GetAllProduct(category string, price string, sort string, limit uint, afterCursor, beforeCursor string) ([]models.Product, int64, string, string, error)
	GetProductByID(ctx context.Context, id uint) (*models.Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*models.Product, error)
	GetActiveProductBySlug(ctx context.Context, slug string) (*models.Product, error)
	UpdateProduct(product *models.Product) error
	DeleteProduct(sku string) error
	GetProductBySKU(ctx context.Context, sku string) (*models.Product, error)
	AddToStock(ctx context.Context, id uint, amount uint) error
	RemoveFromStock(ctx context.Context, id uint, amount uint) error
	GetProductsBySellerID(ctx context.Context, sellerID uint, limit int, search string, afterCursor, beforeCursor string) ([]models.Product, int64, string, string, error)
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

func (r *productRepository) applyProductListFilters(q *gorm.DB, category, price string) *gorm.DB {
	if category != "" {
		parts := strings.Split(category, ",")
		ids := make([]uint, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if id, err := strconv.ParseUint(p, 10, 32); err == nil {
				ids = append(ids, uint(id))
			}
		}
		if len(ids) > 0 {
			q = q.Where("category_id IN ?", ids)
		}
	}
	if price != "" {
		prices := strings.Split(price, ",")
		if len(prices) == 2 {
			minPrice, errMin := strconv.Atoi(prices[0])
			maxPrice, errMax := strconv.Atoi(prices[1])
			if errMin == nil && errMax == nil {
				q = q.Where("price BETWEEN ? AND ?", minPrice, maxPrice)
			}
		}
	}
	return q
}

func encodeProductListCursor(p models.Product, sortKey, fp string) (string, error) {
	c := cursor.ProductListCursor{
		ID:    p.ID,
		CA:    p.CreatedAt.UnixMilli(),
		Price: p.Price,
		Sort:  sortKey,
		FP:    fp,
	}
	return cursor.Encode(&c)
}

func reverseProductsInPlace(products []models.Product) {
	for i, j := 0, len(products)-1; i < j; i, j = i+1, j-1 {
		products[i], products[j] = products[j], products[i]
	}
}

// GetAllProduct uses keyset pagination. Pass empty afterCursor and beforeCursor for the first page.
func (r *productRepository) GetAllProduct(category string, price string, sort string, limit uint, afterCursor, beforeCursor string) ([]models.Product, int64, string, string, error) {
	sortKey := cursor.NormalizeProductSort(sort)
	fp := cursor.ProductFilterFP(category, price)
	if limit < 1 {
		limit = 12
	}
	fetch := int(limit) + 1

	base := r.db.Model(&models.Product{}).Where("status = ?", models.StatusActive)
	base = r.applyProductListFilters(base, category, price)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, "", "", err
	}

	var products []models.Product

	mode, anchor := resolveProductListPaging(beforeCursor, afterCursor, sortKey, fp)

	q := r.db.Model(&models.Product{}).Where("status = ?", models.StatusActive)
	q = r.applyProductListFilters(q, category, price)

	axis, desc := parseProductSortKey(sortKey)
	switch mode {
	case productPagingBackward:
		q = productListSeekBefore(q, axis, desc, anchor)
		q = productListOrderBackward(q, axis, desc)
	case productPagingForward:
		q = productListSeekAfter(q, axis, desc, anchor)
		q = productListOrderForward(q, axis, desc)
	default:
		q = productListOrderForward(q, axis, desc)
	}

	q = q.Preload("Category").Limit(fetch)
	if err := q.Find(&products).Error; err != nil {
		return nil, 0, "", "", err
	}

	if mode == productPagingBackward {
		reverseProductsInPlace(products)
	}

	hasExtra := len(products) > int(limit)
	if hasExtra {
		products = products[:limit]
	}

	var nextC, prevC string
	if len(products) == 0 {
		return products, total, "", "", nil
	}

	first := products[0]
	last := products[len(products)-1]

	hasPrev := strings.TrimSpace(afterCursor) != "" || (mode == productPagingBackward && hasExtra)
	if hasPrev {
		if enc, err := encodeProductListCursor(first, sortKey, fp); err == nil {
			prevC = enc
		}
	}
	if hasExtra {
		if enc, err := encodeProductListCursor(last, sortKey, fp); err == nil {
			nextC = enc
		}
	}
	return products, total, nextC, prevC, nil
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

func (r *productRepository) GetActiveProductBySlug(ctx context.Context, slug string) (*models.Product, error) {
	var product models.Product
	err := r.db.Preload("Category").WithContext(ctx).
		Where("slug = ? AND status = ?", slug, models.StatusActive).
		First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) UpdateProduct(product *models.Product) error {
	result := r.db.Save(product)
	return result.Error
}

func (r productRepository) DeleteProduct(sku string) error {
	result := r.db.Where("sku = ?", sku).Delete(&models.Product{})
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

func encodeSellerProductCursor(p models.Product, fp string) (string, error) {
	c := cursor.SellerProductCursor{
		ID: p.ID,
		CA: p.CreatedAt.UnixMilli(),
		FP: fp,
	}
	return cursor.Encode(&c)
}

func (r *productRepository) GetProductsBySellerID(ctx context.Context, sellerID uint, limit int, search string, afterCursor, beforeCursor string) ([]models.Product, int64, string, string, error) {
	fp := cursor.SellerListFP(sellerID, search)
	if limit < 1 {
		limit = 10
	}
	fetch := limit + 1

	baseCount := r.sellerProductsBaseQuery(ctx, sellerID, search)
	var total int64
	if err := baseCount.Count(&total).Error; err != nil {
		return nil, 0, "", "", err
	}

	var products []models.Product

	mode, anchor := resolveSellerListPaging(beforeCursor, afterCursor, fp)

	q := r.sellerProductsBaseQuery(ctx, sellerID, search)
	switch mode {
	case sellerPagingBackward:
		q = sellerSeekAfterNewerThanAnchor(q, anchor)
		q = sellerOrderOldestFirst(q)
	case sellerPagingForward:
		q = sellerSeekBeforeOlderThanAnchor(q, anchor)
		q = sellerOrderNewestFirst(q)
	default:
		q = sellerOrderNewestFirst(q)
	}

	if err := q.Preload("Category").Limit(fetch).Find(&products).Error; err != nil {
		return nil, 0, "", "", err
	}

	if mode == sellerPagingBackward {
		reverseProductsInPlace(products)
	}

	hasExtra := len(products) > limit
	if hasExtra {
		products = products[:limit]
	}

	if len(products) == 0 {
		return products, total, "", "", nil
	}

	first := products[0]
	last := products[len(products)-1]

	var nextC, prevC string
	hasPrev := strings.TrimSpace(afterCursor) != "" || (mode == sellerPagingBackward && hasExtra)
	if hasPrev {
		if enc, err := encodeSellerProductCursor(first, fp); err == nil {
			prevC = enc
		}
	}
	if hasExtra {
		if enc, err := encodeSellerProductCursor(last, fp); err == nil {
			nextC = enc
		}
	}

	return products, total, nextC, prevC, nil
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
