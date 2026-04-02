package api

import (
	"github.com/NOOKX2/e-commerce-backend/configs"
	"github.com/NOOKX2/e-commerce-backend/internal/api/middleware"
	"github.com/NOOKX2/e-commerce-backend/internal/handler"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, userHandler *handler.UserHandler, sellerHandler *handler.SellerHandler, adminHandler *handler.AdminHandler, productHandler *handler.ProductHandler, orderHandler *handler.OrderHandler, userCardHandler *handler.UserCardHandler, uploadHandler *handler.UploadHandler, categoryHandler *handler.CategoryHandler, settingsHandler *handler.SettingsHandler, userRepo repository.UserRepositoryInterface, config *configs.Config) {
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
	v1.Get("/categories", productHandler.GetCategories)
	v1.Get("/platform/public", settingsHandler.GetPublicPlatformSnapshot)

	authRequired := v1.Group("/", middleware.Authentication(config, userRepo))
	authRequired.Get("/profile", userHandler.GetUserProfile)
	authRequired.Put("/profile", userHandler.UpdateProfile)

	authRequired.Put("/products/:id", productHandler.UpdateProduct)

	authRequired.Get("/upload/sign-url", uploadHandler.GetSignedURL)

	orderRoute := authRequired.Group("/orders")
	orderRoute.Get("/", orderHandler.GetUserOrders)
	orderRoute.Post("/", orderHandler.CreateOrder)
	orderRoute.Get("/:id", orderHandler.GetOrderByID)

	cardRoute := authRequired.Group("/cards")
	cardRoute.Post("/", userCardHandler.CreatedUserCard)
	cardRoute.Get("/", userCardHandler.GetCardByUserID)
	cardRoute.Delete("/:cardID", userCardHandler.DeleteCard)

	admin := v1.Group("/admin", middleware.Authentication(config, userRepo), middleware.RoleRequired("admin"))
	admin.Get("/dashboard", adminHandler.GetDashboard)
	admin.Post("/categories", categoryHandler.Create)
	admin.Get("/categories", categoryHandler.List)
	admin.Put("/categories/:id", categoryHandler.Update)
	admin.Delete("/categories/:id", categoryHandler.Delete)
	admin.Get("/users", adminHandler.GetUsers)
	admin.Get("/users/:id", adminHandler.GetUserByID)
	admin.Patch("/users/:id", adminHandler.UpdateUserDetails)
	admin.Patch("/users/:id/status", adminHandler.UpdateUserStatus)
	admin.Get("/products", adminHandler.GetProducts)
	admin.Get("/products/:slug", adminHandler.GetProductBySlug)
	admin.Patch("/products/:id/status", adminHandler.UpdateProductStatus)
	admin.Get("/orders", adminHandler.GetOrders)
	admin.Get("/orders/:id", adminHandler.GetOrderByID)
	admin.Patch("/orders/:id/status", adminHandler.UpdateOrderStatus)
	admin.Get("/settings/platform", settingsHandler.GetAdminPlatformSettings)
	admin.Put("/settings/platform", settingsHandler.PutAdminPlatformSettings)

	seller := v1.Group("/seller", middleware.Authentication(config, userRepo), middleware.RoleRequired("seller"))
	seller.Get("/", orderHandler.GetDashboardSummary)
	seller.Get("/products/:slug", productHandler.GetSellerProductBySlug)
	seller.Get("/products", productHandler.GetProductsBySellerID)
	seller.Get("/orders", orderHandler.GetOrderBySellerID)
	seller.Get("/orders/:id", orderHandler.GetSellerOrderDetails)
	seller.Get("/customers", orderHandler.GetSellerCustomers)
	seller.Get("/customers/:id", orderHandler.GetCustomerDetail)
	seller.Post("/products", productHandler.AddProduct)
	seller.Put("/products/:sku", productHandler.UpdateProductBySKU)
	seller.Delete("/products/:sku", productHandler.DeleteProduct)
	seller.Get("/settings/shop", settingsHandler.GetSellerShopSettings)
	seller.Put("/settings/shop", settingsHandler.PutSellerShopSettings)

}
