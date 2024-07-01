package videoconsumer

import (
	"io/ioutil"
	"zebra/models"
)

func ProcessVideo(videoFile string) ([]byte, error) {
	video, err := ioutil.ReadFile(videoFile)
	if err != nil {
		return nil, err
	}

	// save video to database
	createVideo := models.Video{
		OrginalFilePath: videoFile,
		FilePath:        videoFile,
		Failed:          false,
	}
	createVideo.Save()

	// send video to processing service
	return video, nil
}
