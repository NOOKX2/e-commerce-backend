package api

import (
	"github.com/NOOKX2/e-commerce-backend/configs"
	"github.com/NOOKX2/e-commerce-backend/internal/api/middleware"
	"github.com/NOOKX2/e-commerce-backend/internal/handler"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, userHandler *handler.UserHandler, sellerHandler *handler.SellerHandler, adminHandler *handler.AdminHandler, config *configs.Config) {
	v1 := app.Group("/api/v1")

	v1.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "ok"})
	})

	auth := v1.Group("/auth")

	auth.Post("/register", userHandler.Register)
	auth.Post("/login", userHandler.Login)

	authRequired := v1.Group("/", middleware.Authentication(config))

	authRequired.Get("/profile", userHandler.GetUserProfile)

	admin := v1.Group("/admin", middleware.Authentication(config), middleware.RoleRequired("admin"))
	admin.Get("/dashboard", adminHandler.GetDashboard)

	seller := v1.Group("/seller", middleware.Authentication(config), middleware.RoleRequired("seller"))
	seller.Get("/products", sellerHandler.GetProductsPage)
}
