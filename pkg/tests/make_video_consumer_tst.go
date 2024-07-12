package tests

import (
	"testing"
	"zebra/internal/video_consumer"
	appmock "zebra/pkg/mock"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

func VideoConsumerTst(t *testing.T) {

	const videoUploadTopic = "video.upload.topic"

	appmock.SetupMockKafka(t, videoUploadTopic, "test_key", VIDEO_TEST_FILE_NAME)

	testMessage := &sarama.ConsumerMessage{
		Value: []byte(VIDEO_TEST_FILE_NAME),
	}

	// Simulate the consumer handling the message
	handler := &video_consumer.DefaultMessageHandler{}
	lastMsg := handler.HandleMessage(testMessage)
	// Verify the handler processed the message correctly
	assert.Equal(t, string(testMessage.Value), string(lastMsg.Value), "Message value should match")

}
