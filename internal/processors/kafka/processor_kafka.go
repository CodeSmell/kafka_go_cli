package kafka

import (
	"context"
	"kafka_go_cli/internal/config"
	"kafka_go_cli/internal/processor"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// KafkaProcessor is a stub Processor implementation.
// Kafka publishing behavior can be implemented later without changing the wiring.
type KafkaProcessor struct {
	logger *slog.Logger
}

// the init function is part of Go and is called when the package is imported
// we are using that to register with the factory
func init() {
	processor.Register("kafka", func(ctx context.Context, logger *slog.Logger, settings config.Settings) (processor.Processor, error) {
		_ = ctx
		_ = settings
		return New(logger), nil
	})
}

type KafkaConfig struct {
	Brokers        string // "localhost:9092,localhost:9093"
	Topic          string // Kafka topic to publish to
	ProducerID     string // Optional: producer.id for batching
	FlushInterval  int    // ms before auto-flush (0 = manual only)
	Acks           string // 0, 1 or  "all" for strongest durability
	Retries        int    // Number of retries for failed sends
	RetryDelay     int    // ms between retries
	MaxInflight    int    // Max messages in-flight (unacknowledged)
	BatchSizeBytes int    // Max batch size in bytes
	BatchDelay     int    // ms to wait before sending batch
}

// New is a constructor that creates a new Kafka processor.
func New(logger *slog.Logger) *KafkaProcessor {
	return &KafkaProcessor{logger: logger}
}

// Implementation of FileProcessor that can publish to Kafka
func (k *KafkaProcessor) Process(ctx context.Context, message processor.Message) error {
	return nil
}

// Close is a no-op implementation of the Close method.
func (k *KafkaProcessor) Close() error {
	return nil
}

func LoadKafkaConfig(configPath string) (KafkaConfig, error) {
	return KafkaConfig{}, nil
}

func setupKafkaProducer() (*kafka.Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{})
	if err != nil {
		return nil, err
	}
	return p, nil
}
