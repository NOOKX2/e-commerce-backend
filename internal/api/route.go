package api

import (
	"github.com/NOOKX2/e-commerce-backend/configs"
	"github.com/NOOKX2/e-commerce-backend/internal/api/middleware"
	"github.com/NOOKX2/e-commerce-backend/internal/handler"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, userHandler *handler.UserHandler, sellerHandler *handler.SellerHandler, adminHandler *handler.AdminHandler, productHandler *handler.ProductHandler, orderHandler *handler.OrderHandler, userCardHandler *handler.UserCardHandler, uploadHandler *handler.UploadHandler, config *configs.Config) {
	v1 := app.Group("/api/v1")

	v1.Get("/", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "ok"})
	})

	auth := v1.Group("/auth")

	auth.Post("/register", userHandler.Register)
	auth.Post("/login", userHandler.Login)

	v1.Get("/products", productHandler.GetAllProduct)
	v1.Get("/products/id/:id", productHandler.GetProductByID)
	v1.Get("/products/:slug", productHandler.GetProductBySlug)

	authRequired := v1.Group("/", middleware.Authentication(config))
	authRequired.Get("/profile", userHandler.GetUserProfile)

	authRequired.Put("/products/:id", productHandler.UpdateProduct)
	authRequired.Delete("/products/:id", productHandler.DeleteProduct)

	authRequired.Get("/upload/sign-url", uploadHandler.GetSignedURL)

	orderRoute := authRequired.Group("/orders")
	orderRoute.Get("/", orderHandler.GetUserOrders)
	orderRoute.Post("/", orderHandler.CreateOrder)
	orderRoute.Get("/:id", orderHandler.GetOrderByID)

	cardRoute := authRequired.Group("/cards")
	cardRoute.Post("/", userCardHandler.CreatedUserCard)
	cardRoute.Get("/", userCardHandler.GetCardByUserID)
	cardRoute.Delete("/:cardID", userCardHandler.DeleteCard)

	admin := v1.Group("/admin", middleware.Authentication(config), middleware.RoleRequired("admin"))
	admin.Get("/dashboard", adminHandler.GetDashboard)

	seller := v1.Group("/seller", middleware.Authentication(config), middleware.RoleRequired("seller"))
	seller.Get("/products", productHandler.GetProductsBySellerID)
	seller.Get("/categories", productHandler.GetCategories)
	seller.Post("/products", productHandler.AddProduct)    
    seller.Put("/products/:id", productHandler.UpdateProduct) 
    seller.Delete("/products/:id", productHandler.DeleteProduct)
}
