package file

import "kafka_go_cli/internal/model"

// ParseMessage converts raw file contents into a structured model.Message.
//
// Parser is intentionally a stub for now: it returns body-only.
// The file format (keys/headers/body delimiters) can be implemented later
// without changing the poller/processor wiring.
func ParseMessage(content string) (model.Message, error) {
	return model.Message{Body: content}, nil
}
