package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
	"gorm.io/gorm"
)

type AdminRepositoryInterface interface {
	GetDashboardSummary(ctx context.Context) (map[string]interface{}, error)
	ListUsers(ctx context.Context, role, search string, page, limit int) ([]models.User, int64, error)
	UpdateUserStatus(ctx context.Context, userID uint, status string) error
	ListProducts(ctx context.Context, page, limit int, search, status string) ([]models.Product, int64, error)
	UpdateProductStatus(ctx context.Context, productID uint, status string) error
	ListOrders(ctx context.Context, page, limit int, search string) ([]models.Order, int64, error)
	GetOrderByID(ctx context.Context, orderID uint) (*models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID uint, status string) error
	GetProductBySlug(ctx context.Context, slug string) (*models.Product, error)
	GetUserByID(ctx context.Context, id uint) (*models.User, error)
	UpdateUserDetails(ctx context.Context, id uint, req request.UpdateUserDetailsRequest) error
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepositoryInterface {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetDashboardSummary(ctx context.Context) (map[string]interface{}, error) {
	var summary struct {
		PlatformGMV   float64 `gorm:"column:platform_gmv"`
		TotalOrders   int     `gorm:"column:total_orders"`
		TotalUsers    int     `gorm:"column:total_users"`
		TotalSellers  int     `gorm:"column:total_sellers"`
		TotalProducts int     `gorm:"column:total_products"`
	}

	// Keep it simple + fast: independent aggregates via scalar subqueries.
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE((SELECT SUM(total_amount) FROM orders), 0) AS platform_gmv,
			(SELECT COUNT(*) FROM orders) AS total_orders,
			(SELECT COUNT(*) FROM users) AS total_users,
			(SELECT COUNT(*) FROM users WHERE role = 'seller') AS total_sellers,
			(SELECT COUNT(*) FROM products) AS total_products
	`).Scan(&summary).Error
	if err != nil {
		return nil, err
	}

	type RecentOrder struct {
		ID     uint    `json:"id"`
		User   string  `json:"user"`
		Amount float64 `json:"amount"`
		Status string  `json:"status"`
		Date   string  `json:"date"`
	}

	var recentOrders []RecentOrder
	r.db.WithContext(ctx).Table("orders").
		Select(`
			orders.id,
			u.name as user,
			orders.total_amount as amount,
			orders.status,
			orders.created_at as date
		`).
		Joins("JOIN users u ON u.id = orders.user_id").
		Order("orders.created_at DESC").
		Limit(5).
		Scan(&recentOrders)

	var recentProducts []models.Product
	r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(5).
		Find(&recentProducts)

	return map[string]interface{}{
		"platformGMV":   summary.PlatformGMV,
		"totalOrders":   summary.TotalOrders,
		"totalUsers":    summary.TotalUsers,
		"totalSellers":  summary.TotalSellers,
		"totalProducts": summary.TotalProducts,
		"recentOrders":  recentOrders,
		"recentProducts": recentProducts,
	}, nil
}

func (r *adminRepository) userListQuery(ctx context.Context, role, search string) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&models.User{})
	if role != "" {
		q = q.Where("role = ?", role)
	}
	if search != "" {
		term := "%" + search + "%"
		q = q.Where("(name ILIKE ? OR email ILIKE ?)", term, term)
	}
	return q
}

func (r *adminRepository) ListUsers(ctx context.Context, role, search string, page, limit int) ([]models.User, int64, error) {
	var total int64
	if err := r.userListQuery(ctx, role, search).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	var users []models.User
	if err := r.userListQuery(ctx, role, search).Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *adminRepository) UpdateUserStatus(ctx context.Context, userID uint, status string) error {
	result := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *adminRepository) productListQuery(ctx context.Context, search, status string) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&models.Product{})
	search = strings.TrimSpace(search)
	status = strings.TrimSpace(strings.ToLower(status))
	if search != "" {
		term := "%" + search + "%"
		idLike := "%" + strings.TrimSpace(search) + "%"
		q = q.Where("(name ILIKE ? OR sku ILIKE ? OR CAST(id AS TEXT) LIKE ?)", term, term, idLike)
	}
	if status != "" && status != "all" {
		q = q.Where("LOWER(status) = ?", status)
	}
	return q
}

func (r *adminRepository) ListProducts(ctx context.Context, page, limit int, search, status string) ([]models.Product, int64, error) {
	var total int64
	if err := r.productListQuery(ctx, search, status).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	var products []models.Product
	if err := r.productListQuery(ctx, search, status).
		Preload("Seller").Preload("Category").Order("created_at DESC").Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *adminRepository) UpdateProductStatus(ctx context.Context, productID uint, status string) error {
	result := r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", productID).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *adminRepository) orderListQuery(ctx context.Context, search string) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&models.Order{})
	search = strings.TrimSpace(search)
	if search != "" {
		term := "%" + search + "%"
		idLike := "%" + search + "%"
		q = q.Where(
			"(shipping_receiver_name ILIKE ? OR status ILIKE ? OR CAST(orders.id AS TEXT) LIKE ?)",
			term, term, idLike,
		)
	}
	return q
}

func (r *adminRepository) ListOrders(ctx context.Context, page, limit int, search string) ([]models.Order, int64, error) {
	var total int64
	if err := r.orderListQuery(ctx, search).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	var orders []models.Order
	if err := r.orderListQuery(ctx, search).
		Preload("Items").Preload("Items.Product").Preload("Items.Product.Seller").
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (r *adminRepository) GetOrderByID(ctx context.Context, orderID uint) (*models.Order, error) {
	var order models.Order
	if err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("Items.Product").
		Preload("Items.Product.Seller").
		First(&order, orderID).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *adminRepository) UpdateOrderStatus(ctx context.Context, orderID uint, status string) error {
	result := r.db.WithContext(ctx).Model(&models.Order{}).Where("id = ?", orderID).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *adminRepository) GetProductBySlug(ctx context.Context, slug string) (*models.Product, error) {
	var product models.Product
    // แนะนำให้ Preload ข้อมูล Category และ Seller มาด้วย เพื่อให้ฝั่ง Frontend นำไปโชว์ได้
    if err := r.db.WithContext(ctx).Preload("Category").Preload("Seller").Where("slug = ?", slug).First(&product).Error; err != nil {
        return nil, err
    }
    return &product, nil
}

func (r* adminRepository) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err 
	}
	return &user, nil
}

func (r *adminRepository) UpdateUserDetails(ctx context.Context, id uint, req request.UpdateUserDetailsRequest) error {
	var user models.User
	// 1. เช็คว่ามี User ไหม
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return errors.New("user not found")
	}

	// 2. อัปเดตเฉพาะฟิลด์ที่อนุญาต
	user.Name = req.Name
	user.Role = req.Role
	user.Status = req.Status

	// 3. เซฟลงฐานข้อมูล
	if err := r.db.WithContext(ctx).Save(&user).Error; err != nil {
		return err
	}
	return nil
}