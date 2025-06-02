package main

import (
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

func newRedisClient() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	return redis.NewClient(&redis.Options{Addr: addr})
}

func newKafkaWriter() *kafka.Writer {
	addr := os.Getenv("KAFKA_ADDR")
	if addr == "" {
		addr = "localhost:9092"
	}
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{addr},
		Topic:   "events",
	})
}
