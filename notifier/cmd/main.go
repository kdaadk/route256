package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"route256/notifier/config"
	"route256/notifier/internal/consumer"
	"route256/notifier/internal/kafka"
)

func main() {
	slog.Info("Starting notifier service...")

	cfg := config.Load()

	kafkaConsumer, err := kafka.NewConsumer([]string{cfg.KafkaBrokers}, cfg.GroupID)
	if err != nil {
		slog.Error("Error creating consumer: %v", err)
	}

	handler := consumer.NewHandler()

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			if err := kafkaConsumer.Consume(ctx, []string{cfg.KafkaTopic}, handler); err != nil {
				slog.Info("Error from consumer: %v", err)
				time.Sleep(100 * time.Millisecond)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-handler.Ready
	slog.Info("Consumer is ready")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
	cancel()
	wg.Wait()

	slog.Info("Shutting down notifier service")
}
