//go:build integration
// +build integration

package kafka

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"kafka_go_cli/internal/model"

	// confluent-kafka-go client for consuming messages in the test
	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// redpanda testcontainers module for running a Kafka-compatible container in the test
	tcredpanda "github.com/testcontainers/testcontainers-go/modules/redpanda"
)

const defaultKafkaTestImage = "redpandadata/redpanda:v23.3.3"

func kafkaTestImage() string {
	if v := strings.TrimSpace(os.Getenv("KAFKA_IT_IMAGE")); v != "" {
		return v
	}
	return defaultKafkaTestImage
}

func TestIntegration_KafkaProcessorWithContainer(t *testing.T) {
	ctx := context.Background()
	container, err := tcredpanda.Run(ctx, kafkaTestImage(), tcredpanda.WithAutoCreateTopics())
	if err != nil {
		t.Skipf("skipping integration test: redpanda container unavailable: %v", err)
	}
	defer func() {
		_ = container.Terminate(ctx)
	}()

	seedBroker, err := container.KafkaSeedBroker(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, seedBroker)

	// we need to set the address family to v4 in the test because
	// on some platforms (like macOS with Colima) localhost
	// may resolve to ::1 which can cause connection issues for
	// Kafka clients that don't handle IPv6 well
	topic := fmt.Sprintf("it-topic-%d", time.Now().UnixNano())
	settings := model.Settings{
		ProcessorConfig: map[string]interface{}{
			"brokers":               seedBroker,
			"topic":                 topic,
			"acks":                  "all",
			"broker-address-family": "v4",
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	kafkaProcessor, err := New(ctx, logger, settings)
	require.NoError(t, err)
	defer func() {
		_ = kafkaProcessor.Close()
	}()

	expectedKey := "order-123"
	expectedValue := "hello from integration test"

	msg := model.Message{
		Keys: []model.KeyValue{{Key: "id", Value: expectedKey}},
		Body: expectedValue,
	}

	// processor should publish the message without error
	err = kafkaProcessor.Process(ctx, msg)
	require.NoError(t, err)

	// consume the message using a real Kafka consumer to verify it was published correctly
	consumer, err := ckafka.NewConsumer(&ckafka.ConfigMap{
		"bootstrap.servers":     seedBroker,
		"group.id":              fmt.Sprintf("it-group-%d", time.Now().UnixNano()),
		"auto.offset.reset":     "earliest",
		"broker.address.family": "v4",
	})
	require.NoError(t, err)
	defer consumer.Close()

	err = consumer.Subscribe(topic, nil)
	require.NoError(t, err)

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		ev := consumer.Poll(1_000)
		if ev == nil {
			continue
		}
		switch msg := ev.(type) {
		case *ckafka.Message:
			assert.Equal(t, expectedValue, string(msg.Value))
			assert.Equal(t, expectedKey, string(msg.Key))
			return
		case ckafka.Error:
			t.Fatalf("kafka consumer error: %v", msg)
		}
	}

	t.Fatal("timed out waiting for consumed message")
}
