package handler

import (
	"fmt"
	"strconv"

	"github.com/NOOKX2/e-commerce-backend/internal/service"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
	"github.com/gofiber/fiber/v2"
)

type ProductHandler struct {
	ProductService service.ProductServiceInterface
}

func NewProductHandler(svc service.ProductServiceInterface) *ProductHandler {
	return &ProductHandler{ProductService: svc}
}

func (h *ProductHandler) AddProduct(c *fiber.Ctx) error {
	ctx := c.UserContext()
	req := new(request.ProductRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request " + err.Error(),
		})
	}

	sellerIDFloat, ok := c.Locals("userID").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "User ID not found",
		})
	}

	sellerID := uint(sellerIDFloat)

	productInput := service.CreateProductInput{
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		SellerID:    sellerID,
	}

	product, err := h.ProductService.AddProduct(ctx, productInput)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Create product successful",
		"product": product,
	})
}

func (h *ProductHandler) GetAllProduct(c *fiber.Ctx) error {
	products, err := h.ProductService.GetAllProduct()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error loading products" + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Get all products successful",
		"products": products,
	})
}

func (h *ProductHandler) GetProductByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid data type. ID must be integer",
		})
	}
	id := uint(id64)
	product, err := h.ProductService.GetProductByID(id)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot retrieve product" + err.Error(),
		})
	}

	if product == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Product ID not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Get product by ID successfully",
		"product": product,
	})
}

func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid data type. ID must be integer",
		})
	}
	productID := uint(id64)
	sellerIDFloat, ok := c.Locals("userID").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "authentication context is missing"})
	}
	sellerID := uint(sellerIDFloat)

	productReq := new(request.UpdateProductRequest)
	if err := c.BodyParser(productReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	
	updatedProduct, err := h.ProductService.UpdateProduct(productID, sellerID, productReq)

	if err != nil {
		fmt.Println(err.Error())
		switch err.Error() {
		case service.ErrProductNotFound.Error():
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case service.ErrForbidden.Error():
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		default:
			// Error อื่นๆ ที่ไม่คาดคิด (เช่น DB down)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "an unexpected error occurred"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(updatedProduct)
}

func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid data type. ID must be integer",
		})
	}
	productID := uint(id64)
	sellerIDFloat, ok := c.Locals("userID").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "authentication context is missing"})
	}
	sellerID := uint(sellerIDFloat)

	if err := h.ProductService.DeleteProduct(productID, sellerID); err != nil {
		switch err.Error() {
		case service.ErrProductNotFound.Error():
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case service.ErrForbidden.Error():
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		default:
			// Error อื่นๆ ที่ไม่คาดคิด (เช่น DB down)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "an unexpected error occurred"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message" : "Delete product successfully.",
	})
}