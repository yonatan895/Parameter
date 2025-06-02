package main

import (
	"os"
	"testing"
)

func TestNewRedisClient(t *testing.T) {
	os.Setenv("REDIS_ADDR", "127.0.0.1:9999")
	c := newRedisClient()
	if c == nil {
		t.Fatal("client nil")
	}
	if c.Options().Addr != "127.0.0.1:9999" {
		t.Fatalf("expected addr 127.0.0.1:9999 got %s", c.Options().Addr)
	}
	_ = c.Close()
}

func TestNewKafkaWriter(t *testing.T) {
	os.Setenv("KAFKA_ADDR", "127.0.0.1:9093")
	w := newKafkaWriter()
	if w == nil {
		t.Fatal("writer nil")
	}
	if len(w.Stats().Brokers) == 0 || w.Stats().Brokers[0].BrokerAddress != "127.0.0.1:9093" {
		t.Fatalf("unexpected broker address")
	}
	_ = w.Close()
}
