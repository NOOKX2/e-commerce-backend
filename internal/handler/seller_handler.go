package handler

import "github.com/gofiber/fiber/v2"

type SellerHandler struct {
}

func NewSellerHandler() *SellerHandler {
	return &SellerHandler{}
}

func (h *SellerHandler) GetProductsPage(c *fiber.Ctx) error{
	userID := c.Locals("userID")
	
	return c.JSON(fiber.Map{
		"message": "Welcome to products page",
		"seller_id": userID,
	})
}
