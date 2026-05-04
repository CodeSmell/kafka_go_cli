package file

import (
	"kafka_go_cli/internal/model"
	"strings"
)

const keyDelimiter = "--key"
const headerDelimiter = "--header"

// ParseMessage converts raw file contents into a structured model.Message.
//
// Parser is intentionally a stub for now: it returns body-only.
// The file format (keys/headers/body delimiters) can be implemented later
// without changing the poller/processor wiring.
func ParseMessage(content string) model.Message {
	message := model.Message{}
	remainingContent := splitOutKey(content, &message)
	message.Body = remainingContent

	theBody := splitOutHeaders(remainingContent, &message)

	message.Body = strings.TrimSpace(theBody)
	return message
}

// extracts the value of the message key
// as one value from start to the first occurrence of the key delimiter
func splitOutKey(content string, msg *model.Message) string {
	// split content based on delimiters
	// into key and the rest of the content
	parts := strings.SplitN(content, keyDelimiter, 2)
	if len(parts) < 2 {
		return content
	}

	keyPart := strings.TrimSpace(parts[0])
	// add the key to the message struct
	if keyPart != "" {
		msg.Keys = append(msg.Keys, model.KeyValue{Key: "key", Value: keyPart})
	}
	// return the remaining content after the key part
	return parts[1]
}

// takes the message content
// splits out the header section based on the header delimiter
// assumes the key is already handled and removed from the content
func splitOutHeaders(content string, msg *model.Message) string {
	// split content based on delimiters
	// into headers and the rest of the content
	parts := strings.SplitN(content, headerDelimiter, 2)
	if len(parts) < 2 {
		return content
	}

	// build headers from the header part
	msg.Headers = buildHeaders(parts[0])
	// return the remaining content after the header part
	return parts[1]
}

func buildHeaders(headerPart string) []model.KeyValue {
	// each line of the headers content
	// is expected to be in the format "HeaderKey: HeaderValue"
	headerLines := strings.Split(headerPart, "\n")
	headers := make([]model.KeyValue, 0, len(headerLines))
	for _, line := range headerLines {
		line = strings.TrimSpace(line)
		// skip empty lines
		if line == "" {
			continue
		}
		// skip lines with no colon (invalid header format)
		kv := strings.SplitN(line, ":", 2)
		if len(kv) != 2 {
			continue
		}
		// add the header to the message struct
		headers = append(headers, model.KeyValue{
			Key:   strings.TrimSpace(kv[0]),
			Value: strings.TrimSpace(kv[1]),
		})
	}
	// if no valid headers were found,
	// return nil instead of an empty slice
	if len(headers) == 0 {
		return nil
	}
	return headers
}
