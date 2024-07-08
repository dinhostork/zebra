package final_messages

import (
	"encoding/json"
	"log"
	configs "zebra/configs/database"
	"zebra/models"
	"zebra/pkg/utils"
	"zebra/shared"

	"github.com/IBM/sarama"
)

func SendErrorMessage(video models.Video, failMessage string) {
	producer, err := shared.InitKafkaProducer(shared.ERROR_TOPIC)

	if err != nil {
		log.Printf("Error creating error producer: %v", err)
		return
	}

	defer producer.Close()

	parsedVideo, err := json.Marshal(struct {
		Video       models.Video `json:"video"`
		FailMessage string       `json:"fail_message"`
	}{
		Video:       video,
		FailMessage: failMessage,
	})

	log.Printf("Sending error message: %v", string(parsedVideo))
	if err != nil {
		log.Printf("Error parsing video to JSON: %v", err)
		return
	}

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: shared.ERROR_TOPIC,
		Value: sarama.StringEncoder(parsedVideo),
	})

	if err != nil {
		log.Printf("Error sending error message: %v", err)
	}

	utils.RemoveFile(*video.TranscodedPath)
	db := configs.GetDB()
	defer db.Close()

}
