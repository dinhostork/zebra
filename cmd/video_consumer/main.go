// main.go
package main

import (
	"fmt"
	"log"
	"zebra/internal/video_consumer"
	"zebra/shared"

	"github.com/IBM/sarama"
)

func main() {
	fmt.Println("Starting video consumer")
	shared.LoadEnv()

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

	handler := &video_consumer.DefaultMessageHandler{}
	video_consumer.HandleMessages(partitionConsumer, handler)
}
