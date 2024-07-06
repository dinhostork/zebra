package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"zebra/shared"

	"github.com/IBM/sarama"
)

func main() {
	fmt.Println("Starting video watermark service")
	shared.LoadEnv("../..")

	consumer, err := shared.InitKafkaConsumer()
	if err != nil {
		log.Fatalf("Failed to initialize Kafka consumer: %v", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(shared.VIDEO_WATERMARK_TOPIC, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to start partition consumer: %v", err)
	}
	defer partitionConsumer.Close()

	shared.HandleMessages(partitionConsumer, processMessage)
}

func processMessage(msg *sarama.ConsumerMessage) {
	videoFile := string(msg.Value)
	fmt.Printf("Received video file: %s\n", videoFile)

	fmt.Printf("Adding watermark to video '%s'\n", videoFile)
	if err := addWatermark(videoFile); err != nil {
		log.Printf("Failed to add watermark to video '%s': %v", videoFile, err)
		return
	}
	fmt.Printf("Watermark added to video '%s'\n", videoFile)

	// At this point, we would save the watermarked video file to a storage service
	if err := saveWatermarkedVideo(videoFile); err != nil {
		log.Printf("Error saving watermarked video file: %v", err)
		return
	}

	fmt.Printf("Watermarked video '%s' saved\n", videoFile)
}

func addWatermark(videoFile string) error {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	// Separate the original video's base name without extension
	baseName := strings.TrimSuffix(filepath.Base(videoFile), filepath.Ext(videoFile))
	outputFile := filepath.Join(currentDir, baseName+"_watermarked.mp4")
	watermarkFile := filepath.Join(currentDir, "watermark.png") // Full path to the watermark image file

	// Calculate video duration
	totalVideoDuration, err := getVideoDuration(videoFile)
	if err != nil {
		return fmt.Errorf("error getting video duration: %v", err)
	}

	// Calculate timestamps for watermark appearance
	part1Start := 0.0
	part1End := totalVideoDuration / 4.0
	part2Start := part1End
	part2End := totalVideoDuration / 2.0
	part3Start := part2End
	part3End := totalVideoDuration * 3.0 / 4.0
	part4Start := part3End
	part4End := totalVideoDuration

	// Create intermediate filenames
	intermediateFile1 := filepath.Join(currentDir, baseName+"_part1.mp4")
	intermediateFile2 := filepath.Join(currentDir, baseName+"_part2.mp4")
	intermediateFile3 := filepath.Join(currentDir, baseName+"_part3.mp4")
	intermediateFile4 := filepath.Join(currentDir, baseName+"_part4.mp4")

	// Defer function to remove intermediate files on function exit or error
	defer func() {
		os.Remove(intermediateFile1)
		os.Remove(intermediateFile2)
		os.Remove(intermediateFile3)
		os.Remove(intermediateFile4)
	}()

	// Create ffmpeg command for each watermark position
	cmd1 := exec.Command("ffmpeg", "-i", videoFile, "-i", watermarkFile, "-filter_complex",
		fmt.Sprintf("[0:v][1:v]overlay=5:5:enable='between(t,%f,%f)'", part1Start, part1End),
		"-c:a", "copy", "-c:v", "libx264", "-preset", "veryfast", intermediateFile1)

	cmd2 := exec.Command("ffmpeg", "-i", intermediateFile1, "-i", watermarkFile, "-filter_complex",
		fmt.Sprintf("[0:v][1:v]overlay='(main_w-overlay_w-5):5':enable='between(t,%f,%f)'", part2Start, part2End),
		"-c:a", "copy", "-c:v", "libx264", "-preset", "veryfast", intermediateFile2)

	cmd3 := exec.Command("ffmpeg", "-i", intermediateFile2, "-i", watermarkFile, "-filter_complex",
		fmt.Sprintf("[0:v][1:v]overlay=5:'(main_h-overlay_h-5)':enable='between(t,%f,%f)'", part3Start, part3End),
		"-c:a", "copy", "-c:v", "libx264", "-preset", "veryfast", intermediateFile3)

	cmd4 := exec.Command("ffmpeg", "-i", intermediateFile3, "-i", watermarkFile, "-filter_complex",
		fmt.Sprintf("[0:v][1:v]overlay='(main_w-overlay_w-5):(main_h-overlay_h-5)':enable='between(t,%f,%f)'", part4Start, part4End),
		"-c:a", "copy", "-c:v", "libx264", "-preset", "veryfast", intermediateFile4)

	// Execute commands sequentially
	for _, cmd := range []*exec.Cmd{cmd1, cmd2, cmd3, cmd4} {
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error adding watermark to video: %v. Output: %s", err, output)
		}
	}

	// Rename the final intermediate file to outputFile
	err = os.Rename(intermediateFile4, outputFile)
	if err != nil {
		return fmt.Errorf("error renaming final watermarked video: %v", err)
	}

	return nil
}

func saveWatermarkedVideo(videoFile string) error {
	// Implement logic to save the watermarked video to a storage service
	// For now, we'll just log a message
	fmt.Printf("Saving watermarked video '%s'\n", videoFile)
	return nil
}

func getVideoDuration(videoFile string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", videoFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("error getting video duration: %v. Output: %s", err, output)
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing video duration: %v", err)
	}

	return duration, nil
}
