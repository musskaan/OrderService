package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	database "orderService.com/go-orderService-grpc/db"
	"orderService.com/go-orderService-grpc/model"
	o "orderService.com/go-orderService-grpc/proto/order"
	u "orderService.com/go-orderService-grpc/proto/user"
)

type UserServiceServer struct {
	DB *gorm.DB
	u.UserServiceServer
}

type OrderServiceServer struct {
	DB                    *gorm.DB
	CatalogServiceAPI     string
	FulfillmentServiceAPI string
	o.OrderServiceServer
}

func main() {
	go createOrderServer()
	createUserServer()
}

func createOrderServer() {
	lis2, err := net.Listen("tcp", ":8002")
	if err != nil {
		log.Fatalf("Failed to listen: 8002, %v", err)
	}

	catalogServiceAPIUrl := "http://localhost:8080/api/v1/restaurants/"
	fulfillmentServiceAPIUrl := "http://localhost:9090/api/v1/deliveries"

	oServer := grpc.NewServer()
	db := database.Connection()

	o.RegisterOrderServiceServer(oServer, &OrderServiceServer{DB: db, CatalogServiceAPI: catalogServiceAPIUrl, FulfillmentServiceAPI: fulfillmentServiceAPIUrl})
	err = oServer.Serve(lis2)
	if err != nil {
		log.Fatalf("Failed to serve 8002: %v", err)
	}
}

func createUserServer() {
	lis1, err := net.Listen("tcp", ":8001")
	if err != nil {
		log.Fatalf("Failed to listen: 8001, %v", err)
	}

	uServer := grpc.NewServer()
	db := database.Connection()

	u.RegisterUserServiceServer(uServer, &UserServiceServer{DB: db})
	err = uServer.Serve(lis1)
	if err != nil {
		log.Fatalf("Failed to serve 8001: %v", err)
	}
}

func (userServer *UserServiceServer) Register(_ context.Context, req *u.RegisterUserRequest) (*u.RegisterUserResponse, error) {
	if req.Username == "" || req.Password == "" || req.Address == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Missing required user fields")
	}

	if req.Address.City == "" || req.Address.Street == "" || req.Address.State == "" || req.Address.Zipcode == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid address data")
	}

	address := &model.Address{
		Street:  req.Address.Street,
		City:    req.Address.City,
		State:   req.Address.State,
		Zipcode: req.Address.Zipcode,
	}

	hashedPassword, err := HashPassword(req.Password)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error hashing password")
	}

	user := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Address:  address,
	}

	err = userServer.DB.Create(&user).Error

	if err != nil {
		errorString := fmt.Sprintf("error storing the user: %v", err)
		return nil, status.Errorf(codes.Unknown, errorString)
	}

	response := &u.RegisterUserResponse{
		Username: user.Username,
		Address:  req.Address,
		Message:  "Yayy! User Registered Sccessfully!",
	}

	return response, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func (orderServer *OrderServiceServer) Create(ctx context.Context, req *o.CreateOrderRequest) (*o.CreateOrderResponse, error) {
	username, password, ok := extractCredentials(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "Credentials not found")
	}

	user, err := database.GetUserByUsername(orderServer.DB, username)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	res, err := isAuthenticated(user.Password, password)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error while decrypting password")
	}

	if !res {
		return nil, status.Errorf(codes.InvalidArgument, "Unauthorized: Invalid username or password")
	}

	totalPrice, err := calculateOrderTotal(req, orderServer.CatalogServiceAPI)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, err.Error())
	}

	itemsByte, err := json.Marshal(req.MenuItems)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, err.Error())
	}

	itemStr := string(itemsByte)

	order := &model.Order{
		Username:     username,
		RestaurantId: req.RestaurantId,
		MenuItems:    itemStr,
		TotalPrice:   totalPrice,
	}

	err = orderServer.DB.Create(&order).Error
	if err != nil {
		errorString := fmt.Sprintf("error storing the order: %v", err)
		return nil, status.Errorf(codes.Unknown, errorString)
	}

	restaurantAddress, err := fetchRestaurantAddress(req.RestaurantId, orderServer.CatalogServiceAPI)
	if err != nil {
		return nil, status.Errorf(codes.Canceled, err.Error())
	}

	requestBody, _ := json.Marshal(map[string]any{
		"orderId":       order.Id,
		"dropAddress":   user.Address,
		"pickupAddress": restaurantAddress,
	})

	reqBody := bytes.NewBuffer(requestBody)

	resp, err := http.Post(orderServer.FulfillmentServiceAPI, "application/json", reqBody)

	switch resp.StatusCode {
	case http.StatusConflict:
		return nil, status.Errorf(codes.Aborted, "OrderAlreadyAssignedException: %s", parseResponse(resp.Body))
	case http.StatusNotFound:
		return nil, status.Errorf(codes.NotFound, "NoDeliveryExecutiveNearbyException: %s", parseResponse(resp.Body))
	case http.StatusInternalServerError:
		log.Println("inside")
		return nil, status.Errorf(codes.Internal, "Internal Server Error: %s", parseResponse(resp.Body))
	default:
		break
	}

	response := &o.CreateOrderResponse{
		Id:           order.Id,
		Username:     username,
		RestaurantId: req.RestaurantId,
		MenuItems:    req.MenuItems,
		TotalPrice:   totalPrice,
	}

	return response, nil
}

func parseResponse(body io.Reader) string {
	responseBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return fmt.Sprintf("Error reading response body: %s", err.Error())
	}
	return string(responseBytes)
}

func fetchRestaurantAddress(restaurantId string, url string) (*model.Address, error) {
	// apiStr := fmt.Sprintf("http://localhost:8080/api/v1/restaurants/%v", r)
	apiStr := fmt.Sprintf(url + restaurantId)
	resp, _ := http.Get(apiStr)

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, status.Errorf(codes.NotFound, "NoSuchElementException: %s", parseResponse(resp.Body))
	case http.StatusInternalServerError:
		return nil, status.Errorf(codes.Internal, "Internal Server Error: %s", parseResponse(resp.Body))
	default:
		break
	}

	body := parseResponse(resp.Body)

	var response struct {
		Data struct {
			Restaurant struct {
				Address *model.Address `json:"address"`
			} `json:"restaurant"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return nil, err
	}

	return response.Data.Restaurant.Address, nil
}

func isAuthenticated(storedHashPassword, password string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(storedHashPassword), []byte(password)); err != nil {
		return false, nil
	}

	return true, nil
}

func extractCredentials(ctx context.Context) (string, string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "", false
	}

	authHeaders, ok := md["authorization"]

	if !ok || len(authHeaders) == 0 {
		return "", "", false
	}

	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Basic ") {
		return "", "", false
	}

	decoded, err := base64.StdEncoding.DecodeString(authHeader[6:])
	if err != nil {
		return "", "", false
	}

	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return "", "", false
	}

	return credentials[0], credentials[1], true
}

func calculateOrderTotal(req *o.CreateOrderRequest, url string) (float64, error) {
	restaurantID := req.RestaurantId
	total := 0.0

	for menuItemName, quantity := range req.MenuItems {
		apiString := fmt.Sprintf(url + restaurantID + "/menuItems/" + menuItemName)

		resp, _ := http.Get(apiString)

		if resp.StatusCode != http.StatusOK {
			return 0.0, errors.New("item not found")
		}

		body := parseResponse(resp.Body)

		var response struct {
			Data struct {
				MenuItem struct {
					Price float64 `json:"price"`
				} `json:"menu_item"`
			} `json:"data"`
		}

		if err := json.Unmarshal([]byte(body), &response); err != nil {
			return 0.0, err
		}

		price := response.Data.MenuItem.Price
		total += price * float64(quantity)
	}

	return total, nil
}
