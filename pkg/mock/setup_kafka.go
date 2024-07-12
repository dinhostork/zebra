package appmock

import (
	"testing"
	"zebra/shared"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	log "github.com/sirupsen/logrus"
)

// Set up a mock Kafka broker, consumer, and producer for testing
func SetupMockKafka(t *testing.T, topic string, key string, value string) (*mocks.Consumer, *mocks.SyncProducer, *sarama.MockBroker) {
	log.SetLevel(log.DebugLevel)
	sarama.Logger = log.StandardLogger()

	// Given a mock Kafka broker
	mockBroker := sarama.NewMockBroker(t, 0)

	// Set up mock consumer
	mockConsumer := mocks.NewConsumer(t, nil)
	mockProducer := mocks.NewSyncProducer(t, nil)

	// Set up expectations
	partition := int32(0)
	mockConsumer.ExpectConsumePartition(topic, partition, sarama.OffsetOldest).YieldMessage(&sarama.ConsumerMessage{
		Key:       []byte(key),
		Value:     []byte(value),
		Topic:     topic,
		Partition: partition,
		Offset:    0,
	})

	// Set the mock Kafka consumer and producer in shared package
	shared.SetKafkaConsumer(mockConsumer)
	shared.SetKafkaProducer(mockProducer)
	mockProducer.ExpectSendMessageAndSucceed()

	return mockConsumer, mockProducer, mockBroker
}
