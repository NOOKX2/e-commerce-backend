package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
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
			ShippingAddress:       *shippingAddress,
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
