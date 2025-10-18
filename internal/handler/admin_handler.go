package handler

import (
	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

func (h *AdminHandler) GetDashboard(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	return c.JSON(fiber.Map{
		"message":  "Welcome to Admin Dashboard",
		"admin_id": userID,
	})
}
