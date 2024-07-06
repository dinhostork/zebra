package video_transcode

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/IBM/sarama"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

func ProcessMessage(msg *sarama.ConsumerMessage) {
	videoFile := string(msg.Value)
	fmt.Printf("Received video file: %s\n", videoFile)

	fmt.Printf("Transcoding video '%s'\n", videoFile)
	if err := TranscodeVideo(videoFile); err != nil {
		log.Printf("Failed to transcode video '%s': %v", videoFile, err)
		return
	}
	fmt.Printf("Video '%s' transcoded successfully\n", videoFile)

	// At this point, you might want to save the original video file or do other processing
}

func TranscodeVideo(videoFile string) error {
	formats := []string{".mp4"}
	for _, format := range formats {
		if err := transcodeToFormat(videoFile, format); err != nil {
			return fmt.Errorf("error transcoding video to %s: %v", format, err)
		}
	}
	return nil
}

func transcodeToFormat(videoFile, format string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	baseName := strings.TrimSuffix(filepath.Base(videoFile), filepath.Ext(videoFile))
	outputFile := filepath.Join(currentDir, baseName+format)

	stream := ffmpeg_go.Input(videoFile)

	switch format {
	case ".mp4":
		stream = stream.Output(outputFile, ffmpeg_go.KwArgs{"c:v": "libx264", "crf": 22, "preset": "fast", "vcodec": "h264"})
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	stream = stream.GlobalArgs("-hide_banner", "-y")

	if err := stream.Run(); err != nil {
		return fmt.Errorf("error running ffmpeg for %s: %v", format, err)
	}
	return nil
}
