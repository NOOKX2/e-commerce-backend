package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

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
	CostPrice   float64
	Description string
	SellerID    uint
	ImageUrl    string
	Category    string
	Quantity    uint
	ImageHash   string
}

type ProductServiceInterface interface {
	AddProduct(ctx context.Context, input CreateProductInput) (*models.Product, error)
	GetAllProduct(category, price, sort, pageStr, limitStr string) ([]models.Product, int64, error)
	GetProductByID(ctx context.Context, id uint) (*models.Product, error)
	GetProductBySlug(ctx context.Context, slug string) (*models.Product, error)
	GetProductBySlugForSeller(ctx context.Context, slug string, sellerID uint) (*models.Product, error)
	UpdateProduct(ctx context.Context, productID uint, sellerID uint, productReq *request.UpdateProductRequest) (*models.Product, error)
	DeleteProduct(ctx context.Context, sku string, sellerID uint) error
	AddToStock(ctx context.Context, id uint, amount uint) error
	RemoveFromStock(ctx context.Context, id uint, amount uint) error
	GetProductsBySellerID(ctx context.Context, sellerID uint, page, limit int, search string) ([]models.Product, map[string]interface{}, error)
	GetAllCategories(ctx context.Context) ([]models.Category, error)
	UpdateProductBySKU(ctx context.Context, sellerID uint, sku string, req request.UpdateProductRequest) (*models.Product, error)
}

type ProductService struct {
	repo           repository.ProductRepositoryInterface
	uploadService  UploadService
	settingsRepo   repository.SettingsRepositoryInterface
}

func NewProductService(repo repository.ProductRepositoryInterface, uploadService UploadService, settingsRepo repository.SettingsRepositoryInterface) *ProductService {
	return &ProductService{
		repo:          repo,
		uploadService: uploadService,
		settingsRepo:  settingsRepo,
	}
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

	if input.SKU == "" {
		input.SKU = fmt.Sprintf("%s-%d", utils.Slugify(input.Name), time.Now().Unix()%10000)
	}

	existingProduct, err := s.repo.GetProductBySKU(ctx, input.SKU)
	if err == nil {
		err = s.repo.AddToStock(ctx, existingProduct.ID, input.Quantity)
		if err != nil {
			fmt.Println(err)
			return nil, fmt.Errorf("failed to update stock: %w", err)
		}

		updatedProduct, _ := s.repo.GetProductBySKU(ctx, input.SKU)
		return updatedProduct, nil
	}

	normalizedName := utils.NormalizeName(input.Name)

	existingProduct, err = s.repo.GetProductByNormalizedName(ctx, input.SellerID, normalizedName)
	if err == nil {

		if err := s.AddToStock(ctx, existingProduct.ID, input.Quantity); err != nil {
			return nil, fmt.Errorf("failed to update stock by name: %w", err)
		}

		existingProduct.Price = input.Price
		existingProduct.CostPrice = input.CostPrice
		if input.Description != "" {
			existingProduct.Description = input.Description
		}

		if err := s.repo.UpdateProduct(existingProduct); err != nil {
			return nil, fmt.Errorf("failed to update product info: %w", err)
		}

		return s.repo.GetProductByID(ctx, existingProduct.ID)
	}

	categoryName := input.Category
	if categoryName == "" {
		categoryName = "General"
	}
	categorySlug := utils.Slugify(categoryName)
	categoryID, err := s.repo.GetOrCreateCategory(ctx, categoryName, categorySlug)

	if err != nil {
		return nil, fmt.Errorf("failed to handle category: %w", err)
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

	status := models.StatusActive
	if manual, err := s.settingsRepo.IsManualProductApprovalEnabled(ctx); err == nil && manual {
		status = models.StatusPending
	} else if err != nil {
		return nil, fmt.Errorf("platform settings: %w", err)
	}

	product := &models.Product{
		Name:           input.Name,
		NormalizedName: normalizedName,
		SKU:            input.SKU,
		Price:          input.Price,
		Description:    input.Description,
		SellerID:       input.SellerID,
		ImageURL:       input.ImageUrl,
		CategoryID:     categoryID,
		Slug:           uniqueSlug,
		Quantity:       input.Quantity,
		CostPrice:      input.CostPrice,
		ImageHash:      input.ImageHash,
		Status:         status,
	}
	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	if input.ImageHash != "" && input.ImageUrl != "" {

		errMedia := s.uploadService.SaveMediaRecord(input.ImageHash, input.ImageUrl)
		if errMedia != nil {

			fmt.Printf("Log: Failed to save media record for hash %s: %v\n", input.ImageHash, errMedia)
		} else {
			fmt.Printf("Log: Media record saved successfully for hash: %s\n", input.ImageHash)
		}
	}

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
	product, err := s.repo.GetActiveProductBySlug(ctx, slug)

	if err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) GetProductBySlugForSeller(ctx context.Context, slug string, sellerID uint) (*models.Product, error) {
	product, err := s.repo.GetProductBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if product.SellerID != sellerID {
		return nil, ErrForbidden
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

func (s *ProductService) DeleteProduct(ctx context.Context, sku string, sellerID uint) error {
	product, err := s.repo.GetProductBySKU(ctx, sku)

	if err != nil {
		return ErrProductNotFound
	}

	if product.SellerID != sellerID {
		return ErrForbidden
	}

	if err := s.repo.DeleteProduct(sku); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

func (s *ProductService) AddToStock(ctx context.Context, id uint, amount uint) error {
	return s.repo.AddToStock(ctx, id, amount)
}

func (s *ProductService) RemoveFromStock(ctx context.Context, id uint, amount uint) error {
	return s.repo.RemoveFromStock(ctx, id, amount)
}

func (s *ProductService) GetProductsBySellerID(ctx context.Context, sellerID uint, page, limit int, search string) ([]models.Product, map[string]interface{}, error) {
	products, total, err := s.repo.GetProductsBySellerID(ctx, sellerID, page, limit, search)
	if err != nil {
		return nil, nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	meta := map[string]interface{}{
		"total_pages":  totalPages,
		"current_page": page,
		"total_items":  total,
	}

	if products == nil {
		return []models.Product{}, meta, nil
	}

	return products, meta, nil
}

func (s *ProductService) GetAllCategories(ctx context.Context) ([]models.Category, error) {
	return s.repo.GetAllCategories(ctx)
}

func (s *ProductService) UpdateProductBySKU(ctx context.Context, sellerID uint, sku string, req request.UpdateProductRequest) (*models.Product, error) {
	existingProduct, err := s.repo.GetProductBySKU(ctx, sku)
	if err != nil {
		return nil, err
	}

	if existingProduct.SellerID != sellerID {
		return nil, errors.New("unauthorized: you don't own this product")
	}

	category, err := s.repo.GetCategoryBySlug(ctx, req.Category)
	if err != nil {
		return nil, fmt.Errorf("invalid category: %s", req.Category)
	}

	existingProduct.Name = req.Name
	existingProduct.Price = req.Price
	existingProduct.Description = req.Description
	existingProduct.CategoryID = category.ID
	existingProduct.SalePrice = req.SalePrice
	existingProduct.CostPrice = req.CostPrice
	existingProduct.Quantity = req.Quantity
	existingProduct.Status = models.ProductStatus(req.Status)

	if (req.ImageHash != "" && req.ImageUrl != "") {
		if err := s.uploadService.SaveMediaRecord(req.ImageHash, req.ImageUrl); err != nil {
			return nil, err
		}

		existingProduct.ImageHash = req.ImageHash
		existingProduct.ImageURL = req.ImageUrl
	}
	

	fmt.Println(existingProduct.Price)

	updateProduct, err := s.repo.UpdateProductBySKU(ctx, existingProduct.SKU, existingProduct.SellerID, existingProduct)
	if err != nil {
		return nil, err
	}

	return updateProduct, nil
}
