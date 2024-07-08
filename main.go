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
	envPath := flag.String("envPath", ".env", "path to .env file")
	envType := flag.String("envType", "development", "Specify the environment: development or production")

	flag.Parse()

	if *envType == "production" {
		go startService("video_transcode", "./build/video_transcode", *envPath)
		go startService("video_watermark", "./build/video_watermark", *envPath)
		go startService("video_consumer", "./build/video_consumer", *envPath)
	}

	shared.LoadEnv(*envPath)
	startAPIServer()
}

func startAPIServer() {
	port := os.Getenv("API_PORT")
	fmt.Println("Starting API server on port", port)
	routes.SetupRoutes()
}

func startService(name, path, envPath string) {
	fmt.Printf("Starting %s service\n", name)
	cmd := exec.Command(path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), fmt.Sprintf("envPath=%s", envPath))

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to start service %s: %v", name, err)
	}
}
