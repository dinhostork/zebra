package handleupload

import (
	"io/ioutil"
	"net/http"
	"zebra/shared"

	"github.com/IBM/sarama"
)

// HandleUpload handles the video upload request
func HandleUpload(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("video")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	tempFile, err := ioutil.TempFile("", "video-*.mp4")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tempFile.Write(fileBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	producer, err := shared.InitKafkaProducer(shared.VIDEO_UPLOAD_TOPIC)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer producer.Close()

	// Send the video file path to the video upload topic
	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: shared.VIDEO_UPLOAD_TOPIC,
		Value: sarama.StringEncoder(tempFile.Name()),
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
