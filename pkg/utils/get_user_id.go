package utils

import (
	"errors"
	"github.com/gofiber/fiber/v2"
)

func GetUserIDFromContext(c *fiber.Ctx) (uint, error) {
	rawUserID := c.Locals("user_id")

	if rawUserID == nil {
		return 0, errors.New("user id not found in context")
	}

	var userID uint

	switch v := rawUserID.(type) {
	case float64:
		userID = uint(v)
	case int:
		userID = uint(v)
	case uint:
		userID = v
	case int64:
		userID = uint(v)
	default:
		return 0, errors.New("user id format is invalid")
	}

	if userID == 0 {
		return 0, errors.New("invalid user id (0)")
	}

	return userID, nil
}