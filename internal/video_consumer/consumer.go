package video_consumer

import (
	"fmt"
	"log"
	"zebra/models"
	"zebra/shared"

	"github.com/IBM/sarama"
)

func HandleMessages(partitionConsumer sarama.PartitionConsumer) {
	shared.HandleMessages(partitionConsumer, handleMessage)
}

func handleMessage(msg *sarama.ConsumerMessage) {
	videoFile := string(msg.Value)
	fmt.Printf("Received video file: %s\n", videoFile)

	video := models.Video{
		OriginalFilePath: videoFile,
		FilePath:         videoFile,
		Failed:           false,
	}
	if err := video.Save(); err != nil {
		log.Printf("Error saving video to database: %v", err)
		return
	}

	fmt.Printf("Video '%s' saved to database\n", videoFile)
	if err := sendMessageToTranscodeService(videoFile); err != nil {
		log.Printf("Error sending message to transcode service: %v", err)
		return
	}

	fmt.Printf("Video '%s' sent to transcoding service\n", videoFile)
}

func sendMessageToTranscodeService(videoFile string) error {
	producer, err := shared.InitKafkaProducer(shared.VIDEO_TRANSCODE_TOPIC)
	if err != nil {
		return fmt.Errorf("error creating producer: %v", err)
	}
	defer producer.Close()

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: shared.VIDEO_TRANSCODE_TOPIC,
		Value: sarama.StringEncoder(videoFile),
	})
	if err != nil {
		return fmt.Errorf("error sending message to transcode service: %v", err)
	}

	return nil
}
