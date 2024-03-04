package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	u "orderService.com/go-orderService-grpc/proto/user"
)

func setupTestDB(t *testing.T) (sqlmock.Sqlmock, *UserServiceServer) {
	mockDB, mock, err := sqlmock.New()
	assert.Nil(t, err, "Failed to create mock DB: %v", err)
	defer mockDB.Close()

	dialect := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})

	gormDb, err := gorm.Open(dialect, &gorm.Config{})
	assert.Nil(t, err, "Failed to open GORM DB: %v", err)

	userServer := &UserServiceServer{DB: gormDb}

	mock.ExpectBegin()

	return mock, userServer
}

func TestRegisterUser_InvalidUserData_UsernameEmpty_ReturnsError(t *testing.T) {
	mock, userServer := setupTestDB(t)

	user := &u.RegisterUserRequest{
		Password: "password",
		Address: &u.Address{
			Street:  "street",
			City:    "city",
			State:   "state",
			Zipcode: "zip",
		},
	}

	got, err := userServer.Register(context.Background(), user)

	mock.ExpectRollback()
	assert.NotNil(t, err)
	assert.Nil(t, got)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, codes.InvalidArgument, statusErr.Code(), "Expected InvalidArgument error")
}

func TestRegisterUser_InvalidUserData_PasswordEmpty_ReturnsError(t *testing.T) {
	mock, userServer := setupTestDB(t)

	user := &u.RegisterUserRequest{
		Username: "user",
		Address: &u.Address{
			Street:  "street",
			City:    "city",
			State:   "state",
			Zipcode: "zip",
		},
	}

	got, err := userServer.Register(context.Background(), user)

	mock.ExpectRollback()
	assert.NotNil(t, err)
	assert.Nil(t, got)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, codes.InvalidArgument, statusErr.Code(), "Expected InvalidArgument error")
}

func TestRegisterUser_InvalidUserData_NullAddress_ReturnsError(t *testing.T) {
	mock, userServer := setupTestDB(t)

	user := &u.RegisterUserRequest{
		Username: "user",
		Password: "password",
		Address:  nil,
	}

	got, err := userServer.Register(context.Background(), user)

	mock.ExpectRollback()
	assert.NotNil(t, err)
	assert.Nil(t, got)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, codes.InvalidArgument, statusErr.Code(), "Expected InvalidArgument error")
}

func TestRegisterUser_InvalidUserData_EmptyStreet_ReturnsError(t *testing.T) {
	mock, userServer := setupTestDB(t)

	user := &u.RegisterUserRequest{
		Username: "user",
		Password: "password",
		Address: &u.Address{
			Street:  "",
			City:    "city",
			State:   "state",
			Zipcode: "zip",
		},
	}

	got, err := userServer.Register(context.Background(), user)

	mock.ExpectRollback()
	assert.NotNil(t, err)
	assert.Nil(t, got)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, codes.InvalidArgument, statusErr.Code(), "Expected InvalidArgument error")
}

func TestRegisterUser_InvalidUserData_EmptyCity_ReturnsError(t *testing.T) {
	mock, userServer := setupTestDB(t)

	user := &u.RegisterUserRequest{
		Username: "user",
		Password: "password",
		Address: &u.Address{
			Street:  "street",
			City:    "",
			State:   "state",
			Zipcode: "zip",
		},
	}

	got, err := userServer.Register(context.Background(), user)

	mock.ExpectRollback()
	assert.NotNil(t, err)
	assert.Nil(t, got)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, codes.InvalidArgument, statusErr.Code(), "Expected InvalidArgument error")
}

func TestRegisterUser_InvalidUserData_EmptyState_ReturnsError(t *testing.T) {
	mock, userServer := setupTestDB(t)

	user := &u.RegisterUserRequest{
		Username: "user",
		Password: "password",
		Address: &u.Address{
			Street:  "street",
			City:    "city",
			State:   "",
			Zipcode: "zip",
		},
	}

	got, err := userServer.Register(context.Background(), user)

	mock.ExpectRollback()
	assert.NotNil(t, err)
	assert.Nil(t, got)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, codes.InvalidArgument, statusErr.Code(), "Expected InvalidArgument error")
}

func TestRegisterUser_InvalidUserData_EmptyZipCode_ReturnsError(t *testing.T) {
	mock, userServer := setupTestDB(t)

	user := &u.RegisterUserRequest{
		Username: "user",
		Password: "password",
		Address: &u.Address{
			Street:  "street",
			City:    "city",
			State:   "state",
			Zipcode: "",
		},
	}

	got, err := userServer.Register(context.Background(), user)

	mock.ExpectRollback()
	assert.NotNil(t, err)
	assert.Nil(t, got)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, codes.InvalidArgument, statusErr.Code(), "Expected InvalidArgument error")
}

func TestRegisterUser_UnexpectedDatabaseError_ReturnsError(t *testing.T) {
	mock, userServer := setupTestDB(t)

	user := &u.RegisterUserRequest{
		Username: "user",
		Password: "password",
		Address: &u.Address{
			Street:  "street",
			City:    "city",
			State:   "state",
			Zipcode: "zip",
		},
	}

	mock.ExpectQuery("INSERT").WillReturnError(fmt.Errorf("some database error"))

	mock.ExpectCommit()

	got, err := userServer.Register(context.Background(), user)

	mock.ExpectRollback()
	assert.NotNil(t, err)
	assert.Nil(t, got)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok, "Expected gRPC status error")
	assert.Equal(t, codes.Unknown, statusErr.Code(), "Expected InvalidArgument error")
}

func TestRegisterUser_Success(t *testing.T) {
	mockDB, mock, err := sqlmock.New()

	assert.Nil(t, err, "Failed to create mock DB: %v", err)
	defer mockDB.Close()

	dialect := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})

	gormDb, err := gorm.Open(dialect, &gorm.Config{})
	assert.Nil(t, err, "Failed to open GORM DB: %v", err)

	userServer := &UserServiceServer{DB: gormDb}

	mock.ExpectBegin()

	user := &u.RegisterUserRequest{
		Username: "user",
		Password: "password",
		Address: &u.Address{
			Street:  "street",
			City:    "city",
			State:   "state",
			Zipcode: "zip",
		},
	}

	rows := sqlmock.NewRows([]string{"username", "password", "address_street", "address_city", "address_state", "address_zipcode"}).AddRow("user", "password", "street", "city", "state", "zip")
	mock.ExpectQuery("INSERT").WillReturnRows(rows)

	mock.ExpectCommit()

	got, err := userServer.Register(context.Background(), user)

	mock.ExpectRollback()
	assert.Nil(t, err)
	assert.Equal(t, user.Username, got.Username)
	assert.Equal(t, user.Address, got.Address)
	assert.Equal(t, "Yayy! User Registered Sccessfully!", got.Message)
}
