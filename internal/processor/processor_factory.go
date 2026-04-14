package processor

import (
	"context"
	"fmt"
	"log/slog"

	"kafka_go_cli/internal/config"
	"kafka_go_cli/internal/processors/noop"
)

// The factory NewProcessor method creates a processor based on the configured type.
func NewProcessor(ctx context.Context, logger *slog.Logger, settings config.Settings) (Processor, error) {
	logger.Info("creating processor", "type", settings.ProcessorType)

	switch settings.ProcessorType {
	case "kafka":
		return nil, fmt.Errorf("kafka processor not yet implemented")

	case "pulsar":
		return nil, fmt.Errorf("pulsar processor not yet implemented")

	case "noop":
		return noop.New(logger), nil

	default:
		return nil, fmt.Errorf("unknown processor type: %s", settings.ProcessorType)
	}
}
