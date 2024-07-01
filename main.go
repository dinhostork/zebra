package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	strategy := os.Getenv("STRATEGY")
	if strategy == "" {
		fmt.Println("STRATEGY not set in the environment file.")
		return
	}

	var consumer sarama.PartitionConsumer
	if strategy == "kafka" || strategy == "both" {
		consumer, err = initKafkaConsumer()
		if err != nil {
			panic(err)
		}
	}

	if strategy == "api" || strategy == "both" {
		go startAPIServer()
	}

	if strategy == "kafka" || strategy == "both" {
		go startKafkaConsumer(consumer)
	}

	select {}
}

func initKafkaConsumer() (sarama.PartitionConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	brokers := []string{"localhost:9092"}
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	consumer, err := master.ConsumePartition("zebra-watermark", 0, sarama.OffsetNewest)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

func startAPIServer() {
	fmt.Println("Starting API server on :8080")
	http.HandleFunc("/upload", handleUpload)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
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

	// TODO - SEND MESSAGE TO KAFKA

	fmt.Fprint(w, "Video processed successfully!")
}

func startKafkaConsumer(_ sarama.PartitionConsumer) {
	fmt.Println("Starting Kafka consumer")
	// for {
	// 	select {
	// 	case message := <-consumer.Messages():

	// 		// TODO - PROCESS VIDEO
	// 	case err := <-consumer.Errors():
	// 		fmt.Println(err.Error())
	// 	}
	// }
}
