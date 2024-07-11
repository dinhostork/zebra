package shared

import (
	"log"

	"github.com/IBM/sarama"
)

// HandleMessages consumes messages from a Kafka partition consumer
func HandleMessages(partitionConsumer sarama.PartitionConsumer, processMessage func(msg *sarama.ConsumerMessage) *sarama.ConsumerMessage) {
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			processMessage(msg)
		case err := <-partitionConsumer.Errors():
			log.Printf("Error consuming message: %v", err)
		}
	}
}
