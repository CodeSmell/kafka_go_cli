package kafka

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"kafka_go_cli/internal/model"
	"kafka_go_cli/internal/processor"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepKafkaProducerForTest(t *testing.T) {
	// Save the original newKafkaProducer function
	// and restore it after the test
	originalMethod := newKafkaProducer
	t.Cleanup(func() {
		newKafkaProducer = originalMethod
	})
}

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestKafkaConfigParamsRegistered(t *testing.T) {
	// relying on Go initialization
	// when we import processors the init function in processor_kafka.go
	// should have registered the config params with the factory
	params := processor.GetConfigParams("kafka")
	require.NotEmpty(t, params)

	// making a map for easier assertions
	flags := make([]string, 0, len(params))
	for _, p := range params {
		flags = append(flags, p.Flag)
	}

	expectedFlags := []string{
		"kafka-topic",
		"kafka-brokers",
		"kafka-acks",
		"kafka-retries",
		"kafka-retry-delay",
		"kafka-max-in-flight",
		"kafka-batch-size",
		"kafka-batch-delay",
		"kafka-producer-id",
	}

	assert.Len(t, flags, len(expectedFlags))
	assert.ElementsMatch(t, expectedFlags, flags)
}

func TestLoadKafkaConfigDefaults(t *testing.T) {
	cfg, err := LoadKafkaConfig(nil)
	require.NoError(t, err)

	assert.Equal(t, "localhost:9092", cfg.Brokers)
	assert.Equal(t, "default-topic", cfg.Topic)
	assert.Equal(t, "", cfg.ProducerID)
	assert.Equal(t, "all", cfg.Acks)
	assert.Equal(t, 3, cfg.Retries)
	assert.Equal(t, 100, cfg.RetryDelay)
	assert.Equal(t, 5, cfg.MaxInflight)
	assert.Equal(t, 16384, cfg.BatchSizeBytes)
	assert.Equal(t, 10, cfg.BatchDelay)
}

func TestLoadKafkaConfigOverrides(t *testing.T) {
	raw := map[string]interface{}{
		"brokers":       "kafka1:9092,kafka2:9092",
		"topic":         "orders",
		"producer-id":   "cli-1",
		"acks":          "1",
		"retries":       7,
		"retry-delay":   float64(250),
		"max-in-flight": 9,
		"batch-size":    float64(4096),
		"batch-delay":   25,
	}

	cfg, err := LoadKafkaConfig(raw)
	require.NoError(t, err)

	assert.Equal(t, "kafka1:9092,kafka2:9092", cfg.Brokers)
	assert.Equal(t, "orders", cfg.Topic)
	assert.Equal(t, "cli-1", cfg.ProducerID)
	assert.Equal(t, "1", cfg.Acks)
	assert.Equal(t, 7, cfg.Retries)
	assert.Equal(t, 250, cfg.RetryDelay)
	assert.Equal(t, 9, cfg.MaxInflight)
	assert.Equal(t, 4096, cfg.BatchSizeBytes)
	assert.Equal(t, 25, cfg.BatchDelay)
}

func TestNewUsesNewKafkaProducer(t *testing.T) {
	prepKafkaProducerForTest(t)
	newKafkaProducer = func(cfg KafkaConfig, logger *slog.Logger) (*ckafka.Producer, error) {
		// we aren't really making a producer
		return nil, nil
	}

	mySettings := model.Settings{
		ProcessorConfig: map[string]interface{}{
			"brokers": "kafka:9092",
			"topic":   "foo",
		},
	}

	proc, err := New(context.Background(), createTestLogger(), mySettings)
	require.NoError(t, err)
	assert.Nil(t, proc.producer)
	assert.Equal(t, "kafka:9092", proc.config.Brokers)
	assert.Equal(t, "foo", proc.config.Topic)
}

func TestNewReturnsNewKafkaProducerError(t *testing.T) {
	prepKafkaProducerForTest(t)
	newKafkaProducer = func(cfg KafkaConfig, logger *slog.Logger) (*ckafka.Producer, error) {
		// simulate an error connecting to Kafka
		return nil, errors.New("boom")
	}

	_, err := New(context.Background(), createTestLogger(), model.Settings{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to setup Kafka producer")
}

func TestConvertToKafkaMessage(t *testing.T) {
	// create a simple KafkaProcessor to access the conversion method
	p := &KafkaProcessor{config: KafkaConfig{Topic: "foo"}}

	myFileMessage := model.Message{
		Keys:    []model.KeyValue{{Key: "id", Value: "k-1"}},
		Headers: []model.KeyValue{{Key: "header1", Value: "h1"}, {Key: "header2", Value: "h2"}},
		Body:    "hello",
	}

	kafkaMessage := p.convertToKafkaMessage(myFileMessage)

	require.NotNil(t, kafkaMessage)
	assert.Equal(t, "foo", *kafkaMessage.TopicPartition.Topic)
	assert.Equal(t, "k-1", string(kafkaMessage.Key))
	assert.Equal(t, "h1", string(kafkaMessage.Headers[0].Value))
	assert.Equal(t, "h2", string(kafkaMessage.Headers[1].Value))
	assert.Equal(t, "hello", string(kafkaMessage.Value))
	assert.Equal(t, int32(ckafka.PartitionAny), kafkaMessage.TopicPartition.Partition)
}

func TestConvertToKafkaMessageNoKey(t *testing.T) {
	p := &KafkaProcessor{config: KafkaConfig{Topic: "orders"}}
	kafkaMessage := p.convertToKafkaMessage(model.Message{Body: "hello"})

	require.NotNil(t, kafkaMessage)
	assert.Equal(t, "", string(kafkaMessage.Key))
	assert.Nil(t, kafkaMessage.Headers)
	assert.Equal(t, "hello", string(kafkaMessage.Value))
}
