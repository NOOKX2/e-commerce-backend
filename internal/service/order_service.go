package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
	"github.com/NOOKX2/e-commerce-backend/pkg/response"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/paymentintent"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"gorm.io/gorm"
)

type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, userID uint, shippingAddress *models.ShippingAddress, items []models.OrderItem, paymentMethodID string, shouldSaveCard bool) (*models.Order, error)
	GetAllOrders(ctx context.Context) ([]models.Order, error)
	GetUserOrders(ctx context.Context, userID uint) ([]models.Order, error)
	GetOrderByID(ctx context.Context, orderID uint, userID uint) (*models.Order, error)
	GetOrderBySellerID(ctx context.Context, sellerID uint, page, limit int, search string) ([]response.SellerOrderResponse, map[string]interface{}, error)
	GetOrderDetailsBySellerID(ctx context.Context, orderID uint, sellerID uint) (*response.SellerOrderDetailResponse, error)
	GetCustomersBySellerID(ctx context.Context, sellerID uint, page, limit int, search string) ([]response.SellerCustomerResponse, map[string]interface{}, error)
	GetCustomerDetailBySellerID(ctx context.Context, sellerID uint, customerID uint) (*response.CustomerDetailResponse, error)
	GetDashboardSummary(ctx context.Context, sellerID uint) (map[string]interface{}, error)
}

type OrderService struct {
	orderRepo       repository.OrderRepositoryInterface
	productRepo     ProductServiceInterface
	userCardService UserCardServiceInterface
	userService     UserServiceInterface
}

func NewOrderService(orderRepo repository.OrderRepositoryInterface, productRepo ProductServiceInterface, userCardService UserCardServiceInterface, userService UserServiceInterface) OrderServiceInterface {
	return &OrderService{
		orderRepo:       orderRepo,
		productRepo:     productRepo,
		userCardService: userCardService,
		userService:     userService,
	}
}

func (osv *OrderService) CreateOrder(ctx context.Context, userID uint, shippingAddress *models.ShippingAddress, items []models.OrderItem, paymentMethodID string, shouldSaveCard bool) (*models.Order, error) {
	if len(items) == 0 {
		return nil, ErrOrderNotFound
	}

	user, err := osv.userService.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("User id not found: %v", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if user.Status == models.UserStatusSuspended {
		return nil, ErrAccountSuspended
	}

	for _, item := range items {
		product, err := osv.productRepo.GetProductByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("Product id not found: %v", err)
		}

		if product.SellerID == userID {
			return nil, fmt.Errorf("Seller cannot buy your own product")
		}

		if product.Quantity < item.Quantity {
			return nil, fmt.Errorf("%s not enough in stock", product.Name)
		}
	}

	stripeCustomerID := user.StripeCustomerID
	if strings.TrimSpace(stripeCustomerID) == "" {
		params := &stripe.CustomerParams{
			Email: stripe.String(user.Email),
		}
		newCustomer, err := customer.New(params)
		if err != nil {
			return nil, fmt.Errorf("failed to create stripe customer: %v", err)
		}

		stripeCustomerID = newCustomer.ID
		err = osv.userService.UpdateStripeCustomerID(ctx, userID, stripeCustomerID)
		if err != nil {
			fmt.Printf("Warning: Could not save customer ID to DB: %v\n", err)
		}

	} else {
		stripeCustomerID = user.StripeCustomerID
	}

	var userCardID *uint
	var stripePaymentIntentID string

	if shouldSaveCard {
		attachParams := &stripe.PaymentMethodAttachParams{
			Customer: stripe.String(stripeCustomerID),
		}

		_, err = paymentmethod.Attach(paymentMethodID, attachParams)
		if err != nil {
			return nil, fmt.Errorf("Cannot attach paymentMethodID and attachParams: %v", err)
		}

		saveReq := request.SaveCardRequest{
			PaymentMethodID: paymentMethodID,
		}

		cardID, err := osv.userCardService.SaveNewCard(ctx, userID, saveReq)
		if err != nil {
			fmt.Printf("Warning: Failed to save card: %v\n", err)
		} else {
			userCardID = &cardID

		}
		stripePaymentIntentID = paymentMethodID
	} else {
		stripePaymentIntentID = paymentMethodID

		existingCard, err := osv.userCardService.GetCardByStripePaymentMethodID(ctx, userID, stripePaymentIntentID)
		if err != nil {
			return nil, fmt.Errorf("Card not found: %v", err)
		}

		userCardID = &existingCard.ID
		stripePaymentIntentID = existingCard.StripePaymentMethodID
	}

	return osv.orderRepo.WithTransaction(ctx, func(repo repository.OrderRepositoryInterface) (*models.Order, error) {
		var totalAmount float64
		processedItems := make([]models.OrderItem, len(items))

		for i, item := range items {
			product, err := osv.productRepo.GetProductByID(ctx, item.ProductID)

			if err != nil {
				return nil, fmt.Errorf("item with product ID %d not found", item.ProductID)
			}

			if err := osv.productRepo.RemoveFromStock(ctx, product.ID, item.Quantity); err != nil {
				return nil, fmt.Errorf("failed to update stock for product %d: %w", item.ProductID, err)
			}

			totalAmount += product.Price * float64(item.Quantity)
			processedItems[i] = models.OrderItem{
				ProductID:       product.ID,
				Quantity:        item.Quantity,
				PriceAtPurchase: product.Price,
			}
		}

		stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
		paymentIntentParams := &stripe.PaymentIntentParams{
			Amount:        stripe.Int64(int64(totalAmount * 100)),
			Currency:      stripe.String(string(stripe.CurrencyTHB)),
			Customer:      stripe.String(stripeCustomerID),
			PaymentMethod: stripe.String(stripePaymentIntentID),
			Confirm:       stripe.Bool(true),
			AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
				Enabled:        stripe.Bool(true),
				AllowRedirects: stripe.String("never"),
			},
		}

		if !shouldSaveCard {
			paymentIntentParams.OffSession = stripe.Bool(true)
		}

		paymentIntent, err := paymentintent.New(paymentIntentParams)
		if err != nil {
			return nil, fmt.Errorf("stripe error: %w", err)
		}

		newOrder := &models.Order{
			UserID:                userID,
			Status:                "complete",
			TotalAmount:           totalAmount,

			ShippingEmail:         shippingAddress.Email,
            ShippingReceiverName:  shippingAddress.ReceiverName,
            ShippingPhoneNumber:   shippingAddress.PhoneNumber,
            ShippingStreetAddress: shippingAddress.StreetAddress,
            ShippingSubDistrict:   shippingAddress.SubDistrict,
            ShippingDistrict:      shippingAddress.District,
            ShippingProvince:      shippingAddress.Province,
            ShippingPostalCode:    shippingAddress.PostalCode,

			UserCardID:            userCardID,
			StripePaymentIntentID: &paymentIntent.ID,
			Items:                 processedItems,
		}

		createdOrder, err := osv.orderRepo.CreateOrder(ctx, newOrder)
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

func (os *OrderService) GetOrderBySellerID(ctx context.Context, sellerID uint, page, limit int, search string) ([]response.SellerOrderResponse, map[string]interface{}, error) {
	orders, total, err := os.orderRepo.GetOrderBySellerIDPaginated(ctx, sellerID, page, limit, search)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve user orders: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	meta := map[string]interface{}{
		"total_pages":  totalPages,
		"current_page": page,
		"total_items":  total,
	}

	if orders == nil {
		return []response.SellerOrderResponse{}, meta, nil
	}

	var sellerOrderResponse []response.SellerOrderResponse
	for _, order := range orders {
		productName := "Unknown Product"
		if len(order.Items) > 0 {
			productName = order.Items[0].Product.Name
			if len(order.Items) > 1 {
				productName = fmt.Sprintf("%s (+%d)", productName, len(order.Items)-1)
			}
		}
		formattedDate := order.CreatedAt.Format("02/01/2006")

		sellerOrderResponse = append(sellerOrderResponse, response.SellerOrderResponse{
			ID:       order.ID,
			Product:  productName,
			Customer: order.ShippingReceiverName,
			Date:     formattedDate,
			Amount:   order.TotalAmount,
			Status:   order.Status,
		})
	}

	return sellerOrderResponse, meta, nil
}

func (s *OrderService) GetOrderDetailsBySellerID(ctx context.Context, orderID uint, sellerID uint) (*response.SellerOrderDetailResponse, error) {
	order, err := s.orderRepo.GetOrderDetailsBySellerID(ctx, orderID, sellerID)
	if err != nil {
		return nil, err
	}

	orderItems := []response.SellerOrderItemDTO{}
	var sellerSubtotal float64 = 0;

	for _, item := range order.Items {
		itemTotal := item.PriceAtPurchase * float64(item.Quantity)
		sellerSubtotal += itemTotal

		orderItems = append(orderItems, response.SellerOrderItemDTO{
            ProductID: item.ProductID,
            Name:      item.Product.Name,
            SKU:       item.Product.SKU,
            ImageURL:  item.Product.ImageURL,
            Price:     item.PriceAtPurchase,
            Quantity:  item.Quantity,
            Total:     itemTotal,
        })
	}

	fullAddress := fmt.Sprintf("%s, %s, %s, %s %s",
        order.ShippingStreetAddress,
        order.ShippingSubDistrict,
        order.ShippingDistrict,
        order.ShippingProvince,
        order.ShippingPostalCode,
    )

	res := &response.SellerOrderDetailResponse{
        OrderID:  order.ID,
        Status:   order.Status,
        PlacedAt: order.CreatedAt.Format("January 02, 2006 at 03:04 PM"), 
        CustomerInfo: response.CustomerInfoDTO{
            Name:        order.ShippingReceiverName,
            PhoneNumber: order.ShippingPhoneNumber,
            Email:       order.ShippingEmail, 
        },
        ShippingAddress: response.ShippingAddressDTO{
            AddressLine: fullAddress,
        },
        Items:          orderItems,
        SellerSubtotal: sellerSubtotal,
    }

    return res, nil
}

func (s *OrderService) GetCustomersBySellerID(ctx context.Context, sellerID uint, page, limit int, search string) ([]response.SellerCustomerResponse, map[string]interface{}, error) {
	customers, total, err := s.orderRepo.GetCustomersBySellerIDPaginated(ctx, sellerID, page, limit, search)
	if err != nil {
		return nil, nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	meta := map[string]interface{}{
		"total_pages":  totalPages,
		"current_page": page,
		"total_items":  total,
	}

	if customers == nil {
		customers = []response.SellerCustomerResponse{}
	}

	return customers, meta, nil
}

func (s *OrderService) GetCustomerDetailBySellerID(ctx context.Context, sellerID uint, customerID uint) (*response.CustomerDetailResponse, error) {
	orders, err := s.orderRepo.GetCustomerOrdersForSeller(ctx, sellerID, customerID)
	if err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		return nil, errors.New("customer not found or has no orders with this seller")
	}

	var totalSpent float64 = 0
	var orderHistory []response.CustomerOrder

	for _, order := range orders {
		var orderTotal float64 = 0
		var itemsCount int = 0

		for _, item := range order.Items {
			orderTotal += item.PriceAtPurchase * float64(item.Quantity)
			itemsCount += int(item.Quantity)
		}

		totalSpent += orderTotal

		orderHistory = append(orderHistory, response.CustomerOrder{
			OrderID:    order.ID,
			Date:       order.CreatedAt.Format("Jan 02, 2006"), 
			Total:      orderTotal,
			Status:     order.Status,
			ItemsCount: itemsCount,
		})
	}

	latestOrder := orders[0]
	firstOrder := orders[len(orders)-1]

	fullLocation := fmt.Sprintf("%s, %s", latestOrder.ShippingProvince, latestOrder.ShippingPostalCode)

	res := &response.CustomerDetailResponse{
		ID:           customerID,
		Name:         latestOrder.ShippingReceiverName,
		Email:        latestOrder.ShippingEmail, 
		PhoneNumber:  latestOrder.ShippingPhoneNumber,
		Location:     fullLocation,
		JoinedDate:   firstOrder.CreatedAt.Format("January 02, 2006"),
		TotalSpent:   totalSpent,
		TotalOrders:  len(orders),
		Status:       "Active", 
		OrderHistory: orderHistory,
	}

	return res, nil
}

func (s *OrderService) GetDashboardSummary(ctx context.Context, sellerID uint) (map[string]interface{}, error) {
    // เรียกใช้ Repository
    summary, err := s.orderRepo.GetSellerDashboardSummary(ctx, sellerID)
    if err != nil {
        return nil, err
    }

    // คุณสามารถเพิ่ม Logic คำนวณเปอร์เซ็นต์การเติบโตตรงนี้ได้ในอนาคต
    return summary, nil
}