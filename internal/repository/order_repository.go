package repository

import (
	"context"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"gorm.io/gorm"
)

type OrderRepositoryInterface interface {
	CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error)
	GetAllOrders(ctx context.Context) ([]models.Order, error)
	GetUserOrders(ctx context.Context, userID uint) ([]models.Order, error)
	GetOrderById(ctx context.Context, orderID uint, userID uint) (*models.Order, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepositoryInterface {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error) {
	itemsToCreate := order.Items
	order.Items = nil
	
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(itemsToCreate) > 0 {
		for i := range itemsToCreate {
			itemsToCreate[i].OrderID = order.ID 
			itemsToCreate[i].ID = 0              
		}

		if err := tx.Create(&itemsToCreate).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return order, nil
}

func (r *orderRepository) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	var orders []models.Order
	result := r.db.WithContext(ctx).Preload("OrderItems").Find(&orders)
	return orders, result.Error
}

func (r *orderRepository) GetUserOrders(ctx context.Context, userID uint) ([]models.Order, error) {
	var userOrders []models.Order
	result := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product"). 
		Where("user_id = ?", userID).   
		Order("created_at DESC").     
		Find(&userOrders)
	return userOrders, result.Error
}

func (r *orderRepository) GetOrderById(ctx context.Context, orderID uint, userID uint) (*models.Order, error) {
	var order models.Order
	result := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product"). 
		Where("user_id = ?", userID).   
		Order("created_at DESC").     
		Find(&order, orderID)
	return &order, result.Error
}