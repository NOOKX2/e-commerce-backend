package middleware

import "github.com/gofiber/fiber/v2"

func RoleRequired(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("userRole").(string)

		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Role not found in context",
			})
		}

		if userRole != requiredRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You don't have hermission to access this resource",
			})
		}

		return c.Next()
	}
}


