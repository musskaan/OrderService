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
	host     = "localhost"
	port     = 5433
	username = "postgres"
	password = "pgpswd"
	dbName   = "OrderServiceDB"
	sslMode  = "disable"
)

func Connection() *gorm.DB {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", host, port, username, password, dbName, sslMode)

	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}

	log.Println("Connected to the database")

	err = db.AutoMigrate(&model.User{}, &model.Order{})

	if err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}

	return db
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
