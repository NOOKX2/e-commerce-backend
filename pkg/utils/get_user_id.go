package utils

import (
	"errors"
	"github.com/gofiber/fiber/v2"
)

// ParseUserID converts a JWT claim or context value (e.g. float64 from MapClaims) to uint.
func ParseUserID(v interface{}) (uint, error) {
	if v == nil {
		return 0, errors.New("user id not found in context")
	}

	var userID uint

	switch v := v.(type) {
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

func GetUserIDFromContext(c *fiber.Ctx) (uint, error) {
	return ParseUserID(c.Locals("user_id"))
}