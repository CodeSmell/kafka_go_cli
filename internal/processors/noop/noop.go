package noop

import (
	"context"
	"log/slog"
)

// NoOpProcessor implements the Processor interface but does nothing.
// Useful for testing and as a default processor.
type NoOpProcessor struct {
	logger *slog.Logger
}

// New creates a new no-op processor.
func New(logger *slog.Logger) *NoOpProcessor {
	return &NoOpProcessor{logger: logger}
}

// Process is a no-op implementation of the Process method.
func (n *NoOpProcessor) Process(ctx context.Context, content string) error {
	n.logger.Debug("no-op is processing a file")
	return nil
}

// Close is a no-op implementation of the Close method.
func (n *NoOpProcessor) Close() error {
	return nil
}
