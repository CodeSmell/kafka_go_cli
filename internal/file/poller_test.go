package file

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDirectoryPollerBuilder(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	processor := &testProcessor{}

	poller, err := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithDeleteFiles(true).
		WithKeepRunning(true).
		WithProcessor(processor).
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, poller)

	assert.True(t, poller.keepRunning)
	assert.True(t, poller.deleteFiles)
	assert.Equal(t, tmpDir, poller.dirPath)
	assert.Equal(t, processor, poller.processor)
}

func TestDirectoryPollerBuilderDefaults(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	// Create poller with minimal config - should use defaults
	poller, err := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithProcessor(&testProcessor{}).
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, poller)

	// Verify the builder applied expected defaults
	assert.False(t, poller.keepRunning)
	assert.False(t, poller.deleteFiles)
	assert.Equal(t, time.Second, poller.pollIntervalMillis)
	assert.Equal(t, -1, poller.maxPollCycles)
	assert.Equal(t, 1, poller.workerCount)
}

func TestDirectoryPollerBuilderWithoutProcessor(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	// Build without calling WithProcessor() - processor will be nil
	poller, err := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		Build()

	assert.Error(t, err)
	assert.Nil(t, poller)
}

func TestDirectoryPollerBuilderWithoutMessageLocation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// Build without calling WithMessageLocation() - message location will be empty
	poller, err := NewDirectoryPollerBuilder(logger).
		WithProcessor(&testProcessor{}).
		Build()

	assert.Error(t, err)
	assert.Nil(t, poller)
}

// testProcessor is a test implementation of FileProcessor that does nothing.
type testProcessor struct{}

// Process implements the FileProcessor interface for testing.
func (t *testProcessor) Process(ctx context.Context, content string) error {
	return nil
}
