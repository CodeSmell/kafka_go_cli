// Package processor defines the Processor interface and factory.
// It coordinates implementations in the processors/ subpackage.
package processor

import "context"

// FileProcessor handles the content of each file found during polling.
// Implementations can publish to Kafka, write to logs, or perform other actions.
type FileProcessor interface {
	// Process handles file content and returns an error if processing fails.
	Process(ctx context.Context, content string) error
}

// Processor extends FileProcessor with lifecycle management for graceful shutdown.
type Processor interface {
	FileProcessor
	// Handles shutdown logic, such as flushing buffers or closing connections.
	Close() error
}
