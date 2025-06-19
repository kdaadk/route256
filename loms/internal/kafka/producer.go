package kafka

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	myLogger "github.com/kdaadk/route256/pkg/logger"
	"net"
	"os"
	"time"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewKafkaProducer(topic string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Retry.Backoff = 1 * time.Second
	config.Net.DialTimeout = 10 * time.Second

	broker := getBrokerAddress()

	producer, err := sarama.NewSyncProducer([]string{broker}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w (broker: %s)", err, broker)
	}

	myLogger.InfoContext(context.Background(), fmt.Sprintf("Connected to Kafka broker at %v", broker))
	return &KafkaProducer{producer: producer, topic: topic}, nil
}

func (kp *KafkaProducer) SendMessage(key, value string) error {
	msg := &sarama.ProducerMessage{
		Topic:     kp.topic,
		Key:       sarama.StringEncoder(key),
		Value:     sarama.StringEncoder(value),
		Timestamp: time.Now(),
	}

	partition, offset, err := kp.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	myLogger.InfoContext(context.Background(), fmt.Sprintf("Produced message to %s[%d]@%d", kp.topic, partition, offset))
	return nil
}

func (kp *KafkaProducer) Close() error {
	if err := kp.producer.Close(); err != nil {
		return fmt.Errorf("failed to close producer: %w", err)
	}
	return nil
}

func getBrokerAddress() string {
	if broker := os.Getenv("KAFKA_BROKER"); broker != "" {
		return broker
	}

	if _, err := net.LookupHost("kafka0"); err == nil {
		return "kafka0:29092"
	}

	return "localhost:9092"
}
