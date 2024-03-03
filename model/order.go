package model

type Order struct {
	Id           int64   `json:"id" gorm:"primaryKey;autoIncrement:true"`
	RestaurantId string  `json:"restaurant_id"`
	Username     string  `json:"username"`
	TotalPrice   float64 `json:"total_price"`
	MenuItems    string  `json:"menu_items"`
}
