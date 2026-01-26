package handler

import (
	"fmt"
	"strconv"

	"github.com/NOOKX2/e-commerce-backend/internal/service"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
	"github.com/NOOKX2/e-commerce-backend/pkg/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type UserCardHandler struct {
	UserCardService service.UserCardServiceInterface
}

func NewUserCardHandler(us service.UserCardServiceInterface) *UserCardHandler {
	return &UserCardHandler{UserCardService: us}
}

func (h *UserCardHandler) CreatedUserCard(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized access: " + err.Error(),
		})
	}

	var req request.SaveCardRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Validation failed: " + err.Error(),
		})
	}

	cardID, err := h.UserCardService.SaveNewCard(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Card saved successfully",
		"cardID":  cardID,
	})
}

func (h *UserCardHandler) GetCardByUserID(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized access: " + err.Error(),
		})
	}

	cards, err := h.UserCardService.GetCardByUserID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch cards: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"cards":   cards,
	})
}

func (h *UserCardHandler) DeleteCard(c *fiber.Ctx) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized access: " + err.Error(),
		})
	}

	cardIDParam := c.Params("cardID")
	cardID, err := strconv.Atoi(cardIDParam)
	if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "success": false,
            "error":   "Invalid card ID format",
        })
    }

	err = h.UserCardService.DeleteCard(c.UserContext(), uint(cardID), userID)
	if err != nil {
		fmt.Println("error delete card from handler", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "error":   err.Error(),
        })
    }

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "success": true,
        "message": "Card deleted successfully",
    })
}
