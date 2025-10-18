package main

import (
	"log"

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

	dbConnection, err := db.NewDatabaseConnection(config)
	if err != nil {
		log.Fatalf("Fatal error: database connection failed: %v", err)
	}

	if err := dbConnection.AutoMigrate(&domain.User{}); err != nil {
		log.Fatalf("failed to run auto-migrations: %v", err)
	}

	userRepository := repository.NewUserRepository(dbConnection)
	userService := service.NewUserService(userRepository, config)
	userHandler := handler.NewUserHandler(userService)

	sellerHandler := handler.NewSellerHandler()
	adminHandler := handler.NewAdminHandler()

	app := fiber.New()

	api.SetupRoutes(app, userHandler, sellerHandler, adminHandler, config)

	log.Printf("Server is starting on port %s", config.ApiPort)
	err = app.Listen(":" + config.ApiPort)
	if err != nil {
		log.Fatalf("Fatal error: server failed to start: %v", err)
	}

}
