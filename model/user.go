package model

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zipcode string `json:"zipcode"`
}

type User struct {
	Id       int64    `gorm:"primaryKey;autoIncrement:true" json:"id"`
	Username string   `json:"username" gorm:"unique"`
	Password string   `json:"password"`
	Address  *Address `json:"address" gorm:"embedded"`
}
