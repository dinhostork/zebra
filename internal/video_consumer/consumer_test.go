package video_consumer_test

import (
	"testing"
	"zebra/internal/video_consumer"
	appmock "zebra/pkg/mock"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

const ONE_SECOND = 1000000000

var videoUploadTopic = "video.upload.topic"

func TestVideoConsumer(t *testing.T) {
	appmock.SetupMockKafka(t, videoUploadTopic, "test_key", "test_video_file.mp4")

	testMessage := &sarama.ConsumerMessage{
		Value: []byte("test_video_file.mp4"),
	}

	// Simulate the consumer handling the message
	handler := &video_consumer.DefaultMessageHandler{}
	lastMsg := handler.HandleMessage(testMessage)
	// Verify the handler processed the message correctly
	assert.Equal(t, string(testMessage.Value), string(lastMsg.Value), "Message value should match")
}
