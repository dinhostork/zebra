package shared

import (
	"github.com/joho/godotenv"
)

// LoadEnv loads the .env file
func LoadEnv(path string) {
	// loads dot env from the root directory
	err := godotenv.Load(path)
	if err != nil {
		println(err)
		panic("Error loading .env file" + err.Error())
	}
}
