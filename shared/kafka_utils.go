package shared

import (
	"os"

	"github.com/IBM/sarama"
)

var (
	mockProducer sarama.SyncProducer
	mockConsumer sarama.Consumer
)

func InitKafkaProducer(topic string) (sarama.SyncProducer, error) {
	LoadEnv()
	kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaPort := os.Getenv("KAFKA_PORT")

	if mockProducer != nil {
		return mockProducer, nil
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{kafkaHost + ":" + kafkaPort}, config)
	if err != nil {
		return nil, err
	}

	return producer, nil
}

func InitKafkaConsumer() (sarama.Consumer, error) {
	LoadEnv()
	kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaPort := os.Getenv("KAFKA_PORT")

	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	brokers := []string{kafkaHost + ":" + kafkaPort}
	if mockConsumer != nil {
		return mockConsumer, nil
	}
	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

func SetKafkaProducer(producer sarama.SyncProducer) {
	mockProducer = producer
}

func SetKafkaConsumer(consumer sarama.Consumer) {
	mockConsumer = consumer
}
