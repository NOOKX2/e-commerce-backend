package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"gorm.io/gorm"
)

type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, userID uint, shippingAddress *models.ShippingAddress, items []models.OrderItem, paymentIntentID string) (*models.Order, error)
	GetAllOrders(ctx context.Context) ([]models.Order, error)
	GetUserOrders(ctx context.Context, userID uint) ([]models.Order, error)
	GetOrderByID(ctx context.Context, orderID uint, userID uint) (*models.Order, error)
}

type OrderService struct {
	orderRepo   repository.OrderRepositoryInterface
	productRepo ProductServiceInterface
}

func NewOrderService(orderRepo repository.OrderRepositoryInterface, productRepo ProductServiceInterface) OrderServiceInterface {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (os *OrderService) CreateOrder(ctx context.Context, userID uint, shippingAddress *models.ShippingAddress, items []models.OrderItem, paymentIntentID string) (*models.Order, error) {
	if len(items) == 0 {
		return nil, ErrOrderNotFound
	}

	return os.orderRepo.WithTransaction(ctx, func(repo repository.OrderRepositoryInterface) (*models.Order, error) {
		var totalAmount float64
		processedItems := make([]models.OrderItem, len(items))

		for i, item := range items {
			product, err := os.productRepo.GetProductByID(ctx, item.ProductID)

			if err != nil {
				return nil, fmt.Errorf("item with product ID %d not found", item.ProductID)
			}

			err = os.productRepo.RemoveFromStock(ctx, product.ID, item.Quantity)

			if err != nil {
				return nil, fmt.Errorf("failed to update stock for product %d: %w", item.ProductID, err)
			}

			totalAmount += product.Price * float64(item.Quantity)

			processedItems[i] = models.OrderItem{
				ProductID:       product.ID,
				Quantity:        item.Quantity,
				PriceAtPurchase: product.Price,
			}
		}

		newOrder := &models.Order{
			UserID:                userID,
			Status:                "pending",
			TotalAmount:           totalAmount,
			ShippingAddress:       *shippingAddress,
			StripePaymentIntentID: &paymentIntentID,
			Items:                 processedItems,
		}

		createdOrder, err := os.orderRepo.CreateOrder(ctx, newOrder)
		if err != nil {
			return nil, fmt.Errorf("failed to save order %w", err)
		}

		return createdOrder, nil
	})

}

func (os *OrderService) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	orders, err := os.orderRepo.GetAllOrders(ctx)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (os *OrderService) GetUserOrders(ctx context.Context, userID uint) ([]models.Order, error) {
	orders, err := os.orderRepo.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user orders: %w", err)
	}

	return orders, nil
}

func (os *OrderService) GetOrderByID(ctx context.Context, orderID uint, userID uint) (*models.Order, error) {
	order, err := os.orderRepo.GetOrderById(ctx, orderID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}

		return nil, fmt.Errorf("failed to retrieve order: %w", err)
	}

	return order, nil
}
