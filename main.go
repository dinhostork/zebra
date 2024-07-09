package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"zebra/routes"
	"zebra/shared"
)

func main() {

	envType := flag.String("envType", "development", "Specify the environment: development or production")

	flag.Parse()

	if *envType == "production" {
		go startService("video_transcode", "./build/video_transcode")
		go startService("video_watermark", "./build/video_watermark")
		go startService("video_consumer", "./build/video_consumer")
	}

	shared.LoadEnv()
	startAPIServer()
}

func startAPIServer() {
	port := os.Getenv("API_PORT")
	fmt.Println("Starting API server on port", port)
	routes.SetupRoutes()
}

func startService(name, path string) {
	cmd := exec.Command(path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to start service %s: %v", name, err)
	}
}
