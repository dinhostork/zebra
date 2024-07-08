#!/bin/bash
# Copy .env file to build folder
cp .env.production build/.env

# Build video_consumer
GOOS=linux GOARCH=amd64 go build -o build/video_consumer ./cmd/video_consumer

# Build video_transcode
GOOS=linux GOARCH=amd64 go build -o build/video_transcode ./cmd/video_transcode

# Build video_watermark
GOOS=linux GOARCH=amd64 go build -o build/video_watermark ./cmd/video_watermark

# Build main
GOOS=linux GOARCH=amd64 go build -o build/main .