package configs

import (
	"os"
	"zebra/shared"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	// Disable warnings for sqlite3 package
	// #cgo CFLAGS: -Wno-return-local-addr
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
	db *gorm.DB
)

func Connect() {
	shared.LoadEnv()
	user := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PASSWORD")
	host := os.Getenv("MYSQL_HOST")
	database := os.Getenv("MYSQL_DATABASE")
	var err error

	connectionString := user + ":" + password + "@tcp(" + host + ")/" + database + "?charset=utf8&parseTime=True&loc=Local"

	rootDir, err := shared.GetRootDir()
	if err != nil {
		panic(err)
	}

	if os.Getenv("ENV") == "test" {
		db, err = gorm.Open("sqlite3", rootDir+"/test.db")
	} else {
		db, err = gorm.Open("mysql", connectionString)
	}
	if err != nil {
		panic(err)
	}
}

func GetDB() *gorm.DB {
	if db == nil || db.DB().Ping() != nil {
		Connect()
	}
	return db
}
