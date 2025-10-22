package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/NOOKX2/e-commerce-backend/configs"
	"github.com/NOOKX2/e-commerce-backend/internal/api"
	"github.com/NOOKX2/e-commerce-backend/internal/db"
	"github.com/NOOKX2/e-commerce-backend/internal/domain"
	"github.com/NOOKX2/e-commerce-backend/internal/handler"
	"github.com/NOOKX2/e-commerce-backend/internal/repository"
	"github.com/NOOKX2/e-commerce-backend/internal/service"
	"github.com/gofiber/fiber/v2"
)

func main() {
	config, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatalf("Fatal error: could not load config file %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbConnection, err := db.NewDatabaseConnection(config)
	if err != nil {
		log.Fatalf("Fatal error: database connection failed: %v", err)
	}

	if err := dbConnection.AutoMigrate(&domain.User{}, &domain.Product{}); err != nil {
		log.Fatalf("failed to run auto-migrations: %v", err)
	}

	userRepository := repository.NewUserRepository(dbConnection)
	userService := service.NewUserService(userRepository, config)
	userHandler := handler.NewUserHandler(userService)

	sellerHandler := handler.NewSellerHandler()
	adminHandler := handler.NewAdminHandler()

	productRepository := repository.NewProductRepository(dbConnection)
	productService := service.NewProductService(productRepository)
	productHandler := handler.NewProductHandler(productService)

	app := fiber.New()

	api.SetupRoutes(app, userHandler, sellerHandler, adminHandler, productHandler, config)

	log.Printf("Server is starting on port %s", config.ApiPort)
	err = app.Listen(":" + config.ApiPort)
	<-ctx.Done()
	if err != nil {
		log.Fatalf("Fatal error: server failed to start: %v", err)
	}

}
