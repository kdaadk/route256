package consumer

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	myLogger "github.com/kdaadk/route256/pkg/logger"
)

type Handler struct {
	Ready chan bool
}

func NewHandler() *Handler {
	return &Handler{
		Ready: make(chan bool),
	}
}

func (h *Handler) Setup(session sarama.ConsumerGroupSession) error {
	close(h.Ready)
	return nil
}

func (h *Handler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		myLogger.InfoContext(context.Background(), fmt.Sprintf("Message received: Topic=%s, Partition=%d, Offset=%d, Key=%s, Value=%s",
			message.Topic, message.Partition, message.Offset, string(message.Key), string(message.Value)))

		session.MarkMessage(message, "")
		session.Commit()
	}
	return nil
}
