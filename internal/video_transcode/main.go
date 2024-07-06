package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"zebra/shared"

	"github.com/IBM/sarama"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

func main() {
	fmt.Println("Starting video transcode service")
	shared.LoadEnv("../..")

	consumer, err := shared.InitKafkaConsumer()
	if err != nil {
		log.Fatalf("Failed to initialize Kafka consumer: %v", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(shared.VIDEO_TRANSCODE_TOPIC, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to start partition consumer: %v", err)
	}
	defer partitionConsumer.Close()

	shared.HandleMessages(partitionConsumer, processMessage)
}

func processMessage(msg *sarama.ConsumerMessage) {
	videoFile := string(msg.Value)
	fmt.Printf("Received video file: %s\n", videoFile)

	fmt.Printf("Transcoding video '%s'\n", videoFile)
	if err := transcodeVideo(videoFile); err != nil {
		log.Printf("Failed to transcode video '%s': %v", videoFile, err)
		return
	}
	fmt.Printf("Video '%s' transcoded successfully\n", videoFile)

	// at this point, we would save the original video file to a storage service
	if err := saveOriginalVideo(videoFile); err != nil {
		log.Printf("Error saving original video file: %v", err)
		return
	}

	if err := sendMessageToWatermarkService(videoFile); err != nil {
		log.Printf("Error sending message to watermark service: %v", err)
		return
	}

	fmt.Printf("Video '%s' sent to watermark service\n", videoFile)
}

func transcodeVideo(videoFile string) error {
	formats := []string{".mp4"}
	for _, format := range formats {
		if err := transcodeToFormat(videoFile, format); err != nil {
			return fmt.Errorf("error transcoding video to %s: %v", format, err)
		}
	}
	return nil
}

func transcodeToFormat(videoFile, format string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	baseName := strings.TrimSuffix(filepath.Base(videoFile), filepath.Ext(videoFile))
	outputFile := filepath.Join(currentDir, baseName+format)

	stream := ffmpeg_go.Input(videoFile)

	switch format {
	case ".mp4":
		stream = stream.Output(outputFile, ffmpeg_go.KwArgs{"c:v": "libx264", "crf": 23, "preset": "medium"})
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	stream = stream.GlobalArgs("-hide_banner", "-y")

	if err := stream.Run(); err != nil {
		return fmt.Errorf("error running ffmpeg for %s: %v", format, err)
	}
	return nil
}

func sendMessageToWatermarkService(videoFile string) error {
	producer, err := shared.InitKafkaProducer(shared.VIDEO_WATERMARK_TOPIC)
	if err != nil {
		return fmt.Errorf("error creating producer: %v", err)
	}
	defer producer.Close()

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: shared.VIDEO_WATERMARK_TOPIC,
		Value: sarama.StringEncoder(videoFile),
	})
	if err != nil {
		return fmt.Errorf("error sending message to watermark service: %v", err)
	}

	return nil
}

func saveOriginalVideo(videoFile string) error {
	// Save the original video file somewhere
	return nil
}
