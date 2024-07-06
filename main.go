package main

import (
	"fmt"
	"os"
	"zebra/routes"
	"zebra/shared"
)

func main() {
	shared.LoadEnv(".")
	startAPIServer()
}

func startAPIServer() {
	port := os.Getenv("API_PORT")
	fmt.Println("Starting API server on port", port)
	routes.SetupRoutes()
}
