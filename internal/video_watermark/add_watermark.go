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
	"zebra/pkg/final_messages"
	"zebra/pkg/utils"
	"zebra/shared"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

func ProcessMessage(msg *sarama.ConsumerMessage) {
	videoID := string(msg.Value)

	fmt.Printf("Received video file: %s\n", videoID)

	// Get the video file from the database
	videoFile, err := models.GetVideoById(videoID)

	if err != nil {
		log.Printf("Failed to get video '%s' from database: %v", videoID, err)
		return
	}

	fmt.Printf("Adding watermark to video '%s'\n", strconv.Itoa(int(videoFile.ID)))
	if err := addWatermark(*videoFile); err != nil {
		log.Printf("Failed to add watermark to video '%s': %v", strconv.Itoa(int(videoFile.ID)), err)
		videoFile.Failed = true
		errorMessage := fmt.Sprintf("Failed to add watermark to video: %v", err)
		videoFile.FailedMessage = &errorMessage
		final_messages.SendErrorMessage(*videoFile, err.Error())
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
	watermarkFile := filepath.Join(currentDir, "watermark.svg") // Full path to the watermark image file

	// Get video resolution
	width, height, err := getVideoResolution(*video.TranscodedPath)
	if err != nil {
		return fmt.Errorf("error getting video resolution: %v", err)
	}

	// Resize watermark based on video resolution
	resizedWatermarkFile := filepath.Join(currentDir, baseName+"_resized_watermark.png")
	if err := resizeWatermark(watermarkFile, resizedWatermarkFile, width, height); err != nil {
		return fmt.Errorf("error resizing watermark: %v", err)
	}

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
		os.Remove(resizedWatermarkFile)
	}()

	// Create ffmpeg command for each watermark position
	cmd1 := exec.Command("ffmpeg", "-i", *video.TranscodedPath, "-i", resizedWatermarkFile, "-filter_complex",
		fmt.Sprintf("[0:v][1:v]overlay=5:5:enable='between(t,%f,%f)'", part1Start, part1End),
		"-c:a", "copy", "-c:v", "libx264", "-preset", "veryfast", intermediateFile1)

	cmd2 := exec.Command("ffmpeg", "-i", intermediateFile1, "-i", resizedWatermarkFile, "-filter_complex",
		fmt.Sprintf("[0:v][1:v]overlay='(main_w-overlay_w-5):5':enable='between(t,%f,%f)'", part2Start, part2End),
		"-c:a", "copy", "-c:v", "libx264", "-preset", "veryfast", intermediateFile2)

	cmd3 := exec.Command("ffmpeg", "-i", intermediateFile2, "-i", resizedWatermarkFile, "-filter_complex",
		fmt.Sprintf("[0:v][1:v]overlay=5:'(main_h-overlay_h-5)':enable='between(t,%f,%f)'", part3Start, part3End),
		"-c:a", "copy", "-c:v", "libx264", "-preset", "veryfast", intermediateFile3)

	cmd4 := exec.Command("ffmpeg", "-i", intermediateFile3, "-i", resizedWatermarkFile, "-filter_complex",
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
	if err := saveWatermarkedVideo(video, outputFile); err != nil {
		return fmt.Errorf("error saving watermarked video: %v", err)
	}
	if err != nil {
		return fmt.Errorf("error renaming final watermarked video: %v", err)
	}

	return nil
}

func getVideoResolution(videoFile string) (int, int, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "csv=s=x:p=0", videoFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, fmt.Errorf("error getting video resolution: %v. Output: %s", err, output)
	}

	resolutionStr := strings.TrimSpace(string(output))
	resolution := strings.Split(resolutionStr, "x")
	if len(resolution) != 2 {
		return 0, 0, fmt.Errorf("invalid resolution format: %s", resolutionStr)
	}

	width, err := strconv.Atoi(resolution[0])
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing video width: %v", err)
	}

	height, err := strconv.Atoi(resolution[1])
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing video height: %v", err)
	}

	return width, height, nil
}

func resizeWatermark(inputFile, outputFile string, videoWidth, videoHeight int) error {
	// Assume original watermark size is 100x100 pixels
	originalWidth := 100
	originalHeight := 100

	// Determine new watermark size based on video resolution
	scaleFactor := float64(videoWidth) / 10.0 / float64(originalWidth)
	newWidth := int(float64(originalWidth) * scaleFactor)
	newHeight := int(float64(originalHeight) * scaleFactor)

	cmd := exec.Command("ffmpeg", "-i", inputFile, "-vf", fmt.Sprintf("scale=%d:%d", newWidth, newHeight), outputFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error resizing watermark: %v. Output: %s", err, output)
	}

	return nil
}

func saveWatermarkedVideo(video models.Video, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file %s: %v", path, err)
	}

	filename := filepath.Base(path)

	url, err := shared.UploadVideoToS3(file, filename, true)
	if err != nil {
		return fmt.Errorf("error uploading watermarked video to S3: %v", err)
	}

	video.Url = &url
	now := time.Now()
	video.ProcessedAt = &now
	models.UpdateVideo(video)
	fmt.Printf("Saving watermarked video '%s'\n", strconv.Itoa(int(video.ID)))

	log.Printf("Watermarked video path: %s\n", path)
	// Remove the original video file
	utils.RemoveFile(path)
	// Send success message
	final_messages.SendSuccessMessage(video)
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
