package handler

import (
	"strconv"
	"strings"

	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/internal/service"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
	"github.com/NOOKX2/e-commerce-backend/pkg/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	AdminService service.AdminServiceInterface
}

func NewAdminHandler(svc service.AdminServiceInterface) *AdminHandler {
	return &AdminHandler{AdminService: svc}
}

func parseAdminPagination(c *fiber.Ctx) (page, limit int) {
	page, _ = strconv.Atoi(c.Query("page", "1"))
	limit, _ = strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	return page, limit
}

func (h *AdminHandler) GetDashboard(c *fiber.Ctx) error {
	summary, err := h.AdminService.GetDashboardSummary(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    summary,
	})
}

func (h *AdminHandler) GetUsers(c *fiber.Ctx) error {
	role := strings.TrimSpace(strings.ToLower(c.Query("role", "")))
	if role == "all" {
		role = ""
	}
	search := strings.TrimSpace(c.Query("search", ""))
	page, limit := parseAdminPagination(c)
	users, meta, err := h.AdminService.ListUsers(c.Context(), role, search, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    users,
		"meta":    meta,
	})
}

func (h *AdminHandler) UpdateUserStatus(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil || userID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid user id"})
	}

	var req request.UpdateUserStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	if err := validator.New().Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "validation failed: " + err.Error()})
	}

	adminID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "unauthorized"})
	}

	st := strings.ToLower(strings.TrimSpace(req.Status))
	if uint(userID) == adminID && (st == string(models.UserStatusSuspended) || st == string(models.UserStatusBanned)) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "cannot suspend or ban your own account",
		})
	}

	if err := h.AdminService.UpdateUserStatus(c.Context(), uint(userID), req.Status); err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "invalid user status" {
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "user status updated"})
}

// GET /v1/admin/users/:id
func (h *AdminHandler) GetUserByID(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil || userID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid user id"})
	}

	user, err := h.AdminService.GetUserByID(c.Context(), uint(userID))

	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "user not found" || err.Error() == "record not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "data": user})
}

// PUT /v1/admin/users/:id
func (h *AdminHandler) UpdateUserDetails(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil || userID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid user id"})
	}

	var req request.UpdateUserDetailsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}

	// ตรวจสอบ Validation (ชื่อห้ามว่าง, Role/Status ต้องตรงตามที่กำหนด)
	if err := validator.New().Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "validation failed: " + err.Error()})
	}

	adminID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "error": "unauthorized"})
	}
	if uint(userID) == adminID && (req.Status == models.UserStatusSuspended || req.Status == models.UserStatusBanned) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "cannot suspend or ban your own account",
		})
	}

	if err := h.AdminService.UpdateUserDetails(c.Context(), uint(userID), req); err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "user not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "user details updated"})
}

func (h *AdminHandler) GetProducts(c *fiber.Ctx) error {
	page, limit := parseAdminPagination(c)
	search := strings.TrimSpace(c.Query("search", ""))
	status := strings.TrimSpace(strings.ToLower(c.Query("status", "")))
	products, meta, err := h.AdminService.ListProducts(c.Context(), page, limit, search, status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "data": products, "meta": meta})
}

func (h *AdminHandler) GetProductBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid product slug",
		})
	}

	product, err := h.AdminService.GetProductBySlug(c.Context(), slug)
	if err != nil {
		status := fiber.StatusInternalServerError
		// จำลองการเช็ค Error เหมือน GetOrderByID ถ้าไม่เจอให้ส่ง 404
		if err.Error() == "product not found" || err.Error() == "record not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    product,
	})
}

func (h *AdminHandler) UpdateProductStatus(c *fiber.Ctx) error {
	productID, err := strconv.Atoi(c.Params("id"))
	if err != nil || productID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid product id"})
	}

	var req request.UpdateProductStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	if err := validator.New().Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "validation failed: " + err.Error()})
	}

	if err := h.AdminService.UpdateProductStatus(c.Context(), uint(productID), req.Status); err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "invalid product status" {
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "product status updated"})
}

func (h *AdminHandler) GetOrders(c *fiber.Ctx) error {
	page, limit := parseAdminPagination(c)
	search := strings.TrimSpace(c.Query("search", ""))
	orders, meta, err := h.AdminService.ListOrders(c.Context(), page, limit, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "data": orders, "meta": meta})
}

func (h *AdminHandler) GetOrderByID(c *fiber.Ctx) error {
	orderID, err := strconv.Atoi(c.Params("id"))
	if err != nil || orderID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid order id"})
	}
	order, err := h.AdminService.GetOrderByID(c.Context(), uint(orderID))
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "order not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "data": order})
}

func (h *AdminHandler) UpdateOrderStatus(c *fiber.Ctx) error {
	orderID, err := strconv.Atoi(c.Params("id"))
	if err != nil || orderID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid order id"})
	}

	var req request.UpdateOrderStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "invalid request body"})
	}
	if err := validator.New().Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "validation failed: " + err.Error()})
	}

	if err := h.AdminService.UpdateOrderStatus(c.Context(), uint(orderID), req.Status); err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "invalid order status" {
			status = fiber.StatusBadRequest
		}
		if err.Error() == "order not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true, "message": "order status updated"})
}
