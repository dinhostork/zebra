package video_transcode_test

import (
	"os"
	"testing"
	"zebra/internal/video_consumer"
	appmock "zebra/pkg/mock"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

func TestVideoTranscode(t *testing.T) {
	const (
		videoTranscodeTopic = "video.transcode.topic"
		videoUploadTopic    = "video.upload.topic"
		videoFileName       = "test_video_file.mp4"
	)

	// Open the test video file
	file, err := os.Open("../../testdata/" + videoFileName)
	if err != nil {
		t.Fatalf("Failed to open test video file: %v", err)
	}
	defer file.Close()

	// Set up mock Kafka for video upload
	appmock.SetupMockKafka(t, videoUploadTopic, "test_key", file.Name())

	// Create a video consumer message
	video := &sarama.ConsumerMessage{
		Value: []byte(file.Name()),
	}

	// Handle the video message
	videoHandler := &video_consumer.DefaultMessageHandler{}
	videoHandler.HandleMessage(video)

	// Set up mock Kafka for video transcode
	consumer, _, _ := appmock.SetupMockKafka(t, videoTranscodeTopic, "test_key", videoFileName)

	// Consume the message from the video transcode topic
	partition, err := consumer.ConsumePartition(videoTranscodeTopic, 0, sarama.OffsetOldest)
	if err != nil {
		t.Fatalf("Failed to consume partition: %v", err)
	}
	defer partition.Close()

	// Read the message from the partition
	msg := <-partition.Messages()

	// Assert that the video file is the same as the test video file
	assert.Equal(t, videoFileName, string(msg.Value))
}
