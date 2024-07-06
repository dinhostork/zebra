package configs

import (
	"os"
	"zebra/shared"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	db *gorm.DB
)

func Connect() {
	shared.LoadEnv("../..")

	user := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PASSWORD")
	host := os.Getenv("MYSQL_HOST")
	database := os.Getenv("MYSQL_DATABASE")

	connectionString := user + ":" + password + "@tcp(" + host + ")/" + database + "?charset=utf8&parseTime=True&loc=Local"

	d, err := gorm.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}
	db = d
}

func GetDB() *gorm.DB {
	return db
}
