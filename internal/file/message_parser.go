package file

import "kafka_go_cli/internal/processor"

// ParseMessage converts raw file contents into a structured processor.Message.
//
// Parser is intentionally a stub for now: it returns body-only.
// The file format (keys/headers/body delimiters) can be implemented later
// without changing the poller/processor wiring.
func ParseMessage(content string) (processor.Message, error) {
	return processor.Message{Body: content}, nil
}
