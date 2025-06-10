package kafka

import (
	"context"
	"log"
	"sync"

	"github.com/IBM/sarama"
)

type Consumer struct {
	ready    chan bool
	config   *sarama.Config
	consumer sarama.ConsumerGroup
}

func NewConsumer(brokers []string, groupId string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_5_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRange}
	config.Consumer.Offsets.AutoCommit.Enable = false

	consumer, err := sarama.NewConsumerGroup(brokers, groupId, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		ready:    make(chan bool),
		config:   config,
		consumer: consumer,
	}, nil
}

func (c *Consumer) Consume(ctx context.Context, topics []string, handler sarama.ConsumerGroupHandler) error {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := c.consumer.Consume(ctx, topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			c.ready = make(chan bool)
		}
	}()

	<-c.ready
	<-ctx.Done()
	wg.Wait()
	return c.consumer.Close()
}
