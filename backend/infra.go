package main

import (
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

// newRedisClient creates and returns a Redis client using the REDIS_ADDR
// environment variable. If the variable is empty it falls back to
// localhost.
func newRedisClient() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	return redis.NewClient(&redis.Options{Addr: addr})
}

// newKafkaWriter creates a Kafka writer configured from the KAFKA_ADDR
// environment variable. It publishes messages to the "events" topic.
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
