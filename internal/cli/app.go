package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	// loads config package for side-effect of
	// registering config struct with viper
	"kafka_go_cli/internal/config"
	"kafka_go_cli/internal/file"
	"kafka_go_cli/internal/logging"

	"github.com/spf13/cobra"
)

var errCheckFailed = errors.New("configuration check failed")
var errMissingMessageLocation = errors.New("message-location is required")

// runE is the single entry point for the CLI.
// It supports two modes:
// - default: run the scanner
// - --check: validate resolved configuration and exit
func runE(cmd *cobra.Command, args []string) error {
	checkOnly, _ := cmd.Flags().GetBool("check")

	settings, err := config.Load(cmd)
	if err != nil {
		if checkOnly {
			printProblems(cmd, []string{fmt.Sprintf("config load failed: %v", err)})
			return errCheckFailed
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "error: %v\n", err)
		return err
	}

	if checkOnly {
		return runCheck(cmd, settings)
	}

	logger, err := logging.New(settings.LogLevel)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "error: %v\n", err)
		return err
	}

	if settings.MessageLocation == "" {
		fmt.Fprintln(cmd.ErrOrStderr(), "error: message-location is required")
		_ = cmd.Usage()
		logger.Error("invalid configuration", "error", errMissingMessageLocation)
		return errMissingMessageLocation
	}

	logResolvedSettings(logger, settings)
	return runScan(cmd, logger, settings)
}

func logResolvedSettings(logger *slog.Logger, settings config.Settings) {
	logger.Info("--- resolved configuration ---")
	if settings.ConfigFile == "" {
		logger.Info("config", "used", false)
	} else {
		logger.Info("config", "used", true, "path", settings.ConfigFile)
	}

	logger.Info("log-level", "value", settings.LogLevel)
	logger.Info("message-location", "value", settings.MessageLocation)
	logger.Info("run-once", "value", settings.RunOnce)
	logger.Info("no-delete-files", "value", settings.NoDeleteFiles)
	logger.Info("delay", "value", settings.Delay)
	logger.Info("max-cycles", "value", settings.MaxCycles)
	logger.Info("no-op", "value", settings.NoOp)
}

func runCheck(cmd *cobra.Command, settings config.Settings) error {
	problems := make([]string, 0, 4)

	if _, err := logging.New(settings.LogLevel); err != nil {
		problems = append(problems, fmt.Sprintf("log-level invalid: %v", err))
	}

	if settings.MessageLocation == "" {
		problems = append(problems, "message-location is required")
	} else if err := file.ValidateDirectory(settings.MessageLocation); err != nil {
		problems = append(problems, fmt.Sprintf("message-location invalid: %v", err))
	}

	if len(problems) > 0 {
		printProblems(cmd, problems)
		return errCheckFailed
	}

	// Success output is intentionally short.
	fmt.Fprintln(cmd.OutOrStdout(), "OK")
	if settings.ConfigFile != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Config: %s\n", settings.ConfigFile)
	}
	return nil
}

func printProblems(cmd *cobra.Command, problems []string) {
	fmt.Fprintln(cmd.OutOrStdout(), "PROBLEMS:")
	for _, p := range problems {
		fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", p)
	}
}

func runScan(cmd *cobra.Command, logger *slog.Logger, settings config.Settings) error {
	logger.Info("--- starting directory scan ---")

	processor := &NoOpProcessor{}
	poller, err := file.NewDirectoryPollerBuilder(logger).
		WithMessageLocation(settings.MessageLocation).
		WithDeleteFiles(!settings.NoDeleteFiles).
		WithKeepRunning(!settings.RunOnce).
		WithMaxPollCycles(settings.MaxCycles).
		WithPollInterval(time.Duration(settings.Delay) * time.Millisecond).
		WithNoOpProcessor(settings.NoOp).
		WithProcessor(processor).
		Build()
	if err != nil {
		logger.Error("failed to build directory poller", "error", err)
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fileCount, err := poller.PollDirectory(ctx)
	if err != nil {
		logger.Error("directory scan failed", "error", err)
		return err
	}

	logger.Info("directory scan completed", "total-file-count", fileCount)
	fmt.Fprintf(cmd.OutOrStdout(), "Found %d file(s) in %s\n", fileCount, settings.MessageLocation)
	return nil
}

// NoOpProcessor is a temporary FileProcessor that does nothing.
// TODO: Replace with real implementations (KafkaFileProcessor, PulsarFileProcessor, etc.)
type NoOpProcessor struct{}

// Process implements the FileProcessor interface with a no-op operation.
func (n *NoOpProcessor) Process(ctx context.Context, content string) error {
	// TODO: Implement actual processing (publish to Kafka, Pulsar, etc.)
	return nil
}
