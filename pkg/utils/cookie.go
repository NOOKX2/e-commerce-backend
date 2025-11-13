package utils

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func GenerateCookie(token string) *fiber.Cookie {
	return &fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		Expires:  time.Now().Add(72 * time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	}
}
