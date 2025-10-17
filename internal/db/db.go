package db

import (
	"fmt"

	"github.com/NOOKX2/e-commerce-backend/configs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)



func NewDatabaseConnection(config *configs.Config) (*gorm.DB, error) {

	dbInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost,
		config.DBPort,
		config.DBUser,
		config.DBPassword,
		config.DBName,
	)
	
	var err error
	db, err := gorm.Open(postgres.Open(dbInfo), &gorm.Config{})

	if err != nil {
		return nil, fmt.Errorf("Could not connect database: %w ", err)
	}

	fmt.Println("Connect to database successfully")

	return db, nil
}
