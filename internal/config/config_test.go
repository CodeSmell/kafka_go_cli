package config

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// drives testing of the runE function's config loading logic by simulating command execution
// with various args and config files that are specified in test functions.
func loadWithArgs(t *testing.T, args ...string) (Settings, error) {
	t.Helper()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	// this mimics the flags defined in NewRootCommand
	cmd.PersistentFlags().String("config", "", "Path to config file (optional)")
	cmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	cmd.PersistentFlags().String("message-location", "", "Directory path to scan for message files")
	cmd.PersistentFlags().Bool("run-once", false, "Run a single scan cycle then exits")
	cmd.PersistentFlags().Bool("no-delete-files", true, "Do not delete files after successful processing")
	cmd.PersistentFlags().Int("delay", 0, "Number of ms to wait between polling cycles")
	cmd.PersistentFlags().Int("max-cycles", 0, "Max number of times to poll (default -1 polls indefinitely)")
	cmd.PersistentFlags().String("processor", "noop", "Processor type (kafka, pulsar, noop)")

	var got Settings
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		settings, err := Load(cmd)
		got = settings
		return err
	}

	cmd.SetArgs(args)
	err := cmd.Execute()
	return got, err
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	assert.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestLoadNewFieldsDefaults(t *testing.T) {
	s, err := loadWithArgs(t)
	assert.Equal(t, "", s.ConfigFile)
	assert.NoError(t, err)
	assert.Equal(t, "", s.MessageLocation)
	assert.False(t, s.RunOnce)
	assert.True(t, s.NoDeleteFiles)
	assert.Equal(t, 1000, s.Delay)
	assert.Equal(t, -1, s.MaxCycles)
	assert.Equal(t, "noop", s.ProcessorType)
}

func TestLoadPrecedence_ConfigOverridesDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := writeTempConfig(t, "run_once: true\nlog_level: debug\ndelay: 500\nmax_cycles: 10\nprocessor: kafka\nmessage_location: "+tmpDir+"\n")

	s, err := loadWithArgs(t, "--config", cfgPath)
	assert.NoError(t, err)
	assert.Equal(t, cfgPath, s.ConfigFile)
	assert.Equal(t, "debug", s.LogLevel)
	assert.Equal(t, tmpDir, s.MessageLocation)
	assert.True(t, s.RunOnce)
	assert.True(t, s.NoDeleteFiles) // default is true, not overridden by config
	assert.Equal(t, 500, s.Delay)
	assert.Equal(t, 10, s.MaxCycles)
	assert.Equal(t, "kafka", s.ProcessorType)
}

func TestLoadPrecedence_EnvOverridesConfig(t *testing.T) {
	t.Setenv("KAFKA_GO_CLI_LOG_LEVEL", "warn")
	t.Setenv("KAFKA_GO_CLI_DELAY", "1500")
	t.Setenv("KAFKA_GO_CLI_PROCESSOR", "noop")

	cfgPath := writeTempConfig(t, "log_level: debug\ndelay: 500\nprocessor: kafka\n")
	s, err := loadWithArgs(t, "--config", cfgPath)
	assert.NoError(t, err)
	assert.Equal(t, "warn", s.LogLevel)
	assert.Equal(t, 1500, s.Delay)
	assert.Equal(t, "noop", s.ProcessorType)
}

func TestLoadPrecedence_CLIOverridesEnvAndConfig(t *testing.T) {
	t.Setenv("KAFKA_GO_CLI_LOG_LEVEL", "warn")
	t.Setenv("KAFKA_GO_CLI_DELAY", "1500")
	t.Setenv("KAFKA_GO_CLI_PROCESSOR", "pulsar")

	cfgPath := writeTempConfig(t, "log_level: debug\ndelay: 500\nprocessor: noop\n")
	s, err := loadWithArgs(t, "--config", cfgPath, "--log-level", "error", "--delay", "2000", "--max-cycles", "20", "--processor", "kafka")
	assert.NoError(t, err)
	assert.Equal(t, "error", s.LogLevel)
	assert.Equal(t, 2000, s.Delay)
	assert.Equal(t, 20, s.MaxCycles)
	assert.Equal(t, "kafka", s.ProcessorType)
}

func TestLoadExplicitConfigPathMissingIsError(t *testing.T) {
	_, err := loadWithArgs(t, "--config", "/path/does/not/exist.yaml")
	assert.Error(t, err)
}
