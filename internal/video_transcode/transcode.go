package video_transcode

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"zebra/models"

	"github.com/google/uuid"

	"github.com/IBM/sarama"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

func ProcessMessage(msg *sarama.ConsumerMessage) {
	video_id := string(msg.Value)
	fmt.Printf("Received video file: %s\n", video_id)
	videoFile, db := models.GetVideoById(video_id)
	defer db.Close()

	fmt.Printf("Transcoding video '%s'\n", video_id)
	if err := TranscodeVideo(*videoFile); err != nil {
		log.Printf("Failed to transcode video '%s': %v", strconv.Itoa(int(videoFile.ID)), err)
		return
	}
	fmt.Printf("Video '%s' transcoded successfully\n", strconv.Itoa(int(videoFile.ID)))
	defer RemoveFile(videoFile.TempFilePath)
}

func TranscodeVideo(video models.Video) error {
	formats := []string{".mp4"}
	for _, format := range formats {
		if err := transcodeToFormat(video, format); err != nil {
			return fmt.Errorf("error transcoding video to %s: %v", format, err)
		}
	}
	return nil
}

func transcodeToFormat(video models.Video, format string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	// create a unique name for the transcoded file
	baseName := uuid.New().String()

	//baseName := strings.TrimSuffix(filepath.Base(video.TempFilePath), filepath.Ext(video.TempFilePath))
	outputFile := filepath.Join(currentDir, baseName+format)

	stream := ffmpeg_go.Input(video.TempFilePath)

	switch format {
	case ".mp4":
		stream = stream.Output(outputFile, ffmpeg_go.KwArgs{"c:v": "libx264", "crf": 22, "preset": "fast", "vcodec": "h264"})
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	stream = stream.GlobalArgs("-hide_banner", "-y")

	if err := stream.Run(); err != nil {
		return fmt.Errorf("error running ffmpeg for %s: %v", format, err)
	}

	// save transcoded path to database
	video.TranscodedPath = &outputFile
	models.UpdateVideo(video)
	return nil
}

func RemoveFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("error removing file %s: %v", filePath, err)
	}
	return nil
}
