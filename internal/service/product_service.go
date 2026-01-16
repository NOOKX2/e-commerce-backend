package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
	"github.com/NOOKX2/e-commerce-backend/pkg/utils"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type CreateProductInput struct {
	Name        string
	SKU         string
	Price       float64
	Description string
	SellerID    uint
	ImageUrl    string
	Category    string
	Quantity    uint
}

type ProductServiceInterface interface {
	AddProduct(ctx context.Context, input CreateProductInput) (*models.Product, error)
	GetAllProduct(category, price, sort, pageStr, limitStr string) ([]models.Product, int64, error)
	GetProductByID(ctx context.Context, id uint) (*models.Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*models.Product, error)
	UpdateProduct(ctx context.Context, productID uint, sellerID uint, productReq *request.UpdateProductRequest) (*models.Product, error)
	DeleteProduct(ctx context.Context, productID uint, sellerID uint) error
	AddToStock(ctx context.Context, id uint, amount uint ) error
	RemoveFromStock(ctx context.Context, id uint, amount uint) error
}

type ProductService struct {
	repo repository.ProductRepositoryInterface
}

func NewProductService(repo repository.ProductRepositoryInterface) ProductServiceInterface {
	return &ProductService{repo: repo}
}

func (s *ProductService) generateUniqueSlug(ctx context.Context, baseSlug string, sku string) (string, error) {
	existingProduct, err := s.repo.GetProductBySKU(ctx, sku)
	if err == nil {
		return existingProduct.Slug, nil
	}

	finalSlug := baseSlug
	for i := 1; ; i++ {
		_, err := s.repo.GetProductBySlug(ctx, finalSlug)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break

			}
			return "", fmt.Errorf("failed to check for existing slug: %w", err)

		}
		finalSlug = fmt.Sprintf("%s-%d", baseSlug, i)

	}
	return finalSlug, nil
}

func (s *ProductService) AddProduct(ctx context.Context, input CreateProductInput) (*models.Product, error) {
	existingProduct, err := s.repo.GetProductBySKU(ctx, input.SKU)
	if err == nil {
		err = s.repo.AddToStock(ctx, existingProduct.ID, input.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to update stock: %w", err)
		}

		updatedProduct, _ := s.repo.GetProductBySKU(ctx, input.SKU)
		return updatedProduct, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if input.Name == "" {
		return nil, errors.New("Product name cannot be empty")
	}

	if input.Price <= 0 {
		return nil, errors.New("Product price must be a pisitive value")
	}

	baseSlug := utils.Slugify(input.Name)
	uniqueSlug, err := s.generateUniqueSlug(ctx, baseSlug, input.SKU)
	if err != nil {
		return nil, err
	}

	product := &models.Product{
		Name:        input.Name,
		SKU:         input.SKU,
		Price:       input.Price,
		Description: input.Description,
		SellerID:    input.SellerID,
		ImageURL:    input.ImageUrl,
		Category:    input.Category,
		Slug:        uniqueSlug,
		Quantity:    input.Quantity,
	}

	err = s.repo.Create(ctx, product)

	return product, err
}

func (s *ProductService) GetAllProduct(category, price, sort, pageStr, limitStr string) ([]models.Product, int64, error) {
	page, err := strconv.ParseUint(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.ParseUint(limitStr, 10, 64)
	if err != nil || limit < 1 {
		limit = 12
	}

	offset := (page - 1) * limit

	products, total, err := s.repo.GetAllProduct(category, price, sort, uint(limit), uint(offset))
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil

}

func (s *ProductService) GetProductByID(ctx context.Context, id uint) (*models.Product, error) {
	product, err := s.repo.GetProductByID(ctx, id)

	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) GetProductBySlug(ctx context.Context, slug string) (*models.Product, error) {
	product, err := s.repo.GetProductBySlug(ctx, slug)

	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) getProductForUpdate(ctx context.Context, productID uint, sellerID uint) (*models.Product, error) {
	existingProduct, err := s.repo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	if existingProduct == nil {
		return nil, ErrProductNotFound
	}

	if existingProduct.SellerID != sellerID {
		return nil, ErrForbidden
	}

	return existingProduct, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, productID uint, sellerID uint, productReq *request.UpdateProductRequest) (*models.Product, error) {
	existingProduct, err := s.getProductForUpdate(ctx, productID, sellerID)

	if err != nil {
		return nil, err
	}

	if err := copier.Copy(existingProduct, productReq); err != nil {
		return nil, errors.New("Error update data" + err.Error())
	}

	if err := s.repo.UpdateProduct(existingProduct); err != nil {
		return nil, err
	}

	return existingProduct, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, productID uint, sellerID uint) error {
	product, err := s.repo.GetProductByID(ctx, productID)

	if err != nil {
		return ErrProductNotFound
	}

	if product.SellerID != sellerID {
		return ErrForbidden
	}

	if err := s.repo.DeleteProduct(productID); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

func (s * ProductService) AddToStock(ctx context.Context, id uint, amount uint) error {
	return s.repo.AddToStock(ctx, id, amount)
}

func (s * ProductService) RemoveFromStock(ctx context.Context, id uint, amount uint) error {
	return s.repo.RemoveFromStock(ctx, id, amount)
}