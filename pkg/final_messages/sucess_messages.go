package final_messages

import (
	"encoding/json"
	"log"
	configs "zebra/configs/database"
	"zebra/models"
	"zebra/shared"

	"github.com/IBM/sarama"
)

func SendSuccessMessage(video models.Video) {
	producer, err := shared.InitKafkaProducer(shared.SUCCESS_TOPIC)
	if err != nil {
		log.Printf("Error creating success producer: %v", err)
		return
	}

	defer producer.Close()

	parsedVideo, err := json.Marshal(video)

	log.Printf("Sending success message: %v", string(parsedVideo))
	if err != nil {
		log.Printf("Error parsing video to JSON: %v", err)
		return
	}

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: shared.SUCCESS_TOPIC,
		Value: sarama.StringEncoder(parsedVideo),
	})

	if err != nil {
		log.Printf("Error sending success message: %v", err)
	}

	db := configs.GetDB()
	defer db.Close()

}
