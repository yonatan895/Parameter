package kafka

import "context"

type WriterConfig struct {
	Brokers []string
	Topic   string
}

type Message struct {
	Value []byte
}

type Broker struct {
	BrokerAddress string
}

type WriterStats struct {
	Brokers []Broker
}

type Writer struct {
	brokers []string
}

func NewWriter(cfg WriterConfig) *Writer {
	return &Writer{brokers: cfg.Brokers}
}

func (w *Writer) WriteMessages(ctx context.Context, msgs ...Message) error { return nil }

func (w *Writer) Close() error { return nil }

func (w *Writer) Stats() WriterStats {
	stats := WriterStats{}
	for _, b := range w.brokers {
		stats.Brokers = append(stats.Brokers, Broker{BrokerAddress: b})
	}
	return stats
}
