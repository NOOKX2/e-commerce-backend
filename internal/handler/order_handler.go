package handler

import (
	"errors"
	"fmt"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/internal/service"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
	"github.com/NOOKX2/e-commerce-backend/pkg/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CreateOrderRequest struct {
	ShippingAddress string `json:"shippingAddress" validate:"required"`
	Items           []struct {
		ProductID uint `json:"productId" validate:"required,gt=0"`
		Quantity  uint `json:"quantity"  validate:"required,gt=0"`
	} `json:"items" validate:"required,min=1,dive"`

	PaymentIntentID string `json:"paymentIntentId" validate:"required"`
}

type CreatePaymentIntentRequest struct {
	Items []request.CartItemRequest `json:"items" validate:"required,min=1"`
}

type OrderHandler struct {
	OrderService service.OrderServiceInterface
}

func NewOrderHandler(os service.OrderServiceInterface) *OrderHandler {
	return &OrderHandler{OrderService: os}
}

func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {

	userID, err := utils.GetUserIDFromContext(c)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": "false",
			"error": "Unauthorized access" + err.Error(),
		})
	}

	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": "false",
			"error":   "Invalid request body: " + err.Error(),
		})
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": "false",
			"error":   "Validation failed: " + err.Error()})
	}

	orderItems := make([]models.OrderItem, len(req.Items))
	for i, item := range req.Items {
		orderItems[i] = models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	for _, item := range orderItems {
		fmt.Println(item)
	}

	createdOrder, err := h.OrderService.CreateOrder(
		c.Context(),
		userID,
		req.ShippingAddress,
		orderItems,
		req.PaymentIntentID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": "false",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": "true",
		"order":   createdOrder,
	})
}

// func (h * OrderHandler) CreatePaymentInend(c *fiber.Ctx) error {

// }

func (h *OrderHandler) GetUserOrders(c *fiber.Ctx) error {
	

	userID, err := utils.GetUserIDFromContext(c)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": "false",
			"error": "Unauthorized access" + err.Error(),
		})
	}

	orders, err := h.OrderService.GetUserOrders(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": "false",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": "true",
		"orders":  orders,
	})
}

func (h *OrderHandler) GetOrderByID(c *fiber.Ctx) error {
	

	userID, err := utils.GetUserIDFromContext(c)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": "false",
			"error": "Unauthorized access" + err.Error(),
		})
	}

	orderID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": "false",
			"error":   "Invalit order ID format. Order ID must be an integer",
		})
	}

	if orderID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": "false",
			"error":   "order id must be positive integer",
		})
	}

	order, err := h.OrderService.GetOrderByID(c.Context(), uint(orderID), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) ||
			errors.Is(err, service.ErrOrderNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": "false",
				"error":   "Order not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "false",
			"error":  err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": "true",
		"order":   order,
	})
}
