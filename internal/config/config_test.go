package config

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"kafka_go_cli/internal/model"
	"kafka_go_cli/internal/processor"
)

// drives testing of the runE function's config loading logic by simulating command execution
// with various args and config files that are specified in test functions.
func loadWithArgs(t *testing.T, args ...string) (model.Settings, error) {
	return loadWithArgsAndExtraFlags(t, nil, args...)
}

func loadWithArgsAndExtraFlags(t *testing.T, extraFlags []string, args ...string) (model.Settings, error) {
	t.Helper()

	cmd := &cobra.Command{Use: "test"}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	// this mimics registering the flags in Cobra (NewRootCommand)
	cmd.PersistentFlags().String("config", "", "Path to config file (optional)")
	cmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	cmd.PersistentFlags().String("message-location", "", "Directory path to scan for message files")
	cmd.PersistentFlags().Bool("run-once", false, "Run a single scan cycle then exits")
	cmd.PersistentFlags().Bool("no-delete-files", true, "Do not delete files after successful processing")
	cmd.PersistentFlags().Int("delay", 0, "Number of ms to wait between polling cycles")
	cmd.PersistentFlags().Int("max-cycles", 0, "Max number of times to poll (default -1 polls indefinitely)")
	cmd.PersistentFlags().String("processor", "noop", "Processor type (kafka, pulsar, noop)")

	// this mimics registration of processor-specific flags in Cobra (NewRootCommand)
	for _, flagName := range extraFlags {
		cmd.PersistentFlags().String(flagName, "", "test")
	}

	var got model.Settings
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		settings, err := Load(cmd)
		got = settings
		return err
	}

	cmd.SetArgs(args)
	err := cmd.Execute()
	return got, err
}

func writeTempConfigFile(t *testing.T, content string) string {
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
	cfgPath := writeTempConfigFile(t, "run_once: true\nlog_level: debug\ndelay: 500\nmax_cycles: 10\nprocessor: kafka\nmessage_location: "+tmpDir+"\n")

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

	cfgPath := writeTempConfigFile(t, "log_level: debug\ndelay: 500\nprocessor: kafka\n")
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

	cfgPath := writeTempConfigFile(t, "log_level: debug\ndelay: 500\nprocessor: noop\n")
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

func TestLoad_ProcessorSpecificSettings(t *testing.T) {
	// simulate registering processor-specific config params
	// that would normally happen in the processor package's init() function
	// and be available when config needs them during loading
	processor.RegisterConfigParams("mock", []model.ConfigParam{
		{Name: "param1", Flag: "mock-param1", Description: "mock param1"},
		{Name: "param2", Flag: "mock-param2", Description: "mock param2"},
		{Name: "param3", Flag: "mock-param3", Description: "mock param3"},
	})

	t.Setenv("KAFKA_GO_CLI_MOCK_PARAM3", "env")

	cfgPath := writeTempConfigFile(t, "processor: mock\nmock:\n  param1: configFile\n")

	extraFlags := []string{"mock-param1", "mock-param2", "mock-param3"}
	s, err := loadWithArgsAndExtraFlags(t, extraFlags, "--config", cfgPath, "--mock-param2", "cli")

	assert.NoError(t, err)
	assert.Equal(t, "mock", s.ProcessorType)
	assert.Equal(t, "configFile", s.ProcessorConfig["param1"])
	assert.Equal(t, "cli", s.ProcessorConfig["param2"])
	assert.Equal(t, "env", s.ProcessorConfig["param3"])
}

func TestLoad_ProcessorSpecificSettings_WithOverrides(t *testing.T) {
	// simulate registering processor-specific config params
	// that would normally happen in the processor package's init() function
	// and be available when config needs them during loading
	processor.RegisterConfigParams("mock", []model.ConfigParam{
		{Name: "param1", Flag: "mock-param1", Description: "mock param1"},
		{Name: "param2", Flag: "mock-param2", Description: "mock param2"},
		{Name: "param3", Flag: "mock-param3", Description: "mock param3"},
	})

	t.Setenv("KAFKA_GO_CLI_MOCK_PARAM1", "env1")
	t.Setenv("KAFKA_GO_CLI_MOCK_PARAM2", "env2")

	cfgPath := writeTempConfigFile(t, "processor: mock\nmock:\n  param1: configFile\n")

	extraFlags := []string{"mock-param1", "mock-param2", "mock-param3"}
	s, err := loadWithArgsAndExtraFlags(t, extraFlags, "--config", cfgPath, "--mock-param2", "cli2")

	assert.NoError(t, err)
	assert.Equal(t, "mock", s.ProcessorType)
	// env should override config file
	assert.Equal(t, "env1", s.ProcessorConfig["param1"])
	// CLI should override env
	assert.Equal(t, "cli2", s.ProcessorConfig["param2"])
}
