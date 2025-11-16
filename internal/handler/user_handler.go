package handler

import (
	"errors"
	"fmt"

	"github.com/NOOKX2/e-commerce-backend/internal/service"
	"github.com/NOOKX2/e-commerce-backend/pkg/request"
	"github.com/NOOKX2/e-commerce-backend/pkg/response"
	"github.com/NOOKX2/e-commerce-backend/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService service.UserServiceInterface 
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

	token, createdUser, err := h.userService.Register(req.Email, req.Password, req.Name)
	fmt.Printf("Error Type: %T, Message: %v\n", err, err)
	if err != nil {
		if errors.Is(err, service.ErrUserExisted){
			return  c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"status": "false",
				"errorType": "User Exist",
				"message": "User with this email already exist.",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	registerResponse := response.UserResponse{
		ID:    createdUser.ID,
		Email: createdUser.Email,
		Name:  createdUser.Name,
	}
	fmt.Println(registerResponse)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Register successfully",
		"token": token,
		"user": registerResponse,
	})
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	req := new(request.LoginRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body" + err.Error()})
	}

	token, user, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound){
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status": "false",
				"errorType": "User not found",
				"message": "User with this email not found",
			})
		}

		if errors.Is(err, service.ErrPasswordIncorrect){
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status": "false",
				"errorType": "Pssword incorrect",
				"message": "Password incorrect",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
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
