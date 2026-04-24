package noop

import (
	"context"
	"log/slog"

	"kafka_go_cli/internal/model"
	"kafka_go_cli/internal/processor"
)

// NoOpProcessor implements the Processor interface but does nothing.
// Useful for testing and as a default processor.
type NoOpProcessor struct {
	logger *slog.Logger
}

// the init function is part of Go and is called when the package is imported
// we are using that to register with the factory
func init() {
	// Register the NoOp constructor with the factory
	processor.Register("noop", func(ctx context.Context, logger *slog.Logger, settings model.Settings) (processor.Processor, error) {
		return New(logger), nil
	})
	// NoOp processor has no configuration parameters to register
}

// New is a constructor that creates a new no-op processor.
func New(logger *slog.Logger) *NoOpProcessor {
	return &NoOpProcessor{logger: logger}
}

// Process is a no-op implementation of the Process method.
func (n *NoOpProcessor) Process(ctx context.Context, message model.Message) error {
	n.logger.Debug("no-op is processing a file")
	return nil
}

// Close is a no-op implementation of the Close method.
func (n *NoOpProcessor) Close() error {
	return nil
}
