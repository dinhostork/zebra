package handlers

import "github.com/IBM/sarama"

type MessageHandler interface {
	HandleMessage(msg *sarama.ConsumerMessage) *sarama.ConsumerMessage
}

