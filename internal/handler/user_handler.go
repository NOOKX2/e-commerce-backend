package handler

import (
	"fmt"

	"github.com/NOOKX2/e-commerce-backend/internal/service"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
	"github.com/NOOKX2/e-commerce-backend/pkg/response"
	"github.com/NOOKX2/e-commerce-backend/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService service.UserServiceInterface // Handler จะคุยกับ Service
}

func NewUserHandler(svc service.UserServiceInterface) *UserHandler {
	return &UserHandler{userService: svc}
}

func (h *UserHandler) Register(c *fiber.Ctx) error {
	req := new(request.RegisterRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request " + err.Error(),
		})
	}

	createdUser, err := h.userService.Register(req.Email, req.Password, req.Name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	res := response.RegisterResponse{
		ID:    createdUser.ID,
		Email: createdUser.Email,
		Name:  createdUser.Name,
	}
	fmt.Println(res)

	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	req := new(request.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body" + err.Error()})
	}

	token, user, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	cookie := utils.GenerateCookie(token)
	c.Cookie(cookie)

	
	 userResponse := response.UserResponse{
	 	ID:    user.ID,
	 	Name:  user.Name, 
	 	Email: user.Email,
	 	Role:  string(user.Role),
	 }
	
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Login successful",
		"token":   token,
		"user": userResponse,
	})
}

func (h *UserHandler) GetUserProfile(c *fiber.Ctx) error {
	userIDFloat, ok := c.Locals("userID").(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "User ID not found in cookie",
		})
	}
	userID := uint(userIDFloat)

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "could not retrieve user",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	response := response.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  string(user.Role),
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Get user by ID successful",
		"response": response,
	})
}
