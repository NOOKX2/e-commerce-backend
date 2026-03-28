package handler

import (
	"github.com/NOOKX2/e-commerce-backend/internal/service"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
	"github.com/NOOKX2/e-commerce-backend/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

type SettingsHandler struct {
	SettingsService service.SettingsServiceInterface
}

func NewSettingsHandler(svc service.SettingsServiceInterface) *SettingsHandler {
	return &SettingsHandler{SettingsService: svc}
}

// GET /api/v1/admin/settings/platform
func (h *SettingsHandler) GetAdminPlatformSettings(c *fiber.Ctx) error {
	s, err := h.SettingsService.GetPlatformSettings(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "data": s})
}

// PUT /api/v1/admin/settings/platform
func (h *SettingsHandler) PutAdminPlatformSettings(c *fiber.Ctx) error {
	var req request.UpdatePlatformSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	s, err := h.SettingsService.UpdatePlatformSettings(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "data": s})
}

// GET /api/v1/platform/public — no auth (storefront banner)
func (h *SettingsHandler) GetPublicPlatformSnapshot(c *fiber.Ctx) error {
	s, err := h.SettingsService.GetPlatformSettings(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"maintenanceMode": s.MaintenanceMode,
			"siteName":        s.SiteName,
		},
	})
}

// GET /api/v1/seller/settings/shop
func (h *SettingsHandler) GetSellerShopSettings(c *fiber.Ctx) error {
	sellerID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "unauthorized"})
	}
	shop, err := h.SettingsService.GetSellerShop(c.Context(), sellerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "data": shop})
}

// PUT /api/v1/seller/settings/shop
func (h *SettingsHandler) PutSellerShopSettings(c *fiber.Ctx) error {
	sellerID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "unauthorized"})
	}
	var req request.UpdateSellerShopRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	shop, err := h.SettingsService.UpdateSellerShop(c.Context(), sellerID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "data": shop})
}
