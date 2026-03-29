package middleware

import (
	"github.com/NOOKX2/e-commerce-backend/configs"
	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"github.com/NOOKX2/e-commerce-backend/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

func Authentication(config *configs.Config, users repository.UserRepositoryInterface) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies("session_token")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "jwt token not found",
			})
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "unexpected signing method")
			}

			return []byte(config.JWTSecret), nil

		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired JWT" + err.Error(),
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid JWT claims data format",
			})
		}

		userID, err := utils.ParseUserID(claims["user_id"])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		user, err := users.GetUserByID(userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "could not verify user",
			})
		}
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "user not found",
			})
		}

		if user.Status == models.UserStatusBanned {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success":   false,
				"errorType": "account_banned",
				"message":   "This account has been banned.",
			})
		}

		c.Locals("user_id", claims["user_id"])
		c.Locals("user_role", claims["user_role"])

		return c.Next()
	}
}
