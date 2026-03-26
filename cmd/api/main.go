package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/NOOKX2/e-commerce-backend/configs"
	"github.com/NOOKX2/e-commerce-backend/internal/api"
	"github.com/NOOKX2/e-commerce-backend/internal/db"
	"github.com/NOOKX2/e-commerce-backend/internal/handler"
	"github.com/NOOKX2/e-commerce-backend/internal/models"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"github.com/NOOKX2/e-commerce-backend/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	config, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatalf("Fatal error: could not load config file %v", err)
	}

	r2Config := configs.LoadR2Config()

	

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbConnection, err := db.NewDatabaseConnection(config)
	if err != nil {
		log.Fatalf("Fatal error: database connection failed: %v", err)
	}

	if err := dbConnection.AutoMigrate(&models.User{}, &models.Product{}, &models.Order{}, &models.OrderItem{}, &models.ShippingAddress{}, &models.UserCard{}, &models.Media{}); err != nil {
		log.Fatalf("failed to run auto-migrations: %v", err)
	}

	uploadRepository := repository.NewUploadRepository(dbConnection)
	uploadService := service.NewUploadService(r2Config, uploadRepository)
	uploadHandler := handler.NewUploadHandler(uploadService)

	userRepository := repository.NewUserRepository(dbConnection)
	userService := service.NewUserService(userRepository, config)
	userHandler := handler.NewUserHandler(userService)

	sellerHandler := handler.NewSellerHandler()
	adminHandler := handler.NewAdminHandler()

	productRepository := repository.NewProductRepository(dbConnection)
	productService := service.NewProductService(productRepository, uploadService)
	productHandler := handler.NewProductHandler(productService)

	userCardRepository := repository.NewUserCardRepository(dbConnection)
	userCardService := service.NewUserCardService(userCardRepository)
	userCardHandler := handler.NewUserCardHandler(userCardService)

	orderRepository := repository.NewOrderRepository(dbConnection)
	orderService := service.NewOrderService(orderRepository, productService, userCardService, userService)
	orderHandler := handler.NewOrderHandler(orderService)

	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000, http://localhost:3001, http://localhost:3004",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, PATCH, HEAD",
		AllowCredentials: true,
	}))

	api.SetupRoutes(app, userHandler, sellerHandler, adminHandler, productHandler, orderHandler, userCardHandler, uploadHandler, config)

	log.Printf("Server is starting on port %s", config.ApiPort)
	err = app.Listen(":" + config.ApiPort)
	<-ctx.Done()
	if err != nil {
		log.Fatalf("Fatal error: server failed to start: %v", err)
	}

}
