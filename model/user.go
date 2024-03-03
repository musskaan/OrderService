package model

type Address struct {
	Street  string
	City    string
	State   string
	Zipcode string
}

type User struct {
	Id       int64    `gorm:"primaryKey;autoIncrement:true" json:"id"`
	Username string   `json:"username" gorm:"unique"`
	Password string   `json:"password"`
	Address  *Address `json:"address" gorm:"embedded"`
}
