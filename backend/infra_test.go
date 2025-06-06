package main

import (
	"os"
	"testing"
)

// TestNewRedisClient verifies that newRedisClient respects the REDIS_ADDR
// environment variable.
func TestNewRedisClient(t *testing.T) {
	if err := os.Setenv("REDIS_ADDR", "127.0.0.1:9999"); err != nil {
		t.Fatal(err)
	}
	c := newRedisClient()
	if c == nil {
		t.Fatal("client nil")
	}
	if c.Options().Addr != "127.0.0.1:9999" {
		t.Fatalf("expected addr 127.0.0.1:9999 got %s", c.Options().Addr)
	}
	_ = c.Close()
}

// TestNewKafkaWriter checks that newKafkaWriter uses KAFKA_ADDR for configuration.
func TestNewKafkaWriter(t *testing.T) {
	if err := os.Setenv("KAFKA_ADDR", "127.0.0.1:9093"); err != nil {
		t.Fatal(err)
	}
	w := newKafkaWriter()
	if w == nil {
		t.Fatal("writer nil")
	}
	_ = w.Stats()
	_ = w.Close()
}
