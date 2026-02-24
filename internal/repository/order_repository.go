package repository

import (
	"context"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"gorm.io/gorm"
)

type OrderRepositoryInterface interface {
	WithTransaction(ctx context.Context, fn func(repo OrderRepositoryInterface) (*models.Order, error)) (*models.Order, error)
	CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error)
	GetAllOrders(ctx context.Context) ([]models.Order, error)
	GetUserOrders(ctx context.Context, userID uint) ([]models.Order, error)
	GetOrderById(ctx context.Context, orderID uint, userID uint) (*models.Order, error)
	GetOrderBySellerID(ctx context.Context, sellerID uint) ([]models.Order, error)
	GetOrderDetailsBySellerID(ctx context.Context, orderID uint, sellerID uint) (*models.Order, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepositoryInterface {
	return &orderRepository{db: db}
}

func (r *orderRepository) WithTransaction(ctx context.Context, fn func(repo OrderRepositoryInterface) (*models.Order, error)) (*models.Order, error) {
	var resultOrder *models.Order

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &orderRepository{db: tx}

		var err error
		resultOrder, err = fn(txRepo)

		return err
	})

	return resultOrder, err
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

func (r *orderRepository) GetOrderBySellerID(ctx context.Context, sellerID uint) ([]models.Order, error) {
	var orders []models.Order

	if err := r.db.WithContext(ctx).
		Distinct("orders.*").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("products.seller_id = ?", sellerID).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN products ON products.id = order_items.product_id").
				Where("products.seller_id = ?", sellerID).
				Preload("Product") 
		}).
		Order("orders.created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepository) GetOrderDetailsBySellerID(ctx context.Context, orderID uint, sellerID uint) (*models.Order, error) {
	var order models.Order

	if err := r.db.WithContext(ctx).
        Distinct("orders.*").
        Joins("JOIN order_items ON order_items.order_id = orders.id").
        Joins("JOIN products ON products.id = order_items.product_id").
        Where("orders.id = ? AND products.seller_id = ?", orderID, sellerID).

		Preload("Items", func(db *gorm.DB) *gorm.DB {
            return db.Joins("JOIN products ON products.id = order_items.product_id").
                Where("products.seller_id = ?", sellerID).
                Preload("Product") 
        }).

		First(&order).Error; err != nil {
			return nil, err
		}

		return &order, nil
}