package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"
)

// FileProcessor handles the content of each file found during polling.
// Implementations can publish to Kafka, write to logs, or perform other actions.
type FileProcessor interface {
	// Process handles file content and returns an error if processing fails.
	Process(ctx context.Context, content string) error
}

// DirectoryPoller scans a directory for files and processes them with a configurable strategy.
type DirectoryPoller struct {
	dirPath            string
	keepRunning        bool
	deleteFiles        bool
	pollIntervalMillis time.Duration
	maxPollCycles      int // -1 or 0 = unlimited
	workerCount        int // number of concurrent workers
	processor          FileProcessor
	logger             *slog.Logger
}

// DirectoryPollerBuilder provides a fluent API for constructing a DirectoryPoller.
type DirectoryPollerBuilder struct {
	messageLocation    string
	keepRunning        bool
	deleteFiles        bool
	pollIntervalMillis time.Duration
	maxPollCycles      int
	workerCount        int
	processor          FileProcessor
	logger             *slog.Logger
}

// NewDirectoryPollerBuilder creates a new builder for constructing a DirectoryPoller.
// The logger is required and passed explicitly, following the project's dependency injection pattern.
func NewDirectoryPollerBuilder(logger *slog.Logger) *DirectoryPollerBuilder {
	return &DirectoryPollerBuilder{
		logger:             logger,
		keepRunning:        false,
		deleteFiles:        false,
		pollIntervalMillis: 1000 * time.Millisecond,
		maxPollCycles:      -1,
		workerCount:        1,
	}
}

func (b *DirectoryPollerBuilder) WithKeepRunning(keep bool) *DirectoryPollerBuilder {
	b.keepRunning = keep
	return b
}

func (b *DirectoryPollerBuilder) WithDeleteFiles(delete bool) *DirectoryPollerBuilder {
	b.deleteFiles = delete
	return b
}

func (b *DirectoryPollerBuilder) WithMessageLocation(location string) *DirectoryPollerBuilder {
	b.messageLocation = location
	return b
}

func (b *DirectoryPollerBuilder) WithPollInterval(interval time.Duration) *DirectoryPollerBuilder {
	b.pollIntervalMillis = interval
	return b
}

func (b *DirectoryPollerBuilder) WithMaxPollCycles(maxCycles int) *DirectoryPollerBuilder {
	b.maxPollCycles = maxCycles
	return b
}

func (b *DirectoryPollerBuilder) WithWorkerCount(count int) *DirectoryPollerBuilder {
	b.workerCount = count
	return b
}

func (b *DirectoryPollerBuilder) WithProcessor(processor FileProcessor) *DirectoryPollerBuilder {
	b.processor = processor
	return b
}

// Build constructs and returns a configured DirectoryPoller.
// Returns an error if no processor or message location are set.
func (b *DirectoryPollerBuilder) Build() (*DirectoryPoller, error) {
	if b.processor == nil {
		return nil, fmt.Errorf("a file processor is required")
	}
	if b.messageLocation == "" {
		return nil, fmt.Errorf("a message location is required")
	}
	return &DirectoryPoller{
		dirPath:            b.messageLocation,
		keepRunning:        b.keepRunning,
		deleteFiles:        b.deleteFiles,
		pollIntervalMillis: b.pollIntervalMillis,
		maxPollCycles:      b.maxPollCycles,
		workerCount:        b.workerCount,
		processor:          b.processor,
		logger:             b.logger,
	}, nil
}

// DirectoryPoller methods

// PollDirectory scans a directory for files and processes them according to the poller configuration.
// It validates the directory, enters a polling loop, and processes files with the configured FileProcessor.
// Context can be used to signal cancellation to stop polling gracefully.
//
// TODO: Implement full polling loop with:
// - Directory validation
// - File iteration and filtering (regular files only)
// - Context cancellation support
// - Poll cycle tracking and shouldContinuePolling logic
func (dp *DirectoryPoller) PollDirectory(ctx context.Context, dirPath string) (int, error) {
	dp.logger.Info("polling directory...")
	if err := ValidateDirectory(dirPath); err != nil {
		return 0, err
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}

	return count, nil
}

// shouldContinuePolling determines if the polling loop should continue.
// Returns false when maxPollCycles is reached, otherwise respects keepRunning flag.
// If continuing, sleeps for the configured poll interval.
//
// TODO: Implement with:
// - maxPollCycles limit check
// - keepRunning flag check
// - sleep between cycles
func (dp *DirectoryPoller) shouldContinuePolling(pollCycles int) bool {
	// TODO: Implement
	return false
}

// processFile reads the file content and passes it to the processor.
// Errors are logged but don't stop polling (resilience pattern).
//
// TODO: Implement with:
// - File content reading
// - Processor.Process() call
// - Resilient error handling (log but continue)
// - deleteFile() call on success
func (dp *DirectoryPoller) processFile(ctx context.Context, filePath string) error {
	// TODO: Implement
	return nil
}

// deleteFile removes the file if the deleteFiles flag is enabled.
// Errors are logged but don't cause the polling to stop.
//
// TODO: Implement with:
// - deleteFiles flag check
// - os.Remove() call
// - Error logging without propagation
func (dp *DirectoryPoller) deleteFile(filePath string) {
	// TODO: Implement
}

// ValidateDirectory checks if a path exists and is a directory.
// Used as a package-level utility for simple directory validation.
func ValidateDirectory(path string) error {
	if path == "" {
		return fmt.Errorf("directory path is required")
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("path must be a directory: %s", path)
	}

	return nil
}
