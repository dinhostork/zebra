package video_transcode

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"zebra/models"
	"zebra/pkg/final_messages"
	"zebra/shared"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

func ProcessMessage(msg *sarama.ConsumerMessage) {
	videoID := string(msg.Value)

	fmt.Printf("Received video file: %s\n", videoID)
	videoFile, err := models.GetVideoById(videoID)
	if err != nil {
		log.Printf("Failed to get video '%s' from database: %v", videoID, err)
		return
	}

	fmt.Printf("Transcoding video '%s'\n", videoID)
	if err := TranscodeVideo(*videoFile); err != nil {
		log.Printf("Failed to transcode video '%s': %v", strconv.Itoa(int(videoFile.ID)), err)
		videoFile.Failed = true
		errorMessage := fmt.Sprintf("Failed to transcode video: %v", err)
		videoFile.FailedMessage = &errorMessage
		final_messages.SendErrorMessage(*videoFile, err.Error())
		return
	}
	fmt.Printf("Video '%s' transcoded successfully\n", strconv.Itoa(int(videoFile.ID)))

	SendToWatermarkService(*videoFile)
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

	// create a unique name for the transcoded file
	baseName := uuid.New().String()
	root, err := shared.GetRootDir()
	if err != nil {
		return fmt.Errorf("error getting root directory: %v", err)
	}

	outputFile := filepath.Join(root, "temp", "transcoded", baseName+format)

	stream := ffmpeg_go.Input(video.TempFilePath)

	switch format {
	case ".mp4":
		stream = stream.Output(outputFile, ffmpeg_go.KwArgs{"c:v": "libxfffsfs264", "crf": 22, "preset": "fast", "vcodec": "h264"})
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	stream = stream.GlobalArgs("-hide_banner", "-y")

	if err := stream.Run(); err != nil {
		return fmt.Errorf("error running ffmpeg for %s: %v", format, err)
	}

	// upload transcoded file to S3
	file, err := os.Open(outputFile)
	if err != nil {
		return fmt.Errorf("error opening file %s: %v", outputFile, err)
	}

	url, err := shared.UploadVideoToS3(file, baseName+format, false)
	if err != nil {
		return fmt.Errorf("error uploading file to S3: %v", err)
	}

	// save transcoded path to database
	video.TranscodedPath = &outputFile
	video.TranscodedUrl = &url
	if _, err := models.UpdateVideo(video); err != nil {
		return fmt.Errorf("error updating video in database: %v", err)
	}

	return nil
}

func RemoveFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("error removing file %s: %v", filePath, err)
	}
	return nil
}

func SendToWatermarkService(video models.Video) {
	producer, err := shared.InitKafkaProducer(shared.VIDEO_WATERMARK_TOPIC)
	if err != nil {
		log.Printf("Error creating producer: %v", err)
		video.Failed = true
		errorMessage := fmt.Sprintf("Error creating producer: %v", err)
		video.FailedMessage = &errorMessage
		final_messages.SendErrorMessage(video, err.Error())
		return
	}

	defer producer.Close()

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: shared.VIDEO_WATERMARK_TOPIC,
		Value: sarama.StringEncoder(strconv.Itoa(int(video.ID))),
	})
	if err != nil {
		log.Printf("Error sending message to watermark service: %v", err)
		video.Failed = true
		errorMessage := fmt.Sprintf("Error sending message to watermark service: %v", err)
		video.FailedMessage = &errorMessage
		final_messages.SendErrorMessage(video, err.Error())
		return
	}

	fmt.Printf("Video '%s' sent to watermark service\n", strconv.Itoa(int(video.ID)))
}
