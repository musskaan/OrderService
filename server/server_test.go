package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	o "orderService.com/go-orderService-grpc/proto/order"
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

func TestCreateOrder_AuthorizationScenarios(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		expectedCode codes.Code
	}{
		{
			name:         "No Authorization Header",
			ctx:          context.Background(),
			expectedCode: codes.Unauthenticated,
		},
		{
			name:         "Invalid Authorization Header",
			ctx:          metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer token")),
			expectedCode: codes.Unauthenticated,
		},
		{
			name:         "Malformed Authorization Header",
			ctx:          metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Basic invalid_base64")),
			expectedCode: codes.Unauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, _, err := sqlmock.New()
			assert.Nil(t, err, "Failed to create mock DB: %v", err)
			defer mockDB.Close()

			dialect := postgres.New(postgres.Config{
				Conn:       mockDB,
				DriverName: "postgres",
			})

			gormDb, err := gorm.Open(dialect, &gorm.Config{})
			assert.Nil(t, err, "Failed to open GORM DB: %v", err)

			orderRequest := &o.CreateOrderRequest{
				RestaurantId: "test_restaurant",
				MenuItems:    map[string]int32{"item1": 2, "item2": 1},
			}

			orderServiceServer := &OrderServiceServer{DB: gormDb}

			response, err := orderServiceServer.Create(tt.ctx, orderRequest)

			assert.NotNil(t, err)
			assert.Nil(t, response)
			statusErr, ok := status.FromError(err)
			assert.True(t, ok, "Expected gRPC status error")
			assert.Equal(t, tt.expectedCode, statusErr.Code(), "Expected error code")
		})
	}
}

func TestCreateOrder_GetUserByUsernameScenarios(t *testing.T) {
	tests := []struct {
		name            string
		ctx             context.Context
		username        string
		password        string
		extractOK       bool
		dbError         error
		expectedCode    codes.Code
		expectedMessage string
	}{
		{
			name:            "Valid Credentials",
			ctx:             metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Basic dXNlcm5hbWU6cGFzc3dvcmQ=")),
			username:        "username",
			password:        "password",
			extractOK:       true,
			dbError:         nil,
			expectedCode:    codes.NotFound,
			expectedMessage: "user not found: record not found",
		},
		{
			name:            "DB Error",
			ctx:             metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Basic dXNlcm5hbWU6cGFzc3dvcmQ=")),
			username:        "username",
			password:        "password",
			extractOK:       true,
			dbError:         gorm.ErrRecordNotFound,
			expectedCode:    codes.NotFound,
			expectedMessage: "user not found: record not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			assert.Nil(t, err, "Failed to create mock DB: %v", err)
			defer mockDB.Close()

			dialect := postgres.New(postgres.Config{
				Conn:       mockDB,
				DriverName: "postgres",
			})

			gormDb, err := gorm.Open(dialect, &gorm.Config{})
			assert.Nil(t, err, "Failed to open GORM DB: %v", err)

			mock.ExpectQuery("SELECT * FROM \"users\" WHERE username = ?").WithArgs("username").WillReturnError(gorm.ErrRecordNotFound)

			orderServiceServer := &OrderServiceServer{DB: gormDb}

			response, err := orderServiceServer.Create(tt.ctx, &o.CreateOrderRequest{})

			assert.NotNil(t, err)
			assert.Nil(t, response)
			statusErr, ok := status.FromError(err)
			assert.True(t, ok, "Expected gRPC status error")
			assert.Equal(t, tt.expectedCode, statusErr.Code(), "Expected error code")
		})
	}
}
