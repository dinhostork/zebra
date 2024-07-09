package shared

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadEnv loads the .env file from the root directory of the project
func LoadEnv() {
	// Get the root directory of the project
	rootDir, err := GetRootDir()
	if err != nil {
		println(err)
		panic("Error getting project root directory: " + err.Error())
	}

	// Load .env file from the root directory
	err = godotenv.Load(filepath.Join(rootDir, ".env"))
	if err != nil {
		println(err)
		panic("Error loading .env file: " + err.Error())
	}
}

// getRootDir returns the root directory of the project
func GetRootDir() (string, error) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Find the nearest directory containing go.mod (root directory)
	rootDir := wd
	for {
		if _, err := os.Stat(filepath.Join(rootDir, "go.mod")); err == nil {
			return rootDir, nil
		}
		if rootDir == "/" || rootDir == "" {
			break
		}
		rootDir = filepath.Dir(rootDir)
	}

	return "", fmt.Errorf("go.mod file not found in the project root directory or its parents")
}
