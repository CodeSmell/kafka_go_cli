package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"kafka_go_cli/internal/processor"

	"github.com/stretchr/testify/assert"
)

// mockFileProcessor is a mock FileProcessor for testing that tracks calls and captures content.
type mockFileProcessor struct {
	// called tracks whether Process was called
	called bool
	// capturedContent stores all content passed to Process in order
	capturedContent []processor.Message
	// callCount tracks how many times Process was called
	callCount int
	// processFunc is an optional custom function to execute on Process calls
	processFunc func(ctx context.Context, content processor.Message) error
}

// Process implements the FileProcessor interface for the mock.
func (m *mockFileProcessor) Process(ctx context.Context, content processor.Message) error {
	m.called = true
	m.callCount++
	m.capturedContent = append(m.capturedContent, content)

	if m.processFunc != nil {
		return m.processFunc(ctx, content)
	}
	return nil
}

func createTempFile(t *testing.T, tmpDir string, fileName string, content string) {
	testFilePath := tmpDir + "/" + fileName
	testContent := content
	err := os.WriteFile(testFilePath, []byte(testContent), 0644)
	assert.NoError(t, err)
}

// Note: Any test that calls PollDirectory should set a MaxPollCycles limit
// to ensure the test doesn't run indefinitely.
func TestDirectoryPollerBuilder(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	processor := &mockFileProcessor{}

	poller, err := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithDeleteFiles(true).
		WithKeepRunning(true).
		WithMaxPollCycles(5).
		WithPollInterval(500 * time.Millisecond).
		WithProcessor(processor).
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, poller)

	assert.True(t, poller.keepRunning)
	assert.True(t, poller.deleteFiles)
	assert.Equal(t, tmpDir, poller.dirPath)
	assert.Equal(t, 5, poller.maxPollCycles)
	assert.Equal(t, 500*time.Millisecond, poller.pollIntervalMillis)
	assert.Equal(t, processor, poller.processor)
}

func TestDirectoryPollerBuilderDefaults(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	// Create poller with minimal config - should use Go zero values (not config defaults)
	poller, err := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithProcessor(&mockFileProcessor{}).
		Build()

	assert.NoError(t, err)
	assert.NotNil(t, poller)

	// Verify the builder uses Go zero values (defaults come from config layer)
	assert.False(t, poller.keepRunning)
	assert.False(t, poller.deleteFiles)
	assert.Equal(t, time.Duration(0), poller.pollIntervalMillis)
	assert.Equal(t, 0, poller.maxPollCycles)
	assert.Equal(t, 0, poller.workerCount)
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
		WithProcessor(&mockFileProcessor{}).
		Build()

	assert.Error(t, err)
	assert.Nil(t, poller)
}

func TestPollDirectoryWithMaxCyclesAndMultiFiles(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	// Create some test files
	for i := 1; i <= 3; i++ {
		createTempFile(t, tmpDir, "file"+fmt.Sprint(i)+".txt", "foo bar "+fmt.Sprint(i))
	}

	mockProcessor := &mockFileProcessor{}
	poller, _ := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithKeepRunning(true).
		WithMaxPollCycles(2).
		WithDeleteFiles(false).
		WithPollInterval(10 * time.Millisecond).
		WithProcessor(mockProcessor).
		Build()

	count, err := poller.PollDirectory(context.Background())

	assert.NoError(t, err)
	// The poller should find files on both cycles (3 files * 2 cycles)
	assert.Equal(t, 6, count)
	assert.Equal(t, 6, mockProcessor.callCount)
	assert.Equal(t, "foo bar 1", mockProcessor.capturedContent[0].Body)
	assert.Equal(t, "foo bar 2", mockProcessor.capturedContent[1].Body)
	assert.Equal(t, "foo bar 3", mockProcessor.capturedContent[2].Body)
	assert.FileExists(t, tmpDir+"/file1.txt")
	assert.FileExists(t, tmpDir+"/file2.txt")
	assert.FileExists(t, tmpDir+"/file3.txt")
}

func TestPollDirectoryInvalidDirectory(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	mockProcessor := &mockFileProcessor{}
	poller, _ := NewDirectoryPollerBuilder(logger).
		WithMessageLocation("/nonexistent/path").
		WithMaxPollCycles(1).
		WithProcessor(mockProcessor).
		Build()

	ctx := context.Background()
	_, err := poller.PollDirectory(ctx)

	assert.Error(t, err)
	assert.False(t, mockProcessor.called)
}

func TestPollDirectoryContextCancellation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	createTempFile(t, tmpDir, "testFile1.txt", "foo bar 1")

	mockProcessor := &mockFileProcessor{}
	poller, _ := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithMaxPollCycles(100). // Large cycle count, but will be interrupted
		WithPollInterval(100 * time.Millisecond).
		WithKeepRunning(true).
		WithProcessor(mockProcessor).
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
	assert.True(t, mockProcessor.called)
	assert.Equal(t, "foo bar 1", mockProcessor.capturedContent[0].Body)
}

func TestPollDirectoryEmptyDirectory(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	mockProcessor := &mockFileProcessor{}
	poller, _ := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithKeepRunning(true).
		WithMaxPollCycles(1).
		WithProcessor(mockProcessor).
		Build()

	count, err := poller.PollDirectory(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.False(t, mockProcessor.called)
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

	mockProcessor := &mockFileProcessor{}
	poller, _ := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithMaxPollCycles(1).
		WithProcessor(mockProcessor).
		Build()

	count, err := poller.PollDirectory(context.Background())

	assert.NoError(t, err)
	// Should only count the 2 files in root, not the subdir or file inside it
	assert.Equal(t, 2, count)
	assert.Equal(t, 2, mockProcessor.callCount)
}

func TestProcessFileWithDeletion(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	createTempFile(t, tmpDir, "test_file.txt", "test message content")

	// Verify file exists before polling
	assert.FileExists(t, tmpDir+"/test_file.txt")

	mockProcessor := &mockFileProcessor{}
	poller, _ := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithMaxPollCycles(1).
		WithDeleteFiles(true).
		WithProcessor(mockProcessor).
		Build()

	count, err := poller.PollDirectory(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.True(t, mockProcessor.called)
	assert.Equal(t, 1, mockProcessor.callCount)
	assert.Equal(t, "test message content", mockProcessor.capturedContent[0].Body)
	assert.NoFileExists(t, tmpDir+"/test_file.txt", "file should be deleted after processing")
}

func TestProcessorErrorHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	tmpDir := t.TempDir()

	// Create test files
	createTempFile(t, tmpDir, "file1.txt", "content1")
	createTempFile(t, tmpDir, "file2.txt", "content2")

	mockProcessor := &mockFileProcessor{
		processFunc: func(ctx context.Context, content processor.Message) error {
			if content.Body == "content1" {
				return fmt.Errorf("processor error on first file")
			}
			return nil
		},
	}

	poller, _ := NewDirectoryPollerBuilder(logger).
		WithMessageLocation(tmpDir).
		WithMaxPollCycles(1).
		WithProcessor(mockProcessor).
		Build()

	count, err := poller.PollDirectory(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Equal(t, 2, mockProcessor.callCount)
	assert.Equal(t, 2, len(mockProcessor.capturedContent))
}
