package service

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
	"gorm.io/gorm"
)

type AdminServiceInterface interface {
	GetDashboardSummary(ctx context.Context) (map[string]interface{}, error)
	ListUsers(ctx context.Context, role, search string, page, limit int) ([]models.User, map[string]interface{}, error)
	UpdateUserStatus(ctx context.Context, userID uint, status string) error
	ListProducts(ctx context.Context, page, limit int, search, status string) ([]models.Product, map[string]interface{}, error)
	UpdateProductStatus(ctx context.Context, productID uint, status string) error
	ListOrders(ctx context.Context, page, limit int, search string) ([]models.Order, map[string]interface{}, error)
	GetOrderByID(ctx context.Context, orderID uint) (*models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID uint, status string) error
	GetProductBySlug(ctx context.Context, slug string) (*models.Product, error)
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	UpdateUserDetails(ctx context.Context, id uint, req request.UpdateUserDetailsRequest) error
}

type AdminService struct {
	repo repository.AdminRepositoryInterface
}

func NewAdminService(repo repository.AdminRepositoryInterface) AdminServiceInterface {
	return &AdminService{repo: repo}
}

func adminPaginationMeta(total int64, page, limit int) map[string]interface{} {
	if limit < 1 {
		limit = 10
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages < 1 {
		totalPages = 1
	}
	return map[string]interface{}{
		"total":         total,
		"total_pages":   totalPages,
		"current_page":  page,
		"limit":         limit,
	}
}

func (s *AdminService) GetDashboardSummary(ctx context.Context) (map[string]interface{}, error) {
	return s.repo.GetDashboardSummary(ctx)
}

func (s *AdminService) ListUsers(ctx context.Context, role, search string, page, limit int) ([]models.User, map[string]interface{}, error) {
	role = strings.TrimSpace(strings.ToLower(role))
	search = strings.TrimSpace(search)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	users, total, err := s.repo.ListUsers(ctx, role, search, page, limit)
	if err != nil {
		return nil, nil, err
	}
	return users, adminPaginationMeta(total, page, limit), nil
}

func (s *AdminService) UpdateUserStatus(ctx context.Context, userID uint, status string) error {
	status = strings.TrimSpace(strings.ToLower(status))
	switch status {
	case string(models.UserStatusActive), string(models.UserStatusSuspended), string(models.UserStatusBanned):
	default:
		return errors.New("invalid user status")
	}
	return s.repo.UpdateUserStatus(ctx, userID, status)
}

func (s *AdminService) ListProducts(ctx context.Context, page, limit int, search, status string) ([]models.Product, map[string]interface{}, error) {
	search = strings.TrimSpace(search)
	status = strings.TrimSpace(strings.ToLower(status))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	products, total, err := s.repo.ListProducts(ctx, page, limit, search, status)
	if err != nil {
		return nil, nil, err
	}
	return products, adminPaginationMeta(total, page, limit), nil
}

func (s *AdminService) UpdateProductStatus(ctx context.Context, productID uint, status string) error {
	status = strings.TrimSpace(strings.ToLower(status))
	switch status {
	case string(models.StatusActive), string(models.StatusInactive), string(models.StatusDraft), string(models.StatusArchived), string(models.StatusPending):
	default:
		return errors.New("invalid product status")
	}
	return s.repo.UpdateProductStatus(ctx, productID, status)
}

func (s *AdminService) ListOrders(ctx context.Context, page, limit int, search string) ([]models.Order, map[string]interface{}, error) {
	search = strings.TrimSpace(search)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	orders, total, err := s.repo.ListOrders(ctx, page, limit, search)
	if err != nil {
		return nil, nil, err
	}
	return orders, adminPaginationMeta(total, page, limit), nil
}

func (s *AdminService) GetOrderByID(ctx context.Context, orderID uint) (*models.Order, error) {
	order, err := s.repo.GetOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	return order, nil
}

func (s *AdminService) UpdateOrderStatus(ctx context.Context, orderID uint, status string) error {
	status = strings.TrimSpace(strings.ToLower(status))
	switch status {
	case "pending", "processing", "complete", "cancelled":
	default:
		return errors.New("invalid order status")
	}
	if err := s.repo.UpdateOrderStatus(ctx, orderID, status); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("order not found")
		}
		return err
	}
	return nil
}

func (s *AdminService) GetProductBySlug(ctx context.Context, slug string) (*models.Product, error) {
	return s.repo.GetProductBySlug(ctx, slug)
}

func (s *AdminService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *AdminService) UpdateUserDetails(ctx context.Context, id uint, req request.UpdateUserDetailsRequest) error {
	return s.repo.UpdateUserDetails(ctx, id, req)
}