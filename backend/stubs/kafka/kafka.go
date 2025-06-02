package kafka

import "context"

type WriterConfig struct {
	Brokers []string
	Topic   string
}

type Message struct {
	Value []byte
}

type BrokerStats struct {
	BrokerAddress string
}

type Stats struct {
	Brokers []BrokerStats
}

type Writer struct {
	cfg WriterConfig
}

func NewWriter(cfg WriterConfig) *Writer {
	return &Writer{cfg: cfg}
}

func (w *Writer) WriteMessages(ctx context.Context, msgs ...Message) error { return nil }

func (w *Writer) Stats() Stats {
	var stats Stats
	for _, addr := range w.cfg.Brokers {
		stats.Brokers = append(stats.Brokers, BrokerStats{BrokerAddress: addr})
	}
	return stats
}

func (w *Writer) Close() error { return nil }
