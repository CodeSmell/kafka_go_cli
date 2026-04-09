package file

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Note: Any test that calls PollDirectory should set a MaxPollCycles limit
// to ensure the test doesn't run indefinitely.
func TestDirectoryPollerBuilder(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	processor := &testProcessor{}

	poller, err := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithDeleteFiles(true).
		WithKeepRunning(true).
		WithMaxPollCycles(5).
		WithPollInterval(500 * time.Millisecond).
		WithNoOpProcessor(true).
		WithProcessor(processor).
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, poller)

	assert.True(t, poller.keepRunning)
	assert.True(t, poller.deleteFiles)
	assert.Equal(t, tmpDir, poller.dirPath)
	assert.Equal(t, 5, poller.maxPollCycles)
	assert.Equal(t, 500*time.Millisecond, poller.pollIntervalMillis)
	assert.True(t, poller.noOp)
	assert.Equal(t, processor, poller.processor)
}

func TestDirectoryPollerBuilderDefaults(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	// Create poller with minimal config - should use Go zero values (not config defaults)
	poller, err := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithProcessor(&testProcessor{}).
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, poller)

	// Verify the builder uses Go zero values (defaults come from config layer)
	assert.False(t, poller.keepRunning)
	assert.False(t, poller.deleteFiles)
	assert.Equal(t, time.Duration(0), poller.pollIntervalMillis)
	assert.Equal(t, 0, poller.maxPollCycles)
	assert.Equal(t, 0, poller.workerCount)
	assert.False(t, poller.noOp)
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

func TestPollDirectoryWithMaxCycles(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	// Create some test files
	for i := 1; i <= 3; i++ {
		file, _ := os.Create(tmpDir + "/file" + string(rune(i)))
		file.Close()
	}

	poller, _ := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithKeepRunning(true).
		WithMaxPollCycles(2).
		WithDeleteFiles(false).
		WithPollInterval(10 * time.Millisecond).
		WithProcessor(&testProcessor{}).
		Build()

	count, err := poller.PollDirectory(context.Background())

	assert.NoError(t, err)
	// The poller should find files on both cycles (3 files * 2 cycles)
	assert.Equal(t, 6, count)
}

func TestPollDirectoryInvalidDirectory(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	poller, _ := NewDirectoryPollerBuilder(logger).
		WithMessageLocation("/nonexistent/path").
		WithMaxPollCycles(1).
		WithProcessor(&testProcessor{}).
		Build()

	ctx := context.Background()
	_, err := poller.PollDirectory(ctx)

	assert.Error(t, err)
}

func TestPollDirectoryContextCancellation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	file, _ := os.Create(tmpDir + "/foo")
	file.Close()

	processor := &testProcessor{}
	poller, _ := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithMaxPollCycles(100). // Large cycle count, but will be interrupted
		WithPollInterval(100 * time.Millisecond).
		WithKeepRunning(true).
		WithProcessor(processor).
		Build()

	// Create a context that cancels after 50ms
	// we set up polling to run much longer
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// This should return quickly due to context cancellation,
	// not after 100 cycles
	count, err := poller.PollDirectory(ctx)

	assert.NoError(t, err)
	// Should have completed at least 1 cycle but not all 100
	assert.Greater(t, count, 0)
	assert.Less(t, count, 100)
}

func TestPollDirectoryEmptyDirectory(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	processor := &testProcessor{}
	poller, _ := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithKeepRunning(true).
		WithMaxPollCycles(1).
		WithProcessor(processor).
		Build()

	count, err := poller.PollDirectory(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestPollDirectoryIgnoresSubdirectories(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	// Create 2 files
	os.Create(tmpDir + "/file1.txt")
	os.Create(tmpDir + "/file2.txt")
	// Create 1 subdirectory
	os.Mkdir(tmpDir+"/subdir", 0755)
	os.Create(tmpDir + "/subdir/file3.txt")

	processor := &testProcessor{}
	poller, _ := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithMaxPollCycles(1).
		WithProcessor(processor).
		Build()

	count, err := poller.PollDirectory(context.Background())

	assert.NoError(t, err)
	// Should only count the 2 files in root, not the subdir or file inside it
	assert.Equal(t, 2, count)
}

// testProcessor is a test implementation of FileProcessor that does nothing.
type testProcessor struct{}

// Process implements the FileProcessor interface for testing.
func (t *testProcessor) Process(ctx context.Context, content string) error {
	return nil
}
