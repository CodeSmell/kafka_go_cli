package kafka

import (
	"context"
	"fmt"
	"kafka_go_cli/internal/model"
	"kafka_go_cli/internal/processor"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// KafkaProcessor publishes messages to Kafka brokers
type KafkaProcessor struct {
	logger   *slog.Logger
	producer *kafka.Producer
	config   KafkaConfig
}

type KafkaConfig struct {
	Brokers        string // "localhost:9092,localhost:9093"
	Topic          string // Kafka topic to publish to
	AddressFamily  string // Optional: broker address family (for example v4)
	ProducerID     string // Optional: producer.id for batching
	FlushInterval  int    // ms before auto-flush (0 = manual only)
	Acks           string // 0, 1 or  "all" for strongest durability
	Retries        int    // Number of retries for failed sends
	RetryDelay     int    // ms between retries
	MaxInflight    int    // Max messages in-flight (unacknowledged)
	BatchSizeBytes int    // Max batch size in bytes
	BatchDelay     int    // ms to wait before sending batch
}

// newKafkaProducer is a seam or hook that
// defaults to setupKafkaProducer method allowing
// it to be replaced in unit tests.
var newKafkaProducer = setupKafkaProducer

// the init function is part of Go and is called when the package is imported
// we are using that to register with the factory
func init() {
	processor.Register("kafka", func(ctx context.Context, logger *slog.Logger, settings model.Settings) (processor.Processor, error) {
		kafkaProc, err := New(ctx, logger, settings)
		if err != nil {
			logger.Error("failed to create Kafka processor", "error", err)
			return nil, err
		}
		return kafkaProc, nil
	})

	// Register Kafka processor's configuration parameters
	processor.RegisterConfigParams("kafka", []model.ConfigParam{
		{Name: "topic", Flag: "kafka-topic", Description: "Kafka topic to publish to"},
		{Name: "brokers", Flag: "kafka-brokers", Description: "Kafka brokers (comma-separated)"},
		{Name: "broker-address-family", Flag: "kafka-broker-address-family", Description: "Kafka broker address family (for example v4)"},
		{Name: "acks", Flag: "kafka-acks", Description: "Kafka acks setting (0, 1, or all)"},
		{Name: "retries", Flag: "kafka-retries", Description: "Number of retries for failed sends"},
		{Name: "retry-delay", Flag: "kafka-retry-delay", Description: "Milliseconds between retries"},
		{Name: "max-in-flight", Flag: "kafka-max-in-flight", Description: "Max messages in-flight"},
		{Name: "batch-size", Flag: "kafka-batch-size", Description: "Max batch size in bytes"},
		{Name: "batch-delay", Flag: "kafka-batch-delay", Description: "Milliseconds to wait before sending batch"},
		{Name: "producer-id", Flag: "kafka-producer-id", Description: "Producer ID for batching"},
	})
}

// New creates a new Kafka processor with a connected producer
func New(ctx context.Context, logger *slog.Logger, settings model.Settings) (*KafkaProcessor, error) {
	cfg, err := LoadKafkaConfig(settings.ProcessorConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load Kafka config: %w", err)
	}

	producer, err := newKafkaProducer(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to setup Kafka producer: %w", err)
	}

	logger.Info("Kafka processor initialized", "brokers", cfg.Brokers, "topic", cfg.Topic)

	return &KafkaProcessor{
		logger:   logger,
		producer: producer,
		config:   cfg,
	}, nil
}

// Process implements the Processor interface to publish messages to Kafka
func (k *KafkaProcessor) Process(ctx context.Context, message model.Message) error {

	record := k.convertToKafkaMessage(message)

	// Send asynchronously — deliver reports go to DeliveryReports channel
	deliveryReports := make(chan kafka.Event, 1)
	err := k.producer.Produce(record, deliveryReports)
	if err != nil {
		k.logger.Error("failed to produce message", "error", err)
		return err
	}

	// Wait for delivery report (blocking until message is delivered or fails)
	event := <-deliveryReports
	m := event.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		k.logger.Error("failed to deliver message",
			"topic", *m.TopicPartition.Topic,
			"error", m.TopicPartition.Error)
		return m.TopicPartition.Error
	}

	k.logger.Info("message published successfully",
		"topic", *m.TopicPartition.Topic,
		"partition", m.TopicPartition.Partition,
		"offset", m.TopicPartition.Offset)

	return nil
}

func (k *KafkaProcessor) convertToKafkaMessage(message model.Message) *kafka.Message {
	// Extract key from message (first key if available)
	var key string
	if len(message.Keys) > 0 {
		key = message.Keys[0].Value
	}

	var headers []kafka.Header
	if len(message.Headers) > 0 {
		headers = make([]kafka.Header, 0, len(message.Headers))
		for _, h := range message.Headers {
			headers = append(headers, kafka.Header{Key: h.Key, Value: []byte(h.Value)})
		}
	}

	record := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &k.config.Topic,
			Partition: kafka.PartitionAny,
		},
		Key:     []byte(key),
		Headers: headers,
		Value:   []byte(message.Body),
	}

	return record
}

// Close flushes pending messages and closes the producer
func (k *KafkaProcessor) Close() error {
	if k.producer != nil {
		k.logger.Info("closing Kafka producer...")
		remaining := k.producer.Flush(30000) // 30 second timeout
		if remaining > 0 {
			k.logger.Warn("messages remained in queue after flush", "count", remaining)
		}
		k.producer.Close()
	}
	return nil
}

// setupKafkaProducer creates and configures a Kafka producer
func setupKafkaProducer(cfg KafkaConfig, logger *slog.Logger) (*kafka.Producer, error) {
	configMap := &kafka.ConfigMap{
		"bootstrap.servers":                     cfg.Brokers,
		"client.id":                             cfg.ProducerID,
		"acks":                                  cfg.Acks,
		"retries":                               cfg.Retries,
		"retry.backoff.ms":                      cfg.RetryDelay,
		"max.in.flight.requests.per.connection": cfg.MaxInflight,
		"batch.size":                            cfg.BatchSizeBytes,
		"linger.ms":                             cfg.BatchDelay,
	}
	if cfg.AddressFamily != "" {
		_ = configMap.SetKey("broker.address.family", cfg.AddressFamily)
	}

	logger.Debug("creating Kafka producer", "brokers", cfg.Brokers)

	p, err := kafka.NewProducer(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	return p, nil
}

// Helper functions to extract values from raw config map
func getStringOrDefault(cfg map[string]interface{}, key string, defaultValue string) string {
	if val, ok := cfg[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func getIntOrDefault(cfg map[string]interface{}, key string, defaultValue int) int {
	if val, ok := cfg[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

// LoadKafkaConfig extracts Kafka configuration from the raw processor config map
func LoadKafkaConfig(rawConfig map[string]interface{}) (KafkaConfig, error) {
	if rawConfig == nil {
		rawConfig = make(map[string]interface{})
	}

	return KafkaConfig{
		Brokers:        getStringOrDefault(rawConfig, "brokers", "localhost:9092"),
		AddressFamily:  getStringOrDefault(rawConfig, "broker.address.family", getStringOrDefault(rawConfig, "broker-address-family", "")),
		Topic:          getStringOrDefault(rawConfig, "topic", "default-topic"),
		ProducerID:     getStringOrDefault(rawConfig, "producer-id", ""),
		Acks:           getStringOrDefault(rawConfig, "acks", "all"),
		Retries:        getIntOrDefault(rawConfig, "retries", 3),
		RetryDelay:     getIntOrDefault(rawConfig, "retry-delay", 100),
		MaxInflight:    getIntOrDefault(rawConfig, "max-in-flight", 5),
		BatchSizeBytes: getIntOrDefault(rawConfig, "batch-size", 16384),
		BatchDelay:     getIntOrDefault(rawConfig, "batch-delay", 10),
	}, nil
}
