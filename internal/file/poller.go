package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"kafka_go_cli/internal/processor"
)

// DirectoryPoller scans a directory for files and processes them with a configurable strategy.
type DirectoryPoller struct {
	dirPath            string
	keepRunning        bool
	deleteFiles        bool
	pollIntervalMillis time.Duration
	maxPollCycles      int // -1 = unlimited
	processor          processor.FileProcessor
	logger             *slog.Logger
	workerCount        int // number of concurrent workers
}

// DirectoryPollerBuilder provides a fluent API for constructing a DirectoryPoller.
type DirectoryPollerBuilder struct {
	messageLocation    string
	keepRunning        bool
	deleteFiles        bool
	pollIntervalMillis time.Duration
	maxPollCycles      int
	processor          processor.FileProcessor
	logger             *slog.Logger
	workerCount        int
}

// NewDirectoryPollerBuilder creates a new builder for constructing a DirectoryPoller.
// The logger is required and passed explicitly, following the project's dependency injection pattern.
func NewDirectoryPollerBuilder(logger *slog.Logger) *DirectoryPollerBuilder {
	return &DirectoryPollerBuilder{
		logger: logger,
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

func (b *DirectoryPollerBuilder) WithProcessor(p processor.FileProcessor) *DirectoryPollerBuilder {
	b.processor = p
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
		processor:          b.processor,
		logger:             b.logger,
		workerCount:        b.workerCount,
	}, nil
}

// DirectoryPoller methods
//
// PollDirectory scans a directory for files and processes them according to the poller configuration.
// It validates the directory, enters a polling loop, and processes files with the configured FileProcessor.
// Context can be used to signal cancellation to stop polling gracefully.
func (dp *DirectoryPoller) PollDirectory(ctx context.Context) (int, error) {
	if err := ValidateDirectory(dp.dirPath); err != nil {
		return 0, err
	}

	totalFileCount := 0
	keepRunning := true
	pollCycles := 0

	// the while loop in Go uses a for loop
	for keepRunning {
		pollCycles++
		dp.logger.Info("polling directory...", "poll-count", pollCycles)

		// get the files in the directory
		// available for this polling cycle
		entries, err := os.ReadDir(dp.dirPath)
		if err != nil {
			return 0, err
		}

		count := 0
		for _, entry := range entries {
			// avoid processing directories
			if !entry.IsDir() {
				count++
				dp.processFile(ctx, dp.dirPath+"/"+entry.Name())
			}
		}

		totalFileCount += count
		keepRunning = dp.shouldContinuePolling(pollCycles)

		if keepRunning {
			// Use select to allow context cancellation to interrupt
			// the polling of the directory, enabling graceful shutdown.
			select {
			case <-ctx.Done():
				dp.logger.Info("polling interrupted via context cancellation")
				return totalFileCount, nil
			case <-time.After(dp.pollIntervalMillis):
				// Sleep interval completed normally
			}
		}
	}

	return totalFileCount, nil
}

// shouldContinuePolling determines if the polling loop should continue.
func (dp *DirectoryPoller) shouldContinuePolling(pollCycles int) bool {
	continuePolling := true

	// run once will take precedence over max cycles if both are set
	if !dp.keepRunning {
		continuePolling = false
	} else if dp.maxPollCycles > 0 && pollCycles >= dp.maxPollCycles {
		continuePolling = false
	}

	return continuePolling
}

// processFile reads the file content, parses it and passes it to the processor.
// Errors are logged but don't stop polling (resilience pattern).
func (dp *DirectoryPoller) processFile(ctx context.Context, filePath string) error {
	// read file content (ReadFile handles all open/close operations)
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		dp.logger.Error("failed to read the file", "file", filePath, "error", err)
	}

	fileContents := string(bytes)

	dp.logger.Debug("processing file", "file", filePath, "size", len(bytes))

	msg := ParseMessage(fileContents)

	// let the processor handle the file content
	if err := dp.processor.Process(ctx, msg); err != nil {
		dp.logger.Error("failed to process the file", "file", filePath, "error", err)
	}

	// delete file
	dp.deleteFile(filePath)

	return nil
}

// deleteFile removes the file if the deleteFiles flag is enabled.
// Errors are logged but don't cause the polling to stop.
func (dp *DirectoryPoller) deleteFile(filePath string) {
	if dp.deleteFiles {
		if err := os.Remove(filePath); err != nil {
			dp.logger.Error("failed to delete file", "file", filePath, "error", err)
		}
	}
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
