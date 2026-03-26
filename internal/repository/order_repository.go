package repository

import (
	"context"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/pkg/response"
	"gorm.io/gorm"
)

type OrderRepositoryInterface interface {
	WithTransaction(ctx context.Context, fn func(repo OrderRepositoryInterface) (*models.Order, error)) (*models.Order, error)
	CreateOrder(ctx context.Context, order *models.Order) (*models.Order, error)
	GetAllOrders(ctx context.Context) ([]models.Order, error)
	GetUserOrders(ctx context.Context, userID uint) ([]models.Order, error)
	GetOrderById(ctx context.Context, orderID uint, userID uint) (*models.Order, error)
	GetOrderBySellerIDPaginated(ctx context.Context, sellerID uint, page, limit int, search string) ([]models.Order, int64, error)
	GetOrderDetailsBySellerID(ctx context.Context, orderID uint, sellerID uint) (*models.Order, error)
	GetCustomersBySellerIDPaginated(ctx context.Context, sellerID uint, page, limit int, search string) ([]response.SellerCustomerResponse, int64, error)
	GetCustomerOrdersForSeller(ctx context.Context, sellerID uint, customerID uint) ([]models.Order, error)
	GetSellerDashboardSummary(ctx context.Context, sellerID uint) (map[string]interface{}, error)
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

func (r *orderRepository) GetOrderBySellerIDPaginated(ctx context.Context, sellerID uint, page, limit int, search string) ([]models.Order, int64, error) {
	var orders []models.Order
	offset := (page - 1) * limit

	// 1. สร้าง Base Query (ร่างต้น)
	baseQuery := r.db.WithContext(ctx).Table("orders").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("products.seller_id = ?", sellerID)

	if search != "" {
		searchTerm := "%" + search + "%"
		baseQuery = baseQuery.Where(`(
            CAST(orders.id AS TEXT) ILIKE ? OR 
            orders.shipping_receiver_name ILIKE ? OR 
            CAST(orders.total_amount AS TEXT) ILIKE ? OR
            orders.status ILIKE ? OR
            products.name ILIKE ?
        )`, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// 2. โคลนร่างต้นไปใช้นับจำนวน
	var total int64
	if err := baseQuery.Session(&gorm.Session{}).Distinct("orders.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 3. โคลนร่างต้นไปใช้ดึงข้อมูล
	if err := baseQuery.Session(&gorm.Session{}).Distinct("orders.*").
		Order("orders.created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN products ON products.id = order_items.product_id").
				Where("products.seller_id = ?", sellerID).
				Preload("Product")
		}).
		Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
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


func (r *orderRepository) GetCustomersBySellerIDPaginated(ctx context.Context, sellerID uint, page, limit int, search string) ([]response.SellerCustomerResponse, int64, error) {
    var customers []response.SellerCustomerResponse
    offset := (page - 1) * limit

    // 1. สร้าง Base Query ร่างต้น (รวม Joins และ SellerID Filter)
    baseQuery := r.db.WithContext(ctx).Table("orders").
        Joins("JOIN order_items ON order_items.order_id = orders.id").
        Joins("JOIN products ON products.id = order_items.product_id").
        Where("products.seller_id = ?", sellerID)

    // --- 🔍 เพิ่มเงื่อนไขการค้นหา (Search Logic) ---
    if search != "" {
        searchTerm := "%" + search + "%"
        // ค้นหาจากชื่อผู้รับ หรือ อีเมล (ใช้ ILIKE สำหรับ Postgres เพื่อให้ไม่สนตัวเล็กตัวใหญ่)
        baseQuery = baseQuery.Where("(orders.shipping_receiver_name ILIKE ? OR orders.shipping_email ILIKE ?)", searchTerm, searchTerm)
    }

    // 2. นับจำนวนลูกค้าที่ไม่ซ้ำกัน (Distinct User IDs) สำหรับ Pagination
    // ใช้ Session(&gorm.Session{}) เพื่อแยก Query ออกมาไม่ให้รบกวนกัน
    var total int64
    if err := baseQuery.Session(&gorm.Session{}).
        Distinct("orders.user_id").
        Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // ถ้าไม่มีข้อมูลเลย ให้คืนค่ากลับไปทันที ประหยัดทรัพยากร
    if total == 0 {
        return customers, 0, nil
    }

    // 3. ดึงข้อมูลรายละเอียดลูกค้าพร้อมการทำ Aggregation
    if err := baseQuery.Session(&gorm.Session{}).
        Select(`
            orders.user_id as id,
            MAX(orders.shipping_receiver_name) as name,
            MAX(orders.shipping_email) as email,
            COUNT(DISTINCT orders.id) as total_orders,
            SUM(order_items.price_at_purchase * order_items.quantity) as total_spent,
            MAX(orders.created_at) as last_order_date,
            MAX(orders.shipping_province) as location,
            'Active' as status
        `).
        Group("orders.user_id").
        Order("last_order_date DESC").
        Limit(limit).
        Offset(offset).
        Scan(&customers).Error; err != nil {
        return nil, 0, err
    }

    return customers, total, nil
}

func (r *orderRepository) GetCustomerOrdersForSeller(ctx context.Context, sellerID uint, customerID uint) ([]models.Order, error) {
	var orders []models.Order

	if err := r.db.WithContext(ctx).
		Distinct("orders.*").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("orders.user_id = ? AND products.seller_id = ?", customerID, sellerID).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN products ON products.id = order_items.product_id").
				Where("products.seller_id = ?", sellerID)
		}).
		Order("orders.created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *orderRepository) GetSellerDashboardSummary(ctx context.Context, sellerID uint) (map[string]interface{}, error) {
	var summary struct {
		TotalRevenue float64 `gorm:"column:total_revenue"`
		ActiveOrders int     `gorm:"column:active_orders"`
		NewCustomers int     `gorm:"column:new_customers"`
	}

	// คำนวณยอดเฉพาะสินค้าของ Seller นี้
	err := r.db.WithContext(ctx).Table("orders").
		Select(`
            COALESCE(SUM(order_items.price_at_purchase * order_items.quantity), 0) as total_revenue,
            COUNT(DISTINCT CASE WHEN orders.status NOT IN ('completed', 'cancelled') THEN orders.id END) as active_orders,
            COUNT(DISTINCT orders.user_id) as new_customers
        `).
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Joins("JOIN products ON products.id = order_items.product_id").
		Where("products.seller_id = ?", sellerID).
		Scan(&summary).Error

	if err != nil {
		return nil, err
	}
	type RecentOrder struct {
		ID       string  `json:"id"`
		Product  string  `json:"product"`
		Customer string  `json:"customer"`
		Date     string  `json:"date"`
		Amount   float64 `json:"amount"`
		Status   string  `json:"status"`
	}
	var recentOrders []RecentOrder

	// ดึงเฉพาะออเดอร์ที่มีสินค้าของ Seller นี้ และเอาข้อมูลล่าสุด
	r.db.WithContext(ctx).Table("orders").
		Select(`
            orders.id, 
            p.name as product, 
            u.name as customer, 
            orders.created_at as date, 
            (oi.price_at_purchase * oi.quantity) as amount, 
            orders.status
        `).
		Joins("JOIN order_items oi ON oi.order_id = orders.id").
		Joins("JOIN products p ON p.id = oi.product_id").
		Joins("JOIN users u ON u.id = orders.user_id").
		Where("p.seller_id = ?", sellerID).
		Order("orders.created_at DESC").
		Limit(5).
		Scan(&recentOrders)

	return map[string]interface{}{
		"totalRevenue": summary.TotalRevenue,
		"activeOrders": summary.ActiveOrders,
		"newCustomers": summary.NewCustomers,
		"recentOrders": recentOrders,
	}, nil
}
