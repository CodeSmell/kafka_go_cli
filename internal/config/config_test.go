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

	cmd.PersistentFlags().String("config", "", "Path to config file (optional)")
	cmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	cmd.PersistentFlags().String("message-location", "", "Directory path to scan for message files")
	cmd.PersistentFlags().Bool("run-once", false, "Run a single scan cycle then exits")
	cmd.PersistentFlags().Bool("no-delete-files", true, "Do not delete files after successful processing")

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

func TestLoadPrecedence_ConfigOverridesDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := writeTempConfig(t, "run_once: true\nlog_level: debug\nmessage_location: "+tmpDir+"\n")

	s, err := loadWithArgs(t, "--config", cfgPath)
	assert.NoError(t, err)
	assert.Equal(t, "debug", s.LogLevel)
	assert.True(t, s.RunOnce)
	assert.True(t, s.NoDeleteFiles) // default is true, not overridden by config
	assert.Equal(t, tmpDir, s.MessageLocation)
	assert.Equal(t, cfgPath, s.ConfigFile)
}

func TestLoadPrecedence_EnvOverridesConfig(t *testing.T) {
	t.Setenv("KAFKA_GO_CLI_LOG_LEVEL", "warn")

	cfgPath := writeTempConfig(t, "log_level: debug\n")
	s, err := loadWithArgs(t, "--config", cfgPath)
	assert.NoError(t, err)
	assert.Equal(t, "warn", s.LogLevel)
}

func TestLoadPrecedence_CLIOverridesEnvAndConfig(t *testing.T) {
	t.Setenv("KAFKA_GO_CLI_LOG_LEVEL", "warn")

	cfgPath := writeTempConfig(t, "log_level: debug\n")
	s, err := loadWithArgs(t, "--config", cfgPath, "--log-level", "error")
	assert.NoError(t, err)
	assert.Equal(t, "error", s.LogLevel)
}

func TestLoadExplicitConfigPathMissingIsError(t *testing.T) {
	_, err := loadWithArgs(t, "--config", "/path/does/not/exist.yaml")
	assert.Error(t, err)
}
