package main

import (
	"fmt"
	"zebra/models"
	"zebra/shared"

	"github.com/IBM/sarama"
)

func main() {
	fmt.Println("Starting video consumer")
	shared.LoadEnv("../..")
	consumer, err := shared.InitKafkaConsumer()
	if err != nil {
		panic(err)
	}

	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(shared.VIDEO_UPLOAD_TOPIC, 0, sarama.OffsetNewest)
	if err != nil {
		panic(err)
	}

	defer partitionConsumer.Close()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			videoFile := string(msg.Value)
			fmt.Printf("Received video file: %s\n", videoFile)

			// save video to database
			createVideo := models.Video{
				OrginalFilePath: videoFile,
				FilePath:        videoFile,
				Failed:          false,
			}
			createVideo.Save()

			// send video to transcoding service
			// (placeholder for your transcoding logic)
			fmt.Printf("Video '%s' saved to database\n", videoFile)

			// remove from Kafka topic

		case err := <-partitionConsumer.Errors():
			fmt.Printf("Error consuming message: %v\n", err)
		}
	}
}
