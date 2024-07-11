package shared

import (
	"os"

	"github.com/IBM/sarama"
)

var mockProducer sarama.SyncProducer

func InitKafkaProducer(topic string) (sarama.SyncProducer, error) {
	LoadEnv()
	kafkaHost := os.Getenv("KAFKA_HOST")
	kafaPort := os.Getenv("KAFKA_PORT")

	if mockProducer != nil {
		return mockProducer, nil
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{kafkaHost + ":" + kafaPort}, config)
	if err != nil {
		return nil, err
	}

	return producer, nil
}

func InitKafkaConsumer() (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	brokers := []string{"localhost:9092"}
	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

func SetKafkaProducer(producer sarama.SyncProducer) {
	mockProducer = producer
}
