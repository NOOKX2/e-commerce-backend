package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
	"github.com/NOOKX2/e-commerce-backend/pkg/utils"
	"gorm.io/gorm"
)

var (
	ErrCategoryNotFound = errors.New("category not found")
)

type CategoryServiceInterface interface {
	Create(ctx context.Context, req request.CreateCategoryRequest) (*models.Category, error)
	List(ctx context.Context) ([]models.Category, error)
	Update(ctx context.Context, id uint, req request.UpdateCategoryRequest) (*models.Category, error)
	Delete(ctx context.Context, id uint) error
}

type CategoryService struct {
	repo repository.CategoryRepositoryInterface
}

func NewCategoryService(repo repository.CategoryRepositoryInterface) CategoryServiceInterface {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Create(ctx context.Context, req request.CreateCategoryRequest) (*models.Category, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}

	slugVal := strings.TrimSpace(req.Slug)
	if slugVal == "" {
		slugVal = utils.Slugify(name)
	}

	category := &models.Category{
		Name: name,
		Slug: slugVal,
	}

	if err := s.repo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

func (s *CategoryService) List(ctx context.Context) ([]models.Category, error) {
	return s.repo.List(ctx)
}

func (s *CategoryService) Update(ctx context.Context, id uint, req request.UpdateCategoryRequest) (*models.Category, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}

	slugVal := strings.TrimSpace(req.Slug)
	if slugVal == "" {
		slugVal = utils.Slugify(name)
	}

	existing.Name = name
	existing.Slug = slugVal

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *CategoryService) Delete(ctx context.Context, id uint) error {
	// ensure exists for consistent errors
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

