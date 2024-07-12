package video_transcode_test

import (
	"log"
	"os"
	"testing"
	appmock "zebra/pkg/mock"
	"zebra/pkg/tests"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

func TestVideoTranscode(t *testing.T) {
	const (
		videoTranscodeTopic = "video.transcode.topic"
		videoUploadTopic    = "video.upload.topic"
	)

	// Open the test video file
	file, err := os.Open("../../testdata/" + tests.VIDEO_TEST_FILE_NAME)
	if err != nil {
		t.Fatalf("Failed to open test video file: %v", err)
	}
	defer file.Close()

	tests.VideoConsumerTst(t)

	// Set up mock Kafka for video transcode
	consumer, _, _ := appmock.SetupMockKafka(t, videoTranscodeTopic, "test_key", tests.VIDEO_TEST_FILE_NAME)

	// Consume the message from the video transcode topic
	partition, err := consumer.ConsumePartition(videoTranscodeTopic, 0, sarama.OffsetOldest)
	if err != nil {
		t.Fatalf("Failed to consume partition: %v", err)
	}
	defer partition.Close()

	// Read the message from the partition
	msg := <-partition.Messages()

	// Assert that the video file is the same as the test video file
	log.Printf("Received video file: %s\n", string(msg.Value))
	assert.Equal(t, tests.VIDEO_TEST_FILE_NAME, string(msg.Value))
}
