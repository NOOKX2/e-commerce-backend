package handler

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

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

	var req request.ProductRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid JSON: " + err.Error(),
		})
	}

	sellerID, err := utils.GetUserIDFromContext(c)
	fmt.Println("seller id product", sellerID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": "false",
			"error":   "Unauthorized access" + err.Error(),
		})
	}

	sku := req.SKU
	if sku == "" {
		sku = fmt.Sprintf("%s-%d", utils.Slugify(req.Name), time.Now().Unix()%10000)
	}

	productInput := service.CreateProductInput{
		SKU:         sku,
		Name:        req.Name,
		Price:       req.Price,
		CostPrice:   req.CostPrice,
		Description: req.Description,
		SellerID:    sellerID,
		ImageUrl:    req.ImageUrl,
		Category:    req.Category,
		Quantity:    req.Quantity,
		ImageHash:   req.ImageHash,
	}

	product, err := h.ProductService.AddProduct(ctx, productInput)
	if err != nil {
		fmt.Println("error here",err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	fmt.Println(product)
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

	limitInt, _ := strconv.Atoi(limitQuery)

	products, total, err := h.ProductService.GetAllProduct(categoryQuery, priceQuery, sortQuery, pageQuery, limitQuery)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error loading products" + err.Error(),
		})
	}

	productResponses := response.ToProductResponses(products)
	totalPages := int(math.Ceil(float64(total) / float64(limitInt)))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": "true",
		"message": "Get all products successful",
		"data":    productResponses,
		"meta": fiber.Map{
			"total_items": total,
			"total_pages": totalPages,
			"limit":       limitInt,
		},
	})
}

func (h *ProductHandler) GetProductByID(c *fiber.Ctx) error {
	productID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": "false",
			"error":   "Product not found" + err.Error(),
		})
	}
	product, err := h.ProductService.GetProductByID(c.UserContext(), productID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Cannot retrieve product" + err.Error(),
		})
	}

	if product == nil {
		fmt.Println("error is here")
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
			fmt.Println("error is here")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": "false",
				"message": "Product not found " + err.Error(),
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
	productID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": "false",
			"error":   "Product not found" + err.Error(),
		})
	}

	sellerIDFloat, ok := c.Locals("userID").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "product id not found"})
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
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "an unexpected error occurred"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Delete product successfully.",
	})
}

func (h *ProductHandler) GetProductsBySellerID(c *fiber.Ctx) error {
	ctx := c.UserContext()
	sellerID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": "false",
			"error":   "SellerID not found" + err.Error(),
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	search := c.Query("search", "")

	products, meta, err := h.ProductService.GetProductsBySellerID(ctx, sellerID, page, limit, search)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": false,
			"error":  err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":   true,
		"products": products,
		"meta":     meta,
	})

}

func (h *ProductHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.ProductService.GetAllCategories(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
            "error": "Failed to fetch categories: " + err.Error(),
        })
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "success": true,
        "data":    categories,
    })
}