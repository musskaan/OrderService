package database

import (
	"errors"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"orderService.com/go-orderService-grpc/model"
)

const (
	host                 = "localhost"
	port                 = 5433
	username             = "postgres"
	password             = "pgpswd"
	orderServiceDbName   = "OrderServiceDB"
	catalogServiceDbName = "RestaurantOrderingSystem"
	sslMode              = "disable"
)

type Databases struct {
	OrderServiceDb   *gorm.DB
	CatalogServiceDb *gorm.DB
}

func Connection() Databases {
	orderServiceConnectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, username, password, orderServiceDbName, sslMode)
	catalogServiceConnectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, username, password, catalogServiceDbName, sslMode)

	orderServiceDb, err := gorm.Open(postgres.Open(orderServiceConnectionString), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}

	catalogServiceDb, err := gorm.Open(postgres.Open(catalogServiceConnectionString), &gorm.Config{})

	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}

	log.Println("Connected to the database")

	err = orderServiceDb.AutoMigrate(&model.User{}, &model.Order{})

	if err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}

	return Databases{
		OrderServiceDb:   orderServiceDb,
		CatalogServiceDb: catalogServiceDb,
	}
}

func GetUserByUsername(db *gorm.DB, username string) (*model.User, error) {
	var user model.User

	err := db.Where("username = ?", username).Find(&user).Error
	if err != nil {
		errStr := fmt.Sprintf("user not found: %v", err)
		return nil, errors.New(errStr)
	}

	return &user, nil
}
