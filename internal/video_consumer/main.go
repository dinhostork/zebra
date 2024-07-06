package main

import (
	"fmt"
	"log"
	"zebra/models"
	"zebra/shared"

	"github.com/IBM/sarama"
)

func main() {
	fmt.Println("Starting video consumer")
	shared.LoadEnv("../..")

	consumer, err := shared.InitKafkaConsumer()
	if err != nil {
		log.Fatalf("Failed to initialize Kafka consumer: %v", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(shared.VIDEO_UPLOAD_TOPIC, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to start partition consumer: %v", err)
	}
	defer partitionConsumer.Close()

	handleMessages(partitionConsumer)
}

func handleMessages(partitionConsumer sarama.PartitionConsumer) {
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			processMessage(msg)
		case err := <-partitionConsumer.Errors():
			log.Printf("Error consuming message: %v", err)
		}
	}
}

func processMessage(msg *sarama.ConsumerMessage) {
	videoFile := string(msg.Value)
	fmt.Printf("Received video file: %s\n", videoFile)

	createVideo := models.Video{
		OrginalFilePath: videoFile,
		FilePath:        videoFile,
		Failed:          false,
	}
	if err := createVideo.Save(); err != nil {
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
