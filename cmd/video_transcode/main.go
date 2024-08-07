package main

import (
	"fmt"
	"log"
	"zebra/internal/video_transcode"
	"zebra/shared"

	"github.com/IBM/sarama"
)

func main() {

	fmt.Println("Starting video transcode service")
	shared.LoadEnv()

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

	handler := &video_transcode.DefaultMessageHandler{}
	video_transcode.HandleMessages(partitionConsumer, handler)
}
