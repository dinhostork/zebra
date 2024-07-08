package video_watermark

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"zebra/models"
	"zebra/shared"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

func ProcessMessage(msg *sarama.ConsumerMessage) {
	video_id := string(msg.Value)

	fmt.Printf("Received video file: %s\n", video_id)

	// Get the video file from the database
	videoFile, db := models.GetVideoById(video_id)

	defer db.Close()

	fmt.Printf("Adding watermark to video '%s'\n", strconv.Itoa(int(videoFile.ID)))
	if err := addWatermark(*videoFile); err != nil {
		log.Printf("Failed to add watermark to video '%s': %v", strconv.Itoa(int(videoFile.ID)), err)
		return
	}
	fmt.Printf("Watermark added to video '%s'\n", strconv.Itoa(int(videoFile.ID)))

}

func addWatermark(video models.Video) error {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	// Create a unique name for the watermarked file
	baseName := uuid.New().String()

	// Separate the original video's base name without extension
	outputFile := filepath.Join(currentDir, baseName+".mp4")    // Full path to the watermarked video file
	watermarkFile := filepath.Join(currentDir, "watermark.png") // Full path to the watermark image file

	// Calculate video duration
	totalVideoDuration, err := getVideoDuration(*video.TranscodedPath)
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
	cmd1 := exec.Command("ffmpeg", "-i", *video.TranscodedPath, "-i", watermarkFile, "-filter_complex",
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

	// saves the watermarked video path to the database
	saveWatermarkedVideo(video, outputFile)
	if err != nil {
		return fmt.Errorf("error renaming final watermarked video: %v", err)
	}

	return nil
}

func saveWatermarkedVideo(video models.Video, path string) error {
	file, err := os.Open(path)

	if err != nil {
		return fmt.Errorf("error opening file %s: %v", path, err)
	}

	url, err := shared.UploadVideoToS3(file, path, true)
	if err != nil {
		return fmt.Errorf("error uploading watermarked video to S3: %v", err)
	}

	video.Path = &url
	now := time.Now()
	video.ProcessedAt = &now
	models.UpdateVideo(video)
	fmt.Printf("Saving watermarked video '%s'\n", strconv.Itoa(int(video.ID)))

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
