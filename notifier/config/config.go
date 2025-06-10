package config

import (
	"os"
)

type Config struct {
	KafkaBrokers string
	KafkaTopic   string
	GroupID      string
}

func Load() *Config {
	return &Config{
		KafkaBrokers: getEnv("KAFKA_BROKERS", "kafka0:29092"),
		KafkaTopic:   getEnv("KAFKA_TOPIC", "loms.order-events"),
		GroupID:      getEnv("KAFKA_GROUP_ID", "notifier-group"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
