package testutil

// can be used by tests in other packages
// except from the processor package to avoid circular imports
import (
	"context"
	"kafka_go_cli/internal/model"
)

// MockProcessor is a mock FileProcessor for testing that tracks calls and captures content.
type MockProcessor struct {
	// Called tracks whether Process was called
	Called bool
	// CapturedContent stores all content passed to Process in order
	CapturedContent []model.Message
	// CallCount tracks how many times Process was called
	CallCount int
	// ProcessFunc is an optional custom function to execute on Process calls
	ProcessFunc func(ctx context.Context, content model.Message) error
	// CloseCount tracks how many times Close was called
	CloseCount int
}

func init() {

}

// Process implements the FileProcessor interface for the mock.
func (m *MockProcessor) Process(ctx context.Context, content model.Message) error {
	m.Called = true
	m.CallCount++
	m.CapturedContent = append(m.CapturedContent, content)

	if m.ProcessFunc != nil {
		return m.ProcessFunc(ctx, content)
	}
	return nil
}

func (m *MockProcessor) Close() error {
	// No resources to clean up in this mock, but method is required by Processor interface
	m.CloseCount++
	return nil
}
