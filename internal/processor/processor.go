// Package processor defines the Processor interface and factory.
// It coordinates implementations in the processors/ subpackage.
package processor

import "context"

// KeyValue represents a single key/value pair.
// Used for both message keys and headers.
type KeyValue struct {
	Key   string
	Value string
}

// Message is the structured representation of a polled file.
// Multiple keys and headers (including duplicates) are allowed.
type Message struct {
	Keys    []KeyValue
	Headers []KeyValue
	Body    string
}

// FileProcessor handles the content of each file found during polling.
// Implementations can publish to Kafka, write to logs, or perform other actions.
type FileProcessor interface {
	// Process handles file content and returns an error if processing fails.
	Process(ctx context.Context, message Message) error
}

// Processor extends FileProcessor with lifecycle management for graceful shutdown.
type Processor interface {
	FileProcessor
	// Handles shutdown logic, such as flushing buffers or closing connections.
	Close() error
}
