package handler

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/NOOKX2/e-commerce-backend/internal/service"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
	"github.com/NOOKX2/e-commerce-backend/pkg/response"
	"github.com/NOOKX2/e-commerce-backend/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
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

	sellerID, err := utils.GetUserIDFromContext(c)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": "false",
			"error":   "Unauthorized access" + err.Error(),
		})
	}

	productInput := service.CreateProductInput{
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		SellerID:    sellerID,
		ImageUrl:    req.ImageUrl,
		Category:    req.Category,
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
	categoryQuery := c.Query("category")
	priceQuery := c.Query("price")
	sortQuery := c.Query("sort")
	pageQuery := c.Query("page", "1")
	limitQuery := c.Query("limit", "12")

	products, err := h.ProductService.GetAllProduct(categoryQuery, priceQuery, sortQuery, pageQuery, limitQuery)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error loading products" + err.Error(),
		})
	}

	productResponses := response.ToProductResponses(products)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": "true",
		"message": "Get all products successful",
		"data":    productResponses,
	})
}

func (h *ProductHandler) GetProductByID(c *fiber.Ctx) error {
	productID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": "false",
			"error": "Product not found" + err.Error(),
		})
	}
	product, err := h.ProductService.GetProductByID(c.UserContext(), productID)

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
		"status":  "success",
		"message": "Get product by ID successfully",
		"data":    product,
	})
}

func (h *ProductHandler) GetProductBySlug(c *fiber.Ctx) error {
	slug := c.Params("slug")
	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": "false",
			"message": "Slug parameter is required",
		})
	}

	product, err := h.ProductService.GetProductBySlug(c.UserContext(), slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": "false",
				"message": "Product not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": "false",
			"message": err.Error(),
		})
	}

	productResponse := response.ToProductResponse(*product)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": "true",
		"message": "Get product by slug successfully",
		"data":    productResponse,
	})
}

func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	ctx := c.UserContext()
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

	updatedProduct, err := h.ProductService.UpdateProduct(ctx, productID, sellerID, productReq)

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
	ctx := c.UserContext()
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

	if err := h.ProductService.DeleteProduct(ctx, productID, sellerID); err != nil {
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
		"message": "Delete product successfully.",
	})
}
